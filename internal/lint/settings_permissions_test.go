package lint

import (
	"testing"
)

func TestValidatePermissions(t *testing.T) {
	tests := []struct {
		name         string
		perms        any
		wantErrors   int
		wantSeverity string // if set, check first error has this severity
	}{
		{
			name:         "nil permissions",
			perms:        nil,
			wantErrors:   1,
			wantSeverity: "error",
		},
		{
			name:         "not an object",
			perms:        "not an object",
			wantErrors:   1,
			wantSeverity: "error",
		},
		{
			name:         "not an object (array)",
			perms:        []any{"Bash"},
			wantErrors:   1,
			wantSeverity: "error",
		},
		{
			name:       "empty object is valid",
			perms:      map[string]any{},
			wantErrors: 0,
		},
		{
			name: "valid allow only",
			perms: map[string]any{
				"allow": []any{"Bash(npm*)", "Read", "Write"},
			},
			wantErrors: 0,
		},
		{
			name: "valid deny only",
			perms: map[string]any{
				"deny": []any{"Bash(rm*)", "Bash(sudo*)"},
			},
			wantErrors: 0,
		},
		{
			name: "valid allow and deny",
			perms: map[string]any{
				"allow": []any{"Bash(npm*)", "Read", "Edit", "Glob", "Grep"},
				"deny":  []any{"Bash(rm*)"},
			},
			wantErrors: 0,
		},
		{
			name: "unknown key in permissions",
			perms: map[string]any{
				"allow":  []any{"Read"},
				"permit": []any{"Write"},
			},
			wantErrors:   1,
			wantSeverity: "error",
		},
		{
			name: "allow not an array",
			perms: map[string]any{
				"allow": "Read",
			},
			wantErrors:   1,
			wantSeverity: "error",
		},
		{
			name: "deny not an array",
			perms: map[string]any{
				"deny": 42,
			},
			wantErrors:   1,
			wantSeverity: "error",
		},
		{
			name: "empty string entry",
			perms: map[string]any{
				"allow": []any{""},
			},
			wantErrors:   1,
			wantSeverity: "error",
		},
		{
			name: "non-string entry",
			perms: map[string]any{
				"allow": []any{42},
			},
			wantErrors:   1,
			wantSeverity: "error",
		},
		{
			name: "unknown tool name is suggestion",
			perms: map[string]any{
				"allow": []any{"FakeTool"},
			},
			wantErrors:   1,
			wantSeverity: "suggestion",
		},
		{
			name: "unknown tool with pattern is suggestion",
			perms: map[string]any{
				"deny": []any{"FakeTool(rm*)"},
			},
			wantErrors:   1,
			wantSeverity: "suggestion",
		},
		{
			name: "MCP tool is valid",
			perms: map[string]any{
				"allow": []any{"mcp__my_server_tool"},
			},
			wantErrors: 0,
		},
		{
			name: "all known tools are valid",
			perms: map[string]any{
				"allow": []any{
					"Bash", "Read", "Write", "Edit", "Glob", "Grep",
					"Task", "Skill", "WebSearch", "WebFetch",
					"TodoRead", "TodoWrite", "TaskOutput", "AskUser",
				},
			},
			wantErrors: 0,
		},
		{
			name: "mixed valid and unknown",
			perms: map[string]any{
				"allow": []any{"Read", "UnknownTool", "Edit"},
			},
			wantErrors:   1,
			wantSeverity: "suggestion",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validatePermissions(tt.perms, "settings.json")
			if len(errs) != tt.wantErrors {
				t.Errorf("validatePermissions() error count = %d, want %d", len(errs), tt.wantErrors)
				for _, e := range errs {
					t.Logf("  - [%s] %s (source: %s)", e.Severity, e.Message, e.Source)
				}
			}
			if tt.wantSeverity != "" && len(errs) > 0 {
				if errs[0].Severity != tt.wantSeverity {
					t.Errorf("first error severity = %q, want %q", errs[0].Severity, tt.wantSeverity)
				}
			}
		})
	}
}

func TestExtractToolName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Bash", "Bash"},
		{"Bash(npm*)", "Bash"},
		{"Read", "Read"},
		{"Write(/path/**)", "Write"},
		{"mcp__my_server_tool", "mcp__"},
		{"mcp__foo", "mcp__"},
		{"Edit", "Edit"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := extractToolName(tt.input)
			if got != tt.want {
				t.Errorf("extractToolName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
