package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/git"
	"github.com/dotcommander/cclint/internal/lint"
)

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
		if e == nil && filepath.IsLocal(relative) {
			if !discovery.IsExcluded(filepath.ToSlash(relative), cfg.Exclude) {
				selected = append(selected, filepath.Join(cfg.Root, relative))
			}
		}
	}
	changes.Files = selected
	if changes.HasRelevantDeletion() {
		for _, deleted := range changes.DeletedFiles {
			absolute := filepath.Join(repoRoot, deleted)
			relative, e := filepath.Rel(canonicalRoot, absolute)
			if e == nil && filepath.IsLocal(relative) {
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
