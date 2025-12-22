package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/carlrannaberg/cclint/internal/config"
	"github.com/carlrannaberg/cclint/internal/cli"
	"github.com/carlrannaberg/cclint/internal/outputters"
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
	// Load configuration
	config, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Run skills linter
	summary, err := cli.LintSkills(config.Root, config.Quiet, config.Verbose)
	if err != nil {
		return fmt.Errorf("error running skills linter: %w", err)
	}

	// Format and output results
	outputter := outputters.NewOutputter(config)
	if err := outputter.Format(summary, config.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}
