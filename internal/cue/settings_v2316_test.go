package cue

import (
	"testing"
)

// TestValidateSettings_v2311Through2316Fields tests the settings fields added
// or backfilled across the v2.1.311–v2.1.316 sweep: the every-mode write block
// under permissions (v2.1.314) and the sandbox command-pattern exclusion list
// (pre-existing key, backfill surfaced by the v2.1.314 description fix).
func TestValidateSettings_v2311Through2316Fields(t *testing.T) {
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
			name: "valid blockWritesOutsideWorkingDirectories true",
			data: map[string]any{
				"permissions": map[string]any{
					"blockWritesOutsideWorkingDirectories": true,
				},
			},
		},
		{
			name: "valid read and write blocks together",
			data: map[string]any{
				"permissions": map[string]any{
					"blockReadsOutsideWorkingDirectories":  true,
					"blockWritesOutsideWorkingDirectories": false,
				},
			},
		},
		{
			name: "valid sandbox excludedCommands patterns",
			data: map[string]any{
				"sandbox": map[string]any{
					"excludedCommands": []any{"git status", "make *"},
				},
			},
		},
		{
			name: "valid sandbox excludedCommands empty",
			data: map[string]any{
				"sandbox": map[string]any{
					"excludedCommands": []any{},
				},
			},
		},
		{
			name:      "reject blockWritesOutsideWorkingDirectories wrong type",
			data:      map[string]any{"permissions": map[string]any{"blockWritesOutsideWorkingDirectories": "yes"}},
			wantError: true,
		},
		{
			name:      "reject excludedCommands wrong element type",
			data:      map[string]any{"sandbox": map[string]any{"excludedCommands": []any{42}}},
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
