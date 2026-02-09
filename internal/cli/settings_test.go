package cli

import (
	"testing"
)

func TestLintSettings(t *testing.T) {
	// Test with empty directory
	summary, err := LintSettings("testdata/empty", false, false, true)
	if err != nil {
		t.Fatalf("LintSettings() error = %v", err)
	}
	if summary == nil {
		t.Fatal("LintSettings() returned nil summary")
	}
}

func TestValidateSettingsSpecific(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]any
		wantErrorCount int
	}{
		{
			name:           "no hooks",
			data:           map[string]any{},
			wantErrorCount: 0,
		},
		{
			name: "valid hooks",
			data: map[string]any{
				"hooks": map[string]any{
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
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid hooks",
			data: map[string]any{
				"hooks": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "valid permissions with allow and deny",
			data: map[string]any{
				"permissions": map[string]any{
					"allow": []any{"Bash(npm*)", "Read", "Edit"},
					"deny":  []any{"Bash(rm*)"},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid permissions structure",
			data: map[string]any{
				"permissions": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "valid hooks and valid permissions together",
			data: map[string]any{
				"hooks": map[string]any{
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
				"permissions": map[string]any{
					"allow": []any{"Bash(npm*)"},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid mcpServers with command and args",
			data: map[string]any{
				"mcpServers": map[string]any{
					"my-server": map[string]any{
						"command": "npx",
						"args":    []any{"-y", "@modelcontextprotocol/server-filesystem"},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid mcpServers structure",
			data: map[string]any{
				"mcpServers": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "mcpServers missing command",
			data: map[string]any{
				"mcpServers": map[string]any{
					"bad-server": map[string]any{
						"args": []any{"--flag"},
					},
				},
			},
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateSettingsSpecific(tt.data, "settings.json")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateSettingsSpecific() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - %s: %s", err.Severity, err.Message)
				}
			}
		})
	}
}
