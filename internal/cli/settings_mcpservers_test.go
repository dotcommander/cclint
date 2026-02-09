package cli

import (
	"testing"
)

func TestValidateMCPServers(t *testing.T) {
	tests := []struct {
		name           string
		mcpServers     any
		wantErrorCount int
	}{
		{
			name:           "not an object",
			mcpServers:     "not an object",
			wantErrorCount: 1,
		},
		{
			name:           "nil value",
			mcpServers:     nil,
			wantErrorCount: 1,
		},
		{
			name:           "empty object is valid",
			mcpServers:     map[string]any{},
			wantErrorCount: 0,
		},
		{
			name: "valid server with command and args",
			mcpServers: map[string]any{
				"filesystem": map[string]any{
					"command": "npx",
					"args":    []any{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid server with command only",
			mcpServers: map[string]any{
				"simple-server": map[string]any{
					"command": "/usr/local/bin/mcp-server",
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid server with env vars",
			mcpServers: map[string]any{
				"api-server": map[string]any{
					"command": "node",
					"args":    []any{"server.js"},
					"env": map[string]any{
						"API_KEY":  "sk-test-123",
						"NODE_ENV": "production",
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid server with cwd",
			mcpServers: map[string]any{
				"local-server": map[string]any{
					"command": "python",
					"args":    []any{"server.py"},
					"cwd":     "/home/user/mcp-servers",
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid server with all fields",
			mcpServers: map[string]any{
				"full-server": map[string]any{
					"command": "node",
					"args":    []any{"index.js", "--port", "3000"},
					"env": map[string]any{
						"DEBUG": "true",
					},
					"cwd": "/opt/mcp",
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "missing command field",
			mcpServers: map[string]any{
				"no-cmd": map[string]any{
					"args": []any{"--flag"},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "empty command string",
			mcpServers: map[string]any{
				"empty-cmd": map[string]any{
					"command": "",
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "command is not a string",
			mcpServers: map[string]any{
				"bad-cmd": map[string]any{
					"command": 42,
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "server config is not an object",
			mcpServers: map[string]any{
				"bad-server": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "args is not an array",
			mcpServers: map[string]any{
				"bad-args": map[string]any{
					"command": "node",
					"args":    "not an array",
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "args contains non-string element",
			mcpServers: map[string]any{
				"bad-arg-elem": map[string]any{
					"command": "node",
					"args":    []any{"valid", 42, "also-valid"},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "env is not an object",
			mcpServers: map[string]any{
				"bad-env": map[string]any{
					"command": "node",
					"env":     "not an object",
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "env value is not a string",
			mcpServers: map[string]any{
				"bad-env-val": map[string]any{
					"command": "node",
					"env": map[string]any{
						"GOOD_KEY": "good-value",
						"BAD_KEY":  42,
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "cwd is not a string",
			mcpServers: map[string]any{
				"bad-cwd": map[string]any{
					"command": "node",
					"cwd":     42,
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "multiple servers mixed valid and invalid",
			mcpServers: map[string]any{
				"good-server": map[string]any{
					"command": "npx",
					"args":    []any{"-y", "mcp-server"},
				},
				"bad-server": map[string]any{
					"args": []any{"--flag"},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "multiple errors in one server",
			mcpServers: map[string]any{
				"very-bad": map[string]any{
					"command": 42,
					"args":    "not array",
					"env":     "not object",
					"cwd":     123,
				},
			},
			wantErrorCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateMCPServers(tt.mcpServers, "settings.json")
			if len(errs) != tt.wantErrorCount {
				t.Errorf("validateMCPServers() error count = %d, want %d", len(errs), tt.wantErrorCount)
				for _, e := range errs {
					t.Logf("  - [%s] %s (source: %s)", e.Severity, e.Message, e.Source)
				}
			}
			// Verify all errors use anthropic-docs source
			for _, e := range errs {
				if e.Source != "anthropic-docs" {
					t.Errorf("expected source 'anthropic-docs', got %q for: %s", e.Source, e.Message)
				}
			}
		})
	}
}
