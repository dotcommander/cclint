package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigNearestAncestorAndExplicitRoot(t *testing.T) {
	root := t.TempDir()
	child := filepath.Join(root, "src", "nested")
	if err := os.MkdirAll(child, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, ".cclintrc.json"), []byte(`{"quiet":true,"concurrency":17}`), 0644); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(child); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	cfg, err := LoadConfig("")
	if err != nil {
		t.Fatal(err)
	}
	if !cfg.Quiet || cfg.Concurrency != 17 {
		t.Fatalf("ancestor config not loaded: %+v", cfg)
	}
	cfg, err = LoadConfig(child)
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Quiet || cfg.Concurrency != 10 || cfg.Root != child || !cfg.RootExplicit {
		t.Fatalf("explicit lookup escaped root: %+v", cfg)
	}
}

func TestConfigEnvironmentAliases(t *testing.T) {
	t.Setenv("CCLINT_FAIL_ON", "warning")
	t.Setenv("CCLINT_FOLLOW_SYMLINKS", "true")
	t.Setenv("CCLINT_SHOW_SCORES", "true")
	t.Setenv("CCLINT_SHOW_IMPROVEMENTS", "true")
	t.Setenv("CCLINT_NO_CYCLE_CHECK", "true")
	t.Setenv("CCLINT_RULES_STRICT", "false")
	t.Setenv("CCLINT_SCHEMAS_ENABLED", "false")
	t.Setenv("CCLINT_OUTPUT", "env-report.json")
	cfg, err := LoadConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FailOn != "warning" || !cfg.FollowSymlinks || !cfg.ShowScores || !cfg.ShowImprovements || !cfg.NoCycleCheck || cfg.Rules.Strict || cfg.Schemas.Enabled || cfg.Output != "env-report.json" {
		t.Fatalf("aliases not applied: %+v", cfg)
	}
	t.Setenv("CCLINT_FAILON", "suggestion")
	cfg, err = LoadConfig(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if cfg.FailOn != "suggestion" {
		t.Fatalf("legacy alias precedence changed: %s", cfg.FailOn)
	}
}

func TestConfigRootExplicit(t *testing.T) {
	root := t.TempDir()
	cfg, err := LoadConfig(root)
	if err != nil || !cfg.RootExplicit {
		t.Fatalf("argument root not explicit: %+v %v", cfg, err)
	}
	t.Setenv("CCLINT_ROOT", root)
	cfg, err = LoadConfig("")
	if err != nil || !cfg.RootExplicit || cfg.Root != root {
		t.Fatalf("environment root not explicit: %+v %v", cfg, err)
	}
	t.Setenv("CCLINT_ROOT", "")
	if err := os.WriteFile(filepath.Join(root, ".cclintrc.json"), []byte(`{"root":"/configured/root"}`), 0644); err != nil {
		t.Fatal(err)
	}
	old, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(old) })
	cfg, err = LoadConfig("")
	if err != nil || !cfg.RootExplicit || cfg.Root != "/configured/root" {
		t.Fatalf("file root not explicit: %+v %v", cfg, err)
	}
}
