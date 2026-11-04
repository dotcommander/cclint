package lint

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

func TestParallelBatchPreservesSequentialOrder(t *testing.T) {
	t.Parallel()
	root := selectionTestRoot(t)
	agentDir := filepath.Join(root, ".claude", "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	names := []string{"alpha", "bravo", "charlie", "delta", "echo"}
	for _, name := range names {
		contents := "---\nname: " + name + "\ndescription: Test agent " + name + "\n---\nBody.\n"
		if err := os.WriteFile(filepath.Join(agentDir, name+".md"), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	sequential, err := NewLinterContextWithOptions(root, LinterOptions{Quiet: true})
	if err != nil {
		t.Fatal(err)
	}
	parallel, err := NewLinterContextWithOptions(root, LinterOptions{Quiet: true, Parallel: true, Concurrency: 4})
	if err != nil {
		t.Fatal(err)
	}

	entry, ok := LinterForType(discovery.FileTypeAgent)
	if !ok {
		t.Fatal("agent linter missing from registry")
	}
	want := lintBatch(sequential, entry.New(sequential.RootPath))
	got := lintBatch(parallel, entry.New(parallel.RootPath))

	if len(got.Results) != len(want.Results) || len(got.Results) != len(names) {
		t.Fatalf("parallel results = %d, want %d (sequential %d)", len(got.Results), len(names), len(want.Results))
	}
	for i := range want.Results {
		if got.Results[i].File != want.Results[i].File || got.Results[i].Success != want.Results[i].Success {
			t.Errorf("results[%d] = (%s, %v), want (%s, %v)",
				i, got.Results[i].File, got.Results[i].Success, want.Results[i].File, want.Results[i].Success)
		}
		if len(got.Results[i].Errors) != len(want.Results[i].Errors) ||
			len(got.Results[i].Warnings) != len(want.Results[i].Warnings) ||
			len(got.Results[i].Suggestions) != len(want.Results[i].Suggestions) {
			t.Errorf("results[%d] issue counts differ between parallel and sequential", i)
		}
	}
	if got.TotalFiles != want.TotalFiles || got.TotalErrors != want.TotalErrors ||
		got.TotalWarnings != want.TotalWarnings || got.TotalSuggestions != want.TotalSuggestions {
		t.Errorf("totals differ: parallel (%d/%d/%d/%d) vs sequential (%d/%d/%d/%d)",
			got.TotalFiles, got.TotalErrors, got.TotalWarnings, got.TotalSuggestions,
			want.TotalFiles, want.TotalErrors, want.TotalWarnings, want.TotalSuggestions)
	}
}

func TestSkipSchemaValidationDropsCUEOnlyFindings(t *testing.T) {
	t.Parallel()
	// model: bogus fails the embedded #Model CUE union (error) while the Go
	// model check only warns — a CUE-only error fixture.
	root := selectionTestRoot(t)
	agentDir := filepath.Join(root, ".claude", "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	contents := "---\nname: bogus-model\ndescription: Fails CUE model union.\nmodel: bogus\n---\nBody.\n"
	if err := os.WriteFile(filepath.Join(agentDir, "bogus.md"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	withSchema, err := NewLinterContextWithOptions(root, LinterOptions{Quiet: true})
	if err != nil {
		t.Fatal(err)
	}
	entry, ok := LinterForType(discovery.FileTypeAgent)
	if !ok {
		t.Fatal("agent linter missing from registry")
	}
	summary := lintBatch(withSchema, entry.New(withSchema.RootPath))
	if summary.TotalErrors == 0 {
		t.Fatal("fixture produced no CUE errors without the skip flag; fixture is not CUE-only")
	}

	skipped, err := NewLinterContextWithOptions(root, LinterOptions{Quiet: true, SkipSchemaValidation: true})
	if err != nil {
		t.Fatal(err)
	}
	if skipped.Validator != nil {
		t.Error("SkipSchemaValidation left a validator on the context")
	}
	skippedSummary := lintBatch(skipped, entry.New(skipped.RootPath))
	if skippedSummary.TotalErrors != 0 {
		t.Errorf("SkipSchemaValidation kept %d errors, want 0", skippedSummary.TotalErrors)
		for _, r := range skippedSummary.Results {
			for _, e := range r.Errors {
				t.Logf("  error: %s", e.Message)
			}
		}
	}
}

func TestRelaxedRulesDemoteWarningsRunAllLinters(t *testing.T) {
	t.Parallel()
	root := selectionTestRoot(t)
	writeComponentFile(t, root, discovery.FileTypeAgent)

	issues := []cue.ValidationError{
		{File: ".claude/agents/test.md", Message: "an error", Severity: cue.SeverityError},
		{File: ".claude/agents/test.md", Message: "a warning", Severity: cue.SeverityWarning},
	}
	strict := NewOrchestrator(&config.Config{Rules: config.RulesConfig{Strict: true}, Root: root, Quiet: true},
		OrchestratorConfig{BaselinePath: ".cclintbaseline.json"}).
		WithLinters([]LinterEntry{fakeEntry(discovery.FileTypeAgent, "agents", issues, nil, nil)})
	strictResult, err := strict.Run()
	if err != nil {
		t.Fatal(err)
	}

	relaxed := NewOrchestrator(&config.Config{Root: root, Quiet: true, Rules: config.RulesConfig{Strict: false}},
		OrchestratorConfig{BaselinePath: ".cclintbaseline.json"}).
		WithLinters([]LinterEntry{fakeEntry(discovery.FileTypeAgent, "agents", issues, nil, nil)})
	relaxedResult, err := relaxed.Run()
	if err != nil {
		t.Fatal(err)
	}

	if strictResult.TotalWarnings != 1 || strictResult.TotalSuggestions != 0 {
		t.Fatalf("strict totals = (%d warnings, %d suggestions), want (1, 0)",
			strictResult.TotalWarnings, strictResult.TotalSuggestions)
	}
	if relaxedResult.TotalWarnings != 0 || relaxedResult.TotalSuggestions != 1 {
		t.Errorf("relaxed totals = (%d warnings, %d suggestions), want (0, 1)",
			relaxedResult.TotalWarnings, relaxedResult.TotalSuggestions)
	}
	if relaxedResult.TotalErrors != 1 {
		t.Errorf("relaxed TotalErrors = %d, want 1 (errors are untouched)", relaxedResult.TotalErrors)
	}
	demoted := relaxedResult.Summaries[0].Results[0].Warnings
	suggestions := relaxedResult.Summaries[0].Results[0].Suggestions
	if len(demoted) != 0 {
		t.Errorf("relaxed result kept %d warnings", len(demoted))
	}
	if len(suggestions) != 1 || suggestions[0].Message != "a warning" || suggestions[0].Severity != cue.SeveritySuggestion {
		t.Errorf("demoted suggestion = %+v, want warning demoted to suggestion", suggestions)
	}
}

func TestRelaxedRulesDemoteWarningsProcessSelected(t *testing.T) {
	t.Parallel()
	summary := &LintSummary{ComponentType: "agents"}
	summary.Results = []LintResult{{
		File:    "agents/test.md",
		Type:    "agents",
		Success: true,
		Errors: []cue.ValidationError{
			{File: "agents/test.md", Message: "an error", Severity: cue.SeverityError},
		},
		Warnings: []cue.ValidationError{
			{File: "agents/test.md", Message: "a warning", Severity: cue.SeverityWarning},
		},
	}}

	o := NewOrchestrator(&config.Config{Root: t.TempDir(), Quiet: true, Rules: config.RulesConfig{Strict: false}},
		OrchestratorConfig{BaselinePath: ".cclintbaseline.json"})
	result, err := o.ProcessSelected(summary)
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalWarnings != 0 || result.TotalSuggestions != 1 || result.TotalErrors != 1 {
		t.Errorf("ProcessSelected relaxed totals = (%d errors, %d warnings, %d suggestions), want (1, 0, 1)",
			result.TotalErrors, result.TotalWarnings, result.TotalSuggestions)
	}
	if len(summary.Results[0].Warnings) != 0 {
		t.Errorf("ProcessSelected left %d warnings on the result", len(summary.Results[0].Warnings))
	}
	demoted := summary.Results[0].Suggestions
	if len(demoted) != 1 || demoted[0].Message != "a warning" || demoted[0].Severity != cue.SeveritySuggestion {
		t.Errorf("ProcessSelected demoted suggestion = %+v, want the warning demoted", demoted)
	}
}

func TestLinterOptionsSchemaExtensionsThreaded(t *testing.T) {
	t.Parallel()
	root := selectionTestRoot(t)
	agentDir := filepath.Join(root, ".claude", "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	extFile := filepath.Join(root, "agent-ext.cue")
	if err := os.WriteFile(extFile, []byte("#Agent: {maxTurns: int & >100}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	for name, maxTurns := range map[string]int{"violating": 50, "compliant": 200} {
		contents := "---\nname: " + name + "\ndescription: Extension fixture.\nmaxTurns: " + strconv.Itoa(maxTurns) + "\n---\nBody.\n"
		if err := os.WriteFile(filepath.Join(agentDir, name+".md"), []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	ctx, err := NewLinterContextWithOptions(root, LinterOptions{
		Quiet:            true,
		SchemaExtensions: map[string]any{"agent": "agent-ext.cue"},
	})
	if err != nil {
		t.Fatalf("NewLinterContextWithOptions with extensions: %v", err)
	}
	entry, ok := LinterForType(discovery.FileTypeAgent)
	if !ok {
		t.Fatal("agent linter missing from registry")
	}
	summary := lintBatch(ctx, entry.New(ctx.RootPath))

	var violated, compliant bool
	for _, r := range summary.Results {
		hasMaxTurnsError := false
		for _, e := range r.Errors {
			if strings.Contains(e.Message, "maxTurns") {
				hasMaxTurnsError = true
			}
		}
		if filepath.Base(r.File) == "violating.md" {
			violated = true
			if !hasMaxTurnsError {
				t.Errorf("extension constraint not enforced on violating fixture: %+v", r.Errors)
			}
		}
		if filepath.Base(r.File) == "compliant.md" {
			compliant = true
			if hasMaxTurnsError {
				t.Errorf("compliant fixture rejected by extension: %+v", r.Errors)
			}
		}
	}
	if !violated || !compliant {
		t.Fatalf("expected both fixtures linted (violated=%v, compliant=%v)", violated, compliant)
	}
}

// Regression: a missing extension file on the LintFiles (selected-file) path
// must fail the run, not silently proceed without the configured constraints.
func TestLintFilesMissingSchemaExtensionFails(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	agentDir := filepath.Join(root, ".claude", "agents")
	if err := os.MkdirAll(agentDir, 0o755); err != nil {
		t.Fatal(err)
	}
	contents := "---\nname: ok\ndescription: Valid agent.\n---\nBody.\n"
	if err := os.WriteFile(filepath.Join(agentDir, "ok.md"), []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}

	summary, err := LintFiles([]string{filepath.Join(agentDir, "ok.md")}, root, "", true, false,
		false, map[string]any{"agent": "does-not-exist.cue"})
	if err != nil {
		t.Fatalf("LintFiles returned an invocation error; context-creation failures should be recorded results: %v", err)
	}
	if summary.TotalErrors == 0 {
		t.Fatal("missing extension file was silently ignored; want a recorded error")
	}
	found := false
	for _, r := range summary.Results {
		for _, e := range r.Errors {
			if strings.Contains(e.Message, "load schema extensions") && strings.Contains(e.Message, "does-not-exist.cue") {
				found = true
			}
		}
	}
	if !found {
		t.Errorf("no error names the extension failure; results: %+v", summary.Results)
	}
}
