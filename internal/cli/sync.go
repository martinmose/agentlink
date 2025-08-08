package cli

import (
	"fmt"
	"os"

	"github.com/martinmose/agentlink/internal/config"
	"github.com/martinmose/agentlink/internal/symlink"
	"github.com/spf13/cobra"
)

var syncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Create/fix symlinks based on configuration",
	Long: `Create or fix symlinks to keep instruction files in sync.

Reads .agentlink.yaml in current directory, or falls back to global config
at ~/.config/agentlink/config.yaml. Creates or fixes symlinks so they point
to the configured source file.`,
	RunE: runSync,
}

func init() {
	rootCmd.AddCommand(syncCmd)
}

func runSync(cmd *cobra.Command, args []string) error {
	// Find config file
	configPath, isProject := config.FindConfigPath()
	
	// Load or create config
	cfg, err := loadOrCreateConfig(configPath, isProject)
	if err != nil {
		return err
	}

	if verbose {
		if isProject {
			printInfo("Using project config: %s", configPath)
		} else {
			printInfo("Using global config: %s", configPath)
		}
	}

	// Create symlink manager
	manager := symlink.NewManager(dryRun, force, verbose)

	// Validate source file
	if err := manager.ValidateSource(cfg.Source); err != nil {
		printError("Source validation failed: %v", err)
		return err
	}

	printOK("Source: %s", cfg.Source)

	// Process each link
	hasErrors := false
	for _, linkPath := range cfg.Links {
		if err := processLink(manager, linkPath, cfg.Source); err != nil {
			printError("Failed to process %s: %v", linkPath, err)
			hasErrors = true
		}
	}

	if hasErrors {
		return fmt.Errorf("sync completed with errors")
	}

	if dryRun {
		printInfo("Dry run completed - no changes made")
	}

	return nil
}

func loadOrCreateConfig(configPath string, isProject bool) (*config.Config, error) {
	// Try to load existing config
	if _, err := os.Stat(configPath); err == nil {
		return config.LoadConfig(configPath)
	}

	// If it's a project config and doesn't exist, error
	if isProject {
		printError("No .agentlink.yaml found in current directory")
		printInfo("Run 'agentlink init' to create one")
		return nil, fmt.Errorf("no project config found")
	}

	// Create default global config
	printInfo("Creating default global config at %s", configPath)
	if !dryRun {
		if err := config.CreateDefaultGlobalConfig(configPath); err != nil {
			printError("Failed to create default config: %v", err)
			return nil, err
		}
	}

	printWarning("Please edit %s to configure your source and links", configPath)
	return nil, fmt.Errorf("created default config - please edit it first")
}

func processLink(manager *symlink.Manager, linkPath, sourcePath string) error {
	if verbose {
		printInfo("Processing link: %s", linkPath)
	}

	action, err := manager.FixLink(linkPath, sourcePath)
	if err != nil {
		return err
	}

	switch action {
	case "skip":
		if verbose {
			printSkip("%s already links to %s", linkPath, sourcePath)
		}
	case "create":
		printCreate("%s -> %s", linkPath, sourcePath)
	case "fix":
		printOK("Fixed %s -> %s", linkPath, sourcePath)
	case "replace":
		printOK("Replaced %s -> %s", linkPath, sourcePath)
	case "fix broken":
		printOK("Fixed broken %s -> %s", linkPath, sourcePath)
	}

	return nil
}