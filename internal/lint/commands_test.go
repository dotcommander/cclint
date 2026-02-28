package lint

import (
	"strings"
	"testing"
)

func TestValidateCommandSpecific(t *testing.T) {
	tests := []struct {
		name          string
		data          map[string]any
		filePath      string
		contents      string
		wantErrCount  int
		wantSuggCount int
	}{
		{
			name: "valid command",
			data: map[string]any{
				"name":          "test-cmd",
				"description":   "Test command",
				"allowed-tools": "Task",
			},
			filePath:      "commands/test-cmd.md",
			contents:      "---\nname: test-cmd\ndescription: Test command\nallowed-tools: Task\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "invalid name format",
			data: map[string]any{
				"name": "TestCmd",
			},
			filePath:      "commands/test.md",
			contents:      "---\nname: TestCmd\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "unknown field",
			data: map[string]any{
				"name": "test",
				"foo":  "bar",
			},
			filePath:      "commands/test.md",
			contents:      "---\nname: test\nfoo: bar\n---\n",
			wantErrCount:  0,
			wantSuggCount: 1,
		},
		{
			name: "valid without name (derived from filename)",
			data: map[string]any{
				"description": "Test",
			},
			filePath:      "commands/test.md",
			contents:      "---\ndescription: Test\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "valid dispatcher tools Task Skill AskUserQuestion",
			data: map[string]any{
				"allowed-tools": "Task, Skill, AskUserQuestion",
			},
			filePath:      "commands/test.md",
			contents:      "---\nallowed-tools: Task, Skill, AskUserQuestion\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
		{
			name: "Bash forbidden in allowed-tools",
			data: map[string]any{
				"allowed-tools": "Task, Bash",
			},
			filePath:      "commands/test.md",
			contents:      "---\nallowed-tools: Task, Bash\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "multiple forbidden tools Write Read Edit",
			data: map[string]any{
				"allowed-tools": "Write, Read, Edit",
			},
			filePath:      "commands/test.md",
			contents:      "---\nallowed-tools: Write, Read, Edit\n---\n",
			wantErrCount:  3,
			wantSuggCount: 0,
		},
		{
			name: "wildcard forbidden in allowed-tools",
			data: map[string]any{
				"allowed-tools": "*",
			},
			filePath:      "commands/test.md",
			contents:      "---\nallowed-tools: \"*\"\n---\n",
			wantErrCount:  1,
			wantSuggCount: 0,
		},
		{
			name: "Task scoped with parens is ok",
			data: map[string]any{
				"allowed-tools": "Task(coder-agent)",
			},
			filePath:      "commands/test.md",
			contents:      "---\nallowed-tools: Task(coder-agent)\n---\n",
			wantErrCount:  0,
			wantSuggCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateCommandSpecific(tt.data, tt.filePath, tt.contents)

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
				t.Errorf("validateCommandSpecific() errors = %d, want %d", errCount, tt.wantErrCount)
				for _, e := range errors {
					if e.Severity == "error" {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}
			if suggCount != tt.wantSuggCount {
				t.Errorf("validateCommandSpecific() suggestions = %d, want %d", suggCount, tt.wantSuggCount)
			}
		})
	}
}

func TestValidateCommandBestPractices(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		contents     string
		data         map[string]any
		wantContains []string
	}{
		{
			name:     "XML tags in description",
			filePath: "commands/test.md",
			contents: "---\ndescription: Test <tag>content</tag>\n---\n",
			data: map[string]any{
				"description": "Test <tag>content</tag>",
			},
			wantContains: []string{"XML-like tags"},
		},
		{
			name:         "implementation section",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\n## Implementation\nSteps here",
			data:         map[string]any{"name": "test"},
			wantContains: []string{"implementation steps"},
		},
		{
			name:         "Task() without allowed-tools",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\nTask(agent): do something",
			data:         map[string]any{"name": "test"},
			wantContains: []string{"allowed-tools"},
		},
		{
			name:         "bloat sections in thin command",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\nTask(agent): do\n## Quick Reference\n",
			data:         map[string]any{"name": "test"},
			wantContains: []string{"Quick Reference"},
		},
		{
			name:         "excessive examples",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\n```bash\nfoo\n```\n```bash\nbar\n```\n```bash\nbaz\n```",
			data:         map[string]any{"name": "test"},
			wantContains: []string{"code examples"},
		},
		{
			name:         "success criteria without checkboxes",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\n---\n## Success\nAll tests pass",
			data:         map[string]any{"name": "test"},
			wantContains: []string{"checkbox format"},
		},
		{
			name:         "Skill without Task delegation",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\nallowed-tools: Skill\n---\nSkill(some-skill)",
			data:         map[string]any{"name": "test", "allowed-tools": "Skill"},
			wantContains: []string{"Skill() without Task() delegation"},
		},
		{
			name:         "Skill with Task in body is valid",
			filePath:     "commands/test.md",
			contents:     "---\nname: test\nallowed-tools: Task, Skill\n---\nTask(agent): run\nSkill(some-skill)",
			data:         map[string]any{"name": "test", "allowed-tools": "Task, Skill"},
			wantContains: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validateCommandBestPractices(tt.filePath, tt.contents, tt.data)

			for _, want := range tt.wantContains {
				found := false
				for _, sugg := range suggestions {
					if strings.Contains(sugg.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateCommandBestPractices() should contain suggestion about %q", want)
					for _, s := range suggestions {
						t.Logf("  Got: %s", s.Message)
					}
				}
			}
		})
	}
}

func TestKnownCommandFields(t *testing.T) {
	expected := []string{"name", "description", "allowed-tools", "argument-hint", "model", "disable-model-invocation"}
	for _, field := range expected {
		if !knownCommandFields[field] {
			t.Errorf("knownCommandFields missing expected field: %s", field)
		}
	}
}

func TestValidateCommandPreprocessing(t *testing.T) {
	tests := []struct {
		name         string
		contents     string
		wantErrCount int
		wantContains []string
	}{
		{
			name:         "no preprocessing directives",
			contents:     "---\nname: test\n---\nJust a normal command body",
			wantErrCount: 0,
		},
		{
			name:         "valid preprocessing directive",
			contents:     "---\nname: test\n---\n!cat README.md",
			wantErrCount: 0,
		},
		{
			name:         "valid echo directive",
			contents:     "---\nname: test\n---\n!echo hello world",
			wantErrCount: 0,
		},
		{
			name:         "empty preprocessing directive",
			contents:     "---\nname: test\n---\n!",
			wantErrCount: 1,
			wantContains: []string{"Empty preprocessing directive"},
		},
		{
			name:         "empty preprocessing with whitespace only",
			contents:     "---\nname: test\n---\n!   ",
			wantErrCount: 1,
			wantContains: []string{"Empty preprocessing directive"},
		},
		{
			name:         "dangerous rm -rf /",
			contents:     "---\nname: test\n---\n!rm -rf /",
			wantErrCount: 1,
			wantContains: []string{"Dangerous preprocessing command"},
		},
		{
			name:         "dangerous rm -f -r /",
			contents:     "---\nname: test\n---\n!rm -f -r /",
			wantErrCount: 1,
			wantContains: []string{"Dangerous preprocessing command"},
		},
		{
			name:         "dangerous mkfs command",
			contents:     "---\nname: test\n---\n!mkfs.ext4 /dev/sda1",
			wantErrCount: 1,
			wantContains: []string{"Dangerous preprocessing command", "mkfs"},
		},
		{
			name:         "dangerous dd to device",
			contents:     "---\nname: test\n---\n!dd if=/dev/zero of=/dev/sda",
			wantErrCount: 1,
			wantContains: []string{"Dangerous preprocessing command", "dd"},
		},
		{
			name:         "safe rm on specific path",
			contents:     "---\nname: test\n---\n!rm -rf /tmp/mydir",
			wantErrCount: 0,
		},
		{
			name:         "directive inside code block is ignored",
			contents:     "---\nname: test\n---\n```bash\n!\n```",
			wantErrCount: 0,
		},
		{
			name:         "directive inside frontmatter is ignored",
			contents:     "---\nname: test\n!echo bad\n---\nBody here",
			wantErrCount: 0,
		},
		{
			name:         "multiple directives mixed valid and invalid",
			contents:     "---\nname: test\n---\n!cat file.txt\n!\n!echo hello",
			wantErrCount: 1,
			wantContains: []string{"Empty preprocessing directive"},
		},
		{
			name:         "dangerous chmod 777 /",
			contents:     "---\nname: test\n---\n!chmod 777 /",
			wantErrCount: 1,
			wantContains: []string{"Dangerous preprocessing command"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validateCommandPreprocessing("commands/test.md", tt.contents)

			errCount := 0
			for _, issue := range issues {
				if issue.Severity == "error" {
					errCount++
				}
			}

			if errCount != tt.wantErrCount {
				t.Errorf("validateCommandPreprocessing() errors = %d, want %d", errCount, tt.wantErrCount)
				for _, issue := range issues {
					t.Logf("  %s: %s (line %d)", issue.Severity, issue.Message, issue.Line)
				}
			}

			for _, want := range tt.wantContains {
				found := false
				for _, issue := range issues {
					if strings.Contains(issue.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateCommandPreprocessing() should contain message about %q", want)
					for _, issue := range issues {
						t.Logf("  Got: %s", issue.Message)
					}
				}
			}
		})
	}
}

func TestValidateCommandSubstitution(t *testing.T) {
	tests := []struct {
		name         string
		contents     string
		data         map[string]any
		wantWarnings int
		wantSuggs    int
		wantContains []string
	}{
		{
			name:         "no substitution variables",
			contents:     "---\nname: test\n---\nJust a normal body",
			data:         map[string]any{"name": "test"},
			wantWarnings: 0,
			wantSuggs:    0,
		},
		{
			name:         "$ARGUMENTS with argument-hint",
			contents:     "---\nname: test\nargument-hint: <query>\n---\nSearch for $ARGUMENTS",
			data:         map[string]any{"name": "test", "argument-hint": "<query>"},
			wantWarnings: 0,
			wantSuggs:    0,
		},
		{
			name:         "$ARGUMENTS without argument-hint",
			contents:     "---\nname: test\n---\nSearch for $ARGUMENTS",
			data:         map[string]any{"name": "test"},
			wantWarnings: 0,
			wantSuggs:    1,
			wantContains: []string{"argument-hint"},
		},
		{
			name:         "sequential $1 $2 $3 with hint",
			contents:     "---\nname: test\nargument-hint: <a> <b> <c>\n---\nUse $1 then $2 then $3",
			data:         map[string]any{"name": "test", "argument-hint": "<a> <b> <c>"},
			wantWarnings: 0,
			wantSuggs:    0,
		},
		{
			name:         "sequential $1 $2 without hint",
			contents:     "---\nname: test\n---\nUse $1 then $2",
			data:         map[string]any{"name": "test"},
			wantWarnings: 0,
			wantSuggs:    1,
			wantContains: []string{"argument-hint"},
		},
		{
			name:         "$2 without $1 (gap from start)",
			contents:     "---\nname: test\nargument-hint: <a> <b>\n---\nUse $2 here",
			data:         map[string]any{"name": "test", "argument-hint": "<a> <b>"},
			wantWarnings: 1,
			wantSuggs:    0,
			wantContains: []string{"$2 used without $1"},
		},
		{
			name:         "$1 and $3 without $2 (gap in middle)",
			contents:     "---\nname: test\nargument-hint: <a> <b> <c>\n---\nUse $1 and $3",
			data:         map[string]any{"name": "test", "argument-hint": "<a> <b> <c>"},
			wantWarnings: 1,
			wantSuggs:    0,
			wantContains: []string{"gap"},
		},
		{
			name:         "high positional arg $10",
			contents:     "---\nname: test\nargument-hint: many args\n---\nUse $1 $2 $3 $4 $5 $6 $7 $8 $9 $10",
			data:         map[string]any{"name": "test", "argument-hint": "many args"},
			wantWarnings: 1,
			wantContains: []string{"High positional argument $10"},
		},
		{
			name:         "high positional arg $15",
			contents:     "---\nname: test\nargument-hint: many args\n---\nUse $1 $2 $3 $4 $5 $6 $7 $8 $9 $10 $11 $12 $13 $14 $15",
			data:         map[string]any{"name": "test", "argument-hint": "many args"},
			wantWarnings: 1,
			wantContains: []string{"High positional argument $15"},
		},
		{
			name:         "$0 is ignored (not user positional)",
			contents:     "---\nname: test\nargument-hint: <a>\n---\nUse $0 and $1",
			data:         map[string]any{"name": "test", "argument-hint": "<a>"},
			wantWarnings: 0,
			wantSuggs:    0,
		},
		{
			name:         "mixed $ARGUMENTS and $1 with hint",
			contents:     "---\nname: test\nargument-hint: <query>\n---\nAll: $ARGUMENTS, first: $1",
			data:         map[string]any{"name": "test", "argument-hint": "<query>"},
			wantWarnings: 0,
			wantSuggs:    0,
		},
		{
			name:         "$5 without $1 (gap from start, high jump)",
			contents:     "---\nname: test\nargument-hint: <a>\n---\nUse $5 here",
			data:         map[string]any{"name": "test", "argument-hint": "<a>"},
			wantWarnings: 1,
			wantContains: []string{"$5 used without $1"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			issues := validateCommandSubstitution("commands/test.md", tt.contents, tt.data)

			warnCount := 0
			suggCount := 0
			for _, issue := range issues {
				switch issue.Severity {
				case "warning":
					warnCount++
				case "suggestion":
					suggCount++
				}
			}

			if warnCount != tt.wantWarnings {
				t.Errorf("validateCommandSubstitution() warnings = %d, want %d", warnCount, tt.wantWarnings)
				for _, issue := range issues {
					t.Logf("  %s: %s (line %d)", issue.Severity, issue.Message, issue.Line)
				}
			}
			if suggCount != tt.wantSuggs {
				t.Errorf("validateCommandSubstitution() suggestions = %d, want %d", suggCount, tt.wantSuggs)
				for _, issue := range issues {
					t.Logf("  %s: %s (line %d)", issue.Severity, issue.Message, issue.Line)
				}
			}

			for _, want := range tt.wantContains {
				found := false
				for _, issue := range issues {
					if strings.Contains(issue.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateCommandSubstitution() should contain message about %q", want)
					for _, issue := range issues {
						t.Logf("  Got: %s", issue.Message)
					}
				}
			}
		})
	}
}

func TestExtractBody(t *testing.T) {
	tests := []struct {
		name     string
		contents string
		want     string
	}{
		{
			name:     "with frontmatter",
			contents: "---\nname: test\n---\nBody content here",
			want:     "\nBody content here",
		},
		{
			name:     "without frontmatter",
			contents: "Just plain content",
			want:     "Just plain content",
		},
		{
			name:     "empty frontmatter",
			contents: "---\n---\nBody",
			want:     "\nBody",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractBody(tt.contents)
			if got != tt.want {
				t.Errorf("extractBody() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestCollectPositionalArgs(t *testing.T) {
	tests := []struct {
		name    string
		matches [][]string
		want    []int
	}{
		{
			name:    "empty matches",
			matches: nil,
			want:    nil,
		},
		{
			name:    "$1 and $2",
			matches: [][]string{{"$1", "1"}, {"$2", "2"}},
			want:    []int{1, 2},
		},
		{
			name:    "$0 is excluded",
			matches: [][]string{{"$0", "0"}, {"$1", "1"}},
			want:    []int{1},
		},
		{
			name:    "duplicates collapsed",
			matches: [][]string{{"$1", "1"}, {"$1", "1"}, {"$2", "2"}},
			want:    []int{1, 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectPositionalArgs(tt.matches)
			if len(got) != len(tt.want) {
				t.Errorf("collectPositionalArgs() len = %d, want %d", len(got), len(tt.want))
				return
			}
			// Sort both for comparison (collectPositionalArgs does not sort)
			gotMap := make(map[int]bool)
			for _, v := range got {
				gotMap[v] = true
			}
			for _, w := range tt.want {
				if !gotMap[w] {
					t.Errorf("collectPositionalArgs() missing %d, got %v", w, got)
				}
			}
		})
	}
}
