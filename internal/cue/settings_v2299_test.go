package cue

import (
	"testing"
)

// TestValidateSettings_v2299Through2308Fields tests the settings fields added
// across v2.1.299–v2.1.308: idle compaction toggles, sandbox user-namespace /
// PTY / git-config / nested-sandbox relaxations, the managed ripgrep path,
// sandbox.network Unix-socket / localhost-binding / mach-lookup allowances,
// hook model overrides, single-name enabledMcpjsonServers / disabledMcpjsonServers,
// and the managed-only forceLoginOrgUUID gate.
func TestValidateSettings_v2299Through2308Fields(t *testing.T) {
	t.Parallel()

	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name: "valid idleCompaction true",
			data: map[string]any{
				"idleCompaction": true,
			},
		},
		{
			name: "valid precomputeCompactionEnabled false",
			data: map[string]any{
				"precomputeCompactionEnabled": false,
			},
		},
		{
			name: "valid autoContinueAtUsageLimit true",
			data: map[string]any{
				"autoContinueAtUsageLimit": true,
			},
		},
		{
			name: "valid sandbox namespace pty gitconfig and relaxations",
			data: map[string]any{
				"sandbox": map[string]any{
					"allowNestedUserNamespaces":    true,
					"allowPty":                     true,
					"allowGitConfig":               false,
					"enableWeakerNestedSandbox":    true,
					"enableWeakerNetworkIsolation": false,
					"ripgrep":                      "/opt/bin/rg",
				},
			},
		},
		{
			name: "valid sandbox network socket binding and mach lookup",
			data: map[string]any{
				"sandbox": map[string]any{
					"network": map[string]any{
						"allowAllUnixSockets": true,
						"allowLocalBinding":   false,
						"allowMachLookup":     "com.apple.*",
					},
				},
			},
		},
		{
			name: "valid prompt hook with model override",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"hooks": []map[string]any{
								{
									"type":    "prompt",
									"prompt":  "Review this output",
									"if":      "Bash(git *)",
									"timeout": 30,
									"model":   "haiku",
								},
							},
						},
					},
				},
			},
			wantError: false,
		},
		{
			name: "valid enabledMcpjsonServers single string",
			data: map[string]any{
				"enabledMcpjsonServers": "filesystem",
			},
		},
		{
			name: "valid disabledMcpjsonServers single string",
			data: map[string]any{
				"disabledMcpjsonServers": "github",
			},
		},
		{
			name: "valid forceLoginOrgUUID string",
			data: map[string]any{
				"forceLoginOrgUUID": "some-org-uuid",
			},
		},
		{
			name:      "reject idleCompaction wrong type",
			data:      map[string]any{"idleCompaction": "yes"},
			wantError: true,
		},
		{
			name:      "reject allowNestedUserNamespaces wrong type",
			data:      map[string]any{"sandbox": map[string]any{"allowNestedUserNamespaces": "true"}},
			wantError: true,
		},
		{
			name: "reject hook model wrong type",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"hooks": []map[string]any{
								{
									"type":   "prompt",
									"prompt": "Review this output",
									"model":  5,
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name:      "reject enabledMcpjsonServers wrong type",
			data:      map[string]any{"enabledMcpjsonServers": 42},
			wantError: true,
		},
		{
			name:      "reject forceLoginOrgUUID wrong type",
			data:      map[string]any{"forceLoginOrgUUID": true},
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
