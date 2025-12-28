package cli

import (
	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/cue"
)

// FilterResults filters issues based on baseline, returning filtered issues and count of ignored issues
func FilterResults(summary *LintSummary, b *baseline.Baseline) (int, int, int) {
	if b == nil {
		return 0, 0, 0 // No baseline, no filtering
	}

	var totalIgnored, errorsIgnored, suggestionsIgnored int

	// Filter issues in each result
	for i := range summary.Results {
		result := &summary.Results[i]

		// Filter errors
		filteredErrors := make([]cue.ValidationError, 0, len(result.Errors))
		for _, err := range result.Errors {
			if b.IsKnown(err) {
				errorsIgnored++
				totalIgnored++
			} else {
				filteredErrors = append(filteredErrors, err)
			}
		}
		result.Errors = filteredErrors

		// Filter warnings
		filteredWarnings := make([]cue.ValidationError, 0, len(result.Warnings))
		for _, warn := range result.Warnings {
			if b.IsKnown(warn) {
				totalIgnored++
			} else {
				filteredWarnings = append(filteredWarnings, warn)
			}
		}
		result.Warnings = filteredWarnings

		// Filter suggestions
		filteredSuggestions := make([]cue.ValidationError, 0, len(result.Suggestions))
		for _, sugg := range result.Suggestions {
			if b.IsKnown(sugg) {
				suggestionsIgnored++
				totalIgnored++
			} else {
				filteredSuggestions = append(filteredSuggestions, sugg)
			}
		}
		result.Suggestions = filteredSuggestions

		// Update success status based on filtered errors
		if len(result.Errors) == 0 {
			result.Success = true
		} else {
			result.Success = false
		}
	}

	// Recalculate summary totals
	var totalErrors, totalSuggestions, successfulFiles, failedFiles int
	for _, result := range summary.Results {
		totalErrors += len(result.Errors)
		totalSuggestions += len(result.Suggestions)
		if result.Success {
			successfulFiles++
		} else {
			failedFiles++
		}
	}

	summary.TotalErrors = totalErrors
	summary.TotalSuggestions = totalSuggestions
	summary.SuccessfulFiles = successfulFiles
	summary.FailedFiles = failedFiles

	return totalIgnored, errorsIgnored, suggestionsIgnored
}

// CollectAllIssues collects all validation errors from a summary (for baseline creation)
func CollectAllIssues(summary *LintSummary) []cue.ValidationError {
	var issues []cue.ValidationError

	for _, result := range summary.Results {
		issues = append(issues, result.Errors...)
		issues = append(issues, result.Warnings...)
		issues = append(issues, result.Suggestions...)
	}

	return issues
}
