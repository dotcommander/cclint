package cli

import (
	"testing"
)

func TestValidateHookCommandSecurity(t *testing.T) {
	tests := []struct {
		name             string
		command          string
		wantWarningCount int
	}{
		{
			name:             "safe command",
			command:          `echo "test"`,
			wantWarningCount: 0,
		},
		{
			name:             "unquoted variable",
			command:          `echo $VAR`,
			wantWarningCount: 1,
		},
		{
			name:             "path traversal",
			command:          `cat ../../../etc/passwd`,
			wantWarningCount: 1,
		},
		{
			name:             "hardcoded absolute path",
			command:          `cat "/Users/test/file.txt"`,
			wantWarningCount: 1,
		},
		{
			name:             "env file access",
			command:          `cat .env`,
			wantWarningCount: 1,
		},
		{
			name:             "git directory access",
			command:          `cat .git/config`,
			wantWarningCount: 1,
		},
		{
			name:             "credentials file",
			command:          `cat credentials.json`,
			wantWarningCount: 1,
		},
		{
			name:             "ssh directory",
			command:          `ls .ssh/`,
			wantWarningCount: 1,
		},
		{
			name:             "aws config",
			command:          `cat .aws/credentials`,
			wantWarningCount: 2, // credentials + aws config
		},
		{
			name:             "ssh private key",
			command:          `cat ~/.ssh/id_rsa`,
			wantWarningCount: 2, // ssh + private key
		},
		{
			name:             "eval command",
			command:          `eval "dangerous"`,
			wantWarningCount: 1,
		},
		{
			name:             "command substitution",
			command:          `echo $(whoami)`,
			wantWarningCount: 1,
		},
		{
			name:             "backtick substitution",
			command:          "echo `whoami`",
			wantWarningCount: 1,
		},
		{
			name:             "redirect to dev",
			command:          `echo test > /dev/null`,
			wantWarningCount: 1,
		},
		{
			name:             "multiple issues",
			command:          `eval cat $VAR ../../../etc/passwd`,
			wantWarningCount: 3, // eval + unquoted var + path traversal
		},
		{
			name:             "safe with CLAUDE_PROJECT_DIR",
			command:          `cat "$CLAUDE_PROJECT_DIR/file.txt"`,
			wantWarningCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := hookContext{EventName: "TestEvent", HookIdx: 0, InnerIdx: 0, FilePath: "test.json"}
			warnings := validateHookCommandSecurity(tt.command, ctx)
			if len(warnings) != tt.wantWarningCount {
				t.Errorf("validateHookCommandSecurity() warning count = %d, want %d", len(warnings), tt.wantWarningCount)
				for _, warn := range warnings {
					t.Logf("  - %s", warn.Message)
				}
			}
		})
	}
}
