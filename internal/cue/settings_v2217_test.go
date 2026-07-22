package cue

import (
	"testing"
)

// TestValidateSettings_v2217Fields tests settings fields added in v2.1.205 through v2.1.217:
// emojiCompletionEnabled, axScreenReader, vimInsertModeRemaps, disableAutoMode,
// and sandbox.filesystem.disabled.
func TestValidateSettings_v2217Fields(t *testing.T) {
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
			name: "valid emojiCompletionEnabled (v2.1.217)",
			data: map[string]any{
				"emojiCompletionEnabled": true,
			},
			wantError: false,
		},
		{
			name: "valid axScreenReader (v2.1.208)",
			data: map[string]any{
				"axScreenReader": true,
			},
			wantError: false,
		},
		{
			name: "valid vimInsertModeRemaps map (v2.1.208)",
			data: map[string]any{
				"vimInsertModeRemaps": map[string]any{"jj": "Escape"},
			},
			wantError: false,
		},
		{
			name: "valid disableAutoMode disable (v2.1.207)",
			data: map[string]any{
				"disableAutoMode": "disable",
			},
			wantError: false,
		},
		{
			name: "valid sandbox.filesystem.disabled (v2.1.216)",
			data: map[string]any{
				"sandbox": map[string]any{"filesystem": map[string]any{"disabled": true}},
			},
			wantError: false,
		},
		{
			name: "reject disableAutoMode value outside enum",
			data: map[string]any{
				"disableAutoMode": "enable",
			},
			wantError: true,
		},
		{
			name: "reject emojiCompletionEnabled wrong type",
			data: map[string]any{
				"emojiCompletionEnabled": "yes",
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
