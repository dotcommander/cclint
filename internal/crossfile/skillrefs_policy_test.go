package crossfile

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

func TestSkillReferenceInventoryUsesPolicy(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		policy ReferencePolicy
		linked bool
		want   int
	}{
		{name: "included orphan", want: 1},
		{name: "excluded file", policy: ReferencePolicy{Exclude: []string{"skills/test-skill/references/**"}}},
		{name: "linked directory skipped", linked: true},
		{name: "linked directory followed", linked: true, policy: ReferencePolicy{FollowSymlinks: true}, want: 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			dir := filepath.Join(root, "skills/test-skill/references")
			if tc.linked {
				target := filepath.Join(root, "assets")
				if err := os.MkdirAll(target, 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Dir(dir), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(target, dir); err != nil {
					t.Fatal(err)
				}
			} else if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "unused.md"), []byte("Reference"), 0644); err != nil {
				t.Fatal(err)
			}
			v := NewCrossFileValidator([]discovery.File{{
				Type: discovery.FileTypeSkill, RelPath: "skills/test-skill/SKILL.md",
			}}).WithReferencePolicy(tc.policy)
			if got := v.ValidateSkillReferences(root); len(got) != tc.want {
				t.Fatalf("inventory diagnostics = %v; want %d", got, tc.want)
			}
			// Mentioned but excluded/skipped references must not become phantom errors.
			v.skills["test-skill"] = discovery.File{RelPath: "skills/test-skill/SKILL.md", Contents: "Read references/unused.md"}
			if got := v.ValidateSkillReferences(root); len(got) != 0 {
				t.Fatalf("ineligible reference became a phantom: %v", got)
			}
		})
	}
}
