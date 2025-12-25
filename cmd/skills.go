package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/dotcommander/cclint/internal/cli"
)

var skillsCmd = &cobra.Command{
	Use:   "skills",
	Short: "Lint only skill definition files",
	Long: `The skills command scans for and validates skill definition files in your project.

Supported file patterns:
- .claude/skills/**/SKILL.md
- skills/**/SKILL.md

Validation checks:
- SKILL.md file exists in skill directory
- Markdown structure is valid`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSkillsLint(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(skillsCmd)
}

func runSkillsLint() error {
	return runComponentLint("skills", cli.LintSkills)
}
