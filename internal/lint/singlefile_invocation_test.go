package lint

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
)

func TestSingleFileInvocationReusesValidatorsWithinRootOnly(t *testing.T) {
	t.Parallel()
	first, second := selectionTestRoot(t), selectionTestRoot(t)
	firstAgent := writeSelectionFile(t, first, "agents/first.md", "first\n")
	anotherAgent := writeSelectionFile(t, first, "agents/another.md", "another\n")
	secondAgent := writeSelectionFile(t, second, "agents/second.md", "second\n")
	firstSkill := "first-skill-only"
	secondSkill := "second-skill-only"
	writeSelectionFile(t, first, "skills/"+firstSkill+"/SKILL.md", "---\nname: "+firstSkill+"\n---\nSkill.\n")
	writeSelectionFile(t, second, "skills/"+secondSkill+"/SKILL.md", "---\nname: "+secondSkill+"\n---\nSkill.\n")

	run := &singleFileInvocation{}
	newContext := func(path, root string, invocation *singleFileInvocation) *SingleFileLinterContext {
		t.Helper()
		ctx, err := newSingleFileLinterContext(SingleFileRequest{
			FilePath: path, RootPath: root, invocation: invocation,
		})
		if err != nil {
			t.Fatal(err)
		}
		return ctx
	}
	a := newContext(firstAgent, first, run)
	b := newContext(anotherAgent, first, run)
	c := newContext(secondAgent, second, run)
	if a.Validator != b.Validator || b.Validator != c.Validator {
		t.Fatal("selected files did not share one CUE validator")
	}
	firstIndex, secondIndex := a.EnsureCrossFileValidator(), c.EnsureCrossFileValidator()
	if firstIndex == nil || secondIndex == nil || firstIndex == secondIndex || b.EnsureCrossFileValidator() != firstIndex {
		t.Fatal("cross-file indexes were not reused per root and isolated across roots")
	}
	if issues := firstIndex.ValidateAgent(firstAgent, "Skill: "+secondSkill, nil); len(issues) != 1 {
		t.Fatalf("other root's skill unexpectedly resolved: %+v", issues)
	}
	if issues := secondIndex.ValidateAgent(secondAgent, "Skill: "+firstSkill, nil); len(issues) != 1 {
		t.Fatalf("first root's skill unexpectedly resolved: %+v", issues)
	}

	// A later invocation observes new files and constructs fresh validators.
	lateSkill := "late-skill"
	writeSelectionFile(t, first, filepath.Join("skills", lateSkill, "SKILL.md"), "---\nname: "+lateSkill+"\n---\nSkill.\n")
	fresh := newContext(firstAgent, first, &singleFileInvocation{})
	if fresh.Validator == a.Validator || fresh.EnsureCrossFileValidator() == firstIndex {
		t.Fatal("independent invocation inherited a previous validator")
	}
	if issues := fresh.EnsureCrossFileValidator().ValidateAgent(firstAgent, "Skill: "+lateSkill, nil); len(issues) != 0 {
		t.Fatalf("new invocation missed a new skill: %+v", issues)
	}
	standaloneA := newContext(firstAgent, first, nil)
	standaloneB := newContext(anotherAgent, first, nil)
	if standaloneA.Validator == standaloneB.Validator || standaloneA.EnsureCrossFileValidator() == standaloneB.EnsureCrossFileValidator() {
		t.Fatal("standalone contexts unexpectedly share validators")
	}
}

func TestSingleFileInvocationRepeatsSchemaLoadWarningPerFile(t *testing.T) {
	t.Parallel()
	root := selectionTestRoot(t)
	paths := []string{
		writeSelectionFile(t, root, "agents/first.md", "first\n"),
		writeSelectionFile(t, root, "agents/second.md", "second\n"),
	}
	// Simulate a failed schema load without depending on the embedded schemas.
	run := &singleFileInvocation{validator: cue.NewValidator(), schemaErr: errSchemaFixture{}}
	for _, path := range paths {
		ctx, err := newSingleFileLinterContext(SingleFileRequest{FilePath: path, RootPath: root, invocation: run})
		if err != nil {
			t.Fatal(err)
		}
		if ctx.Validator != run.validator || len(ctx.Warnings) != 1 || !strings.Contains(ctx.Warnings[0].Message, "fixture schema failure") || ctx.Warnings[0].File != ctx.File.RelPath {
			t.Fatalf("file %s lost its own schema warning: %+v", path, ctx.Warnings)
		}
	}
}

type errSchemaFixture struct{}

func (errSchemaFixture) Error() string { return "fixture schema failure" }
