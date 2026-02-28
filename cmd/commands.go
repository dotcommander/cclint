package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/dotcommander/cclint/internal/lint"
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
			exitFunc(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(commandsCmd)
}

func runCommandsLint() error {
	return runComponentLint("commands", lint.LintCommands)
}