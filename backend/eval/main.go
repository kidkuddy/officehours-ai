// Command eval runs the diagnoser against a labeled set of founder profiles and
// reports maturity-stage classification accuracy (BUILD_SPEC §9).
//
// Usage:
//
//	DATABASE_URL=... eval \
//	  -set backend/eval/labeled_set.json \
//	  -seed-dir seed \
//	  [-ohctl-dir /path/to/ohctl] [-keep]
//
// It connects directly to Postgres, creates an ephemeral user per case, execs
// the diagnoser (claude headless with ohctl on PATH), reads back the persisted
// stage, and prints a JSON report. Ephemeral users are deleted unless -keep.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"officehours/internal/agent"
	"officehours/internal/auth"

	"github.com/jackc/pgx/v5/pgxpool"
)

type labeledCase struct {
	ID            string `json:"id"`
	ExpectedStage string `json:"expected_stage"`
	Text          string `json:"text"`
}

type caseResult struct {
	ID             string `json:"id"`
	ExpectedStage  string `json:"expected_stage"`
	PredictedStage string `json:"predicted_stage"`
	Correct        bool   `json:"correct"`
	Error          string `json:"error,omitempty"`
}

func main() {
	var (
		setPath  = flag.String("set", "labeled_set.json", "path to labeled set JSON")
		seedDir  = flag.String("seed-dir", "seed", "seed directory (contains agents/diagnoser.md)")
		ohctlDir = flag.String("ohctl-dir", "", "directory containing the ohctl binary (prepended to PATH)")
		keep     = flag.Bool("keep", false, "keep ephemeral users instead of deleting")
		out      = flag.String("out", "", "optional path to write the JSON report")
	)
	flag.Parse()

	if err := run(*setPath, *seedDir, *ohctlDir, *keep, *out); err != nil {
		fmt.Fprintf(os.Stderr, "eval: %v\n", err)
		os.Exit(1)
	}
}

func run(setPath, seedDir, ohctlDir string, keep bool, outPath string) error {
	raw, err := os.ReadFile(setPath)
	if err != nil {
		return fmt.Errorf("read set: %w", err)
	}
	var cases []labeledCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		return fmt.Errorf("parse set: %w", err)
	}

	def, err := agent.LoadByKey(filepath.Join(seedDir, "agents"), "diagnoser")
	if err != nil {
		return fmt.Errorf("load diagnoser: %w (is %s/agents/diagnoser.md present?)", err, seedDir)
	}

	ctx := context.Background()
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		return fmt.Errorf("DATABASE_URL is not set")
	}
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return fmt.Errorf("connect db: %w", err)
	}
	defer pool.Close()

	results := make([]caseResult, 0, len(cases))
	correct := 0
	for _, c := range cases {
		res := caseResult{ID: c.ID, ExpectedStage: c.ExpectedStage}

		userID, cleanup, err := ephemeralUser(ctx, pool, c)
		if err != nil {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}

		prompt := agent.RenderDiagnoser(def, agent.PromptContext{
			UserID:         userID,
			OnboardingText: c.Text,
		})
		if err := execDiagnoser(ctx, prompt, url, seedDir, ohctlDir); err != nil {
			res.Error = err.Error()
		}

		// Read back the persisted stage.
		var stage string
		if qerr := pool.QueryRow(ctx,
			`select stage from profiles where user_id=$1`, userID).Scan(&stage); qerr == nil {
			res.PredictedStage = stage
			res.Correct = stage == c.ExpectedStage
			if res.Correct {
				correct++
			}
		} else if res.Error == "" {
			res.Error = fmt.Sprintf("read stage: %v", qerr)
		}

		if !keep {
			cleanup()
		}
		results = append(results, res)
		fmt.Fprintf(os.Stderr, "eval: %s expected=%s predicted=%s correct=%v\n",
			c.ID, c.ExpectedStage, res.PredictedStage, res.Correct)
	}

	report := map[string]any{
		"total":    len(cases),
		"correct":  correct,
		"accuracy": ratio(correct, len(cases)),
		"results":  results,
	}
	b, _ := json.MarshalIndent(report, "", "  ")
	fmt.Println(string(b))
	if outPath != "" {
		if err := os.WriteFile(outPath, b, 0o644); err != nil {
			return fmt.Errorf("write report: %w", err)
		}
	}
	return nil
}

// ephemeralUser creates a throwaway user + profile (with company_text) and
// returns its id and a cleanup func.
func ephemeralUser(ctx context.Context, pool *pgxpool.Pool, c labeledCase) (string, func(), error) {
	hash, err := auth.HashPassword("eval-" + c.ID)
	if err != nil {
		return "", nil, err
	}
	email := fmt.Sprintf("eval+%s+%d@officehours.local", c.ID, time.Now().UnixNano())
	var id string
	err = pool.QueryRow(ctx,
		`insert into users (email, password_hash, name) values ($1,$2,$3) returning id`,
		email, hash, "Eval "+c.ID).Scan(&id)
	if err != nil {
		return "", nil, fmt.Errorf("create user: %w", err)
	}
	_, err = pool.Exec(ctx,
		`insert into profiles (user_id, company_text) values ($1,$2)
		 on conflict (user_id) do update set company_text=excluded.company_text`,
		id, c.Text)
	if err != nil {
		return "", nil, fmt.Errorf("create profile: %w", err)
	}
	cleanup := func() {
		_, _ = pool.Exec(ctx, `delete from users where id=$1`, id)
	}
	return id, cleanup, nil
}

func execDiagnoser(ctx context.Context, prompt, dbURL, seedDir, ohctlDir string) error {
	workdir, err := os.MkdirTemp("", "oh-eval-")
	if err != nil {
		return err
	}
	defer os.RemoveAll(workdir)

	path := os.Getenv("PATH")
	if ohctlDir != "" {
		path = ohctlDir + string(os.PathListSeparator) + path
	}
	env := []string{
		"PATH=" + path,
		"DATABASE_URL=" + dbURL,
		"SEED_DIR=" + seedDir,
	}
	if k := os.Getenv("ANTHROPIC_API_KEY"); k != "" {
		env = append(env, "ANTHROPIC_API_KEY="+k)
	}

	execCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	_, err = agent.Exec(execCtx, agent.ExecOptions{Prompt: prompt, WorkDir: workdir, Env: env})
	return err
}

func ratio(a, b int) float64 {
	if b == 0 {
		return 0
	}
	return float64(a) / float64(b)
}
