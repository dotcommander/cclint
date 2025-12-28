package format

import (
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
		name              string
		input             string
		expectedFM        string
		expectedBody      string
		expectedHasFM     bool
		expectError       bool
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
			name:              "no frontmatter",
			input:             "Just body content",
			expectedFM:        "",
			expectedBody:      "Just body content",
			expectedHasFM:     false,
			expectError:       false,
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
			fm, body, hasFM, err := parseFrontmatterRaw(tt.input)

			if tt.expectError && err == nil {
				t.Error("Expected error but got nil")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}

			if hasFM != tt.expectedHasFM {
				t.Errorf("hasFrontmatter = %v, expected %v", hasFM, tt.expectedHasFM)
			}

			if !tt.expectError {
				if fm != tt.expectedFM {
					t.Errorf("frontmatter = %q, expected %q", fm, tt.expectedFM)
				}
				if body != tt.expectedBody {
					t.Errorf("body = %q, expected %q", body, tt.expectedBody)
				}
			}
		})
	}
}
