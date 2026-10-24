package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"testing"
)

func TestChangedUntrackedPathsPreserveFilenameBytes(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("requires real Git repositories")
	}
	for _, committed := range []bool{false, true} {
		name := "unborn"
		if committed {
			name = "committed"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			run := func(args ...string) {
				t.Helper()
				cmd := exec.CommandContext(t.Context(), "git", args...)
				cmd.Dir = root
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("git %q: %v: %s", args, err, out)
				}
			}
			run("init", "--quiet")
			if committed {
				if err := os.WriteFile(filepath.Join(root, "seed.txt"), nil, 0o644); err != nil {
					t.Fatal(err)
				}
				run("add", "--", "seed.txt")
				run("-c", "user.name=Test", "-c", "user.email=test@example.invalid", "-c", "commit.gpgsign=false", "commit", "--quiet", "-m", "fixture")
			}
			canonical, err := filepath.EvalSymlinks(root)
			if err != nil {
				t.Fatal(err)
			}
			accepted := []string{
				"agents/ leading.md", "agents/trailing .md", "agents/trailing /file.md",
				"agents/tab\tname.md", "agents/line\nbreak.md", "agents/quote\"name.md",
				"agents/日本語.md", "agents/-leading.md", "agents/data.json",
			}
			excluded := []string{"agents/config.yaml", "agents/config.yml", "agents/notes.txt", "README.md"}
			var want []string
			for _, rel := range append(slices.Clone(accepted), excluded...) {
				path := filepath.Join(root, rel)
				if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, nil, 0o644); err != nil {
					t.Fatal(err)
				}
			}
			for _, rel := range accepted {
				want = append(want, filepath.Join(canonical, rel))
			}
			changes, err := GetChangedFileChanges(root)
			if err != nil {
				t.Fatal(err)
			}
			slices.Sort(want)
			slices.Sort(changes.Files)
			if !slices.Equal(changes.Files, want) {
				t.Errorf("files = %q, want %q", changes.Files, want)
			}
			if len(changes.DeletedFiles) != 0 {
				t.Errorf("untracked files produced deletions: %q", changes.DeletedFiles)
			}
		})
	}
}
