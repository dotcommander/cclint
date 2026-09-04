package cmd

import (
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/lint"
)

// lintBaselineOutcome carries ignored-issue counts produced by baseline
// filtering. Zero values mean the run did not use a baseline.
type lintBaselineOutcome struct {
	total       int
	errors      int
	suggestions int
}

// lintOutcome is the CLI-owned result shape shared by every lint source.
// It separates how a result was produced from the invariant post-lint closure:
// format, report baseline filtering, remind, and apply the failure policy.
type lintOutcome struct {
	errors      int
	warnings    int
	suggestions int
	baseline    lintBaselineOutcome
}

func fullLintOutcome(result *lint.Result) lintOutcome {
	return lintOutcome{
		errors:      result.TotalErrors,
		warnings:    result.TotalWarnings,
		suggestions: result.TotalSuggestions,
		baseline: lintBaselineOutcome{
			total:       result.BaselineIgnored,
			errors:      result.ErrorsIgnored,
			suggestions: result.SuggestionsIgnored,
		},
	}
}

func componentLintOutcome(result *lint.Result, summary *lint.LintSummary) lintOutcome {
	outcome := summaryLintOutcome(summary)
	outcome.baseline = lintBaselineOutcome{
		total:       result.BaselineIgnored,
		errors:      result.ErrorsIgnored,
		suggestions: result.SuggestionsIgnored,
	}
	return outcome
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

	printBaselineSummary(outcome.baseline.total, outcome.baseline.errors, outcome.baseline.suggestions, cfg.Quiet)
	printValidationReminder(cfg)
	applyFailurePolicy(cfg, opts, outcome.errors, outcome.warnings, outcome.suggestions)
	return nil
}
