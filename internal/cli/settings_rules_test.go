package cli

import (
	"testing"
)

func TestValidateRules(t *testing.T) {
	tests := []struct {
		name           string
		rules          any
		wantErrorCount int
		wantSeverity   string // check first error severity if set
	}{
		{
			name:           "not an array",
			rules:          "not an array",
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "nil value",
			rules:          nil,
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "empty array is valid",
			rules:          []any{},
			wantErrorCount: 0,
		},
		{
			name:           "valid simple glob",
			rules:          []any{"rules/*.md"},
			wantErrorCount: 0,
		},
		{
			name:           "valid doublestar glob",
			rules:          []any{"**/*.md"},
			wantErrorCount: 0,
		},
		{
			name:           "valid multiple patterns",
			rules:          []any{"rules/*.md", "**/*.md", "docs/rules.md"},
			wantErrorCount: 0,
		},
		{
			name:           "valid brace expansion",
			rules:          []any{"rules/*.{md,txt}"},
			wantErrorCount: 0,
		},
		{
			name:           "empty string entry",
			rules:          []any{""},
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "non-string entry",
			rules:          []any{42},
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "absolute path warns",
			rules:          []any{"/etc/rules/*.md"},
			wantErrorCount: 1,
			wantSeverity:   "warning",
		},
		{
			name:           "leading slash warns",
			rules:          []any{"/rules/*.md"},
			wantErrorCount: 1,
			wantSeverity:   "warning",
		},
		{
			name:           "invalid glob syntax unmatched bracket",
			rules:          []any{"rules/[*.md"},
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "invalid glob syntax unbalanced brace",
			rules:          []any{"rules/{a,b.md"},
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "mixed valid and invalid",
			rules:          []any{"rules/*.md", "", "**/*.txt"},
			wantErrorCount: 1, // empty string
		},
		{
			name:           "mixed valid and absolute",
			rules:          []any{"rules/*.md", "/absolute/path.md"},
			wantErrorCount: 1, // absolute path warning
			wantSeverity:   "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateRules(tt.rules, "settings.json")
			if len(errs) != tt.wantErrorCount {
				t.Errorf("validateRules() error count = %d, want %d", len(errs), tt.wantErrorCount)
				for _, e := range errs {
					t.Logf("  - [%s] %s (source: %s)", e.Severity, e.Message, e.Source)
				}
			}
			if tt.wantSeverity != "" && len(errs) > 0 {
				if errs[0].Severity != tt.wantSeverity {
					t.Errorf("first error severity = %q, want %q", errs[0].Severity, tt.wantSeverity)
				}
			}
		})
	}
}

func TestValidateMatcherToolName(t *testing.T) {
	tests := []struct {
		name            string
		toolNamePattern string
		wantErrorCount  int
		wantSeverity    string // check first error severity if set
	}{
		{
			name:            "empty string returns nothing",
			toolNamePattern: "",
			wantErrorCount:  0,
		},
		{
			name:            "valid plain tool name",
			toolNamePattern: "Bash",
			wantErrorCount:  0,
		},
		{
			name:            "valid tool with glob pattern",
			toolNamePattern: "Bash(npm*)",
			wantErrorCount:  0,
		},
		{
			name:            "valid tool with complex glob",
			toolNamePattern: "Write(/src/**/*.go)",
			wantErrorCount:  0,
		},
		{
			name:            "valid Read tool",
			toolNamePattern: "Read",
			wantErrorCount:  0,
		},
		{
			name:            "valid Edit tool",
			toolNamePattern: "Edit",
			wantErrorCount:  0,
		},
		{
			name:            "valid MCP tool",
			toolNamePattern: "mcp__my_server_tool",
			wantErrorCount:  0,
		},
		{
			name:            "unknown tool name is suggestion",
			toolNamePattern: "FakeTool",
			wantErrorCount:  1,
			wantSeverity:    "suggestion",
		},
		{
			name:            "unknown tool with pattern is suggestion",
			toolNamePattern: "FakeTool(rm*)",
			wantErrorCount:  1,
			wantSeverity:    "suggestion",
		},
		{
			name:            "unclosed parenthesis",
			toolNamePattern: "Bash(npm*",
			wantErrorCount:  1,
			wantSeverity:    "error",
		},
		{
			name:            "empty glob inside parens",
			toolNamePattern: "Bash()",
			wantErrorCount:  1,
			wantSeverity:    "warning",
		},
		{
			name:            "invalid glob in parens",
			toolNamePattern: "Bash([invalid)",
			wantErrorCount:  1,
			wantSeverity:    "error",
		},
		{
			name:            "unknown tool AND unclosed paren",
			toolNamePattern: "FakeTool(pattern",
			wantErrorCount:  2, // suggestion + error
		},
		{
			name:            "unknown tool AND invalid glob",
			toolNamePattern: "FakeTool([bad)",
			wantErrorCount:  2, // suggestion + error
		},
		{
			name:            "valid Glob tool with star",
			toolNamePattern: "Glob(**/*.ts)",
			wantErrorCount:  0,
		},
		{
			name:            "valid Grep tool",
			toolNamePattern: "Grep",
			wantErrorCount:  0,
		},
		{
			name:            "valid Task tool with pattern",
			toolNamePattern: "Task(quality*)",
			wantErrorCount:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateMatcherToolName(tt.toolNamePattern, "test location", "test.json")
			if len(errs) != tt.wantErrorCount {
				t.Errorf("validateMatcherToolName(%q) error count = %d, want %d", tt.toolNamePattern, len(errs), tt.wantErrorCount)
				for _, e := range errs {
					t.Logf("  - [%s] %s (source: %s)", e.Severity, e.Message, e.Source)
				}
			}
			if tt.wantSeverity != "" && len(errs) > 0 {
				if errs[0].Severity != tt.wantSeverity {
					t.Errorf("first error severity = %q, want %q", errs[0].Severity, tt.wantSeverity)
				}
			}
		})
	}
}

func TestValidateSettingsSpecificWithRules(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]any
		wantErrorCount int
	}{
		{
			name: "valid rules array",
			data: map[string]any{
				"rules": []any{"rules/*.md", "**/*.md"},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid rules type",
			data: map[string]any{
				"rules": "not an array",
			},
			wantErrorCount: 1,
		},
		{
			name: "rules with absolute path",
			data: map[string]any{
				"rules": []any{"/etc/rules/*.md"},
			},
			wantErrorCount: 1,
		},
		{
			name: "rules with invalid glob",
			data: map[string]any{
				"rules": []any{"rules/[*.md"},
			},
			wantErrorCount: 1,
		},
		{
			name:           "no rules key is valid",
			data:           map[string]any{},
			wantErrorCount: 0,
		},
		{
			name: "rules combined with hooks",
			data: map[string]any{
				"rules": []any{"rules/*.md"},
				"hooks": map[string]any{
					"PreToolUse": []any{
						map[string]any{
							"matcher": map[string]any{
								"toolName": "Bash(npm*)",
							},
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
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateSettingsSpecific(tt.data, "settings.json")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateSettingsSpecific() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - [%s] %s (source: %s)", err.Severity, err.Message, err.Source)
				}
			}
		})
	}
}
