package output

import (
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/lint"
)

func TestBuildFlatIssues(t *testing.T) {
	t.Parallel()

	errA := cue.ValidationError{File: "a.md", Message: "missing field", Severity: "error", Line: 1}
	warnA := cue.ValidationError{File: "a.md", Message: "deprecated usage", Severity: "warning", Line: 2}
	suggA := cue.ValidationError{File: "a.md", Message: "consider adding docs", Severity: "suggestion", Line: 3}
	errB := cue.ValidationError{File: "b.md", Message: "invalid format", Severity: "error", Line: 5}

	summary := &lint.LintSummary{
		ComponentType: "agent",
		Results: []lint.LintResult{
			{
				File:        "a.md",
				Errors:      []cue.ValidationError{errA},
				Warnings:    []cue.ValidationError{warnA},
				Suggestions: []cue.ValidationError{suggA},
				Success:     false,
			},
			{
				File:    "b.md",
				Errors:  []cue.ValidationError{errB},
				Success: false,
			},
		},
	}

	t.Run("total count", func(t *testing.T) {
		t.Parallel()
		issues := BuildFlatIssues(summary)
		if len(issues) != 4 {
			t.Errorf("expected 4 issues, got %d", len(issues))
		}
	})

	t.Run("ordering within result A", func(t *testing.T) {
		t.Parallel()
		issues := BuildFlatIssues(summary)
		if issues[0].Severity != SeverityError {
			t.Errorf("index 0: expected error, got %s", issues[0].Severity)
		}
		if issues[1].Severity != SeverityWarning {
			t.Errorf("index 1: expected warning, got %s", issues[1].Severity)
		}
		if issues[2].Severity != SeveritySuggestion {
			t.Errorf("index 2: expected suggestion, got %s", issues[2].Severity)
		}
		if issues[0].ResultIndex != 0 || issues[1].ResultIndex != 0 || issues[2].ResultIndex != 0 {
			t.Errorf("result A issues should have ResultIndex==0")
		}
	})

	t.Run("result B error has ResultIndex 1", func(t *testing.T) {
		t.Parallel()
		issues := BuildFlatIssues(summary)
		if issues[3].ResultIndex != 1 {
			t.Errorf("result B error: expected ResultIndex==1, got %d", issues[3].ResultIndex)
		}
		if issues[3].Severity != SeverityError {
			t.Errorf("result B error: expected SeverityError, got %s", issues[3].Severity)
		}
	})

	t.Run("issuesForResult", func(t *testing.T) {
		t.Parallel()
		issues := BuildFlatIssues(summary)
		a := issuesForResult(issues, 0)
		b := issuesForResult(issues, 1)
		if len(a) != 3 {
			t.Errorf("issuesForResult(0): expected 3, got %d", len(a))
		}
		if len(b) != 1 {
			t.Errorf("issuesForResult(1): expected 1, got %d", len(b))
		}
	})

	t.Run("countBySeverity", func(t *testing.T) {
		t.Parallel()
		issues := BuildFlatIssues(summary)
		if n := countBySeverity(issues, SeverityError); n != 2 {
			t.Errorf("SeverityError count: expected 2, got %d", n)
		}
		if n := countBySeverity(issues, SeverityWarning); n != 1 {
			t.Errorf("SeverityWarning count: expected 1, got %d", n)
		}
		if n := countBySeverity(issues, SeveritySuggestion); n != 1 {
			t.Errorf("SeveritySuggestion count: expected 1, got %d", n)
		}
	})

	t.Run("BuildFlatIssues nil", func(t *testing.T) {
		t.Parallel()
		issues := BuildFlatIssues(nil)
		if issues != nil {
			t.Errorf("expected nil for nil summary, got %v", issues)
		}
	})

	t.Run("severityErrors warning message", func(t *testing.T) {
		t.Parallel()
		issues := BuildFlatIssues(summary)
		aIssues := issuesForResult(issues, 0)
		warnings := severityErrors(aIssues, SeverityWarning)
		if len(warnings) != 1 {
			t.Fatalf("expected 1 warning, got %d", len(warnings))
		}
		if warnings[0].Message != warnA.Message {
			t.Errorf("expected message %q, got %q", warnA.Message, warnings[0].Message)
		}
	})
}
