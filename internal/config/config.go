package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the agentlink configuration
type Config struct {
	Source string   `yaml:"source"`
	Links  []string `yaml:"links"`
}

// LoadConfig loads configuration from the given path
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", path, err)
	}

	// Validate config
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config in %s: %w", path, err)
	}

	// Expand paths
	if err := config.ExpandPaths(filepath.Dir(path)); err != nil {
		return nil, fmt.Errorf("failed to expand paths in %s: %w", path, err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Source == "" {
		return fmt.Errorf("source cannot be empty")
	}
	if len(c.Links) == 0 {
		return fmt.Errorf("links cannot be empty")
	}
	return nil
}

// ExpandPaths expands ~ and makes relative paths absolute based on configDir
func (c *Config) ExpandPaths(configDir string) error {
	var err error
	
	// Expand source path
	c.Source, err = expandPath(c.Source, configDir)
	if err != nil {
		return fmt.Errorf("failed to expand source path: %w", err)
	}

	// Expand link paths
	for i, link := range c.Links {
		c.Links[i], err = expandPath(link, configDir)
		if err != nil {
			return fmt.Errorf("failed to expand link path %s: %w", link, err)
		}
	}

	return nil
}

// expandPath expands ~ and makes relative paths absolute
func expandPath(path, baseDir string) (string, error) {
	// Handle ~ expansion
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", fmt.Errorf("failed to get home directory: %w", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Make relative paths absolute
	if !filepath.IsAbs(path) {
		path = filepath.Join(baseDir, path)
	}

	return filepath.Clean(path), nil
}

// FindConfigPath finds the appropriate config file path
// Returns project config (.agentlink.yaml) if exists, otherwise global config path
func FindConfigPath() (string, bool) {
	// Check for project config first
	projectConfig := ".agentlink.yaml"
	if _, err := os.Stat(projectConfig); err == nil {
		abs, _ := filepath.Abs(projectConfig)
		return abs, true
	}

	// Return global config path (may not exist yet)
	homeDir, _ := os.UserHomeDir()
	globalConfig := filepath.Join(homeDir, ".config", "agentlink", "config.yaml")
	return globalConfig, false
}

// CreateDefaultGlobalConfig creates a default global config with examples
func CreateDefaultGlobalConfig(path string) error {
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	defaultConfig := `# Agentlink global configuration
# This file was auto-created. Uncomment and modify as needed.

# Example: Use a file in your Claude config as source
# source: ~/.config/claude/CLAUDE.md
# links:
#   - ~/.config/opencode/AGENTS.md
#   - ~/.config/some-tool/INSTRUCTIONS.md

source: ~/.config/agentlink/INSTRUCTIONS.md
links:
  - ~/.config/agentlink/CLAUDE.md
  - ~/.config/agentlink/AGENTS.md
`

	if err := os.WriteFile(path, []byte(defaultConfig), 0644); err != nil {
		return fmt.Errorf("failed to write default config: %w", err)
	}

	return nil
}

// CreateProjectConfig creates a project config file
func CreateProjectConfig(path string) error {
	config := `# Choose the file you actually edit as the source:
source: CLAUDE.md
links:
  - AGENTS.md                    # Root level
  - OPENCODE.md                  # Root level
  # - .agent/AGENTS.md           # Inside .agent directory  
  # - .codex/instructions.md     # Different name and location
  # - config/ai/GEMINI.md        # Nested directories
`

	if err := os.WriteFile(path, []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write project config: %w", err)
	}

	return nil
}