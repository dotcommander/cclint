package lint

import "github.com/dotcommander/cclint/internal/cue"

// attachKind is retained for existing callers; severity determines the slice.
type attachKind int

const (
	attachAsError attachKind = iota
	attachAsSuggestion
)

// skillIssueIndex retains the first result for each file, including results
// appended by an earlier issue in this batch.
type skillIssueIndex struct {
	summary *LintSummary
	byFile  map[string]int
}

func newSkillIssueIndex(summary *LintSummary) *skillIssueIndex {
	index := &skillIssueIndex{summary: summary, byFile: make(map[string]int, len(summary.Results))}
	for i, result := range summary.Results {
		if _, exists := index.byFile[result.File]; !exists {
			index.byFile[result.File] = i
		}
	}
	return index
}

func (index *skillIssueIndex) attach(issue cue.ValidationError, createIfMissing bool) bool {
	if i, ok := index.byFile[issue.File]; ok {
		result := &index.summary.Results[i]
		categorizeIssues(result, []cue.ValidationError{issue})
		result.Success = len(result.Errors) == 0
		return false
	}
	if !createIfMissing {
		return false
	}
	entry := LintResult{File: issue.File, Type: "skill", Success: true}
	categorizeIssues(&entry, []cue.ValidationError{issue})
	entry.Success = len(entry.Errors) == 0
	index.byFile[issue.File] = len(index.summary.Results)
	index.summary.Results = append(index.summary.Results, entry)
	return true
}

// attachIssueToSummary preserves the direct helper API used by existing tests.
func attachIssueToSummary(summary *LintSummary, issue cue.ValidationError, _ attachKind, createIfMissing bool) bool {
	return newSkillIssueIndex(summary).attach(issue, createIfMissing)
}

func (index *skillIssueIndex) applyOrphanedSkills(ctx *LinterContext) {
	for _, issue := range ctx.CrossValidator.FindOrphanedSkills() {
		index.attach(issue, false)
	}
}

func (index *skillIssueIndex) applyGhostTriggers(ctx *LinterContext) {
	for _, issue := range ctx.CrossValidator.ValidateTriggerMaps(ctx.RootPath) {
		index.attach(issue, true)
	}
}

func (index *skillIssueIndex) applyTriggerConflicts(ctx *LinterContext) {
	for _, issue := range ctx.CrossValidator.DetectTriggerConflicts(ctx.RootPath) {
		index.attach(issue, true)
	}
}

func (index *skillIssueIndex) applySkillRefIssues(ctx *LinterContext) {
	for _, issue := range ctx.CrossValidator.ValidateSkillReferences(ctx.RootPath) {
		index.attach(issue, true)
	}
}

// Keep direct helpers available to package callers and tests; the batch path
// builds one index and shares it across all four issue sources.
func applyOrphanedSkills(ctx *LinterContext, summary *LintSummary) {
	newSkillIssueIndex(summary).applyOrphanedSkills(ctx)
}

func applyGhostTriggers(ctx *LinterContext, summary *LintSummary) {
	newSkillIssueIndex(summary).applyGhostTriggers(ctx)
}

func applyTriggerConflicts(ctx *LinterContext, summary *LintSummary) {
	newSkillIssueIndex(summary).applyTriggerConflicts(ctx)
}

func applySkillRefIssues(ctx *LinterContext, summary *LintSummary) {
	newSkillIssueIndex(summary).applySkillRefIssues(ctx)
}
