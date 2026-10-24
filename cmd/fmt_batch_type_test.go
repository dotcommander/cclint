package cmd

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const batchTypeComponentBody = "---\ndescription: Sample.\nname: sample\n---\n\nBody.\n"

// writeBatchTypeFixture creates one component file inside the isolated root.
func writeBatchTypeFixture(t *testing.T, root, relativePath, content string) string {
	t.Helper()
	path := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

// writeBatchTypeTree builds a mixed component tree in which every discovered
// file needs formatting, so any wrong selection rewrites unrelated bytes.
func writeBatchTypeTree(t *testing.T) (root, agentPath, commandPath, skillPath string) {
	t.Helper()
	root = t.TempDir()
	agentPath = writeBatchTypeFixture(t, root, filepath.Join(".claude", "agents", "reviewer.md"),
		"---\ndescription: Reviews changes.\nname: reviewer\n---\n\nReviews.\n")
	commandPath = writeBatchTypeFixture(t, root, filepath.Join(".claude", "commands", "ship.md"),
		"---\ndescription: Ships code.\nname: ship\n---\n\nShips.\n")
	skillPath = writeBatchTypeFixture(t, root, filepath.Join(".claude", "skills", "deploy", "SKILL.md"),
		"---\ndescription: Deploys code.\nname: deploy\n---\n\nDeploys.\n")
	return root, agentPath, commandPath, skillPath
}

// TestFmtBatchTypeFiltersMixedDiscovery pins that no-path --type selection
// discovers only the requested formatting component type from a mixed tree;
// discovered skills, commands, and rules must never leak into the batch.
func TestFmtBatchTypeFiltersMixedDiscovery(t *testing.T) {
	t.Parallel()

	root, agentPath, commandPath, _ := writeBatchTypeTree(t)

	tests := []struct {
		name     string
		typeFlag string
		want     string
	}{
		{"singular agent", "agent", agentPath},
		{"plural agent alias", "agents", agentPath},
		{"singular command", "command", commandPath},
		{"plural command alias", "commands", commandPath},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			files, err := collectFilesToFormat(&fmtCommand{Type: tt.typeFlag}, root, root, nil)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(files, []string{tt.want}) {
				t.Fatalf("--type %s selected %v, want [%s] only", tt.typeFlag, files, tt.want)
			}
		})
	}
}

// TestFmtBatchTypeRejectsContradictoryPositional pins alias-normalized
// agreement between positional component types and --type before formatting.
func TestFmtBatchTypeRejectsContradictoryPositional(t *testing.T) {
	t.Parallel()

	root, agentPath, _, _ := writeBatchTypeTree(t)

	t.Run("matching plural alias discovers the type", func(t *testing.T) {
		t.Parallel()

		files, err := collectFilesToFormat(&fmtCommand{Type: "agent", Args: []string{"agents"}}, root, root, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(files, []string{agentPath}) {
			t.Fatalf("matching alias selected %v, want [%s]", files, agentPath)
		}
	})

	t.Run("contradictory singular positional is rejected", func(t *testing.T) {
		t.Parallel()

		_, err := collectFilesToFormat(&fmtCommand{Type: "agent", Args: []string{"skills"}}, root, root, nil)
		if err == nil || !strings.Contains(err.Error(), "conflicts with --type agent") {
			t.Fatalf("error = %v, want a type conflict", err)
		}
	})

	t.Run("contradictory plural positional is rejected", func(t *testing.T) {
		t.Parallel()

		_, err := collectFilesToFormat(&fmtCommand{Type: "skill", Args: []string{"agents"}}, root, root, nil)
		if err == nil || !strings.Contains(err.Error(), "conflicts with --type skill") {
			t.Fatalf("error = %v, want a type conflict", err)
		}
	})

	t.Run("type without formatting capability is rejected", func(t *testing.T) {
		t.Parallel()

		_, err := collectFilesToFormat(&fmtCommand{Type: "output-style"}, root, root, nil)
		if err == nil || !strings.Contains(err.Error(), "unknown component type: output-style") {
			t.Fatalf("error = %v, want unknown component type", err)
		}
	})
}

// TestFmtBatchTypeHonorsExclude pins that --type discovery still applies the
// configured exclude patterns like untyped discovery.
func TestFmtBatchTypeHonorsExclude(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	keepPath := writeBatchTypeFixture(t, root, filepath.Join(".claude", "agents", "keep.md"), batchTypeComponentBody)
	writeBatchTypeFixture(t, root, filepath.Join(".claude", "agents", "skip.md"), batchTypeComponentBody)

	files, err := collectFilesToFormat(&fmtCommand{Type: "agent"}, root, root, []string{".claude/agents/skip.md"})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(files, []string{keepPath}) {
		t.Fatalf("--type discovery ignored exclude: selected %v, want [%s]", files, keepPath)
	}
}

// TestFmtBatchTypeExplicitSelectionsKeepForce pins that explicit --file and
// positional path selections bypass the --type discovery filter and keep the
// later per-file force, while an empty --type preserves the type union.
func TestFmtBatchTypeExplicitSelectionsKeepForce(t *testing.T) {
	t.Parallel()

	root, agentPath, _, skillPath := writeBatchTypeTree(t)
	relativeSkill := filepath.Join(".claude", "skills", "deploy", "SKILL.md")

	t.Run("explicit file is not filtered by type", func(t *testing.T) {
		t.Parallel()

		files, err := collectFilesToFormat(&fmtCommand{Files: []string{skillPath}, Type: "agent"}, root, root, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(files, []string{skillPath}) {
			t.Fatalf("explicit --file selection = %v, want [%s]", files, skillPath)
		}
	})

	t.Run("positional path keeps type force", func(t *testing.T) {
		t.Parallel()

		files, err := collectFilesToFormat(&fmtCommand{Args: []string{relativeSkill}, Type: "agent"}, root, root, nil)
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(files, []string{skillPath}) {
			t.Fatalf("positional path selection = %v, want [%s]", files, skillPath)
		}
	})

	t.Run("empty type keeps union discovery", func(t *testing.T) {
		t.Parallel()

		files, err := collectFilesToFormat(&fmtCommand{Args: []string{"agents", "skills"}}, root, root, nil)
		if err != nil {
			t.Fatal(err)
		}
		got := append([]string(nil), files...)
		sort.Strings(got)
		want := []string{agentPath, skillPath}
		sort.Strings(want)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("union discovery = %v, want %v", got, want)
		}
	})
}

// TestFmtBatchTypeWriteScopesToType runs the actual fmt --write flow on an
// isolated fixture: --type agents rewrites only the discovered agent file
// while unrelated discovered components keep their exact bytes, and a direct
// explicit --file selection still receives the --type force.
func TestFmtBatchTypeWriteScopesToType(t *testing.T) {
	t.Parallel()

	root, agentPath, commandPath, skillPath := writeBatchTypeTree(t)
	before := map[string]string{}
	for name, path := range map[string]string{"agent": agentPath, "command": commandPath, "skill": skillPath} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		before[name] = string(content)
	}

	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relativeRoot, err := filepath.Rel(cwd, root)
	if err != nil {
		t.Fatal(err)
	}
	quiet := true
	if err := runFmt(executionOptions{root: &relativeRoot, quiet: &quiet}, &fmtCommand{Type: "agents", Write: true}); err != nil {
		t.Fatal(err)
	}

	wantAgent := "---\nname: reviewer\ndescription: Reviews changes.\n---\n\nReviews.\n"
	agentAfter, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(agentAfter) != wantAgent {
		t.Fatalf("--type agents rewrote agent file to %q, want %q", agentAfter, wantAgent)
	}
	for name, path := range map[string]string{"command": commandPath, "skill": skillPath} {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(content) != before[name] {
			t.Fatalf("unrelated discovered %s bytes changed: %q", name, content)
		}
	}

	// Explicit --file selection is deliberate: --type still forces the
	// component type while bypassing the discovery filter.
	if err := runFmt(executionOptions{root: &relativeRoot, quiet: &quiet}, &fmtCommand{Files: []string{commandPath}, Type: "agent", Write: true}); err != nil {
		t.Fatal(err)
	}
	wantCommand := "---\nname: ship\ndescription: Ships code.\n---\n\nShips.\n"
	forced, err := os.ReadFile(commandPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(forced) != wantCommand {
		t.Fatalf("explicit --file force produced %q, want %q", forced, wantCommand)
	}
	agentFinal, err := os.ReadFile(agentPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(agentFinal) != wantAgent {
		t.Fatalf("second run changed unrelated agent bytes: %q", agentFinal)
	}
}
