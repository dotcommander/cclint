package cue

import (
	"testing"
)

func TestValidateSettings_v2252Through2260Fields(t *testing.T) {
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
			name: "valid timeFormat preset",
			data: map[string]any{
				"timeFormat": "24-hour-utc",
				"timeZone":   "UTC",
			},
		},
		{
			name: "valid strftime pattern",
			data: map[string]any{
				"timeFormat": "%Y-%m-%d %H:%M",
				"timeZone":   "America/New_York",
			},
		},
		{
			name: "valid permissions read block and mode",
			data: map[string]any{
				"permissions": map[string]any{
					"defaultMode":                         "acceptEdits",
					"blockReadsOutsideWorkingDirectories": true,
				},
			},
		},
		{
			name: "valid managed HTTP MCP server",
			data: map[string]any{
				"managedMcpServers": map[string]any{
					"company-tools": map[string]any{
						"type": "http",
						"url":  "https://example.com/mcp",
					},
				},
			},
		},
		{
			name: "valid managed SSE MCP server with headers",
			data: map[string]any{
				"managedMcpServers": map[string]any{
					"company-tools": map[string]any{
						"type":    "sse",
						"url":     "https://example.com/sse",
						"headers": map[string]any{"X-Example": "value"},
					},
				},
			},
		},
		{
			name:      "reject timeFormat wrong type",
			data:      map[string]any{"timeFormat": 24},
			wantError: true,
		},
		{
			name:      "reject timeZone wrong type",
			data:      map[string]any{"timeZone": false},
			wantError: true,
		},
		{
			name:      "reject managed MCP command entry",
			data:      map[string]any{"managedMcpServers": map[string]any{"bad": map[string]any{"type": "http", "url": "https://example.com", "command": "run"}}},
			wantError: true,
		},
		{
			name:      "reject managed MCP unsupported transport",
			data:      map[string]any{"managedMcpServers": map[string]any{"bad": map[string]any{"type": "stdio", "url": "https://example.com"}}},
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
