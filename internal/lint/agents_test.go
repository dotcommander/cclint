package lint

import (
	"strings"
	"testing"
)

func TestHasEditingTools(t *testing.T) {
	tests := []struct {
		name  string
		tools any
		want  bool
	}{
		{"wildcard", "*", true},
		{"string with Edit", "Read, Edit, Bash", true},
		{"string with Write", "Write", true},
		{"string with MultiEdit", "MultiEdit", true},
		{"string without editing", "Read, Bash", false},
		{"array with Edit", []any{"Read", "Edit"}, true},
		{"array without editing", []any{"Read", "Bash"}, false},
		{"empty string", "", false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasEditingTools(tt.tools)
			if got != tt.want {
				t.Errorf("hasEditingTools(%v) = %v, want %v", tt.tools, got, tt.want)
			}
		})
	}
}

func TestKnownAgentFields(t *testing.T) {
	expected := []string{"name", "description", "model", "color", "tools", "disallowedTools", "permissionMode", "maxTurns", "effort", "initialPrompt", "skills", "hooks", "memory", "mcpServers", "isolation", "background", "requiredMcpServers", "criticalSystemReminder_EXPERIMENTAL"}
	for _, field := range expected {
		if !knownAgentFields[field] {
			t.Errorf("knownAgentFields missing expected field: %s", field)
		}
	}
}

// TestValidateAgentNoUnknownFieldSuggestion guards against the v2.1.156 false
// positive where valid agent frontmatter keys were flagged as unknown.
func TestValidateAgentNoUnknownFieldSuggestion(t *testing.T) {
	fields := []string{"requiredMcpServers", "criticalSystemReminder_EXPERIMENTAL"}
	data := map[string]any{
		"name":                                "test",
		"description":                         "test. Use PROACTIVELY when testing.",
		"requiredMcpServers":                  []any{"filesystem"},
		"criticalSystemReminder_EXPERIMENTAL": "Always confirm before deleting.",
	}
	contents := "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nrequiredMcpServers:\n  - filesystem\ncriticalSystemReminder_EXPERIMENTAL: Always confirm before deleting.\n---\n"

	errors := validateAgentSpecific(data, "agents/test.md", contents)
	for _, e := range errors {
		for _, f := range fields {
			if strings.Contains(e.Message, "Unknown frontmatter field") && strings.Contains(e.Message, f) {
				t.Errorf("field %q should be known but got suggestion: %s", f, e.Message)
			}
		}
	}
}
