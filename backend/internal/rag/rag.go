// Package rag implements the full-text retrieval pipeline: chunking of
// md/txt/pdf files, tsvector indexing, and FTS query via ts_rank.
// See BUILD_SPEC §3 (RAG indexing) and §2 (chunks table).
package rag

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ledongthuc/pdf"
)

// ChunkSize is the target chunk length in characters (paragraph-aware).
const ChunkSize = 800

// IndexResult summarizes an indexing run.
type IndexResult struct {
	Collection string   `json:"collection"`
	Folder     string   `json:"folder"`
	Files      int      `json:"files"`
	Chunks     int      `json:"chunks"`
	Skipped    []string `json:"skipped"`
	UserID     *string  `json:"user_id,omitempty"`
}

// QueryHit is one FTS result.
type QueryHit struct {
	ChunkID    string  `json:"chunk_id"`
	DocumentID string  `json:"document_id"`
	Filename   string  `json:"filename"`
	Collection string  `json:"collection"`
	Ord        int     `json:"ord"`
	Rank       float64 `json:"rank"`
	Content    string  `json:"content"`
}

// IndexFolder walks folder, extracts text from md/txt/pdf, chunks it, and stores
// documents + chunks (with tsvector) for the given collection. userID may be nil
// for shared KB. PDF extraction failures are skipped and reported, not fatal.
func IndexFolder(ctx context.Context, pool *pgxpool.Pool, folder, collection string, userID *string) (*IndexResult, error) {
	if collection == "" {
		return nil, fmt.Errorf("rag: collection is required")
	}
	info, err := os.Stat(folder)
	if err != nil {
		return nil, fmt.Errorf("rag: stat folder: %w", err)
	}

	res := &IndexResult{Collection: collection, Folder: folder, UserID: userID, Skipped: []string{}}

	var files []string
	if info.IsDir() {
		err = filepath.WalkDir(folder, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				return nil
			}
			if supportedExt(path) {
				files = append(files, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("rag: walk: %w", err)
		}
	} else if supportedExt(folder) {
		files = []string{folder}
	}

	for _, f := range files {
		text, err := ExtractText(f)
		if err != nil {
			res.Skipped = append(res.Skipped, fmt.Sprintf("%s: %v", f, err))
			fmt.Fprintf(os.Stderr, "rag: warning: skipping %s: %v\n", f, err)
			continue
		}
		n, err := IndexFile(ctx, pool, f, text, collection, userID)
		if err != nil {
			res.Skipped = append(res.Skipped, fmt.Sprintf("%s: %v", f, err))
			fmt.Fprintf(os.Stderr, "rag: warning: index failed %s: %v\n", f, err)
			continue
		}
		res.Files++
		res.Chunks += n
	}
	return res, nil
}

// IndexFile stores a single document (by path/text) and its chunks.
func IndexFile(ctx context.Context, pool *pgxpool.Pool, path, text, collection string, userID *string) (int, error) {
	mime := mimeForExt(path)
	filename := filepath.Base(path)

	tx, err := pool.Begin(ctx)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback(ctx)

	var docID string
	err = tx.QueryRow(ctx,
		`insert into documents (user_id, collection, filename, mime, path)
		 values ($1,$2,$3,$4,$5) returning id`,
		userID, collection, filename, mime, path).Scan(&docID)
	if err != nil {
		return 0, err
	}

	chunks := Chunk(text, ChunkSize)
	for i, c := range chunks {
		_, err = tx.Exec(ctx,
			`insert into chunks (document_id, collection, user_id, ord, content, tsv)
			 values ($1,$2,$3,$4,$5, to_tsvector('english',$5))`,
			docID, collection, userID, i, c)
		if err != nil {
			return 0, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, err
	}
	return len(chunks), nil
}

// Query runs an FTS query over a collection (and optionally a user), ordered by
// ts_rank desc, limited to k.
func Query(ctx context.Context, pool *pgxpool.Pool, collection, q string, k int, userID *string) ([]QueryHit, error) {
	if collection == "" {
		return nil, fmt.Errorf("rag: collection is required")
	}
	if k <= 0 {
		k = 5
	}

	sql := `select c.id, c.document_id, coalesce(d.filename,''), c.collection, c.ord,
	               ts_rank(c.tsv, plainto_tsquery('english',$2)) as rank, c.content
	        from chunks c
	        left join documents d on d.id = c.document_id
	        where c.collection = $1 and c.tsv @@ plainto_tsquery('english',$2)`
	args := []any{collection, q}
	if userID != nil {
		sql += ` and c.user_id = $3`
		args = append(args, *userID)
		sql += ` order by rank desc limit $4`
		args = append(args, k)
	} else {
		sql += ` order by rank desc limit $3`
		args = append(args, k)
	}

	rows, err := pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	hits := []QueryHit{}
	for rows.Next() {
		var h QueryHit
		if err := rows.Scan(&h.ChunkID, &h.DocumentID, &h.Filename, &h.Collection, &h.Ord, &h.Rank, &h.Content); err != nil {
			return nil, err
		}
		hits = append(hits, h)
	}
	return hits, rows.Err()
}

// Chunk splits text into ~size-char paragraph-aware chunks. Paragraphs are kept
// whole when they fit; oversized paragraphs are hard-split.
func Chunk(text string, size int) []string {
	text = strings.TrimSpace(text)
	if text == "" {
		return nil
	}
	if size <= 0 {
		size = ChunkSize
	}

	paras := splitParagraphs(text)
	var chunks []string
	var cur strings.Builder

	flush := func() {
		if cur.Len() > 0 {
			chunks = append(chunks, strings.TrimSpace(cur.String()))
			cur.Reset()
		}
	}

	for _, p := range paras {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		// Oversized paragraph: hard-split into size-pieces.
		if len(p) > size {
			flush()
			for len(p) > size {
				cut := size
				// Try to break on a space near the boundary.
				if idx := strings.LastIndex(p[:cut], " "); idx > size/2 {
					cut = idx
				}
				chunks = append(chunks, strings.TrimSpace(p[:cut]))
				p = strings.TrimSpace(p[cut:])
			}
			if p != "" {
				cur.WriteString(p)
			}
			continue
		}
		// Would adding this paragraph overflow the current chunk?
		if cur.Len() > 0 && cur.Len()+len(p)+2 > size {
			flush()
		}
		if cur.Len() > 0 {
			cur.WriteString("\n\n")
		}
		cur.WriteString(p)
	}
	flush()
	return chunks
}

func splitParagraphs(text string) []string {
	// Split on blank lines.
	var paras []string
	var cur strings.Builder
	sc := bufio.NewScanner(strings.NewReader(text))
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.TrimSpace(line) == "" {
			if cur.Len() > 0 {
				paras = append(paras, cur.String())
				cur.Reset()
			}
			continue
		}
		if cur.Len() > 0 {
			cur.WriteString("\n")
		}
		cur.WriteString(line)
	}
	if cur.Len() > 0 {
		paras = append(paras, cur.String())
	}
	return paras
}

// ExtractText reads text from a md/txt/pdf file.
func ExtractText(path string) (string, error) {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".md", ".txt", ".markdown", ".text":
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(b), nil
	case ".pdf":
		return extractPDF(path)
	default:
		b, err := os.ReadFile(path)
		if err != nil {
			return "", err
		}
		return string(b), nil
	}
}

func extractPDF(path string) (string, error) {
	f, r, err := pdf.Open(path)
	if err != nil {
		return "", fmt.Errorf("pdf open: %w", err)
	}
	defer f.Close()
	var sb strings.Builder
	totalPage := r.NumPage()
	for i := 1; i <= totalPage; i++ {
		p := r.Page(i)
		if p.V.IsNull() {
			continue
		}
		txt, err := p.GetPlainText(nil)
		if err != nil {
			return "", fmt.Errorf("pdf page %d: %w", i, err)
		}
		sb.WriteString(txt)
		sb.WriteString("\n")
	}
	if strings.TrimSpace(sb.String()) == "" {
		return "", fmt.Errorf("pdf yielded no text")
	}
	return sb.String(), nil
}

func supportedExt(path string) bool {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".txt", ".markdown", ".text", ".pdf":
		return true
	}
	return false
}

func mimeForExt(path string) string {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown":
		return "text/markdown"
	case ".txt", ".text":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	}
	return "application/octet-stream"
}
