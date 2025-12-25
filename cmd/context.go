package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/dotcommander/cclint/internal/cli"
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
	return runComponentLint("context", cli.LintContext)
}