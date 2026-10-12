package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunFmtRejectsMalformedInput(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		name    string
		content string
		message string
	}{
		{"invalid YAML", "---\nname: [example\ndescription: Test skill\n---\n\n# Example\n", "yaml:"},
		{"unclosed frontmatter", "---\nname: example\ndescription: Test skill\n", "unclosed frontmatter"},
		{"duplicate key", "---\nname: example\nname: other\n---\n\n# Example\n", "already defined"},
	}
	modes := []struct {
		name string
		cmd  fmtCommand
	}{
		{"dry-run", fmtCommand{}},
		{"check", fmtCommand{Check: true}},
		{"diff", fmtCommand{Diff: true}},
		{"write", fmtCommand{Write: true}},
	}
	for _, input := range inputs {
		t.Run(input.name, func(t *testing.T) {
			t.Parallel()

			for _, mode := range modes {
				t.Run(mode.name, func(t *testing.T) {
					t.Parallel()

					for _, quiet := range []bool{false, true} {
						name := "normal"
						if quiet {
							name = "quiet"
						}
						t.Run(name, func(t *testing.T) {
							t.Parallel()

							root := t.TempDir()
							directory := filepath.Join(root, ".claude", "skills", "example")
							if err := os.MkdirAll(directory, 0o755); err != nil {
								t.Fatal(err)
							}
							path := filepath.Join(directory, "SKILL.md")
							if err := os.WriteFile(path, []byte(input.content), 0o600); err != nil {
								t.Fatal(err)
							}
							command := mode.cmd
							command.Files = []string{path}
							err := runFmt(executionOptions{root: &root, quiet: &quiet}, &command)
							if err == nil || !strings.Contains(err.Error(), input.message) || !strings.Contains(err.Error(), path) {
								t.Fatalf("runFmt error = %v, want path and %q", err, input.message)
							}
							after, err := os.ReadFile(path)
							if err != nil {
								t.Fatal(err)
							}
							if string(after) != input.content {
								t.Fatalf("failed formatting changed input: %q", after)
							}
							assertNoFormatterTemps(t, directory)
						})
					}
				})
			}
		})
	}
}

func TestRunFmtRejectsMissingExplicitFile(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	path := filepath.Join(root, ".claude", "agents", "missing.md")
	quiet := true
	err := runFmt(executionOptions{root: &root, quiet: &quiet}, &fmtCommand{
		Check: true,
		Files: []string{path},
	})
	if err == nil || !strings.Contains(err.Error(), path) {
		t.Fatalf("runFmt error = %v, want missing path %q", err, path)
	}
}
