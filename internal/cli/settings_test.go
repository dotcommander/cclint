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
		hooks          interface{}
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
			hooks: map[string]interface{}{
				"UnknownEvent": []interface{}{},
			},
			wantErrorCount: 1,
		},
		{
			name: "invalid hook configuration",
			hooks: map[string]interface{}{
				"PreToolUse": "not an array",
			},
			wantErrorCount: 1,
		},
		{
			name: "missing matcher",
			hooks: map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"hooks": []interface{}{},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "missing hooks field",
			hooks: map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "invalid hooks type",
			hooks: map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks":   "not an array",
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "missing hook type",
			hooks: map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{},
						},
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "invalid hook type value",
			hooks: map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"PostToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"Stop": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"Stop": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"Setup": []interface{}{
					map[string]interface{}{
						"matcher": "init",
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"Setup": []interface{}{
					map[string]interface{}{
						"matcher": "maintenance",
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"SubagentStart": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"PostToolUseFailure": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"TeammateIdle": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"TaskCompleted": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"PostToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
			hooks: map[string]interface{}{
				"PreToolUse": []interface{}{
					map[string]interface{}{
						"matcher": map[string]interface{}{},
						"hooks": []interface{}{
							map[string]interface{}{
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
		data           map[string]interface{}
		wantErrorCount int
	}{
		{
			name:           "no hooks",
			data:           map[string]interface{}{},
			wantErrorCount: 0,
		},
		{
			name: "valid hooks",
			data: map[string]interface{}{
				"hooks": map[string]interface{}{
					"PreToolUse": []interface{}{
						map[string]interface{}{
							"matcher": map[string]interface{}{},
							"hooks": []interface{}{
								map[string]interface{}{
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
			data: map[string]interface{}{
				"hooks": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "valid permissions with allow and deny",
			data: map[string]interface{}{
				"permissions": map[string]interface{}{
					"allow": []interface{}{"Bash(npm*)", "Read", "Edit"},
					"deny":  []interface{}{"Bash(rm*)"},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid permissions structure",
			data: map[string]interface{}{
				"permissions": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "valid hooks and valid permissions together",
			data: map[string]interface{}{
				"hooks": map[string]interface{}{
					"PreToolUse": []interface{}{
						map[string]interface{}{
							"matcher": map[string]interface{}{},
							"hooks": []interface{}{
								map[string]interface{}{
									"type":    "command",
									"command": "echo test",
								},
							},
						},
					},
				},
				"permissions": map[string]interface{}{
					"allow": []interface{}{"Bash(npm*)"},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid mcpServers with command and args",
			data: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"my-server": map[string]interface{}{
						"command": "npx",
						"args":   []interface{}{"-y", "@modelcontextprotocol/server-filesystem"},
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "invalid mcpServers structure",
			data: map[string]interface{}{
				"mcpServers": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "mcpServers missing command",
			data: map[string]interface{}{
				"mcpServers": map[string]interface{}{
					"bad-server": map[string]interface{}{
						"args": []interface{}{"--flag"},
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
		mcpServers     interface{}
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
			mcpServers:     map[string]interface{}{},
			wantErrorCount: 0,
		},
		{
			name: "valid server with command and args",
			mcpServers: map[string]interface{}{
				"filesystem": map[string]interface{}{
					"command": "npx",
					"args":   []interface{}{"-y", "@modelcontextprotocol/server-filesystem", "/tmp"},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid server with command only",
			mcpServers: map[string]interface{}{
				"simple-server": map[string]interface{}{
					"command": "/usr/local/bin/mcp-server",
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid server with env vars",
			mcpServers: map[string]interface{}{
				"api-server": map[string]interface{}{
					"command": "node",
					"args":   []interface{}{"server.js"},
					"env": map[string]interface{}{
						"API_KEY":  "sk-test-123",
						"NODE_ENV": "production",
					},
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid server with cwd",
			mcpServers: map[string]interface{}{
				"local-server": map[string]interface{}{
					"command": "python",
					"args":   []interface{}{"server.py"},
					"cwd":    "/home/user/mcp-servers",
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "valid server with all fields",
			mcpServers: map[string]interface{}{
				"full-server": map[string]interface{}{
					"command": "node",
					"args":   []interface{}{"index.js", "--port", "3000"},
					"env": map[string]interface{}{
						"DEBUG": "true",
					},
					"cwd": "/opt/mcp",
				},
			},
			wantErrorCount: 0,
		},
		{
			name: "missing command field",
			mcpServers: map[string]interface{}{
				"no-cmd": map[string]interface{}{
					"args": []interface{}{"--flag"},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "empty command string",
			mcpServers: map[string]interface{}{
				"empty-cmd": map[string]interface{}{
					"command": "",
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "command is not a string",
			mcpServers: map[string]interface{}{
				"bad-cmd": map[string]interface{}{
					"command": 42,
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "server config is not an object",
			mcpServers: map[string]interface{}{
				"bad-server": "not an object",
			},
			wantErrorCount: 1,
		},
		{
			name: "args is not an array",
			mcpServers: map[string]interface{}{
				"bad-args": map[string]interface{}{
					"command": "node",
					"args":   "not an array",
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "args contains non-string element",
			mcpServers: map[string]interface{}{
				"bad-arg-elem": map[string]interface{}{
					"command": "node",
					"args":   []interface{}{"valid", 42, "also-valid"},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "env is not an object",
			mcpServers: map[string]interface{}{
				"bad-env": map[string]interface{}{
					"command": "node",
					"env":    "not an object",
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "env value is not a string",
			mcpServers: map[string]interface{}{
				"bad-env-val": map[string]interface{}{
					"command": "node",
					"env": map[string]interface{}{
						"GOOD_KEY": "good-value",
						"BAD_KEY":  42,
					},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "cwd is not a string",
			mcpServers: map[string]interface{}{
				"bad-cwd": map[string]interface{}{
					"command": "node",
					"cwd":    42,
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "multiple servers mixed valid and invalid",
			mcpServers: map[string]interface{}{
				"good-server": map[string]interface{}{
					"command": "npx",
					"args":   []interface{}{"-y", "mcp-server"},
				},
				"bad-server": map[string]interface{}{
					"args": []interface{}{"--flag"},
				},
			},
			wantErrorCount: 1,
		},
		{
			name: "multiple errors in one server",
			mcpServers: map[string]interface{}{
				"very-bad": map[string]interface{}{
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
		perms      interface{}
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
			perms:      []interface{}{"Bash"},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name:       "empty object is valid",
			perms:      map[string]interface{}{},
			wantErrors: 0,
		},
		{
			name: "valid allow only",
			perms: map[string]interface{}{
				"allow": []interface{}{"Bash(npm*)", "Read", "Write"},
			},
			wantErrors: 0,
		},
		{
			name: "valid deny only",
			perms: map[string]interface{}{
				"deny": []interface{}{"Bash(rm*)", "Bash(sudo*)"},
			},
			wantErrors: 0,
		},
		{
			name: "valid allow and deny",
			perms: map[string]interface{}{
				"allow": []interface{}{"Bash(npm*)", "Read", "Edit", "Glob", "Grep"},
				"deny":  []interface{}{"Bash(rm*)"},
			},
			wantErrors: 0,
		},
		{
			name: "unknown key in permissions",
			perms: map[string]interface{}{
				"allow":  []interface{}{"Read"},
				"permit": []interface{}{"Write"},
			},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "allow not an array",
			perms: map[string]interface{}{
				"allow": "Read",
			},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "deny not an array",
			perms: map[string]interface{}{
				"deny": 42,
			},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "empty string entry",
			perms: map[string]interface{}{
				"allow": []interface{}{""},
			},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "non-string entry",
			perms: map[string]interface{}{
				"allow": []interface{}{42},
			},
			wantErrors: 1,
			wantSeverity: "error",
		},
		{
			name: "unknown tool name is suggestion",
			perms: map[string]interface{}{
				"allow": []interface{}{"FakeTool"},
			},
			wantErrors: 1,
			wantSeverity: "suggestion",
		},
		{
			name: "unknown tool with pattern is suggestion",
			perms: map[string]interface{}{
				"deny": []interface{}{"FakeTool(rm*)"},
			},
			wantErrors: 1,
			wantSeverity: "suggestion",
		},
		{
			name: "MCP tool is valid",
			perms: map[string]interface{}{
				"allow": []interface{}{"mcp__my_server_tool"},
			},
			wantErrors: 0,
		},
		{
			name: "all known tools are valid",
			perms: map[string]interface{}{
				"allow": []interface{}{
					"Bash", "Read", "Write", "Edit", "Glob", "Grep",
					"Task", "Skill", "WebSearch", "WebFetch",
					"TodoRead", "TodoWrite", "TaskOutput", "AskUser",
				},
			},
			wantErrors: 0,
		},
		{
			name: "mixed valid and unknown",
			perms: map[string]interface{}{
				"allow": []interface{}{"Read", "UnknownTool", "Edit"},
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

func TestValidateComponentHooks(t *testing.T) {
	validHook := []interface{}{
		map[string]interface{}{
			"matcher": map[string]interface{}{},
			"hooks": []interface{}{
				map[string]interface{}{
					"type":    "command",
					"command": "echo test",
				},
			},
		},
	}

	tests := []struct {
		name           string
		hooks          interface{}
		wantErrorCount int
	}{
		{
			name: "accepts PreToolUse",
			hooks: map[string]interface{}{
				"PreToolUse": validHook,
			},
			wantErrorCount: 0,
		},
		{
			name: "accepts PostToolUse",
			hooks: map[string]interface{}{
				"PostToolUse": validHook,
			},
			wantErrorCount: 0,
		},
		{
			name: "accepts Stop",
			hooks: map[string]interface{}{
				"Stop": validHook,
			},
			wantErrorCount: 0,
		},
		{
			name: "rejects SessionStart",
			hooks: map[string]interface{}{
				"SessionStart": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects Setup",
			hooks: map[string]interface{}{
				"Setup": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects SubagentStart",
			hooks: map[string]interface{}{
				"SubagentStart": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects TeammateIdle for components",
			hooks: map[string]interface{}{
				"TeammateIdle": validHook,
			},
			wantErrorCount: 1,
		},
		{
			name: "rejects TaskCompleted for components",
			hooks: map[string]interface{}{
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
