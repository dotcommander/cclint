package cue

import (
	"testing"
)

func TestValidateSettings_v2261Through2265Fields(t *testing.T) {
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
			name: "valid bash and task output caps",
			data: map[string]any{
				"bashOutputMaxChars": 50000,
				"taskOutputMaxChars": 64000,
			},
		},
		{
			name: "valid forceLoginGatewayUrl",
			data: map[string]any{
				"forceLoginGatewayUrl": "https://gateway.example.com",
			},
		},
		{
			name:      "reject bashOutputMaxChars wrong type",
			data:      map[string]any{"bashOutputMaxChars": "50000"},
			wantError: true,
		},
		{
			name:      "reject taskOutputMaxChars wrong type",
			data:      map[string]any{"taskOutputMaxChars": true},
			wantError: true,
		},
		{
			name:      "reject forceLoginGatewayUrl wrong type",
			data:      map[string]any{"forceLoginGatewayUrl": 42},
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
