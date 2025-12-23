package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/outputters"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Lint only CLAUDE.md context files",
	Long: `The context command validates CLAUDE.md context files in your project.

Supported file patterns:
- .claude/CLAUDE.md
- CLAUDE.md

Validation checks:
- Rule compliance
- Section structure
- Reference validity`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runContextLint(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(contextCmd)
}

func runContextLint() error {
	// Load configuration
	config, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Run context linter
	summary, err := cli.LintContext(config.Root, config.Quiet, config.Verbose)
	if err != nil {
		return fmt.Errorf("error running context linter: %w", err)
	}

	// Format and output results
	outputter := outputters.NewOutputter(config)
	if err := outputter.Format(summary, config.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}