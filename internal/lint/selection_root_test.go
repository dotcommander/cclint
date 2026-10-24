package lint

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestDirectorySelectionPreservesProjectRoot(t *testing.T) {
	t.Parallel()
	root := selectionTestRoot(t)
	want := []string{
		".claude-plugin/plugin.json",
		".claude/agents/hidden-agent.md",
		".claude/commands/hidden-command.md",
		".claude/settings.json",
		"agents/visible-agent.md",
		"commands/visible-command.md",
		"skills/example/SKILL.md",
	}
	for _, path := range want {
		writeSelectionFile(t, root, path, "fixture\n")
	}
	writeSelectionFile(t, root, ".cache/agents/ignored.md", "fixture\n")
	writeSelectionFile(t, root, ".git/CLAUDE.md", "fixture\n")
	writeSelectionFile(t, root, "notes/readme.md", "fixture\n")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relativeRoot, err := filepath.Rel(cwd, root)
	if err != nil {
		t.Fatal(err)
	}
	slices.Sort(want)
	for _, tc := range []struct {
		name, path, explicitRoot string
	}{
		{"automatic absolute", root, ""},
		{"automatic relative", relativeRoot, ""},
		{"explicit absolute", ".", root},
		{"explicit relative", ".", relativeRoot},
		{"root as positional argument", root, root},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			files, err := expandDirectoriesAtRoot([]string{tc.path}, tc.explicitRoot, "")
			if err != nil {
				t.Fatal(err)
			}
			got := make([]string, 0, len(files))
			for _, file := range files {
				relative, err := filepath.Rel(root, file.Path)
				if err != nil {
					t.Fatal(err)
				}
				got = append(got, filepath.ToSlash(relative))
			}
			slices.Sort(got)
			if !slices.Equal(got, want) {
				t.Fatalf("selected = %v, want %v", got, want)
			}
		})
	}
}

func TestDirectorySelectionRetainsHintsAndOverrides(t *testing.T) {
	t.Parallel()
	root := selectionTestRoot(t)
	command := writeSelectionFile(t, root, "nested/command/example.md", "fixture\n")
	writeSelectionFile(t, root, "skills/example/SKILL.md", "fixture\n")
	writeSelectionFile(t, root, "skills/example/references/help.md", "fixture\n")
	writeSelectionFile(t, root, "notes/readme.md", "fixture\n")
	writeSelectionFile(t, root, "notes/.cache/ignored.md", "fixture\n")
	for _, tc := range []struct {
		name, directory, override, hint string
	}{
		{"singular directory", "nested/command", "", "command"},
		{"skill filename guard", "skills", "", ""},
		{"forced type", "notes", "agent", "agent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			files, err := expandDirectoriesAtRoot([]string{tc.directory}, root, tc.override)
			if err != nil || len(files) != 1 {
				t.Fatalf("files = %v, err = %v; want one file", files, err)
			}
			if files[0].TypeHint != tc.hint {
				t.Fatalf("hint = %q, want %q", files[0].TypeHint, tc.hint)
			}
		})
	}
	files, err := expandDirectoriesAtRoot([]string{command}, root, "command")
	if err != nil || len(files) != 1 || files[0] != (fileWithHint{Path: command, TypeHint: "command"}) {
		t.Fatalf("non-directory pass-through = %v, %v", files, err)
	}
}

func TestLintFilesSelectedProjectRoot(t *testing.T) {
	t.Parallel()
	root := selectionTestRoot(t)
	writeSelectionFile(t, root, ".claude/agents/example.md", "---\nname: example\ndescription: Test agent\n---\n## Foundation\nTest.\n")
	writeSelectionFile(t, root, ".claude/commands/example.md", "---\ndescription: Test command\nallowed-tools: Task\n---\nTask(example): Test.\n")
	summary, err := LintFiles([]string{root}, "", "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalFiles != 2 || len(summary.Results) != 2 {
		t.Fatalf("selected %d files, %d results; want two", summary.TotalFiles, len(summary.Results))
	}
	for _, result := range summary.Results {
		if result.Type == "unknown" {
			t.Fatalf("selected file lost its type: %+v", result)
		}
	}
}

func selectionTestRoot(t *testing.T) string {
	t.Helper()
	root, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	return root
}

func writeSelectionFile(t *testing.T, root, relative, contents string) string {
	t.Helper()
	path := filepath.Join(root, filepath.FromSlash(relative))
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(contents), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
