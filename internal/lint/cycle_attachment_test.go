package lint

import (
	"testing"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

func TestBatchCycleReuseAcrossSkillSnapshotAndAgent(t *testing.T) {
	t.Parallel()
	files := []discovery.File{
		{RelPath: "agents/alpha.md", Type: discovery.FileTypeAgent, Contents: "Task(beta)"},
		{RelPath: "agents/beta.md", Type: discovery.FileTypeAgent, Contents: "Task(alpha)"},
	}
	ctx := &LinterContext{RootPath: t.TempDir(), CrossValidator: crossfile.NewCrossFileValidator(files)}
	NewSkillLinter(ctx.RootPath).PostProcessBatch(ctx, &LintSummary{})
	if len(ctx.cycles) == 0 {
		t.Fatal("skill batch did not populate invocation cycle cache")
	}
	first := &ctx.cycles[0]
	// Replacing the validator makes recomputation produce no cycles. The agent
	// must reuse the skill batch's original analysis, not merely match its output.
	ctx.CrossValidator = crossfile.NewCrossFileValidator(nil)
	summary := &LintSummary{TotalFiles: 2, SuccessfulFiles: 2, Results: []LintResult{
		{File: "agents/alpha.md", Success: true}, {File: "agents/beta.md", Success: true},
	}}
	NewAgentLinter().PostProcessBatch(ctx, summary)
	if &ctx.batchCycles()[0] != first || summary.TotalErrors != 2 {
		t.Fatalf("agent batch did not reuse skill analysis: %+v", summary)
	}
	fresh := &LinterContext{CrossValidator: crossfile.NewCrossFileValidator(nil)}
	if len(fresh.batchCycles()) != 0 {
		t.Fatal("separate invocation inherited cycle analysis")
	}
	disabled := &LinterContext{CrossValidator: crossfile.NewCrossFileValidator(files), NoCycleCheck: true}
	if disabled.batchCycles() != nil {
		t.Fatal("disabled cycle check ran")
	}
	disabled.NoCycleCheck = false
	if len(disabled.batchCycles()) == 0 {
		t.Fatal("disabled cycle check consumed lazy analysis")
	}
}

func TestIndexedIssueAttachmentPreservesFirstResultAndAppendOrder(t *testing.T) {
	t.Parallel()
	summary := &LintSummary{Results: []LintResult{{File: "skills/a/SKILL.md", Success: true}, {File: "skills/a/SKILL.md", Success: true}}}
	index := newSkillIssueIndex(summary)
	for _, message := range []string{"first", "second"} {
		index.attach(cue.ValidationError{File: "skills/a/SKILL.md", Message: message, Severity: cue.SeverityError}, true)
	}
	for _, message := range []string{"new", "later"} {
		index.attach(cue.ValidationError{File: "skills/b/SKILL.md", Message: message, Severity: cue.SeveritySuggestion}, true)
	}
	if len(summary.Results) != 3 || len(summary.Results[0].Errors) != 2 || len(summary.Results[1].Errors) != 0 ||
		summary.Results[0].Errors[0].Message != "first" || summary.Results[0].Errors[1].Message != "second" ||
		len(summary.Results[2].Suggestions) != 2 || summary.Results[2].Suggestions[1].Message != "later" {
		t.Fatalf("attachment changed first-match or append behavior: %+v", summary.Results)
	}
}

func TestIndexedCycleAttachmentsKeepAgentFirstAndSkillAll(t *testing.T) {
	t.Parallel()
	agents := &LintSummary{SuccessfulFiles: 2, Results: []LintResult{{File: "agents/alpha.md", Success: true}, {File: ".claude/agents/alpha.md", Success: true}}}
	addCycleToSummary(agents, "alpha", "alpha -> alpha")
	if agents.TotalErrors != 1 || agents.FailedFiles != 1 || len(agents.Results[1].Errors) != 0 {
		t.Fatalf("agent first-match attribution changed: %+v", agents)
	}
	skills := &LintSummary{Results: []LintResult{{File: "skills/one/SKILL.md"}, {File: ".claude/skills/one/SKILL.md"}}}
	cycle := crossfile.Cycle{Path: []string{"skill:one", "skill:one"}}
	attachSkillCycles(skills, []crossfile.Cycle{cycle, cycle})
	for _, result := range skills.Results {
		if len(result.Errors) != 1 || result.Errors[0].File != result.File {
			t.Fatalf("skill matching or duplicate suppression changed: %+v", skills.Results)
		}
	}
}
