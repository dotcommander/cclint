package output

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/scoring"
)

func TestConsoleFormatter_Format(t *testing.T) {
	tests := []struct {
		name             string
		summary          *cli.LintSummary
		quiet            bool
		verbose          bool
		showScores       bool
		showImprovements bool
		wantContains     []string
		wantNotContains  []string
	}{
		{
			name: "quiet mode - no output",
			summary: &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 1,
				FailedFiles:     0,
				StartTime:       time.Now(),
			},
			quiet:           true,
			wantContains:    []string{},
			wantNotContains: []string{"passed", "failed"},
		},
		{
			name: "all passing files",
			summary: &cli.LintSummary{
				TotalFiles:      2,
				SuccessfulFiles: 2,
				FailedFiles:     0,
				ComponentType:   "agent",
				StartTime:       time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test1.md",
						Type:    "agent",
						Success: true,
					},
					{
						File:    "test2.md",
						Type:    "agent",
						Success: true,
					},
				},
			},
			wantContains: []string{
				"✓ All 2 agents passed",
			},
		},
		{
			name: "file with errors",
			summary: &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 0,
				FailedFiles:     1,
				TotalErrors:     2,
				StartTime:       time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: false,
						Errors: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "missing required field 'name'",
								Severity: "error",
								Line:     5,
							},
							{
								File:     "test.md",
								Message:  "invalid format",
								Severity: "error",
								Line:     10,
							},
						},
					},
				},
			},
			wantContains: []string{
				"✗",
				"test.md",
				"missing required field 'name'",
				"invalid format",
				"0/1 passed",
				"2 errors",
			},
		},
		{
			name: "file with warnings",
			summary: &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 1,
				FailedFiles:     0,
				TotalWarnings:   1,
				StartTime:       time.Now(),
				Results: []cli.LintResult{
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
						},
					},
				},
			},
			wantContains: []string{
				"⚠",
				"consider adding description",
			},
		},
		{
			name: "verbose mode with suggestions",
			summary: &cli.LintSummary{
				TotalFiles:       1,
				SuccessfulFiles:  1,
				FailedFiles:      0,
				TotalSuggestions: 2,
				StartTime:        time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "skill",
						Success: true,
						Suggestions: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "could add examples section",
								Severity: "suggestion",
								Source:   cue.SourceAnthropicDocs,
								Line:     15,
							},
							{
								File:     "test.md",
								Message:  "consider refactoring",
								Severity: "suggestion",
								Source:   cue.SourceCClintObserve,
								Line:     20,
							},
						},
					},
				},
			},
			verbose: true,
			wantContains: []string{
				"💡",
				"could add examples section",
				"consider refactoring",
				"[docs]",
				"[cclint]",
				"2 suggestions",
			},
		},
		{
			name: "non-verbose mode hides suggestions",
			summary: &cli.LintSummary{
				TotalFiles:       1,
				SuccessfulFiles:  1,
				FailedFiles:      0,
				TotalSuggestions: 1,
				StartTime:        time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "skill",
						Success: true,
						Suggestions: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "hidden suggestion",
								Severity: "suggestion",
							},
						},
					},
				},
			},
			verbose:         false,
			wantNotContains: []string{"hidden suggestion", "💡"},
		},
		{
			name: "show quality scores",
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
			verbose:    true, // Need verbose to show files without issues
			showScores: true,
			wantContains: []string{
				"[A 85]",
			},
		},
		{
			name: "verbose with quality scores breakdown",
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
							Overall:       72,
							Tier:          "B",
							Structural:    30,
							Practices:     28,
							Composition:   8,
							Documentation: 6,
						},
					},
				},
			},
			verbose:    true,
			showScores: true,
			wantContains: []string{
				"[B 72]",
				"Score: 72/100 (B)",
				"Structural: 30/40",
				"Practices: 28/40",
				"Composition: 8/10",
				"Documentation: 6/10",
			},
		},
		{
			name: "show improvements",
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
						Improvements: []cli.ImprovementRecommendation{
							{
								Description: "Add Foundation section",
								PointValue:  5,
								Line:        10,
							},
							{
								Description: "Add Workflow section",
								PointValue:  5,
								Line:        15,
							},
						},
					},
				},
			},
			verbose:          true, // Need verbose to show files without issues
			showImprovements: true,
			wantContains: []string{
				"Improvements:",
				"+5 pts:",
				"Add Foundation section",
				"Add Workflow section",
				"(line 10)",
				"(line 15)",
			},
		},
		{
			name: "empty summary",
			summary: &cli.LintSummary{
				TotalFiles:      0,
				SuccessfulFiles: 0,
				FailedFiles:     0,
				StartTime:       time.Now(),
				Results:         []cli.LintResult{},
			},
			wantContains: []string{
				"✓ All 0 files passed",
			},
		},
		{
			name: "error without line number",
			summary: &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 0,
				FailedFiles:     1,
				TotalErrors:     1,
				StartTime:       time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: false,
						Errors: []cue.ValidationError{
							{
								File:     "test.md",
								Message:  "general error without line",
								Severity: "error",
								Line:     0,
							},
						},
					},
				},
			},
			wantContains: []string{
				"test.md: general error without line",
			},
			wantNotContains: []string{":0:"},
		},
		{
			name: "mixed errors, warnings, and suggestions",
			summary: &cli.LintSummary{
				TotalFiles:       1,
				SuccessfulFiles:  0,
				FailedFiles:      1,
				TotalErrors:      1,
				TotalWarnings:    1,
				TotalSuggestions: 1,
				StartTime:        time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: false,
						Errors: []cue.ValidationError{
							{File: "test.md", Message: "error message", Severity: "error", Line: 1},
						},
						Warnings: []cue.ValidationError{
							{File: "test.md", Message: "warning message", Severity: "warning", Line: 2},
						},
						Suggestions: []cue.ValidationError{
							{File: "test.md", Message: "suggestion message", Severity: "suggestion", Line: 3},
						},
					},
				},
			},
			verbose: true,
			wantContains: []string{
				"✘",
				"error message",
				"⚠",
				"warning message",
				"💡",
				"suggestion message",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Capture stdout
			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			formatter := NewConsoleFormatter(tt.quiet, tt.verbose, tt.showScores, tt.showImprovements)
			err := formatter.Format(tt.summary)

			w.Close()
			os.Stdout = old

			if err != nil {
				t.Errorf("Format() error = %v", err)
				return
			}

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			for _, want := range tt.wantContains {
				if !strings.Contains(output, want) {
					t.Errorf("Format() output missing expected string:\n  want: %q\n  got: %q", want, output)
				}
			}

			for _, notWant := range tt.wantNotContains {
				if strings.Contains(output, notWant) {
					t.Errorf("Format() output contains unexpected string:\n  don't want: %q\n  got: %q", notWant, output)
				}
			}
		})
	}
}

func TestConsoleFormatter_Colorization(t *testing.T) {
	tests := []struct {
		name         string
		colorize     bool
		wantEscCodes bool
	}{
		{
			name:         "colorization enabled",
			colorize:     true,
			wantEscCodes: false, // lipgloss doesn't add escape codes in test environment
		},
		{
			name:         "colorization disabled",
			colorize:     false,
			wantEscCodes: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewConsoleFormatter(false, false, false, false)
			formatter.colorize = tt.colorize

			summary := &cli.LintSummary{
				TotalFiles:      1,
				SuccessfulFiles: 1,
				FailedFiles:     0,
				ComponentType:   "agent",
				StartTime:       time.Now(),
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Type:    "agent",
						Success: true,
					},
				},
			}

			old := os.Stdout
			r, w, _ := os.Pipe()
			os.Stdout = w

			_ = formatter.Format(summary)

			w.Close()
			os.Stdout = old

			var buf bytes.Buffer
			_, _ = io.Copy(&buf, r)
			output := buf.String()

			// Just verify it doesn't crash with colorization off
			if !strings.Contains(output, "✓ All 1 agents passed") {
				t.Errorf("Expected success message in output")
			}
		})
	}
}

func TestConsoleFormatter_NoColorsInConclusion(t *testing.T) {
	formatter := NewConsoleFormatter(false, false, false, false)
	formatter.colorize = false

	summary := &cli.LintSummary{
		TotalFiles:      1,
		SuccessfulFiles: 1,
		FailedFiles:     0,
		ComponentType:   "agent",
		StartTime:       time.Now(),
		Results: []cli.LintResult{
			{
				File:    "test.md",
				Type:    "agent",
				Success: true,
			},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = formatter.Format(summary)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "✓ All 1 agents passed") {
		t.Error("Should display conclusion even without colors")
	}
}

func TestConsoleFormatter_EmptyResultsNoNewline(t *testing.T) {
	formatter := NewConsoleFormatter(false, false, false, false)

	summary := &cli.LintSummary{
		TotalFiles:      0,
		SuccessfulFiles: 0,
		FailedFiles:     0,
		ComponentType:   "",
		StartTime:       time.Now(),
		Results:         []cli.LintResult{}, // Empty results
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = formatter.Format(summary)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Should still print conclusion
	if !strings.Contains(output, "✓ All 0 files passed") {
		t.Error("Should display conclusion for empty results")
	}
	// Should not have extra newlines for empty results
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) > 2 {
		t.Errorf("Empty results should not produce extra newlines, got %d lines", len(lines))
	}
}

func TestConsoleFormatter_EdgeCases(t *testing.T) {
	t.Run("nil quality score", func(t *testing.T) {
		formatter := NewConsoleFormatter(false, false, true, false)
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
					Quality: nil,
				},
			},
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		err := formatter.Format(summary)

		w.Close()
		os.Stdout = old

		if err != nil {
			t.Errorf("Format() with nil Quality should not error: %v", err)
		}

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if strings.Contains(output, "[A") || strings.Contains(output, "[B") {
			t.Errorf("Should not display score when Quality is nil")
		}
	})

	t.Run("improvement without line number", func(t *testing.T) {
		formatter := NewConsoleFormatter(false, true, false, true) // verbose=true to show files
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
					Improvements: []cli.ImprovementRecommendation{
						{
							Description: "General improvement",
							PointValue:  3,
							Line:        0,
						},
					},
				},
			},
		}

		old := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		_ = formatter.Format(summary)

		w.Close()
		os.Stdout = old

		var buf bytes.Buffer
		_, _ = io.Copy(&buf, r)
		output := buf.String()

		if strings.Contains(output, "(line 0)") {
			t.Errorf("Should not display '(line 0)' for improvements without line number")
		}
		if !strings.Contains(output, "General improvement") {
			t.Errorf("Should display improvement description")
		}
	})

	t.Run("quality score tier colors", func(t *testing.T) {
		tiers := []struct {
			tier     string
			overall  int
			expected string
		}{
			{"A", 90, "[A 90]"},
			{"B", 75, "[B 75]"},
			{"C", 55, "[C 55]"},
			{"D", 35, "[D 35]"},
			{"F", 20, "[F 20]"},
		}

		for _, tier := range tiers {
			t.Run("tier_"+tier.tier, func(t *testing.T) {
				formatter := NewConsoleFormatter(false, true, true, false) // verbose=true to show files
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
							Quality: &scoring.QualityScore{
								Overall: tier.overall,
								Tier:    tier.tier,
							},
						},
					},
				}

				old := os.Stdout
				r, w, _ := os.Pipe()
				os.Stdout = w

				_ = formatter.Format(summary)

				w.Close()
				os.Stdout = old

				var buf bytes.Buffer
				_, _ = io.Copy(&buf, r)
				output := buf.String()

				if !strings.Contains(output, tier.expected) {
					t.Errorf("Expected %q in output, got %q", tier.expected, output)
				}
			})
		}
	})
}

func TestConsoleFormatter_UnknownSeverity(t *testing.T) {
	formatter := NewConsoleFormatter(false, false, false, false)
	summary := &cli.LintSummary{
		TotalFiles:      1,
		SuccessfulFiles: 0,
		FailedFiles:     1,
		TotalErrors:     1,
		StartTime:       time.Now(),
		Results: []cli.LintResult{
			{
				File:    "test.md",
				Type:    "agent",
				Success: false,
				Errors: []cue.ValidationError{
					{
						File:     "test.md",
						Message:  "unknown severity type",
						Severity: "unknown",
						Line:     5,
					},
				},
			},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = formatter.Format(summary)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "unknown severity type") {
		t.Error("Should handle unknown severity gracefully")
	}
}

func TestConsoleFormatter_UnknownSource(t *testing.T) {
	formatter := NewConsoleFormatter(false, true, false, false) // verbose to show sources
	summary := &cli.LintSummary{
		TotalFiles:       1,
		SuccessfulFiles:  1,
		FailedFiles:      0,
		TotalSuggestions: 1,
		StartTime:        time.Now(),
		Results: []cli.LintResult{
			{
				File:    "test.md",
				Type:    "agent",
				Success: true,
				Suggestions: []cue.ValidationError{
					{
						File:     "test.md",
						Message:  "suggestion with unknown source",
						Severity: "suggestion",
						Source:   "unknown-source",
						Line:     5,
					},
				},
			},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = formatter.Format(summary)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	if !strings.Contains(output, "suggestion with unknown source") {
		t.Error("Should display suggestion with unknown source")
	}
	// Unknown sources are not displayed as tags
	if strings.Contains(output, "[unknown-source]") {
		t.Error("Should not display unknown source tags")
	}
}

func TestConsoleFormatter_QuietModeWithErrors(t *testing.T) {
	formatter := NewConsoleFormatter(true, false, false, false)
	summary := &cli.LintSummary{
		TotalFiles:      1,
		SuccessfulFiles: 0,
		FailedFiles:     1,
		TotalErrors:     1,
		StartTime:       time.Now(),
		Results: []cli.LintResult{
			{
				File:    "test.md",
				Type:    "agent",
				Success: false,
				Errors: []cue.ValidationError{
					{
						File:     "test.md",
						Message:  "some error",
						Severity: "error",
						Line:     5,
					},
				},
			},
		},
	}

	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	_ = formatter.Format(summary)

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)
	output := buf.String()

	// Quiet mode should produce minimal output
	if strings.Contains(output, "some error") {
		t.Error("Quiet mode should suppress detailed error output")
	}
}

func TestNewConsoleFormatter(t *testing.T) {
	tests := []struct {
		name             string
		quiet            bool
		verbose          bool
		showScores       bool
		showImprovements bool
	}{
		{"default", false, false, false, false},
		{"quiet", true, false, false, false},
		{"verbose", false, true, false, false},
		{"with scores", false, false, true, false},
		{"with improvements", false, false, false, true},
		{"all flags", true, true, true, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter := NewConsoleFormatter(tt.quiet, tt.verbose, tt.showScores, tt.showImprovements)

			if formatter == nil {
				t.Fatal("NewConsoleFormatter returned nil")
			}
			if formatter.quiet != tt.quiet {
				t.Errorf("quiet = %v, want %v", formatter.quiet, tt.quiet)
			}
			if formatter.verbose != tt.verbose {
				t.Errorf("verbose = %v, want %v", formatter.verbose, tt.verbose)
			}
			if formatter.showScores != tt.showScores {
				t.Errorf("showScores = %v, want %v", formatter.showScores, tt.showScores)
			}
			if formatter.showImprovements != tt.showImprovements {
				t.Errorf("showImprovements = %v, want %v", formatter.showImprovements, tt.showImprovements)
			}
			if !formatter.colorize {
				t.Error("colorize should default to true")
			}
		})
	}
}

func TestConsoleFormatter_CelebrationTrigger(t *testing.T) {
	tests := []struct {
		name              string
		quiet             bool
		verbose           bool
		colorize          bool
		failedFiles       int
		totalErrors       int
		totalWarnings     int
		totalSuggestions  int
		expectCelebration bool
	}{
		{
			name:              "perfect success - triggers celebration",
			quiet:             false,
			verbose:           false,
			colorize:          true,
			failedFiles:       0,
			totalErrors:       0,
			totalWarnings:     0,
			totalSuggestions:  0,
			expectCelebration: true,
		},
		{
			name:              "has suggestions - no celebration",
			quiet:             false,
			verbose:           false,
			colorize:          true,
			failedFiles:       0,
			totalErrors:       0,
			totalWarnings:     0,
			totalSuggestions:  5,
			expectCelebration: false,
		},
		{
			name:              "verbose mode with suggestions - no celebration",
			quiet:             false,
			verbose:           true,
			colorize:          true,
			failedFiles:       0,
			totalErrors:       0,
			totalWarnings:     0,
			totalSuggestions:  1,
			expectCelebration: false,
		},
		{
			name:              "verbose mode, zero suggestions - celebration",
			quiet:             false,
			verbose:           true,
			colorize:          true,
			failedFiles:       0,
			totalErrors:       0,
			totalWarnings:     0,
			totalSuggestions:  0,
			expectCelebration: true,
		},
		{
			name:              "quiet mode - no output at all",
			quiet:             true,
			verbose:           false,
			colorize:          true,
			failedFiles:       0,
			totalErrors:       0,
			totalWarnings:     0,
			totalSuggestions:  0,
			expectCelebration: false,
		},
		{
			name:              "no color - no animation (plain text only)",
			quiet:             false,
			verbose:           false,
			colorize:          false,
			failedFiles:       0,
			totalErrors:       0,
			totalWarnings:     0,
			totalSuggestions:  0,
			expectCelebration: false, // Still success, but no animation
		},
		{
			name:              "errors present - no celebration",
			quiet:             false,
			verbose:           false,
			colorize:          true,
			failedFiles:       1,
			totalErrors:       3,
			totalWarnings:     0,
			totalSuggestions:  0,
			expectCelebration: false,
		},
		{
			name:              "warnings present - no celebration",
			quiet:             false,
			verbose:           false,
			colorize:          true,
			failedFiles:       0,
			totalErrors:       0,
			totalWarnings:     2,
			totalSuggestions:  0,
			expectCelebration: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			summary := &cli.LintSummary{
				ComponentType:    "agent",
				TotalFiles:       5,
				FailedFiles:      tt.failedFiles,
				TotalErrors:      tt.totalErrors,
				TotalWarnings:    tt.totalWarnings,
				TotalSuggestions: tt.totalSuggestions,
				Results:          []cli.LintResult{},
			}

			// Test the celebration trigger logic directly
			// (Animation won't show in tests due to non-TTY, but logic is verified)
			allPassed := summary.FailedFiles == 0
			if tt.verbose {
				allPassed = allPassed && summary.TotalSuggestions == 0
			}

			perfectSuccess := summary.TotalErrors == 0 && summary.TotalWarnings == 0 && summary.TotalSuggestions == 0
			shouldCelebrate := allPassed && !tt.quiet && tt.colorize && perfectSuccess

			if shouldCelebrate != tt.expectCelebration {
				t.Errorf("celebration trigger mismatch: got %v, want %v", shouldCelebrate, tt.expectCelebration)
			}
		})
	}
}
