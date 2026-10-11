package cue

import (
	"testing"
)

func TestValidateSettings_v2282Through2296Fields(t *testing.T) {
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
			name: "valid allowedProviders all members",
			data: map[string]any{
				"allowedProviders": []any{
					"anthropic", "customEndpoint", "bedrock", "vertex",
					"foundry", "anthropicAws", "mantle", "gateway",
				},
			},
		},
		{
			name: "valid availableModelsMatch prefix",
			data: map[string]any{
				"availableModelsMatch": "prefix",
			},
		},
		{
			name: "valid availableModelsMatch exact",
			data: map[string]any{
				"availableModelsMatch": "exact",
			},
		},
		{
			name: "valid deniedModels array",
			data: map[string]any{
				"deniedModels": []any{"opus", "claude-opus-5"},
			},
		},
		{
			name: "valid maxProseWidth 100",
			data: map[string]any{
				"maxProseWidth": 100,
			},
		},
		{
			name: "valid allowClaudeInChromeWithManagedMcp true",
			data: map[string]any{
				"allowClaudeInChromeWithManagedMcp": true,
			},
		},
		{
			name: "valid attribution false",
			data: map[string]any{
				"attribution": false,
			},
		},
		{
			name: "valid attribution object",
			data: map[string]any{
				"attribution": map[string]any{"commit": "x"},
			},
		},
		{
			name: "valid modelSettings autoCompactWindow auto",
			data: map[string]any{
				"modelSettings": map[string]any{
					"sonnet": map[string]any{"autoCompactWindow": "auto"},
				},
			},
		},
		{
			name: "valid modelSettings autoCompactWindow 200000",
			data: map[string]any{
				"modelSettings": map[string]any{
					"sonnet": map[string]any{"autoCompactWindow": 200000},
				},
			},
		},
		{
			name:      "reject allowedProviders invalid member openai",
			data:      map[string]any{"allowedProviders": []any{"openai"}},
			wantError: true,
		},
		{
			name:      "reject availableModelsMatch loose",
			data:      map[string]any{"availableModelsMatch": "loose"},
			wantError: true,
		},
		{
			name:      "reject deniedModels bare string",
			data:      map[string]any{"deniedModels": "opus"},
			wantError: true,
		},
		{
			name:      "reject maxProseWidth below min",
			data:      map[string]any{"maxProseWidth": 39},
			wantError: true,
		},
		{
			name:      "reject maxProseWidth non-int string",
			data:      map[string]any{"maxProseWidth": "100"},
			wantError: true,
		},
		{
			name:      "reject attribution number",
			data:      map[string]any{"attribution": 42},
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
