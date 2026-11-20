package cue

import (
	"testing"
)

// TestValidateSettings_v2326Through2329Fields tests the settings fields added
// across the v2.1.326–v2.1.329 sweep: the mcp_tool hook fields (server, tool,
// input) with the mcp_tool conditional and onFailure coverage, the
// sandbox.filesystem write/read path controls (v2.1.326+), the managed HTTP
// hook URL and env-var allowlists (binary-verified v2.1.328), and the
// attributionProject billing-group keys (v2.1.327).
func TestValidateSettings_v2326Through2329Fields(t *testing.T) {
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
			name: "valid mcp_tool hook full",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"hooks": []map[string]any{
								{
									"type":      "mcp_tool",
									"server":    "github",
									"tool":      "create_issue",
									"input":     "${tool_input.file_path}",
									"onFailure": "block",
									"timeout":   30,
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid mcp_tool hook object input",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"hooks": []map[string]any{
								{
									"type":   "mcp_tool",
									"server": "github",
									"tool":   "create_issue",
									"input":  map[string]any{"title": "x"},
								},
							},
						},
					},
				},
			},
		},
		{
			name: "valid mcp_tool hook minimal",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"hooks": []map[string]any{
								{
									"type":   "mcp_tool",
									"server": "github",
									"tool":   "create_issue",
								},
							},
						},
					},
				},
			},
		},
		{
			name: "reject mcp_tool missing server",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"hooks": []map[string]any{
								{
									"type": "mcp_tool",
									"tool": "create_issue",
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "reject mcp_tool missing tool",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"hooks": []map[string]any{
								{
									"type":   "mcp_tool",
									"server": "github",
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "reject mcp_tool server wrong type",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"hooks": []map[string]any{
								{
									"type":   "mcp_tool",
									"server": 42,
									"tool":   "x",
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "reject mcp_tool invalid onFailure",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUse": []map[string]any{
						{
							"hooks": []map[string]any{
								{
									"type":      "mcp_tool",
									"server":    "s",
									"tool":      "t",
									"onFailure": "retry",
								},
							},
						},
					},
				},
			},
			wantError: true,
		},
		{
			name: "valid attributionProject",
			data: map[string]any{
				"attributionProject": "claude_proj_abc123",
			},
		},
		{
			name:      "reject attributionProject wrong type",
			data:      map[string]any{"attributionProject": 42},
			wantError: true,
		},
		{
			name: "valid attributionProjectPolicy",
			data: map[string]any{
				"attributionProjectPolicy": map[string]any{"default": "claude_proj_x"},
			},
		},
		{
			name:      "reject attributionProjectPolicy wrong type",
			data:      map[string]any{"attributionProjectPolicy": map[string]any{"default": 42}},
			wantError: true,
		},
		{
			name: "valid sandbox filesystem controls",
			data: map[string]any{
				"sandbox": map[string]any{
					"filesystem": map[string]any{
						"allowWrite":                []any{"/tmp"},
						"denyWrite":                 []any{"/etc"},
						"denyRead":                  []any{"/secrets"},
						"allowRead":                 []any{"/secrets/public"},
						"allowManagedReadPathsOnly": true,
						"disabled":                  false,
					},
				},
			},
		},
		{
			name: "reject filesystem allowWrite wrong element",
			data: map[string]any{
				"sandbox": map[string]any{
					"filesystem": map[string]any{
						"allowWrite": []any{42},
					},
				},
			},
			wantError: true,
		},
		{
			name: "reject filesystem allowManagedReadPathsOnly wrong type",
			data: map[string]any{
				"sandbox": map[string]any{
					"filesystem": map[string]any{
						"allowManagedReadPathsOnly": "yes",
					},
				},
			},
			wantError: true,
		},
		{
			name: "valid http hook allowlists",
			data: map[string]any{
				"allowedHttpHookUrls":    []any{"https://example.com/*"},
				"httpHookAllowedEnvVars": []any{"TOKEN"},
			},
		},
		{
			name:      "reject allowedHttpHookUrls wrong type",
			data:      map[string]any{"allowedHttpHookUrls": "https://example.com"},
			wantError: true,
		},
		{
			name:      "reject httpHookAllowedEnvVars wrong element",
			data:      map[string]any{"httpHookAllowedEnvVars": []any{1}},
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
