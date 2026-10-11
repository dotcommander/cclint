package cue

import (
	"testing"
)

func TestClaudeMDNullTitleDescription(t *testing.T) {
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
			name: "null title and description pass",
			data: map[string]any{
				"title":       nil,
				"description": nil,
			},
			wantError: false,
		},
		{
			name: "non-string title rejected",
			data: map[string]any{
				"title": 123,
			},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs, err := v.ValidateClaudeMD(tt.data)
			if err != nil {
				t.Fatalf("ValidateClaudeMD returned error: %v", err)
			}
			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateClaudeMD() hasErrors = %v, want %v", hasErrors, tt.wantError)
				for _, e := range errs {
					t.Logf("  Error: %s", e.Message)
				}
			}
		})
	}
}
