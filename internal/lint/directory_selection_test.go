package lint

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLintFilesRootRelativeDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, ".claude", "agents")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "helper.md")
	content := "---\nname: helper\ndescription: Review source files.\nmodel: sonnet\n---\n\n## Foundation\nRead source.\n\n## Workflow\n1. Explain findings.\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relativeRoot, err := filepath.Rel(cwd, root)
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name string
		root string
		path string
	}{
		{"root-relative directory", root, ".claude/agents/"},
		{"relative root directory", relativeRoot, ".claude/agents/"},
		{"absolute directory", root, directory},
		{"absolute directory without root", "", directory},
		{"cwd-relative directory without root", "", filepath.Join(relativeRoot, ".claude", "agents")},
		{"root-relative file", root, ".claude/agents/helper.md"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			summary, err := LintFiles([]string{tt.path}, tt.root, "", true, false)
			if err != nil {
				t.Fatal(err)
			}
			if summary.TotalFiles != 1 || summary.TotalErrors != 0 || len(summary.Results) != 1 {
				t.Fatalf("expected one valid selected file; got %+v", summary)
			}
			result := summary.Results[0]
			if result.Type != "agent" || !result.Success {
				t.Fatalf("expected successful agent result; got %+v", result)
			}
			if filepath.Base(result.File) != "helper.md" {
				t.Fatalf("selected %q, want helper.md", result.File)
			}
		})
	}
}

func TestDirectorySelectionPreservesNonDirectoryInputs(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	inputs := []string{".claude/agents/missing.md", "missing-directory/"}
	got, err := expandDirectoriesAtRoot(inputs, root, "")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != len(inputs) {
		t.Fatalf("got %d paths, want %d", len(got), len(inputs))
	}
	for i, path := range inputs {
		if got[i].Path != path {
			t.Errorf("path = %q, want original %q", got[i].Path, path)
		}
	}
}
