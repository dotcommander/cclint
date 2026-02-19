package crossfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dotcommander/cclint/internal/discovery"
)

// ---------------------------------------------------------------------------
// IsTriggerMap
// ---------------------------------------------------------------------------

func TestIsTriggerMap(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     bool
	}{
		{
			name: "trigger column header",
			contents: `| Trigger | Agent |
|---------|-------|
| refactor | poet-agent |`,
			want: true,
		},
		{
			name: "trigger column header case-insensitive uppercase",
			contents: `| TRIGGER | Skill |
|---------|-------|
| test    | code-testing-qa |`,
			want: true,
		},
		{
			name: "trigger column with extra text in header",
			contents: `| Trigger word | Skill name |
|-------------|-----------|
| meeseeks    | meeseeks-methodology |`,
			want: true,
		},
		{
			name: "table without trigger column",
			contents: `| Name | Value |
|------|-------|
| foo  | bar   |`,
			want: false,
		},
		{
			name: "no table at all",
			contents: `# Just a heading

Some prose without any tables.`,
			want: false,
		},
		{
			name: "empty string",
			contents: "",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsTriggerMap(tt.contents)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// IsSeparatorRow
// ---------------------------------------------------------------------------

func TestIsSeparatorRow(t *testing.T) {
	tests := []struct {
		name string
		line string
		want bool
	}{
		{name: "simple separator", line: "|---|---|", want: true},
		{name: "separator with colons", line: "|:---|:---:|---:|", want: true},
		{name: "separator with spaces", line: "| --- | --- |", want: true},
		{name: "data row", line: "| foo | bar |", want: false},
		{name: "header row", line: "| Trigger | Agent |", want: false},
		{name: "not a table row", line: "some text", want: false},
		{name: "just dashes no pipe", line: "---", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSeparatorRow(tt.line)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// IsLikelySkillName
// ---------------------------------------------------------------------------

func TestIsLikelySkillName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{name: "hyphenated skill name", input: "arch-database-core", want: true},
		{name: "two-part skill name", input: "code-testing", want: true},
		{name: "agent name", input: "poet-agent", want: true},
		{name: "single word no hyphen", input: "refactor", want: false},
		{name: "short string", input: "go", want: false},
		{name: "three chars no hyphen", input: "foo", want: false},
		{name: "single char", input: "a", want: false},
		{name: "model name sonnet no hyphen", input: "sonnet", want: false},
		{name: "model name haiku no hyphen", input: "haiku", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsLikelySkillName(tt.input)
			assert.Equal(t, tt.want, got)
		})
	}
}

// ---------------------------------------------------------------------------
// ParseTriggerTable
// ---------------------------------------------------------------------------

func TestParseTriggerTable(t *testing.T) {
	tests := []struct {
		name      string
		contents  string
		wantRefs  []TriggerRef
		wantCount int
	}{
		{
			name: "extracts skill refs from bare names",
			contents: `| Trigger | Skill |
|---------|-------|
| test coverage | code-testing-qa |
| refactor | code-clean-code |`,
			wantRefs: []TriggerRef{
				{File: "test.md", RefType: "skill", RefName: "code-testing-qa"},
				{File: "test.md", RefType: "skill", RefName: "code-clean-code"},
			},
		},
		{
			name: "extracts agent refs from Task() patterns",
			contents: `| Trigger | Agent |
|---------|-------|
| run tests | Task(quality-agent) |
| commit changes | Task(commit-agent) |`,
			wantRefs: []TriggerRef{
				{File: "test.md", RefType: "agent", RefName: "quality-agent"},
				{File: "test.md", RefType: "agent", RefName: "commit-agent"},
			},
		},
		{
			name: "handles backtick-wrapped Task()",
			contents: "| Trigger | Agent |\n|---------|-------|\n| commit | `Task(commit-agent)` |",
			wantRefs: []TriggerRef{
				{File: "test.md", RefType: "agent", RefName: "commit-agent"},
			},
		},
		{
			name: "ignores prose words without hyphens",
			contents: `| Trigger | Description |
|---------|-------------|
| foo bar | Run the refactor task now |`,
			wantRefs: nil,
		},
		{
			name: "skips separator rows",
			contents: `| Trigger | Skill |
|---------|-------|
| meeseeks | meeseeks-methodology |`,
			wantRefs: []TriggerRef{
				{File: "test.md", RefType: "skill", RefName: "meeseeks-methodology"},
			},
		},
		{
			name: "deduplicates repeated refs",
			contents: `| Trigger | Skill |
|---------|-------|
| test | code-testing-qa |
| tests | code-testing-qa |`,
			wantCount: 1,
		},
		{
			name: "no trigger table returns empty",
			contents: `| Name | Value |
|------|-------|
| foo  | bar   |`,
			wantRefs: nil,
		},
		{
			name: "names in Never Do column are not extracted",
			contents: `| Trigger | Route To | Never Do |
|---------|----------|----------|
| "lint me", "golangci-lint" | ` + "`Task(poet-agent)`" + ` | ` + "`Bash(\"golangci-lint\")`" + ` direct |`,
			wantRefs: []TriggerRef{
				{File: "test.md", RefType: "agent", RefName: "poet-agent"},
			},
		},
		{
			name: "routing column with only Skill header works as before",
			contents: `| Trigger | Skill |
|---------|-------|
| test coverage | code-testing-qa |`,
			wantRefs: []TriggerRef{
				{File: "test.md", RefType: "skill", RefName: "code-testing-qa"},
			},
		},
		{
			name: "references/ path in routing column is not extracted as skill",
			contents: `| Trigger | Agent |
|---------|-------|
| "meeseeks" | agent-claude-code (Read references/meeseeks-methodology.md) |`,
			wantRefs: []TriggerRef{
				{File: "test.md", RefType: "skill", RefName: "agent-claude-code"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTriggerTable("test.md", tt.contents)

			if tt.wantRefs != nil {
				require.Equal(t, tt.wantRefs, got)
			}
			if tt.wantCount > 0 {
				assert.Len(t, got, tt.wantCount)
			}
			if tt.wantRefs == nil && tt.wantCount == 0 {
				assert.Empty(t, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// discoverReferenceFiles
// ---------------------------------------------------------------------------

func TestDiscoverReferenceFiles(t *testing.T) {
	root := t.TempDir()

	// Create a references file under skills/my-skill/references/
	skillRefDir := filepath.Join(root, "skills", "my-skill", "references")
	require.NoError(t, os.MkdirAll(skillRefDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(skillRefDir, "triggers.md"), []byte("content"), 0o644))

	// Create a references file under .claude/skills/other-skill/references/
	claudeRefDir := filepath.Join(root, ".claude", "skills", "other-skill", "references")
	require.NoError(t, os.MkdirAll(claudeRefDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(claudeRefDir, "routing.md"), []byte("content"), 0o644))

	got := discoverReferenceFiles(root)
	assert.Len(t, got, 2)
}

// ---------------------------------------------------------------------------
// ValidateTriggerMaps integration
// ---------------------------------------------------------------------------

func TestValidateTriggerMaps(t *testing.T) {
	t.Run("ghost skill detected", func(t *testing.T) {
		root := t.TempDir()

		// Create a reference file with a trigger map referencing a ghost skill
		refDir := filepath.Join(root, "skills", "agent-routing", "references")
		require.NoError(t, os.MkdirAll(refDir, 0o755))

		triggerContent := `# Trigger Map

| Trigger | Skill |
|---------|-------|
| test | ghost-skill |
| refactor | real-skill |
`
		require.NoError(t, os.WriteFile(filepath.Join(refDir, "triggers.md"), []byte(triggerContent), 0o644))

		// Build a validator that knows about real-skill but not ghost-skill
		realSkillFile := discovery.File{
			RelPath:  "skills/real-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "",
		}
		v := NewCrossFileValidator([]discovery.File{realSkillFile})

		errs := v.ValidateTriggerMaps(root)
		require.Len(t, errs, 1)
		assert.Contains(t, errs[0].Message, "ghost-skill")
		assert.Equal(t, "error", errs[0].Severity)
	})

	t.Run("real skill passes", func(t *testing.T) {
		root := t.TempDir()

		refDir := filepath.Join(root, "skills", "agent-routing", "references")
		require.NoError(t, os.MkdirAll(refDir, 0o755))

		triggerContent := `| Trigger | Skill |
|---------|-------|
| test | code-testing-qa |
`
		require.NoError(t, os.WriteFile(filepath.Join(refDir, "triggers.md"), []byte(triggerContent), 0o644))

		realSkillFile := discovery.File{
			RelPath:  "skills/code-testing-qa/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "",
		}
		v := NewCrossFileValidator([]discovery.File{realSkillFile})

		errs := v.ValidateTriggerMaps(root)
		assert.Empty(t, errs)
	})

	t.Run("ghost agent detected", func(t *testing.T) {
		root := t.TempDir()

		refDir := filepath.Join(root, "skills", "dispatch", "references")
		require.NoError(t, os.MkdirAll(refDir, 0o755))

		triggerContent := `| Trigger | Agent |
|---------|-------|
| deploy | Task(ghost-agent) |
`
		require.NoError(t, os.WriteFile(filepath.Join(refDir, "routing.md"), []byte(triggerContent), 0o644))

		v := NewCrossFileValidator([]discovery.File{}) // no agents registered

		errs := v.ValidateTriggerMaps(root)
		require.Len(t, errs, 1)
		assert.Contains(t, errs[0].Message, "ghost-agent")
	})

	t.Run("built-in agent types are not flagged", func(t *testing.T) {
		root := t.TempDir()

		refDir := filepath.Join(root, "skills", "dispatch", "references")
		require.NoError(t, os.MkdirAll(refDir, 0o755))

		// Explore, haiku, sonnet are built-in types
		triggerContent := `| Trigger | Agent |
|---------|-------|
| explore | Task(Explore) |
| small task | Task(haiku) |
| coding | Task(sonnet) |
`
		require.NoError(t, os.WriteFile(filepath.Join(refDir, "builtins.md"), []byte(triggerContent), 0o644))

		v := NewCrossFileValidator([]discovery.File{})

		errs := v.ValidateTriggerMaps(root)
		assert.Empty(t, errs, "built-in agents should not be flagged as ghosts")
	})

	t.Run("non-trigger-map files are skipped", func(t *testing.T) {
		root := t.TempDir()

		refDir := filepath.Join(root, "skills", "my-skill", "references")
		require.NoError(t, os.MkdirAll(refDir, 0o755))

		// A reference file with a table but no Trigger column — should be skipped
		content := `| Name | Value |
|------|-------|
| missing-skill | something |
`
		require.NoError(t, os.WriteFile(filepath.Join(refDir, "config.md"), []byte(content), 0o644))

		v := NewCrossFileValidator([]discovery.File{})
		errs := v.ValidateTriggerMaps(root)
		assert.Empty(t, errs)
	})

	t.Run("no reference files returns empty", func(t *testing.T) {
		root := t.TempDir()
		v := NewCrossFileValidator([]discovery.File{})
		errs := v.ValidateTriggerMaps(root)
		assert.Empty(t, errs)
	})
}
