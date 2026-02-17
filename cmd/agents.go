package cmd

import (
	"fmt"
	"os"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/spf13/cobra"
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
			exitFunc(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(agentsCmd)
}

func runAgentsLint() error {
	return runComponentLint("agents", cli.LintAgents)
}