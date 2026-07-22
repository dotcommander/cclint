package cmd

import (
	"fmt"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/lint"
)

// runTypeLint runs the linter for a specific file type.
func runTypeLint(opts executionOptions, ft discovery.FileType) error {
	entry, ok := lint.LinterForType(ft)
	if !ok {
		return fmt.Errorf("no linter for type %s", ft)
	}
	return runComponentLint(opts, entry)
}

// runComponentLint is the generic function that handles config loading,
// linter execution, and output formatting for any component type.
// This follows the Single Responsibility Principle by separating
// orchestration from component-specific linting logic.
func runComponentLint(opts executionOptions, entry lint.LinterEntry) error {
	cfg, err := loadCLIConfig(opts)
	if err != nil {
		return err
	}

	result, err := runOrchestratedLint(cfg, opts, []lint.LinterEntry{entry})
	if err != nil {
		return fmt.Errorf("error running %s linter: %w", entry.Name, err)
	}

	summary := &lint.LintSummary{}
	if len(result.Summaries) > 0 {
		summary = result.Summaries[0]
	}

	if err := formatSummaryOutput(cfg, summary); err != nil {
		return err
	}

	printBaselineSummary(result.BaselineIgnored, result.ErrorsIgnored, result.SuggestionsIgnored, cfg.Quiet)
	printValidationReminder(cfg)
	applyFailurePolicy(cfg, opts, summary.TotalErrors, summary.TotalWarnings, summary.TotalSuggestions)

	return nil
}
