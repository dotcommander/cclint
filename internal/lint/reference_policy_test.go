package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSkillReferencePolicyWiring(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name      string
		exclude   []string
		follow    bool
		wantGhost bool
	}{
		{name: "default skips linked references"},
		{name: "follow includes in-root references", follow: true, wantGhost: true},
		{name: "exclude skill subtree", follow: true, exclude: []string{"skills/routing/**"}},
		{name: "exclude reference path", follow: true, exclude: []string{"skills/routing/references/routes.md"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for path, content := range map[string]string{
				"skills/included/SKILL.md": "Included skill documentation",
				"assets/routes.md":         "| Trigger | Target |\n|---|---|\n| review | missing-skill |\n",
			} {
				full := filepath.Join(root, path)
				if err := os.MkdirAll(filepath.Dir(full), 0755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(full, []byte(content), 0644); err != nil {
					t.Fatal(err)
				}
			}
			ref := filepath.Join(root, "skills/routing/references/routes.md")
			if err := os.MkdirAll(filepath.Dir(ref), 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(filepath.Join(root, "assets/routes.md"), ref); err != nil {
				t.Fatal(err)
			}
			ctx, err := NewLinterContextWithOptions(root, LinterOptions{
				Exclude: tc.exclude, FollowSymlinks: tc.follow,
				SkipSchemaValidation: true, NoCycleCheck: true,
			})
			if err != nil {
				t.Fatal(err)
			}
			summary := &LintSummary{}
			NewSkillLinter(root).PostProcessBatch(ctx, summary)
			ghost := false
			for _, result := range summary.Results {
				if result.File == "skills/routing/references/routes.md" && len(result.Errors) > 0 {
					ghost = true
				}
			}
			if ghost != tc.wantGhost {
				t.Fatalf("ghost = %v, want %v; results: %+v", ghost, tc.wantGhost, summary.Results)
			}
		})
	}
}
