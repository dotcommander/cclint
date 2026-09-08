package cue

import (
	"strings"
	"testing"
)

func TestModelUnionGenerated(t *testing.T) {
	t.Parallel()

	union := modelUnionCUE()
	if !strings.HasPrefix(union, "#Model:") {
		t.Fatalf("modelUnionCUE() must start with #Model:, got %q", union)
	}
	for _, alias := range modelAliases {
		quoted := `"` + alias + `"`
		if !strings.Contains(union, quoted) {
			t.Errorf("modelUnionCUE() missing alias %s in %q", quoted, union)
		}
	}
	if !strings.Contains(union, claudeModelRegexCUE) {
		t.Errorf("modelUnionCUE() missing regex branch %s in %q", claudeModelRegexCUE, union)
	}
}

func TestClaudeMDAcceptsFullModelID(t *testing.T) {
	t.Parallel()

	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatalf("LoadSchemas: %v", err)
	}

	errs, err := v.ValidateClaudeMD(map[string]any{
		"model": "claude-opus-4-5",
	})
	if err != nil {
		t.Fatalf("ValidateClaudeMD(full model ID): unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("ValidateClaudeMD(full model ID): expected no errors, got %d: %v", len(errs), errs)
	}

	errs, err = v.ValidateAgent(map[string]any{
		"name":        "test-agent",
		"description": "test",
		"model":       "opus",
	})
	if err != nil {
		t.Fatalf("ValidateAgent(valid alias): unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("ValidateAgent(valid alias): expected no errors, got %d: %v", len(errs), errs)
	}

	errs, err = v.ValidateAgent(map[string]any{
		"name":        "test-agent",
		"description": "test",
		"model":       "claude-opus-5",
	})
	if err != nil {
		t.Fatalf("ValidateAgent(claude-opus-5): unexpected error: %v", err)
	}
	if len(errs) != 0 {
		t.Fatalf("ValidateAgent(claude-opus-5): expected no errors, got %d: %v", len(errs), errs)
	}

	errs, err = v.ValidateAgent(map[string]any{
		"name":        "test-agent",
		"description": "test",
		"model":       "notamodel",
	})
	if err != nil {
		t.Fatalf("ValidateAgent(invalid model): unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatalf("ValidateAgent(invalid model): expected a validation error, got none")
	}
}

func TestModelUnionAcceptsSuffixAliases(t *testing.T) {
	t.Parallel()

	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatalf("LoadSchemas: %v", err)
	}

	for _, model := range []string{"opusplan[1m]", "haiku[1m]", "best[1m]"} {
		errs, err := v.ValidateAgent(map[string]any{
			"name":        "test-agent",
			"description": "test",
			"model":       model,
		})
		if err != nil {
			t.Fatalf("model %q: unexpected error: %v", model, err)
		}
		if len(errs) != 0 {
			t.Errorf("model %q: unexpected validation errors: %v", model, errs)
		}
	}

	// Suffix is [1m]-only on aliases; other suffixes must still fail the union.
	errs, err := v.ValidateAgent(map[string]any{
		"name":        "test-agent",
		"description": "test",
		"model":       "opusplan[2m]",
	})
	if err != nil {
		t.Fatalf("opusplan[2m]: unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Error("opusplan[2m]: expected validation errors, got none")
	}
}
