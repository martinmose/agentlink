package cli

import (
	"fmt"
	"os"

	"github.com/martinmose/agentlink/internal/config"
	"github.com/martinmose/agentlink/internal/symlink"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check status of symlinks",
	Long: `Check the status of each symlink defined in the configuration.

Reports the status of each link (OK, missing, wrong target, not a symlink, broken)
and exits with non-zero code if any problems are found.`,
	RunE: runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Find config file
	configPath, isProject := config.FindConfigPath()
	
	// Load config (don't create if missing)
	if _, err := os.Stat(configPath); err != nil {
		if isProject {
			printError("No .agentlink.yaml found in current directory")
			printInfo("Run 'agentlink init' to create one")
		} else {
			printError("No global config found at %s", configPath)
			printInfo("Run 'agentlink sync' to create a default config")
		}
		return fmt.Errorf("no config found")
	}

	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		printError("Failed to load config: %v", err)
		return err
	}

	if verbose {
		if isProject {
			printInfo("Checking project config: %s", configPath)
		} else {
			printInfo("Checking global config: %s", configPath)
		}
	}

	// Create symlink manager
	manager := symlink.NewManager(false, false, verbose)

	// Check each link
	hasProblems := false
	
	// Check source file
	sourceStatus := "OK"
	if err := manager.ValidateSource(cfg.Source); err != nil {
		sourceStatus = fmt.Sprintf("ERROR: %v", err)
		hasProblems = true
	}

	// Print header
	fmt.Printf("Source: %s [%s]\n", cfg.Source, sourceStatus)
	fmt.Printf("Links:\n")
	maxPathLen := 0
	
	// Calculate max path length for formatting
	for _, linkPath := range cfg.Links {
		if len(linkPath) > maxPathLen {
			maxPathLen = len(linkPath)
		}
	}

	for _, linkPath := range cfg.Links {
		info := manager.CheckLink(linkPath, cfg.Source)
		
		_ = info.Status.String() // We handle status display in the switch below

		if info.Status != symlink.StatusOK {
			hasProblems = true
		}

		// Format the output nicely
		fmt.Printf("  %-*s -> ", maxPathLen, linkPath)
		
		switch info.Status {
		case symlink.StatusOK:
			fmt.Printf("%s ✓\n", cfg.Source)
		case symlink.StatusMissing:
			fmt.Printf("missing\n")
		case symlink.StatusWrongTarget:
			fmt.Printf("%s (expected %s) ✗\n", info.Target, cfg.Source)
		case symlink.StatusNotSymlink:
			fmt.Printf("not a symlink ✗\n")
		case symlink.StatusBroken:
			if info.Error != nil {
				fmt.Printf("broken: %v ✗\n", info.Error)
			} else {
				fmt.Printf("broken ✗\n")
			}
		}
	}

	if hasProblems || sourceStatus != "OK" {
		fmt.Printf("\nFound problems. Run 'agentlink sync' to fix them.\n")
		return fmt.Errorf("configuration has problems")
	}

	fmt.Printf("\nAll links are correctly configured ✓\n")
	return nil
}