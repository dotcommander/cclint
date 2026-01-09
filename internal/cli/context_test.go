package cli

import (
	"testing"
)

func TestLintContext(t *testing.T) {
	// Test with empty directory
	summary, err := LintContext("testdata/empty", false, false, true)
	if err != nil {
		t.Fatalf("LintContext() error = %v", err)
	}
	if summary == nil {
		t.Fatal("LintContext() returned nil summary")
	}
}

func TestParseMarkdownSections(t *testing.T) {
	tests := []struct {
		name         string
		content      string
		wantSections int
	}{
		{
			name:         "empty content",
			content:      "",
			wantSections: 0,
		},
		{
			name: "single h1 section",
			content: `# CLAUDE.md

Project instructions.`,
			wantSections: 1,
		},
		{
			name: "single h2 section",
			content: `## Build & Run

Instructions here.`,
			wantSections: 1,
		},
		{
			name: "multiple sections",
			content: `# CLAUDE.md

Project context

## Build & Run

Build instructions

## Testing

Test instructions`,
			wantSections: 3,
		},
		{
			name: "with frontmatter",
			content: `---
title: Project
---

# CLAUDE.md

Content here

## Section 2

More content`,
			wantSections: 2,
		},
		{
			name: "empty sections",
			content: `# Section 1

## Section 2

## Section 3`,
			wantSections: 3,
		},
		{
			name: "mixed heading levels",
			content: `# Main Title

Some content

## Subsection

Subsection content

# Another Main Section

More content`,
			wantSections: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sections := parseMarkdownSections(tt.content)
			if len(sections) != tt.wantSections {
				t.Errorf("parseMarkdownSections() section count = %d, want %d", len(sections), tt.wantSections)
			}

			// Verify section structure
			for i, section := range sections {
				sectionMap, ok := section.(map[string]interface{})
				if !ok {
					t.Errorf("Section %d is not a map", i)
					continue
				}

				if _, hasHeading := sectionMap["heading"]; !hasHeading {
					t.Errorf("Section %d missing 'heading' field", i)
				}

				if _, hasContent := sectionMap["content"]; !hasContent {
					t.Errorf("Section %d missing 'content' field", i)
				}
			}
		})
	}
}

func TestValidateContextSpecific(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]interface{}
		contents       string
		wantErrorCount int
	}{
		{
			name:           "no sections",
			data:           map[string]interface{}{},
			contents:       "",
			wantErrorCount: 1, // suggestion for no sections
		},
		{
			name: "empty sections array",
			data: map[string]interface{}{
				"sections": []interface{}{},
			},
			contents:       "",
			wantErrorCount: 1, // suggestion for no sections
		},
		{
			name: "valid sections",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"heading": "Build & Run",
						"content": "Instructions here",
					},
					map[string]interface{}{
						"heading": "Testing",
						"content": "Test instructions",
					},
				},
			},
			contents:       "",
			wantErrorCount: 0,
		},
		{
			name: "section missing heading",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"content": "Content without heading",
					},
				},
			},
			contents:       "",
			wantErrorCount: 1,
		},
		{
			name: "section missing content",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"heading": "Empty Section",
					},
				},
			},
			contents:       "",
			wantErrorCount: 1,
		},
		{
			name: "section with empty heading",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"heading": "",
						"content": "Content",
					},
				},
			},
			contents:       "",
			wantErrorCount: 1,
		},
		{
			name: "section with empty content",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"heading": "Heading",
						"content": "",
					},
				},
			},
			contents:       "",
			wantErrorCount: 1,
		},
		{
			name: "multiple issues",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"heading": "Valid",
						"content": "Valid content",
					},
					map[string]interface{}{
						"heading": "",
						"content": "",
					},
					map[string]interface{}{},
				},
			},
			contents:       "",
			wantErrorCount: 4, // empty heading, empty content, missing heading, missing content
		},
		// Binary @include tests (Claude Code 2.1.2+)
		{
			name: "binary include - png",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"heading": "Images",
						"content": "See image below",
					},
				},
			},
			contents:       "@include ./assets/logo.png",
			wantErrorCount: 1, // binary include warning
		},
		{
			name: "binary include - multiple",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"heading": "Assets",
						"content": "Various files",
					},
				},
			},
			contents:       "@include ./docs/manual.pdf\n@include ./assets/icon.jpg",
			wantErrorCount: 2, // two binary include warnings
		},
		{
			name: "valid include - markdown",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"heading": "Docs",
						"content": "See below",
					},
				},
			},
			contents:       "@include ./docs/API.md",
			wantErrorCount: 0, // markdown is fine
		},
		{
			name: "mixed includes - text and binary",
			data: map[string]interface{}{
				"sections": []interface{}{
					map[string]interface{}{
						"heading": "Resources",
						"content": "Files",
					},
				},
			},
			contents:       "@include ./config.json\n@include ./logo.png\n@include ./README.md",
			wantErrorCount: 1, // only png is binary
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateContextSpecific(tt.data, "CLAUDE.md", tt.contents)
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateContextSpecific() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - %s: %s", err.Severity, err.Message)
				}
			}
		})
	}
}
