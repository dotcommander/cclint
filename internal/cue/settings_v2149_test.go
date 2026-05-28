package cue

import (
	"testing"
)

// TestValidateSettings_v2149Fields tests the new managed-settings fields added in v2.1.149–v2.1.154.
func TestValidateSettings_v2149Fields(t *testing.T) {
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
			name: "valid allowAllClaudeAiMcps (v2.1.149+)",
			data: map[string]any{
				"allowAllClaudeAiMcps": true,
			},
			wantError: false,
		},
		{
			name: "valid pluginSuggestionMarketplaces (v2.1.152+)",
			data: map[string]any{
				"pluginSuggestionMarketplaces": []map[string]any{
					{"source": "github", "repo": "acme/approved-plugins"},
				},
			},
			wantError: false,
		},
		{
			name: "invalid pluginSuggestionMarketplaces bad source enum",
			data: map[string]any{
				"pluginSuggestionMarketplaces": []map[string]any{
					{"source": "ftp", "url": "ftp://example.com"},
				},
			},
			wantError: true,
		},
		{
			name: "valid allowedMcpServers (v2.1.154+)",
			data: map[string]any{
				"allowedMcpServers": []string{"mcp-server-*", "approved-mcp"},
			},
			wantError: false,
		},
		{
			name: "valid deniedMcpServers (v2.1.154+)",
			data: map[string]any{
				"deniedMcpServers": []string{"untrusted-*"},
			},
			wantError: false,
		},
		{
			name: "valid all four new fields together",
			data: map[string]any{
				"allowAllClaudeAiMcps": true,
				"pluginSuggestionMarketplaces": []map[string]any{
					{"source": "github", "repo": "acme/approved-plugins"},
				},
				"allowedMcpServers": []string{"mcp-server-*", "approved-mcp"},
				"deniedMcpServers":  []string{"untrusted-*"},
			},
			wantError: false,
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
