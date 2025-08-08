package cli

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/martinmose/agentlink/internal/config"
	"github.com/spf13/cobra"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Check environment and permissions",
	Long: `Check the environment for potential issues with agentlink.

Performs various sanity checks including:
- Operating system and version
- Symlink support
- Config directory permissions
- PATH and binary location`,
	RunE: runDoctor,
}

func init() {
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	fmt.Printf("Agentlink Doctor\n")
	fmt.Printf("================\n\n")

	hasIssues := false

	// Check OS
	fmt.Printf("Operating System: %s %s\n", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		fmt.Printf("⚠️  Windows support is best-effort only\n")
		hasIssues = true
	} else {
		fmt.Printf("✓ Supported platform\n")
	}
	fmt.Printf("\n")

	// Check symlink support
	fmt.Printf("Symlink Support:\n")
	if err := checkSymlinkSupport(); err != nil {
		fmt.Printf("✗ Symlinks not supported: %v\n", err)
		hasIssues = true
	} else {
		fmt.Printf("✓ Symlinks are supported\n")
	}
	fmt.Printf("\n")

	// Check binary location
	fmt.Printf("Binary Location:\n")
	if exePath, err := os.Executable(); err != nil {
		fmt.Printf("⚠️  Could not determine binary location: %v\n", err)
	} else {
		fmt.Printf("Binary: %s\n", exePath)
		if isInPath(exePath) {
			fmt.Printf("✓ Binary is in PATH\n")
		} else {
			fmt.Printf("⚠️  Binary is not in PATH\n")
		}
	}
	fmt.Printf("\n")

	// Check config directories
	fmt.Printf("Configuration:\n")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("✗ Cannot determine home directory: %v\n", err)
		hasIssues = true
	} else {
		fmt.Printf("Home directory: %s\n", homeDir)
		
		configDir := filepath.Join(homeDir, ".config", "agentlink")
		if err := checkDirectoryAccess(configDir, true); err != nil {
			fmt.Printf("✗ Config directory issue: %v\n", err)
			hasIssues = true
		} else {
			fmt.Printf("✓ Config directory accessible: %s\n", configDir)
		}
	}
	fmt.Printf("\n")

	// Check current directory project config
	fmt.Printf("Project Configuration:\n")
	if _, err := os.Stat(".git"); err == nil {
		fmt.Printf("✓ Git repository detected\n")
	} else {
		fmt.Printf("⚠️  No .git directory (not in a git repository)\n")
	}
	
	if _, err := os.Stat(".agentlink.yaml"); err == nil {
		fmt.Printf("✓ Project config found: .agentlink.yaml\n")
		
		// Try to load and validate it
		if cfg, err := config.LoadConfig(".agentlink.yaml"); err != nil {
			fmt.Printf("✗ Project config is invalid: %v\n", err)
			hasIssues = true
		} else {
			fmt.Printf("✓ Project config is valid\n")
			if verbose {
				fmt.Printf("  Source: %s\n", cfg.Source)
				fmt.Printf("  Links: %d configured\n", len(cfg.Links))
			}
		}
	} else {
		fmt.Printf("⚠️  No project config (.agentlink.yaml)\n")
	}
	fmt.Printf("\n")

	// Check global config
	fmt.Printf("Global Configuration:\n")
	globalConfig, _ := config.FindConfigPath()
	if _, err := os.Stat(globalConfig); err == nil {
		fmt.Printf("✓ Global config found: %s\n", globalConfig)
		
		// Try to load and validate it
		if cfg, err := config.LoadConfig(globalConfig); err != nil {
			fmt.Printf("✗ Global config is invalid: %v\n", err)
			hasIssues = true
		} else {
			fmt.Printf("✓ Global config is valid\n")
			if verbose {
				fmt.Printf("  Source: %s\n", cfg.Source)
				fmt.Printf("  Links: %d configured\n", len(cfg.Links))
			}
		}
	} else {
		fmt.Printf("⚠️  No global config found: %s\n", globalConfig)
		fmt.Printf("    (This is normal - will be created when needed)\n")
	}
	fmt.Printf("\n")

	// Summary
	if hasIssues {
		fmt.Printf("🔧 Some issues found. See messages above for details.\n")
		return fmt.Errorf("environment check found issues")
	} else {
		fmt.Printf("✅ Environment looks good!\n")
		return nil
	}
}

func checkSymlinkSupport() error {
	// Create a temporary file and symlink to test
	tmpDir := os.TempDir()
	testFile := filepath.Join(tmpDir, "agentlink_test_target")
	testLink := filepath.Join(tmpDir, "agentlink_test_link")

	// Clean up any existing test files
	os.Remove(testFile)
	os.Remove(testLink)

	// Create test target file
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("cannot create test file: %w", err)
	}
	defer os.Remove(testFile)

	// Try to create symlink
	if err := os.Symlink(testFile, testLink); err != nil {
		return fmt.Errorf("cannot create symlinks: %w", err)
	}
	defer os.Remove(testLink)

	// Verify symlink works
	if _, err := os.Readlink(testLink); err != nil {
		return fmt.Errorf("cannot read symlink: %w", err)
	}

	return nil
}

func isInPath(binaryPath string) bool {
	pathEnv := os.Getenv("PATH")
	if pathEnv == "" {
		return false
	}

	binaryDir := filepath.Dir(binaryPath)

	for _, dir := range filepath.SplitList(pathEnv) {
		if dir == binaryDir {
			return true
		}
	}

	// Also check if we can execute the command by name
	_, err := exec.LookPath("agentlink")
	return err == nil
}

func checkDirectoryAccess(dirPath string, createIfMissing bool) error {
	info, err := os.Stat(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			if createIfMissing {
				if err := os.MkdirAll(dirPath, 0755); err != nil {
					return fmt.Errorf("cannot create directory: %w", err)
				}
				return nil
			}
			return fmt.Errorf("directory does not exist")
		}
		return fmt.Errorf("cannot stat directory: %w", err)
	}

	if !info.IsDir() {
		return fmt.Errorf("path exists but is not a directory")
	}

	// Test write access by creating a temporary file
	testFile := filepath.Join(dirPath, ".agentlink_test")
	if err := os.WriteFile(testFile, []byte("test"), 0644); err != nil {
		return fmt.Errorf("directory is not writable: %w", err)
	}
	os.Remove(testFile)

	return nil
}