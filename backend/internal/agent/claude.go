package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// ExecOptions configure a headless agent invocation.
type ExecOptions struct {
	Prompt   string
	WorkDir  string   // --add-dir target; also the process cwd
	Env      []string // extra env (DATABASE_URL, ANTHROPIC_API_KEY, PATH with ohctl)
	Provider string   // "claude" (default) | "gemini"; routes ExecWithProvider
}

// claudeJSON is the shape of `claude -p ... --output-format json` output.
// claude emits {"type":"result","result":"<final text>", ...}.
type claudeJSON struct {
	Type    string `json:"type"`
	Subtype string `json:"subtype"`
	Result  string `json:"result"`
	IsError bool   `json:"is_error"`
}

// Exec runs claude in headless mode and returns the assistant's final text.
//
//	claude -p "<prompt>" --output-format json --dangerously-skip-permissions --add-dir <workdir>
func Exec(ctx context.Context, opts ExecOptions) (string, error) {
	if opts.WorkDir == "" {
		return "", fmt.Errorf("agent: WorkDir is required")
	}
	args := []string{
		"-p", opts.Prompt,
		"--output-format", "json",
		"--dangerously-skip-permissions",
		"--add-dir", opts.WorkDir,
	}
	cmd := exec.CommandContext(ctx, "claude", args...)
	cmd.Dir = opts.WorkDir
	cmd.Env = append(os.Environ(), opts.Env...)

	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		var ee *exec.ExitError
		if as(err, &ee) {
			stderr = strings.TrimSpace(string(ee.Stderr))
		}
		if stderr != "" {
			return "", fmt.Errorf("agent: claude exec: %w: %s", err, stderr)
		}
		return "", fmt.Errorf("agent: claude exec: %w", err)
	}

	text, perr := parseClaudeOutput(out)
	if perr != nil {
		// Fall back to raw output if JSON parse fails but we got something.
		raw := strings.TrimSpace(string(out))
		if raw != "" {
			return raw, nil
		}
		return "", perr
	}
	return text, nil
}

// parseClaudeOutput extracts the final assistant text from claude JSON output.
func parseClaudeOutput(out []byte) (string, error) {
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return "", fmt.Errorf("agent: empty claude output")
	}
	// Single JSON object (non-streaming json format).
	var single claudeJSON
	if err := json.Unmarshal([]byte(trimmed), &single); err == nil && single.Result != "" {
		if single.IsError {
			return "", fmt.Errorf("agent: claude returned error result: %s", single.Result)
		}
		return single.Result, nil
	}
	// Streaming JSON (one object per line) — take the last result-typed line.
	lines := strings.Split(trimmed, "\n")
	for i := len(lines) - 1; i >= 0; i-- {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}
		var obj claudeJSON
		if err := json.Unmarshal([]byte(line), &obj); err == nil {
			if obj.Type == "result" && obj.Result != "" {
				return obj.Result, nil
			}
		}
	}
	return "", fmt.Errorf("agent: could not parse claude output")
}

// as is a tiny errors.As helper to avoid importing errors twice across files.
func as(err error, target any) bool {
	switch t := target.(type) {
	case **exec.ExitError:
		if ee, ok := err.(*exec.ExitError); ok {
			*t = ee
			return true
		}
	}
	return false
}
