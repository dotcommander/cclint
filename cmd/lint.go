package cmd

import (
	"fmt"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/outputters"
)

// LinterFunc is the function signature for component linters.
// It takes root path, quiet mode, verbose mode and returns a summary.
type LinterFunc func(rootPath string, quiet bool, verbose bool) (*cli.LintSummary, error)

// runComponentLint is the generic function that handles config loading,
// linter execution, and output formatting for any component type.
// This follows the Single Responsibility Principle by separating
// orchestration from component-specific linting logic.
func runComponentLint(linterName string, linter LinterFunc) error {
	// Load configuration
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Run component-specific linter
	summary, err := linter(cfg.Root, cfg.Quiet, cfg.Verbose)
	if err != nil {
		return fmt.Errorf("error running %s linter: %w", linterName, err)
	}

	// Format and output results
	outputter := outputters.NewOutputter(cfg)
	if err := outputter.Format(summary, cfg.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	return nil
}
