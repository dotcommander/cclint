package lint

import (
	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/cue"
)

// IgnoredCounts records issues removed by one baseline-filtering operation.
type IgnoredCounts struct {
	Total       int
	Errors      int
	Warnings    int
	Suggestions int
}

func (r *Result) accumulateIgnored(ignored IgnoredCounts) {
	r.BaselineIgnored += ignored.Total
	r.ErrorsIgnored += ignored.Errors
	r.WarningsIgnored += ignored.Warnings
	r.SuggestionsIgnored += ignored.Suggestions
}

// FilterResults removes known issues and returns their counts by severity.
func FilterResults(summary *LintSummary, b *baseline.Baseline) IgnoredCounts {
	if b == nil {
		return IgnoredCounts{} // No baseline, no filtering
	}

	var ignored IgnoredCounts

	// Filter issues in each result
	for i := range summary.Results {
		result := &summary.Results[i]

		// Filter errors
		filteredErrors, errign := filterIssues(result.Errors, b.IsKnown)
		result.Errors = filteredErrors
		ignored.Errors += errign

		// Filter warnings
		filteredWarnings, warnign := filterIssues(result.Warnings, b.IsKnown)
		result.Warnings = filteredWarnings
		ignored.Warnings += warnign

		// Filter suggestions
		filteredSuggestions, suggign := filterIssues(result.Suggestions, b.IsKnown)
		result.Suggestions = filteredSuggestions
		ignored.Suggestions += suggign

		// Update success status based on filtered errors
		result.Success = len(result.Errors) == 0
	}

	// Recalculate summary totals
	recalculateTotals(summary)

	ignored.Total = ignored.Errors + ignored.Warnings + ignored.Suggestions
	return ignored
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
