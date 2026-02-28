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

// TestLintSingleSkillIntegrated tests lintSingleSkill integration
func TestLintSingleSkillIntegrated(t *testing.T) {
	ctx := &SingleFileLinterContext{
		Quiet:   true,
		Verbose: false,
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

	result := lintSingleSkill(ctx)
	if result.File == "" {
		t.Error("lintSingleSkill() returned empty result")
	}
}

// TestLintSingleSettingsIntegrated tests lintSingleSettings integration
func TestLintSingleSettingsIntegrated(t *testing.T) {
	ctx := &SingleFileLinterContext{
		Quiet:   true,
		Verbose: false,
		Validator: cue.NewValidator(),
		File: discovery.File{
			RelPath: ".claude/settings.json",
			Contents: `{"theme": "dark"}`,
		},
	}

	result := lintSingleSettings(ctx)
	if result.File == "" {
		t.Error("lintSingleSettings() returned empty result")
	}
}

// TestLintSingleContextIntegrated tests lintSingleContext integration
func TestLintSingleContextIntegrated(t *testing.T) {
	ctx := &SingleFileLinterContext{
		Quiet:   true,
		Verbose: false,
		Validator: cue.NewValidator(),
		File: discovery.File{
			RelPath: "CLAUDE.md",
			Contents: `# CLAUDE.md

## Build & Run

Instructions.
`,
		},
	}

	result := lintSingleContext(ctx)
	if result.File == "" {
		t.Error("lintSingleContext() returned empty result")
	}
}

// TestLintSinglePluginIntegrated tests lintSinglePlugin integration
func TestLintSinglePluginIntegrated(t *testing.T) {
	ctx := &SingleFileLinterContext{
		Quiet:   true,
		Verbose: false,
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

	result := lintSinglePlugin(ctx)
	if result.File == "" {
		t.Error("lintSinglePlugin() returned empty result")
	}
}
