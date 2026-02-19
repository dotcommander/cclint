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
// extractMentionedRefs
// ---------------------------------------------------------------------------

func TestExtractMentionedRefs(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     []string
	}{
		{
			name:     "bare prose mention",
			contents: "See references/foo.md for details.",
			want:     []string{"foo.md"},
		},
		{
			name:     "markdown link format",
			contents: "See [foo](references/foo.md) for details.",
			want:     []string{"foo.md"},
		},
		{
			name:     "Read tool invocation",
			contents: "Read(references/foo.md) returns the content.",
			want:     []string{"foo.md"},
		},
		{
			name:     "multiple distinct files",
			contents: "See references/foo.md and references/bar.md also references/baz.md.",
			want:     []string{"bar.md", "baz.md", "foo.md"},
		},
		{
			name:     "duplicate mentions deduplicated",
			contents: "references/foo.md is mentioned here and again references/foo.md.",
			want:     []string{"foo.md"},
		},
		{
			name:     "bare references/ directory without filename ignored",
			contents: "Files live in the references/ directory.",
			want:     nil,
		},
		{
			name:     "no references at all",
			contents: "# Skill\n\nThis skill does nothing.",
			want:     nil,
		},
		{
			name:     "underscore and hyphen in filename",
			contents: "See references/my-long_name.md.",
			want:     []string{"my-long_name.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractMentionedRefs(tt.contents)
			if len(tt.want) == 0 {
				assert.Empty(t, got)
			} else {
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// ValidateSkillReferences integration tests
// ---------------------------------------------------------------------------

func TestValidateSkillReferences(t *testing.T) {
	t.Run("referenced file exists - no error", func(t *testing.T) {
		root := t.TempDir()
		skillDir := filepath.Join(root, "skills", "my-skill")
		refsDir := filepath.Join(skillDir, "references")
		require.NoError(t, os.MkdirAll(refsDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refsDir, "foo.md"), []byte("content"), 0o644))

		v := NewCrossFileValidator([]discovery.File{{
			RelPath:  "skills/my-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "See references/foo.md for details.",
		}})

		errs := v.ValidateSkillReferences(root)
		assert.Empty(t, errs)
	})

	t.Run("referenced file missing - phantom ref error", func(t *testing.T) {
		root := t.TempDir()
		skillDir := filepath.Join(root, "skills", "my-skill")
		require.NoError(t, os.MkdirAll(skillDir, 0o755))
		// No references/ directory or file created.

		v := NewCrossFileValidator([]discovery.File{{
			RelPath:  "skills/my-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "See references/missing.md for details.",
		}})

		errs := v.ValidateSkillReferences(root)
		require.Len(t, errs, 1)
		assert.Equal(t, "error", errs[0].Severity)
		assert.Contains(t, errs[0].Message, "references/missing.md")
		assert.Contains(t, errs[0].Message, "does not exist")
		assert.Equal(t, "skills/my-skill/SKILL.md", errs[0].File)
	})

	t.Run("file exists in references/ but not mentioned - orphaned ref info", func(t *testing.T) {
		root := t.TempDir()
		skillDir := filepath.Join(root, "skills", "my-skill")
		refsDir := filepath.Join(skillDir, "references")
		require.NoError(t, os.MkdirAll(refsDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refsDir, "orphan.md"), []byte("content"), 0o644))

		v := NewCrossFileValidator([]discovery.File{{
			RelPath:  "skills/my-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "# My Skill\n\nThis skill does nothing special.",
		}})

		errs := v.ValidateSkillReferences(root)
		require.Len(t, errs, 1)
		assert.Equal(t, "info", errs[0].Severity)
		assert.Contains(t, errs[0].Message, "references/orphan.md")
		assert.Contains(t, errs[0].Message, "not mentioned")
	})

	t.Run("skill has no references/ directory - no errors", func(t *testing.T) {
		root := t.TempDir()
		skillDir := filepath.Join(root, "skills", "no-refs-skill")
		require.NoError(t, os.MkdirAll(skillDir, 0o755))

		v := NewCrossFileValidator([]discovery.File{{
			RelPath:  "skills/no-refs-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "# Skill with no references directory.",
		}})

		errs := v.ValidateSkillReferences(root)
		assert.Empty(t, errs)
	})

	t.Run("multiple phantom and orphaned refs in same skill", func(t *testing.T) {
		root := t.TempDir()
		skillDir := filepath.Join(root, "skills", "multi-skill")
		refsDir := filepath.Join(skillDir, "references")
		require.NoError(t, os.MkdirAll(refsDir, 0o755))
		// Create existing file and orphan file.
		require.NoError(t, os.WriteFile(filepath.Join(refsDir, "exists.md"), []byte("content"), 0o644))
		require.NoError(t, os.WriteFile(filepath.Join(refsDir, "orphan.md"), []byte("content"), 0o644))

		v := NewCrossFileValidator([]discovery.File{{
			RelPath:  "skills/multi-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "See references/exists.md and references/phantom.md.",
		}})

		errs := v.ValidateSkillReferences(root)
		// Expect one phantom error and one orphan info.
		require.Len(t, errs, 2)

		var errCount, infoCount int
		for _, e := range errs {
			switch e.Severity {
			case "error":
				errCount++
				assert.Contains(t, e.Message, "phantom.md")
			case "info":
				infoCount++
				assert.Contains(t, e.Message, "orphan.md")
			}
		}
		assert.Equal(t, 1, errCount, "expected one phantom ref error")
		assert.Equal(t, 1, infoCount, "expected one orphaned ref info")
	})

	t.Run("markdown link format detected", func(t *testing.T) {
		root := t.TempDir()
		skillDir := filepath.Join(root, "skills", "link-skill")
		refsDir := filepath.Join(skillDir, "references")
		require.NoError(t, os.MkdirAll(refsDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refsDir, "guide.md"), []byte("content"), 0o644))

		v := NewCrossFileValidator([]discovery.File{{
			RelPath:  "skills/link-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "Read [the guide](references/guide.md) first.",
		}})

		errs := v.ValidateSkillReferences(root)
		assert.Empty(t, errs)
	})

	t.Run("prose format references/foo.md detected", func(t *testing.T) {
		root := t.TempDir()
		skillDir := filepath.Join(root, "skills", "prose-skill")
		refsDir := filepath.Join(skillDir, "references")
		require.NoError(t, os.MkdirAll(refsDir, 0o755))
		require.NoError(t, os.WriteFile(filepath.Join(refsDir, "patterns.md"), []byte("content"), 0o644))

		v := NewCrossFileValidator([]discovery.File{{
			RelPath:  "skills/prose-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "See references/patterns.md for all patterns.",
		}})

		errs := v.ValidateSkillReferences(root)
		assert.Empty(t, errs)
	})

	t.Run("empty validator - no skills - no errors", func(t *testing.T) {
		root := t.TempDir()
		v := NewCrossFileValidator([]discovery.File{})
		errs := v.ValidateSkillReferences(root)
		assert.Empty(t, errs)
	})
}
