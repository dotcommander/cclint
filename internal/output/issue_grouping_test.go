package output

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/textutil"
)

func TestIssueGroupingDuplicatePathsAndImprovementOnly(t *testing.T) {
	t.Parallel()
	summary := &lint.LintSummary{
		TotalFiles: 4, FailedFiles: 2, SuccessfulFiles: 2,
		TotalErrors: 1, TotalWarnings: 1, TotalSuggestions: 1,
		StartTime: time.Now(),
		Results: []lint.LintResult{
			{File: "same.md", Type: "agent", Errors: []cue.ValidationError{{File: "same.md", Message: "first error"}}},
			{File: "empty.md", Type: "agent", Success: true},
			{File: "same.md", Type: "agent", Warnings: []cue.ValidationError{{File: "same.md", Message: "second warning"}}, Suggestions: []cue.ValidationError{{File: "same.md", Message: "second suggestion"}}},
			{File: "improved.md", Type: "agent", Success: true, Improvements: []textutil.ImprovementRecommendation{{Description: "Add examples", PointValue: 3}}},
		},
	}
	groups := groupIssuesByResult(BuildFlatIssues(summary), len(summary.Results))
	if len(groups) != 4 || len(groups[0]) != 1 || len(groups[1]) != 0 || len(groups[2]) != 2 || len(groups[3]) != 0 ||
		groups[0][0].Severity != SeverityError || groups[2][0].Severity != SeverityWarning || groups[2][1].Severity != SeveritySuggestion {
		t.Fatalf("result-index groups differ from flat issue ordering: %+v", groups)
	}

	var console bytes.Buffer
	cf := NewConsoleFormatter(false, true, false, false)
	cf.colorize = false
	cf.writer = &console
	if err := cf.Format(summary); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(console.String(), "same.md\n    ✘ same.md: first error") ||
		!strings.Contains(console.String(), "same.md\n    ⚠ same.md: second warning\n    💡 same.md: second suggestion") ||
		strings.Count(console.String(), "first error") != 1 || strings.Count(console.String(), "second warning") != 1 {
		t.Fatalf("console issue order or result association changed: %s", console.String())
	}

	path := filepath.Join(t.TempDir(), "report.md")
	mf := NewMarkdownFormatter(false, false, path)
	mf.showImprovements = true
	if err := mf.Format(summary); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	markdown := string(data)
	if strings.Contains(markdown, "empty.md") || strings.Count(markdown, "- [same.md](#samemd)") != 2 ||
		strings.Count(markdown, "- **same.md** - first error") != 1 || strings.Count(markdown, "- **same.md** - second warning") != 1 ||
		!strings.Contains(markdown, "- [improved.md](#improvedmd)") || !strings.Contains(markdown, "### improved.md\n\nStatus: ✅\n\nType: `agent`\n\n#### Improvements\n\n- **+3 pts:** Add examples") {
		t.Fatalf("markdown result association or improvement-only rendering changed: %s", markdown)
	}
}
