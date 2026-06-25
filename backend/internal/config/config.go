// Package config loads runtime configuration from environment variables and
// the features.yaml file (see BUILD_SPEC §6).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Features mirrors config/features.yaml.
type Features struct {
	Learn struct {
		Enabled bool `yaml:"enabled"`
	} `yaml:"learn"`
	DataRoom struct {
		Enabled bool     `yaml:"enabled"`
		Accept  []string `yaml:"accept"`
	} `yaml:"data_room"`
	Advisors struct {
		UnlockAll bool `yaml:"unlock_all"`
	} `yaml:"advisors"`
}

// Config holds resolved env + feature configuration.
type Config struct {
	DatabaseURL     string
	JWTSecret       string
	AnthropicAPIKey string
	UploadDir       string
	SeedDir         string
	ConfigDir       string
	Port            string
	Features        Features

	// Agent backend selection and Gemini/Vertex settings.
	AgentProvider       string // "claude" (default) | "gemini"
	GoogleCloudProject  string
	GoogleCloudLocation string
	GoogleUseVertexAI   string // "true"/"false" pass-through for GOOGLE_GENAI_USE_VERTEXAI
}

// getenv returns the env value or a fallback default.
func getenv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// Load reads env vars and loads features.yaml from CONFIG_DIR.
func Load() (*Config, error) {
	c := &Config{
		DatabaseURL:     os.Getenv("DATABASE_URL"),
		JWTSecret:       os.Getenv("JWT_SECRET"),
		AnthropicAPIKey: os.Getenv("ANTHROPIC_API_KEY"),
		UploadDir:       getenv("UPLOAD_DIR", "/data/uploads"),
		SeedDir:         getenv("SEED_DIR", "/seed"),
		ConfigDir:       getenv("CONFIG_DIR", "/config"),
		Port:            getenv("PORT", "8080"),

		AgentProvider:       getenv("AGENT_PROVIDER", "claude"),
		GoogleCloudProject:  os.Getenv("GOOGLE_CLOUD_PROJECT"),
		GoogleCloudLocation: os.Getenv("GOOGLE_CLOUD_LOCATION"),
		GoogleUseVertexAI:   os.Getenv("GOOGLE_GENAI_USE_VERTEXAI"),
	}

	features, err := LoadFeatures(filepath.Join(c.ConfigDir, "features.yaml"))
	if err != nil {
		return nil, err
	}
	c.Features = features
	return c, nil
}

// LoadFeatures reads and parses a features.yaml file at the given path.
func LoadFeatures(path string) (Features, error) {
	var f Features
	data, err := os.ReadFile(path)
	if err != nil {
		return f, fmt.Errorf("config: read features.yaml at %s: %w", path, err)
	}
	if err := yaml.Unmarshal(data, &f); err != nil {
		return f, fmt.Errorf("config: parse features.yaml: %w", err)
	}
	return f, nil
}
