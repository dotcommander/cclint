package lint

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
			name: "no headings",
			content: `Just some plain text
without any headings at all
even multiple lines`,
			wantSections: 0,
		},
		{
			name: "empty content sections",
			content: `# Section 1

## Section 2

## Section 3`,
			wantSections: 3,
		},
		{
			name: "extra spaces in heading",
			content: `## Build & Run

Content here`,
			wantSections: 1,
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
				sectionMap, ok := section.(map[string]any)
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

func TestParseMarkdownSectionsContent(t *testing.T) {
	content := `# Main Title

First section content

## Subsection

Subsection body text`

	sections := parseMarkdownSections(content)
	if len(sections) != 2 {
		t.Fatalf("expected 2 sections, got %d", len(sections))
	}

	s0 := sections[0].(map[string]any)
	if s0["heading"] != "Main Title" {
		t.Errorf("section 0 heading = %q, want %q", s0["heading"], "Main Title")
	}
	if s0["content"] == "" {
		t.Error("section 0 content should not be empty")
	}

	s1 := sections[1].(map[string]any)
	if s1["heading"] != "Subsection" {
		t.Errorf("section 1 heading = %q, want %q", s1["heading"], "Subsection")
	}
	if s1["content"] == "" {
		t.Error("section 1 content should not be empty")
	}
}

func TestValidateContextSections(t *testing.T) {
	tests := []struct {
		name           string
		sections       []any
		wantErrorCount int
	}{
		{
			name: "all sections valid",
			sections: []any{
				map[string]any{"heading": "Build & Commands", "content": "go build ./..."},
				map[string]any{"heading": "Architecture", "content": "Clean arch with layers"},
			},
			wantErrorCount: 0,
		},
		{
			name:           "empty sections slice",
			sections:       []any{},
			wantErrorCount: 0,
		},
		{
			name: "section missing content",
			sections: []any{
				map[string]any{"heading": "Empty Section", "content": ""},
			},
			wantErrorCount: 1,
		},
		{
			name: "section missing heading",
			sections: []any{
				map[string]any{"heading": "", "content": "Some content"},
			},
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateContextSections(tt.sections, "CLAUDE.md")
			if len(errs) != tt.wantErrorCount {
				t.Errorf("validateContextSections() error count = %d, want %d", len(errs), tt.wantErrorCount)
				for _, e := range errs {
					t.Logf("  - [%s] %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestCheckBinaryIncludes(t *testing.T) {
	tests := []struct {
		name           string
		contents       string
		wantErrorCount int
	}{
		{
			name:           "no includes",
			contents:       "Just some markdown content",
			wantErrorCount: 0,
		},
		{
			name:           "png include",
			contents:       "@include ./assets/logo.png",
			wantErrorCount: 1,
		},
		{
			name:           "jpg img tag style",
			contents:       "@include ./images/banner.jpg",
			wantErrorCount: 1,
		},
		{
			name:           "text file ok",
			contents:       "@include ./docs/README.md",
			wantErrorCount: 0,
		},
		{
			name:           "no extension",
			contents:       "@include ./Makefile",
			wantErrorCount: 0,
		},
		{
			name:           "multiple binary includes",
			contents:       "@include ./docs/manual.pdf\n@include ./assets/icon.jpg",
			wantErrorCount: 2,
		},
		{
			name:           "pdf include",
			contents:       "@include ./docs/spec.pdf",
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := checkBinaryIncludes(tt.contents, "CLAUDE.md")
			if len(errs) != tt.wantErrorCount {
				t.Errorf("checkBinaryIncludes() error count = %d, want %d", len(errs), tt.wantErrorCount)
				for _, e := range errs {
					t.Logf("  - [%s] %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestValidateContextSpecific(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]any
		contents       string
		wantErrorCount int
	}{
		{
			name: "clean - valid sections no binary",
			data: map[string]any{
				"sections": []any{
					map[string]any{"heading": "Build & Run", "content": "go build ./..."},
					map[string]any{"heading": "Testing", "content": "go test ./..."},
				},
			},
			contents:       "@include ./docs/API.md",
			wantErrorCount: 0,
		},
		{
			name: "with binary include",
			data: map[string]any{
				"sections": []any{
					map[string]any{"heading": "Images", "content": "See below"},
				},
			},
			contents:       "@include ./assets/logo.png",
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateContextSpecific(tt.data, "CLAUDE.md", tt.contents)
			if len(errs) != tt.wantErrorCount {
				t.Errorf("validateContextSpecific() error count = %d, want %d", len(errs), tt.wantErrorCount)
				for _, e := range errs {
					t.Logf("  - [%s] %s", e.Severity, e.Message)
				}
			}
		})
	}
}
