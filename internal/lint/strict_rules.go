package lint

import "github.com/dotcommander/cclint/internal/cue"

// demoteSummaryWarnings moves warning-severity findings to suggestions for
// relaxed (rules.strict=false) runs. Errors are untouched.
func demoteSummaryWarnings(summary *LintSummary) {
	for i := range summary.Results {
		result := &summary.Results[i]
		for _, w := range result.Warnings {
			w.Severity = cue.SeveritySuggestion
			result.Suggestions = append(result.Suggestions, w)
		}
		result.Warnings = nil
	}
	recalculateTotals(summary)
}
