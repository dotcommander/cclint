package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/textutil"
)

// Integration tests for linting workflows

func TestLintAgentsIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal project structure
	claudeDir := filepath.Join(tmpDir, ".claude")
	agentsDir := filepath.Join(claudeDir, "agents")
	if err := os.MkdirAll(agentsDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create a valid agent
	agentFile := filepath.Join(agentsDir, "test-agent.md")
	agentContent := `---
name: test-agent
description: A test agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test foundation

## Workflow

1. Do work
`
	if err := os.WriteFile(agentFile, []byte(agentContent), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := LintAgents(tmpDir, true, false, false)
	if err != nil {
		t.Fatalf("LintAgents() error = %v", err)
	}

	if summary.TotalFiles != 1 {
		t.Errorf("LintAgents() TotalFiles = %d, want 1", summary.TotalFiles)
	}
}

func TestLintCommandsIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	claudeDir := filepath.Join(tmpDir, ".claude")
	commandsDir := filepath.Join(claudeDir, "commands")
	if err := os.MkdirAll(commandsDir, 0755); err != nil {
		t.Fatal(err)
	}

	cmdFile := filepath.Join(commandsDir, "test.md")
	cmdContent := `---
allowed-tools: Task
---
Task(test-specialist): do something
`
	if err := os.WriteFile(cmdFile, []byte(cmdContent), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := LintCommands(tmpDir, true, false, false)
	if err != nil {
		t.Fatalf("LintCommands() error = %v", err)
	}

	if summary.TotalFiles != 1 {
		t.Errorf("LintCommands() TotalFiles = %d, want 1", summary.TotalFiles)
	}
}

func TestLintSkillsIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	claudeDir := filepath.Join(tmpDir, ".claude")
	skillsDir := filepath.Join(claudeDir, "skills/test-skill")
	if err := os.MkdirAll(skillsDir, 0755); err != nil {
		t.Fatal(err)
	}

	skillFile := filepath.Join(skillsDir, "SKILL.md")
	skillContent := `---
name: test-skill
description: A test skill that helps with testing. Use when running tests.
---

# Test Skill

Content here
`
	if err := os.WriteFile(skillFile, []byte(skillContent), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := LintSkills(tmpDir, true, false, false)
	if err != nil {
		t.Fatalf("LintSkills() error = %v", err)
	}

	if summary.TotalFiles != 1 {
		t.Errorf("LintSkills() TotalFiles = %d, want 1", summary.TotalFiles)
	}
}

func TestLintRulesIntegration(t *testing.T) {
	tmpDir := t.TempDir()

	claudeDir := filepath.Join(tmpDir, ".claude")
	rulesDir := filepath.Join(claudeDir, "rules")
	if err := os.MkdirAll(rulesDir, 0755); err != nil {
		t.Fatal(err)
	}

	ruleFile := filepath.Join(rulesDir, "test.md")
	ruleContent := `---
paths: "**/*.go"
---

# Test Rule

Content
`
	if err := os.WriteFile(ruleFile, []byte(ruleContent), 0644); err != nil {
		t.Fatal(err)
	}

	summary, err := LintRules(tmpDir, true, false, false)
	if err != nil {
		t.Fatalf("LintRules() error = %v", err)
	}

	if summary.TotalFiles != 1 {
		t.Errorf("LintRules() TotalFiles = %d, want 1", summary.TotalFiles)
	}
}

func TestSingleFileContextEnsure(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude/agents")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(claudeDir, "test.md")
	if err := os.WriteFile(testFile, []byte("test content"), 0644); err != nil {
		t.Fatal(err)
	}

	ctx, err := NewSingleFileLinterContext(testFile, tmpDir, "", false, false)
	if err != nil {
		t.Fatalf("NewSingleFileLinterContext() error = %v", err)
	}

	// First call should initialize
	validator1 := ctx.EnsureCrossFileValidator()

	// Second call should return cached
	validator2 := ctx.EnsureCrossFileValidator()

	if validator1 != validator2 {
		t.Error("EnsureCrossFileValidator() should return cached validator")
	}
}

func TestImprovementRecommendation(t *testing.T) {
	// Test that ImprovementRecommendation structure works
	rec := textutil.ImprovementRecommendation{
		Description: "Missing examples",
		PointValue:  5,
		Line:        10,
		Severity:    "medium",
	}

	if rec.PointValue != 5 {
		t.Errorf("ImprovementRecommendation.PointValue = %d, want 5", rec.PointValue)
	}
}

func TestLintResultStructure(t *testing.T) {
	result := LintResult{
		File:    "test.md",
		Type:    "agent",
		Success: true,
	}

	if result.File != "test.md" {
		t.Errorf("LintResult.File = %q, want %q", result.File, "test.md")
	}

	if result.Type != "agent" {
		t.Errorf("LintResult.Type = %q, want %q", result.Type, "agent")
	}
}

func TestLintSummaryAggregation(t *testing.T) {
	summary := LintSummary{
		TotalFiles:       5,
		SuccessfulFiles:  3,
		FailedFiles:      2,
		TotalErrors:      10,
		TotalWarnings:    5,
		TotalSuggestions: 8,
	}

	if summary.TotalFiles != 5 {
		t.Errorf("LintSummary.TotalFiles = %d, want 5", summary.TotalFiles)
	}

	if summary.SuccessfulFiles+summary.FailedFiles != summary.TotalFiles {
		t.Error("LintSummary: SuccessfulFiles + FailedFiles should equal TotalFiles")
	}
}

func TestFindProjectRootForFile(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude/agents")
	if err := os.MkdirAll(claudeDir, 0755); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(claudeDir, "test.md")

	root, err := findProjectRootForFile(testFile)
	if err != nil {
		t.Errorf("findProjectRootForFile() error = %v", err)
	}

	if root == "" {
		t.Error("findProjectRootForFile() returned empty root")
	}
}

func TestSingleFileLinterTypes(t *testing.T) {
	tmpDir := t.TempDir()
	claudeDir := filepath.Join(tmpDir, ".claude")

	tests := []struct {
		name     string
		dir      string
		filename string
		content  string
		fileType discovery.FileType
	}{
		{
			name:     "agent",
			dir:      "agents",
			filename: "test.md",
			content:  "---\nname: test\ndescription: test\n---\nBody",
			fileType: discovery.FileTypeAgent,
		},
		{
			name:     "command",
			dir:      "commands",
			filename: "cmd.md",
			content:  "---\nallowed-tools: Task\n---\nBody",
			fileType: discovery.FileTypeCommand,
		},
		{
			name:     "skill",
			dir:      "skills/test",
			filename: "SKILL.md",
			content:  "---\nname: test\n---\nBody",
			fileType: discovery.FileTypeSkill,
		},
		{
			name:     "settings",
			dir:      "",
			filename: "settings.json",
			content:  `{"theme": "dark"}`,
			fileType: discovery.FileTypeSettings,
		},
		{
			name:     "context",
			dir:      "",
			filename: "CLAUDE.md",
			content:  "# Context",
			fileType: discovery.FileTypeContext,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := filepath.Join(claudeDir, tt.dir)
			if err := os.MkdirAll(dir, 0755); err != nil {
				t.Fatal(err)
			}

			filePath := filepath.Join(dir, tt.filename)
			if err := os.WriteFile(filePath, []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}

			ctx, err := NewSingleFileLinterContext(filePath, tmpDir, "", false, false)
			if err != nil {
				t.Fatalf("NewSingleFileLinterContext() error = %v", err)
			}

			if ctx.File.Type != tt.fileType {
				t.Errorf("NewSingleFileLinterContext() FileType = %v, want %v", ctx.File.Type, tt.fileType)
			}
		})
	}
}
