package lint

import (
	"strings"
	"testing"
)

func TestValidateAgentMaxTurnsDontAskInfo(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantInfo bool
	}{
		{
			name: "maxTurns with dontAsk emits info",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"maxTurns":       10,
				"permissionMode": "dontAsk",
			},
			wantInfo: true,
		},
		{
			name: "maxTurns without dontAsk no info",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"maxTurns":       10,
				"permissionMode": "default",
			},
			wantInfo: false,
		},
		{
			name: "dontAsk without maxTurns no info",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"permissionMode": "dontAsk",
			},
			wantInfo: false,
		},
		{
			name: "neither maxTurns nor dontAsk no info",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
			},
			wantInfo: false,
		},
		{
			name: "maxTurns with bypassPermissions no info",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"maxTurns":       5,
				"permissionMode": "bypassPermissions",
			},
			wantInfo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents := "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\n---\n"
			errors := validateAgentSpecific(tt.data, "agents/test.md", contents)

			foundInfo := false
			for _, e := range errors {
				if e.Severity == "info" && strings.Contains(e.Message, "maxTurns with permissionMode 'dontAsk'") {
					foundInfo = true
					break
				}
			}

			if foundInfo != tt.wantInfo {
				t.Errorf("validateAgentSpecific() info about maxTurns+dontAsk = %v, want %v", foundInfo, tt.wantInfo)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestValidModelPattern(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"haiku", true},
		{"sonnet", true},
		{"opus", true},
		{"inherit", true},
		{"opusplan", true},
		{"sonnet[1m]", true},
		{"haiku[2]", true},
		{"opus[v3]", true},
		{"sonnet[latest]", true},
		{"unknown-model", false},
		{"turbo-3", false},
		{"", false},
		{"SONNET", false},
		{"Haiku", false},
		{"sonnet[]", false},
		{"fast", false},
		{"claude-opus-4-5", true},
		{"claude-sonnet-4-6", true},
		{"claude-haiku-4-5-20251001", true},
		{"claude-", false},
		{"claude-OPUS", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := validModelPattern.MatchString(tt.model)
			if got != tt.want {
				t.Errorf("validModelPattern.MatchString(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}
