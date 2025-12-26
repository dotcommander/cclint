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
	fileFlag         []string // Explicit file paths (--file flag)
	typeFlag         string   // Force component type (--type flag)
)

var rootCmd = &cobra.Command{
	Use:   "cclint [files...]",
	Short: "Claude Code Lint - A comprehensive linting tool for Claude Code projects",
	Long: `CCLint is a linting tool for Claude Code projects that validates agent files,
command files, settings, and documentation according to established patterns.

USAGE MODES:

  Full scan (default):
    cclint                    Lint all component types
    cclint agents             Lint only agents
    cclint commands           Lint only commands

  Single-file mode:
    cclint ./agents/foo.md    Lint a specific file
    cclint path/to/file.md    Lint by path
    cclint a.md b.md c.md     Lint multiple files

  Explicit file mode (for edge cases):
    cclint --file agents      Lint a file literally named "agents"
    cclint --type agent x.md  Override type detection

EXAMPLES:

  # Lint a single agent
  cclint ./agents/my-agent.md

  # Lint multiple files
  cclint ./agents/a.md ./commands/b.md

  # Force type for file outside standard path
  cclint --type skill ./custom/methodology.md

⚠️  NOTE: cclint is a work in progress. Its suggestions should be validated:
   • Cross-reference with official docs: docs.anthropic.com, docs.claude.com
   • Clear violations (fake flags, >220 lines agents) are reliable
   • Style suggestions should be verified against official documentation`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		// Determine mode: single-file vs full scan
		filesToLint := collectFilesToLint(args)

		if len(filesToLint) > 0 {
			// Single-file mode
			if err := runSingleFileLint(filesToLint); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(2) // Exit 2 for invocation errors
			}
		} else {
			// Full scan mode
			if err := runLint(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
			}
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

	// Existing flags
	rootCmd.PersistentFlags().StringVarP(&rootPath, "root", "r", "", "Project root directory (auto-detected if not specified)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().BoolVarP(&showScores, "scores", "s", false, "Show quality scores (0-100) for each component")
	rootCmd.PersistentFlags().BoolVarP(&showImprovements, "improvements", "i", false, "Show specific improvements with point values")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "console", "Output format for reports (console|json|markdown)")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file for reports (requires --format)")
	rootCmd.PersistentFlags().StringVarP(&failOn, "fail-on", "", "error", "Fail build on specified level (error|warning|suggestion)")

	// Single-file mode flags
	rootCmd.Flags().StringArrayVar(&fileFlag, "file", nil, "Explicit file path(s) to lint (use for files with subcommand names)")
	rootCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Force component type (agent|command|skill|settings|context|plugin)")

	// Viper bindings
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

// collectFilesToLint determines which files to lint based on args and flags.
//
// Priority:
//  1. --file flag (explicit files, bypasses subcommand detection)
//  2. Args that look like file paths
//  3. Empty (full scan mode)
//
// Known subcommands (agents, commands, etc.) are NOT treated as file paths
// unless --file flag is used.
func collectFilesToLint(args []string) []string {
	var files []string

	// 1. --file flag takes precedence (explicit file mode)
	if len(fileFlag) > 0 {
		return fileFlag
	}

	// 2. Check args for file paths
	for _, arg := range args {
		// Skip known subcommands (they'll be handled by Cobra)
		if cli.IsKnownSubcommand(arg) {
			continue
		}

		// Check if it looks like a file path
		if cli.LooksLikePath(arg) {
			files = append(files, arg)
		}
	}

	return files
}

// runSingleFileLint lints specific files and outputs results.
//
// Exit codes:
//   - 0: All files passed (no errors)
//   - 1: One or more files had lint errors
//   - 2: Invocation error (file not found, invalid type, etc.)
func runSingleFileLint(files []string) error {
	// Load configuration for output settings
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Override config with flag values
	cfg.Quiet = quiet
	cfg.Verbose = verbose

	// Lint files
	summary, err := cli.LintFiles(files, rootPath, typeFlag, cfg.Quiet, cfg.Verbose)
	if err != nil {
		return err
	}

	// Output results
	outputter := outputters.NewOutputter(cfg)
	if err := outputter.Format(summary, cfg.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Print validation reminder (unless quiet mode)
	if !cfg.Quiet {
		fmt.Println("\n⚠️  Validate suggestions against docs.anthropic.com or docs.claude.com")
	}

	// Exit with error if any file had errors
	if summary.TotalErrors > 0 {
		os.Exit(1)
	}

	return nil
}