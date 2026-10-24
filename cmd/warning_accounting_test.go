package cmd

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/lint"
)

func TestWarningSummaryAggregation(t *testing.T) {
	t.Parallel()
	summary := &ComponentSummary{TierCounts: make(map[string]int), TopIssues: make(map[string]int)}
	results := []lint.LintResult{
		{File: "warning.md", Warnings: []cue.ValidationError{
			{Message: "Missing Foundation section", Severity: cue.SeverityWarning},
			{Message: "Missing Workflow section", Severity: cue.SeverityWarning},
		}},
		{File: "mixed.md",
			Errors:      []cue.ValidationError{{Message: "Missing Foundation section"}},
			Suggestions: []cue.ValidationError{{Message: "Missing Foundation section"}}},
	}
	aggregateResults(summary, results)
	if summary.TopIssues[issueMissingFoundation] != 3 || summary.TopIssues[issueMissingWorkflow] != 1 ||
		len(summary.TopIssues) != 2 || len(summary.AllResults) != 2 {
		t.Fatalf("warning categories missing or duplicated: %+v", summary)
	}
	if len(results[0].Warnings) != 2 || len(results[1].Errors) != 1 || len(results[1].Suggestions) != 1 {
		t.Fatal("aggregation mutated diagnostic slices")
	}
}

// Child processes isolate the real stderr renderer from other parallel tests.
func TestBaselineWarningReport(t *testing.T) {
	t.Parallel()
	if mode := os.Getenv("CCLINT_TEST_BASELINE_WARNINGS"); mode != "" {
		root := t.TempDir()
		known := []cue.ValidationError{
			{File: "test.md", Message: "known error", Source: "test", Severity: cue.SeverityError},
			{File: "test.md", Message: "known warning", Source: "test", Severity: cue.SeverityWarning},
			{File: "test.md", Message: "known suggestion", Source: "test", Severity: cue.SeveritySuggestion},
		}
		path := filepath.Join(root, "baseline.json")
		if err := baseline.CreateBaseline(known).SaveBaseline(path); err != nil {
			t.Fatal(err)
		}
		cfg := &config.Config{Root: root, Quiet: mode == "quiet"}
		summary := &lint.LintSummary{Results: []lint.LintResult{{
			File: "test.md", Type: "agent", Errors: known[:1],
			Warnings: known[1:2], Suggestions: known[2:],
		}}}
		result, err := lint.NewOrchestrator(cfg, lint.OrchestratorConfig{
			UseBaseline: true, BaselinePath: path,
		}).ProcessSelected(summary)
		if err != nil {
			t.Fatal(err)
		}
		if err := reportLintOutcome(cfg, executionOptions{}, fullLintOutcome(result), func() error { return nil }); err != nil {
			t.Fatal(err)
		}
		return
	}
	if testing.Short() {
		t.Skip("subprocess stderr integration")
	}
	for _, mode := range []string{"normal", "quiet"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(t.Context(), 4*time.Second)
			t.Cleanup(cancel)
			child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestBaselineWarningReport$")
			child.Env = append(os.Environ(), "CCLINT_TEST_BASELINE_WARNINGS="+mode)
			data, err := child.CombinedOutput()
			if err != nil {
				t.Fatalf("report subprocess: %v\n%s", err, data)
			}
			want := "3 baseline issues ignored (1 errors, 1 warnings, 1 suggestions)"
			if mode == "normal" && !strings.Contains(string(data), want) {
				t.Fatalf("warning baseline summary missing: %s", data)
			}
			if mode == "quiet" && strings.Contains(string(data), "baseline issues ignored") {
				t.Fatalf("quiet baseline summary leaked: %s", data)
			}
		})
	}
}
