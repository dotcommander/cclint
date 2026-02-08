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
		data          map[string]any
		filePath      string
		contents      string
		wantErrCount  int
		wantSuggCount int
	}{
		{
			name: "valid agent",
			data: map[string]any{
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
			data:          map[string]any{"description": "test. Use PROACTIVELY when testing."},
			filePath:      "agents/test.md",
			contents:      "---\ndescription: test. Use PROACTIVELY when testing.\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name:          "missing description",
			data:          map[string]any{"name": "test"},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "invalid name format",
			data: map[string]any{
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
			data: map[string]any{
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
			data: map[string]any{
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
			data: map[string]any{
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
			data: map[string]any{
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
			data: map[string]any{
				"name":        "test",
				"description": "A test agent for doing things.",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: A test agent for doing things.\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "valid memory scope user",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"memory":      "user",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmemory: user\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid memory scope project",
			data: map[string]any{
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
			name: "valid memory scope local",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"memory":      "local",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmemory: local\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "invalid memory scope",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"memory":      "global",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmemory: global\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "valid permissionMode default",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"permissionMode": "default",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\npermissionMode: default\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid permissionMode bypassPermissions",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"permissionMode": "bypassPermissions",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\npermissionMode: bypassPermissions\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid permissionMode delegate",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"permissionMode": "delegate",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\npermissionMode: delegate\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "invalid permissionMode",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"permissionMode": "yolo",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\npermissionMode: yolo\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "valid maxTurns positive integer",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"maxTurns":    10,
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmaxTurns: 10\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "invalid maxTurns zero",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"maxTurns":    0,
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmaxTurns: 0\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "invalid maxTurns negative",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"maxTurns":    -5,
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmaxTurns: -5\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "invalid maxTurns string",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"maxTurns":    "ten",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmaxTurns: ten\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "valid mcpServers array",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"mcpServers":  []any{"filesystem", "github"},
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmcpServers:\n  - filesystem\n  - github\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "mcpServers with empty string element",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"mcpServers":  []any{"filesystem", ""},
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmcpServers:\n  - filesystem\n  - \"\"\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "mcpServers as string instead of array",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"mcpServers":  "filesystem",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmcpServers: filesystem\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "valid model sonnet",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "sonnet",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: sonnet\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model haiku",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "haiku",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: haiku\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model opus",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "opus",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: opus\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model inherit",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "inherit",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: inherit\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model with version suffix",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "sonnet[1m]",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: sonnet[1m]\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid model opusplan",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       "opusplan",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: opusplan\n---\n",
			wantErrCount:  0,
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
		tools any
		want  bool
	}{
		{"wildcard", "*", true},
		{"string with Edit", "Read, Edit, Bash", true},
		{"string with Write", "Write", true},
		{"string with MultiEdit", "MultiEdit", true},
		{"string without editing", "Read, Bash", false},
		{"array with Edit", []any{"Read", "Edit"}, true},
		{"array without editing", []any{"Read", "Bash"}, false},
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
		data         map[string]any
		wantContains []string
	}{
		{
			name:     "XML tags in description",
			filePath: "agents/test.md",
			contents: "---\ndescription: Test <xml>tag</xml>\n---\n",
			data: map[string]any{
				"description": "Test <xml>tag</xml>",
			},
			wantContains: []string{"XML-like tags"},
		},
		{
			name:     "bloat section Quick Reference",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test\n---\n## Quick Reference\n",
			data: map[string]any{
				"name":        "test",
				"description": "test",
			},
			wantContains: []string{"Quick Reference"},
		},
		{
			name:     "missing model",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test\n---\n",
			data: map[string]any{
				"name":        "test",
				"description": "test",
			},
			wantContains: []string{"model"},
		},
		{
			name:     "editing tools without permissionMode",
			filePath: "agents/test.md",
			contents: "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\ntools: Edit\n---\n",
			data: map[string]any{
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
	expected := []string{"name", "description", "model", "color", "tools", "disallowedTools", "permissionMode", "maxTurns", "skills", "hooks", "memory", "mcpServers"}
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

func TestValidateAgentModelValues(t *testing.T) {
	tests := []struct {
		name         string
		model        string
		wantWarnings int
	}{
		{"valid haiku", "haiku", 0},
		{"valid sonnet", "sonnet", 0},
		{"valid opus", "opus", 0},
		{"valid inherit", "inherit", 0},
		{"valid opusplan", "opusplan", 0},
		{"valid sonnet with version", "sonnet[1m]", 0},
		{"valid haiku with version", "haiku[2]", 0},
		{"valid opus with version", "opus[v3]", 0},
		{"invalid unknown-model", "unknown-model", 1},
		{"invalid empty", "", 1},
		{"invalid random", "fast", 1},
		{"invalid arbitrary string", "turbo-3", 1},
		{"invalid partial", "son", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"model":       tt.model,
			}
			contents := "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nmodel: " + tt.model + "\n---\n"

			errors := validateAgentSpecific(data, "agents/test.md", contents)

			warnings := 0
			for _, e := range errors {
				if e.Severity == "warning" && strings.Contains(e.Message, "Unknown model") {
					warnings++
				}
			}

			if warnings != tt.wantWarnings {
				t.Errorf("validateAgentSpecific() model warnings = %d, want %d for model %q", warnings, tt.wantWarnings, tt.model)
				for _, e := range errors {
					if e.Severity == "warning" {
						t.Logf("  Warning: %s", e.Message)
					}
				}
			}
		})
	}
}

func TestValidateAgentMaxTurnsDontAskInfo(t *testing.T) {
	tests := []struct {
		name     string
		data     map[string]any
		wantInfo bool
	}{
		{
			name: "maxTurns with dontAsk emits info",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"maxTurns":       10,
				"permissionMode": "dontAsk",
			},
			wantInfo: true,
		},
		{
			name: "maxTurns without dontAsk no info",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"maxTurns":       10,
				"permissionMode": "default",
			},
			wantInfo: false,
		},
		{
			name: "dontAsk without maxTurns no info",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"permissionMode": "dontAsk",
			},
			wantInfo: false,
		},
		{
			name: "neither maxTurns nor dontAsk no info",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
			},
			wantInfo: false,
		},
		{
			name: "maxTurns with bypassPermissions no info",
			data: map[string]any{
				"name":           "test",
				"description":    "test. Use PROACTIVELY when testing.",
				"maxTurns":       5,
				"permissionMode": "bypassPermissions",
			},
			wantInfo: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents := "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\n---\n"
			errors := validateAgentSpecific(tt.data, "agents/test.md", contents)

			foundInfo := false
			for _, e := range errors {
				if e.Severity == "info" && strings.Contains(e.Message, "maxTurns with permissionMode 'dontAsk'") {
					foundInfo = true
					break
				}
			}

			if foundInfo != tt.wantInfo {
				t.Errorf("validateAgentSpecific() info about maxTurns+dontAsk = %v, want %v", foundInfo, tt.wantInfo)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestValidModelPattern(t *testing.T) {
	tests := []struct {
		model string
		want  bool
	}{
		{"haiku", true},
		{"sonnet", true},
		{"opus", true},
		{"inherit", true},
		{"opusplan", true},
		{"sonnet[1m]", true},
		{"haiku[2]", true},
		{"opus[v3]", true},
		{"sonnet[latest]", true},
		{"unknown-model", false},
		{"turbo-3", false},
		{"", false},
		{"SONNET", false},
		{"Haiku", false},
		{"sonnet[]", false},
		{"fast", false},
	}

	for _, tt := range tests {
		t.Run(tt.model, func(t *testing.T) {
			got := validModelPattern.MatchString(tt.model)
			if got != tt.want {
				t.Errorf("validModelPattern.MatchString(%q) = %v, want %v", tt.model, got, tt.want)
			}
		})
	}
}
