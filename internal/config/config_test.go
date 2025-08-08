package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "config.yaml")
	
	configContent := `source: test.md
links:
  - link1.md
  - link2.md
`
	
	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Load the config
	cfg, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}
	
	// Check values
	expectedSource := filepath.Join(tmpDir, "test.md")
	if cfg.Source != expectedSource {
		t.Errorf("Expected source %s, got %s", expectedSource, cfg.Source)
	}
	
	if len(cfg.Links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(cfg.Links))
	}
	
	expectedLink1 := filepath.Join(tmpDir, "link1.md")
	if cfg.Links[0] != expectedLink1 {
		t.Errorf("Expected first link %s, got %s", expectedLink1, cfg.Links[0])
	}
}

func TestValidateConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: Config{
				Source: "test.md",
				Links:  []string{"link1.md", "link2.md"},
			},
			wantErr: false,
		},
		{
			name: "empty source",
			config: Config{
				Source: "",
				Links:  []string{"link1.md"},
			},
			wantErr: true,
		},
		{
			name: "empty links",
			config: Config{
				Source: "test.md",
				Links:  []string{},
			},
			wantErr: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExpandPaths(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}
	
	tmpDir := t.TempDir()
	
	config := Config{
		Source: "~/test.md",
		Links:  []string{"relative.md", "~/absolute.md"},
	}
	
	err = config.ExpandPaths(tmpDir)
	if err != nil {
		t.Fatalf("Failed to expand paths: %v", err)
	}
	
	expectedSource := filepath.Join(homeDir, "test.md")
	if config.Source != expectedSource {
		t.Errorf("Expected source %s, got %s", expectedSource, config.Source)
	}
	
	expectedRelative := filepath.Join(tmpDir, "relative.md")
	if config.Links[0] != expectedRelative {
		t.Errorf("Expected relative link %s, got %s", expectedRelative, config.Links[0])
	}
	
	expectedAbsolute := filepath.Join(homeDir, "absolute.md")
	if config.Links[1] != expectedAbsolute {
		t.Errorf("Expected absolute link %s, got %s", expectedAbsolute, config.Links[1])
	}
}

func TestFindConfigPath(t *testing.T) {
	// Save current directory
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer os.Chdir(origDir)
	
	// Test in directory without project config
	tmpDir := t.TempDir()
	os.Chdir(tmpDir)
	
	path, isProject := FindConfigPath()
	if isProject {
		t.Error("Expected global config, got project config")
	}
	
	homeDir, _ := os.UserHomeDir()
	expectedPath := filepath.Join(homeDir, ".config", "agentlink", "config.yaml")
	if path != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, path)
	}
	
	// Test in directory with project config
	projectConfig := ".agentlink.yaml"
	os.WriteFile(projectConfig, []byte("source: test.md\nlinks: [test.md]"), 0644)
	
	path, isProject = FindConfigPath()
	if !isProject {
		t.Error("Expected project config, got global config")
	}
	
	expectedProjectPath := filepath.Join(tmpDir, ".agentlink.yaml")
	if path != expectedProjectPath {
		t.Errorf("Expected project path %s, got %s", expectedProjectPath, path)
	}
}