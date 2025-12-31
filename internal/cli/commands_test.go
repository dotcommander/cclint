package cli

import (
	"strings"
	"testing"
)

func TestValidateCommandSpecific(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]interface{}
		filePath      string
		contents      string
		wantErrCount  int
		wantSuggCount int
	}{
		{
			name: "valid command",
			data: map[string]interface{}{
				"name":          "test-cmd",
				"description":   "Test command",
				"allowed-tools": "Task",
			},
			filePath:      "commands/test-cmd.md",
			contents:      "---\nname: test-cmd\ndescription: Test command\nallowed-tools: Task\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "invalid name format",
			data: map[string]interface{}{
				"name": "TestCmd",
			},
			filePath:      "commands/test.md",
			contents:      "---\nname: TestCmd\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "unknown field",
			data: map[string]interface{}{
				"name": "test",
				"foo":  "bar",
			},
			filePath:      "commands/test.md",
			contents:      "---\nname: test\nfoo: bar\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "valid without name (derived from filename)",
			data: map[string]interface{}{
				"description": "Test",
			},
			filePath:      "commands/test.md",
			contents:      "---\ndescription: Test\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateCommandSpecific(tt.data, tt.filePath, tt.contents)

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
				t.Errorf("validateCommandSpecific() errors = %d, want %d", errCount, tt.wantErrCount)
				for _, e := range errors {
					if e.Severity == "error" {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
			if suggCount != tt.wantSuggCount {
				t.Errorf("validateCommandSpecific() suggestions = %d, want %d", suggCount, tt.wantSuggCount)
			}
		})
	}
}

func TestValidateCommandBestPractices(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		contents     string
		data         map[string]interface{}
		wantContains []string
	}{
		{
			name:     "XML tags in description",
			filePath: "commands/test.md",
			contents: "---\ndescription: Test <tag>content</tag>\n---\n",
			data: map[string]interface{}{
				"description": "Test <tag>content</tag>",
			},
			wantContains: []string{"XML-like tags"},
		},
		{
			name:         "implementation section",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\n## Implementation\nSteps here",
			data:         map[string]interface{}{"name": "test"},
			wantContains: []string{"implementation steps"},
		},
		{
			name:         "Task() without allowed-tools",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\nTask(agent): do something",
			data:         map[string]interface{}{"name": "test"},
			wantContains: []string{"allowed-tools"},
		},
		{
			name:         "bloat sections in thin command",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\nTask(agent): do\n## Quick Reference\n",
			data:         map[string]interface{}{"name": "test"},
			wantContains: []string{"Quick Reference"},
		},
		{
			name:         "excessive examples",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\n```bash\nfoo\n```\n```bash\nbar\n```\n```bash\nbaz\n```",
			data:         map[string]interface{}{"name": "test"},
			wantContains: []string{"code examples"},
		},
		{
			name:         "success criteria without checkboxes",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\n## Success\nAll tests pass",
			data:         map[string]interface{}{"name": "test"},
			wantContains: []string{"checkbox format"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validateCommandBestPractices(tt.filePath, tt.contents, tt.data)

			for _, want := range tt.wantContains {
				found := false
				for _, sugg := range suggestions {
					if strings.Contains(sugg.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateCommandBestPractices() should contain suggestion about %q", want)
					for _, s := range suggestions {
						t.Logf("  Got: %s", s.Message)
					}
				}
			}
		})
	}
}

func TestKnownCommandFields(t *testing.T) {
	expected := []string{"name", "description", "allowed-tools", "argument-hint", "model", "disable-model-invocation"}
	for _, field := range expected {
		if !knownCommandFields[field] {
			t.Errorf("knownCommandFields missing expected field: %s", field)
		}
	}
}
