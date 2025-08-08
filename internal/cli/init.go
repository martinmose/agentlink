package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/martinmose/agentlink/internal/config"
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Create .agentlink.yaml in current directory",
	Long: `Create a .agentlink.yaml configuration file in the current directory.

If no .git directory is found, you'll be prompted to confirm creation.`,
	RunE: runInit,
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command, args []string) error {
	configPath := ".agentlink.yaml"

	// Check if config already exists
	if _, err := os.Stat(configPath); err == nil {
		if !force {
			printError(".agentlink.yaml already exists (use --force to overwrite)")
			return fmt.Errorf("config file already exists")
		}
		printWarning("Overwriting existing .agentlink.yaml")
	}

	// Check for .git directory
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		if !force {
			fmt.Print("No .git directory found. Create .agentlink.yaml here anyway? (y/N): ")
			reader := bufio.NewReader(os.Stdin)
			response, err := reader.ReadString('\n')
			if err != nil {
				return fmt.Errorf("failed to read input: %w", err)
			}
			
			response = strings.ToLower(strings.TrimSpace(response))
			if response != "y" && response != "yes" {
				printInfo("Cancelled")
				return nil
			}
		} else {
			printWarning("No .git directory found, but continuing due to --force")
		}
	}

	// Create the config file
	if dryRun {
		printInfo("Would create .agentlink.yaml")
		return nil
	}

	if err := config.CreateProjectConfig(configPath); err != nil {
		printError("Failed to create config file: %v", err)
		return err
	}

	abs, _ := filepath.Abs(configPath)
	printOK("Created %s", abs)
	printInfo("Edit the config file and run 'agentlink sync' after creating your source file")

	return nil
}