package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/config"
)

func ptr[T any](value T) *T { return &value }

func TestApplyCLIOverridesSetsExplicitValues(t *testing.T) {
	oldVersion := Version
	t.Cleanup(func() { Version = oldVersion })
	Version = "v1.2.3-test"

	cfg := &config.Config{}
	applyCLIOverrides(cfg, executionOptions{
		root:         ptr("/override/root"),
		quiet:        ptr(true),
		verbose:      ptr(true),
		scores:       ptr(true),
		improvements: ptr(true),
		format:       ptr("json"),
		output:       ptr("report.json"),
		failOn:       ptr("warning"),
		noCycleCheck: ptr(true),
	})

	if cfg.Version != "v1.2.3-test" || cfg.Root != "/override/root" || !cfg.Quiet || !cfg.Verbose ||
		!cfg.ShowScores || !cfg.ShowImprovements || cfg.Format != "json" || cfg.Output != "report.json" ||
		cfg.FailOn != "warning" || !cfg.NoCycleCheck {
		t.Fatalf("explicit overrides were not applied: %+v", cfg)
	}
}

func TestApplyCLIOverridesPreservesLoadedConfigWithoutFlags(t *testing.T) {
	cfg := &config.Config{
		Root:             "/config/root",
		Quiet:            true,
		Verbose:          true,
		ShowScores:       true,
		ShowImprovements: true,
		Format:           "json",
		Output:           "report.json",
		FailOn:           "warning",
		NoCycleCheck:     true,
	}

	applyCLIOverrides(cfg, executionOptions{})

	if cfg.Root != "/config/root" || !cfg.Quiet || !cfg.Verbose || !cfg.ShowScores ||
		!cfg.ShowImprovements || cfg.Format != "json" || cfg.Output != "report.json" ||
		cfg.FailOn != "warning" || !cfg.NoCycleCheck {
		t.Fatalf("loaded values were overwritten without CLI flags: %+v", cfg)
	}
}

func TestApplyCLIOverridesHonorsExplicitFalse(t *testing.T) {
	cfg := &config.Config{
		Quiet:            true,
		Verbose:          true,
		ShowScores:       true,
		ShowImprovements: true,
		NoCycleCheck:     true,
	}

	applyCLIOverrides(cfg, executionOptions{
		quiet:        ptr(false),
		verbose:      ptr(false),
		scores:       ptr(false),
		improvements: ptr(false),
		noCycleCheck: ptr(false),
	})

	if cfg.Quiet || cfg.Verbose || cfg.ShowScores || cfg.ShowImprovements || cfg.NoCycleCheck {
		t.Fatalf("explicit false did not override loaded true values: %+v", cfg)
	}
}

func TestExecutionOptionsDoNotLeakAcrossExecutions(t *testing.T) {
	first := &config.Config{Quiet: true, Format: "markdown"}
	applyCLIOverrides(first, executionOptions{quiet: ptr(false), format: ptr("json")})
	if first.Quiet || first.Format != "json" {
		t.Fatalf("first execution overrides = %+v", first)
	}

	second := &config.Config{Quiet: true, Format: "markdown"}
	applyCLIOverrides(second, executionOptions{})
	if !second.Quiet || second.Format != "markdown" {
		t.Fatalf("second execution inherited first execution state: %+v", second)
	}
}

func TestLoadCLIConfigAbsentAndExplicitFalsePrecedence(t *testing.T) {
	root := t.TempDir()
	requireWriteFile(t, filepath.Join(root, ".cclintrc.yaml"), "quiet: true\nverbose: true\nno-cycle-check: true\n")

	loaded, err := loadCLIConfig(executionOptions{root: ptr(root)})
	if err != nil {
		t.Fatalf("load config without boolean flags: %v", err)
	}
	if !loaded.Quiet || !loaded.Verbose || !loaded.NoCycleCheck {
		t.Fatalf("absent flags did not preserve loaded true values: %+v", loaded)
	}

	overridden, err := loadCLIConfig(executionOptions{
		root:         ptr(root),
		quiet:        ptr(false),
		verbose:      ptr(false),
		noCycleCheck: ptr(false),
	})
	if err != nil {
		t.Fatalf("load config with explicit false flags: %v", err)
	}
	if overridden.Quiet || overridden.Verbose || overridden.NoCycleCheck {
		t.Fatalf("explicit false flags did not override loaded true values: %+v", overridden)
	}

	loadedAgain, err := loadCLIConfig(executionOptions{root: ptr(root)})
	if err != nil {
		t.Fatalf("load config after explicit false execution: %v", err)
	}
	if !loadedAgain.Quiet || !loadedAgain.Verbose || !loadedAgain.NoCycleCheck {
		t.Fatalf("later execution inherited earlier explicit false values: %+v", loadedAgain)
	}
}

func requireWriteFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
