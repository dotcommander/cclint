package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
)

func writeFailClosedConfig(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0600); err != nil {
		t.Fatal(err)
	}
}

func writeFailClosedConfigJSON(t *testing.T, path string, values map[string]any) {
	t.Helper()
	data, err := json.Marshal(values)
	if err != nil {
		t.Fatal(err)
	}
	writeFailClosedConfig(t, path, string(data))
}

func TestLoadConfigRejectsMalformedPreferredSibling(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	selected := filepath.Join(root, ".cclintrc.json")
	writeFailClosedConfig(t, selected, "{")
	writeFailClosedConfig(t, filepath.Join(root, ".cclintrc.yaml"), "format: markdown\n")

	config, err := LoadConfig(root)
	if config != nil || err == nil || !strings.Contains(err.Error(), selected) {
		t.Fatalf("LoadConfig() = (%v, %v), want selected-file error", config, err)
	}
	if errors.Is(err, os.ErrNotExist) {
		t.Fatalf("parse error incorrectly classified as absence: %v", err)
	}
}

func TestLoadConfigRejectsMalformedNearestAncestor(t *testing.T) {
	// Chdir changes process-global state; this test must run serially.
	root := t.TempDir()
	nearest := filepath.Join(root, "nearest")
	child := filepath.Join(nearest, "child")
	if err := os.MkdirAll(child, 0700); err != nil {
		t.Fatal(err)
	}
	writeFailClosedConfigJSON(t, filepath.Join(root, ".cclintrc.json"), map[string]any{"format": "json"})
	selected := filepath.Join(nearest, ".cclintrc.yaml")
	writeFailClosedConfig(t, selected, "format: [\n")
	oldWD, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(child); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldWD) })

	config, err := LoadConfig("")
	if config != nil || err == nil || !strings.Contains(err.Error(), selected) {
		t.Fatalf("LoadConfig() = (%v, %v), want nearest-ancestor error", config, err)
	}
}

func TestLoadConfigRejectsUnreadableSelectedDirectory(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	selected := filepath.Join(root, ".cclintrc.json")
	if err := os.Mkdir(selected, 0700); err != nil {
		t.Fatal(err)
	}
	writeFailClosedConfig(t, filepath.Join(root, ".cclintrc.yaml"), "format: markdown\n")

	config, err := LoadConfig(root)
	if config != nil || err == nil || !strings.Contains(err.Error(), selected) {
		t.Fatalf("LoadConfig() = (%v, %v), want directory read error", config, err)
	}
	var pathErr *os.PathError
	if !errors.As(err, &pathErr) {
		t.Fatalf("read error must wrap filesystem cause: %v", err)
	}
}

func TestLoadConfigRejectsNonAbsenceStatError(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	selected := filepath.Join(root, ".cclintrc.json")
	if err := os.Symlink(selected, selected); err != nil {
		t.Skipf("self-referential symlink unavailable: %v", err)
	}
	writeFailClosedConfig(t, filepath.Join(root, ".cclintrc.yaml"), "format: markdown\n")

	config, err := LoadConfig(root)
	if config != nil || err == nil || !strings.Contains(err.Error(), selected) || !errors.Is(err, syscall.ELOOP) {
		t.Fatalf("LoadConfig() = (%v, %v), want path-bearing wrapped ELOOP", config, err)
	}
}

func TestLoadConfigExplicitRootDoesNotSearchParent(t *testing.T) {
	t.Parallel()
	parent := t.TempDir()
	root := filepath.Join(parent, "child")
	if err := os.Mkdir(root, 0700); err != nil {
		t.Fatal(err)
	}
	writeFailClosedConfigJSON(t, filepath.Join(parent, ".cclintrc.json"), map[string]any{"format": "markdown", "quiet": true})

	config, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if !config.RootExplicit || config.Root != root || config.Format != formatConsole || config.Quiet {
		t.Fatalf("explicit root unexpectedly used parent configuration: %+v", config)
	}
}

func TestLoadConfigSelectedConfigAndDefaults(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	config, err := LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if config.Format != formatConsole || config.Concurrency != 10 || !config.Schemas.Enabled {
		t.Fatalf("missing config did not retain defaults: %+v", config)
	}

	writeFailClosedConfig(t, filepath.Join(root, ".cclintrc.yml"), "format: markdown\nquiet: true\n")
	config, err = LoadConfig(root)
	if err != nil {
		t.Fatal(err)
	}
	if !config.RootExplicit || config.Root != root || config.Format != formatMarkdown || !config.Quiet || config.Concurrency != 10 {
		t.Fatalf("selected config did not retain overrides and defaults: %+v", config)
	}
}
