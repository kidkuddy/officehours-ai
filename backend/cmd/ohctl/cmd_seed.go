package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"officehours/internal/agent"
	"officehours/internal/auth"
	"officehours/internal/rag"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

// seedDir resolves SEED_DIR (default /seed).
func seedDir() string {
	if v := os.Getenv("SEED_DIR"); v != "" {
		return v
	}
	return "/seed"
}

func runSeedUser(cmd *cobra.Command, _ []string) error {
	email, _ := cmd.Flags().GetString("email")
	password, _ := cmd.Flags().GetString("password")
	name, _ := cmd.Flags().GetString("name")
	if email == "" || password == "" || name == "" {
		return fmt.Errorf("--email, --password and --name are required")
	}

	ctx, cancel := newCtx()
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	id, created, err := seedUser(ctx, pool, email, password, name)
	if err != nil {
		return err
	}
	printJSON(map[string]any{"id": id, "email": email, "name": name, "created": created})
	return nil
}

func runSeedDemo(cmd *cobra.Command, _ []string) error {
	withUser, _ := cmd.Flags().GetBool("with-user")

	ctx, cancel := context.WithTimeout(context.Background(), 300_000_000_000) // 5m
	defer cancel()
	pool, err := openPool(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	dir := seedDir()
	out := map[string]any{}

	// 1. Parse advisor + learn + agent definitions (validation / surfacing).
	advisors := parseDefs(filepath.Join(dir, "advisors"))
	learn := parseDefs(filepath.Join(dir, "learn"))
	agents := parseDefs(filepath.Join(dir, "agents"))
	out["advisors"] = defSummaries(advisors)
	out["learn"] = defSummaries(learn)
	out["agents"] = defSummaries(agents)

	// 2. Index /seed/kb/<collection>/* into their collections (shared KB).
	kbDir := filepath.Join(dir, "kb")
	indexed := []any{}
	entries, err := os.ReadDir(kbDir)
	if err != nil {
		out["kb_error"] = fmt.Sprintf("read kb dir %s: %v", kbDir, err)
	} else {
		for _, e := range entries {
			if !e.IsDir() {
				continue
			}
			collection := e.Name()
			// Re-index cleanly: drop existing shared docs for this collection.
			if err := purgeCollection(ctx, pool, collection); err != nil {
				return fmt.Errorf("purge collection %s: %w", collection, err)
			}
			res, err := rag.IndexFolder(ctx, pool, filepath.Join(kbDir, collection), collection, nil)
			if err != nil {
				return fmt.Errorf("index %s: %w", collection, err)
			}
			indexed = append(indexed, res)
		}
	}
	out["indexed"] = indexed

	// 3. Optional demo user (from /seed/profiles/example-agritech.md if present).
	if withUser {
		id, created, err := seedUser(ctx, pool, "founder@demo.officehours.ai", "demo1234", "Demo Founder")
		if err != nil {
			return err
		}
		// Seed the company_text + default goal so the critical path is ready.
		companyText := readProfileExample(filepath.Join(dir, "profiles", "example-agritech.md"))
		if companyText != "" {
			_, err = pool.Exec(ctx,
				`update profiles set company_text=$2, updated_at=now() where user_id=$1`,
				id, companyText)
			if err != nil {
				return err
			}
		}
		ensureDefaultGoal(ctx, pool, id)
		out["demo_user"] = map[string]any{
			"id": id, "email": "founder@demo.officehours.ai",
			"password": "demo1234", "created": created,
		}
	}

	out["ok"] = true
	printJSON(out)
	return nil
}

// seedUser creates a user (and an empty profile) idempotently by email.
func seedUser(ctx context.Context, pool *pgxpool.Pool, email, password, name string) (id string, created bool, err error) {
	// Already exists?
	err = pool.QueryRow(ctx, `select id from users where email=$1`, email).Scan(&id)
	if err == nil {
		return id, false, nil
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return "", false, err
	}
	err = pool.QueryRow(ctx,
		`insert into users (email, password_hash, name) values ($1,$2,$3) returning id`,
		email, hash, name).Scan(&id)
	if err != nil {
		return "", false, err
	}
	// Create the matching profile row (default Ideation stage).
	_, err = pool.Exec(ctx,
		`insert into profiles (user_id) values ($1) on conflict (user_id) do nothing`, id)
	if err != nil {
		return "", false, err
	}
	return id, true, nil
}

func ensureDefaultGoal(ctx context.Context, pool *pgxpool.Pool, userID string) {
	var exists bool
	_ = pool.QueryRow(ctx,
		`select exists(select 1 from goals where user_id=$1 and title=$2)`,
		userID, "Start an Office Hours session").Scan(&exists)
	if !exists {
		_, _ = pool.Exec(ctx,
			`insert into goals (user_id, title) values ($1,$2)`,
			userID, "Start an Office Hours session")
	}
}

// purgeCollection deletes shared-KB documents (and cascading chunks) for a
// collection, so seed demo can re-run cleanly.
func purgeCollection(ctx context.Context, pool *pgxpool.Pool, collection string) error {
	_, err := pool.Exec(ctx,
		`delete from documents where collection=$1 and user_id is null`, collection)
	return err
}

func parseDefs(dir string) []*agent.Definition {
	defs, err := agent.LoadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ohctl: warning: load defs from %s: %v\n", dir, err)
		return nil
	}
	return defs
}

func defSummaries(defs []*agent.Definition) []map[string]any {
	out := []map[string]any{}
	for _, d := range defs {
		out = append(out, map[string]any{
			"key": d.Key, "name": d.Name, "kind": d.Kind,
			"collection": d.Collection, "enabled": d.Enabled,
			"order": d.Order, "description": d.Description,
		})
	}
	return out
}

func readProfileExample(path string) string {
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(b)
}
