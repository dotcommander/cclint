package cli

import (
	"testing"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/discovery"
)

// Test all linter implementations for basic interface compliance

func TestAgentLinter(t *testing.T) {
	linter := NewAgentLinter()

	if linter.Type() != "agent" {
		t.Errorf("AgentLinter.Type() = %q, want %q", linter.Type(), "agent")
	}

	if linter.FileType() != discovery.FileTypeAgent {
		t.Errorf("AgentLinter.FileType() = %v, want %v", linter.FileType(), discovery.FileTypeAgent)
	}

	// Test ParseContent
	contents := "---\nname: test\ndescription: test agent\n---\nBody"
	data, body, err := linter.ParseContent(contents)
	if err != nil {
		t.Errorf("AgentLinter.ParseContent() error = %v", err)
	}
	if data == nil {
		t.Error("AgentLinter.ParseContent() data is nil")
	}
	if body == "" {
		t.Error("AgentLinter.ParseContent() body is empty")
	}

	// Test ValidateSpecific
	_ = linter.ValidateSpecific(map[string]any{"name": "test", "description": "test"}, "test.md", contents)
}

func TestCommandLinter(t *testing.T) {
	linter := NewCommandLinter()

	if linter.Type() != "command" {
		t.Errorf("CommandLinter.Type() = %q, want %q", linter.Type(), "command")
	}

	if linter.FileType() != discovery.FileTypeCommand {
		t.Errorf("CommandLinter.FileType() = %v, want %v", linter.FileType(), discovery.FileTypeCommand)
	}

	// Test ParseContent
	contents := "---\nname: test\n---\nBody"
	data, body, err := linter.ParseContent(contents)
	if err != nil {
		t.Errorf("CommandLinter.ParseContent() error = %v", err)
	}
	if data == nil {
		t.Error("CommandLinter.ParseContent() data is nil")
	}
	if body == "" {
		t.Error("CommandLinter.ParseContent() body is empty")
	}

	// Test ValidateSpecific
	_ = linter.ValidateSpecific(map[string]any{"name": "test"}, "test.md", contents)
}

func TestSkillLinter(t *testing.T) {
	linter := NewSkillLinter()

	if linter.Type() != "skill" {
		t.Errorf("SkillLinter.Type() = %q, want %q", linter.Type(), "skill")
	}

	if linter.FileType() != discovery.FileTypeSkill {
		t.Errorf("SkillLinter.FileType() = %v, want %v", linter.FileType(), discovery.FileTypeSkill)
	}

	// Test PreValidate
	errors := linter.PreValidate("skills/test/SKILL.md", "content")
	if len(errors) > 0 {
		for _, e := range errors {
			if e.Severity == "error" {
				t.Errorf("SkillLinter.PreValidate() unexpected error: %s", e.Message)
			}
		}
	}

	errors = linter.PreValidate("skills/test/skill.md", "content")
	if len(errors) == 0 {
		t.Error("SkillLinter.PreValidate() should error on non-SKILL.md filename")
	}

	// Test ParseContent
	contents := "---\nname: test\n---\nBody"
	data, body, err := linter.ParseContent(contents)
	if err != nil {
		t.Errorf("SkillLinter.ParseContent() error = %v", err)
	}
	if data == nil {
		t.Error("SkillLinter.ParseContent() data is nil")
	}
	if body == "" {
		t.Error("SkillLinter.ParseContent() body is empty")
	}

	// Test ValidateSpecific
	_ = linter.ValidateSpecific(map[string]any{"name": "test"}, "SKILL.md", contents)

	// Test ValidateBestPractices
	_ = linter.ValidateBestPractices("SKILL.md", contents, map[string]any{"name": "test"})

	// Test Score
	score := linter.Score(contents, map[string]any{"name": "test"}, "body")
	if score == nil {
		t.Error("SkillLinter.Score() returned nil")
	}

	// Test GetImprovements
	improvements := linter.GetImprovements(contents, map[string]any{"name": "test"})
	if improvements == nil {
		t.Error("SkillLinter.GetImprovements() returned nil")
	}
}

func TestSettingsLinter(t *testing.T) {
	linter := NewSettingsLinter()

	if linter.Type() != "settings" {
		t.Errorf("SettingsLinter.Type() = %q, want %q", linter.Type(), "settings")
	}

	if linter.FileType() != discovery.FileTypeSettings {
		t.Errorf("SettingsLinter.FileType() = %v, want %v", linter.FileType(), discovery.FileTypeSettings)
	}

	// Test ParseContent
	contents := `{"theme": "dark"}`
	data, body, err := linter.ParseContent(contents)
	if err != nil {
		t.Errorf("SettingsLinter.ParseContent() error = %v", err)
	}
	if data == nil {
		t.Error("SettingsLinter.ParseContent() data is nil")
	}
	if body != "" {
		t.Error("SettingsLinter.ParseContent() body should be empty for JSON")
	}

	// Test ValidateSpecific
	_ = linter.ValidateSpecific(data, "settings.json", contents)
}

func TestContextLinter(t *testing.T) {
	linter := NewContextLinter()

	if linter.Type() != "context" {
		t.Errorf("ContextLinter.Type() = %q, want %q", linter.Type(), "context")
	}

	if linter.FileType() != discovery.FileTypeContext {
		t.Errorf("ContextLinter.FileType() = %v, want %v", linter.FileType(), discovery.FileTypeContext)
	}

	// Test ParseContent
	contents := "# CLAUDE.md\nProject context"
	data, _, err := linter.ParseContent(contents)
	if err != nil {
		t.Errorf("ContextLinter.ParseContent() error = %v", err)
	}
	if data == nil {
		t.Error("ContextLinter.ParseContent() data is nil")
	}

	// Test ValidateSpecific
	_ = linter.ValidateSpecific(data, "CLAUDE.md", contents)
}

func TestPluginLinter(t *testing.T) {
	linter := NewPluginLinter("")

	if linter.Type() != "plugin" {
		t.Errorf("PluginLinter.Type() = %q, want %q", linter.Type(), "plugin")
	}

	if linter.FileType() != discovery.FileTypePlugin {
		t.Errorf("PluginLinter.FileType() = %v, want %v", linter.FileType(), discovery.FileTypePlugin)
	}

	// Test ParseContent
	contents := `{"name": "test-plugin", "version": "1.0.0"}`
	data, body, err := linter.ParseContent(contents)
	if err != nil {
		t.Errorf("PluginLinter.ParseContent() error = %v", err)
	}
	if data == nil {
		t.Error("PluginLinter.ParseContent() data is nil")
	}
	if body != "" {
		t.Error("PluginLinter.ParseContent() body should be empty for JSON")
	}

	// Test ValidateSpecific
	_ = linter.ValidateSpecific(data, "plugin.json", contents)
}

// Test batch post-processing
func TestBatchPostProcessor(t *testing.T) {
	// The AgentLinter implements BatchPostProcessor for cycle detection
	linter := NewAgentLinter()

	// Create a valid context with CrossValidator
	files := []discovery.File{
		{RelPath: "agents/test.md", Type: discovery.FileTypeAgent, Contents: "test"},
	}
	crossValidator := crossfile.NewCrossFileValidator(files)

	ctx := &LinterContext{
		RootPath:       "/test",
		NoCycleCheck:   false,
		CrossValidator: crossValidator,
	}

	summary := &LintSummary{
		Results: []LintResult{},
	}

	// Call PostProcessBatch
	if bpp, ok := any(linter).(BatchPostProcessor); ok {
		bpp.PostProcessBatch(ctx, summary)
		// Should not panic
	} else {
		t.Error("AgentLinter should implement BatchPostProcessor")
	}
}
