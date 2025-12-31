package output

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/scoring"
)

func TestJSONFormatter_Format(t *testing.T) {
	tests := []struct {
		name       string
		summary    *cli.LintSummary
		quiet      bool
		indent     bool
		outputFile string
		validate   func(t *testing.T, output string)
	}{
		{
			name: "basic json output - compact",
			summary: &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 1,
				FailedFiles:     0,
				TotalErrors:     0,
				TotalWarnings:   0,
				StartTime:       time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: true,
					},
				},
			},
			indent: false,
			validate: func(t *testing.T, output string) {
				var report JSONReport
				if err := json.Unmarshal([]byte(output), &report); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if report.Header.Tool != "cclint" {
					t.Errorf("Tool = %q, want %q", report.Header.Tool, "cclint")
				}
				if report.Summary.TotalFiles != 1 {
					t.Errorf("TotalFiles = %d, want 1", report.Summary.TotalFiles)
				}
				if report.Summary.SuccessfulFiles != 1 {
					t.Errorf("SuccessfulFiles = %d, want 1", report.Summary.SuccessfulFiles)
				}
				if len(report.Results) != 1 {
					t.Fatalf("Results length = %d, want 1", len(report.Results))
				}
				if report.Results[0].File != "test.md" {
					t.Errorf("Results[0].File = %q, want %q", report.Results[0].File, "test.md")
				}
				if !report.Results[0].Success {
					t.Error("Results[0].Success = false, want true")
				}
			},
		},
		{
			name: "indented json output",
			summary: &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 1,
				FailedFiles:     0,
				TotalErrors:     0,
				TotalWarnings:   0,
				StartTime:       time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "command",
						Success: true,
					},
				},
			},
			indent: true,
			validate: func(t *testing.T, output string) {
				if !strings.Contains(output, "\n") {
					t.Error("Indented output should contain newlines")
				}
				if !strings.Contains(output, "  ") {
					t.Error("Indented output should contain spaces for indentation")
				}
			},
		},
		{
			name: "json with errors and warnings",
			summary: &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 0,
				FailedFiles:     1,
				TotalErrors:     2,
				TotalWarnings:   1,
				StartTime:       time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: false,
						Errors: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "missing name",
								Severity: "error",
								Source:   cue.SourceAnthropicDocs,
								Line:     5,
								Column:   10,
							},
							{
								File:     "test.md",
								Message:  "invalid format",
								Severity: "error",
								Source:   cue.SourceCClintObserve,
								Line:     15,
							},
						},
						Warnings: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "consider refactoring",
								Severity: "warning",
								Line:     20,
							},
						},
					},
				},
			},
			indent: true,
			validate: func(t *testing.T, output string) {
				var report JSONReport
				if err := json.Unmarshal([]byte(output), &report); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if report.Summary.TotalErrors != 2 {
					t.Errorf("TotalErrors = %d, want 2", report.Summary.TotalErrors)
				}
				if report.Summary.TotalWarnings != 1 {
					t.Errorf("TotalWarnings = %d, want 1", report.Summary.TotalWarnings)
				}
				if len(report.Results[0].Errors) != 2 {
					t.Errorf("Errors length = %d, want 2", len(report.Results[0].Errors))
				}
				if report.Results[0].Errors[0].Message != "missing name" {
					t.Errorf("Error message = %q, want 'missing name'", report.Results[0].Errors[0].Message)
				}
				if report.Results[0].Errors[0].Source != cue.SourceAnthropicDocs {
					t.Errorf("Error source = %q, want %q", report.Results[0].Errors[0].Source, cue.SourceAnthropicDocs)
				}
				if report.Results[0].Errors[0].Line != 5 {
					t.Errorf("Error line = %d, want 5", report.Results[0].Errors[0].Line)
				}
				if report.Results[0].Errors[0].Column != 10 {
					t.Errorf("Error column = %d, want 10", report.Results[0].Errors[0].Column)
				}
			},
		},
		{
			name: "json with quality scores",
			summary: &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 1,
				FailedFiles:     0,
				StartTime:       time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: true,
						Quality: &scoring.QualityScore{
							Overall:       85,
							Tier:          "A",
							Structural:    35,
							Practices:     38,
							Composition:   7,
							Documentation: 5,
						},
					},
				},
			},
			indent: true,
			validate: func(t *testing.T, output string) {
				var report JSONReport
				if err := json.Unmarshal([]byte(output), &report); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if report.Results[0].Quality == nil {
					t.Fatal("Quality should not be nil")
				}
				if report.Results[0].Quality.Overall != 85 {
					t.Errorf("Overall = %d, want 85", report.Results[0].Quality.Overall)
				}
				if report.Results[0].Quality.Tier != "A" {
					t.Errorf("Tier = %q, want 'A'", report.Results[0].Quality.Tier)
				}
				if report.Results[0].Quality.Structural != 35 {
					t.Errorf("Structural = %d, want 35", report.Results[0].Quality.Structural)
				}
			},
		},
		{
			name: "empty results",
			summary: &cli.LintSummary{
				TotalFiles:      0,
				SuccessfulFiles: 0,
				FailedFiles:     0,
				TotalErrors:     0,
				TotalWarnings:   0,
				StartTime:       time.Now(),
				Results:         []cli.LintResult{},
			},
			indent: true,
			validate: func(t *testing.T, output string) {
				var report JSONReport
				if err := json.Unmarshal([]byte(output), &report); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if report.Summary.TotalFiles != 0 {
					t.Errorf("TotalFiles = %d, want 0", report.Summary.TotalFiles)
				}
				if len(report.Results) != 0 {
					t.Errorf("Results length = %d, want 0", len(report.Results))
				}
			},
		},
		{
			name: "result without quality score",
			summary: &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 1,
				FailedFiles:     0,
				StartTime:       time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: true,
						Quality: nil, // No quality score
					},
				},
			},
			indent: false,
			validate: func(t *testing.T, output string) {
				var report JSONReport
				if err := json.Unmarshal([]byte(output), &report); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if report.Results[0].Quality != nil {
					t.Error("Quality should be nil when not provided")
				}
			},
		},
		{
			name: "multiple files",
			summary: &cli.LintSummary{
				TotalFiles:      3,
				SuccessfulFiles: 2,
				FailedFiles:     1,
				TotalErrors:     1,
				TotalWarnings:   0,
				StartTime:       time.Now(),
				Results: []cli.LintResult{
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
			indent: true,
			validate: func(t *testing.T, output string) {
				var report JSONReport
				if err := json.Unmarshal([]byte(output), &report); err != nil {
					t.Fatalf("Failed to parse JSON: %v", err)
				}

				if len(report.Results) != 3 {
					t.Errorf("Results length = %d, want 3", len(report.Results))
				}
				if report.Results[1].Success {
					t.Error("Results[1].Success = true, want false")
				}
				if len(report.Results[1].Errors) != 1 {
					t.Errorf("Results[1].Errors length = %d, want 1", len(report.Results[1].Errors))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout if not writing to file
			var output string
			if tt.outputFile == "" {
				oldStdout := captureStdout(t, func() {
					formatter := NewJSONFormatter(tt.quiet, tt.indent, tt.outputFile)
					if err := formatter.Format(tt.summary); err != nil {
						t.Fatalf("Format() error = %v", err)
					}
				})
				output = oldStdout
			}

			if tt.validate != nil {
				tt.validate(t, output)
			}
		})
	}
}

func TestJSONFormatter_WriteToFile(t *testing.T) {
	tempDir := t.TempDir()
	outputFile := filepath.Join(tempDir, "output.json")

	summary := &cli.LintSummary{
		TotalFiles:      1,
		SuccessfulFiles: 1,
		FailedFiles:     0,
		StartTime:       time.Now(),
		Results: []cli.LintResult{
			{
				File:    "test.md",
				Type:    "agent",
				Success: true,
			},
		},
	}

	formatter := NewJSONFormatter(false, true, outputFile)
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

	var report JSONReport
	if err := json.Unmarshal(content, &report); err != nil {
		t.Fatalf("Failed to parse JSON from file: %v", err)
	}

	if report.Header.Tool != "cclint" {
		t.Errorf("Tool = %q, want %q", report.Header.Tool, "cclint")
	}
}

func TestJSONFormatter_WriteToFileError(t *testing.T) {
	// Try to write to invalid path
	outputFile := "/invalid/path/that/does/not/exist/output.json"

	summary := &cli.LintSummary{
		TotalFiles: 1,
		StartTime:  time.Now(),
		Results:    []cli.LintResult{},
	}

	formatter := NewJSONFormatter(false, true, outputFile)
	err := formatter.Format(summary)

	if err == nil {
		t.Error("Expected error when writing to invalid path")
	}
	if !strings.Contains(err.Error(), "error writing to file") {
		t.Errorf("Error message = %q, want to contain 'error writing to file'", err.Error())
	}
}

func TestJSONFormatter_DurationFormat(t *testing.T) {
	startTime := time.Now().Add(-123 * time.Millisecond)
	summary := &cli.LintSummary{
		TotalFiles: 1,
		StartTime:  startTime,
		Results:    []cli.LintResult{},
	}

	output := captureStdout(t, func() {
		formatter := NewJSONFormatter(false, true, "")
		if err := formatter.Format(summary); err != nil {
			t.Fatalf("Format() error = %v", err)
		}
	})

	var report JSONReport
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Duration should be a string with time units
	if report.Summary.Duration == "" {
		t.Error("Duration should not be empty")
	}
	// Duration format is like "123ms" or "1.5s"
	if !strings.HasSuffix(report.Summary.Duration, "ms") && !strings.HasSuffix(report.Summary.Duration, "s") {
		t.Errorf("Duration format unexpected: %q", report.Summary.Duration)
	}
}

func TestNewJSONFormatter(t *testing.T) {
	tests := []struct {
		name       string
		quiet      bool
		indent     bool
		outputFile string
	}{
		{"default", false, false, ""},
		{"quiet", true, false, ""},
		{"indented", false, true, ""},
		{"with output file", false, false, "output.json"},
		{"all options", true, true, "report.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewJSONFormatter(tt.quiet, tt.indent, tt.outputFile)

			if formatter == nil {
				t.Fatal("NewJSONFormatter returned nil")
			}
			if formatter.quiet != tt.quiet {
				t.Errorf("quiet = %v, want %v", formatter.quiet, tt.quiet)
			}
			if formatter.indent != tt.indent {
				t.Errorf("indent = %v, want %v", formatter.indent, tt.indent)
			}
			if formatter.outputFile != tt.outputFile {
				t.Errorf("outputFile = %q, want %q", formatter.outputFile, tt.outputFile)
			}
		})
	}
}

func TestJSONFormatter_AllFieldsPopulated(t *testing.T) {
	summary := &cli.LintSummary{
		TotalFiles:      2,
		SuccessfulFiles: 1,
		FailedFiles:     1,
		TotalErrors:     2,
		TotalWarnings:   1,
		StartTime:       time.Now(),
		Results: []cli.LintResult{
			{
				File:     "test1.md",
				Type:     "agent",
				Success:  true,
				Duration: 100,
			},
			{
				File:     "test2.md",
				Type:     "command",
				Success:  false,
				Duration: 150,
				Errors: []cue.ValidationError{
					{
						File:     "test2.md",
						Message:  "error 1",
						Severity: "error",
						Source:   cue.SourceAnthropicDocs,
						Line:     5,
						Column:   10,
					},
					{
						File:     "test2.md",
						Message:  "error 2",
						Severity: "error",
						Source:   cue.SourceCClintObserve,
						Line:     15,
					},
				},
				Warnings: []cue.ValidationError{
					{
						File:     "test2.md",
						Message:  "warning 1",
						Severity: "warning",
						Line:     20,
					},
				},
				Quality: &scoring.QualityScore{
					Overall:       55,
					Tier:          "C",
					Structural:    20,
					Practices:     25,
					Composition:   5,
					Documentation: 5,
				},
			},
		},
	}

	output := captureStdout(t, func() {
		formatter := NewJSONFormatter(false, true, "")
		if err := formatter.Format(summary); err != nil {
			t.Fatalf("Format() error = %v", err)
		}
	})

	var report JSONReport
	if err := json.Unmarshal([]byte(output), &report); err != nil {
		t.Fatalf("Failed to parse JSON: %v", err)
	}

	// Verify header
	if report.Header.Tool != "cclint" {
		t.Errorf("Header.Tool = %q, want 'cclint'", report.Header.Tool)
	}
	if report.Header.Version != "1.0.0" {
		t.Errorf("Header.Version = %q, want '1.0.0'", report.Header.Version)
	}
	if report.Header.Timestamp == "" {
		t.Error("Header.Timestamp should not be empty")
	}

	// Verify summary
	if report.Summary.TotalFiles != 2 {
		t.Errorf("Summary.TotalFiles = %d, want 2", report.Summary.TotalFiles)
	}
	if report.Summary.SuccessfulFiles != 1 {
		t.Errorf("Summary.SuccessfulFiles = %d, want 1", report.Summary.SuccessfulFiles)
	}
	if report.Summary.FailedFiles != 1 {
		t.Errorf("Summary.FailedFiles = %d, want 1", report.Summary.FailedFiles)
	}
	if report.Summary.TotalErrors != 2 {
		t.Errorf("Summary.TotalErrors = %d, want 2", report.Summary.TotalErrors)
	}
	if report.Summary.TotalWarnings != 1 {
		t.Errorf("Summary.TotalWarnings = %d, want 1", report.Summary.TotalWarnings)
	}

	// Verify results
	if len(report.Results) != 2 {
		t.Fatalf("Results length = %d, want 2", len(report.Results))
	}

	// First result
	if report.Results[0].Duration != 100 {
		t.Errorf("Results[0].Duration = %d, want 100", report.Results[0].Duration)
	}

	// Second result with errors and quality
	if len(report.Results[1].Errors) != 2 {
		t.Errorf("Results[1].Errors length = %d, want 2", len(report.Results[1].Errors))
	}
	if report.Results[1].Errors[0].Column != 10 {
		t.Errorf("Results[1].Errors[0].Column = %d, want 10", report.Results[1].Errors[0].Column)
	}
	if report.Results[1].Quality.Overall != 55 {
		t.Errorf("Results[1].Quality.Overall = %d, want 55", report.Results[1].Quality.Overall)
	}
}

// Helper function to capture stdout
func captureStdout(t *testing.T, fn func()) string {
	old := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatal(err)
	}
	return buf.String()
}
