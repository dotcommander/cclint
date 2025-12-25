package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/dotcommander/cclint/internal/cli"
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
	return runComponentLint("settings", cli.LintSettings)
}