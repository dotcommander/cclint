package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/carlrannaberg/cclint/internal/config"
	"github.com/carlrannaberg/cclint/internal/cli"
	"github.com/carlrannaberg/cclint/internal/outputters"
)

var settingsCmd = &cobra.Command{
	Use:   "settings",
	Short: "Lint only settings.json files",
	Long: `The settings command validates Claude settings.json files in your project.

Supported file patterns:
- .claude/settings.json
- claude/settings.json

Validation checks:
- Hook configurations
- Event types
- Command specifications`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSettingsLint(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(settingsCmd)
}

func runSettingsLint() error {
	// Load configuration
	config, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Run settings linter
	summary, err := cli.LintSettings(config.Root, config.Quiet, config.Verbose)
	if err != nil {
		return fmt.Errorf("error running settings linter: %w", err)
	}

	// Format and output results
	outputter := outputters.NewOutputter(config)
	if err := outputter.Format(summary, config.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}