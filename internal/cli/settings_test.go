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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateSettingsSpecific(tt.data, "settings.json")
			if len(errors) != tt.wantErrorCount {
				t.Errorf("validateSettingsSpecific() error count = %d, want %d", len(errors), tt.wantErrorCount)
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
