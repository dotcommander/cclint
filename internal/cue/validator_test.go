package cue

import (
	"strings"
	"testing"
)

// TestNewValidator tests the Validator constructor
func TestNewValidator(t *testing.T) {
	v := NewValidator()
	if v == nil {
		t.Fatal("NewValidator returned nil")
	}
	if v.ctx == nil {
		t.Error("Validator.ctx is nil")
	}
	if v.schemas == nil {
		t.Error("Validator.schemas is nil")
	}
	if len(v.schemas) != 0 {
		t.Errorf("Expected empty schemas map, got %d entries", len(v.schemas))
	}
}

// TestLoadSchemas tests loading embedded CUE schemas
func TestLoadSchemas(t *testing.T) {
	v := NewValidator()
	err := v.LoadSchemas("schemas")

	if err != nil {
		t.Errorf("LoadSchemas failed: %v", err)
	}

	// Check that schemas were loaded
	expectedSchemas := []string{"agent", "command", "skill", "settings", "claude_md"}
	for _, name := range expectedSchemas {
		if _, ok := v.schemas[name]; !ok {
			t.Errorf("Expected schema %q to be loaded", name)
		}
	}
}

// TestLoadSchemas_InvalidDir tests LoadSchemas with invalid directory
func TestLoadSchemas_InvalidDir(t *testing.T) {
	v := NewValidator()
	// This should still succeed because we use embedded FS
	err := v.LoadSchemas("nonexistent")

	// Should still load from embedded schemas
	if err != nil {
		t.Errorf("LoadSchemas should use embedded FS: %v", err)
	}
}

// TestValidateAgent tests agent validation with valid and invalid data
func TestValidateAgent(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name: "valid agent",
			data: map[string]any{
				"name":        "test-agent",
				"description": "A test agent",
				"model":       "sonnet",
				"color":       "blue",
				"tools":       "*",
			},
			wantError: false,
		},
		{
			name: "valid agent minimal",
			data: map[string]any{
				"name":        "minimal-agent",
				"description": "Minimal valid agent",
			},
			wantError: false,
		},
		{
			name: "invalid name - uppercase",
			data: map[string]any{
				"name":        "TestAgent",
				"description": "Agent with uppercase name",
			},
			wantError: true,
		},
		{
			name: "invalid name - spaces",
			data: map[string]any{
				"name":        "test agent",
				"description": "Agent with spaces in name",
			},
			wantError: true,
		},
		{
			name: "missing required name",
			data: map[string]any{
				"description": "Agent without name",
			},
			wantError: true,
		},
		{
			name: "missing required description",
			data: map[string]any{
				"name": "test-agent",
			},
			wantError: true,
		},
		{
			name: "empty description",
			data: map[string]any{
				"name":        "test-agent",
				"description": "",
			},
			wantError: true,
		},
		{
			name: "invalid model",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with invalid model",
				"model":       "gpt-4",
			},
			wantError: true,
		},
		{
			name: "invalid color",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with invalid color",
				"color":       "turquoise",
			},
			wantError: true,
		},
		{
			name: "valid tools as string",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with comma-separated tools",
				"tools":       "Read,Write,Edit",
			},
			wantError: false,
		},
		{
			name: "valid tools as array",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with tool array",
				"tools":       []string{"Read", "Write", "Edit"},
			},
			wantError: false,
		},
		{
			name: "valid with skills",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with auto-load skills",
				"skills":      "skill-one,skill-two",
			},
			wantError: false,
		},
		{
			name: "valid with hooks",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with hooks",
				"hooks": map[string]any{
					"PreToolUse": []map[string]any{
						{
							"matcher": "Bash",
							"hooks": []map[string]any{
								{
									"type":    "command",
									"command": "echo 'test'",
									"timeout": 30,
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid with memory user scope",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with user memory",
				"memory":      "user",
			},
			wantError: false,
		},
		{
			name: "valid with memory project scope",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with project memory",
				"memory":      "project",
			},
			wantError: false,
		},
		{
			name: "valid with memory local scope",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with local memory",
				"memory":      "local",
			},
			wantError: false,
		},
		{
			name: "invalid memory scope",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with bad memory scope",
				"memory":      "global",
			},
			wantError: true,
		},
		{
			name: "valid tools with Task(agent_type)",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with restricted sub-agents",
				"tools":       []string{"Read", "Write", "Task(poet-agent)"},
			},
			wantError: false,
		},
		{
			name: "valid tools with multiple Task refs",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with multiple sub-agent restrictions",
				"tools":       []string{"Read", "Task(bee-agent)", "Task(quality-agent)"},
			},
			wantError: false,
		},
		{
			name: "valid tools Task-only",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with only Task refs",
				"tools":       []string{"Task(worker-1)"},
			},
			wantError: false,
		},
		{
			name: "valid tools string with Task syntax",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with Task in comma string",
				"tools":       "Read,Write,Task(poet-agent)",
			},
			wantError: false,
		},
		{
			name: "valid with Setup hook",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with Setup hook",
				"hooks": map[string]any{
					"Setup": []map[string]any{
						{
							"matcher": "init",
							"hooks": []map[string]any{
								{
									"type":    "command",
									"command": "echo 'setup'",
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
	}

	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := v.ValidateAgent(tt.data)
			if err != nil {
				t.Fatalf("ValidateAgent returned error: %v", err)
			}

			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateAgent() hasErrors = %v, want %v", hasErrors, tt.wantError)
				if hasErrors {
					for _, e := range errs {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
		})
	}
}

// TestValidateCommand tests command validation
func TestValidateCommand(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name: "valid command full",
			data: map[string]any{
				"name":            "test-cmd",
				"description":     "A test command",
				"allowed-tools":   "*",
				"argument-hint":   "<arg>",
				"model":           "sonnet",
			},
			wantError: false,
		},
		{
			name:      "valid command empty",
			data:      map[string]any{},
			wantError: false,
		},
		{
			name: "valid command minimal",
			data: map[string]any{
				"name": "minimal",
			},
			wantError: false,
		},
		{
			name: "invalid name - uppercase",
			data: map[string]any{
				"name": "TestCommand",
			},
			wantError: true,
		},
		{
			name: "invalid name - underscore",
			data: map[string]any{
				"name": "test_command",
			},
			wantError: true,
		},
		{
			name: "valid allowed-tools as array",
			data: map[string]any{
				"name":          "test-cmd",
				"allowed-tools": []string{"Read", "Write"},
			},
			wantError: false,
		},
		{
			name: "valid allowed-tools as string",
			data: map[string]any{
				"name":          "test-cmd",
				"allowed-tools": "Read,Write,Edit",
			},
			wantError: false,
		},
		{
			name: "invalid model",
			data: map[string]any{
				"name":  "test-cmd",
				"model": "claude-invalid",
			},
			wantError: true,
		},
	}

	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := v.ValidateCommand(tt.data)
			if err != nil {
				t.Fatalf("ValidateCommand returned error: %v", err)
			}

			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateCommand() hasErrors = %v, want %v", hasErrors, tt.wantError)
				if hasErrors {
					for _, e := range errs {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
		})
	}
}

// TestValidateSkill tests skill validation
func TestValidateSkill(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name: "valid skill",
			data: map[string]any{
				"name":          "test-skill",
				"description":   "A test skill",
				"allowed-tools": "*",
				"model":         "sonnet",
			},
			wantError: false,
		},
		{
			name: "valid skill minimal",
			data: map[string]any{
				"name":        "minimal-skill",
				"description": "Minimal valid skill",
			},
			wantError: false,
		},
		{
			name: "missing required name",
			data: map[string]any{
				"description": "Skill without name",
			},
			wantError: true,
		},
		{
			name: "missing required description",
			data: map[string]any{
				"name": "test-skill",
			},
			wantError: true,
		},
		{
			name: "empty description",
			data: map[string]any{
				"name":        "test-skill",
				"description": "",
			},
			wantError: true,
		},
		{
			name: "invalid name format",
			data: map[string]any{
				"name":        "Test_Skill",
				"description": "Skill with invalid name",
			},
			wantError: true,
		},
		{
			name: "valid with full model name",
			data: map[string]any{
				"name":        "test-skill",
				"description": "Skill with full model name",
				"model":       "claude-sonnet-4-20250514",
			},
			wantError: false,
		},
		{
			name: "valid allowed-tools array",
			data: map[string]any{
				"name":          "test-skill",
				"description":   "Skill with tool array",
				"allowed-tools": []string{"Read", "Grep", "Glob"},
			},
			wantError: false,
		},
	}

	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := v.ValidateSkill(tt.data)
			if err != nil {
				t.Fatalf("ValidateSkill returned error: %v", err)
			}

			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateSkill() hasErrors = %v, want %v", hasErrors, tt.wantError)
				if hasErrors {
					for _, e := range errs {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
		})
	}
}

// TestValidateSettings tests settings validation
func TestValidateSettings(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name:      "valid empty settings",
			data:      map[string]any{},
			wantError: false,
		},
		{
			name: "valid settings with hooks",
			data: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": []map[string]any{
						{
							"matcher": "Bash",
							"hooks": []map[string]any{
								{
									"type":    "command",
									"command": "echo 'test'",
									"timeout": 30,
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid settings with additional fields",
			data: map[string]any{
				"model":       "sonnet",
				"permissions": map[string]any{},
				"mcp":         map[string]any{},
			},
			wantError: false,
		},
		{
			name: "invalid hook - missing command",
			data: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": []map[string]any{
						{
							"matcher": "Bash",
							"hooks": []map[string]any{
								{
									"type": "command",
									// missing "command" field
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "invalid hook - wrong type value",
			data: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": []map[string]any{
						{
							"matcher": "Bash",
							"hooks": []map[string]any{
								{
									"type":    "invalid",
									"command": "echo 'test'",
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
	}

	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := v.ValidateSettings(tt.data)
			if err != nil {
				t.Fatalf("ValidateSettings returned error: %v", err)
			}

			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateSettings() hasErrors = %v, want %v", hasErrors, tt.wantError)
				if hasErrors {
					for _, e := range errs {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
		})
	}
}

// TestValidateClaudeMD tests CLAUDE.md validation
func TestValidateClaudeMD(t *testing.T) {
	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name:      "valid empty",
			data:      map[string]any{},
			wantError: false,
		},
		{
			name: "valid with metadata",
			data: map[string]any{
				"title":       "My Project",
				"description": "Project documentation",
			},
			wantError: false,
		},
		{
			name: "valid with model",
			data: map[string]any{
				"model": "sonnet",
			},
			wantError: false,
		},
		{
			name: "valid with allowed-tools string",
			data: map[string]any{
				"allowed-tools": "Read,Write,Edit",
			},
			wantError: false,
		},
		{
			name: "valid with allowed-tools array",
			data: map[string]any{
				"allowed-tools": []string{"Read", "Write", "Edit"},
			},
			wantError: false,
		},
		{
			name: "valid with sections",
			data: map[string]any{
				"sections": []map[string]any{
					{
						"heading": "Build & Commands",
						"content": "Instructions here",
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid with rules",
			data: map[string]any{
				"rules": []map[string]any{
					{
						"name":        "test-rule",
						"description": "A test rule",
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid with additional fields",
			data: map[string]any{
				"title":       "Project",
				"description": "Desc",
				"customField": "custom value",
			},
			wantError: false,
		},
	}

	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := v.ValidateClaudeMD(tt.data)
			if err != nil {
				t.Fatalf("ValidateClaudeMD returned error: %v", err)
			}

			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateClaudeMD() hasErrors = %v, want %v", hasErrors, tt.wantError)
				if hasErrors {
					for _, e := range errs {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
		})
	}
}

// TestValidateAgent_NoSchema tests validation when schema is not loaded
func TestValidateAgent_NoSchema(t *testing.T) {
	v := NewValidator()
	// Don't load schemas

	data := map[string]any{
		"name":        "test-agent",
		"description": "Test",
	}

	errs, err := v.ValidateAgent(data)
	if err != nil {
		t.Fatalf("ValidateAgent should not error when schema missing: %v", err)
	}
	if len(errs) > 0 {
		t.Error("ValidateAgent should return no errors when schema not loaded")
	}
}

// TestValidateCommand_NoSchema tests validation when schema is not loaded
func TestValidateCommand_NoSchema(t *testing.T) {
	v := NewValidator()
	// Don't load schemas

	data := map[string]any{
		"name": "test-cmd",
	}

	errs, err := v.ValidateCommand(data)
	if err != nil {
		t.Fatalf("ValidateCommand should not error when schema missing: %v", err)
	}
	if len(errs) > 0 {
		t.Error("ValidateCommand should return no errors when schema not loaded")
	}
}

// TestValidateSkill_NoSchema tests validation when schema is not loaded
func TestValidateSkill_NoSchema(t *testing.T) {
	v := NewValidator()
	// Don't load schemas

	data := map[string]any{
		"name":        "test-skill",
		"description": "Test",
	}

	errs, err := v.ValidateSkill(data)
	if err != nil {
		t.Fatalf("ValidateSkill should not error when schema missing: %v", err)
	}
	if len(errs) > 0 {
		t.Error("ValidateSkill should return no errors when schema not loaded")
	}
}

// TestValidateSettings_NoSchema tests validation when schema is not loaded
func TestValidateSettings_NoSchema(t *testing.T) {
	v := NewValidator()
	// Don't load schemas

	data := map[string]any{}

	errs, err := v.ValidateSettings(data)
	if err != nil {
		t.Fatalf("ValidateSettings should not error when schema missing: %v", err)
	}
	if len(errs) > 0 {
		t.Error("ValidateSettings should return no errors when schema not loaded")
	}
}

// TestValidateClaudeMD_NoSchema tests validation when schema is not loaded
func TestValidateClaudeMD_NoSchema(t *testing.T) {
	v := NewValidator()
	// Don't load schemas

	data := map[string]any{}

	errs, err := v.ValidateClaudeMD(data)
	if err != nil {
		t.Fatalf("ValidateClaudeMD should not error when schema missing: %v", err)
	}
	if len(errs) > 0 {
		t.Error("ValidateClaudeMD should return no errors when schema not loaded")
	}
}

// TestValidationError tests the ValidationError structure
func TestValidationError(t *testing.T) {
	err := ValidationError{
		File:     "/path/to/file.md",
		Message:  "Test error message",
		Severity: "error",
		Source:   SourceAnthropicDocs,
		Line:     10,
		Column:   5,
	}

	if err.File != "/path/to/file.md" {
		t.Errorf("Expected File = %q, got %q", "/path/to/file.md", err.File)
	}
	if err.Message != "Test error message" {
		t.Errorf("Expected Message = %q, got %q", "Test error message", err.Message)
	}
	if err.Severity != "error" {
		t.Errorf("Expected Severity = %q, got %q", "error", err.Severity)
	}
	if err.Source != SourceAnthropicDocs {
		t.Errorf("Expected Source = %q, got %q", SourceAnthropicDocs, err.Source)
	}
	if err.Line != 10 {
		t.Errorf("Expected Line = %d, got %d", 10, err.Line)
	}
	if err.Column != 5 {
		t.Errorf("Expected Column = %d, got %d", 5, err.Column)
	}
}

// TestParseFrontmatter tests frontmatter parsing
func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		content     string
		wantData    map[string]any
		wantBody    string
		wantError   bool
	}{
		{
			name: "valid frontmatter",
			content: `---
name: test-agent
description: Test agent
---
# Agent Body

Content here`,
			wantData: map[string]any{
				"name":        "test-agent",
				"description": "Test agent",
			},
			wantBody:  "\n# Agent Body\n\nContent here",
			wantError: false,
		},
		{
			name: "no frontmatter",
			content: `# Just Content

No frontmatter here`,
			wantData:  map[string]any{},
			wantBody:  "# Just Content\n\nNo frontmatter here",
			wantError: false,
		},
		{
			name:    "empty content",
			content: "",
			wantData: map[string]any{},
			wantBody:  "",
			wantError: false,
		},
		{
			name: "invalid YAML",
			content: `---
name: test
  invalid: yaml: syntax
---
Body`,
			wantData:  nil,
			wantBody:  "",
			wantError: true,
		},
		{
			name: "frontmatter with complex data",
			content: `---
name: test-agent
tools:
  - Read
  - Write
hooks:
  PreToolUse:
    - matcher: Bash
      hooks:
        - type: command
          command: echo test
---
Body content`,
			wantData: map[string]any{
				"name": "test-agent",
				"tools": []any{"Read", "Write"},
				"hooks": map[string]any{
					"PreToolUse": []any{
						map[string]any{
							"matcher": "Bash",
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "echo test",
								},
							},
						},
					},
				},
			},
			wantBody:  "\nBody content",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fm, err := ParseFrontmatter(tt.content)

			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if fm == nil {
				t.Fatal("ParseFrontmatter returned nil")
			}

			if fm.Body != tt.wantBody {
				t.Errorf("Body = %q, want %q", fm.Body, tt.wantBody)
			}

			// Check data keys match
			if len(fm.Data) != len(tt.wantData) {
				t.Errorf("Data has %d keys, want %d", len(fm.Data), len(tt.wantData))
			}

			for key := range tt.wantData {
				if _, ok := fm.Data[key]; !ok {
					t.Errorf("Missing expected key %q in data", key)
				}
			}
		})
	}
}

// TestValidateFile tests file validation
func TestValidateFile(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	tests := []struct {
		name      string
		path      string
		content   string
		fileType  string
		wantError bool
	}{
		{
			name: "valid agent file",
			path: "/test/agent.md",
			content: `---
name: test-agent
description: Test agent
---
# Agent content`,
			fileType:  "agent",
			wantError: false,
		},
		{
			name: "valid command file",
			path: "/test/command.md",
			content: `---
name: test-cmd
---
# Command content`,
			fileType:  "command",
			wantError: false,
		},
		{
			name: "valid skill file",
			path: "/test/skill.md",
			content: `---
name: test-skill
description: Test skill
---
# Skill content`,
			fileType:  "skill",
			wantError: false,
		},
		{
			name: "invalid agent - missing required field",
			path: "/test/agent.md",
			content: `---
name: test-agent
---
# Missing description`,
			fileType:  "agent",
			wantError: true,
		},
		{
			name: "invalid frontmatter YAML",
			path: "/test/agent.md",
			content: `---
invalid: yaml: syntax
---
Content`,
			fileType:  "agent",
			wantError: true,
		},
		{
			name:      "unknown file type",
			path:      "/test/unknown.md",
			content:   "# Content",
			fileType:  "unknown",
			wantError: true,
		},
		{
			name: "valid settings file",
			path: "/test/settings.json",
			content: `---
hooks:
  PreToolUse:
    - matcher: Bash
      hooks:
        - type: command
          command: echo test
---`,
			fileType:  "settings",
			wantError: false,
		},
		{
			name: "valid claude_md file",
			path: "/test/CLAUDE.md",
			content: `---
title: My Project
model: sonnet
---
# Project docs`,
			fileType:  "claude_md",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := v.ValidateFile(tt.path, tt.content, tt.fileType)

			if tt.fileType == "unknown" {
				if err == nil {
					t.Error("Expected error for unknown file type")
				}
				return
			}

			if err != nil && !strings.Contains(tt.name, "invalid frontmatter") {
				t.Fatalf("ValidateFile returned error: %v", err)
			}

			hasErrors := len(errs) > 0 || err != nil
			if hasErrors != tt.wantError {
				t.Errorf("ValidateFile() hasErrors = %v, want %v", hasErrors, tt.wantError)
				if len(errs) > 0 {
					for _, e := range errs {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
		})
	}
}

// TestExtractErrorsFromCUE tests error extraction from CUE validation
func TestExtractErrorsFromCUE(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	// Trigger a validation error
	invalidData := map[string]any{
		"name":        "InvalidName",  // uppercase not allowed
		"description": "Test",
	}

	errs, err := v.ValidateAgent(invalidData)
	if err != nil {
		t.Fatalf("ValidateAgent returned error: %v", err)
	}

	if len(errs) == 0 {
		t.Error("Expected validation errors, got none")
	}

	for _, e := range errs {
		if e.Message == "" {
			t.Error("ValidationError has empty message")
		}
		if e.Severity != "error" {
			t.Errorf("Expected severity = %q, got %q", "error", e.Severity)
		}
		if e.Source != SourceAnthropicDocs {
			t.Errorf("Expected source = %q, got %q", SourceAnthropicDocs, e.Source)
		}
	}
}

// TestSourceConstants tests the source constants
func TestSourceConstants(t *testing.T) {
	if SourceAnthropicDocs != "anthropic-docs" {
		t.Errorf("SourceAnthropicDocs = %q, want %q", SourceAnthropicDocs, "anthropic-docs")
	}
	if SourceCClintObserve != "cclint-observation" {
		t.Errorf("SourceCClintObserve = %q, want %q", SourceCClintObserve, "cclint-observation")
	}
}

// TestValidateAgainstSchema_InvalidData tests encoding errors
func TestValidateAgainstSchema_InvalidData(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	// Create data with complex types that might cause encoding issues
	// CUE should handle this gracefully
	data := map[string]any{
		"name":        "test-agent",
		"description": "Test with complex nested structures",
		"nested": map[string]any{
			"deep": map[string]any{
				"value": []int{1, 2, 3},
			},
		},
	}

	errs, err := v.ValidateAgent(data)
	if err != nil {
		t.Fatalf("ValidateAgent should handle complex data: %v", err)
	}
	// Nested data should be allowed (schema allows additional fields)
	if len(errs) > 0 {
		t.Logf("Got validation errors: %v", errs)
	}
}

// TestValidateAgainstSchema_MissingDefinition tests non-existent schema definition
func TestValidateAgainstSchema_MissingDefinition(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	// Manually test with a schema type that doesn't have a definition
	// This tests the path where def.Exists() returns false
	schema, ok := v.schemas["agent"]
	if !ok {
		t.Fatal("Agent schema not loaded")
	}

	// Try to validate against a non-existent definition
	errs, err := v.validateAgainstSchema(schema, map[string]any{}, "nonexistent")
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	// Should return no errors when definition doesn't exist
	if len(errs) > 0 {
		t.Errorf("Expected no errors for missing definition, got %d", len(errs))
	}
}

// TestValidateAgainstSchema_ConcreteValidation tests concrete value validation
func TestValidateAgainstSchema_ConcreteValidation(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	// Test with incomplete data that fails concrete validation
	// This should trigger the cue.Concrete(true) validation path
	data := map[string]any{
		"name": "test-agent",
		// Missing required description
	}

	errs, err := v.ValidateAgent(data)
	if err != nil {
		t.Fatalf("ValidateAgent returned error: %v", err)
	}

	// Should have validation errors for missing required field
	if len(errs) == 0 {
		t.Error("Expected validation errors for missing required field")
	}
}

// TestValidateAgent_EdgeCases tests edge cases for agent validation
func TestValidateAgent_EdgeCases(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name: "name at max length",
			data: map[string]any{
				"name":        strings.Repeat("a", 64),
				"description": "Test",
			},
			wantError: false,
		},
		{
			name: "name exceeds max length",
			data: map[string]any{
				"name":        strings.Repeat("a", 65),
				"description": "Test",
			},
			wantError: true,
		},
		{
			name: "description at max length",
			data: map[string]any{
				"name":        "test-agent",
				"description": strings.Repeat("a", 1024),
			},
			wantError: false,
		},
		{
			name: "description exceeds max length",
			data: map[string]any{
				"name":        "test-agent",
				"description": strings.Repeat("a", 1025),
			},
			wantError: true,
		},
		{
			name: "all valid models",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Test",
				"model":       "sonnet",
			},
			wantError: false,
		},
		{
			name: "all valid colors",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Test",
				"color":       "red",
			},
			wantError: false,
		},
		{
			name: "valid permission modes",
			data: map[string]any{
				"name":           "test-agent",
				"description":    "Test",
				"permissionMode": "acceptEdits",
			},
			wantError: false,
		},
		{
			name: "extra custom fields",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Test",
				"customField": "custom value",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, err := v.ValidateAgent(tt.data)
			if err != nil {
				t.Fatalf("ValidateAgent returned error: %v", err)
			}

			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateAgent() hasErrors = %v, want %v", hasErrors, tt.wantError)
				if hasErrors {
					for _, e := range errs {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
		})
	}
}
