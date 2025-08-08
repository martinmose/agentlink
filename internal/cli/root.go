package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	// Build-time variables
	version = "dev"
	commit  = "none"
	date    = "unknown"
	
	// Command flags
	dryRun  bool
	force   bool
	verbose bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "agentlink",
	Version: version,
	Short:   "Keep your AI instruction files in sync with zero magic — just symlinks",
	Long: `Agentlink keeps AI instruction files in sync across different tools using symlinks.

Different tools want different files at project root: AGENTS.md, CLAUDE.md, GEMINI.md, etc.
Agentlink solves this by maintaining one source file and creating symlinks to it.`,
}

// Execute adds all child commands to the root command and sets flags appropriately.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&dryRun, "dry-run", false, "show what would be done without making changes")
	rootCmd.PersistentFlags().BoolVarP(&force, "force", "f", false, "force replacement of conflicting files")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

// printInfo prints an info message
func printInfo(format string, args ...interface{}) {
	fmt.Printf("[info] "+format+"\n", args...)
}

// printOK prints a success message
func printOK(format string, args ...interface{}) {
	fmt.Printf("[ok] "+format+"\n", args...)
}

// printCreate prints a create message
func printCreate(format string, args ...interface{}) {
	fmt.Printf("[create] "+format+"\n", args...)
}

// printSkip prints a skip message
func printSkip(format string, args ...interface{}) {
	fmt.Printf("[skip] "+format+"\n", args...)
}

// printError prints an error message
func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[error] "+format+"\n", args...)
}

// printWarning prints a warning message
func printWarning(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "[warning] "+format+"\n", args...)
}