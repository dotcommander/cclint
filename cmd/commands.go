package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/outputters"
)

var commandsCmd = &cobra.Command{
	Use:   "commands",
	Short: "Lint only command definition files",
	Long: `The commands command scans for and validates command definition files in your project.

Supported file patterns:
- .claude/commands/**/*.md
- commands/**/*.md

Validation checks:
- Required fields: name, description
- Valid model specification
- Tool permissions through allowed-tools`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runCommandsLint(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(commandsCmd)
}

func runCommandsLint() error {
	// Load configuration
	config, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Run commands linter
	summary, err := cli.LintCommands(config.Root, config.Quiet, config.Verbose)
	if err != nil {
		return fmt.Errorf("error running commands linter: %w", err)
	}

	// Format and output results
	outputter := outputters.NewOutputter(config)
	if err := outputter.Format(summary, config.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}