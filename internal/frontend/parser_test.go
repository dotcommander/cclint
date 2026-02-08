package frontend

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseYAMLFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantData    map[string]any
		wantBody    string
		wantErr     bool
		description string
	}{
		{
			name: "valid_simple_frontmatter",
			input: `---
name: test-agent
model: sonnet
---
# Agent Content

This is the body.`,
			wantData: map[string]any{
				"name":  "test-agent",
				"model": "sonnet",
			},
			wantBody:    "\n# Agent Content\n\nThis is the body.",
			wantErr:     false,
			description: "Valid frontmatter with simple string fields",
		},
		{
			name:  "no_frontmatter",
			input: "# Just Markdown\n\nNo frontmatter here.",
			wantData: map[string]any{},
			wantBody: "# Just Markdown\n\nNo frontmatter here.",
			wantErr:  false,
			description: "Plain markdown with no frontmatter",
		},
		{
			name: "missing_closing_delimiter",
			input: `---
name: test
model: sonnet
# Missing closing ---`,
			wantData: map[string]any{
				"name": "test",
				"model": "sonnet",
			},
			wantBody:    "",
			wantErr:     false,
			description: "Malformed frontmatter missing closing delimiter gets parsed anyway due to SplitN behavior",
		},
		{
			name: "empty_frontmatter",
			input: `---
---
# Content`,
			wantData:    nil,
			wantBody:    "\n# Content",
			wantErr:     false,
			description: "Empty frontmatter block",
		},
		{
			name: "complex_nested_yaml",
			input: `---
name: complex-agent
model: sonnet
triggers:
  - keyword: analyze
  - keyword: review
metadata:
  version: 1.0.0
  tags:
    - testing
    - analysis
config:
  max_tokens: 4000
  temperature: 0.7
---
# Complex Agent

Body content.`,
			wantData: map[string]any{
				"name":  "complex-agent",
				"model": "sonnet",
				"triggers": []any{
					map[string]any{"keyword": "analyze"},
					map[string]any{"keyword": "review"},
				},
				"metadata": map[string]any{
					"version": "1.0.0",
					"tags":    []any{"testing", "analysis"},
				},
				"config": map[string]any{
					"max_tokens":  4000,
					"temperature": 0.7,
				},
			},
			wantBody:    "\n# Complex Agent\n\nBody content.",
			wantErr:     false,
			description: "Complex nested YAML with arrays and objects",
		},
		{
			name: "frontmatter_with_special_characters",
			input: `---
name: "agent-with-quotes"
description: 'Single quotes work too'
symbol: "@special"
path: /path/to/file.md
emoji: "🚀"
multiline: |
  Line 1
  Line 2
---
Body`,
			wantData: map[string]any{
				"name":        "agent-with-quotes",
				"description": "Single quotes work too",
				"symbol":      "@special",
				"path":        "/path/to/file.md",
				"emoji":       "🚀",
				"multiline":   "Line 1\nLine 2\n",
			},
			wantBody:    "\nBody",
			wantErr:     false,
			description: "Frontmatter with special characters and multiline strings",
		},
		{
			name: "body_content_preservation",
			input: `---
name: test
---
# Heading

Some text with **bold** and *italic*.

- List item 1
- List item 2

` + "```go\ncode block\n```" + `

More content.`,
			wantData: map[string]any{
				"name": "test",
			},
			wantBody: `
# Heading

Some text with **bold** and *italic*.

- List item 1
- List item 2

` + "```go\ncode block\n```" + `

More content.`,
			wantErr:     false,
			description: "Body content with markdown formatting is preserved",
		},
		{
			name: "frontmatter_with_leading_whitespace",
			input: `  	---
name: test
---
Body`,
			wantData: map[string]any{
				"name": "test",
			},
			wantBody:    "\nBody",
			wantErr:     false,
			description: "Frontmatter with leading spaces and tabs is recognized",
		},
		{
			name: "triple_dashes_in_body",
			input: `---
name: test
---
Body with --- in it should work fine.`,
			wantData: map[string]any{
				"name": "test",
			},
			wantBody:    "\nBody with --- in it should work fine.",
			wantErr:     false,
			description: "Triple dashes in body content are preserved",
		},
		{
			name: "numeric_values",
			input: `---
count: 42
price: 19.99
enabled: true
disabled: false
---
Body`,
			wantData: map[string]any{
				"count":    42,
				"price":    19.99,
				"enabled":  true,
				"disabled": false,
			},
			wantBody:    "\nBody",
			wantErr:     false,
			description: "Numeric and boolean values are correctly typed",
		},
		{
			name: "empty_string",
			input: "",
			wantData: map[string]any{},
			wantBody: "",
			wantErr:  false,
			description: "Empty string returns empty frontmatter and body",
		},
		{
			name: "only_opening_delimiter",
			input: `---
name: test`,
			wantData: map[string]any{},
			wantBody: `---
name: test`,
			wantErr:     false,
			description: "Only opening delimiter without closing returns content as body",
		},
		{
			name: "malformed_yaml",
			input: `---
name: test
invalid: [unclosed
---
Body`,
			wantData:    nil,
			wantBody:    "",
			wantErr:     true,
			description: "Malformed YAML returns error",
		},
		{
			name: "array_values",
			input: `---
tags:
  - go
  - testing
  - linter
numbers: [1, 2, 3, 4, 5]
---
Body`,
			wantData: map[string]any{
				"tags":    []any{"go", "testing", "linter"},
				"numbers": []any{1, 2, 3, 4, 5},
			},
			wantBody:    "\nBody",
			wantErr:     false,
			description: "Array values in different formats",
		},
		{
			name: "null_values",
			input: `---
name: test
optional: null
empty: ~
---
Body`,
			wantData: map[string]any{
				"name":     "test",
				"optional": nil,
				"empty":    nil,
			},
			wantBody:    "\nBody",
			wantErr:     false,
			description: "Null values are handled correctly",
		},
		{
			name: "frontmatter_only_no_body",
			input: `---
name: test
model: sonnet
---`,
			wantData: map[string]any{
				"name":  "test",
				"model": "sonnet",
			},
			wantBody:    "",
			wantErr:     false,
			description: "Frontmatter with no body content",
		},
		{
			name: "newlines_in_frontmatter",
			input: `---

name: test

model: sonnet

---
Body`,
			wantData: map[string]any{
				"name":  "test",
				"model": "sonnet",
			},
			wantBody:    "\nBody",
			wantErr:     false,
			description: "Extra newlines in frontmatter are handled",
		},
		{
			name: "unicode_content",
			input: `---
name: 日本語
emoji: "🎉🎊"
chinese: 中文
---
Body with unicode: Ñoño café`,
			wantData: map[string]any{
				"name":    "日本語",
				"emoji":   "🎉🎊",
				"chinese": "中文",
			},
			wantBody:    "\nBody with unicode: Ñoño café",
			wantErr:     false,
			description: "Unicode characters in frontmatter and body",
		},
		{
			name: "deeply_nested_structure",
			input: `---
level1:
  level2:
    level3:
      level4:
        value: deep
---
Body`,
			wantData: map[string]any{
				"level1": map[string]any{
					"level2": map[string]any{
						"level3": map[string]any{
							"level4": map[string]any{
								"value": "deep",
							},
						},
					},
				},
			},
			wantBody:    "\nBody",
			wantErr:     false,
			description: "Deeply nested YAML structures",
		},
		{
			name: "mixed_content_types",
			input: `---
string: "hello"
number: 42
float: 3.14
bool: true
null_value: null
array: [1, "two", 3.0, true]
object:
  nested: value
---
Body`,
			wantData: map[string]any{
				"string":     "hello",
				"number":     42,
				"float":      3.14,
				"bool":       true,
				"null_value": nil,
				"array":      []any{1, "two", 3.0, true},
				"object": map[string]any{
					"nested": "value",
				},
			},
			wantBody:    "\nBody",
			wantErr:     false,
			description: "Mixed content types in frontmatter",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseYAMLFrontmatter(tt.input)

			if tt.wantErr {
				require.Error(t, err, "Expected error for: %s", tt.description)
				return
			}

			require.NoError(t, err, "Unexpected error for: %s", tt.description)
			require.NotNil(t, result, "Result should not be nil for: %s", tt.description)

			assert.Equal(t, tt.wantData, result.Data, "Data mismatch for: %s", tt.description)
			assert.Equal(t, tt.wantBody, result.Body, "Body mismatch for: %s", tt.description)
		})
	}
}

func TestParseYAMLFrontmatter_EdgeCases(t *testing.T) {
	t.Run("multiple_documents_in_body", func(t *testing.T) {
		input := `---
name: test
---
# Document 1

---

# Document 2`
		result, err := ParseYAMLFrontmatter(input)
		require.NoError(t, err)
		assert.Equal(t, "test", result.Data["name"])
		// Body should include the remaining content with --- separators
		assert.Contains(t, result.Body, "# Document 1")
		assert.Contains(t, result.Body, "---")
		assert.Contains(t, result.Body, "# Document 2")
	})

	t.Run("whitespace_only_body", func(t *testing.T) {
		input := `---
name: test
---


  `
		result, err := ParseYAMLFrontmatter(input)
		require.NoError(t, err)
		assert.Equal(t, "test", result.Data["name"])
		// Body should preserve whitespace
		assert.NotEmpty(t, result.Body)
	})

	t.Run("frontmatter_with_comments", func(t *testing.T) {
		input := `---
# This is a comment
name: test
model: sonnet # inline comment
---
Body`
		result, err := ParseYAMLFrontmatter(input)
		require.NoError(t, err)
		assert.Equal(t, "test", result.Data["name"])
		assert.Equal(t, "sonnet", result.Data["model"])
	})

	t.Run("windows_line_endings", func(t *testing.T) {
		input := "---\r\nname: test\r\nmodel: sonnet\r\n---\r\nBody"
		result, err := ParseYAMLFrontmatter(input)
		require.NoError(t, err)
		assert.Equal(t, "test", result.Data["name"])
		assert.Equal(t, "sonnet", result.Data["model"])
	})

	t.Run("very_large_frontmatter", func(t *testing.T) {
		// Create a large frontmatter with many fields using valid field names
		input := "---\n"
		for i := 0; i < 100; i++ {
			input += "field_" + string(rune('a'+i%26)) + string(rune('0'+i%10)) + ": value" + string(rune('0'+i%10)) + "\n"
		}
		input += "---\nBody"

		result, err := ParseYAMLFrontmatter(input)
		require.NoError(t, err)
		assert.NotNil(t, result.Data)
		assert.Equal(t, "\nBody", result.Body)
	})
}

func TestFrontmatter_DataAccessPatterns(t *testing.T) {
	input := `---
name: test-agent
model: sonnet
count: 42
enabled: true
tags:
  - testing
  - go
---
Body`

	result, err := ParseYAMLFrontmatter(input)
	require.NoError(t, err)

	t.Run("string_field_access", func(t *testing.T) {
		name, ok := result.Data["name"].(string)
		assert.True(t, ok)
		assert.Equal(t, "test-agent", name)
	})

	t.Run("numeric_field_access", func(t *testing.T) {
		count, ok := result.Data["count"].(int)
		assert.True(t, ok)
		assert.Equal(t, 42, count)
	})

	t.Run("boolean_field_access", func(t *testing.T) {
		enabled, ok := result.Data["enabled"].(bool)
		assert.True(t, ok)
		assert.Equal(t, true, enabled)
	})

	t.Run("array_field_access", func(t *testing.T) {
		tags, ok := result.Data["tags"].([]any)
		assert.True(t, ok)
		assert.Len(t, tags, 2)
		assert.Equal(t, "testing", tags[0])
		assert.Equal(t, "go", tags[1])
	})

	t.Run("missing_field_access", func(t *testing.T) {
		value, exists := result.Data["nonexistent"]
		assert.False(t, exists)
		assert.Nil(t, value)
	})
}

func BenchmarkParseYAMLFrontmatter(b *testing.B) {
	input := `---
name: benchmark-agent
model: sonnet
triggers:
  - keyword: test
  - keyword: benchmark
config:
  max_tokens: 4000
  temperature: 0.7
---
# Benchmark Agent

This is a benchmark test with some body content.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseYAMLFrontmatter(input)
	}
}

func BenchmarkParseYAMLFrontmatter_NoFrontmatter(b *testing.B) {
	input := `# Just Markdown

No frontmatter here. This should be fast.`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseYAMLFrontmatter(input)
	}
}

func BenchmarkParseYAMLFrontmatter_LargeFrontmatter(b *testing.B) {
	input := "---\n"
	for i := 0; i < 100; i++ {
		input += "field_" + string(rune('a'+i%26)) + string(rune('0'+i%10)) + ": value" + string(rune('0'+i%10)) + "\n"
	}
	input += "---\nBody"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ParseYAMLFrontmatter(input)
	}
}
