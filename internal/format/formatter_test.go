package format

import (
	"fmt"
	"strings"
	"testing"
)

func TestAgentFormatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "reorder frontmatter fields",
			input: `---
tools: [Read, Write]
model: sonnet
description: Test agent
name: test-agent
---

# Test Agent

Content here.
`,
			expected: `---
name: test-agent
description: Test agent
model: sonnet
tools:
  - Read
  - Write
---
# Test Agent

Content here.
`,
		},
		{
			name: "trim trailing whitespace",
			input: `---
name: test-agent
description: Test agent
---

# Test Agent

Content here.
`,
			expected: `---
name: test-agent
description: Test agent
---
# Test Agent

Content here.
`,
		},
		{
			name: "normalize blank line after frontmatter",
			input: `---
name: test-agent
description: Test agent
---


# Test Agent

Content here.
`,
			expected: `---
name: test-agent
description: Test agent
---
# Test Agent

Content here.
`,
		},
		{
			name: "ensure file ends with single newline",
			input: `---
name: test-agent
description: Test agent
---

# Test Agent

Content here.

`,
			expected: `---
name: test-agent
description: Test agent
---
# Test Agent

Content here.
`,
		},
		{
			name: "no frontmatter",
			input: `# Test Agent

Content here.
`,
			expected: `# Test Agent

Content here.
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &AgentFormatter{}
			result, err := formatter.Format(tt.input)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Format() mismatch\nGot:\n%s\n\nExpected:\n%s", result, tt.expected)
			}
		})
	}
}

func TestCommandFormatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "reorder frontmatter fields",
			input: `---
allowed-tools: [Task]
description: Test command
name: test-command
---

# Test Command

Delegate to agent.
`,
			expected: `---
name: test-command
description: Test command
allowed-tools:
  - Task
---
# Test Command

Delegate to agent.
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &CommandFormatter{}
			result, err := formatter.Format(tt.input)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Format() mismatch\nGot:\n%s\n\nExpected:\n%s", result, tt.expected)
			}
		})
	}
}

func TestSkillFormatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "normalize skill frontmatter",
			input: `---
version: 1.0
name: test-skill
description: Test skill
author: Someone
---

# Test Skill

Methodology here.
`,
			expected: `---
name: test-skill
description: Test skill
author: Someone
version: 1
---
# Test Skill

Methodology here.
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &SkillFormatter{}
			result, err := formatter.Format(tt.input)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Format() mismatch\nGot:\n%s\n\nExpected:\n%s", result, tt.expected)
			}
		})
	}
}

func TestDiff(t *testing.T) {
	original := `---
name: test
---

Content.
`
	formatted := `---
name: test
---
Content.
`

	diff := Diff(original, formatted, "test.md")
	if diff == "" {
		t.Error("Expected non-empty diff")
	}

	if !strings.Contains(diff, "--- test.md") {
		t.Error("Diff should contain file header")
	}
}

func TestDiffIdentical(t *testing.T) {
	content := `---
name: test
---
Content.
`

	diff := Diff(content, content, "test.md")
	if diff != "" {
		t.Errorf("Expected empty diff for identical content, got: %s", diff)
	}
}

func TestNormalizeFrontmatter(t *testing.T) {
	tests := []struct {
		name           string
		yaml           string
		priorityFields []string
		expectedOrder  []string // Expected order of keys in output
	}{
		{
			name: "priority fields first",
			yaml: `tools: [Read]
description: Test
name: test
model: sonnet`,
			priorityFields: []string{"name", "description", "model", "tools"},
			expectedOrder:  []string{"name", "description", "model", "tools"},
		},
		{
			name: "alphabetical for non-priority",
			yaml: `zebra: last
description: Test
name: test
alpha: first`,
			priorityFields: []string{"name", "description"},
			expectedOrder:  []string{"name", "description", "alpha", "zebra"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := normalizeFrontmatter(tt.yaml, tt.priorityFields)
			if err != nil {
				t.Fatalf("normalizeFrontmatter() error = %v", err)
			}

			// Check field order
			lines := strings.Split(result, "\n")
			var foundKeys []string
			for _, line := range lines {
				if strings.Contains(line, ":") {
					parts := strings.SplitN(line, ":", 2)
					key := strings.TrimSpace(parts[0])
					if key != "" && !strings.HasPrefix(key, "-") {
						foundKeys = append(foundKeys, key)
					}
				}
			}

			if len(foundKeys) != len(tt.expectedOrder) {
				t.Errorf("Expected %d keys, got %d: %v", len(tt.expectedOrder), len(foundKeys), foundKeys)
			}

			for i, expectedKey := range tt.expectedOrder {
				if i >= len(foundKeys) || foundKeys[i] != expectedKey {
					t.Errorf("Expected key[%d] = %s, got %s", i, expectedKey, foundKeys[i])
				}
			}
		})
	}
}

func TestParseFrontmatterRaw(t *testing.T) {
	tests := []struct {
		name          string
		input         string
		expectedFM    string
		expectedBody  string
		expectedHasFM bool
		expectError   bool
	}{
		{
			name: "valid frontmatter",
			input: `---
name: test
---
Body content`,
			expectedFM:    "\nname: test\n",
			expectedBody:  "\nBody content",
			expectedHasFM: true,
			expectError:   false,
		},
		{
			name:          "no frontmatter",
			input:         "Just body content",
			expectedFM:    "",
			expectedBody:  "Just body content",
			expectedHasFM: false,
			expectError:   false,
		},
		{
			name: "unclosed frontmatter",
			input: `---
name: test
Body without closing`,
			expectedFM:    "",
			expectedBody:  "---\nname: test\nBody without closing",
			expectedHasFM: false,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseFrontmatterRaw(tt.input)

			if tt.expectError && result.err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && result.err != nil {
				t.Errorf("Unexpected error: %v", result.err)
			}

			if result.hasFrontmatter != tt.expectedHasFM {
				t.Errorf("hasFrontmatter = %v, expected %v", result.hasFrontmatter, tt.expectedHasFM)
			}

			if !tt.expectError {
				if result.frontmatter != tt.expectedFM {
					t.Errorf("frontmatter = %q, expected %q", result.frontmatter, tt.expectedFM)
				}
				if result.body != tt.expectedBody {
					t.Errorf("body = %q, expected %q", result.body, tt.expectedBody)
				}
			}
		})
	}
}

func TestNewComponentFormatter(t *testing.T) {
	tests := []struct {
		name          string
		componentType string
		expectedType  string
	}{
		{
			name:          "agent formatter",
			componentType: "agent",
			expectedType:  "*format.AgentFormatter",
		},
		{
			name:          "command formatter",
			componentType: "command",
			expectedType:  "*format.CommandFormatter",
		},
		{
			name:          "skill formatter",
			componentType: "skill",
			expectedType:  "*format.SkillFormatter",
		},
		{
			name:          "generic formatter for unknown type",
			componentType: "unknown",
			expectedType:  "*format.GenericFormatter",
		},
		{
			name:          "generic formatter for empty string",
			componentType: "",
			expectedType:  "*format.GenericFormatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewComponentFormatter(tt.componentType)
			actualType := strings.TrimPrefix(strings.TrimPrefix(strings.ReplaceAll(strings.Replace(fmt.Sprintf("%T", formatter), "github.com/dotcommander/cclint/internal/", "", 1), ".", ""), "*"), "format")
			expectedShort := strings.TrimPrefix(strings.TrimPrefix(tt.expectedType, "*format."), "*")

			// Simple type check by formatting
			switch tt.componentType {
			case "agent":
				if _, ok := formatter.(*AgentFormatter); !ok {
					t.Errorf("Expected *AgentFormatter, got %T", formatter)
				}
			case "command":
				if _, ok := formatter.(*CommandFormatter); !ok {
					t.Errorf("Expected *CommandFormatter, got %T", formatter)
				}
			case "skill":
				if _, ok := formatter.(*SkillFormatter); !ok {
					t.Errorf("Expected *SkillFormatter, got %T", formatter)
				}
			default:
				if _, ok := formatter.(*GenericFormatter); !ok {
					t.Errorf("Expected *GenericFormatter, got %T", formatter)
				}
			}

			// Verify formatter works
			result, err := formatter.Format("# Test\nContent\n")
			if err != nil {
				t.Errorf("Format() error = %v", err)
			}
			if result == "" {
				t.Error("Format() returned empty string")
			}

			_ = actualType
			_ = expectedShort
		})
	}
}

func TestGenericFormatter(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name: "with frontmatter",
			input: `---
version: 2.0
name: generic-test
description: A generic file
author: Tester
---

# Generic Content

Some content.
`,
			expected: `---
name: generic-test
description: A generic file
author: Tester
version: 2
---
# Generic Content

Some content.
`,
		},
		{
			name: "without frontmatter",
			input: `# Generic Content

Just markdown.
`,
			expected: `# Generic Content

Just markdown.
`,
		},
		{
			name: "alphabetical ordering for non-priority fields",
			input: `---
name: test
zebra: last
description: Test
alpha: first
---

Content.
`,
			expected: `---
name: test
description: Test
alpha: first
zebra: last
---
Content.
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := &GenericFormatter{}
			result, err := formatter.Format(tt.input)
			if err != nil {
				t.Fatalf("Format() error = %v", err)
			}

			if result != tt.expected {
				t.Errorf("Format() mismatch\nGot:\n%s\n\nExpected:\n%s", result, tt.expected)
			}
		})
	}
}

func TestFormatterErrorPaths(t *testing.T) {
	tests := []struct {
		name      string
		formatter Formatter
		input     string
		wantError bool
	}{
		{
			name:      "agent formatter with invalid YAML",
			formatter: &AgentFormatter{},
			input: `---
name: test
invalid: [unclosed
---
Content`,
			wantError: true,
		},
		{
			name:      "command formatter with invalid YAML",
			formatter: &CommandFormatter{},
			input: `---
name: test
bad: {key: unclosed
---
Content`,
			wantError: true,
		},
		{
			name:      "skill formatter with invalid YAML",
			formatter: &SkillFormatter{},
			input: `---
name: test
error: {{nested
---
Content`,
			wantError: true,
		},
		{
			name:      "generic formatter with invalid YAML",
			formatter: &GenericFormatter{},
			input: `---
name: test
broken: [[[
---
Content`,
			wantError: true,
		},
		{
			name:      "agent formatter with unclosed frontmatter",
			formatter: &AgentFormatter{},
			input: `---
name: test
No closing delimiter`,
			wantError: true,
		},
		{
			name:      "command formatter with unclosed frontmatter",
			formatter: &CommandFormatter{},
			input: `---
name: test`,
			wantError: true,
		},
		{
			name:      "skill formatter with unclosed frontmatter",
			formatter: &SkillFormatter{},
			input: `---
description: test`,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.formatter.Format(tt.input)

			if tt.wantError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			// On error, original content should be returned
			if tt.wantError && err != nil && result != tt.input {
				t.Errorf("Expected original content on error, got different content")
			}
		})
	}
}

func TestNormalizeFrontmatterInvalidYAML(t *testing.T) {
	invalidYAML := `name: test
bad: [unclosed array`

	_, err := normalizeFrontmatter(invalidYAML, []string{"name"})
	if err == nil {
		t.Error("Expected error for invalid YAML")
	}
}

func TestNormalizeFrontmatterEncodingError(t *testing.T) {
	// Test with complex nested structure to ensure encoding works
	complexYAML := `name: test
nested:
  deep:
    value: test
  array:
    - item1
    - item2
list:
  - a
  - b
  - c`

	result, err := normalizeFrontmatter(complexYAML, []string{"name"})
	if err != nil {
		t.Errorf("Unexpected error with complex YAML: %v", err)
	}
	if !strings.Contains(result, "name: test") {
		t.Error("Result should contain name field")
	}
}

func TestDiffEdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		original  string
		formatted string
		filename  string
		wantDiff  bool
	}{
		{
			name:      "formatted has more lines",
			original:  "line1\nline2",
			formatted: "line1\nline2\nline3\nline4",
			filename:  "test.md",
			wantDiff:  true,
		},
		{
			name:      "original has more lines",
			original:  "line1\nline2\nline3\nline4",
			formatted: "line1\nline2",
			filename:  "test.md",
			wantDiff:  true,
		},
		{
			name:      "empty lines differ",
			original:  "line1\n\nline3",
			formatted: "line1\nline2\nline3",
			filename:  "test.md",
			wantDiff:  true,
		},
		{
			name:      "contains minus and plus markers",
			original:  "old content\nremoved line",
			formatted: "new content\nadded line",
			filename:  "file.md",
			wantDiff:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diff := Diff(tt.original, tt.formatted, tt.filename)

			if tt.wantDiff && diff == "" {
				t.Error("Expected non-empty diff")
			}
			if !tt.wantDiff && diff != "" {
				t.Errorf("Expected empty diff, got: %s", diff)
			}

			if tt.wantDiff {
				// Verify diff contains headers
				if !strings.Contains(diff, "--- "+tt.filename) {
					t.Error("Diff should contain original file header")
				}
				if !strings.Contains(diff, "+++ "+tt.filename+" (formatted)") {
					t.Error("Diff should contain formatted file header")
				}

				// Verify diff contains change markers
				if !strings.Contains(diff, "-") && !strings.Contains(diff, "+") {
					t.Error("Diff should contain - or + markers for changes")
				}
			}
		})
	}
}

func TestNormalizeMarkdownEdgeCases(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		hasFrontmatter bool
		expected       string
	}{
		{
			name:           "multiple trailing newlines",
			input:          "Content\n\n\n\n",
			hasFrontmatter: false,
			expected:       "Content\n",
		},
		{
			name:           "no trailing newline",
			input:          "Content",
			hasFrontmatter: false,
			expected:       "Content\n",
		},
		{
			name:           "trailing spaces and tabs",
			input:          "Line with spaces   \t\nAnother line\t",
			hasFrontmatter: false,
			expected:       "Line with spaces\nAnother line\n",
		},
		{
			name:           "multiple leading newlines with frontmatter",
			input:          "\n\n\nContent",
			hasFrontmatter: true,
			expected:       "\nContent\n",
		},
		{
			name:           "no leading newlines with frontmatter",
			input:          "Content",
			hasFrontmatter: true,
			expected:       "\nContent\n",
		},
		{
			name:           "empty content with frontmatter",
			input:          "",
			hasFrontmatter: true,
			expected:       "\n",
		},
		{
			name:           "empty content without frontmatter",
			input:          "",
			hasFrontmatter: false,
			expected:       "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeMarkdown(tt.input, tt.hasFrontmatter)
			if result != tt.expected {
				t.Errorf("normalizeMarkdown() = %q, expected %q", result, tt.expected)
			}
		})
	}
}
