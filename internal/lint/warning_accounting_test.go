package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

func warningBaselineFixture() (*LintSummary, []cue.ValidationError) {
	const file = ".claude/agents/test.md"
	known := []cue.ValidationError{
		{File: file, Message: "known error", Source: "test", Severity: cue.SeverityError},
		{File: file, Message: "known warning", Source: "test", Severity: cue.SeverityWarning},
		{File: file, Message: "known suggestion", Source: "test", Severity: cue.SeveritySuggestion},
	}
	return &LintSummary{ComponentType: "agent", Results: []LintResult{{
		File: file, Type: "agent",
		Errors: []cue.ValidationError{known[0]},
		Warnings: []cue.ValidationError{known[1], {
			File: file, Message: "fresh warning", Source: "test", Severity: cue.SeverityWarning,
		}},
		Suggestions: []cue.ValidationError{known[2]},
	}}}, known
}

func TestBaselineWarningConservation(t *testing.T) {
	t.Parallel()
	summary, known := warningBaselineFixture()
	ignored := FilterResults(summary, baseline.CreateBaseline(known))
	want := IgnoredCounts{Total: 3, Errors: 1, Warnings: 1, Suggestions: 1}
	if ignored != want {
		t.Fatalf("ignored = %+v, want %+v", ignored, want)
	}
	if summary.TotalErrors != 0 || summary.TotalWarnings != 1 || summary.TotalSuggestions != 0 ||
		summary.SuccessfulFiles != 1 || !summary.Results[0].Success {
		t.Fatalf("remaining counters/status: %+v", summary)
	}
	if again := FilterResults(summary, baseline.CreateBaseline(known)); again != (IgnoredCounts{}) {
		t.Fatalf("repeated filtering counted already removed issues: %+v", again)
	}
}

func TestBaselineWarningProducers(t *testing.T) {
	t.Parallel()
	for _, selected := range []bool{false, true} {
		name := "full"
		if selected {
			name = "selected"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			summary, known := warningBaselineFixture()
			path := filepath.Join(root, "baseline.json")
			if err := baseline.CreateBaseline(known).SaveBaseline(path); err != nil {
				t.Fatal(err)
			}
			o := NewOrchestrator(&config.Config{Root: root, Quiet: true},
				OrchestratorConfig{UseBaseline: true, BaselinePath: path})
			var result *Result
			var err error
			if selected {
				result, err = o.ProcessSelected(summary)
			} else {
				writeComponentFile(t, root, discovery.FileTypeAgent)
				o.WithLinters([]LinterEntry{fakeEntry(discovery.FileTypeAgent, "agents",
					CollectAllIssues(summary), nil, nil)})
				result, err = o.Run()
			}
			if err != nil {
				t.Fatal(err)
			}
			if result.BaselineIgnored != 3 || result.ErrorsIgnored != 1 ||
				result.WarningsIgnored != 1 || result.SuggestionsIgnored != 1 ||
				result.BaselineIgnored != result.ErrorsIgnored+result.WarningsIgnored+result.SuggestionsIgnored {
				t.Fatalf("lost ignored severity: %+v", result)
			}
			if result.HasErrors || result.TotalErrors != 0 || result.TotalWarnings != 1 ||
				result.TotalSuggestions != 0 || result.TotalFiles != 1 {
				t.Fatalf("remaining totals/status: %+v", result)
			}
		})
	}
}

func TestBaselineWarningMemory(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.local.md"), []byte("local memory"), 0600); err != nil {
		t.Fatal(err)
	}
	issues := CheckClaudeLocalGitignore(root)
	if len(issues) != 1 || issues[0].Severity != cue.SeverityWarning {
		t.Fatalf("memory warning fixture: %v", issues)
	}
	path := filepath.Join(root, "baseline.json")
	if err := baseline.CreateBaseline(issues).SaveBaseline(path); err != nil {
		t.Fatal(err)
	}
	result, err := NewOrchestrator(&config.Config{Root: root, Quiet: true},
		OrchestratorConfig{UseBaseline: true, BaselinePath: path}).WithLinters([]LinterEntry{}).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.BaselineIgnored != 1 || result.WarningsIgnored != 1 ||
		result.ErrorsIgnored != 0 || result.SuggestionsIgnored != 0 ||
		result.TotalWarnings != 0 || result.TotalFiles != 0 || result.HasErrors {
		t.Fatalf("memory warning accounting: %+v", result)
	}
}
