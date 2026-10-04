package lint

import (
	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestAttachmentsPreserveSeverityAndCounts(t *testing.T) {
	s := &LintSummary{Results: []LintResult{{File: "skill/SKILL.md", Type: "skill"}}}
	warning := cue.ValidationError{File: "refs/map.md", Severity: cue.SeverityWarning, Message: "warning", Source: "test"}
	attachIssueToSummary(s, warning, attachAsError, true)
	attachIssueToSummary(s, cue.ValidationError{File: warning.File, Severity: cue.SeveritySuggestion, Message: "suggestion", Source: "test"}, attachAsError, true)
	recalculateTotals(s)
	if s.TotalFiles != 2 || s.TotalWarnings != 1 || s.TotalErrors != 0 || s.TotalSuggestions != 1 || s.SuccessfulFiles != 2 {
		t.Fatalf("wrong counters: %+v", s)
	}
	FilterResults(s, baseline.CreateBaseline([]cue.ValidationError{warning}))
	if s.TotalWarnings != 0 || !s.Results[1].Success {
		t.Fatalf("filter counters/status: %+v", s)
	}
}
func TestStringListValidationParity(t *testing.T) {
	stringIssues := checkCommandToolAllowlist(map[string]any{"allowed-tools": "Bash, Read"}, "command.md", "")
	listIssues := checkCommandToolAllowlist(map[string]any{"allowed-tools": []any{"Bash", "Read"}}, "command.md", "")
	if len(stringIssues) != 2 || len(listIssues) != len(stringIssues) {
		t.Fatalf("declaration parity: %v / %v", stringIssues, listIssues)
	}
	if issues := validatePathsGlob([]any{"src/**/*.{go,md}", "tests/**"}, "rule.md", ""); len(issues) != 0 {
		t.Fatal(issues)
	}
	if issues := validatePathsGlob([]any{"src/**", 42}, "rule.md", ""); len(issues) != 1 {
		t.Fatal("malformed rule list accepted")
	}
}
func TestNativeBinaryImportsOutsideFences(t *testing.T) {
	issues := checkBinaryIncludes("@images/a.png\n```\n@images/example.png\n```\n", "CLAUDE.md")
	if len(issues) != 1 || issues[0].Severity != cue.SeverityWarning {
		t.Fatalf("binary imports: %v", issues)
	}
}

func TestQuietMemoryChecksParticipateInBaseline(t *testing.T) {
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "CLAUDE.local.md"), []byte("local memory"), 0600); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{Root: root, Quiet: true}
	opts := OrchestratorConfig{CreateBaseline: true, BaselinePath: "baseline.json"}
	created, err := NewOrchestrator(cfg, opts).WithLinters([]LinterEntry{}).Run()
	if err != nil {
		t.Fatal(err)
	}
	if created.TotalWarnings == 0 || created.TotalFiles != 0 {
		t.Fatalf("memory must be diagnostic-only: %+v", created)
	}
	filtered, err := NewOrchestrator(cfg, OrchestratorConfig{UseBaseline: true, BaselinePath: "baseline.json"}).WithLinters([]LinterEntry{}).Run()
	if err != nil {
		t.Fatal(err)
	}
	if filtered.TotalWarnings != 0 || filtered.BaselineIgnored == 0 {
		t.Fatalf("memory baseline missing: %+v", filtered)
	}
}
func TestCUESchemaDiagnosticsUseActualFiles(t *testing.T) {
	validator := cue.NewValidator()
	if err := validator.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"agents/a.md", "agents/b.md"} {
		result := LintResult{}
		runCUEValidation(&result, file, NewAgentLinter(), validator, map[string]any{"name": "BAD", "description": "test", "model": "impossible"})
		found := false
		for _, issue := range result.Errors {
			if issue.File != file {
				t.Fatalf("wrong file: %+v", issue)
			}
			if issue.Source == cue.SourceAnthropicDocs {
				found = true
				if issue.Line != 0 || issue.Column != 0 {
					t.Fatalf("schema coordinates: %+v", issue)
				}
			}
		}
		if !found {
			t.Fatal("schema issue missing")
		}
	}
}

func TestSkillCyclesDeliveredAndCanBeDisabled(t *testing.T) {
	files := []discovery.File{{RelPath: "skills/alpha/SKILL.md", Type: discovery.FileTypeSkill, Contents: "Skill(beta)"}, {RelPath: "skills/beta/SKILL.md", Type: discovery.FileTypeSkill, Contents: "Skill(alpha)"}}
	for _, disabled := range []bool{false, true} {
		ctx := &LinterContext{RootPath: t.TempDir(), CrossValidator: crossfile.NewCrossFileValidator(files), NoCycleCheck: disabled}
		summary := &LintSummary{Results: []LintResult{{File: files[0].RelPath, Type: "skill"}, {File: files[1].RelPath, Type: "skill"}}}
		NewSkillLinter().PostProcessBatch(ctx, summary)
		recalculateTotals(summary)
		for _, result := range summary.Results {
			cycles := 0
			for _, issue := range result.Errors {
				if strings.Contains(issue.Message, "Circular dependency") {
					cycles++
				}
			}
			if disabled && cycles != 0 || !disabled && cycles != 1 {
				t.Fatalf("disabled=%v file=%s cycles=%d", disabled, result.File, cycles)
			}
		}
	}
}
func TestScriptAttributionAndPluginRoot(t *testing.T) {
	root := t.TempDir()
	scriptDir := filepath.Join(root, "skills", "alpha", "scripts")
	if err := os.MkdirAll(scriptDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(scriptDir, "check.py"), []byte("print('hello')"), 0600); err != nil {
		t.Fatal(err)
	}
	issues := NewSkillLinter(root).ValidateBestPractices("skills/alpha/SKILL.md", "body", map[string]any{})
	scripts := 0
	for _, issue := range issues {
		if strings.Contains(issue.Message, "Script '") {
			scripts++
			if issue.File != "skills/alpha/scripts/check.py" || issue.Line != 0 {
				t.Fatalf("script attribution: %+v", issue)
			}
		}
	}
	if scripts != 2 {
		t.Fatalf("script issues=%d", scripts)
	}
	if err := os.MkdirAll(filepath.Join(root, "plugin", "commands"), 0700); err != nil {
		t.Fatal(err)
	}
	if issues := validatePluginPathsExist(map[string]any{"commands": "./commands"}, root, "plugin/.claude-plugin/plugin.json", ""); len(issues) != 0 {
		t.Fatal(issues)
	}
}
func TestImportCyclesDeterministic(t *testing.T) {
	graph := NewImportGraph()
	graph.AddFile("/project/z.md", []string{"a.md", "b.md"})
	graph.AddFile("/project/a.md", []string{"z.md"})
	graph.AddFile("/project/b.md", []string{"z.md"})
	want := graph.DetectCycles()
	for i := 0; i < 10; i++ {
		if got := graph.DetectCycles(); !reflect.DeepEqual(got, want) {
			t.Fatalf("unstable cycles: %v / %v", got, want)
		}
	}
}
