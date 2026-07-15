package cmd

import (
	"testing"

	"github.com/dotcommander/cclint/internal/config"
)

func TestApplyCLIOverridesSetsVersion(t *testing.T) {
	oldVersion := Version
	oldRootPath := rootPath
	oldOutputFormat := outputFormat
	oldFailOn := failOn
	t.Cleanup(func() {
		Version = oldVersion
		rootPath = oldRootPath
		outputFormat = oldOutputFormat
		failOn = oldFailOn
	})

	Version = "v1.2.3-test"
	rootPath = "/override/root"
	outputFormat = "json"
	failOn = "warning"

	cfg := &config.Config{}
	applyCLIOverrides(cfg)

	if cfg.Version != "v1.2.3-test" {
		t.Fatalf("cfg.Version = %q, want v1.2.3-test", cfg.Version)
	}
	if cfg.Root != "/override/root" {
		t.Fatalf("cfg.Root = %q, want /override/root", cfg.Root)
	}
	if cfg.Format != "json" {
		t.Fatalf("cfg.Format = %q, want json", cfg.Format)
	}
	if cfg.FailOn != "warning" {
		t.Fatalf("cfg.FailOn = %q, want warning", cfg.FailOn)
	}
}

func TestApplyCLIOverridesPreservesLoadedConfigWithoutFlags(t *testing.T) {
	oldCLIChanged := cliChanged
	oldRootPath := rootPath
	oldQuiet := quiet
	oldOutputFormat := outputFormat
	oldNoCycleCheck := noCycleCheck
	t.Cleanup(func() {
		cliChanged = oldCLIChanged
		rootPath = oldRootPath
		quiet = oldQuiet
		outputFormat = oldOutputFormat
		noCycleCheck = oldNoCycleCheck
	})

	cliChanged = map[string]bool{}
	rootPath = "/cli/root"
	quiet = false
	outputFormat = "console"
	noCycleCheck = false
	cfg := &config.Config{
		Root:         "/config/root",
		Quiet:        true,
		Format:       "json",
		NoCycleCheck: true,
	}

	applyCLIOverrides(cfg)

	if cfg.Root != "/config/root" || !cfg.Quiet || cfg.Format != "json" || !cfg.NoCycleCheck {
		t.Fatalf("config values were overwritten without CLI flags: %+v", cfg)
	}
}

func TestApplyCLIOverridesHonorsExplicitFalse(t *testing.T) {
	oldCLIChanged := cliChanged
	oldQuiet := quiet
	t.Cleanup(func() {
		cliChanged = oldCLIChanged
		quiet = oldQuiet
	})

	explicitFalse := false
	app := cli{Quiet: &explicitFalse}
	app.apply()
	cfg := &config.Config{Quiet: true}

	applyCLIOverrides(cfg)

	if cfg.Quiet {
		t.Fatal("explicit --quiet=false did not override loaded config")
	}
}
