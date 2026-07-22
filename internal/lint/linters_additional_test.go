package lint

import (
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

// TestPluginLinterScore tests plugin linter scoring
func TestPluginLinterScore(t *testing.T) {
	linter := NewPluginLinter("")
	data := map[string]any{
		"name":        "test-plugin",
		"description": "A comprehensive test plugin",
		"version":     "1.0.0",
	}

	score := linter.Score(`{"name":"test-plugin"}`, data, "")
	if score == nil {
		t.Error("Score() returned nil")
	}
}

// TestPluginLinterGetImprovements tests plugin linter improvements
func TestPluginLinterGetImprovements(t *testing.T) {
	linter := NewPluginLinter("")
	data := map[string]any{
		"name":        "test-plugin",
		"description": "Short",
	}

	improvements := linter.GetImprovements(`{}`, data)
	if improvements == nil {
		t.Error("GetImprovements() returned nil")
	}
}

// TestPluginLinterValidateCUE tests plugin linter CUE validation
func TestPluginLinterValidateCUE(t *testing.T) {
	linter := NewPluginLinter("")
	validator := cue.NewValidator()

	errors, _ := linter.ValidateCUE(validator, map[string]any{"name": "test"})
	// Should return empty slice (no CUE validation for plugins)
	if len(errors) != 0 {
		t.Errorf("ValidateCUE() returned %d errors, want 0", len(errors))
	}
}

// TestSettingsLinterValidateCUE tests settings linter CUE validation
func TestSettingsLinterValidateCUE(t *testing.T) {
	linter := NewSettingsLinter()
	validator := cue.NewValidator()

	errors, _ := linter.ValidateCUE(validator, map[string]any{"theme": "dark"})
	// Should return empty slice (no CUE validation for settings)
	if len(errors) != 0 {
		t.Errorf("ValidateCUE() returned %d errors, want 0", len(errors))
	}
}

// TestContextLinterValidateCUE tests context linter CUE validation
func TestContextLinterValidateCUE(t *testing.T) {
	linter := NewContextLinter()
	validator := cue.NewValidator()

	errors, _ := linter.ValidateCUE(validator, map[string]any{"sections": []any{}})
	// Should return empty slice (no CUE validation for context)
	if len(errors) != 0 {
		t.Errorf("ValidateCUE() returned %d errors, want 0", len(errors))
	}
}

// TestRegistrySingleSkillIntegrated tests single-file registry dispatch for skills.
func TestRegistrySingleSkillIntegrated(t *testing.T) {
	ctx := &SingleFileLinterContext{
		Quiet:     true,
		Verbose:   false,
		Validator: cue.NewValidator(),
		File: discovery.File{
			RelPath: "skills/test-skill/SKILL.md",
			Contents: `---
name: test-skill
---

## Quick Reference

Test content.
`,
		},
	}

	entry, _ := LinterForType(discovery.FileTypeSkill)
	result := lintComponent(ctx, entry.New(ctx.RootPath))
	if result.File == "" {
		t.Error("skill registry dispatch returned empty result")
	}
}

// TestRegistrySingleSettingsIntegrated tests single-file registry dispatch for settings.
func TestRegistrySingleSettingsIntegrated(t *testing.T) {
	ctx := &SingleFileLinterContext{
		Quiet:     true,
		Verbose:   false,
		Validator: cue.NewValidator(),
		File: discovery.File{
			RelPath:  ".claude/settings.json",
			Contents: `{"theme": "dark"}`,
		},
	}

	entry, _ := LinterForType(discovery.FileTypeSettings)
	result := lintComponent(ctx, entry.New(ctx.RootPath))
	if result.File == "" {
		t.Error("settings registry dispatch returned empty result")
	}
}

// TestRegistrySingleContextIntegrated tests single-file registry dispatch for context files.
func TestRegistrySingleContextIntegrated(t *testing.T) {
	ctx := &SingleFileLinterContext{
		Quiet:     true,
		Verbose:   false,
		Validator: cue.NewValidator(),
		File: discovery.File{
			RelPath: "CLAUDE.md",
			Contents: `# CLAUDE.md

## Build & Run

Instructions.
`,
		},
	}

	entry, _ := LinterForType(discovery.FileTypeContext)
	result := lintComponent(ctx, entry.New(ctx.RootPath))
	if result.File == "" {
		t.Error("context registry dispatch returned empty result")
	}
}

// TestRegistrySinglePluginIntegrated tests single-file registry dispatch for plugins.
func TestRegistrySinglePluginIntegrated(t *testing.T) {
	ctx := &SingleFileLinterContext{
		Quiet:     true,
		Verbose:   false,
		Validator: cue.NewValidator(),
		File: discovery.File{
			RelPath: ".claude-plugin/plugin.json",
			Contents: `{
  "name": "test-plugin",
  "description": "A comprehensive test plugin for validation purposes",
  "version": "1.0.0",
  "author": {"name": "Test Author"}
}`,
		},
	}

	entry, _ := LinterForType(discovery.FileTypePlugin)
	result := lintComponent(ctx, entry.New(ctx.RootPath))
	if result.File == "" {
		t.Error("plugin registry dispatch returned empty result")
	}
}
