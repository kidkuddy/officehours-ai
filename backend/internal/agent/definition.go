// Package agent handles agent definitions (markdown + frontmatter), prompt
// rendering for each job type, and the claude headless exec wrapper.
// See BUILD_SPEC §1 (claude exec) and §5 (agent md format).
package agent

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Definition is a parsed agent markdown file (frontmatter + body).
type Definition struct {
	Key         string `yaml:"key" json:"key"`
	Name        string `yaml:"name" json:"name"`
	Kind        string `yaml:"kind" json:"kind"` // advisor | concept | system
	Collection  string `yaml:"collection" json:"collection"`
	Enabled     bool   `yaml:"enabled" json:"enabled"`
	Order       int    `yaml:"order" json:"order"`
	Description string `yaml:"description" json:"description"`
	// Provider optionally overrides AGENT_PROVIDER for this agent
	// ("claude" | "gemini"). Empty means use the env default.
	Provider string `yaml:"provider" json:"provider,omitempty"`
	Body     string `yaml:"-" json:"-"` // system prompt body
}

// ParseFile reads and parses a single agent markdown file.
func ParseFile(path string) (*Definition, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return Parse(b)
}

// Parse parses markdown bytes with YAML frontmatter delimited by `---`.
func Parse(b []byte) (*Definition, error) {
	s := string(b)
	s = strings.ReplaceAll(s, "\r\n", "\n")
	if !strings.HasPrefix(s, "---") {
		return nil, fmt.Errorf("agent: missing frontmatter")
	}
	// Strip leading "---\n".
	rest := strings.TrimPrefix(s, "---")
	rest = strings.TrimPrefix(rest, "\n")
	idx := strings.Index(rest, "\n---")
	if idx < 0 {
		return nil, fmt.Errorf("agent: unterminated frontmatter")
	}
	front := rest[:idx]
	body := rest[idx+len("\n---"):]
	body = strings.TrimPrefix(body, "\n")

	var d Definition
	if err := yaml.Unmarshal([]byte(front), &d); err != nil {
		return nil, fmt.Errorf("agent: parse frontmatter: %w", err)
	}
	d.Body = strings.TrimSpace(body)
	return &d, nil
}

// LoadDir parses all *.md files in dir, returning enabled+disabled definitions
// sorted by Order then Name.
func LoadDir(dir string) ([]*Definition, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	var defs []*Definition
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
			continue
		}
		d, err := ParseFile(filepath.Join(dir, e.Name()))
		if err != nil {
			fmt.Fprintf(os.Stderr, "agent: warning: skip %s: %v\n", e.Name(), err)
			continue
		}
		defs = append(defs, d)
	}
	sort.SliceStable(defs, func(i, j int) bool {
		if defs[i].Order != defs[j].Order {
			return defs[i].Order < defs[j].Order
		}
		return defs[i].Name < defs[j].Name
	})
	return defs, nil
}

// LoadByKey finds an enabled definition by key within dir.
func LoadByKey(dir, key string) (*Definition, error) {
	defs, err := LoadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, d := range defs {
		if d.Key == key {
			return d, nil
		}
	}
	return nil, fmt.Errorf("agent: no definition with key %q in %s", key, dir)
}
