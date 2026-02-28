package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/git"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/outputters"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"golang.org/x/term"
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
	typeFlag         string // Force component type (--type flag)
	diffMode         bool     // Lint only changed files (--diff)
	stagedMode       bool     // Lint only staged files (--staged)
	noCycleCheck     bool     // Disable circular dependency detection
	useBaseline      bool     // Use baseline filtering
	createBaseline   bool     // Create/update baseline file
	baselinePath     string   // Custom baseline file path

	// exitFunc is the function called to exit the program.
	// It can be overridden in tests to prevent actual process termination.
	exitFunc = os.Exit
)

var rootCmd = &cobra.Command{
	Use:     "cclint [files|dirs...]",
	Short:   "Claude Code Lint - A comprehensive linting tool for Claude Code projects",
	Version: Version,
	Long: `CCLint is a linting tool for Claude Code projects that validates agent files,
command files, settings, and documentation according to established patterns.

USAGE MODES:

  Full scan (default):
    cclint                    Lint all component types
    cclint agents             Lint only agents
    cclint agents commands    Lint multiple types

  File and directory mode:
    cclint ./agents/foo.md    Lint a specific file
    cclint path/to/file.md    Lint by path
    cclint a.md b.md c.md     Lint multiple files
    cclint ./commands/        Lint all files in a directory
    cclint ./command/         Singular dir names auto-detected

  Git integration mode:
    cclint --staged           Lint only staged files (pre-commit)
    cclint --diff             Lint all uncommitted changes

  Baseline mode (gradual adoption):
    cclint --baseline-create  Create baseline from current issues
    cclint --baseline         Lint with baseline filtering

  Type override:
    cclint --type agent x.md  Override type detection

EXAMPLES:

  # Lint a single agent
  cclint ./agents/my-agent.md

  # Lint multiple files
  cclint ./agents/a.md ./commands/b.md

  # Lint an entire directory
  cclint ./commands/

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
		if diffMode || stagedMode {
			// Git integration mode
			if err := runGitLint(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				exitFunc(1)
			}
			return
		}

		classified, err := classifyArgs(args)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			exitFunc(2)
			return
		}

		switch {
		case len(classified.filePaths) > 0:
			// File/directory mode
			if err := runSingleFileLint(classified.filePaths); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				exitFunc(2)
			}
		case len(classified.typeFilters) > 0:
			// Type filter mode
			for _, ft := range classified.typeFilters {
				if err := runTypeLint(ft); err != nil {
					fmt.Fprintf(os.Stderr, "Error: %v\n", err)
					exitFunc(1)
				}
			}
		default:
			// Full scan mode
			if err := runLint(); err != nil {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
				exitFunc(1)
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		exitFunc(1)
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
	rootCmd.Flags().StringVarP(&typeFlag, "type", "t", "", "Force component type (agent|command|skill|settings|context|plugin|rule|output-style)")

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

// shouldFail checks if the lint run should exit with failure based on the --fail-on level.
func shouldFail(cfg *config.Config, errors, warnings, suggestions int) bool {
	switch cfg.FailOn {
	case "suggestion":
		if suggestions > 0 {
			return true
		}
		fallthrough
	case "warning":
		if warnings > 0 {
			return true
		}
		fallthrough
	default: // "error"
		return errors > 0
	}
}

func initConfig() {
	// Config loading is handled by config.LoadConfig — this hook only
	// registers environment variable support so viper flag bindings work
	// before LoadConfig is called.
	viper.SetEnvPrefix("CCLINT")
	viper.AutomaticEnv()
}

// startSpinner starts a braille spinner on stderr showing elapsed time.
// It returns a stop func that clears the line when called.
// If verbose, quiet, or stderr is not a TTY, returns a no-op stop func.
func startSpinner(cfg *config.Config) func() {
	if cfg.Verbose || cfg.Quiet || !term.IsTerminal(int(os.Stderr.Fd())) {
		return func() {}
	}

	frames := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	done := make(chan struct{})

	go func() {
		start := time.Now()
		tick := time.NewTicker(100 * time.Millisecond)
		defer tick.Stop()

		frame := 0
		for {
			select {
			case <-done:
				return
			case <-tick.C:
				elapsed := int(time.Since(start).Seconds())
				fmt.Fprintf(os.Stderr, "\r%s cclint %ds", frames[frame%len(frames)], elapsed)
				frame++
			}
		}
	}()

	return func() {
		close(done)
		// Clear the spinner line completely
		fmt.Fprintf(os.Stderr, "\r%-20s\r", "")
	}
}

func runLint() error {
	// Load configuration
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Create and run the orchestrator
	orchestrator := lint.NewOrchestrator(cfg, lint.OrchestratorConfig{
		RootPath:       rootPath,
		UseBaseline:    useBaseline,
		CreateBaseline: createBaseline,
		BaselinePath:   baselinePath,
	})

	stop := startSpinner(cfg)
	result, err := orchestrator.Run()
	stop()

	if err != nil {
		return err
	}

	// Output all results using the outputter
	outputter := outputters.NewOutputter(cfg)
	if err := outputter.FormatAll(result.Summaries, result.StartTime); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Print baseline filtering summary
	if result.BaselineIgnored > 0 && !cfg.Quiet {
		fmt.Fprintf(os.Stderr, "\n%d baseline issues ignored (%d errors, %d suggestions)\n",
			result.BaselineIgnored, result.ErrorsIgnored, result.SuggestionsIgnored)
	}

	// Print validation reminder (verbose only)
	if !cfg.Quiet && cfg.Verbose {
		fmt.Fprintln(os.Stderr, "\n  Validate suggestions against docs.anthropic.com or docs.claude.com")
	}

	// Exit with error based on --fail-on level
	if shouldFail(cfg, result.TotalErrors, 0, result.TotalSuggestions) {
		exitFunc(1)
	}

	return nil
}

// classifiedArgs holds the result of classifying command-line arguments.
type classifiedArgs struct {
	typeFilters []discovery.FileType
	filePaths   []string
}

// classifyArgs classifies each argument as either a type filter or a file/directory path.
//
// An arg is a type filter if discovery.ParseFileType succeeds AND os.Stat fails
// (i.e., it's a known type name that doesn't exist as a file/directory on disk).
// Everything else is treated as a file/directory path.
//
// Mixing type filters with file paths is an error.
func classifyArgs(args []string) (*classifiedArgs, error) {
	result := &classifiedArgs{}
	for _, arg := range args {
		ft, parseErr := discovery.ParseFileType(arg)
		_, statErr := os.Stat(arg)
		if parseErr == nil && statErr != nil {
			// Known type name, doesn't exist on disk → type filter
			result.typeFilters = append(result.typeFilters, ft)
		} else {
			// Everything else → file/directory path
			result.filePaths = append(result.filePaths, arg)
		}
	}
	if len(result.typeFilters) > 0 && len(result.filePaths) > 0 {
		return nil, fmt.Errorf("cannot mix type filters (%v) and file paths; use one or the other",
			result.typeFilters)
	}
	return result, nil
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
	summary, err := lint.LintFiles(files, rootPath, typeFlag, cfg.Quiet, cfg.Verbose)
	if err != nil {
		return err
	}

	// Output results
	outputter := outputters.NewOutputter(cfg)
	if err := outputter.Format(summary, cfg.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Print validation reminder (verbose only)
	if !cfg.Quiet && cfg.Verbose {
		fmt.Println("\n⚠️  Validate suggestions against docs.anthropic.com or docs.claude.com")
	}

	// Exit with error based on --fail-on level
	if shouldFail(cfg, summary.TotalErrors, summary.TotalWarnings, summary.TotalSuggestions) {
		exitFunc(1)
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
	summary, err := lint.LintFiles(files, gitRoot, "", cfg.Quiet, cfg.Verbose)
	if err != nil {
		return err
	}

	// Output results
	outputter := outputters.NewOutputter(cfg)
	if err := outputter.Format(summary, cfg.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Print validation reminder (verbose only)
	if !cfg.Quiet && cfg.Verbose {
		fmt.Println("\n⚠️  Validate suggestions against docs.anthropic.com or docs.claude.com")
	}

	// Exit with error based on --fail-on level
	if shouldFail(cfg, summary.TotalErrors, summary.TotalWarnings, summary.TotalSuggestions) {
		exitFunc(1)
	}

	return nil
}
