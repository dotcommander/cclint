package lint

import (
	"strings"
	"testing"
)

// TestValidateAgentSpecificModelFields covers model and maxTurns field cases
// exercised through validateAgentSpecific. Standalone model-pattern and
// maxTurns+dontAsk interaction tests follow below.
func TestValidateAgentSpecificModelFields(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]any
		filePath      string
		contents      string
		wantErrCount  int
		wantSuggCount int
	}{
		{
			name: "valid maxTurns positive integer",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"maxTurns":    10,
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmaxTurns: 10\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "invalid maxTurns zero",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"maxTurns":    0,
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmaxTurns: 0\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "invalid maxTurns negative",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"maxTurns":    -5,
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmaxTurns: -5\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "invalid maxTurns string",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"maxTurns":    "ten",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmaxTurns: ten\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "valid model sonnet",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "sonnet",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: sonnet\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model haiku",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "haiku",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: haiku\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model opus",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "opus",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: opus\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model inherit",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "inherit",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: inherit\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model with version suffix",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "sonnet[1m]",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: sonnet[1m]\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model opusplan",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "opusplan",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: opusplan\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateAgentSpecific(tt.data, tt.filePath, tt.contents)

			errCount := 0
			suggCount := 0
			for _, e := range errors {
				if e.Severity == "error" {
					errCount++
				} else if e.Severity == "suggestion" {
					suggCount++
				}
			}

			if errCount != tt.wantErrCount {
				t.Errorf("validateAgentSpecific() errors = %d, want %d", errCount, tt.wantErrCount)
				for _, e := range errors {
					if e.Severity == "error" {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
			if suggCount < tt.wantSuggCount {
				t.Errorf("validateAgentSpecific() suggestions = %d, want at least %d", suggCount, tt.wantSuggCount)
			}
		})
	}
}

func TestValidateAgentModelValues(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		wantWarnings int
	}{
		{"valid haiku", "haiku", 0},
		{"valid sonnet", "sonnet", 0},
		{"valid opus", "opus", 0},
		{"valid inherit", "inherit", 0},
		{"valid opusplan", "opusplan", 0},
		{"valid sonnet with version", "sonnet[1m]", 0},
		{"valid haiku with version", "haiku[2]", 0},
		{"valid opus with version", "opus[v3]", 0},
		{"invalid unknown-model", "unknown-model", 1},
		{"invalid empty", "", 1},
		{"invalid random", "fast", 1},
		{"invalid arbitrary string", "turbo-3", 1},
		{"invalid partial", "son", 1},
		{"valid full model claude-opus-4-5", "claude-opus-4-5", 0},
		{"valid full model claude-sonnet-4-6", "claude-sonnet-4-6", 0},
		{"valid full model with date", "claude-haiku-4-5-20251001", 0},
		{"invalid claude- bare prefix", "claude-", 1},
		{"invalid claude uppercase", "claude-OPUS", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       tt.model,
			}
			contents := "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: " + tt.model + "\n---\n"

			errors := validateAgentSpecific(data, "agents/test.md", contents)

			warnings := 0
			for _, e := range errors {
				if e.Severity == "warning" && strings.Contains(e.Message, "Unknown model") {
					warnings++
				}
			}

			if warnings != tt.wantWarnings {
				t.Errorf("validateAgentSpecific() model warnings = %d, want %d for model %q", warnings, tt.wantWarnings, tt.model)
				for _, e := range errors {
					if e.Severity == "warning" {
						t.Logf("  Warning: %s", e.Message)
					}
				}
			}
		})
	}
}
