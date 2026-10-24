package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
)

func TestReflectTraversalPolicy(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("workspace integration fixture")
	}
	for _, tt := range []struct {
		name    string
		exclude []string
		want    int
	}{
		{"selected_entries", []string{"kb/file-skip.md", "kb/excluded/**"}, 3},
		{"whole_kb", []string{"kb/**"}, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			for _, rel := range []string{
				"kb/kept-entry.md",
				"kb/file-skip.md",
				"kb/excluded/nested/bad.md",
				"kb/.work/scratch.md",
				"kb/README.md",
			} {
				path := filepath.Join(root, filepath.FromSlash(rel))
				if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(path, []byte("plain\n"), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			outside := t.TempDir()
			target := filepath.Join(outside, "target.md")
			if err := os.WriteFile(target, []byte("plain\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			for _, link := range []struct{ name, target string }{
				{"internal.md", filepath.Join(root, "kb", "README.md")},
				{"external.md", target},
				{"dangling.md", filepath.Join(outside, "missing.md")},
				{"directory.md", outside},
			} {
				if err := os.Symlink(link.target, filepath.Join(root, "kb", link.name)); err != nil {
					t.Fatal(err)
				}
			}
			cfg := &config.Config{Root: root, Quiet: true, Exclude: tt.exclude, FollowSymlinks: true}
			result, err := NewOrchestrator(cfg, OrchestratorConfig{}).WithLinters(nil).Run()
			if err != nil {
				t.Fatal(err)
			}
			var issues []cue.ValidationError
			for _, summary := range result.Summaries {
				issues = append(issues, CollectAllIssues(summary)...)
			}
			if len(issues) != tt.want || result.TotalFiles != 0 || result.TotalErrors != 0 {
				t.Fatalf("issues=%+v, result=%+v; want %d diagnostic-only issues", issues, result, tt.want)
			}
			for _, issue := range issues {
				if filepath.ToSlash(issue.File) != "kb/kept-entry.md" || issue.Source != cue.SourceCClintObserve {
					t.Errorf("unexpected diagnostic: %+v", issue)
				}
			}
			if tt.want != 0 && (result.TotalWarnings != 1 || result.TotalSuggestions != 2) {
				t.Fatalf("admitted-entry severities changed: %+v", result)
			}
		})
	}
}
