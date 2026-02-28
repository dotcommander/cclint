package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
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
		{
			name: "yaml frontmatter stripped",
			content: `---
title: My Project
---

## Build

Instructions here.

## Testing

Test stuff.`,
			wantSections: 2,
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
		name             string
		sections         []any
		wantErrorCount   int
		wantSeverities   []string
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
			wantSeverities: []string{cue.SeverityWarning},
		},
		{
			name: "section missing heading",
			sections: []any{
				map[string]any{"heading": "", "content": "Some content"},
			},
			wantErrorCount: 1,
			wantSeverities: []string{cue.SeverityWarning},
		},
		{
			name: "both heading and content missing",
			sections: []any{
				map[string]any{"heading": "", "content": ""},
			},
			wantErrorCount: 2,
			wantSeverities: []string{cue.SeverityWarning, cue.SeverityWarning},
		},
		{
			name: "h1 title without content is ok",
			sections: []any{
				map[string]any{"heading": "My Project", "content": "", "level": 1},
				map[string]any{"heading": "Build", "content": "go build", "level": 2},
			},
			wantErrorCount: 0,
		},
		{
			name: "h2 without content still warns",
			sections: []any{
				map[string]any{"heading": "My Project", "content": "", "level": 1},
				map[string]any{"heading": "Empty Section", "content": "", "level": 2},
			},
			wantErrorCount: 1,
			wantSeverities: []string{cue.SeverityWarning},
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
			for i, wantSev := range tt.wantSeverities {
				if i >= len(errs) {
					break
				}
				if errs[i].Severity != wantSev {
					t.Errorf("errs[%d].Severity = %q, want %q", i, errs[i].Severity, wantSev)
				}
			}
		})
	}
}

func TestCheckBinaryIncludes(t *testing.T) {
	tests := []struct {
		name             string
		contents         string
		wantErrorCount   int
		wantSeverities   []string
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
			wantSeverities: []string{cue.SeverityWarning},
		},
		{
			name:           "jpg img tag style",
			contents:       "@include ./images/banner.jpg",
			wantErrorCount: 1,
			wantSeverities: []string{cue.SeverityWarning},
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
			wantSeverities: []string{cue.SeverityWarning, cue.SeverityWarning},
		},
		{
			name:           "pdf include",
			contents:       "@include ./docs/spec.pdf",
			wantErrorCount: 1,
			wantSeverities: []string{cue.SeverityWarning},
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
			for i, wantSev := range tt.wantSeverities {
				if i >= len(errs) {
					break
				}
				if errs[i].Severity != wantSev {
					t.Errorf("errs[%d].Severity = %q, want %q", i, errs[i].Severity, wantSev)
				}
			}
		})
	}
}

func TestValidateContextSpecific(t *testing.T) {
	tests := []struct {
		name             string
		data             map[string]any
		contents         string
		wantErrorCount   int
		wantSeverities   []string
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
			wantSeverities: []string{cue.SeverityWarning},
		},
		{
			name:           "nil sections produces suggestion",
			data:           map[string]any{},
			contents:       "",
			wantErrorCount: 1,
			wantSeverities: []string{cue.SeveritySuggestion},
		},
		{
			name: "empty sections slice produces suggestion",
			data: map[string]any{
				"sections": []any{},
			},
			contents:       "",
			wantErrorCount: 1,
			wantSeverities: []string{cue.SeveritySuggestion},
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
			for i, wantSev := range tt.wantSeverities {
				if i >= len(errs) {
					break
				}
				if errs[i].Severity != wantSev {
					t.Errorf("errs[%d].Severity = %q, want %q", i, errs[i].Severity, wantSev)
				}
			}
		})
	}
}

func TestLintContextWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// CLAUDE.md with valid sections and a binary @include
	content := "# My Project\n\nProject overview here.\n\n## Build\n\n`go build ./...`\n\n## Testing\n\n`go test ./...`\n\n@include ./assets/logo.png\n"
	claudeMdPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMdPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := LintContext(tmpDir, false, false, true)
	if err != nil {
		t.Fatalf("LintContext() error = %v", err)
	}
	if summary == nil {
		t.Fatal("LintContext() returned nil summary")
	}

	if summary.TotalFiles != 1 {
		t.Errorf("LintContext() TotalFiles = %d, want 1", summary.TotalFiles)
	}

	// The binary @include should produce at least one warning
	if summary.TotalWarnings < 1 {
		t.Errorf("LintContext() TotalWarnings = %d, want >= 1 (binary include warning)", summary.TotalWarnings)
	}

	// Verify the binary-include warning appears in results
	found := false
	for _, result := range summary.Results {
		for _, w := range result.Warnings {
			if w.Source == "binary-include" {
				found = true
				break
			}
		}
		if found {
			break
		}
	}
	if !found {
		t.Error("LintContext() expected binary-include warning in results")
	}
}

// Regression: h1 title followed by h2 subsections should not warn about missing content.
// See: cclint ~/.config/opencode/CLAUDE.md produced false positive "Section 0: missing content".
func TestLintContextH1TitleNoFalsePositive(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	// h1 title with no body content, followed immediately by h2 sections
	content := "# OpenCode Configuration\n\n## Custom Commands\n\nCommands are .md files.\n\n## Plugins\n\nPlugin setup.\n"
	if err := os.WriteFile(filepath.Join(claudeDir, "CLAUDE.md"), []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := LintContext(tmpDir, false, false, true)
	if err != nil {
		t.Fatalf("LintContext() error = %v", err)
	}

	if summary.TotalWarnings != 0 {
		t.Errorf("LintContext() TotalWarnings = %d, want 0 (h1 title should not warn)", summary.TotalWarnings)
		for _, result := range summary.Results {
			for _, w := range result.Warnings {
				t.Logf("  warning: %s", w.Message)
			}
		}
	}
}
