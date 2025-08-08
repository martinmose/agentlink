package cli

import (
	"fmt"
	"os"

	"github.com/martinmose/agentlink/internal/config"
	"github.com/martinmose/agentlink/internal/symlink"
	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Remove managed symlinks",
	Long: `Remove symlinks that are managed by agentlink.

Only removes symlinks that point to the configured source file.
Never removes the source file itself or regular files.`,
	RunE: runClean,
}

func init() {
	rootCmd.AddCommand(cleanCmd)
}

func runClean(cmd *cobra.Command, args []string) error {
	// Find config file
	configPath, isProject := config.FindConfigPath()
	
	// Load config (don't create if missing)
	if _, err := os.Stat(configPath); err != nil {
		if isProject {
			printError("No .agentlink.yaml found in current directory")
			printInfo("Run 'agentlink init' to create one")
		} else {
			printError("No global config found at %s", configPath)
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
			printInfo("Cleaning project links from: %s", configPath)
		} else {
			printInfo("Cleaning global links from: %s", configPath)
		}
	}

	// Create symlink manager
	manager := symlink.NewManager(dryRun, force, verbose)

	printInfo("Source: %s (will NOT be removed)", cfg.Source)

	// Process each link
	removedCount := 0
	skippedCount := 0

	for _, linkPath := range cfg.Links {
		if verbose {
			printInfo("Processing link: %s", linkPath)
		}

		info := manager.CheckLink(linkPath, cfg.Source)
		
		switch info.Status {
		case symlink.StatusOK:
			// This is a symlink pointing to our source - remove it
			if !dryRun {
				if err := manager.RemoveLink(linkPath, cfg.Source); err != nil {
					printError("Failed to remove %s: %v", linkPath, err)
					continue
				}
			}
			printOK("Removed %s", linkPath)
			removedCount++
			
		case symlink.StatusMissing:
			if verbose {
				printSkip("%s (already missing)", linkPath)
			}
			skippedCount++
			
		case symlink.StatusWrongTarget:
			printWarning("Skipped %s (points to %s, not %s)", linkPath, info.Target, cfg.Source)
			skippedCount++
			
		case symlink.StatusNotSymlink:
			printWarning("Skipped %s (not a symlink)", linkPath)
			skippedCount++
			
		case symlink.StatusBroken:
			// Remove broken symlinks
			if !dryRun {
				if err := os.Remove(linkPath); err != nil {
					printError("Failed to remove broken symlink %s: %v", linkPath, err)
					continue
				}
			}
			printOK("Removed broken symlink %s", linkPath)
			removedCount++
		}
	}

	// Summary
	if dryRun {
		printInfo("Dry run completed - would remove %d symlinks, skip %d items", removedCount, skippedCount)
	} else {
		printInfo("Clean completed - removed %d symlinks, skipped %d items", removedCount, skippedCount)
	}

	return nil
}