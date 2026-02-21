package cli

import (
	"testing"
)

func TestValidateHooks(t *testing.T) {
	tests := []struct {
		name           string
		hooks          any
		wantErrorCount int
	}{
		{
			name:           "nil hooks",
			hooks:          nil,
			wantErrorCount: 1,
		},
		{
			name:           "invalid type",
			hooks:          "not an object",
			wantErrorCount: 1,
		},
		{
			name: "unknown event",
			hooks: map[string]any{
				"UnknownEvent": []any{},
			},
			wantErrorCount: 1,
		},
		{
			name: "invalid hook configuration",
			hooks: map[string]any{
				"PreToolUse": "not an array",
			},
			wantErrorCount: 1,
		},
		{
			name: "missing matcher",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"hooks": []any{},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "missing hooks field",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "invalid hooks type",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks":   "not an array",
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "missing hook type",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "invalid hook type value",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "invalid",
							},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "command hook missing command",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "command",
							},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "prompt hook on unsupported event",
			hooks: map[string]any{
				"PostToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "prompt",
							},
						},
					},
				},
			},
			wantErrorCount: 2, // unsupported event + missing prompt field
		},
		{
			name: "prompt hook missing prompt",
			hooks: map[string]any{
				"Stop": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "prompt",
							},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "valid command hook",
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
			name: "valid prompt hook",
			hooks: map[string]any{
				"Stop": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":   "prompt",
								"prompt": "Test prompt",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid Setup hook with init matcher",
			hooks: map[string]any{
				"Setup": []any{
					map[string]any{
						"matcher": "init",
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo setup init",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid Setup hook with maintenance matcher",
			hooks: map[string]any{
				"Setup": []any{
					map[string]any{
						"matcher": "maintenance",
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo setup maintenance",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid SubagentStart hook",
			hooks: map[string]any{
				"SubagentStart": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo subagent started",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid PostToolUseFailure hook",
			hooks: map[string]any{
				"PostToolUseFailure": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo tool failed",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid TeammateIdle hook",
			hooks: map[string]any{
				"TeammateIdle": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo teammate idle",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid TaskCompleted hook",
			hooks: map[string]any{
				"TaskCompleted": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo task done",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid ConfigChange hook",
			hooks: map[string]any{
				"ConfigChange": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo config changed",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid WorktreeCreate hook",
			hooks: map[string]any{
				"WorktreeCreate": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo worktree created",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid WorktreeRemove hook",
			hooks: map[string]any{
				"WorktreeRemove": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo worktree removed",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid agent hook",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "agent",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "command hook with async true",
			hooks: map[string]any{
				"PostToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo async test",
								"async":   true,
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid hook type rejected",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "webhook",
							},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateHooks(tt.hooks, "settings.json")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateHooks() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - %s: %s", err.Severity, err.Message)
				}
			}
		})
	}
}

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
				"Setup": []any{
					map[string]any{
						"matcher": "init",
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
			name: "accepts PreToolUse",
			hooks: map[string]any{
				"PreToolUse": validHook,
			},
			wantErrorCount: 0,
		},
		{
			name: "accepts PostToolUse",
			hooks: map[string]any{
				"PostToolUse": validHook,
			},
			wantErrorCount: 0,
		},
		{
			name: "accepts Stop",
			hooks: map[string]any{
				"Stop": validHook,
			},
			wantErrorCount: 0,
		},
		{
			name: "rejects SessionStart",
			hooks: map[string]any{
				"SessionStart": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects Setup",
			hooks: map[string]any{
				"Setup": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects SubagentStart",
			hooks: map[string]any{
				"SubagentStart": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects TeammateIdle for components",
			hooks: map[string]any{
				"TeammateIdle": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects TaskCompleted for components",
			hooks: map[string]any{
				"TaskCompleted": validHook,
			},
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
