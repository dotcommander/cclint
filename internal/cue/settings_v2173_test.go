package cue

import (
	"testing"
)

// TestSettingsV2173Fields tests settings fields added in v2.1.163 through v2.1.169:
// fallbackModel, requiredMinimumVersion, requiredMaximumVersion, and disableBundledSkills.
func TestSettingsV2173Fields(t *testing.T) {
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
			name: "valid fallbackModel array",
			data: map[string]any{
				"fallbackModel": []any{"sonnet", "opus[1m]", "default"},
			},
			wantError: false,
		},
		{
			name: "invalid fallbackModel as plain string",
			data: map[string]any{
				"fallbackModel": "sonnet",
			},
			wantError: true,
		},
		{
			name: "valid requiredMinimumVersion and requiredMaximumVersion strings",
			data: map[string]any{
				"requiredMinimumVersion": "2.1.163",
				"requiredMaximumVersion": "2.1.173",
			},
			wantError: false,
		},
		{
			name: "invalid requiredMinimumVersion as number",
			data: map[string]any{
				"requiredMinimumVersion": 2,
			},
			wantError: true,
		},
		{
			name: "valid disableBundledSkills bool",
			data: map[string]any{
				"disableBundledSkills": true,
			},
			wantError: false,
		},
		{
			name: "invalid disableBundledSkills as string",
			data: map[string]any{
				"disableBundledSkills": "yes",
			},
			wantError: true,
		},
		{
			name: "valid unrelated settings key remains accepted",
			data: map[string]any{
				"language": "en",
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
