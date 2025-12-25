package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/outputters"
)

var (
	rootPath         string
	quiet            bool
	verbose          bool
	showScores       bool
	showImprovements bool
	outputFormat     string
	outputFile       string
	failOn           string
)

var rootCmd = &cobra.Command{
	Use:   "cclint",
	Short: "Claude Code Lint - A comprehensive linting tool for Claude Code projects",
	Long: `CCLint is a linting tool for Claude Code projects that validates agent files,
command files, settings, and documentation according to established patterns.

By default, cclint scans the entire project and reports on all validation issues.
Use specialized commands to focus on specific file types.

⚠️  NOTE: cclint is a work in progress. Its suggestions should be validated:
   • Cross-reference with official docs: docs.anthropic.com, docs.claude.com
   • Clear violations (fake flags, >220 lines agents) are reliable
   • Style suggestions should be verified against official documentation`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runLint(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&rootPath, "root", "r", "", "Project root directory (auto-detected if not specified)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&showScores, "scores", "s", false, "Show quality scores (0-100) for each component")
	rootCmd.PersistentFlags().BoolVarP(&showImprovements, "improvements", "i", false, "Show specific improvements with point values")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "console", "Output format for reports (console|json|markdown)")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file for reports (requires --format)")
	rootCmd.PersistentFlags().StringVarP(&failOn, "fail-on", "", "error", "Fail build on specified level (error|warning|suggestion)")

	_ = viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("showScores", rootCmd.PersistentFlags().Lookup("scores"))
	_ = viper.BindPFlag("showImprovements", rootCmd.PersistentFlags().Lookup("improvements"))
	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("fail-on", rootCmd.PersistentFlags().Lookup("fail-on"))
}

func initConfig() {
	configPaths := []string{".cclintrc.json", ".cclintrc.yaml", ".cclintrc.yml"}
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			viper.SetConfigFile(path)
			if err := viper.ReadInConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
				os.Exit(1)
			}
			break
		}
	}
}

func runLint() error {
	// Load configuration
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	outputter := outputters.NewOutputter(cfg)

	// Define all linters to run
	linters := []struct {
		name   string
		linter func(string, bool, bool) (*cli.LintSummary, error)
	}{
		{"agents", cli.LintAgents},
		{"commands", cli.LintCommands},
		{"skills", cli.LintSkills},
		{"settings", cli.LintSettings},
		{"context", cli.LintContext},
		{"plugins", cli.LintPlugins},
	}

	// Track totals across all linters
	var totalFiles, totalErrors, totalSuggestions int
	var hasErrors bool

	// Run all linters
	for _, l := range linters {
		summary, err := l.linter(cfg.Root, cfg.Quiet, cfg.Verbose)
		if err != nil {
			return fmt.Errorf("error running %s linter: %w", l.name, err)
		}

		// Skip empty results (no files of this type)
		if summary.TotalFiles == 0 {
			continue
		}

		// Format and output results for this component type
		if err := outputter.Format(summary, cfg.Format); err != nil {
			return fmt.Errorf("error formatting %s output: %w", l.name, err)
		}

		// Accumulate totals
		totalFiles += summary.TotalFiles
		totalErrors += summary.TotalErrors
		totalSuggestions += summary.TotalSuggestions
		if summary.TotalErrors > 0 {
			hasErrors = true
		}
	}

	// Print validation reminder (unless quiet mode)
	if !cfg.Quiet {
		fmt.Println("\n⚠️  Validate suggestions against docs.anthropic.com or docs.claude.com")
	}

	// Exit with error if any linter found errors
	if hasErrors {
		os.Exit(1)
	}

	return nil
}