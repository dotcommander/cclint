package cue

import (
	"testing"
)

// TestValidateSettings_v2156Fields tests fields added in v2.1.156:
// hook async-rewake controls, autoMode allow/soft_deny/deny/environment,
// and autoMemoryEnabled/autoDreamEnabled toggles.
func TestValidateSettings_v2156Fields(t *testing.T) {
	t.Parallel()

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
			name: "valid hook asyncRewake with rewakeMessage and rewakeSummary",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"matcher": "Bash",
							"hooks": []map[string]any{
								{
									"type":          "command",
									"command":       "echo done",
									"async":         true,
									"asyncRewake":   true,
									"rewakeMessage": "Background check finished:",
									"rewakeSummary": "lint passed",
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid autoMode with all classifier lists",
			data: map[string]any{
				"autoMode": map[string]any{
					"allow":       []string{"Bash(git status)"},
					"soft_deny":   []string{"Bash(rm *)"},
					"hard_deny":   []string{"Bash(rm -rf *)"},
					"deny":        []string{"Bash(curl * | sh)"},
					"environment": []string{"macOS dev laptop"},
				},
			},
			wantError: false,
		},
		{
			name: "valid autoMemoryEnabled and autoDreamEnabled",
			data: map[string]any{
				"autoMemoryEnabled": true,
				"autoDreamEnabled":  false,
			},
			wantError: false,
		},
		{
			name: "invalid autoMemoryEnabled wrong type",
			data: map[string]any{
				"autoMemoryEnabled": "yes",
			},
			wantError: true,
		},
		{
			name: "invalid asyncRewake wrong type",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"matcher": "Bash",
							"hooks": []map[string]any{
								{
									"type":        "command",
									"command":     "echo done",
									"asyncRewake": "true",
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs, err := v.ValidateSettings(tt.data)
			if err != nil {
				t.Fatalf("ValidateSettings returned error: %v", err)
			}
			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateSettings() hasErrors = %v, want %v", hasErrors, tt.wantError)
				for _, e := range errs {
					t.Logf("  Error: %s", e.Message)
				}
			}
		})
	}
}

// TestValidateAgent_v2156Frontmatter tests the requiredMcpServers and
// criticalSystemReminder_EXPERIMENTAL agent frontmatter fields added in v2.1.156.
func TestValidateAgent_v2156Frontmatter(t *testing.T) {
	t.Parallel()

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
			name: "valid requiredMcpServers array",
			data: map[string]any{
				"name":               "test-agent",
				"description":        "Agent gated on MCP servers",
				"requiredMcpServers": []string{"filesystem", "github"},
			},
			wantError: false,
		},
		{
			name: "valid criticalSystemReminder_EXPERIMENTAL string",
			data: map[string]any{
				"name":                                "test-agent",
				"description":                         "Agent with critical reminder",
				"criticalSystemReminder_EXPERIMENTAL": "Always confirm before deleting.",
			},
			wantError: false,
		},
		{
			name: "invalid requiredMcpServers wrong type",
			data: map[string]any{
				"name":               "test-agent",
				"description":        "Agent with bad requiredMcpServers",
				"requiredMcpServers": "filesystem",
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs, err := v.ValidateAgent(tt.data)
			if err != nil {
				t.Fatalf("ValidateAgent returned error: %v", err)
			}
			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateAgent() hasErrors = %v, want %v", hasErrors, tt.wantError)
				for _, e := range errs {
					t.Logf("  Error: %s", e.Message)
				}
			}
		})
	}
}
