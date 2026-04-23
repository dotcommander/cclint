package lint

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
		// valid-event and valid-type cases live in TestValidateHooksValidEvents (settings_hooks_extra_test.go)
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
