package output

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/cue"
)

func TestMarkdownFormatter_Format(t *testing.T) {
	tests := []struct {
		name         string
		summary      *lint.LintSummary
		quiet        bool
		verbose      bool
		outputFile   string
		wantContains []string
	}{
		{
			name: "basic markdown output",
			summary: &lint.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 1,
				FailedFiles:     0,
				TotalErrors:     0,
				TotalWarnings:   0,
				StartTime:       time.Now(),
				Results: []lint.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: true,
					},
				},
			},
			wantContains: []string{
				"# CCLint Report",
				"**Generated:**",
				"**Project:**",
				"**Duration:**",
				"## Summary",
				"| Metric | Count |",
				"| Files Scanned | 1 |",
				"| Successful | 1 |",
				"| Failed | 0 |",
				"| Errors | 0 |",
				"| Warnings | 0 |",
				"## Detailed Results",
				"## Conclusion",
				"✓ All files passed validation!",
			},
		},
		{
			name: "output with errors",
			summary: &lint.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 0,
				FailedFiles:     1,
				TotalErrors:     2,
				TotalWarnings:   0,
				StartTime:       time.Now(),
				Results: []lint.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: false,
						Errors: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "missing required field",
								Severity: "error",
								Source:   cue.SourceAnthropicDocs,
								Line:     5,
							},
							{
								File:     "test.md",
								Message:  "invalid format",
								Severity: "error",
								Source:   cue.SourceCClintObserve,
								Line:     10,
							},
						},
					},
				},
			},
			wantContains: []string{
				"### test.md",
				"Status: ❌",
				"Type: `agent`",
				"#### Errors",
				"- **test.md** - missing required field (line 5) `[docs]`",
				"- **test.md** - invalid format (line 10) `[cclint]`",
				"✗ 1 files failed validation",
			},
		},
		{
			name: "output with warnings",
			summary: &lint.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 1,
				FailedFiles:     0,
				TotalErrors:     0,
				TotalWarnings:   2,
				StartTime:       time.Now(),
				Results: []lint.LintResult{
					{
						File:    "test.md",
						Type:    "command",
						Success: true,
						Warnings: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "consider adding description",
								Severity: "warning",
								Line:     3,
							},
							{
								File:     "test.md",
								Message:  "could improve formatting",
								Severity: "warning",
								Source:   cue.SourceCClintObserve,
							},
						},
					},
				},
			},
			verbose: true,
			wantContains: []string{
				"### test.md",
				"Status: ✅",
				"#### Warnings",
				"- **test.md** - consider adding description (line 3)",
				"- **test.md** - could improve formatting `[cclint]`",
			},
		},
		{
			name: "empty results",
			summary: &lint.LintSummary{
				TotalFiles:      0,
				SuccessfulFiles: 0,
				FailedFiles:     0,
				TotalErrors:     0,
				TotalWarnings:   0,
				StartTime:       time.Now(),
				Results:         []lint.LintResult{},
			},
			wantContains: []string{
				"# CCLint Report",
				"| Files Scanned | 0 |",
				"*No files found to validate.*",
				"✓ All files passed validation!",
			},
		},
		{
			name: "multiple files with table of contents",
			summary: &lint.LintSummary{
				TotalFiles:      3,
				SuccessfulFiles: 2,
				FailedFiles:     1,
				TotalErrors:     1,
				TotalWarnings:   0,
				StartTime:       time.Now(),
				Results: []lint.LintResult{
					{
						File:    "test1.md",
						Type:    "agent",
						Success: true,
					},
					{
						File:    "test2.md",
						Type:    "command",
						Success: false,
						Errors: []cue.ValidationError{
							{
								File:     "test2.md",
								Message:  "error in test2",
								Severity: "error",
								Line:     10,
							},
						},
					},
					{
						File:    "test3.md",
						Type:    "skill",
						Success: true,
					},
				},
			},
			verbose: true,
			wantContains: []string{
				"### Files",
				"- [test1.md](#test1md)",
				"- [test2.md](#test2md)",
				"- [test3.md](#test3md)",
				"### test1.md",
				"### test2.md",
				"### test3.md",
			},
		},
		{
			name: "non-verbose mode hides successful files",
			summary: &lint.LintSummary{
				TotalFiles:      2,
				SuccessfulFiles: 1,
				FailedFiles:     1,
				TotalErrors:     1,
				TotalWarnings:   0,
				StartTime:       time.Now(),
				Results: []lint.LintResult{
					{
						File:    "success.md",
						Type:    "agent",
						Success: true,
					},
					{
						File:    "failed.md",
						Type:    "command",
						Success: false,
						Errors: []cue.ValidationError{
							{
								File:     "failed.md",
								Message:  "error message",
								Severity: "error",
								Line:     5,
							},
						},
					},
				},
			},
			verbose: false,
			wantContains: []string{
				"### failed.md",
				"error message",
			},
		},
		{
			name: "error without line number",
			summary: &lint.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 0,
				FailedFiles:     1,
				TotalErrors:     1,
				StartTime:       time.Now(),
				Results: []lint.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: false,
						Errors: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "general error",
								Severity: "error",
								Line:     0,
							},
						},
					},
				},
			},
			wantContains: []string{
				"- **test.md** - general error",
			},
		},
		{
			name: "mixed errors and warnings",
			summary: &lint.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 0,
				FailedFiles:     1,
				TotalErrors:     1,
				TotalWarnings:   1,
				StartTime:       time.Now(),
				Results: []lint.LintResult{
					{
						File:    "test.md",
						Type:    "skill",
						Success: false,
						Errors: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "critical error",
								Severity: "error",
								Line:     5,
							},
						},
						Warnings: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "minor issue",
								Severity: "warning",
								Line:     10,
							},
						},
					},
				},
			},
			verbose: true,
			wantContains: []string{
				"#### Errors",
				"- **test.md** - critical error (line 5)",
				"#### Warnings",
				"- **test.md** - minor issue (line 10)",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output := captureStdout(t, func() {
				formatter := NewMarkdownFormatter(tt.quiet, tt.verbose, tt.outputFile)
				if err := formatter.Format(tt.summary); err != nil {
					t.Fatalf("Format() error = %v", err)
				}
			})

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Output missing expected string:\n  want: %q\n  got: %q", want, output)
				}
			}
		})
	}
}

func TestMarkdownFormatter_WriteToFile(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.md")

	summary := &lint.LintSummary{
		TotalFiles:      1,
		SuccessfulFiles: 1,
		FailedFiles:     0,
		StartTime:       time.Now(),
		Results: []lint.LintResult{
			{
				File:    "test.md",
				Type:    "agent",
				Success: true,
			},
		},
	}

	formatter := NewMarkdownFormatter(false, false, outputFile)
	if err := formatter.Format(summary); err != nil {
		t.Fatalf("Format() error = %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(outputFile); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %v", err)
	}

	// Read and verify content
	content, err := os.ReadFile(outputFile)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, "# CCLint Report") {
		t.Error("Output file should contain markdown header")
	}
	if !strings.Contains(contentStr, "## Summary") {
		t.Error("Output file should contain summary section")
	}
}

func TestMarkdownFormatter_WriteToFileError(t *testing.T) {
	// Try to write to invalid path
	outputFile := "/invalid/path/that/does/not/exist/output.md"

	summary := &lint.LintSummary{
		TotalFiles: 1,
		StartTime:  time.Now(),
		Results:    []lint.LintResult{},
	}

	formatter := NewMarkdownFormatter(false, false, outputFile)
	err := formatter.Format(summary)

	if err == nil {
		t.Error("Expected error when writing to invalid path")
	}
	if !strings.Contains(err.Error(), "error writing to file") {
		t.Errorf("Error message = %q, want to contain 'error writing to file'", err.Error())
	}
}

func TestMarkdownFormatter_HelperFunctions(t *testing.T) {
	t.Run("getStatusEmoji", func(t *testing.T) {
		tests := []struct {
			success bool
			want    string
		}{
			{true, "✅"},
			{false, "❌"},
		}

		for _, tt := range tests {
			got := getStatusEmoji(tt.success)
			if got != tt.want {
				t.Errorf("getStatusEmoji(%v) = %q, want %q", tt.success, got, tt.want)
			}
		}
	})

	t.Run("createAnchor", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{"test.md", "testmd"},
			{"test file.md", "test-filemd"},
			{"path/to/file.md", "path-to-filemd"},
			{"Test With Spaces.md", "test-with-spacesmd"},
		}

		for _, tt := range tests {
			got := createAnchor(tt.input)
			if got != tt.want {
				t.Errorf("createAnchor(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})

	t.Run("formatSourceTag", func(t *testing.T) {
		tests := []struct {
			input string
			want  string
		}{
			{cue.SourceAnthropicDocs, "docs"},
			{cue.SourceCClintObserve, "cclint"},
			{"unknown-source", "unknown-source"},
			{"", ""},
		}

		for _, tt := range tests {
			got := formatSourceTag(tt.input)
			if got != tt.want {
				t.Errorf("formatSourceTag(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})
}

func TestMarkdownFormatter_VerboseMode(t *testing.T) {
	summary := &lint.LintSummary{
		TotalFiles:      2,
		SuccessfulFiles: 2,
		FailedFiles:     0,
		StartTime:       time.Now(),
		Results: []lint.LintResult{
			{
				File:    "test1.md",
				Type:    "agent",
				Success: true,
			},
			{
				File:    "test2.md",
				Type:    "command",
				Success: true,
			},
		},
	}

	t.Run("verbose mode includes all files", func(t *testing.T) {
		output := captureStdout(t, func() {
			formatter := NewMarkdownFormatter(false, true, "")
			_ = formatter.Format(summary)
		})

		if !strings.Contains(output, "### test1.md") {
			t.Error("Verbose mode should include test1.md")
		}
		if !strings.Contains(output, "### test2.md") {
			t.Error("Verbose mode should include test2.md")
		}
	})

	t.Run("non-verbose mode hides successful files", func(t *testing.T) {
		output := captureStdout(t, func() {
			formatter := NewMarkdownFormatter(false, false, "")
			_ = formatter.Format(summary)
		})

		if strings.Contains(output, "### test1.md") {
			t.Error("Non-verbose mode should not include successful files")
		}
		if strings.Contains(output, "### test2.md") {
			t.Error("Non-verbose mode should not include successful files")
		}
	})
}

func TestMarkdownFormatter_TableOfContents(t *testing.T) {
	t.Run("single file - no TOC", func(t *testing.T) {
		summary := &lint.LintSummary{
			TotalFiles:      1,
			SuccessfulFiles: 0,
			FailedFiles:     1,
			StartTime:       time.Now(),
			Results: []lint.LintResult{
				{
					File:    "test.md",
					Type:    "agent",
					Success: false,
					Errors: []cue.ValidationError{
						{File: "test.md", Message: "error", Severity: "error", Line: 1},
					},
				},
			},
		}

		output := captureStdout(t, func() {
			formatter := NewMarkdownFormatter(false, false, "")
			_ = formatter.Format(summary)
		})

		if strings.Contains(output, "### Files") {
			t.Error("Single file should not generate TOC")
		}
	})

	t.Run("multiple files - includes TOC", func(t *testing.T) {
		summary := &lint.LintSummary{
			TotalFiles:      2,
			SuccessfulFiles: 0,
			FailedFiles:     2,
			StartTime:       time.Now(),
			Results: []lint.LintResult{
				{
					File:    "./path/to/test1.md",
					Type:    "agent",
					Success: false,
					Errors: []cue.ValidationError{
						{File: "test1.md", Message: "error", Severity: "error", Line: 1},
					},
				},
				{
					File:    "./test2.md",
					Type:    "command",
					Success: false,
					Errors: []cue.ValidationError{
						{File: "test2.md", Message: "error", Severity: "error", Line: 1},
					},
				},
			},
		}

		output := captureStdout(t, func() {
			formatter := NewMarkdownFormatter(false, false, "")
			_ = formatter.Format(summary)
		})

		if !strings.Contains(output, "### Files") {
			t.Error("Multiple files should generate TOC")
		}
		if !strings.Contains(output, "- [path/to/test1.md](#path-to-test1md)") {
			t.Error("TOC should include path/to/test1.md with proper anchor")
		}
		if !strings.Contains(output, "- [test2.md](#test2md)") {
			t.Error("TOC should include test2.md with proper anchor")
		}
	})
}

func TestMarkdownFormatter_SourceTags(t *testing.T) {
	summary := &lint.LintSummary{
		TotalFiles:      1,
		SuccessfulFiles: 0,
		FailedFiles:     1,
		TotalErrors:     3,
		StartTime:       time.Now(),
		Results: []lint.LintResult{
			{
				File:    "test.md",
				Type:    "agent",
				Success: false,
				Errors: []cue.ValidationError{
					{
						File:     "test.md",
						Message:  "error with docs source",
						Severity: "error",
						Source:   cue.SourceAnthropicDocs,
						Line:     1,
					},
					{
						File:     "test.md",
						Message:  "error with cclint source",
						Severity: "error",
						Source:   cue.SourceCClintObserve,
						Line:     2,
					},
					{
						File:     "test.md",
						Message:  "error without source",
						Severity: "error",
						Line:     3,
					},
				},
			},
		},
	}

	output := captureStdout(t, func() {
		formatter := NewMarkdownFormatter(false, false, "")
		_ = formatter.Format(summary)
	})

	if !strings.Contains(output, "`[docs]`") {
		t.Error("Should format anthropic-docs source as [docs]")
	}
	if !strings.Contains(output, "`[cclint]`") {
		t.Error("Should format cclint-observation source as [cclint]")
	}

	// Count occurrences of "error without source" line
	lines := strings.Split(output, "\n")
	var errorWithoutSourceLine string
	for _, line := range lines {
		if strings.Contains(line, "error without source") {
			errorWithoutSourceLine = line
			break
		}
	}
	if strings.Contains(errorWithoutSourceLine, "`[") {
		t.Error("Should not add source tag when source is empty")
	}
}

func TestNewMarkdownFormatter(t *testing.T) {
	tests := []struct {
		name       string
		quiet      bool
		verbose    bool
		outputFile string
	}{
		{"default", false, false, ""},
		{"quiet", true, false, ""},
		{"verbose", false, true, ""},
		{"with output file", false, false, "output.md"},
		{"all options", true, true, "report.md"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewMarkdownFormatter(tt.quiet, tt.verbose, tt.outputFile)

			if formatter == nil {
				t.Fatal("NewMarkdownFormatter returned nil")
			}
			if formatter.quiet != tt.quiet {
				t.Errorf("quiet = %v, want %v", formatter.quiet, tt.quiet)
			}
			if formatter.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", formatter.verbose, tt.verbose)
			}
			if formatter.outputFile != tt.outputFile {
				t.Errorf("outputFile = %q, want %q", formatter.outputFile, tt.outputFile)
			}
		})
	}
}

func TestMarkdownFormatter_Separators(t *testing.T) {
	summary := &lint.LintSummary{
		TotalFiles:      2,
		SuccessfulFiles: 0,
		FailedFiles:     2,
		TotalErrors:     2,
		StartTime:       time.Now(),
		Results: []lint.LintResult{
			{
				File:    "test1.md",
				Type:    "agent",
				Success: false,
				Errors: []cue.ValidationError{
					{File: "test1.md", Message: "error 1", Severity: "error", Line: 1},
				},
			},
			{
				File:    "test2.md",
				Type:    "command",
				Success: false,
				Errors: []cue.ValidationError{
					{File: "test2.md", Message: "error 2", Severity: "error", Line: 1},
				},
			},
		},
	}

	output := captureStdout(t, func() {
		formatter := NewMarkdownFormatter(false, false, "")
		_ = formatter.Format(summary)
	})

	// In non-verbose mode, files should have separator
	separatorCount := strings.Count(output, "---")
	if separatorCount < 1 {
		t.Errorf("Expected at least 1 separator (---) in non-verbose mode, got %d", separatorCount)
	}
}
