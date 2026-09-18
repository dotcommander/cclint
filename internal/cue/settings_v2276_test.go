package cue

import (
	"testing"
)

func TestValidateSettings_v2266Through2276Fields(t *testing.T) {
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
			name: "valid claude.ai account sync disabled",
			data: map[string]any{
				"syncClaudeAiSkills":  false,
				"syncClaudeAiPlugins": false,
			},
		},
		{
			name: "valid bashEditDiffEnabled false",
			data: map[string]any{
				"bashEditDiffEnabled": false,
			},
		},
		{
			name: "valid gatewayInternalNetworks",
			data: map[string]any{
				"gatewayInternalNetworks": []any{"203.0.113.0/24"},
			},
		},
		{
			name:      "reject gatewayInternalNetworks wrong type",
			data:      map[string]any{"gatewayInternalNetworks": "203.0.113.0/24"},
			wantError: true,
		},
		{
			name: "valid disableClaudeAiConnectors true",
			data: map[string]any{
				"disableClaudeAiConnectors": true,
			},
		},
		{
			name: "valid maxEffortLevel xhigh",
			data: map[string]any{
				"maxEffortLevel": "xhigh",
			},
		},
		{
			name:      "reject maxEffortLevel ultra",
			data:      map[string]any{"maxEffortLevel": "ultra"},
			wantError: true,
		},
		{
			name: "valid modelSettings",
			data: map[string]any{
				"modelSettings": map[string]any{
					"claude-opus-5": map[string]any{
						"effortLevel":    "high",
						"maxEffortLevel": "max",
					},
				},
			},
		},
		{
			name:      "reject modelSettings effortLevel max",
			data:      map[string]any{"modelSettings": map[string]any{"claude-opus-5": map[string]any{"effortLevel": "max"}}},
			wantError: true,
		},
		{
			name: "valid modelPricing multiplier 10",
			data: map[string]any{
				"modelPricing": map[string]any{
					"multiplier": 10,
				},
			},
		},
		{
			name:      "reject modelPricing multiplier 10.5",
			data:      map[string]any{"modelPricing": map[string]any{"multiplier": 10.5}},
			wantError: true,
		},
		{
			name:      "reject modelPricing multiplier 0",
			data:      map[string]any{"modelPricing": map[string]any{"multiplier": 0}},
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
