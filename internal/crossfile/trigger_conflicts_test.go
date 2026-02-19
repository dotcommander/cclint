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
// ParseTriggerMappings
// ---------------------------------------------------------------------------

func TestParseTriggerMappings(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     []TriggerMapping
	}{
		{
			name: "pairs keywords with skill targets",
			contents: `| Trigger | Skill |
|---------|-------|
| test coverage | code-testing-qa |
| refactor | code-clean-code |`,
			want: []TriggerMapping{
				{File: "test.md", Keyword: "test coverage", Target: "code-testing-qa", RefType: "skill"},
				{File: "test.md", Keyword: "refactor", Target: "code-clean-code", RefType: "skill"},
			},
		},
		{
			name: "pairs keywords with agent Task() targets",
			contents: `| Trigger | Agent |
|---------|-------|
| run tests | Task(quality-agent) |
| commit | Task(commit-agent) |`,
			want: []TriggerMapping{
				{File: "test.md", Keyword: "run tests", Target: "quality-agent", RefType: "agent"},
				{File: "test.md", Keyword: "commit", Target: "commit-agent", RefType: "agent"},
			},
		},
		{
			name: "normalizes keyword to lowercase",
			contents: `| Trigger | Skill |
|---------|-------|
| Cache | code-caching |`,
			want: []TriggerMapping{
				{File: "test.md", Keyword: "cache", Target: "code-caching", RefType: "skill"},
			},
		},
		{
			name: "no trigger table returns empty",
			contents: `| Name | Value |
|------|-------|
| foo  | bar   |`,
			want: nil,
		},
		{
			name:     "empty contents returns empty",
			contents: "",
			want:     nil,
		},
		{
			name: "ignores prose words without hyphens",
			contents: `| Trigger | Description |
|---------|-------------|
| foo bar | Run the task now |`,
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseTriggerMappings("test.md", tt.contents)
			if tt.want == nil {
				assert.Empty(t, got)
			} else {
				require.Equal(t, tt.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// DetectTriggerConflicts
// ---------------------------------------------------------------------------

func TestDetectTriggerConflicts(t *testing.T) {
	t.Run("same keyword different targets emits warning", func(t *testing.T) {
		root := t.TempDir()

		// File 1: "test" → code-testing-qa
		refDir1 := filepath.Join(root, "skills", "routing-a", "references")
		require.NoError(t, os.MkdirAll(refDir1, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refDir1, "triggers.md"), []byte(`| Trigger | Skill |
|---------|-------|
| test | code-testing-qa |
`), 0o644))

		// File 2: "test" → code-testing-advanced (different target!)
		refDir2 := filepath.Join(root, "skills", "routing-b", "references")
		require.NoError(t, os.MkdirAll(refDir2, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refDir2, "triggers.md"), []byte(`| Trigger | Skill |
|---------|-------|
| test | code-testing-advanced |
`), 0o644))

		v := NewCrossFileValidator([]discovery.File{})
		errs := v.DetectTriggerConflicts(root)

		require.Len(t, errs, 1)
		assert.Equal(t, "warning", errs[0].Severity)
		assert.Contains(t, errs[0].Message, "test")
		assert.Contains(t, errs[0].Message, "code-testing-qa")
		assert.Contains(t, errs[0].Message, "code-testing-advanced")
	})

	t.Run("same keyword same target in different files is not a conflict", func(t *testing.T) {
		root := t.TempDir()

		// Both files route "test" → code-testing-qa
		for _, skill := range []string{"routing-a", "routing-b"} {
			refDir := filepath.Join(root, "skills", skill, "references")
			require.NoError(t, os.MkdirAll(refDir, 0o755))
			require.NoError(t, os.WriteFile(filepath.Join(refDir, "triggers.md"), []byte(`| Trigger | Skill |
|---------|-------|
| test | code-testing-qa |
`), 0o644))
		}

		v := NewCrossFileValidator([]discovery.File{})
		errs := v.DetectTriggerConflicts(root)
		assert.Empty(t, errs)
	})

	t.Run("different keywords different targets no conflict", func(t *testing.T) {
		root := t.TempDir()

		refDir := filepath.Join(root, "skills", "routing", "references")
		require.NoError(t, os.MkdirAll(refDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refDir, "triggers.md"), []byte(`| Trigger | Skill |
|---------|-------|
| test | code-testing-qa |
| refactor | code-clean-code |
`), 0o644))

		v := NewCrossFileValidator([]discovery.File{})
		errs := v.DetectTriggerConflicts(root)
		assert.Empty(t, errs)
	})

	t.Run("no trigger tables returns empty", func(t *testing.T) {
		root := t.TempDir()

		refDir := filepath.Join(root, "skills", "my-skill", "references")
		require.NoError(t, os.MkdirAll(refDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refDir, "config.md"), []byte(`| Name | Value |
|------|-------|
| foo  | bar   |
`), 0o644))

		v := NewCrossFileValidator([]discovery.File{})
		errs := v.DetectTriggerConflicts(root)
		assert.Empty(t, errs)
	})

	t.Run("no reference files returns empty", func(t *testing.T) {
		root := t.TempDir()
		v := NewCrossFileValidator([]discovery.File{})
		errs := v.DetectTriggerConflicts(root)
		assert.Empty(t, errs)
	})

	t.Run("single trigger map with unique keywords no conflict", func(t *testing.T) {
		root := t.TempDir()

		refDir := filepath.Join(root, "skills", "routing", "references")
		require.NoError(t, os.MkdirAll(refDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refDir, "triggers.md"), []byte(`| Trigger | Skill |
|---------|-------|
| test | code-testing-qa |
| lint | code-clean-code |
| profile | code-perf-optimization |
`), 0o644))

		v := NewCrossFileValidator([]discovery.File{})
		errs := v.DetectTriggerConflicts(root)
		assert.Empty(t, errs)
	})

	t.Run("case insensitive keyword matching detects conflict", func(t *testing.T) {
		root := t.TempDir()

		// File 1: "Cache" → code-cache-a
		refDir1 := filepath.Join(root, "skills", "routing-a", "references")
		require.NoError(t, os.MkdirAll(refDir1, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refDir1, "triggers.md"), []byte(`| Trigger | Skill |
|---------|-------|
| Cache | code-cache-a |
`), 0o644))

		// File 2: "cache" → code-cache-b (different target, same keyword after lowercase)
		refDir2 := filepath.Join(root, "skills", "routing-b", "references")
		require.NoError(t, os.MkdirAll(refDir2, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refDir2, "triggers.md"), []byte(`| Trigger | Skill |
|---------|-------|
| cache | code-cache-b |
`), 0o644))

		v := NewCrossFileValidator([]discovery.File{})
		errs := v.DetectTriggerConflicts(root)

		require.Len(t, errs, 1)
		assert.Equal(t, "warning", errs[0].Severity)
		assert.Contains(t, errs[0].Message, "cache")
		assert.Contains(t, errs[0].Message, "code-cache-a")
		assert.Contains(t, errs[0].Message, "code-cache-b")
	})
}
