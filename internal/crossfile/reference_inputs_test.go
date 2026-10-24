package crossfile

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

func TestUserScopeSkillReferences(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("isolated filesystem integration")
	}
	for _, filename := range []string{"SKILL.md", "skill.md"} {
		t.Run(filename, func(t *testing.T) {
			t.Parallel()
			dir := t.TempDir()
			writeUserScopeSkill(t, dir, "personal-input-skill", filename)
			v := NewCrossFileValidator([]discovery.File{
				{Type: discovery.FileTypeSkill, RelPath: "skills/project/SKILL.md"},
			})
			v.userScopeSkillDir = dir
			v.userScopeAgentDir = ""

			body := "Skill: personal-input-skill"
			if got := v.ValidateAgent("agents/consumer.md", body, map[string]any{
				"skills": []any{"personal-input-skill"},
			}); len(got) != 0 {
				t.Fatalf("agent body/frontmatter: %v", got)
			}
			if got := v.ValidateCommand("commands/consumer.md", body, nil); len(got) != 0 {
				t.Fatalf("command body: %v", got)
			}
			if got := v.validateTriggerRef(TriggerRef{
				RefType: "skill", RefName: "personal-input-skill", File: "references/triggers.md",
			}); len(got) != 0 {
				t.Fatalf("trigger: %v", got)
			}
			if len(v.skills) != 1 || !v.hasResolvableSkill("project") {
				t.Fatalf("local index changed: %v", v.skills)
			}
			if got := v.FindOrphanedSkills(); len(got) != 1 {
				t.Fatalf("home skill entered orphan analysis: %v", got)
			}
		})
	}
}

func TestUserScopeSkillFallbackBoundaries(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("isolated filesystem integration")
	}
	root := t.TempDir()
	dir := filepath.Join(root, "skills")
	writeUserScopeSkill(t, dir, "personal-input-skill", "SKILL.md")
	writeUserScopeSkill(t, root, "outside", "SKILL.md")
	for _, filename := range []string{"SKILL.md", "skill.md"} {
		if err := os.MkdirAll(filepath.Join(dir, "directory", filename), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(dir, "broken"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(filepath.Join(root, "absent"), filepath.Join(dir, "broken", "SKILL.md")); err != nil {
		t.Fatal(err)
	}
	v := NewCrossFileValidator(nil)
	v.userScopeSkillDir = dir
	for _, name := range []string{
		"", ".", "..", "../outside", `..\outside`, filepath.Join(root, "outside"),
		"missing", "directory", "broken",
	} {
		if v.hasResolvableSkill(name) {
			t.Errorf("unexpected filesystem resolution for %q", name)
		}
	}
	v.userScopeSkillDir = ""
	if got := v.ValidateAgent("agents/consumer.md", "Skill: personal-input-skill", nil); len(got) != 1 ||
		got[0].Severity != cue.SeverityError || got[0].Source != cue.SourceCClintObserve {
		t.Fatalf("unavailable home must retain missing-reference error: %v", got)
	}
	v.skills["../indexed"] = discovery.File{Type: discovery.FileTypeSkill}
	for _, name := range []string{"init", "plugin:external", "../indexed"} {
		if !v.hasResolvableSkill(name) {
			t.Errorf("existing runtime/plugin/index resolution changed for %q", name)
		}
	}
}

func writeUserScopeSkill(t *testing.T, root, name, filename string) {
	t.Helper()
	path := filepath.Join(root, name, filename)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	// Deliberately not parsed: resolution checks presence, not skill contents.
	if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
}

func TestAllowedToolsRepresentationEquivalence(t *testing.T) {
	t.Parallel()
	v := NewCrossFileValidator(nil)
	body := "Read(input)"
	path := "commands/consumer.md"
	scalar := "Read, Task(helper), Write, Read, Task(helper), *"
	want := v.ValidateCommand(path, body, map[string]any{"allowed-tools": scalar})
	if len(want) != 3 {
		t.Fatalf("scalar unused-tool diagnostics: %v", want)
	}
	for i, token := range []string{"'Task(helper)'", "'Write'", "'*'"} {
		if want[i].File != path || want[i].Severity != cue.SeverityInfo ||
			want[i].Source != cue.SourceCClintObserve || !strings.Contains(want[i].Message, token) {
			t.Fatalf("scalar diagnostic %d changed: %+v", i, want[i])
		}
	}
	for _, tc := range []struct {
		name  string
		value any
	}{
		{"strings", []string{"Read", "Task(helper)", "Write", "Read", "Task(helper)", "*"}},
		{"mixed", []any{"Read", nil, "Task(helper)", 42, "Write", "Read", "Task(helper)", "*"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got := v.ValidateCommand(path, body, map[string]any{"allowed-tools": tc.value})
			if !slices.Equal(got, want) {
				t.Fatalf("array diagnostics differ from scalar:\ngot: %+v\nwant: %+v", got, want)
			}
		})
	}
	for _, value := range []any{nil, 42, []int{1}, []any{nil, true, 42}} {
		if got := v.ValidateCommand(path, body, map[string]any{"allowed-tools": value}); len(got) != 0 {
			t.Errorf("unsupported input %T produced diagnostics: %v", value, got)
		}
	}
}
