package output

import (
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/lint"
)

// Severity classifies a FlatIssue. The string values match the severity
// labels the formatters already switch on ("error", "warning", "suggestion").
type Severity string

const (
	SeverityError      Severity = "error"
	SeverityWarning    Severity = "warning"
	SeveritySuggestion Severity = "suggestion"
)

// FlatIssue is the single intermediate representation every formatter consumes.
// It flattens LintSummary.Results[*].{Errors,Warnings,Suggestions} into one
// ordered slice while preserving the originating file, component type, and the
// raw cue.ValidationError. Ordering is stable: results in summary order, and
// within each result Errors then Warnings then Suggestions — identical to the
// pre-IR traversal order in every formatter.
type FlatIssue struct {
	ComponentType string // from LintSummary.ComponentType
	File          string // from LintResult.File
	ResultIndex   int    // index into LintSummary.Results (per summary)
	Severity      Severity
	Err           cue.ValidationError
}

// BuildFlatIssues is the ONLY place that traverses LintResult.Errors/Warnings/
// Suggestions. All formatters route through it so a counting bug is fixed once.
func BuildFlatIssues(summary *lint.LintSummary) []FlatIssue {
	if summary == nil {
		return nil
	}
	var issues []FlatIssue
	for i := range summary.Results {
		result := &summary.Results[i]
		for _, e := range result.Errors {
			issues = append(issues, FlatIssue{
				ComponentType: summary.ComponentType,
				File:          result.File,
				ResultIndex:   i,
				Severity:      SeverityError,
				Err:           e,
			})
		}
		for _, w := range result.Warnings {
			issues = append(issues, FlatIssue{
				ComponentType: summary.ComponentType,
				File:          result.File,
				ResultIndex:   i,
				Severity:      SeverityWarning,
				Err:           w,
			})
		}
		for _, s := range result.Suggestions {
			issues = append(issues, FlatIssue{
				ComponentType: summary.ComponentType,
				File:          result.File,
				ResultIndex:   i,
				Severity:      SeveritySuggestion,
				Err:           s,
			})
		}
	}
	return issues
}

// issuesForResult returns the subset of issues whose ResultIndex == idx.
// Used by formatters that render per-file (console) to avoid re-touching
// LintResult.Errors directly.
func issuesForResult(issues []FlatIssue, idx int) []FlatIssue {
	var out []FlatIssue
	for _, is := range issues {
		if is.ResultIndex == idx {
			out = append(out, is)
		}
	}
	return out
}

// countBySeverity counts issues of a given severity within a slice.
func countBySeverity(issues []FlatIssue, sev Severity) int {
	n := 0
	for _, is := range issues {
		if is.Severity == sev {
			n++
		}
	}
	return n
}

// severityErrors returns the raw cue.ValidationError values of a given severity,
// in slice order. Adapter for callers (markdown) that render []cue.ValidationError.
func severityErrors(issues []FlatIssue, sev Severity) []cue.ValidationError {
	var out []cue.ValidationError
	for _, is := range issues {
		if is.Severity == sev {
			out = append(out, is.Err)
		}
	}
	return out
}
