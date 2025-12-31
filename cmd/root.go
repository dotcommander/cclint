package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/git"
	"github.com/dotcommander/cclint/internal/outputters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X github.com/dotcommander/cclint/cmd.Version=1.0.0"
var Version = "dev"

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
	diffMode         bool     // Lint only changed files (--diff)
	stagedMode       bool     // Lint only staged files (--staged)
	noCycleCheck     bool     // Disable circular dependency detection
	useBaseline      bool     // Use baseline filtering
	createBaseline   bool     // Create/update baseline file
	baselinePath     string   // Custom baseline file path
)

var rootCmd = &cobra.Command{
	Use:     "cclint [files...]",
	Short:   "Claude Code Lint - A comprehensive linting tool for Claude Code projects",
	Version: Version,
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

  Git integration mode:
    cclint --staged           Lint only staged files (pre-commit)
    cclint --diff             Lint all uncommitted changes

  Baseline mode (gradual adoption):
    cclint --baseline-create  Create baseline from current issues
    cclint --baseline         Lint with baseline filtering

  Explicit file mode (for edge cases):
    cclint --file agents      Lint a file literally named "agents"
    cclint --type agent x.md  Override type detection

EXAMPLES:

  # Lint a single agent
  cclint ./agents/my-agent.md

  # Lint multiple files
  cclint ./agents/a.md ./commands/b.md

  # Lint only staged files (pre-commit hook)
  cclint --staged

  # Lint all uncommitted changes
  cclint --diff

  # Create baseline to accept current state
  cclint --baseline-create

  # Lint with baseline (only new issues fail)
  cclint --baseline

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
		} else if diffMode || stagedMode {
			// Git integration mode
			if err := runGitLint(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				os.Exit(1)
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

	// Use -V for version since -v is already used for verbose
	rootCmd.Flags().BoolP("version", "V", false, "Print version information")

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
	rootCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Force component type (agent|command|skill|settings|context|plugin|rule)")

	// Git integration flags
	rootCmd.Flags().BoolVar(&diffMode, "diff", false, "Lint only uncommitted changes (staged + unstaged)")
	rootCmd.Flags().BoolVar(&stagedMode, "staged", false, "Lint only staged files (for pre-commit hooks)")

	// Analysis flags
	rootCmd.PersistentFlags().BoolVar(&noCycleCheck, "no-cycle-check", false, "Disable circular dependency detection")

	// Baseline flags
	rootCmd.PersistentFlags().BoolVar(&useBaseline, "baseline", false, "Use .cclintbaseline.json to filter known issues")
	rootCmd.PersistentFlags().BoolVar(&createBaseline, "baseline-create", false, "Create/update baseline file from current issues")
	rootCmd.PersistentFlags().StringVar(&baselinePath, "baseline-path", ".cclintbaseline.json", "Path to baseline file")

	// Viper bindings
	_ = viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))
	_ = viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	_ = viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	_ = viper.BindPFlag("showScores", rootCmd.PersistentFlags().Lookup("scores"))
	_ = viper.BindPFlag("showImprovements", rootCmd.PersistentFlags().Lookup("improvements"))
	_ = viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	_ = viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	_ = viper.BindPFlag("fail-on", rootCmd.PersistentFlags().Lookup("fail-on"))
	_ = viper.BindPFlag("no-cycle-check", rootCmd.PersistentFlags().Lookup("no-cycle-check"))
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

	// Determine baseline path (relative to project root)
	baselineFile := baselinePath
	if !filepath.IsAbs(baselineFile) {
		baselineFile = filepath.Join(cfg.Root, baselineFile)
	}

	// Load baseline if requested
	var b *baseline.Baseline
	if useBaseline || createBaseline {
		if _, err := os.Stat(baselineFile); err == nil {
			b, err = baseline.LoadBaseline(baselineFile)
			if err != nil && !cfg.Quiet {
				fmt.Fprintf(os.Stderr, "Warning: Failed to load baseline: %v\n", err)
				b = nil
			}
		}
	}

	outputter := outputters.NewOutputter(cfg)

	// Define all linters to run
	linters := []struct {
		name   string
		linter func(string, bool, bool, bool) (*cli.LintSummary, error)
	}{
		{"agents", cli.LintAgents},
		{"commands", cli.LintCommands},
		{"skills", cli.LintSkills},
		{"settings", cli.LintSettings},
		{"context", cli.LintContext},
		{"rules", cli.LintRules},
		// {"plugins", cli.LintPlugins}, // TODO: re-enable when output is less overwhelming
	}

	// Track totals across all linters
	var totalFiles, totalErrors, totalSuggestions int
	var hasErrors bool
	var totalIgnored, errorsIgnored, suggestionsIgnored int
	var allIssues []cue.ValidationError // For baseline creation

	// Run all linters
	for _, l := range linters {
		summary, err := l.linter(cfg.Root, cfg.Quiet, cfg.Verbose, cfg.NoCycleCheck)
		if err != nil {
			return fmt.Errorf("error running %s linter: %w", l.name, err)
		}

		// Skip empty results (no files of this type)
		if summary.TotalFiles == 0 {
			continue
		}

		// Collect issues for baseline creation
		if createBaseline {
			allIssues = append(allIssues, cli.CollectAllIssues(summary)...)
		}

		// Filter with baseline if active
		if useBaseline && b != nil {
			ignored, errIgnored, suggIgnored := cli.FilterResults(summary, b)
			totalIgnored += ignored
			errorsIgnored += errIgnored
			suggestionsIgnored += suggIgnored
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

	// Run project-wide memory checks
	if !cfg.Quiet {
		// Check CLAUDE.local.md gitignore
		gitignoreWarnings := cli.CheckClaudeLocalGitignore(cfg.Root)
		for _, w := range gitignoreWarnings {
			fmt.Printf("warning: %s: %s\n", w.File, w.Message)
		}

		// Check combined memory size
		fd := discovery.NewFileDiscovery(cfg.Root, false)
		allFiles, _ := fd.DiscoverFiles()
		sizeWarnings := cli.CheckCombinedMemorySize(cfg.Root, allFiles)
		for _, w := range sizeWarnings {
			fmt.Printf("warning: %s\n", w.Message)
		}
	}

	// Create/update baseline if requested (do this BEFORE exiting on errors)
	if createBaseline {
		b = baseline.CreateBaseline(allIssues)
		b.CreatedAt = time.Now().UTC().Format(time.RFC3339)

		if err := b.SaveBaseline(baselineFile); err != nil {
			return fmt.Errorf("failed to save baseline: %w", err)
		}

		if !cfg.Quiet {
			fmt.Printf("\nBaseline created: %s (%d issues)\n", baselineFile, len(b.Fingerprints))
		}

		// When creating baseline, exit 0 (success) to accept current state
		return nil
	}

	// Print baseline filtering summary
	if useBaseline && b != nil && totalIgnored > 0 && !cfg.Quiet {
		fmt.Printf("\n%d baseline issues ignored (%d errors, %d suggestions)\n",
			totalIgnored, errorsIgnored, suggestionsIgnored)
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

// runGitLint lints files based on git status (--diff or --staged)
func runGitLint() error {
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Override config with flag values
	cfg.Quiet = quiet
	cfg.Verbose = verbose

	// Determine git root (use current directory if rootPath not specified)
	gitRoot := cfg.Root
	if rootPath == "" {
		// Use current working directory for git operations
		gitRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	// Check if in git repository
	if !git.IsGitRepo(gitRoot) {
		if !cfg.Quiet {
			fmt.Fprintf(os.Stderr, "Warning: Not in a git repository. Falling back to full lint.\n\n")
		}
		return runLint()
	}

	// Get files from git
	var files []string
	if stagedMode {
		files, err = git.GetStagedFiles(gitRoot)
	} else if diffMode {
		files, err = git.GetChangedFiles(gitRoot)
	}
	if err != nil {
		return fmt.Errorf("error getting git files: %w", err)
	}

	if len(files) == 0 {
		if !cfg.Quiet {
			fmt.Println("No files to lint")
		}
		return nil
	}

	// Lint files (pass gitRoot as the root path for correct detection)
	summary, err := cli.LintFiles(files, gitRoot, "", cfg.Quiet, cfg.Verbose)
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
