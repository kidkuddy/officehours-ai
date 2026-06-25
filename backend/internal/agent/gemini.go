package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// geminiJSON is the shape of `gemini -p ... -o json` output on success.
// The CLI emits {"session_id":"...","response":"<final text>","stats":{...}}.
// On failure it emits {"session_id":"...","error":{"type":...,"message":...}}.
type geminiJSON struct {
	SessionID string `json:"session_id"`
	Response  string `json:"response"`
	Error     *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}

// ExecGemini runs the Gemini CLI in headless mode and returns the final text.
//
//	gemini -p "<prompt>" -o json --yolo --skip-trust
//
// Vertex auth comes from the env (GOOGLE_GENAI_USE_VERTEXAI, GOOGLE_CLOUD_PROJECT,
// GOOGLE_CLOUD_LOCATION) plus Google ADC located via CLOUDSDK_CONFIG/
// GOOGLE_APPLICATION_CREDENTIALS, all supplied by the caller in opts.Env.
func ExecGemini(ctx context.Context, opts ExecOptions) (string, error) {
	if opts.WorkDir == "" {
		return "", fmt.Errorf("agent: WorkDir is required")
	}
	args := []string{
		"-p", opts.Prompt,
		"-o", "json",
		"--yolo",       // auto-approve all tool calls (so ohctl runs unattended)
		"--skip-trust", // trust the temp workdir in this headless run
		"--include-directories", opts.WorkDir,
	}
	cmd := exec.CommandContext(ctx, "gemini", args...)
	cmd.Dir = opts.WorkDir
	cmd.Env = append(os.Environ(), opts.Env...)

	out, err := cmd.Output()
	if err != nil {
		stderr := ""
		var ee *exec.ExitError
		if as(err, &ee) {
			stderr = strings.TrimSpace(string(ee.Stderr))
		}
		// The CLI may still print a JSON error body on stdout even on exit!=0.
		if text, perr := parseGeminiOutput(out); perr == nil && text != "" {
			return text, nil
		}
		if stderr != "" {
			return "", fmt.Errorf("agent: gemini exec: %w: %s", err, stderr)
		}
		return "", fmt.Errorf("agent: gemini exec: %w", err)
	}

	text, perr := parseGeminiOutput(out)
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

// parseGeminiOutput extracts the final text from gemini CLI output. It handles
// both the JSON object output mode and plain-text output mode.
func parseGeminiOutput(out []byte) (string, error) {
	trimmed := strings.TrimSpace(string(out))
	if trimmed == "" {
		return "", fmt.Errorf("agent: empty gemini output")
	}
	// The CLI prints some non-JSON banner lines (YOLO mode, ripgrep notes)
	// before the JSON object. Find the first '{' and parse from there.
	if i := strings.Index(trimmed, "{"); i >= 0 {
		candidate := trimmed[i:]
		var g geminiJSON
		if err := json.Unmarshal([]byte(candidate), &g); err == nil {
			if g.Error != nil && g.Error.Message != "" {
				return "", fmt.Errorf("agent: gemini returned error: %s", g.Error.Message)
			}
			if g.Response != "" {
				return g.Response, nil
			}
		}
	}
	// Plain text output mode: strip known banner lines and return the rest.
	var keep []string
	for _, line := range strings.Split(trimmed, "\n") {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		if strings.Contains(l, "YOLO mode is enabled") ||
			strings.Contains(l, "Ripgrep is not available") ||
			strings.HasPrefix(l, "Loaded cached") {
			continue
		}
		keep = append(keep, line)
	}
	text := strings.TrimSpace(strings.Join(keep, "\n"))
	if text == "" {
		return "", fmt.Errorf("agent: could not parse gemini output")
	}
	return text, nil
}
