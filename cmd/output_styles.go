package cmd

import (
	"fmt"
	"os"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/spf13/cobra"
)

var outputStylesCmd = &cobra.Command{
	Use:   "output-styles",
	Short: "Lint only output style files",
	Long: `The output-styles command scans for and validates output style files in your project.

Supported file patterns:
- .claude/output-styles/**/*.md
- output-styles/**/*.md

Validation checks:
- Required frontmatter: name (kebab-case), description
- Optional frontmatter: keep-coding-instructions (boolean)
- Body content must not be empty`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runOutputStylesLint(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			exitFunc(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(outputStylesCmd)
}

func runOutputStylesLint() error {
	return runComponentLint("output-styles", cli.LintOutputStyles)
}
