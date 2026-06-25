package agent

import (
	"context"
	"strings"
)

// Provider identifies which agent backend (CLI) to exec.
const (
	ProviderClaude = "claude"
	ProviderGemini = "gemini"
)

// ResolveProvider returns the provider to use given the AGENT_PROVIDER env value
// and an optional per-agent override (from the prompt md frontmatter). The
// override wins when set to a recognised value; otherwise envDefault is used;
// otherwise claude is the default.
func ResolveProvider(envDefault, override string) string {
	if p := normalizeProvider(override); p != "" {
		return p
	}
	if p := normalizeProvider(envDefault); p != "" {
		return p
	}
	return ProviderClaude
}

func normalizeProvider(v string) string {
	switch strings.ToLower(strings.TrimSpace(v)) {
	case ProviderClaude:
		return ProviderClaude
	case ProviderGemini:
		return ProviderGemini
	default:
		return ""
	}
}

// ExecWithProvider dispatches to the claude or gemini exec path. opts.Provider
// selects the backend; an empty value defaults to claude.
func ExecWithProvider(ctx context.Context, opts ExecOptions) (string, error) {
	switch normalizeProvider(opts.Provider) {
	case ProviderGemini:
		return ExecGemini(ctx, opts)
	default:
		return Exec(ctx, opts)
	}
}
