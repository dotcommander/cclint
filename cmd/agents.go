package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/carlrannaberg/cclint/internal/config"
	"github.com/carlrannaberg/cclint/internal/cli"
	"github.com/carlrannaberg/cclint/internal/outputters"
)

var agentsCmd = &cobra.Command{
	Use:   "agents",
	Short: "Lint only agent definition files",
	Long: `The agents command scans for and validates agent definition files in your project.

Supported file patterns:
- .claude/agents/**/*.md
- agents/**/*.md

Validation checks:
- Required fields: name, description
- Name format: lowercase alphanumeric with hyphens
- Valid model specification
- Valid color selection
- Tool permissions`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runAgentsLint(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)
}

func runAgentsLint() error {
	// Load configuration
	config, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Run agents linter
	summary, err := cli.LintAgents(config.Root, config.Quiet, config.Verbose)
	if err != nil {
		return fmt.Errorf("error running agents linter: %w", err)
	}

	// Format and output results
	outputter := outputters.NewOutputter(config)
	if err := outputter.Format(summary, config.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}