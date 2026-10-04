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
	for i := range summary.Results {
		if summary.Results[i].File == issue.File {
			categorizeIssues(&summary.Results[i], []cue.ValidationError{issue})
			summary.Results[i].Success = len(summary.Results[i].Errors) == 0
			return false
		}
	}
	if !createIfMissing {
		return false
	}
	entry := LintResult{File: issue.File, Type: "skill", Success: true}
	categorizeIssues(&entry, []cue.ValidationError{issue})
	entry.Success = len(entry.Errors) == 0
	summary.Results = append(summary.Results, entry)
	return true
}

// applyOrphanedSkills appends orphan-detection suggestions to existing results.
func applyOrphanedSkills(ctx *LinterContext, summary *LintSummary) {
	for _, orphan := range ctx.CrossValidator.FindOrphanedSkills() {
		// Orphans only attach to existing file results; no fallback entry.
		attachIssueToSummary(summary, orphan, attachAsSuggestion, false)
	}
}

// applyGhostTriggers validates skill/agent refs in trigger map tables and appends errors.
func applyGhostTriggers(ctx *LinterContext, summary *LintSummary) {
	for _, issue := range ctx.CrossValidator.ValidateTriggerMaps(ctx.RootPath) {
		attachIssueToSummary(summary, issue, attachAsError, true)
	}
}

func applyTriggerConflicts(ctx *LinterContext, summary *LintSummary) {
	for _, issue := range ctx.CrossValidator.DetectTriggerConflicts(ctx.RootPath) {
		attachIssueToSummary(summary, issue, attachAsSuggestion, true)
	}
}

func applySkillRefIssues(ctx *LinterContext, summary *LintSummary) {
	for _, issue := range ctx.CrossValidator.ValidateSkillReferences(ctx.RootPath) {
		attachIssueToSummary(summary, issue, attachAsError, true)
	}
}
