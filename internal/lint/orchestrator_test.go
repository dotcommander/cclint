package lint

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

type recordingComponentLinter struct {
	fileType discovery.FileType
	typeName string
	issues   []cue.ValidationError
	contexts *[]*LinterContext
}

func (l *recordingComponentLinter) Type() string                 { return l.typeName }
func (l *recordingComponentLinter) FileType() discovery.FileType { return l.fileType }
func (l *recordingComponentLinter) ParseContent(string) (map[string]any, string, error) {
	return map[string]any{}, "", nil
}
func (l *recordingComponentLinter) ValidateCUE(*cue.Validator, map[string]any) ([]cue.ValidationError, error) {
	return nil, nil
}
func (l *recordingComponentLinter) ValidateSpecific(_ map[string]any, filePath, _ string) []cue.ValidationError {
	issues := make([]cue.ValidationError, len(l.issues))
	copy(issues, l.issues)
	for i := range issues {
		if issues[i].File == "" {
			issues[i].File = filePath
		}
	}
	return issues
}
func (l *recordingComponentLinter) PostProcessBatch(ctx *LinterContext, _ *LintSummary) {
	if l.contexts != nil {
		*l.contexts = append(*l.contexts, ctx)
	}
}

func writeComponentFile(t *testing.T, root string, fileType discovery.FileType) {
	t.Helper()
	var rel string
	switch fileType {
	case discovery.FileTypeAgent:
		rel = ".claude/agents/test.md"
	case discovery.FileTypeCommand:
		rel = ".claude/commands/test.md"
	default:
		t.Fatalf("unsupported fixture type %s", fileType)
	}
	path := filepath.Join(root, rel)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("fixture"), 0o644); err != nil {
		t.Fatal(err)
	}
}

func fakeEntry(fileType discovery.FileType, name string, issues []cue.ValidationError, contexts *[]*LinterContext, roots *[]string) LinterEntry {
	return LinterEntry{
		FileType: fileType,
		Name:     name,
		New: func(rootPath string) ComponentLinter {
			if roots != nil {
				*roots = append(*roots, rootPath)
			}
			return &recordingComponentLinter{fileType: fileType, typeName: name, issues: issues, contexts: contexts}
		},
	}
}

func TestRegistryOrderAndDefaults(t *testing.T) {
	wantRegistry := []struct {
		fileType  discovery.FileType
		name      string
		isDefault bool
	}{
		{discovery.FileTypeAgent, "agents", true},
		{discovery.FileTypeCommand, "commands", true},
		{discovery.FileTypeSkill, "skills", true},
		{discovery.FileTypeSettings, "settings", true},
		{discovery.FileTypeContext, "context", false},
		{discovery.FileTypeRule, "rules", true},
		{discovery.FileTypeOutputStyle, "output-styles", true},
		{discovery.FileTypePlugin, "plugins", true},
	}
	if len(linterRegistry) != len(wantRegistry) {
		t.Fatalf("registry length = %d, want %d", len(linterRegistry), len(wantRegistry))
	}
	for i, want := range wantRegistry {
		got := linterRegistry[i]
		if got.FileType != want.fileType || got.Name != want.name || got.Default != want.isDefault || got.New == nil {
			t.Errorf("registry[%d] = {%s %q %v}, want {%s %q %v}", i, got.FileType, got.Name, got.Default, want.fileType, want.name, want.isDefault)
		}
		lookedUp, ok := LinterForType(want.fileType)
		if !ok || lookedUp.Name != want.name {
			t.Errorf("LinterForType(%s) = {%q, %v}, want %q", want.fileType, lookedUp.Name, ok, want.name)
		}
	}
	if _, ok := LinterForType(discovery.FileTypeUnknown); ok {
		t.Error("LinterForType(unknown) unexpectedly succeeded")
	}

	wantDefaults := []string{"agents", "commands", "skills", "settings", "rules", "output-styles", "plugins"}
	defaults := DefaultLinters()
	if len(defaults) != len(wantDefaults) {
		t.Fatalf("default length = %d, want %d", len(defaults), len(wantDefaults))
	}
	for i, want := range wantDefaults {
		if defaults[i].Name != want {
			t.Errorf("defaults[%d].Name = %q, want %q", i, defaults[i].Name, want)
		}
	}
}

func TestRegistryPluginFactoryReceivesRoot(t *testing.T) {
	entry, ok := LinterForType(discovery.FileTypePlugin)
	if !ok {
		t.Fatal("plugin registry entry missing")
	}
	linter, ok := entry.New("/explicit/root").(*PluginLinter)
	if !ok {
		t.Fatalf("plugin factory returned %T", entry.New("/explicit/root"))
	}
	if linter.RootPath != "/explicit/root" {
		t.Errorf("plugin root = %q, want /explicit/root", linter.RootPath)
	}
}

func TestNewOrchestratorAndWithLinters(t *testing.T) {
	cfg := &config.Config{Root: t.TempDir(), Quiet: true}
	orch := NewOrchestrator(cfg, OrchestratorConfig{BaselinePath: ".cclintbaseline.json"})
	if orch.cfg != cfg || len(orch.linters) != len(DefaultLinters()) {
		t.Fatal("NewOrchestrator did not retain config and defaults")
	}
	custom := []LinterEntry{fakeEntry(discovery.FileTypeAgent, "custom", nil, nil, nil)}
	if got := orch.WithLinters(custom); got != orch || len(orch.linters) != 1 || orch.linters[0].Name != "custom" {
		t.Fatal("WithLinters did not replace linters and return the orchestrator")
	}
}

func TestRunSharesContextAndPropagatesResolvedRoot(t *testing.T) {
	root := t.TempDir()
	writeComponentFile(t, root, discovery.FileTypeAgent)
	writeComponentFile(t, root, discovery.FileTypeCommand)

	var contexts []*LinterContext
	var roots []string
	linters := []LinterEntry{
		fakeEntry(discovery.FileTypeAgent, "agents", nil, &contexts, &roots),
		fakeEntry(discovery.FileTypeCommand, "commands", nil, &contexts, &roots),
	}
	orch := NewOrchestrator(&config.Config{Root: root, Quiet: true}, OrchestratorConfig{BaselinePath: ".cclintbaseline.json"}).WithLinters(linters)
	result, err := orch.Run()
	if err != nil {
		t.Fatalf("Run() error = %v", err)
	}
	if result.TotalFiles != 2 || len(result.Summaries) != 2 {
		t.Fatalf("Run totals = %d files, %d summaries; want 2, 2", result.TotalFiles, len(result.Summaries))
	}
	if len(contexts) != 2 || contexts[0] != contexts[1] {
		t.Fatalf("batch contexts = %#v, want the same context twice", contexts)
	}
	for _, got := range roots {
		if got != root {
			t.Errorf("factory root = %q, want %q", got, root)
		}
	}
}

func TestRunSelectedSubset(t *testing.T) {
	root := t.TempDir()
	writeComponentFile(t, root, discovery.FileTypeAgent)
	writeComponentFile(t, root, discovery.FileTypeCommand)
	var commandCalls int
	command := fakeEntry(discovery.FileTypeCommand, "commands", nil, nil, nil)
	command.New = func(string) ComponentLinter {
		commandCalls++
		return &recordingComponentLinter{fileType: discovery.FileTypeCommand, typeName: "commands"}
	}

	orch := NewOrchestrator(&config.Config{Root: root, Quiet: true}, OrchestratorConfig{}).WithLinters([]LinterEntry{command})
	result, err := orch.Run()
	if err != nil {
		t.Fatal(err)
	}
	if commandCalls != 1 || result.TotalFiles != 1 || len(result.Summaries) != 1 || result.Summaries[0].ComponentType != "commands" {
		t.Fatalf("subset calls command=%d, files=%d summaries=%d", commandCalls, result.TotalFiles, len(result.Summaries))
	}
}

func TestRunAggregatesAndSkipsEmptySummaries(t *testing.T) {
	root := t.TempDir()
	writeComponentFile(t, root, discovery.FileTypeAgent)
	issues := []cue.ValidationError{
		{Message: "error", Severity: cue.SeverityError},
		{Message: "warning", Severity: cue.SeverityWarning},
		{Message: "suggestion", Severity: cue.SeveritySuggestion},
	}
	linters := []LinterEntry{
		fakeEntry(discovery.FileTypeAgent, "agents", issues, nil, nil),
		fakeEntry(discovery.FileTypeCommand, "commands", nil, nil, nil),
	}
	result, err := NewOrchestrator(&config.Config{Root: root, Quiet: true}, OrchestratorConfig{}).WithLinters(linters).Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.TotalFiles != 1 || result.TotalErrors != 1 || result.TotalWarnings != 1 || result.TotalSuggestions != 1 || !result.HasErrors {
		t.Fatalf("unexpected totals: %+v", result)
	}
	if len(result.Summaries) != 1 || result.Summaries[0].ComponentType != "agents" {
		t.Fatalf("summaries = %+v, want only agents", result.Summaries)
	}
}

func TestRunEmptySelectionDoesNotInitializeContext(t *testing.T) {
	result, err := NewOrchestrator(&config.Config{Root: filepath.Join(t.TempDir(), "missing"), Quiet: true}, OrchestratorConfig{}).WithLinters(nil).Run()
	if err != nil {
		t.Fatalf("empty run error = %v", err)
	}
	if result.TotalFiles != 0 || len(result.Summaries) != 0 {
		t.Fatalf("empty run result = %+v", result)
	}
}

func TestResolveBaselinePath(t *testing.T) {
	root := t.TempDir()
	tests := []struct{ path, want string }{
		{filepath.Join(root, "absolute.json"), filepath.Join(root, "absolute.json")},
		{"relative.json", filepath.Join(root, "relative.json")},
	}
	for _, tt := range tests {
		orch := NewOrchestrator(&config.Config{Root: root}, OrchestratorConfig{BaselinePath: tt.path})
		if got := orch.resolveBaselinePath(); got != tt.want {
			t.Errorf("resolveBaselinePath(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestLoadAndSaveBaseline(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "baseline.json")
	orch := NewOrchestrator(&config.Config{Root: root, Quiet: true}, OrchestratorConfig{UseBaseline: true, BaselinePath: path})
	if got, err := orch.loadBaseline(path); err != nil || got != nil {
		t.Fatalf("missing baseline = (%v, %v), want (nil, nil)", got, err)
	}
	issues := []cue.ValidationError{{File: ".claude/agents/test.md", Message: "known", Severity: cue.SeverityError, Source: "test"}}
	if err := orch.saveBaseline(issues, path); err != nil {
		t.Fatal(err)
	}
	got, err := orch.loadBaseline(path)
	if err != nil || got == nil || len(got.Fingerprints) != 1 {
		t.Fatalf("loaded baseline = (%v, %v)", got, err)
	}
}

func TestRunCreatesBaseline(t *testing.T) {
	root := t.TempDir()
	writeComponentFile(t, root, discovery.FileTypeAgent)
	path := filepath.Join(root, "baseline.json")
	issues := []cue.ValidationError{{Message: "new", Severity: cue.SeverityError, Source: "test"}}
	orch := NewOrchestrator(&config.Config{Root: root, Quiet: true}, OrchestratorConfig{CreateBaseline: true, BaselinePath: path}).WithLinters([]LinterEntry{
		fakeEntry(discovery.FileTypeAgent, "agents", issues, nil, nil),
	})
	result, err := orch.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.HasErrors {
		t.Error("baseline creation must clear HasErrors")
	}
	loaded, err := baseline.LoadBaseline(path)
	if err != nil || len(loaded.Fingerprints) != 1 {
		t.Fatalf("created baseline = (%v, %v)", loaded, err)
	}
}

func TestRunFiltersBaseline(t *testing.T) {
	root := t.TempDir()
	writeComponentFile(t, root, discovery.FileTypeAgent)
	path := filepath.Join(root, "baseline.json")
	known := cue.ValidationError{File: ".claude/agents/test.md", Message: "known", Severity: cue.SeverityError, Source: "test"}
	b := baseline.CreateBaseline([]cue.ValidationError{known})
	b.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := b.SaveBaseline(path); err != nil {
		t.Fatal(err)
	}
	issues := []cue.ValidationError{
		{Message: "known", Severity: cue.SeverityError, Source: "test"},
		{Message: "new", Severity: cue.SeverityError, Source: "test"},
	}
	orch := NewOrchestrator(&config.Config{Root: root, Quiet: true}, OrchestratorConfig{UseBaseline: true, BaselinePath: path}).WithLinters([]LinterEntry{
		fakeEntry(discovery.FileTypeAgent, "agents", issues, nil, nil),
	})
	result, err := orch.Run()
	if err != nil {
		t.Fatal(err)
	}
	if result.BaselineIgnored != 1 || result.TotalErrors != 1 || !result.HasErrors {
		t.Fatalf("filtered result = %+v", result)
	}
}
