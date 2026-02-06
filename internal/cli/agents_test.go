package cli

import (
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

func TestValidateAgentSpecific(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]interface{}
		filePath      string
		contents      string
		wantErrCount  int
		wantSuggCount int
	}{
		{
			name: "valid agent",
			data: map[string]interface{}{
				"name":        "test-agent",
				"description": "A test agent. Use PROACTIVELY when testing.",
				"model":       "sonnet",
			},
			filePath:      "agents/test-agent.md",
			contents:      "---\nname: test-agent\ndescription: A test agent. Use PROACTIVELY when testing.\nmodel: sonnet\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name:          "missing name",
			data:          map[string]interface{}{"description": "test. Use PROACTIVELY when testing."},
			filePath:      "agents/test.md",
			contents:      "---\ndescription: test. Use PROACTIVELY when testing.\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name:          "missing description",
			data:          map[string]interface{}{"name": "test"},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "invalid name format",
			data: map[string]interface{}{
				"name":        "TestAgent",
				"description": "test. Use PROACTIVELY when testing.",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: TestAgent\ndescription: test. Use PROACTIVELY when testing.\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "reserved word",
			data: map[string]interface{}{
				"name":        "claude",
				"description": "test. Use PROACTIVELY when testing.",
			},
			filePath:      "agents/claude.md",
			contents:      "---\nname: claude\ndescription: test. Use PROACTIVELY when testing.\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "name doesn't match filename",
			data: map[string]interface{}{
				"name":        "test-agent",
				"description": "test. Use PROACTIVELY when testing.",
			},
			filePath:      "agents/other-name.md",
			contents:      "---\nname: test-agent\ndescription: test. Use PROACTIVELY when testing.\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "invalid color",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"color":       "rainbow",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\ncolor: rainbow\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "unknown field",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"foo":         "bar",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nfoo: bar\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "missing PROACTIVELY in description",
			data: map[string]interface{}{
				"name":        "test",
				"description": "A test agent for doing things.",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: A test agent for doing things.\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "valid memory scope",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"memory":      "project",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmemory: project\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "invalid memory scope",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"memory":      "global",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmemory: global\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateAgentSpecific(tt.data, tt.filePath, tt.contents)

			errCount := 0
			suggCount := 0
			for _, e := range errors {
				if e.Severity == "error" {
					errCount++
				} else if e.Severity == "suggestion" {
					suggCount++
				}
			}

			if errCount != tt.wantErrCount {
				t.Errorf("validateAgentSpecific() errors = %d, want %d", errCount, tt.wantErrCount)
				for _, e := range errors {
					if e.Severity == "error" {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
			if suggCount < tt.wantSuggCount {
				t.Errorf("validateAgentSpecific() suggestions = %d, want at least %d", suggCount, tt.wantSuggCount)
			}
		})
	}
}

func TestHasEditingTools(t *testing.T) {
	tests := []struct {
		name  string
		tools interface{}
		want  bool
	}{
		{"wildcard", "*", true},
		{"string with Edit", "Read, Edit, Bash", true},
		{"string with Write", "Write", true},
		{"string with MultiEdit", "MultiEdit", true},
		{"string without editing", "Read, Bash", false},
		{"array with Edit", []interface{}{"Read", "Edit"}, true},
		{"array without editing", []interface{}{"Read", "Bash"}, false},
		{"empty string", "", false},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasEditingTools(tt.tools)
			if got != tt.want {
				t.Errorf("hasEditingTools(%v) = %v, want %v", tt.tools, got, tt.want)
			}
		})
	}
}

func TestValidateAgentBestPractices(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		contents     string
		data         map[string]interface{}
		wantContains []string
	}{
		{
			name:     "XML tags in description",
			filePath: "agents/test.md",
			contents: "---\ndescription: Test <xml>tag</xml>\n---\n",
			data: map[string]interface{}{
				"description": "Test <xml>tag</xml>",
			},
			wantContains: []string{"XML-like tags"},
		},
		{
			name:     "bloat section Quick Reference",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test\n---\n## Quick Reference\n",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test",
			},
			wantContains: []string{"Quick Reference"},
		},
		{
			name:     "missing model",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test\n---\n",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test",
			},
			wantContains: []string{"model"},
		},
		{
			name:     "editing tools without permissionMode",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\ntools: Edit\n---\n",
			data: map[string]interface{}{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"tools":       "Edit",
			},
			wantContains: []string{"permissionMode"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validateAgentBestPractices(tt.filePath, tt.contents, tt.data)

			for _, want := range tt.wantContains {
				found := false
				for _, sugg := range suggestions {
					if strings.Contains(sugg.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateAgentBestPractices() should contain suggestion about %q", want)
					for _, s := range suggestions {
						t.Logf("  Got: %s", s.Message)
					}
				}
			}
		})
	}
}

func TestKnownAgentFields(t *testing.T) {
	expected := []string{"name", "description", "model", "color", "tools", "disallowedTools", "permissionMode", "skills", "hooks", "memory"}
	for _, field := range expected {
		if !knownAgentFields[field] {
			t.Errorf("knownAgentFields missing expected field: %s", field)
		}
	}
}

func TestAgentLinterPostProcessBatch(t *testing.T) {
	linter := NewAgentLinter()

	tests := []struct {
		name               string
		files              []discovery.File
		noCycleCheck       bool
		wantCycleErrors    int
		wantTotalErrors    int
		wantFailedFiles    int
		wantSuccessFiles   int
	}{
		{
			name: "no cycles detected",
			files: []discovery.File{
				{
					RelPath:  "agents/agent-a.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Skill: skill-a",
				},
				{
					RelPath:  "skills/skill-a/SKILL.md",
					Type:     discovery.FileTypeSkill,
					Contents: "Skill content",
				},
			},
			noCycleCheck:     false,
			wantCycleErrors:  0,
			wantTotalErrors:  0,
			wantFailedFiles:  0,
			wantSuccessFiles: 2,
		},
		{
			name: "cycle detected",
			files: []discovery.File{
				{
					RelPath:  "agents/agent-a.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Task(agent-b)",
				},
				{
					RelPath:  "agents/agent-b.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Task(agent-a)",
				},
			},
			noCycleCheck:     false,
			wantCycleErrors:  2, // One error per agent in cycle
			wantTotalErrors:  2,
			wantFailedFiles:  2,
			wantSuccessFiles: 0,
		},
		{
			name: "cycle check disabled",
			files: []discovery.File{
				{
					RelPath:  "agents/agent-a.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Task(agent-b)",
				},
				{
					RelPath:  "agents/agent-b.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Task(agent-a)",
				},
			},
			noCycleCheck:     true,
			wantCycleErrors:  0, // Cycle check disabled
			wantTotalErrors:  0,
			wantFailedFiles:  0,
			wantSuccessFiles: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			crossValidator := NewCrossFileValidator(tt.files)
			ctx := &LinterContext{
				CrossValidator: crossValidator,
				NoCycleCheck:   tt.noCycleCheck,
			}

			// Create initial summary with all files successful
			summary := &LintSummary{
				TotalFiles:      len(tt.files),
				SuccessfulFiles: len(tt.files),
				FailedFiles:     0,
				TotalErrors:     0,
				Results:         make([]LintResult, len(tt.files)),
			}

			for i, file := range tt.files {
				summary.Results[i] = LintResult{
					File:    file.RelPath,
					Type:    "agent",
					Success: true,
					Errors:  []cue.ValidationError{},
				}
			}

			// Run post-processing
			linter.PostProcessBatch(ctx, summary)

			// Count cycle errors
			cycleErrors := 0
			for _, result := range summary.Results {
				for _, err := range result.Errors {
					if strings.Contains(err.Message, "Circular dependency") {
						cycleErrors++
					}
				}
			}

			if cycleErrors != tt.wantCycleErrors {
				t.Errorf("PostProcessBatch() cycle errors = %d, want %d", cycleErrors, tt.wantCycleErrors)
			}

			if summary.TotalErrors != tt.wantTotalErrors {
				t.Errorf("PostProcessBatch() TotalErrors = %d, want %d", summary.TotalErrors, tt.wantTotalErrors)
			}

			if summary.FailedFiles != tt.wantFailedFiles {
				t.Errorf("PostProcessBatch() FailedFiles = %d, want %d", summary.FailedFiles, tt.wantFailedFiles)
			}

			if summary.SuccessfulFiles != tt.wantSuccessFiles {
				t.Errorf("PostProcessBatch() SuccessfulFiles = %d, want %d", summary.SuccessfulFiles, tt.wantSuccessFiles)
			}
		})
	}
}
