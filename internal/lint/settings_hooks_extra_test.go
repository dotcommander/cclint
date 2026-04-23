package lint

import (
	"testing"
)

// TestValidateHooksValidEvents covers valid-event and valid-type cases extracted
// from TestValidateHooks to keep that file under the 300-line limit.
func TestValidateHooksValidEvents(t *testing.T) {
	// helper: wrap a single hook command under an event key
	hook := func(event string, hookFields map[string]any) map[string]any {
		return map[string]any{
			event: []any{
				map[string]any{
					"matcher": map[string]any{},
					"hooks":   []any{hookFields},
				},
			},
		}
	}
	cmd := func(command string) map[string]any {
		return map[string]any{"type": "command", "command": command}
	}

	tests := []struct {
		name           string
		hooks          map[string]any
		wantErrorCount int
	}{
		{"valid StopFailure hook", hook("StopFailure", cmd("echo stop failed")), 0},
		{"valid SubagentStart hook", hook("SubagentStart", cmd("echo subagent started")), 0},
		{"valid PostToolUseFailure hook", hook("PostToolUseFailure", cmd("echo tool failed")), 0},
		{"valid TeammateIdle hook", hook("TeammateIdle", cmd("echo teammate idle")), 0},
		{"valid TaskCompleted hook", hook("TaskCompleted", cmd("echo task done")), 0},
		{"valid ConfigChange hook", hook("ConfigChange", cmd("echo config changed")), 0},
		{"valid WorktreeCreate hook", hook("WorktreeCreate", cmd("echo worktree created")), 0},
		{"valid TaskCreated hook", hook("TaskCreated", cmd("echo task created")), 0},
		{"valid WorktreeRemove hook", hook("WorktreeRemove", cmd("echo worktree removed")), 0},
		{"valid InstructionsLoaded hook", hook("InstructionsLoaded", cmd("echo instructions loaded")), 0},
		{"valid agent hook", hook("PreToolUse", map[string]any{"type": "agent"}), 0},
		{"valid http hook with url", hook("PreToolUse", map[string]any{"type": "http", "url": "https://example.com/hook"}), 0},
		{"valid mcp_tool hook type accepted (v2.1.118+)", hook("PreToolUse", map[string]any{"type": "mcp_tool"}), 0},
		{"invalid hook type rejected", hook("PreToolUse", map[string]any{"type": "webhook"}), 1},
		{"http hook missing url", hook("PreToolUse", map[string]any{"type": "http"}), 1},
		{
			name: "valid Setup hook event with matcher",
			hooks: map[string]any{
				"Setup": []any{
					map[string]any{
						"matcher": "init",
						"hooks":   []any{cmd("echo setup init")},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "http hook with headers and allowedEnvVars",
			hooks: hook("PostToolUse", map[string]any{
				"type": "http",
				"url":  "https://example.com/webhook",
				"headers": map[string]any{
					"Authorization": "Bearer token",
					"Content-Type":  "application/json",
				},
				"allowedEnvVars": []any{"MY_SECRET", "API_KEY"},
			}),
			wantErrorCount: 0,
		},
		{
			name:           "command hook with async true",
			hooks:          hook("PostToolUse", map[string]any{"type": "command", "command": "echo async test", "async": true}),
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateHooks(tt.hooks, "settings.json")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateHooks() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - %s: %s", err.Severity, err.Message)
				}
			}
		})
	}
}
