package cmd

import (
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/spf13/cobra"
)

var rulesCmd = &cobra.Command{
	Use:   "rules",
	Short: "Lint only rule files in .claude/rules/",
	Long: `The rules command scans for and validates rule files in your project.

Rule files provide modular, topic-specific instructions for Claude Code.
They support optional YAML frontmatter with a 'paths' field for conditional loading.

Patterns checked:
- .claude/rules/**/*.md

Validates:
- File extension (.md required)
- paths: glob pattern syntax
- @import references exist
- Symlink targets exist
- Non-empty content`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := runRulesLint(); err != nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(rulesCmd)
}

func runRulesLint() error {
	return runComponentLint("rules", lint.LintRules)
}
