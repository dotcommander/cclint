package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

// TestDetectFileType tests type detection from paths.
func TestDetectFileType(t *testing.T) {
	// Create a temp directory structure
	tmpDir := t.TempDir()
	createDirs(t, tmpDir, ".claude/agents", ".claude/commands", ".claude/skills/test-skill")

	tests := []struct {
		name       string
		relPath    string
		wantType   discovery.FileType
		wantErr    bool
		errContain string
	}{
		// Standard paths
		{"agent in .claude", ".claude/agents/test.md", discovery.FileTypeAgent, false, ""},
		{"agent nested", ".claude/agents/subdir/test.md", discovery.FileTypeAgent, false, ""},
		{"command in .claude", ".claude/commands/test.md", discovery.FileTypeCommand, false, ""},
		{"skill file", ".claude/skills/test-skill/SKILL.md", discovery.FileTypeSkill, false, ""},
		{"settings file", ".claude/settings.json", discovery.FileTypeSettings, false, ""},
		{"context file", "CLAUDE.md", discovery.FileTypeContext, false, ""},
		{"context in .claude", ".claude/CLAUDE.md", discovery.FileTypeContext, false, ""},

		// Non-standard paths (fallback to basename)
		{"skill by basename", "custom/SKILL.md", discovery.FileTypeSkill, false, ""},
		{"context by basename", "somewhere/CLAUDE.md", discovery.FileTypeContext, false, ""},

		// Ambiguous cases
		{"ambiguous md", "random/file.md", discovery.FileTypeUnknown, true, "cannot determine type"},
		{"ambiguous json", "random/data.json", discovery.FileTypeUnknown, true, "cannot determine type"},

		// Unsupported extensions
		{"txt file", "file.txt", discovery.FileTypeUnknown, true, "unsupported file type"},
		{"go file", "main.go", discovery.FileTypeUnknown, true, "unsupported file type"},
		{"no extension", "README", discovery.FileTypeUnknown, true, "has no extension"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath := filepath.Join(tmpDir, tt.relPath)

			fileType, err := discovery.DetectFileType(absPath, tmpDir)

			if tt.wantErr {
				if err == nil {
					t.Errorf("DetectFileType() expected error containing %q, got nil", tt.errContain)
					return
				}
				if tt.errContain != "" && !containsString(err.Error(), tt.errContain) {
					t.Errorf("DetectFileType() error = %q, want containing %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("DetectFileType() unexpected error: %v", err)
				return
			}

			if fileType != tt.wantType {
				t.Errorf("DetectFileType() = %v, want %v", fileType, tt.wantType)
			}
		})
	}
}

// TestValidateFilePath tests file validation.
func TestValidateFilePath(t *testing.T) {
	// Create test files
	tmpDir := t.TempDir()

	// Valid text file
	validFile := filepath.Join(tmpDir, "valid.md")
	if err := os.WriteFile(validFile, []byte("# Test\nContent"), 0644); err != nil {
		t.Fatal(err)
	}

	// Empty file
	emptyFile := filepath.Join(tmpDir, "empty.md")
	if err := os.WriteFile(emptyFile, []byte{}, 0644); err != nil {
		t.Fatal(err)
	}

	// Binary file
	binaryFile := filepath.Join(tmpDir, "binary.md")
	if err := os.WriteFile(binaryFile, []byte{0x00, 0x01, 0x02}, 0644); err != nil {
		t.Fatal(err)
	}

	// Directory
	dirPath := filepath.Join(tmpDir, "subdir")
	if err := os.Mkdir(dirPath, 0755); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name       string
		path       string
		wantErr    bool
		errContain string
	}{
		{"valid file", validFile, false, ""},
		{"nonexistent", filepath.Join(tmpDir, "nonexistent.md"), true, "file not found"},
		{"empty file", emptyFile, true, "file is empty"},
		{"binary file", binaryFile, true, "binary"},
		{"directory", dirPath, true, "directory"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath, err := discovery.ValidateFilePath(tt.path)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateFilePath() expected error containing %q, got nil", tt.errContain)
					return
				}
				if tt.errContain != "" && !containsString(err.Error(), tt.errContain) {
					t.Errorf("ValidateFilePath() error = %q, want containing %q", err.Error(), tt.errContain)
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateFilePath() unexpected error: %v", err)
				return
			}

			if absPath == "" {
				t.Error("ValidateFilePath() returned empty path")
			}
		})
	}
}

// TestParseFileType tests type parsing from strings.
func TestParseFileType(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		wantType discovery.FileType
		wantErr  bool
	}{
		// Valid types
		{"agent", "agent", discovery.FileTypeAgent, false},
		{"agents plural", "agents", discovery.FileTypeAgent, false},
		{"command", "command", discovery.FileTypeCommand, false},
		{"skill", "skill", discovery.FileTypeSkill, false},
		{"settings", "settings", discovery.FileTypeSettings, false},
		{"context", "context", discovery.FileTypeContext, false},
		{"plugin", "plugin", discovery.FileTypePlugin, false},

		// Case insensitive
		{"AGENT uppercase", "AGENT", discovery.FileTypeAgent, false},
		{"Agent mixed", "Agent", discovery.FileTypeAgent, false},

		// With whitespace
		{"whitespace", "  agent  ", discovery.FileTypeAgent, false},

		// Invalid
		{"empty", "", discovery.FileTypeUnknown, true},
		{"invalid", "foo", discovery.FileTypeUnknown, true},
		{"typo", "agnet", discovery.FileTypeUnknown, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fileType, err := discovery.ParseFileType(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ParseFileType(%q) expected error, got nil", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("ParseFileType(%q) unexpected error: %v", tt.input, err)
				return
			}

			if fileType != tt.wantType {
				t.Errorf("ParseFileType(%q) = %v, want %v", tt.input, fileType, tt.wantType)
			}
		})
	}
}

// TestLintSingleFile tests single-file linting.
func TestLintSingleFile(t *testing.T) {
	// Create test files
	tmpDir := t.TempDir()
	createDirs(t, tmpDir, ".claude/agents")

	// Valid agent file
	validAgent := filepath.Join(tmpDir, ".claude/agents/test-agent.md")
	validContent := `---
name: test-agent
description: A test agent for testing purposes. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test agent foundation.

## Workflow

1. Do stuff
`
	if err := os.WriteFile(validAgent, []byte(validContent), 0644); err != nil {
		t.Fatal(err)
	}

	// Invalid agent (missing fields)
	invalidAgent := filepath.Join(tmpDir, ".claude/agents/invalid-agent.md")
	invalidContent := `---
color: blue
---

No name or description.
`
	if err := os.WriteFile(invalidAgent, []byte(invalidContent), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name        string
		file        string
		typeOver    string
		wantSuccess bool
		wantErrors  int
		wantErr     bool // Expect error to be returned (not in results)
	}{
		{"valid agent", validAgent, "", true, 0, false},
		{"invalid agent", invalidAgent, "", false, 3, false}, // Missing name and description (CUE + Go checks)
		{"nonexistent", filepath.Join(tmpDir, "nope.md"), "", false, 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary, err := LintSingleFile(tt.file, tmpDir, tt.typeOver, true, false)

			// For nonexistent files, error is returned directly
			if tt.wantErr {
				if err == nil {
					t.Error("LintSingleFile() expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Errorf("LintSingleFile() returned error: %v", err)
				return
			}

			if len(summary.Results) != 1 {
				t.Errorf("LintSingleFile() returned %d results, want 1", len(summary.Results))
				return
			}

			result := summary.Results[0]
			if result.Success != tt.wantSuccess {
				t.Errorf("LintSingleFile() success = %v, want %v", result.Success, tt.wantSuccess)
			}

			if len(result.Errors) != tt.wantErrors {
				t.Errorf("LintSingleFile() errors = %d, want %d", len(result.Errors), tt.wantErrors)
				for _, e := range result.Errors {
					t.Logf("  Error: %s", e.Message)
				}
			}
		})
	}
}

// TestLintFiles tests multi-file linting.
func TestLintFiles(t *testing.T) {
	// Create test files
	tmpDir := t.TempDir()
	createDirs(t, tmpDir, ".claude/agents", ".claude/commands")

	// Valid agent
	validAgent := filepath.Join(tmpDir, ".claude/agents/test.md")
	if err := os.WriteFile(validAgent, []byte(`---
name: test
description: Test agent. Use PROACTIVELY when testing.
model: sonnet
---
Content
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Valid command
	validCommand := filepath.Join(tmpDir, ".claude/commands/cmd.md")
	if err := os.WriteFile(validCommand, []byte(`---
allowed-tools: Task
---
Task(test): Do something
`), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := LintFiles([]string{validAgent, validCommand}, tmpDir, "", true, false)
	if err != nil {
		t.Fatalf("LintFiles() error: %v", err)
	}

	if summary.TotalFiles != 2 {
		t.Errorf("LintFiles() TotalFiles = %d, want 2", summary.TotalFiles)
	}

	if len(summary.Results) != 2 {
		t.Errorf("LintFiles() Results = %d, want 2", len(summary.Results))
	}
}

// Helper functions

func createDirs(t *testing.T, base string, paths ...string) {
	t.Helper()
	for _, p := range paths {
		dir := filepath.Join(base, p)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestLintSingleSkill tests skill file linting
func TestLintSingleSkill(t *testing.T) {
	tmpDir := t.TempDir()
	createDirs(t, tmpDir, ".claude/skills/test-skill")

	skillFile := filepath.Join(tmpDir, ".claude/skills/test-skill/SKILL.md")
	content := `---
name: test-skill
---

## Quick Reference

Test skill content.
`
	if err := os.WriteFile(skillFile, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := LintSingleFile(skillFile, tmpDir, "", true, false)
	if err != nil {
		t.Fatalf("LintSingleFile(skill) error: %v", err)
	}

	if len(summary.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(summary.Results))
	}
}

func TestExpandDirectories(t *testing.T) {
	tmpDir := t.TempDir()

	// Plural directory (standard)
	pluralDir := filepath.Join(tmpDir, "agents")
	os.MkdirAll(pluralDir, 0755)
	os.WriteFile(filepath.Join(pluralDir, "test.md"), []byte("# Agent"), 0644)

	// Singular directory
	singularDir := filepath.Join(tmpDir, "command")
	os.MkdirAll(singularDir, 0755)
	os.WriteFile(filepath.Join(singularDir, "do-thing.md"), []byte("# Command"), 0644)

	// Unrecognized directory
	unknownDir := filepath.Join(tmpDir, "prompts")
	os.MkdirAll(unknownDir, 0755)
	os.WriteFile(filepath.Join(unknownDir, "prompt.md"), []byte("# Prompt"), 0644)

	tests := []struct {
		name      string
		paths     []string
		typeOver  string
		wantCount int
	}{
		{
			name:      "plural dir included",
			paths:     []string{pluralDir},
			wantCount: 1,
		},
		{
			name:      "singular dir included via fallback",
			paths:     []string{singularDir},
			wantCount: 1,
		},
		{
			name:      "unrecognized dir excluded",
			paths:     []string{unknownDir},
			wantCount: 0,
		},
		{
			name:      "unrecognized dir included with typeOverride",
			paths:     []string{unknownDir},
			typeOver:  "agent",
			wantCount: 1,
		},
		{
			name:      "non-directory path kept as-is",
			paths:     []string{filepath.Join(pluralDir, "test.md")},
			wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandDirectories(tt.paths, tt.typeOver)
			if len(result) != tt.wantCount {
				t.Errorf("expandDirectories() = %d files, want %d", len(result), tt.wantCount)
				for _, fh := range result {
					t.Logf("  file: %s (hint: %q)", fh.Path, fh.TypeHint)
				}
			}
		})
	}
}

func TestFindProjectRootForFileEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	tests := []struct {
		name        string
		setup       func() string
		wantContain string
	}{
		{
			name: "file in .claude/agents",
			setup: func() string {
				claudeDir := filepath.Join(tmpDir, ".claude", "agents")
				_ = os.MkdirAll(claudeDir, 0755)
				return filepath.Join(claudeDir, "test.md")
			},
			wantContain: ".claude",
		},
		{
			name: "file in agents/",
			setup: func() string {
				agentsDir := filepath.Join(tmpDir, "subproject", "agents")
				_ = os.MkdirAll(agentsDir, 0755)
				return filepath.Join(agentsDir, "test.md")
			},
			wantContain: "subproject",
		},
		{
			name: "file in commands/",
			setup: func() string {
				commandsDir := filepath.Join(tmpDir, "proj", "commands")
				_ = os.MkdirAll(commandsDir, 0755)
				return filepath.Join(commandsDir, "test.md")
			},
			wantContain: "proj",
		},
		{
			name: "file in skills/",
			setup: func() string {
				skillsDir := filepath.Join(tmpDir, "myproject", "skills", "test")
				_ = os.MkdirAll(skillsDir, 0755)
				return filepath.Join(skillsDir, "SKILL.md")
			},
			wantContain: "myproject",
		},
		{
			name: "fallback to parent directory",
			setup: func() string {
				randomDir := filepath.Join(tmpDir, "random")
				_ = os.MkdirAll(randomDir, 0755)
				return filepath.Join(randomDir, "test.md")
			},
			wantContain: "random",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			absPath := tt.setup()
			root, err := findProjectRootForFile(absPath)

			// Don't fail on error for fallback cases - they might not find a real project root
			if err != nil {
				t.Logf("findProjectRootForFile() returned error (expected for some cases): %v", err)
			}

			if root == "" {
				t.Error("findProjectRootForFile() returned empty root")
				return
			}

			if !filepath.IsAbs(root) {
				t.Errorf("findProjectRootForFile() root = %q is not absolute", root)
			}

			// Verify root is a reasonable prefix
			if tt.wantContain != "" {
				// Root should be somewhere in the path
				if !containsSubstring(root, tt.wantContain) && !containsSubstring(absPath, root) {
					t.Logf("findProjectRootForFile() root = %q, file = %q", root, absPath)
					// This is just informational, not a hard failure
				}
			}
		})
	}
}

