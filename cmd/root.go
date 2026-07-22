package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/git"
	"github.com/dotcommander/cclint/internal/lint"
	"golang.org/x/term"
)

// Version is set at build time via ldflags:
//
//	go build -ldflags "-X github.com/dotcommander/cclint/cmd.Version=1.0.0"
var Version = "dev"

// exitFunc is the function called to exit the program.
// It can be overridden in tests to prevent actual process termination.
var exitFunc = os.Exit

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

func runLint(opts executionOptions) error {
	cfg, err := loadCLIConfig(opts)
	if err != nil {
		return err
	}

	result, err := runOrchestratedLint(cfg, opts, nil)
	if err != nil {
		return err
	}

	if err := formatFullRunOutput(cfg, result); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	printBaselineSummary(result.BaselineIgnored, result.ErrorsIgnored, result.SuggestionsIgnored, cfg.Quiet)
	printValidationReminder(cfg)
	applyFailurePolicy(cfg, opts, result.TotalErrors, result.TotalWarnings, result.TotalSuggestions)

	return nil
}

func runRootCommand(opts executionOptions, cmd *lintCommand) error {
	if boolValue(cmd.Diff) || boolValue(cmd.Staged) {
		return runGitLint(opts, cmd)
	}

	classified, err := classifyArgs(cmd.Paths)
	if err != nil {
		return err
	}

	switch {
	case len(classified.filePaths) > 0:
		return runSingleFileLint(opts, cmd, classified.filePaths)
	case len(classified.typeFilters) > 0:
		for _, ft := range classified.typeFilters {
			if err := runTypeLint(opts, ft); err != nil {
				return err
			}
		}
		return nil
	default:
		return runLint(opts)
	}
}

// classifiedArgs holds the result of classifying command-line arguments.
type classifiedArgs struct {
	typeFilters []discovery.FileType
	filePaths   []string
}

// classifyArgs classifies each argument as either a type filter or a file/directory path.
//
// An arg is a type filter if discovery.ParseFileType succeeds (recognized type name).
// Type names always win over directory names; use ./dir/ to force directory mode.
// Everything else is treated as a file/directory path.
//
// Mixing type filters with file paths is an error.
func classifyArgs(args []string) (*classifiedArgs, error) {
	result := &classifiedArgs{}
	for _, arg := range args {
		ft, parseErr := discovery.ParseFileType(arg)
		if parseErr == nil {
			// Known type name → type filter (always wins over directory match)
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
func runSingleFileLint(opts executionOptions, cmd *lintCommand, files []string) error {
	cfg, err := loadCLIConfig(opts)
	if err != nil {
		return err
	}

	summary, err := lint.LintFiles(files, stringValue(opts.root), stringValue(cmd.Type), cfg.Quiet, cfg.Verbose)
	if err != nil {
		return err
	}

	if err := formatSummaryOutput(cfg, summary); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	printValidationReminder(cfg)
	applyFailurePolicy(cfg, opts, summary.TotalErrors, summary.TotalWarnings, summary.TotalSuggestions)

	return nil
}

// runGitLint lints files based on git status (--diff or --staged)
func runGitLint(opts executionOptions, cmd *lintCommand) error {
	cfg, err := loadCLIConfig(opts)
	if err != nil {
		return err
	}

	// Determine git root (use current directory if rootPath not specified)
	gitRoot := cfg.Root
	if opts.root == nil {
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
		return runLint(opts)
	}

	// Get files from git
	var files []string
	if boolValue(cmd.Staged) {
		files, err = git.GetStagedFiles(gitRoot)
	} else if boolValue(cmd.Diff) {
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

	summary, err := lint.LintFiles(files, gitRoot, "", cfg.Quiet, cfg.Verbose)
	if err != nil {
		return err
	}

	if err := formatSummaryOutput(cfg, summary); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	printValidationReminder(cfg)
	applyFailurePolicy(cfg, opts, summary.TotalErrors, summary.TotalWarnings, summary.TotalSuggestions)

	return nil
}
