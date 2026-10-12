package cmd

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/alecthomas/kong"
)

func TestFmtGroupedSelection(t *testing.T) {
	t.Parallel()

	root, files := writeFmtGroupedFixture(t, "# Example\n")
	tests := []struct {
		name string
		cmd  fmtCommand
		want []string
	}{
		{"single selector", fmtCommand{Args: []string{"agents"}}, []string{files["agent"]}},
		{"two selectors", fmtCommand{Args: []string{"agents", "skills"}}, []string{files["agent"], files["skill"]}},
		{"reversed selectors", fmtCommand{Args: []string{"skills", "agents"}}, []string{files["agent"], files["skill"]}},
		{"all selectors", fmtCommand{Args: []string{"agents", "commands", "skills"}}, []string{files["agent"], files["command"], files["skill"]}},
		{"duplicate aliases", fmtCommand{Args: []string{"agents", "agent", "skills", "skill"}}, []string{files["agent"], files["skill"]}},
		{"repeated selector", fmtCommand{Args: []string{"agents", "skills", "agents"}}, []string{files["agent"], files["skill"]}},
		{"positional paths take precedence", fmtCommand{Args: []string{"agents", "skills", files["command"]}}, []string{files["command"]}},
		{"file flag takes precedence", fmtCommand{Files: []string{files["skill"]}, Args: []string{"agents", "skills"}}, []string{files["skill"]}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := collectFilesToFormat(&tt.cmd, root, root, nil)
			if err != nil {
				t.Fatal(err)
			}
			if len(got) != len(tt.want) {
				t.Fatalf("selected %v, want %v", got, tt.want)
			}
			for _, path := range tt.want {
				if !slices.Contains(got, path) {
					t.Fatalf("selected %v, missing %s", got, path)
				}
			}
		})
	}
}

func TestFmtGroupedCLIWrite(t *testing.T) {
	t.Parallel()

	root, files := writeFmtGroupedFixture(t, "# Example  \n")
	var app cli
	parser, err := kong.New(&app, kong.Name("cclint"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := parser.Parse([]string{"--root", root, "fmt", "--write", "--quiet", "agents", "skills"}); err != nil {
		t.Fatal(err)
	}
	if err := runFmt(app.executionOptions(), &app.Fmt); err != nil {
		t.Fatal(err)
	}
	for component, path := range files {
		want := "# Example\n"
		if component == "command" {
			want = "# Example  \n"
		}
		got, err := os.ReadFile(path)
		if err != nil || string(got) != want {
			t.Errorf("%s bytes = %q, %v; want %q", component, got, err, want)
		}
	}
}

func writeFmtGroupedFixture(t *testing.T, content string) (string, map[string]string) {
	t.Helper()

	root := t.TempDir()
	files := map[string]string{
		"agent":   filepath.Join(root, ".claude", "agents", "example.md"),
		"command": filepath.Join(root, ".claude", "commands", "example.md"),
		"skill":   filepath.Join(root, ".claude", "skills", "example", "SKILL.md"),
	}
	for _, path := range files {
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return root, files
}
