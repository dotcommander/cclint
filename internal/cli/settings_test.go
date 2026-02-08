package cli

import (
	"testing"
)

func TestLintSettings(t *testing.T) {
	// Test with empty directory
	summary, err := LintSettings("testdata/empty", false, false, true)
	if err != nil {
		t.Fatalf("LintSettings() error = %v", err)
	}
	if summary == nil {
		t.Fatal("LintSettings() returned nil summary")
	}
}

func TestValidateHooks(t *testing.T) {
	tests := []struct {
		name           string
		hooks          any
		wantErrorCount int
	}{
		{
			name:           "nil hooks",
			hooks:          nil,
			wantErrorCount: 1,
		},
		{
			name:           "invalid type",
			hooks:          "not an object",
			wantErrorCount: 1,
		},
		{
			name: "unknown event",
			hooks: map[string]any{
				"UnknownEvent": []any{},
			},
			wantErrorCount: 1,
		},
		{
			name: "invalid hook configuration",
			hooks: map[string]any{
				"PreToolUse": "not an array",
			},
			wantErrorCount: 1,
		},
		{
			name: "missing matcher",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"hooks": []any{},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "missing hooks field",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "invalid hooks type",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks":   "not an array",
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "missing hook type",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "invalid hook type value",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "invalid",
							},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "command hook missing command",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "command",
							},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "prompt hook on unsupported event",
			hooks: map[string]any{
				"PostToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "prompt",
							},
						},
					},
				},
			},
			wantErrorCount: 2, // unsupported event + missing prompt field
		},
		{
			name: "prompt hook missing prompt",
			hooks: map[string]any{
				"Stop": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "prompt",
							},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "valid command hook",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo test",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid prompt hook",
			hooks: map[string]any{
				"Stop": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":   "prompt",
								"prompt": "Test prompt",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid Setup hook with init matcher",
			hooks: map[string]any{
				"Setup": []any{
					map[string]any{
						"matcher": "init",
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo setup init",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid Setup hook with maintenance matcher",
			hooks: map[string]any{
				"Setup": []any{
					map[string]any{
						"matcher": "maintenance",
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo setup maintenance",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid SubagentStart hook",
			hooks: map[string]any{
				"SubagentStart": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo subagent started",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid PostToolUseFailure hook",
			hooks: map[string]any{
				"PostToolUseFailure": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo tool failed",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid TeammateIdle hook",
			hooks: map[string]any{
				"TeammateIdle": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo teammate idle",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid TaskCompleted hook",
			hooks: map[string]any{
				"TaskCompleted": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo task done",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid agent hook",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "agent",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "command hook with async true",
			hooks: map[string]any{
				"PostToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo async test",
								"async":   true,
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid hook type rejected",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type": "webhook",
							},
						},
					},
				},
			},
			wantErrorCount: 1,
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

func TestValidateHookCommandSecurity(t *testing.T) {
	tests := []struct {
		name            string
		command         string
		wantWarningCount int
	}{
		{
			name:            "safe command",
			command:         `echo "test"`,
			wantWarningCount: 0,
		},
		{
			name:            "unquoted variable",
			command:         `echo $VAR`,
			wantWarningCount: 1,
		},
		{
			name:            "path traversal",
			command:         `cat ../../../etc/passwd`,
			wantWarningCount: 1,
		},
		{
			name:            "hardcoded absolute path",
			command:         `cat "/Users/test/file.txt"`,
			wantWarningCount: 1,
		},
		{
			name:            "env file access",
			command:         `cat .env`,
			wantWarningCount: 1,
		},
		{
			name:            "git directory access",
			command:         `cat .git/config`,
			wantWarningCount: 1,
		},
		{
			name:            "credentials file",
			command:         `cat credentials.json`,
			wantWarningCount: 1,
		},
		{
			name:            "ssh directory",
			command:         `ls .ssh/`,
			wantWarningCount: 1,
		},
		{
			name:            "aws config",
			command:         `cat .aws/credentials`,
			wantWarningCount: 2, // credentials + aws config
		},
		{
			name:            "ssh private key",
			command:         `cat ~/.ssh/id_rsa`,
			wantWarningCount: 2, // ssh + private key
		},
		{
			name:            "eval command",
			command:         `eval "dangerous"`,
			wantWarningCount: 1,
		},
		{
			name:            "command substitution",
			command:         `echo $(whoami)`,
			wantWarningCount: 1,
		},
		{
			name:            "backtick substitution",
			command:         "echo `whoami`",
			wantWarningCount: 1,
		},
		{
			name:            "redirect to dev",
			command:         `echo test > /dev/null`,
			wantWarningCount: 1,
		},
		{
			name:            "multiple issues",
			command:         `eval cat $VAR ../../../etc/passwd`,
			wantWarningCount: 3, // eval + unquoted var + path traversal
		},
		{
			name:            "safe with CLAUDE_PROJECT_DIR",
			command:         `cat "$CLAUDE_PROJECT_DIR/file.txt"`,
			wantWarningCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			warnings := validateHookCommandSecurity(tt.command, "TestEvent", 0, 0, "test.json")
			if len(warnings) != tt.wantWarningCount {
				t.Errorf("validateHookCommandSecurity() warning count = %d, want %d", len(warnings), tt.wantWarningCount)
				for _, warn := range warnings {
					t.Logf("  - %s", warn.Message)
				}
			}
		})
	}
}

func TestValidateSettingsSpecific(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]any
		wantErrorCount int
	}{
		{
			name:           "no hooks",
			data:           map[string]any{},
			wantErrorCount: 0,
		},
		{
			name: "valid hooks",
			data: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": []any{
						map[string]any{
							"matcher": map[string]any{},
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "echo test",
								},
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid hooks",
			data: map[string]any{
				"hooks": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "valid permissions with allow and deny",
			data: map[string]any{
				"permissions": map[string]any{
					"allow": []any{"Bash(npm*)", "Read", "Edit"},
					"deny":  []any{"Bash(rm*)"},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid permissions structure",
			data: map[string]any{
				"permissions": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "valid hooks and valid permissions together",
			data: map[string]any{
				"hooks": map[string]any{
					"PreToolUse": []any{
						map[string]any{
							"matcher": map[string]any{},
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "echo test",
								},
							},
						},
					},
				},
				"permissions": map[string]any{
					"allow": []any{"Bash(npm*)"},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid mcpServers with command and args",
			data: map[string]any{
				"mcpServers": map[string]any{
					"my-server": map[string]any{
						"command": "npx",
						"args":   []any{"-y", "@modelcontextprotocol/server-filesystem"},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid mcpServers structure",
			data: map[string]any{
				"mcpServers": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "mcpServers missing command",
			data: map[string]any{
				"mcpServers": map[string]any{
					"bad-server": map[string]any{
						"args": []any{"--flag"},
					},
				},
			},
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateSettingsSpecific(tt.data, "settings.json")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateSettingsSpecific() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - %s: %s", err.Severity, err.Message)
				}
			}
		})
	}
}

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
					"args":   []any{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
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
					"args":   []any{"server.js"},
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
					"args":   []any{"server.py"},
					"cwd":    "/home/user/mcp-servers",
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid server with all fields",
			mcpServers: map[string]any{
				"full-server": map[string]any{
					"command": "node",
					"args":   []any{"index.js", "--port", "3000"},
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
					"args":   "not an array",
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "args contains non-string element",
			mcpServers: map[string]any{
				"bad-arg-elem": map[string]any{
					"command": "node",
					"args":   []any{"valid", 42, "also-valid"},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "env is not an object",
			mcpServers: map[string]any{
				"bad-env": map[string]any{
					"command": "node",
					"env":    "not an object",
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
					"cwd":    42,
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "multiple servers mixed valid and invalid",
			mcpServers: map[string]any{
				"good-server": map[string]any{
					"command": "npx",
					"args":   []any{"-y", "mcp-server"},
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
					"args":   "not array",
					"env":    "not object",
					"cwd":    123,
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

func TestValidatePermissions(t *testing.T) {
	tests := []struct {
		name       string
		perms      any
		wantErrors int
		wantSeverity string // if set, check first error has this severity
	}{
		{
			name:       "nil permissions",
			perms:      nil,
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name:       "not an object",
			perms:      "not an object",
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name:       "not an object (array)",
			perms:      []any{"Bash"},
			wantErrors: 1,
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
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "allow not an array",
			perms: map[string]any{
				"allow": "Read",
			},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "deny not an array",
			perms: map[string]any{
				"deny": 42,
			},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "empty string entry",
			perms: map[string]any{
				"allow": []any{""},
			},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "non-string entry",
			perms: map[string]any{
				"allow": []any{42},
			},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "unknown tool name is suggestion",
			perms: map[string]any{
				"allow": []any{"FakeTool"},
			},
			wantErrors: 1,
			wantSeverity: "suggestion",
		},
		{
			name: "unknown tool with pattern is suggestion",
			perms: map[string]any{
				"deny": []any{"FakeTool(rm*)"},
			},
			wantErrors: 1,
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
			wantErrors: 1,
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

func TestValidateRules(t *testing.T) {
	tests := []struct {
		name           string
		rules          any
		wantErrorCount int
		wantSeverity   string // check first error severity if set
	}{
		{
			name:           "not an array",
			rules:          "not an array",
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "nil value",
			rules:          nil,
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "empty array is valid",
			rules:          []any{},
			wantErrorCount: 0,
		},
		{
			name:           "valid simple glob",
			rules:          []any{"rules/*.md"},
			wantErrorCount: 0,
		},
		{
			name:           "valid doublestar glob",
			rules:          []any{"**/*.md"},
			wantErrorCount: 0,
		},
		{
			name:           "valid multiple patterns",
			rules:          []any{"rules/*.md", "**/*.md", "docs/rules.md"},
			wantErrorCount: 0,
		},
		{
			name:           "valid brace expansion",
			rules:          []any{"rules/*.{md,txt}"},
			wantErrorCount: 0,
		},
		{
			name:           "empty string entry",
			rules:          []any{""},
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "non-string entry",
			rules:          []any{42},
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "absolute path warns",
			rules:          []any{"/etc/rules/*.md"},
			wantErrorCount: 1,
			wantSeverity:   "warning",
		},
		{
			name:           "leading slash warns",
			rules:          []any{"/rules/*.md"},
			wantErrorCount: 1,
			wantSeverity:   "warning",
		},
		{
			name:           "invalid glob syntax unmatched bracket",
			rules:          []any{"rules/[*.md"},
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "invalid glob syntax unbalanced brace",
			rules:          []any{"rules/{a,b.md"},
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "mixed valid and invalid",
			rules:          []any{"rules/*.md", "", "**/*.txt"},
			wantErrorCount: 1, // empty string
		},
		{
			name:           "mixed valid and absolute",
			rules:          []any{"rules/*.md", "/absolute/path.md"},
			wantErrorCount: 1, // absolute path warning
			wantSeverity:   "warning",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateRules(tt.rules, "settings.json")
			if len(errs) != tt.wantErrorCount {
				t.Errorf("validateRules() error count = %d, want %d", len(errs), tt.wantErrorCount)
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

func TestValidateMatcherToolName(t *testing.T) {
	tests := []struct {
		name           string
		toolNamePattern string
		wantErrorCount int
		wantSeverity   string // check first error severity if set
	}{
		{
			name:           "empty string returns nothing",
			toolNamePattern: "",
			wantErrorCount: 0,
		},
		{
			name:           "valid plain tool name",
			toolNamePattern: "Bash",
			wantErrorCount: 0,
		},
		{
			name:           "valid tool with glob pattern",
			toolNamePattern: "Bash(npm*)",
			wantErrorCount: 0,
		},
		{
			name:           "valid tool with complex glob",
			toolNamePattern: "Write(/src/**/*.go)",
			wantErrorCount: 0,
		},
		{
			name:           "valid Read tool",
			toolNamePattern: "Read",
			wantErrorCount: 0,
		},
		{
			name:           "valid Edit tool",
			toolNamePattern: "Edit",
			wantErrorCount: 0,
		},
		{
			name:           "valid MCP tool",
			toolNamePattern: "mcp__my_server_tool",
			wantErrorCount: 0,
		},
		{
			name:           "unknown tool name is suggestion",
			toolNamePattern: "FakeTool",
			wantErrorCount: 1,
			wantSeverity:   "suggestion",
		},
		{
			name:           "unknown tool with pattern is suggestion",
			toolNamePattern: "FakeTool(rm*)",
			wantErrorCount: 1,
			wantSeverity:   "suggestion",
		},
		{
			name:           "unclosed parenthesis",
			toolNamePattern: "Bash(npm*",
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:           "empty glob inside parens",
			toolNamePattern: "Bash()",
			wantErrorCount: 1,
			wantSeverity:   "warning",
		},
		{
			name:           "invalid glob in parens",
			toolNamePattern: "Bash([invalid)",
			wantErrorCount: 1,
			wantSeverity:   "error",
		},
		{
			name:            "unknown tool AND unclosed paren",
			toolNamePattern: "FakeTool(pattern",
			wantErrorCount:  2, // suggestion + error
		},
		{
			name:            "unknown tool AND invalid glob",
			toolNamePattern: "FakeTool([bad)",
			wantErrorCount:  2, // suggestion + error
		},
		{
			name:           "valid Glob tool with star",
			toolNamePattern: "Glob(**/*.ts)",
			wantErrorCount: 0,
		},
		{
			name:           "valid Grep tool",
			toolNamePattern: "Grep",
			wantErrorCount: 0,
		},
		{
			name:           "valid Task tool with pattern",
			toolNamePattern: "Task(quality*)",
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateMatcherToolName(tt.toolNamePattern, "test location", "test.json")
			if len(errs) != tt.wantErrorCount {
				t.Errorf("validateMatcherToolName(%q) error count = %d, want %d", tt.toolNamePattern, len(errs), tt.wantErrorCount)
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

func TestValidateHooksWithToolNameMatcher(t *testing.T) {
	tests := []struct {
		name           string
		hooks          any
		wantErrorCount int
	}{
		{
			name: "valid hook with known toolName in matcher",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{
							"toolName": "Bash(npm*)",
						},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo pre-tool",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "hook with unknown toolName in matcher produces suggestion",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{
							"toolName": "UnknownTool",
						},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo test",
							},
						},
					},
				},
			},
			wantErrorCount: 1, // suggestion for unknown tool
		},
		{
			name: "hook with unclosed paren in toolName matcher",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{
							"toolName": "Bash(npm*",
						},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo test",
							},
						},
					},
				},
			},
			wantErrorCount: 1, // error for unclosed paren
		},
		{
			name: "hook with string matcher skips toolName validation",
			hooks: map[string]any{
				"Setup": []any{
					map[string]any{
						"matcher": "init",
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo setup",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "hook with empty matcher object skips toolName validation",
			hooks: map[string]any{
				"PreToolUse": []any{
					map[string]any{
						"matcher": map[string]any{},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo test",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "hook with valid MCP toolName in matcher",
			hooks: map[string]any{
				"PostToolUse": []any{
					map[string]any{
						"matcher": map[string]any{
							"toolName": "mcp__my_server_tool",
						},
						"hooks": []any{
							map[string]any{
								"type":    "command",
								"command": "echo post-tool",
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateHooks(tt.hooks, "settings.json")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateHooks() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - [%s] %s (source: %s)", err.Severity, err.Message, err.Source)
				}
			}
		})
	}
}

func TestValidateSettingsSpecificWithRules(t *testing.T) {
	tests := []struct {
		name           string
		data           map[string]any
		wantErrorCount int
	}{
		{
			name: "valid rules array",
			data: map[string]any{
				"rules": []any{"rules/*.md", "**/*.md"},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid rules type",
			data: map[string]any{
				"rules": "not an array",
			},
			wantErrorCount: 1,
		},
		{
			name: "rules with absolute path",
			data: map[string]any{
				"rules": []any{"/etc/rules/*.md"},
			},
			wantErrorCount: 1,
		},
		{
			name: "rules with invalid glob",
			data: map[string]any{
				"rules": []any{"rules/[*.md"},
			},
			wantErrorCount: 1,
		},
		{
			name: "no rules key is valid",
			data: map[string]any{},
			wantErrorCount: 0,
		},
		{
			name: "rules combined with hooks",
			data: map[string]any{
				"rules": []any{"rules/*.md"},
				"hooks": map[string]any{
					"PreToolUse": []any{
						map[string]any{
							"matcher": map[string]any{
								"toolName": "Bash(npm*)",
							},
							"hooks": []any{
								map[string]any{
									"type":    "command",
									"command": "echo test",
								},
							},
						},
					},
				},
			},
			wantErrorCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateSettingsSpecific(tt.data, "settings.json")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateSettingsSpecific() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - [%s] %s (source: %s)", err.Severity, err.Message, err.Source)
				}
			}
		})
	}
}

func TestValidateComponentHooks(t *testing.T) {
	validHook := []any{
		map[string]any{
			"matcher": map[string]any{},
			"hooks": []any{
				map[string]any{
					"type":    "command",
					"command": "echo test",
				},
			},
		},
	}

	tests := []struct {
		name           string
		hooks          any
		wantErrorCount int
	}{
		{
			name: "accepts PreToolUse",
			hooks: map[string]any{
				"PreToolUse": validHook,
			},
			wantErrorCount: 0,
		},
		{
			name: "accepts PostToolUse",
			hooks: map[string]any{
				"PostToolUse": validHook,
			},
			wantErrorCount: 0,
		},
		{
			name: "accepts Stop",
			hooks: map[string]any{
				"Stop": validHook,
			},
			wantErrorCount: 0,
		},
		{
			name: "rejects SessionStart",
			hooks: map[string]any{
				"SessionStart": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects Setup",
			hooks: map[string]any{
				"Setup": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects SubagentStart",
			hooks: map[string]any{
				"SubagentStart": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects TeammateIdle for components",
			hooks: map[string]any{
				"TeammateIdle": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects TaskCompleted for components",
			hooks: map[string]any{
				"TaskCompleted": validHook,
			},
			wantErrorCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := ValidateComponentHooks(tt.hooks, "agent.md")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("ValidateComponentHooks() error count = %d, want %d", len(errors), tt.wantErrorCount)
				for _, err := range errors {
					t.Logf("  - %s: %s", err.Severity, err.Message)
				}
			}
		})
	}
}
