package lint

import (
	"testing"
)

// TestValidateAgentSpecific covers required-field, name, color, unknown-field,
// and missing-PROACTIVELY cases. Optional-field cases (memory, permissionMode,
// mcpServers, isolation, background, effort, initialPrompt, requiredMcpServers)
// live in agent_optional_fields_test.go. Model/maxTurns live in agent_model_test.go.
func TestValidateAgentSpecific(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]any
		filePath      string
		contents      string
		wantErrCount  int
		wantSuggCount int
	}{
		{
			name: "valid agent",
			data: map[string]any{
				"name":        "test-agent",
				"description": "A test agent. Use PROACTIVELY when testing.",
				"model":       "sonnet",
			},
			filePath:      "agents/test-agent.md",
			contents:      "---\nname: test-agent\ndescription: A test agent. Use PROACTIVELY when testing.\nmodel: sonnet\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name:          "missing name",
			data:          map[string]any{"description": "test. Use PROACTIVELY when testing."},
			filePath:      "agents/test.md",
			contents:      "---\ndescription: test. Use PROACTIVELY when testing.\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name:          "missing description",
			data:          map[string]any{"name": "test"},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "invalid name format",
			data: map[string]any{
				"name":        "TestAgent",
				"description": "test. Use PROACTIVELY when testing.",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: TestAgent\ndescription: test. Use PROACTIVELY when testing.\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "reserved word",
			data: map[string]any{
				"name":        "claude",
				"description": "test. Use PROACTIVELY when testing.",
			},
			filePath:      "agents/claude.md",
			contents:      "---\nname: claude\ndescription: test. Use PROACTIVELY when testing.\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "name doesn't match filename",
			data: map[string]any{
				"name":        "test-agent",
				"description": "test. Use PROACTIVELY when testing.",
			},
			filePath:      "agents/other-name.md",
			contents:      "---\nname: test-agent\ndescription: test. Use PROACTIVELY when testing.\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "invalid color",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"color":       "rainbow",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\ncolor: rainbow\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "unknown field",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"foo":         "bar",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nfoo: bar\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "missing PROACTIVELY in description",
			data: map[string]any{
				"name":        "test",
				"description": "A test agent for doing things.",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: A test agent for doing things.\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
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
