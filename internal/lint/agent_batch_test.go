package lint

import (
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

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

func TestAgentLinterPostProcessBatch(t *testing.T) {
	linter := NewAgentLinter()

	tests := []struct {
		name             string
		files            []discovery.File
		noCycleCheck     bool
		wantCycleErrors  int
		wantTotalErrors  int
		wantFailedFiles  int
		wantSuccessFiles int
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
			crossValidator := crossfile.NewCrossFileValidator(tt.files)
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
