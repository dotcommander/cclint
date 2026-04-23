package lint

import (
	"testing"
)

func TestValidateHooksWithToolNameMatcher(t *testing.T) {
	tests := []struct {
		name           string
		hooks          any
		wantErrorCount int
	}{
		{
			name: "valid hook with known toolName in matcher",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{
							"toolName": "Bash(npm*)",
						},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo pre-tool",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "hook with unknown toolName in matcher produces suggestion",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{
							"toolName": "UnknownTool",
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
			wantErrorCount: 1, // suggestion for unknown tool
		},
		{
			name: "hook with unclosed paren in toolName matcher",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{
							"toolName": "Bash(npm*",
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
			wantErrorCount: 1, // error for unclosed paren
		},
		{
			name: "hook with string matcher skips toolName validation",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": "Bash",
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo setup",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "hook with empty matcher object skips toolName validation",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo test",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "hook with valid MCP toolName in matcher",
			hooks: map[string]any{
				"PostToolUse": []any{
					map[string]any{
						"matcher": map[string]any{
							"toolName": "mcp__my_server_tool",
						},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo post-tool",
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
			errors := validateHooks(tt.hooks, "settings.json")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateHooks() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - [%s] %s (source: %s)", err.Severity, err.Message, err.Source)
				}
			}
		})
	}
}

func TestValidateComponentHooks(t *testing.T) {
	validHook := []any{
		map[string]any{
			"matcher": map[string]any{},
			"hooks": []any{
				map[string]any{
					"type":    "command",
					"command": "echo test",
				},
			},
		},
	}

	tests := []struct {
		name           string
		hooks          any
		wantErrorCount int
	}{
		{
			name:           "accepts PreToolUse",
			hooks:          map[string]any{"PreToolUse": validHook},
			wantErrorCount: 0,
		},
		{
			name:           "accepts PostToolUse",
			hooks:          map[string]any{"PostToolUse": validHook},
			wantErrorCount: 0,
		},
		{
			name:           "accepts Stop",
			hooks:          map[string]any{"Stop": validHook},
			wantErrorCount: 0,
		},
		{
			name:           "rejects SessionStart",
			hooks:          map[string]any{"SessionStart": validHook},
			wantErrorCount: 1,
		},
		{
			name:           "rejects Setup",
			hooks:          map[string]any{"Setup": validHook},
			wantErrorCount: 1,
		},
		{
			name:           "rejects SubagentStart",
			hooks:          map[string]any{"SubagentStart": validHook},
			wantErrorCount: 1,
		},
		{
			name:           "rejects TeammateIdle for components",
			hooks:          map[string]any{"TeammateIdle": validHook},
			wantErrorCount: 1,
		},
		{
			name:           "rejects TaskCompleted for components",
			hooks:          map[string]any{"TaskCompleted": validHook},
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateComponentHooks(tt.hooks, "agent.md")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("ValidateComponentHooks() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - %s: %s", err.Severity, err.Message)
				}
			}
		})
	}
}
