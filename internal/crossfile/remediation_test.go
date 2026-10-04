package crossfile

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

func TestSkillReferencesCompleteListsAndNamespaces(t *testing.T) {
	content := "Skill: plugin:plain\n**Skill**: plugin:bold\nSkill(\"plugin:called\")\nSkills:\n- alpha-skill\n  * beta-skill\n- `plugin:list`\nEnd list\n- unrelated-skill\nSkill:\nnot-a-reference\n"
	want := []string{"plugin:plain", "plugin:bold", "plugin:called", "alpha-skill", "beta-skill", "plugin:list"}
	if got := FindSkillReferences(content); !reflect.DeepEqual(got, want) {
		t.Fatalf("references = %v, want %v", got, want)
	}
}

func TestRuntimeSkillExemptionsAcrossChecks(t *testing.T) {
	v := NewCrossFileValidator(nil)
	contents := "Skill(review)\nSkill: plugin:helper\n**Skill**: init\nSkills:\n- focus\n- plugin:listed\n"
	if errs := v.checkSkillReferences("commands/example.md", contents); len(errs) != 0 {
		t.Fatalf("body refs: %v", errs)
	}
	for _, skills := range []any{"review, plugin:helper", []any{"review", "plugin:helper"}, []string{"review", "plugin:helper"}} {
		if errs := v.ValidateAgent("agents/example.md", contents, map[string]any{"skills": skills}); len(errs) != 0 {
			t.Fatalf("agent refs (%T): %v", skills, errs)
		}
	}
	for _, name := range []string{"review", "plugin:helper"} {
		if errs := v.validateTriggerRef(TriggerRef{File: "references/routes.md", RefType: "skill", RefName: name}); len(errs) != 0 {
			t.Fatalf("trigger %s: %v", name, errs)
		}
	}
	if errs := v.checkSkillReferences("commands/example.md", "Skill: missing-local"); len(errs) != 1 {
		t.Fatalf("missing local reference must remain visible: %v", errs)
	}
}

func TestComponentNamesNormalizeForeignSeparators(t *testing.T) {
	for _, path := range []string{"agents/helper.md", `agents\helper.md`} {
		if got := ExtractAgentName(path); got != "helper" {
			t.Fatalf("agent %q = %q", path, got)
		}
	}
	for _, path := range []string{"commands/helper.md", `commands\helper.md`} {
		if got := ExtractCommandName(path); got != "helper" {
			t.Fatalf("command %q = %q", path, got)
		}
	}
	for _, path := range []string{".claude/skills/helper/SKILL.md", `.claude\skills\helper\SKILL.md`} {
		if got := ExtractSkillName(path); got != "helper" {
			t.Fatalf("skill %q = %q", path, got)
		}
	}
}

func TestAgentSkillContentUsesSharedExtractor(t *testing.T) {
	files := []discovery.File{
		{Type: discovery.FileTypeSkill, RelPath: "skills/alpha/SKILL.md", Contents: "--alpha"},
		{Type: discovery.FileTypeSkill, RelPath: "skills/beta/SKILL.md", Contents: "--beta"},
		{Type: discovery.FileTypeSkill, RelPath: "skills/gamma/SKILL.md", Contents: "--gamma"},
		{Type: discovery.FileTypeSkill, RelPath: "skills/delta/SKILL.md", Contents: "--delta"},
	}
	v := NewCrossFileValidator(files)
	got := v.collectAgentSkillContents("**Skill**: alpha\nSkill(\"beta\")\nSkills:\n- gamma\n- delta")
	if len(got) != 4 {
		t.Fatalf("skill contents = %v", got)
	}
	for _, flag := range []string{"alpha", "beta", "gamma", "delta"} {
		if !v.isFlagInAgentOrSkills(flag, "", got) {
			t.Errorf("flag %s not resolved", flag)
		}
	}
}

func TestCyclesStableAcrossInputOrder(t *testing.T) {
	files := []discovery.File{
		{Type: discovery.FileTypeSkill, RelPath: "skills/alpha/SKILL.md", Contents: "Skill(beta)\nSkill(gamma)"},
		{Type: discovery.FileTypeSkill, RelPath: "skills/beta/SKILL.md", Contents: "Skill(alpha)"},
		{Type: discovery.FileTypeSkill, RelPath: "skills/gamma/SKILL.md", Contents: "Skill(alpha)"},
	}
	want := NewCrossFileValidator(files).DetectCycles()
	if len(want) != 2 {
		t.Fatalf("cycles = %v, want two", want)
	}
	for i := 0; i < 20; i++ {
		files[0], files[2] = files[2], files[0]
		if got := NewCrossFileValidator(files).DetectCycles(); !reflect.DeepEqual(got, want) {
			t.Fatalf("unstable cycles = %v, want %v", got, want)
		}
	}
}

func TestTriggerRowsRetainFinalCell(t *testing.T) {
	for _, suffix := range []string{"", " |"} {
		for _, header := range []string{"Target", "Other"} {
			table := "| Trigger | " + header + suffix + "\n| --- | ---" + suffix + "\n| build | build-skill" + suffix + "\n"
			refs := ParseTriggerTable("routes.md", table)
			if len(refs) != 1 || refs[0].RefName != "build-skill" {
				t.Fatalf("refs %q = %v", table, refs)
			}
			mappings := ParseTriggerMappings("routes.md", table)
			if len(mappings) != 1 || mappings[0].Target != "build-skill" {
				t.Fatalf("mappings %q = %v", table, mappings)
			}
		}
		row := "| build | build-skill" + suffix
		if refs := ExtractRefsFromRow("routes.md", row, map[string]bool{}, nil); len(refs) != 1 {
			t.Fatalf("fallback refs = %v", refs)
		}
		if mappings := extractMappingsFromRow("routes.md", row, nil); len(mappings) != 1 {
			t.Fatalf("fallback mappings = %v", mappings)
		}
	}
}

func TestDelegateViaNarrativeDistinguishesExplicitReferences(t *testing.T) {
	v := NewCrossFileValidator([]discovery.File{{Type: discovery.FileTypeAgent, RelPath: "agents/helper.md"}})
	v.userScopeAgentDir = ""
	if errs := v.ValidateSkill("skills/example/SKILL.md", "delegate via script\ndelegate via tool\ndelegate via shell-script\ndelegate via helper", nil); len(errs) != 0 {
		t.Fatalf("narrative prose or known agent incorrectly rejected: %v", errs)
	}
	for _, contents := range []string{"delegate via `missing`", `delegate via "missing"`, "delegate via 'missing'", "delegate via missing-agent"} {
		errs := v.ValidateSkill("skills/example/SKILL.md", contents, nil)
		if len(errs) != 1 || !strings.Contains(errs[0].Message, "doesn't exist") {
			t.Errorf("explicit %q = %v", contents, errs)
		}
	}
}

func TestKnownNarrativeAgentRetainsGraphEdge(t *testing.T) {
	files := []discovery.File{
		{Type: discovery.FileTypeAgent, RelPath: "agents/helper.md", Contents: "Skill(example)"},
		{Type: discovery.FileTypeSkill, RelPath: "skills/example/SKILL.md", Contents: "delegate via helper"},
	}
	if cycles := NewCrossFileValidator(files).DetectCycles(); len(cycles) != 1 {
		t.Fatalf("known narrative reference lost graph edge: %v", cycles)
	}
}

func TestRuntimeTriggerTargetsParseWithoutTruncation(t *testing.T) {
	table := "| Trigger | Skill |\n| --- | --- |\n| review | review |\n| helper | plugin:helper |\n"
	refs := ParseTriggerTable("routes.md", table)
	if len(refs) != 2 || refs[0].RefName != "review" || refs[1].RefName != "plugin:helper" {
		t.Fatalf("runtime refs = %v", refs)
	}
}
