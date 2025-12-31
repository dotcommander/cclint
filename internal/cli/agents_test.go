package cli

import (
	"strings"
	"testing"
)

func TestValidateAgentSpecific(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]interface{}
		filePath      string
		contents      string
		wantErrCount  int
		wantSuggCount int
	}{
		{
			name: "valid agent",
			data: map[string]interface{}{
				"name":        "test-agent",
				"description": "A test agent",
				"model":       "sonnet",
			},
			filePath:      "agents/test-agent.md",
			contents:      "---\nname: test-agent\ndescription: A test agent\nmodel: sonnet\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name:          "missing name",
			data:          map[string]interface{}{"description": "test"},
			filePath:      "agents/test.md",
			contents:      "---\ndescription: test\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name:          "missing description",
			data:          map[string]interface{}{"name": "test"},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "invalid name format",
			data: map[string]interface{}{
				"name":        "TestAgent",
				"description": "test",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: TestAgent\ndescription: test\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "reserved word",
			data: map[string]interface{}{
				"name":        "claude",
				"description": "test",
			},
			filePath:      "agents/claude.md",
			contents:      "---\nname: claude\ndescription: test\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "name doesn't match filename",
			data: map[string]interface{}{
				"name":        "test-agent",
				"description": "test",
			},
			filePath:      "agents/other-name.md",
			contents:      "---\nname: test-agent\ndescription: test\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "invalid color",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test",
				"color":       "rainbow",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test\ncolor: rainbow\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "unknown field",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test",
				"foo":         "bar",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test\nfoo: bar\n---\n",
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

func TestHasEditingTools(t *testing.T) {
	tests := []struct {
		name  string
		tools interface{}
		want  bool
	}{
		{"wildcard", "*", true},
		{"string with Edit", "Read, Edit, Bash", true},
		{"string with Write", "Write", true},
		{"string with MultiEdit", "MultiEdit", true},
		{"string without editing", "Read, Bash", false},
		{"array with Edit", []interface{}{"Read", "Edit"}, true},
		{"array without editing", []interface{}{"Read", "Bash"}, false},
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

func TestValidateAgentBestPractices(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		contents     string
		data         map[string]interface{}
		wantContains []string
	}{
		{
			name:     "XML tags in description",
			filePath: "agents/test.md",
			contents: "---\ndescription: Test <xml>tag</xml>\n---\n",
			data: map[string]interface{}{
				"description": "Test <xml>tag</xml>",
			},
			wantContains: []string{"XML-like tags"},
		},
		{
			name:     "bloat section Quick Reference",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test\n---\n## Quick Reference\n",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test",
			},
			wantContains: []string{"Quick Reference"},
		},
		{
			name:     "missing model",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test\n---\n",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test",
			},
			wantContains: []string{"model"},
		},
		{
			name:     "no PROACTIVELY in description",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test description\n---\n",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test description",
			},
			wantContains: []string{"PROACTIVELY"},
		},
		{
			name:     "editing tools without permissionMode",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test\ntools: Edit\n---\n",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test",
				"tools":       "Edit",
			},
			wantContains: []string{"permissionMode"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validateAgentBestPractices(tt.filePath, tt.contents, tt.data)

			for _, want := range tt.wantContains {
				found := false
				for _, sugg := range suggestions {
					if strings.Contains(sugg.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateAgentBestPractices() should contain suggestion about %q", want)
					for _, s := range suggestions {
						t.Logf("  Got: %s", s.Message)
					}
				}
			}
		})
	}
}

func TestKnownAgentFields(t *testing.T) {
	expected := []string{"name", "description", "model", "color", "tools", "permissionMode", "skills"}
	for _, field := range expected {
		if !knownAgentFields[field] {
			t.Errorf("knownAgentFields missing expected field: %s", field)
		}
	}
}
