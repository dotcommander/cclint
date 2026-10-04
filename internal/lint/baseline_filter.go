package lint

import (
	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/cue"
)

// FilterResults filters issues based on baseline, returning filtered issues and count of ignored issues
func FilterResults(summary *LintSummary, b *baseline.Baseline) (totalIgnored, errorsIgnored, suggestionsIgnored int) {
	if b == nil {
		return 0, 0, 0 // No baseline, no filtering
	}

	var total, errIgn, suggIgn int

	// Filter issues in each result
	for i := range summary.Results {
		result := &summary.Results[i]

		// Filter errors
		filteredErrors, errign := filterIssues(result.Errors, b.IsKnown)
		result.Errors = filteredErrors
		errIgn += errign
		total += errign

		// Filter warnings
		filteredWarnings, warnign := filterIssues(result.Warnings, b.IsKnown)
		result.Warnings = filteredWarnings
		total += warnign

		// Filter suggestions
		filteredSuggestions, suggign := filterIssues(result.Suggestions, b.IsKnown)
		result.Suggestions = filteredSuggestions
		suggIgn += suggign
		total += suggign

		// Update success status based on filtered errors
		result.Success = len(result.Errors) == 0
	}

	// Recalculate summary totals
	recalculateTotals(summary)

	return total, errIgn, suggIgn
}

// filterIssues filters a slice of issues using the provided filter function.
// Returns the filtered slice and the count of ignored issues.
func filterIssues(issues []cue.ValidationError, filter func(cue.ValidationError) bool) ([]cue.ValidationError, int) {
	filtered := make([]cue.ValidationError, 0, len(issues))
	ignored := 0
	for _, issue := range issues {
		if filter(issue) {
			ignored++
		} else {
			filtered = append(filtered, issue)
		}
	}
	return filtered, ignored
}

// recalculateTotals recalculates the summary totals based on the current results.
func recalculateTotals(summary *LintSummary) {
	var totalErrors, totalWarnings, totalSuggestions, successfulFiles, failedFiles int
	physicalFiles := make(map[string]bool)
	for i := range summary.Results {
		result := &summary.Results[i]
		result.Success = len(result.Errors) == 0
		if result.File != "" && result.Type != "memory" {
			physicalFiles[result.File] = physicalFiles[result.File] || !result.Success
		}
		totalErrors += len(result.Errors)
		totalWarnings += len(result.Warnings)
		totalSuggestions += len(result.Suggestions)
		if result.File == "" || result.Type == "memory" {
			continue
		}

	}
	for _, failed := range physicalFiles {
		if failed {
			failedFiles++
		} else {
			successfulFiles++
		}
	}
	summary.TotalFiles = len(physicalFiles)
	summary.TotalWarnings = totalWarnings
	summary.TotalErrors = totalErrors
	summary.TotalSuggestions = totalSuggestions
	summary.SuccessfulFiles = successfulFiles
	summary.FailedFiles = failedFiles
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
