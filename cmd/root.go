package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/git"
	"github.com/dotcommander/cclint/internal/lint"
	"golang.org/x/term"
)

const failOnWarning = "warning"

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
	case failOnWarning:
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
	return runLintWithConfig(cfg, opts)
}

func runLintWithConfig(cfg *config.Config, opts executionOptions) error {
	result, err := runOrchestratedLint(cfg, opts, nil)
	if err != nil {
		return err
	}

	err = reportLintOutcome(cfg, opts, fullLintOutcome(result), func() error {
		return formatFullRunOutput(cfg, result)
	})
	if err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}

func runRootCommand(opts executionOptions, cmd *lintCommand) error {
	if boolValue(cmd.Diff) || boolValue(cmd.Staged) {
		if len(cmd.Paths) > 0 {
			return fmt.Errorf("cannot combine Git selection flags with positional paths or types")
		}
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
		return runTypesLint(opts, classified.typeFilters)
	default:
		if stringValue(cmd.Type) != "" {
			ft, err := discovery.ParseFileType(stringValue(cmd.Type))
			if err != nil {
				return err
			}
			return runTypesLint(opts, []discovery.FileType{ft})
		}
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

	selectionRoot := stringValue(opts.root)
	if selectionRoot == "" && cfg.RootExplicit {
		selectionRoot = cfg.Root
	}
	summary, err := lint.LintFiles(files, selectionRoot, stringValue(cmd.Type), cfg.Quiet, cfg.Verbose)
	if err != nil {
		return err
	}

	if stringValue(opts.root) == "" && !cfg.RootExplicit && summary.ProjectRoot != "" {
		cfg.Root = summary.ProjectRoot
	}
	result, err := processSelected(cfg, opts, summary)
	if err != nil {
		return err
	}
	if err := reportLintOutcome(cfg, opts, fullLintOutcome(result), func() error {
		return formatSummaryOutput(cfg, summary)
	}); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

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
	if stringValue(opts.root) == "" && !cfg.RootExplicit {
		// Use current working directory for git operations
		gitRoot, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	}

	repoRoot, rootErr := git.RepositoryRoot(gitRoot)
	if rootErr == nil && stringValue(opts.root) == "" && !cfg.RootExplicit {
		cfg.Root = repoRoot
		gitRoot = repoRoot
	}
	// Check if in git repository
	if !git.IsGitRepo(gitRoot) {
		if !cfg.Quiet {
			fmt.Fprintf(os.Stderr, "Warning: Not in a git repository. Falling back to full lint.\n\n")
		}
		return runLint(opts)
	}

	// Get files from git
	var changes git.FileChanges
	if boolValue(cmd.Staged) {
		changes, err = git.GetStagedChanges(gitRoot)
	} else if boolValue(cmd.Diff) {
		changes, err = git.GetChangedFileChanges(gitRoot)
	}
	if err != nil {
		return fmt.Errorf("error getting git files: %w", err)
	}

	// Git returns canonical repository paths; keep the explicit root's logical
	// spelling for discovery, diagnostics and exclusions.
	canonicalRoot, canonicalErr := filepath.EvalSymlinks(cfg.Root)
	if canonicalErr != nil {
		return fmt.Errorf("cannot resolve project root: %w", canonicalErr)
	}
	selected := changes.Files[:0]
	for _, file := range changes.Files {
		relative, e := filepath.Rel(canonicalRoot, file)
		if e == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
			selected = append(selected, filepath.Join(cfg.Root, relative))
		}
	}
	changes.Files = selected
	if changes.HasRelevantDeletion() {
		for _, deleted := range changes.DeletedFiles {
			absolute := filepath.Join(repoRoot, deleted)
			relative, e := filepath.Rel(canonicalRoot, absolute)
			if e == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
				cfg.Exclude = append(cfg.Exclude, filepath.ToSlash(relative))
			}
		}
		return runLintWithConfig(cfg, opts)
	}

	if len(changes.Files) == 0 {
		if !cfg.Quiet {
			fmt.Println("No files to lint")
		}
		return nil
	}

	summary, err := lint.LintFiles(changes.Files, cfg.Root, stringValue(cmd.Type), cfg.Quiet, cfg.Verbose)
	if err != nil {
		return err
	}

	result, err := processSelected(cfg, opts, summary)
	if err != nil {
		return err
	}
	if err := reportLintOutcome(cfg, opts, fullLintOutcome(result), func() error {
		return formatSummaryOutput(cfg, summary)
	}); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}
