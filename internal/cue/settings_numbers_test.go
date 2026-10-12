package cue

import (
	"testing"
)

// Regression: encoding/json decodes every JSON number to float64, and CUE
// treats float64(3) as the float 3.0, which cannot unify with `int` schema
// constraints. A settings.json hook with "timeout": 3 (a valid Claude Code
// value; the schema declares `timeout?: *30 | int`) failed with
// "conflicting values 3 and 30" / "mismatched types float and int".
// Integral JSON numbers must validate against int constraints.
func TestValidateSettingsIntegerJSONNumbers(t *testing.T) {
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
			name: "integer hook timeout",
			data: map[string]any{
				"hooks": map[string]any{
					"PostToolUseFailure": []any{
						map[string]any{
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "echo hi",
									"timeout": float64(3),
								},
							},
						},
					},
				},
			},
		},
		{
			name: "integer cleanupPeriodDays",
			data: map[string]any{
				"cleanupPeriodDays": float64(30),
			},
		},
		{
			name: "integer refreshInterval",
			data: map[string]any{
				"refreshInterval": float64(2000),
			},
		},
		{
			// Non-integral floats keep float semantics and still satisfy
			// number-typed fields such as modelPricing.multiplier.
			name: "float multiplier stays valid",
			data: map[string]any{
				"modelPricing": map[string]any{
					"multiplier": float64(1.5),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			errs, err := v.ValidateSettings(tt.data)
			if err != nil {
				t.Fatalf("ValidateSettings() error: %v", err)
			}
			for _, e := range errs {
				if e.Severity == "error" {
					t.Errorf("ValidateSettings() unexpected error: %s", e.Message)
				}
			}
			if tt.wantError && len(errs) == 0 {
				t.Error("ValidateSettings() expected an error, got none")
			}
		})
	}
}

// normalizeJSONNumbers must not mutate the input map: Go-side validators
// reuse encoding/json's float64 representation after CUE validation.
func TestNormalizeJSONNumbersDoesNotMutateInput(t *testing.T) {
	t.Parallel()

	inner := map[string]any{"timeout": float64(3)}
	input := map[string]any{"hooks": []any{inner}}

	got := normalizeJSONNumbers(input).(map[string]any)
	list := got["hooks"].([]any)
	if v, ok := list[0].(map[string]any)["timeout"].(int64); !ok || v != 3 {
		t.Errorf("normalized timeout = %#v, want int64(3)", list[0].(map[string]any)["timeout"])
	}
	if _, ok := inner["timeout"].(float64); !ok {
		t.Errorf("input map was mutated: timeout = %#v, want float64(3)", inner["timeout"])
	}
}
