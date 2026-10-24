package cmd

import (
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/lint"
)

// lintOutcome is the CLI-owned result shape shared by every lint source.
// It separates how a result was produced from the invariant post-lint closure:
// format, report baseline filtering, remind, and apply the failure policy.
type lintOutcome struct {
	errors      int
	warnings    int
	suggestions int
	baseline    lint.IgnoredCounts
}

func fullLintOutcome(result *lint.Result) lintOutcome {
	return lintOutcome{
		errors:      result.TotalErrors,
		warnings:    result.TotalWarnings,
		suggestions: result.TotalSuggestions,
		baseline: lint.IgnoredCounts{
			Total:       result.BaselineIgnored,
			Errors:      result.ErrorsIgnored,
			Warnings:    result.WarningsIgnored,
			Suggestions: result.SuggestionsIgnored,
		},
	}
}

func summaryLintOutcome(summary *lint.LintSummary) lintOutcome {
	return lintOutcome{
		errors:      summary.TotalErrors,
		warnings:    summary.TotalWarnings,
		suggestions: summary.TotalSuggestions,
	}
}

func reportLintOutcome(cfg *config.Config, opts executionOptions, outcome lintOutcome, format func() error) error {
	if err := format(); err != nil {
		return err
	}

	printBaselineSummary(outcome.baseline, cfg.Quiet)
	printValidationReminder(cfg)
	applyFailurePolicy(cfg, opts, outcome.errors, outcome.warnings, outcome.suggestions)
	return nil
}
