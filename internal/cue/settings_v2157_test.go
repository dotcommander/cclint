package cue

import (
	"testing"
)

// TestValidateSettings_v2157Fields tests fields added in v2.1.157:
// the `agent` string field naming the default subagent for dispatched sessions.
func TestValidateSettings_v2157Fields(t *testing.T) {
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
			name: "valid agent string",
			data: map[string]any{
				"agent": "code-reviewer",
			},
			wantError: false,
		},
		{
			name: "invalid agent wrong type (number)",
			data: map[string]any{
				"agent": 42,
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
