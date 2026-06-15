package cue

import (
	"testing"
)

// TestSettingsV2177Fields tests settings fields added in v2.1.174 through v2.1.176:
// wheelScrollAccelerationEnabled, enforceAvailableModels, and footerLinksRegexes.
func TestSettingsV2177Fields(t *testing.T) {
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
			name: "valid wheelScrollAccelerationEnabled bool",
			data: map[string]any{
				"wheelScrollAccelerationEnabled": true,
			},
			wantError: false,
		},
		{
			name: "invalid wheelScrollAccelerationEnabled as string",
			data: map[string]any{
				"wheelScrollAccelerationEnabled": "yes",
			},
			wantError: true,
		},
		{
			name: "valid enforceAvailableModels bool",
			data: map[string]any{
				"enforceAvailableModels": true,
			},
			wantError: false,
		},
		{
			name: "valid availableModels array of strings",
			data: map[string]any{
				"availableModels": []any{"opus", "sonnet", "claude-opus-4-5"},
			},
			wantError: false,
		},
		{
			name: "invalid availableModels as plain string",
			data: map[string]any{
				"availableModels": "opus",
			},
			wantError: true,
		},
		{
			name: "valid footerLinksRegexes array of objects",
			data: map[string]any{
				"footerLinksRegexes": []any{map[string]any{"type": "regex", "pattern": "PR #(\\d+)", "url": "https://example.com/{1}", "label": "PR"}},
			},
			wantError: false,
		},
		{
			name: "invalid footerLinksRegexes as plain string",
			data: map[string]any{
				"footerLinksRegexes": "nope",
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
