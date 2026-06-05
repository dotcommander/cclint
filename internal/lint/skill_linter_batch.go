package lint

import "github.com/dotcommander/cclint/internal/cue"

// attachKind selects which LintResult slice (Errors vs Suggestions) an issue
// is appended to, plus the side effects of the error path (mark Success=false
// on attach, increment FailedFiles on create).
type attachKind int

const (
	attachAsError attachKind = iota
	attachAsSuggestion
)

// attachIssueToSummary finds an existing LintResult matching issue.File and
// appends the issue to its Errors or Suggestions slice. If no match exists
// and createIfMissing is true, a new LintResult entry is created. Returns
// true if a new entry was created (the caller updates FailedFiles for the
// error path).
//
// Order preservation: scans summary.Results in index order, breaks on first
// match — identical semantics to the four prior hand-rolled loops.
func attachIssueToSummary(summary *LintSummary, issue cue.ValidationError, kind attachKind, createIfMissing bool) (created bool) {
	for i, result := range summary.Results {
		if result.File != issue.File {
			continue
		}
		if kind == attachAsError {
			summary.Results[i].Errors = append(summary.Results[i].Errors, issue)
			summary.Results[i].Success = false
		} else {
			summary.Results[i].Suggestions = append(summary.Results[i].Suggestions, issue)
		}
		return false
	}
	if !createIfMissing {
		return false
	}
	entry := LintResult{File: issue.File, Type: "skill"}
	if kind == attachAsError {
		entry.Success = false
		entry.Errors = []cue.ValidationError{issue}
	} else {
		entry.Success = true // suggestions do not fail the build
		entry.Suggestions = []cue.ValidationError{issue}
	}
	summary.Results = append(summary.Results, entry)
	return true
}

// applyOrphanedSkills appends orphan-detection suggestions to existing results.
func applyOrphanedSkills(ctx *LinterContext, summary *LintSummary) {
	for _, orphan := range ctx.CrossValidator.FindOrphanedSkills() {
		summary.TotalSuggestions++
		// Orphans only attach to existing file results; no fallback entry.
		attachIssueToSummary(summary, orphan, attachAsSuggestion, false)
	}
}

// applyGhostTriggers validates skill/agent refs in trigger map tables and appends errors.
func applyGhostTriggers(ctx *LinterContext, summary *LintSummary) {
	for _, gt := range ctx.CrossValidator.ValidateTriggerMaps(ctx.RootPath) {
		summary.TotalErrors++
		summary.FailedFiles++
		// Reference files are not in normal results, so always create a new entry.
		summary.Results = append(summary.Results, LintResult{
			File:    gt.File,
			Type:    "skill",
			Success: false,
			Errors:  []cue.ValidationError{gt},
		})
	}
}

// applyTriggerConflicts detects same-keyword conflicts across files and appends suggestions.
func applyTriggerConflicts(ctx *LinterContext, summary *LintSummary) {
	for _, tc := range ctx.CrossValidator.DetectTriggerConflicts(ctx.RootPath) {
		summary.TotalSuggestions++
		summary.Results = append(summary.Results, LintResult{
			File:        tc.File,
			Type:        "skill",
			Success:     true, // warnings do not fail the build
			Suggestions: []cue.ValidationError{tc},
		})
	}
}

// applySkillRefIssues validates skill reference files for phantom and orphaned refs.
func applySkillRefIssues(ctx *LinterContext, summary *LintSummary) {
	for _, issue := range ctx.CrossValidator.ValidateSkillReferences(ctx.RootPath) {
		if issue.Severity == cue.SeverityError {
			summary.TotalErrors++
			if attachIssueToSummary(summary, issue, attachAsError, true) {
				summary.FailedFiles++
			}
		} else {
			summary.TotalSuggestions++
			attachIssueToSummary(summary, issue, attachAsSuggestion, true)
		}
	}
}
