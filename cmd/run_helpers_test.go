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
