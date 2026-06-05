package lint

import (
	"testing"
)

// TestValidateAgentOptionalFields covers memory, permissionMode, mcpServers,
// isolation, background, effort, initialPrompt, and requiredMcpServers /
// criticalSystemReminder_EXPERIMENTAL validation cases.
// Required-field and name cases live in agent_fields_test.go.
// Model/maxTurns cases live in agent_model_test.go.
func TestValidateAgentOptionalFields(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]any
		filePath      string
		contents      string
		wantErrCount  int
		wantSuggCount int
	}{
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
			name: "valid isolation worktree",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"isolation":   "worktree",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nisolation: worktree\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid background true",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"background":  true,
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nbackground: true\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid background false",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"background":  false,
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nbackground: false\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid effort field",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"effort":      "high",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\neffort: high\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid agent without isolation or background",
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
			name: "valid agent with both isolation and background",
			data: map[string]any{
				"name":        "test",
				"description": "test. Use PROACTIVELY when testing.",
				"isolation":   "worktree",
				"background":  true,
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nisolation: worktree\nbackground: true\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid initialPrompt field",
			data: map[string]any{
				"name":          "test",
				"description":   "test. Use PROACTIVELY when testing.",
				"initialPrompt": "Analyze the current directory and report findings.",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\ninitialPrompt: Analyze the current directory and report findings.\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			// v2.1.156: both fields are real agent frontmatter keys; neither
			// should produce an "unknown frontmatter field" suggestion.
			name: "valid requiredMcpServers and criticalSystemReminder_EXPERIMENTAL fields",
			data: map[string]any{
				"name":                                "test",
				"description":                         "test. Use PROACTIVELY when testing.",
				"requiredMcpServers":                  []any{"filesystem"},
				"criticalSystemReminder_EXPERIMENTAL": "Always confirm before deleting.",
			},
			filePath:      "agents/test.md",
			contents:      "---\nname: test\ndescription: test. Use PROACTIVELY when testing.\nrequiredMcpServers:\n  - filesystem\ncriticalSystemReminder_EXPERIMENTAL: Always confirm before deleting.\n---\n",
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
