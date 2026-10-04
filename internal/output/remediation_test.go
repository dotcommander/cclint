package output

import (
	"encoding/json"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/lint"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestMultiReportFormatsAndWriteFailures(t *testing.T) {
	s := &lint.LintSummary{ComponentType: "agent", TotalFiles: 1, SuccessfulFiles: 1, TotalWarnings: 1, Results: []lint.LintResult{{File: "agents/a.md", Type: "agent", Success: true, Warnings: []cue.ValidationError{{File: "agents/a.md", Message: "warning", Severity: cue.SeverityWarning}}}}}
	for _, format := range []string{"json", "markdown", "console"} {
		t.Run(format, func(t *testing.T) {
			cfg := &config.Config{Format: format, Output: filepath.Join(t.TempDir(), "report"), Root: "project"}
			if err := FormatAll(cfg, []*lint.LintSummary{s}, time.Now()); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(cfg.Output)
			if err != nil || len(data) == 0 {
				t.Fatalf("report: %v", err)
			}
			if format == "json" {
				var report JSONReport
				if err = json.Unmarshal(data, &report); err != nil {
					t.Fatal(err)
				}
				if len(report.Results) != 1 || report.Results[0].Type != "agent" || report.Summary.TotalWarnings != 1 {
					t.Fatal(report)
				}
			}
			cfg.Output = t.TempDir()
			if err := FormatAll(cfg, []*lint.LintSummary{s}, time.Now()); err == nil {
				t.Fatal("write error swallowed")
			}
		})
	}
}
func TestCleanFileVisibleWithRequestedDetails(t *testing.T) {
	result := &lint.LintResult{Success: true}
	if !NewConsoleFormatter(false, false, true, false).shouldShowFile(result, nil) {
		t.Fatal("scores hidden")
	}
	if !NewConsoleFormatter(false, false, false, true).shouldShowFile(result, nil) {
		t.Fatal("improvements hidden")
	}
}
