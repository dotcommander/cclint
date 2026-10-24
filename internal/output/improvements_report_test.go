package output

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/scoring"
	"github.com/dotcommander/cclint/internal/textutil"
)

func TestImprovementReports(t *testing.T) {
	t.Parallel()
	for _, format := range []string{"json", "markdown"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()
			for _, show := range []bool{false, true} {
				name := "disabled"
				if show {
					name = "enabled"
				}
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					recommendations := []textutil.ImprovementRecommendation{
						{Description: "Add an example", PointValue: 4, Line: 7, Severity: "medium"},
						{Description: "Add a reference", PointValue: 2},
					}
					before := slices.Clone(recommendations)
					summary := &lint.LintSummary{
						TotalFiles: 2, SuccessfulFiles: 2, TotalWarnings: 1,
						Results: []lint.LintResult{
							{
								File: "agents/improved-agent.md", Type: "agent", Success: true,
								Quality:      &scoring.QualityScore{Overall: 80, Tier: "B"},
								Improvements: recommendations,
							},
							{
								File: "skills/SKILL.md", Type: "skill", Success: true,
								Warnings: []cue.ValidationError{{
									File: "skills/SKILL.md", Message: "Existing warning",
									Severity: cue.SeverityWarning, Source: cue.SourceCClintObserve,
								}},
							},
						},
					}
					cfg := &config.Config{
						Format: format, Output: filepath.Join(t.TempDir(), "report"),
						Root: "project", Version: "test-version", ShowImprovements: show,
					}
					if err := FormatAll(cfg, []*lint.LintSummary{summary}, time.Now()); err != nil {
						t.Fatal(err)
					}
					data, err := os.ReadFile(cfg.Output)
					if err != nil {
						t.Fatal(err)
					}
					if !slices.Equal(summary.Results[0].Improvements, before) {
						t.Fatal("report mutated source recommendations")
					}
					if format == "json" {
						checkJSONImprovements(t, data, show, before)
					} else {
						checkMarkdownImprovements(t, string(data), show, before)
					}
				})
			}
		})
	}
}

func checkJSONImprovements(t *testing.T, data []byte, show bool, want []textutil.ImprovementRecommendation) {
	t.Helper()
	var report JSONReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if report.Header.Tool != "cclint" || report.Header.Version != "test-version" || report.Summary.TotalFiles != 2 || report.Summary.TotalWarnings != 1 || len(report.Results) != 2 {
		t.Fatalf("existing envelope changed: %+v", report)
	}
	first := report.Results[0]
	if first.File != "agents/improved-agent.md" || first.Type != "agent" || !first.Success || first.Quality == nil || first.Quality.Overall != 80 || len(report.Results[1].Warnings) != 1 {
		t.Fatalf("existing result fields changed: %+v", report.Results)
	}
	var wire struct {
		Results []map[string]json.RawMessage `json:"results"`
	}
	if err := json.Unmarshal(data, &wire); err != nil {
		t.Fatal(err)
	}
	if _, present := wire.Results[1]["improvements"]; present {
		t.Fatal("empty recommendations were not omitted")
	}
	payload, present := wire.Results[0]["improvements"]
	if present != show {
		t.Fatalf("recommendations present=%v; requested=%v", present, show)
	}
	if !show {
		return
	}
	var got []map[string]any
	if err := json.Unmarshal(payload, &got); err != nil {
		t.Fatal(err)
	}
	if len(got) != len(want) {
		t.Fatalf("recommendations=%+v; want %d", got, len(want))
	}
	for i, recommendation := range want {
		if got[i]["description"] != recommendation.Description || got[i]["point_value"] != float64(recommendation.PointValue) {
			t.Errorf("recommendation %d data/order changed: %+v", i, got[i])
		}
		line, hasLine := got[i]["line"]
		severity, hasSeverity := got[i]["severity"]
		if hasLine != (recommendation.Line != 0) || (hasLine && line != float64(recommendation.Line)) || hasSeverity != (recommendation.Severity != "") || (hasSeverity && severity != recommendation.Severity) {
			t.Errorf("recommendation %d optional fields: %+v", i, got[i])
		}
	}
}

func checkMarkdownImprovements(t *testing.T, content string, show bool, want []textutil.ImprovementRecommendation) {
	t.Helper()
	for _, fragment := range []string{
		"### agents/improved-agent.md\n", "[agents/improved-agent.md](",
		"#### Improvements", "+4 pts", "+2 pts", "(line 7)", "[medium]",
		want[0].Description, want[1].Description,
	} {
		if strings.Contains(content, fragment) != show {
			t.Errorf("fragment %q presence differs from requested=%v", fragment, show)
		}
	}
	if strings.Contains(content, "(line 0)") || !strings.Contains(content, "Existing warning") || !strings.Contains(content, "| Warnings | 1 |") {
		t.Errorf("optional line or existing warning output changed: %s", content)
	}
	if show && strings.Index(content, want[0].Description) > strings.Index(content, want[1].Description) {
		t.Error("recommendation order changed")
	}
}
