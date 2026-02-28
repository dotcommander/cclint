package lint

import (
	"testing"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/cue"
)

func TestFilterResults(t *testing.T) {
	tests := []struct {
		name              string
		summary           *LintSummary
		baseline          *baseline.Baseline
		wantTotalIgnored  int
		wantErrorsIgnored int
		wantSuggsIgnored  int
	}{
		{
			name: "no baseline",
			summary: &LintSummary{
				Results: []LintResult{
					{
						Errors: []cue.ValidationError{
							{File: "test.md", Message: "error 1"},
						},
					},
				},
			},
			baseline:          nil,
			wantTotalIgnored:  0,
			wantErrorsIgnored: 0,
			wantSuggsIgnored:  0,
		},
		{
			name: "empty baseline",
			summary: &LintSummary{
				Results: []LintResult{
					{
						Errors: []cue.ValidationError{
							{File: "test.md", Message: "error 1"},
						},
					},
				},
			},
			baseline:          baseline.CreateBaseline([]cue.ValidationError{}),
			wantTotalIgnored:  0,
			wantErrorsIgnored: 0,
			wantSuggsIgnored:  0,
		},
		{
			name: "with baseline matches",
			summary: &LintSummary{
				Results: []LintResult{
					{
						Errors: []cue.ValidationError{
							{File: "test.md", Message: "known error", Source: "test"},
							{File: "test.md", Message: "new error", Source: "test"},
						},
						Suggestions: []cue.ValidationError{
							{File: "test.md", Message: "known suggestion", Source: "test"},
						},
					},
				},
			},
			baseline: baseline.CreateBaseline([]cue.ValidationError{
				{File: "test.md", Message: "known error", Source: "test"},
				{File: "test.md", Message: "known suggestion", Source: "test"},
			}),
			wantTotalIgnored:  2,
			wantErrorsIgnored: 1,
			wantSuggsIgnored:  1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			totalIgnored, errorsIgnored, suggsIgnored := FilterResults(tt.summary, tt.baseline)

			if totalIgnored != tt.wantTotalIgnored {
				t.Errorf("FilterResults() totalIgnored = %d, want %d", totalIgnored, tt.wantTotalIgnored)
			}

			if errorsIgnored != tt.wantErrorsIgnored {
				t.Errorf("FilterResults() errorsIgnored = %d, want %d", errorsIgnored, tt.wantErrorsIgnored)
			}

			if suggsIgnored != tt.wantSuggsIgnored {
				t.Errorf("FilterResults() suggsIgnored = %d, want %d", suggsIgnored, tt.wantSuggsIgnored)
			}
		})
	}
}

func TestCollectAllIssues(t *testing.T) {
	summary := &LintSummary{
		Results: []LintResult{
			{
				Errors: []cue.ValidationError{
					{File: "test1.md", Message: "error 1"},
					{File: "test1.md", Message: "error 2"},
				},
				Warnings: []cue.ValidationError{
					{File: "test1.md", Message: "warning 1"},
				},
				Suggestions: []cue.ValidationError{
					{File: "test1.md", Message: "suggestion 1"},
				},
			},
			{
				Errors: []cue.ValidationError{
					{File: "test2.md", Message: "error 3"},
				},
			},
		},
	}

	issues := CollectAllIssues(summary)

	expectedCount := 5 // 3 errors + 1 warning + 1 suggestion
	if len(issues) != expectedCount {
		t.Errorf("CollectAllIssues() returned %d issues, want %d", len(issues), expectedCount)
	}
}

func TestFilterResultsUpdatesSuccess(t *testing.T) {
	summary := &LintSummary{
		Results: []LintResult{
			{
				Success: false,
				Errors: []cue.ValidationError{
					{File: "test.md", Message: "known error", Source: "test"},
				},
			},
		},
	}

	b := baseline.CreateBaseline([]cue.ValidationError{
		{File: "test.md", Message: "known error", Source: "test"},
	})

	FilterResults(summary, b)

	if !summary.Results[0].Success {
		t.Error("FilterResults() should update Success to true when all errors are filtered")
	}

	if summary.SuccessfulFiles != 1 {
		t.Errorf("FilterResults() SuccessfulFiles = %d, want 1", summary.SuccessfulFiles)
	}

	if summary.FailedFiles != 0 {
		t.Errorf("FilterResults() FailedFiles = %d, want 0", summary.FailedFiles)
	}
}
