package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/alecthomas/kong"
)

func TestFmtExplicitRootSelections(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, ".claude", "skills", "example")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "SKILL.md")
	if err := os.WriteFile(path, []byte("# Example\n"), 0o600); err != nil {
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
	relativePath := filepath.Join(".claude", "skills", "example", "SKILL.md")

	tests := []struct {
		name string
		root string
		cmd  fmtCommand
	}{
		{"positional file", root, fmtCommand{Args: []string{relativePath}}},
		{"positional directory", root, fmtCommand{Args: []string{".claude/skills/"}}},
		{"explicit file", root, fmtCommand{Files: []string{relativePath}}},
		{"explicit file precedence", root, fmtCommand{Files: []string{relativePath}, Args: []string{"nonexistent"}}},
		{"relative root positional", relativeRoot, fmtCommand{Args: []string{relativePath}}},
		{"relative root explicit", relativeRoot, fmtCommand{Files: []string{relativePath}}},
		{"absolute file", root, fmtCommand{Args: []string{path}}},
		{"absolute directory", root, fmtCommand{Args: []string{directory}}},
		{"component discovery", root, fmtCommand{Args: []string{"skills"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := collectFilesToFormat(&tt.cmd, tt.root, tt.root, nil)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, []string{path}) {
				t.Fatalf("selected %v, want [%s]", got, path)
			}
		})
	}
}

func TestFmtSelectionRootVersusCwd(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	// A real cwd entry with the same name must not override an explicit root.
	path := filepath.Join(root, "fmt.go")
	if err := os.WriteFile(path, []byte("fixture"), 0o600); err != nil {
		t.Fatal(err)
	}
	for _, explicit := range []bool{false, true} {
		name := "implicit root"
		selectionRoot := ""
		want := "fmt.go"
		if explicit {
			name = "explicit root"
			selectionRoot = root
			want = path
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got, err := collectFilesToFormat(&fmtCommand{Args: []string{"fmt.go"}}, root, selectionRoot, nil)
			if err != nil || !reflect.DeepEqual(got, []string{want}) {
				t.Fatalf("selected %v, %v; want [%s]", got, err, want)
			}
		})
	}
}

func TestFmtCLIFileSelection(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	relativeFile := filepath.Join(".claude", "skills", "example", "SKILL.md")
	absoluteFile := filepath.Join(t.TempDir(), "outside.md")
	tests := []struct {
		name string
		args []string
		want string
	}{
		{"explicit root", []string{"--root", root, "fmt", "--file", relativeFile}, filepath.Join(root, relativeFile)},
		{"no explicit root", []string{"fmt", "--file", relativeFile}, kong.ExpandPath(relativeFile)},
		{"absolute file", []string{"--root", root, "fmt", "--file", absoluteFile}, absoluteFile},
		{"home relative file", []string{"--root", root, "fmt", "--file", "~/fmt-selection.md"}, kong.ExpandPath("~/fmt-selection.md")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var app cli
			parser, err := kong.New(&app, kong.Name("cclint"))
			if err != nil {
				t.Fatal(err)
			}
			if _, err := parser.Parse(tt.args); err != nil {
				t.Fatal(err)
			}
			got, err := collectFilesToFormat(&app.Fmt, root, stringValue(app.Root), nil)
			if err != nil || !reflect.DeepEqual(got, []string{tt.want}) {
				t.Fatalf("parsed selection = %v, %v; want [%s]", got, err, tt.want)
			}
		})
	}
}

func TestFmtRelativeConfiguredRoot(t *testing.T) {
	t.Parallel()

	for _, fileFlag := range []bool{false, true} {
		name := "positional"
		if fileFlag {
			name = "file flag"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			root := t.TempDir()
			relativeFile := filepath.Join(".claude", "skills", "example", "SKILL.md")
			path := filepath.Join(root, relativeFile)
			if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, []byte("# Example  \n"), 0o600); err != nil {
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
			cmd := fmtCommand{Write: true}
			if fileFlag {
				cmd.Files = []string{relativeFile}
			} else {
				cmd.Args = []string{relativeFile}
			}
			quiet := true
			if err := runFmt(executionOptions{root: &relativeRoot, quiet: &quiet}, &cmd); err != nil {
				t.Fatal(err)
			}
			got, err := os.ReadFile(path)
			if err != nil || string(got) != "# Example\n" {
				t.Fatalf("formatted bytes = %q, %v; want canonical body", got, err)
			}
		})
	}
}

func TestFmtExplicitRootMissingFileSelection(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	want := filepath.Join(root, "missing.md")
	got, err := collectFilesToFormat(&fmtCommand{Files: []string{"missing.md"}}, root, root, nil)
	if err != nil || !reflect.DeepEqual(got, []string{want}) {
		t.Fatalf("explicit file selection = %v, %v; want [%s]", got, err, want)
	}
	// Positional selection still validates existence before returning paths.
	_, err = collectFilesToFormat(&fmtCommand{Args: []string{"missing.md"}}, root, root, nil)
	if err == nil {
		t.Fatal("missing positional file must be rejected")
	}
}

func TestResolvePathArgsCwdRelativeDirectory(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	directory := filepath.Join(root, ".claude", "agents")
	if err := os.MkdirAll(directory, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "helper.md"), []byte("---\nname: helper\ndescription: Review files.\n---\n\nBody.\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relativeDir, err := filepath.Rel(cwd, directory)
	if err != nil {
		t.Fatal(err)
	}

	files, err := resolvePathArgs([]string{relativeDir})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 1 || filepath.Base(files[0]) != "helper.md" {
		t.Fatalf("got %v, want one helper.md", files)
	}
}

// TestFmtDiscoveryHonorsExclude pins that discovery-based fmt selection
// applies .cclintrc exclude patterns like lint discovery; only explicit
// --file/arg selections bypass exclusion.
func TestFmtDiscoveryHonorsExclude(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	for _, rel := range []string{"commands/keep.md", "commands/skip.md"} {
		absPath := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(absPath, []byte("---\nname:   probe\ndescription: d\n---\nbody\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	files, err := collectFilesToFormat(&fmtCommand{}, root, root, []string{"commands/skip.md"})
	if err != nil {
		t.Fatalf("collectFilesToFormat() error = %v", err)
	}
	sawKeep, sawSkip := false, false
	for _, f := range files {
		if strings.HasSuffix(filepath.ToSlash(f), "commands/keep.md") {
			sawKeep = true
		}
		if strings.HasSuffix(filepath.ToSlash(f), "commands/skip.md") {
			sawSkip = true
		}
	}
	if !sawKeep || sawSkip {
		t.Errorf("exclude not honored: keep=%v skip=%v (files: %v)", sawKeep, sawSkip, files)
	}
}
