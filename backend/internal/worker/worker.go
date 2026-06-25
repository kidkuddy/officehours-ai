// Package worker implements the agent_jobs polling pool. Each job renders a
// prompt from the relevant /seed md + context and execs claude headless with
// ohctl on PATH and DATABASE_URL in env (BUILD_SPEC §1, §4).
package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"officehours/internal/agent"
	"officehours/internal/config"
	"officehours/internal/db"
	"officehours/internal/models"
	"officehours/internal/rag"

	"github.com/jackc/pgx/v5"
)

// Worker polls and executes agent_jobs.
type Worker struct {
	Cfg     *config.Config
	DB      *db.DB
	Workers int
	// OhctlDir is the directory containing the ohctl binary to put on PATH.
	OhctlDir string
}

// Run starts the polling pool and blocks until ctx is cancelled.
func (w *Worker) Run(ctx context.Context) {
	if w.DB == nil || w.DB.Pool == nil {
		log.Printf("job-worker: no database; worker disabled")
		<-ctx.Done()
		return
	}
	n := w.Workers
	if n <= 0 {
		n = 2
	}
	log.Printf("job-worker: starting pool of %d", n)
	jobs := make(chan *models.AgentJob)

	// Producers: a single poller fans claimed jobs to N executors.
	for i := 0; i < n; i++ {
		go func(id int) {
			for job := range jobs {
				w.execute(ctx, job)
			}
		}(i)
	}

	ticker := time.NewTicker(1500 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			log.Printf("job-worker: stopping")
			close(jobs)
			return
		case <-ticker.C:
			for {
				job, err := w.claim(ctx)
				if err != nil {
					if err != pgx.ErrNoRows {
						log.Printf("job-worker: claim error: %v", err)
					}
					break
				}
				if job == nil {
					break
				}
				select {
				case jobs <- job:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// claim atomically marks the oldest queued job as running and returns it.
func (w *Worker) claim(ctx context.Context) (*models.AgentJob, error) {
	var j models.AgentJob
	err := w.DB.Pool.QueryRow(ctx,
		`update agent_jobs set status='running', started_at=now()
		 where id = (
		   select id from agent_jobs where status='queued'
		   order by created_at for update skip locked limit 1
		 )
		 returning id, type, user_id, session_id, status, input, output, error, created_at, started_at, finished_at`).
		Scan(&j.ID, &j.Type, &j.UserID, &j.SessionID, &j.Status, &j.Input, &j.Output, &j.Error, &j.CreatedAt, &j.StartedAt, &j.FinishedAt)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return &j, nil
}

func (w *Worker) execute(ctx context.Context, job *models.AgentJob) {
	log.Printf("job-worker: executing job %s type=%s", job.ID, job.Type)
	reply, output, err := w.dispatch(ctx, job)
	if err != nil {
		w.fail(ctx, job, err)
		return
	}
	// For chat jobs, store the assistant reply as a message.
	if reply != "" && job.SessionID != nil &&
		(job.Type == models.JobTypeAdvisor || job.Type == models.JobTypeLearn) {
		_, derr := w.DB.Pool.Exec(ctx,
			`insert into messages (session_id, role, content) values ($1,'assistant',$2)`,
			*job.SessionID, reply)
		if derr != nil {
			w.fail(ctx, job, fmt.Errorf("store reply: %w", derr))
			return
		}
	}
	if output == nil {
		output = map[string]any{}
	}
	if reply != "" {
		output["reply"] = reply
	}
	out, _ := json.Marshal(output)
	_, err = w.DB.Pool.Exec(ctx,
		`update agent_jobs set status='done', output=$2, finished_at=now() where id=$1`,
		job.ID, out)
	if err != nil {
		log.Printf("job-worker: failed to mark job %s done: %v", job.ID, err)
	}
}

func (w *Worker) fail(ctx context.Context, job *models.AgentJob, jerr error) {
	log.Printf("job-worker: job %s failed: %v", job.ID, jerr)
	_, err := w.DB.Pool.Exec(ctx,
		`update agent_jobs set status='error', error=$2, finished_at=now() where id=$1`,
		job.ID, jerr.Error())
	if err != nil {
		log.Printf("job-worker: failed to mark job %s error: %v", job.ID, err)
	}
}

// dispatch renders the right prompt and execs claude. Returns the assistant
// reply (for chat jobs), an optional structured output map, and an error.
func (w *Worker) dispatch(ctx context.Context, job *models.AgentJob) (string, map[string]any, error) {
	switch job.Type {
	case "rag_index":
		return w.runRagIndex(ctx, job)
	case models.JobTypeAdvisor:
		return w.runChat(ctx, job, "advisors")
	case models.JobTypeLearn:
		return w.runChat(ctx, job, "learn")
	case models.JobTypeDiagnoser:
		return w.runDiagnoser(ctx, job)
	case models.JobTypeScorer:
		return w.runScorer(ctx, job)
	default:
		return "", nil, fmt.Errorf("unknown job type %q", job.Type)
	}
}

// runRagIndex indexes a single uploaded document (no claude needed).
func (w *Worker) runRagIndex(ctx context.Context, job *models.AgentJob) (string, map[string]any, error) {
	var in struct {
		DocumentID string `json:"document_id"`
		Path       string `json:"path"`
		Collection string `json:"collection"`
	}
	if err := json.Unmarshal(job.Input, &in); err != nil {
		return "", nil, fmt.Errorf("decode input: %w", err)
	}
	text, err := rag.ExtractText(in.Path)
	if err != nil {
		return "", nil, fmt.Errorf("extract text: %w", err)
	}
	// Chunk and store linked to the already-created document row.
	chunks := rag.Chunk(text, rag.ChunkSize)
	uid := job.UserID
	tx, err := w.DB.Pool.Begin(ctx)
	if err != nil {
		return "", nil, err
	}
	defer tx.Rollback(ctx)
	for i, c := range chunks {
		_, err = tx.Exec(ctx,
			`insert into chunks (document_id, collection, user_id, ord, content, tsv)
			 values ($1,$2,$3,$4,$5, to_tsvector('english',$5))`,
			in.DocumentID, in.Collection, uid, i, c)
		if err != nil {
			return "", nil, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return "", nil, err
	}
	return "", map[string]any{"chunks": len(chunks), "document_id": in.DocumentID}, nil
}

func (w *Worker) runChat(ctx context.Context, job *models.AgentJob, dir string) (string, map[string]any, error) {
	if job.SessionID == nil {
		return "", nil, fmt.Errorf("chat job missing session_id")
	}
	var advisorKey string
	err := w.DB.Pool.QueryRow(ctx,
		`select advisor_key from sessions where id=$1`, *job.SessionID).Scan(&advisorKey)
	if err != nil {
		return "", nil, fmt.Errorf("load session: %w", err)
	}
	def, err := agent.LoadByKey(filepath.Join(w.Cfg.SeedDir, dir), advisorKey)
	if err != nil {
		return "", nil, err
	}
	var in struct {
		Content string `json:"content"`
		Opening bool   `json:"opening"`
	}
	_ = json.Unmarshal(job.Input, &in)
	opening := in.Opening || strings.TrimSpace(in.Content) == ""

	prompt := agent.RenderAdvisor(def, agent.PromptContext{
		UserID:        job.UserID,
		SessionID:     *job.SessionID,
		AdvisorKey:    advisorKey,
		LatestMessage: in.Content,
		Opening:       opening,
	})
	reply, err := w.exec(ctx, prompt, def.Provider)
	if err != nil {
		return "", nil, err
	}
	return reply, map[string]any{"advisor_key": advisorKey}, nil
}

func (w *Worker) runDiagnoser(ctx context.Context, job *models.AgentJob) (string, map[string]any, error) {
	def, err := agent.LoadByKey(filepath.Join(w.Cfg.SeedDir, "agents"), "diagnoser")
	if err != nil {
		return "", nil, err
	}
	var in struct {
		Text string `json:"text"`
	}
	_ = json.Unmarshal(job.Input, &in)
	prompt := agent.RenderDiagnoser(def, agent.PromptContext{
		UserID:         job.UserID,
		OnboardingText: in.Text,
	})
	reply, err := w.exec(ctx, prompt, def.Provider)
	if err != nil {
		return "", nil, err
	}
	return "", map[string]any{"summary": reply}, nil
}

func (w *Worker) runScorer(ctx context.Context, job *models.AgentJob) (string, map[string]any, error) {
	if job.SessionID == nil {
		return "", nil, fmt.Errorf("scorer job missing session_id")
	}
	def, err := agent.LoadByKey(filepath.Join(w.Cfg.SeedDir, "agents"), "scorer")
	if err != nil {
		return "", nil, err
	}
	prompt := agent.RenderScorer(def, agent.PromptContext{
		UserID:    job.UserID,
		SessionID: *job.SessionID,
	})
	reply, err := w.exec(ctx, prompt, def.Provider)
	if err != nil {
		return "", nil, err
	}
	return "", map[string]any{"summary": reply}, nil
}

// exec sets up a temp workdir, builds the env (PATH with ohctl, DATABASE_URL,
// optional ANTHROPIC_API_KEY, Gemini/Vertex vars) and runs the selected agent
// backend headless. providerOverride is the per-agent `provider:` frontmatter
// value (may be empty), which takes precedence over AGENT_PROVIDER.
func (w *Worker) exec(ctx context.Context, prompt, providerOverride string) (string, error) {
	workdir, err := os.MkdirTemp("", "oh-job-")
	if err != nil {
		return "", fmt.Errorf("mkdir tempdir: %w", err)
	}
	defer os.RemoveAll(workdir)

	provider := agent.ResolveProvider(w.Cfg.AgentProvider, providerOverride)
	env := w.jobEnv()

	// Give the agent a generous but bounded timeout.
	execCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	log.Printf("job-worker: exec via provider=%s", provider)
	return agent.ExecWithProvider(execCtx, agent.ExecOptions{
		Prompt:   prompt,
		WorkDir:  workdir,
		Env:      env,
		Provider: provider,
	})
}

// jobEnv builds the environment for the agent subprocess.
func (w *Worker) jobEnv() []string {
	var env []string

	// Ensure ohctl is on PATH.
	path := os.Getenv("PATH")
	if w.OhctlDir != "" {
		path = w.OhctlDir + string(os.PathListSeparator) + path
	}
	env = append(env, "PATH="+path)

	// DATABASE_URL for ohctl.
	if w.Cfg.DatabaseURL != "" {
		env = append(env, "DATABASE_URL="+w.Cfg.DatabaseURL)
	} else if v := os.Getenv("DATABASE_URL"); v != "" {
		env = append(env, "DATABASE_URL="+v)
	}

	// SEED_DIR so ohctl rag query and agents resolve seed.
	env = append(env, "SEED_DIR="+w.Cfg.SeedDir)

	// Fallback API key (Plan B). If unset, claude uses mounted /root/.claude.
	if w.Cfg.AnthropicAPIKey != "" {
		env = append(env, "ANTHROPIC_API_KEY="+w.Cfg.AnthropicAPIKey)
	}

	// Gemini (Vertex AI) settings. Pass through whatever is configured so the
	// gemini CLI can find the project/location and use Vertex with Google ADC.
	env = appendIfSet(env, "GOOGLE_GENAI_USE_VERTEXAI", w.Cfg.GoogleUseVertexAI)
	env = appendIfSet(env, "GOOGLE_CLOUD_PROJECT", w.Cfg.GoogleCloudProject)
	env = appendIfSet(env, "GOOGLE_CLOUD_LOCATION", w.Cfg.GoogleCloudLocation)
	// ADC location: entrypoint stages gcloud config into the node home; honor
	// CLOUDSDK_CONFIG / GOOGLE_APPLICATION_CREDENTIALS if the process has them.
	env = appendIfSet(env, "CLOUDSDK_CONFIG", os.Getenv("CLOUDSDK_CONFIG"))
	env = appendIfSet(env, "GOOGLE_APPLICATION_CREDENTIALS", os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	// Trust the headless temp workdir for the gemini CLI.
	env = append(env, "GEMINI_CLI_TRUST_WORKSPACE=true")
	return env
}

// appendIfSet appends KEY=val only when val is non-empty.
func appendIfSet(env []string, key, val string) []string {
	if val != "" {
		env = append(env, key+"="+val)
	}
	return env
}
