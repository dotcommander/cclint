package cli

import (
	"strings"
	"testing"
)

func TestValidateSkillBestPractices(t *testing.T) {
	tests := []struct {
		name         string
		filePath     string
		contents     string
		fmData       map[string]any
		wantContains []string
	}{
		{
			name:     "first person description",
			filePath: "skills/test/SKILL.md",
			contents: "---\ndescription: I analyze code\n---\n",
			fmData: map[string]any{
				"description": "I analyze code",
			},
			wantContains: []string{"third person"},
		},
		{
			name:     "addresses user",
			filePath: "skills/test/SKILL.md",
			contents: "---\ndescription: You should use this\n---\n",
			fmData: map[string]any{
				"description": "You should use this",
			},
			wantContains: []string{"address the user"},
		},
		{
			name:     "short description",
			filePath: "skills/test/SKILL.md",
			contents: "---\ndescription: Short\n---\n",
			fmData: map[string]any{
				"description": "Short",
			},
			wantContains: []string{"50+"},
		},
		{
			name:     "missing trigger phrases",
			filePath: "skills/test/SKILL.md",
			contents: "---\ndescription: This is a skill that does things without triggers\n---\n",
			fmData: map[string]any{
				"description": "This is a skill that does things without triggers",
			},
			wantContains: []string{"trigger phrases"},
		},
		{
			name:     "invalid semver",
			filePath: "skills/test/SKILL.md",
			contents: "---\nversion: 1.0\n---\n",
			fmData: map[string]any{
				"version": "1.0",
			},
			wantContains: []string{"semver"},
		},
		{
			name:         "missing anti-patterns",
			filePath:     "skills/test/SKILL.md",
			contents:     "---\nname: test\n---\nContent without anti-patterns",
			fmData:       map[string]any{"name": "test"},
			wantContains: []string{"Anti-Patterns"},
		},
		{
			name:         "missing examples",
			filePath:     "skills/test/SKILL.md",
			contents:     "---\nname: test\n---\nContent without examples",
			fmData:       map[string]any{"name": "test"},
			wantContains: []string{"Examples"},
		},
		{
			name:     "argument-hint too long",
			filePath: "skills/test/SKILL.md",
			contents: "---\nargument-hint: " + strings.Repeat("a", 81) + "\n---\n## Anti-Patterns\n## Examples\n",
			fmData: map[string]any{
				"argument-hint": strings.Repeat("a", 81),
			},
			wantContains: []string{"argument-hint", "80"},
		},
		{
			name:     "argument-hint at limit no warning",
			filePath: "skills/test/SKILL.md",
			contents: "---\nargument-hint: " + strings.Repeat("a", 80) + "\n---\n## Anti-Patterns\n## Examples\n",
			fmData: map[string]any{
				"argument-hint": strings.Repeat("a", 80),
			},
			wantContains: []string{}, // no argument-hint warning expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			suggestions := validateSkillBestPractices(tt.filePath, tt.contents, tt.fmData)

			for _, want := range tt.wantContains {
				found := false
				for _, sugg := range suggestions {
					if strings.Contains(sugg.Message, want) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("validateSkillBestPractices() should contain suggestion about %q", want)
					for _, s := range suggestions {
						t.Logf("  Got: %s", s.Message)
					}
				}
			}
		})
	}
}

func TestExtractSkillName(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		filePath string
		want     string
	}{
		{
			name:     "extract from heading",
			content:  "# Test\n\nContent",
			filePath: "skills/test-skill/SKILL.md",
			want:     "test",
		},
		{
			name:     "strip Agent prefix",
			content:  "# Agent Test\n\nContent",
			filePath: "skills/test/SKILL.md",
			want:     "test",
		},
		{
			name:     "strip Command prefix",
			content:  "# Command Test\n\nContent",
			filePath: "skills/test/SKILL.md",
			want:     "test",
		},
		{
			name:     "strip Skill prefix",
			content:  "# Skill Test\n\nContent",
			filePath: "skills/test/SKILL.md",
			want:     "test",
		},
		{
			name:     "strip Patterns suffix",
			content:  "# Test Patterns\n\nContent",
			filePath: "skills/test/SKILL.md",
			want:     "test",
		},
		{
			name:     "fallback to filename",
			content:  "No heading",
			filePath: "skills/fallback-name/SKILL.md",
			want:     "SKILL.md",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractSkillName(tt.content, tt.filePath)
			if got != tt.want {
				t.Errorf("extractSkillName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestKnownSkillFields(t *testing.T) {
	// Core fields from Anthropic docs
	expected := []string{
		"name", "description", "argument-hint", "allowed-tools", "model",
		"context", "agent", "user-invocable", "hooks",
	}
	// agentskills.io spec fields
	expected = append(expected, "license", "compatibility", "metadata")
	// Legacy fields
	expected = append(expected, "version")

	for _, field := range expected {
		if !knownSkillFields[field] {
			t.Errorf("knownSkillFields missing expected field: %s", field)
		}
	}
}

func TestSkillLinterPreValidate(t *testing.T) {
	linter := NewSkillLinter()

	tests := []struct {
		name         string
		filePath     string
		contents     string
		wantErrCount int
	}{
		{
			name:         "valid SKILL.md",
			filePath:     "skills/test/SKILL.md",
			contents:     "content",
			wantErrCount: 0,
		},
		{
			name:         "wrong filename",
			filePath:     "skills/test/readme.md",
			contents:     "content",
			wantErrCount: 1,
		},
		{
			name:         "empty file",
			filePath:     "skills/test/SKILL.md",
			contents:     "   \n  ",
			wantErrCount: 1,
		},
		{
			name:         "windows path SKILL.md",
			filePath:     "skills\\test\\SKILL.md",
			contents:     "content",
			wantErrCount: 0,
		},
		{
			name:         "just SKILL.md",
			filePath:     "SKILL.md",
			contents:     "content",
			wantErrCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := linter.PreValidate(tt.filePath, tt.contents)

			errCount := 0
			for _, e := range errors {
				if e.Severity == "error" {
					errCount++
				}
			}

			if errCount != tt.wantErrCount {
				t.Errorf("PreValidate() errors = %d, want %d", errCount, tt.wantErrCount)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestSkillLinterValidateSpecific(t *testing.T) {
	linter := NewSkillLinter()

	tests := []struct {
		name             string
		data             map[string]any
		contents         string
		wantErrCount     int
		wantWarnings     int
		wantSuggestions  int
	}{
		{
			name: "unknown frontmatter field",
			data: map[string]any{
				"unknown-field": "value",
			},
			contents:        "---\nunknown-field: value\n---\n",
			wantSuggestions: 1,
		},
		{
			name: "reserved word name",
			data: map[string]any{
				"name": "claude",
			},
			contents:     "---\nname: claude\n---\n",
			wantErrCount: 2, // reserved word + name-directory mismatch
		},
		{
			name: "reserved word anthropic",
			data: map[string]any{
				"name": "anthropic",
			},
			contents:     "---\nname: anthropic\n---\n",
			wantErrCount: 2, // reserved word + name-directory mismatch
		},
		{
			name:            "no frontmatter suggestion",
			data:            map[string]any{},
			contents:        "Just content without frontmatter",
			wantSuggestions: 1,
		},
		{
			name: "with frontmatter",
			data: map[string]any{
				"name": "test",
			},
			contents:     "---\nname: test\n---\nContent",
			wantErrCount: 0,
		},
		{
			name: "multiple unknown fields",
			data: map[string]any{
				"unknown1": "val1",
				"unknown2": "val2",
			},
			contents:        "---\nunknown1: val1\nunknown2: val2\n---\n",
			wantSuggestions: 2,
		},
		{
			name: "valid context fork",
			data: map[string]any{
				"name":    "test",
				"context": "fork",
			},
			contents:     "---\nname: test\ncontext: fork\n---\nContent",
			wantErrCount: 0,
		},
		{
			name: "invalid context value",
			data: map[string]any{
				"name":    "test",
				"context": "invalid",
			},
			contents:     "---\nname: test\ncontext: invalid\n---\nContent",
			wantErrCount: 1,
		},
		{
			name: "agent without context fork warns",
			data: map[string]any{
				"name":  "test",
				"agent": "poet-agent",
			},
			contents:     "---\nname: test\nagent: poet-agent\n---\nContent",
			wantErrCount: 0,
			wantWarnings: 1,
		},
		{
			name: "agent with context fork no warning",
			data: map[string]any{
				"name":    "test",
				"context": "fork",
				"agent":   "poet-agent",
			},
			contents:     "---\nname: test\ncontext: fork\nagent: poet-agent\n---\nContent",
			wantErrCount: 0,
			wantWarnings: 0,
		},
		{
			name: "valid user-invocable true",
			data: map[string]any{
				"name":           "test",
				"user-invocable": true,
			},
			contents:     "---\nname: test\nuser-invocable: true\n---\nContent",
			wantErrCount: 0,
		},
		{
			name: "valid user-invocable false",
			data: map[string]any{
				"name":           "test",
				"user-invocable": false,
			},
			contents:     "---\nname: test\nuser-invocable: false\n---\nContent",
			wantErrCount: 0,
		},
		{
			name: "invalid user-invocable type string",
			data: map[string]any{
				"name":           "test",
				"user-invocable": "yes",
			},
			contents:     "---\nname: test\nuser-invocable: \"yes\"\n---\nContent",
			wantErrCount: 1,
		},
		{
			name: "valid disable-model-invocation",
			data: map[string]any{
				"name":                     "test",
				"disable-model-invocation": true,
			},
			contents:     "---\nname: test\ndisable-model-invocation: true\n---\nContent",
			wantErrCount: 0,
		},
		{
			name: "invalid disable-model-invocation type string",
			data: map[string]any{
				"name":                     "test",
				"disable-model-invocation": "true",
			},
			contents:     "---\nname: test\ndisable-model-invocation: \"true\"\n---\nContent",
			wantErrCount: 1,
		},
		{
			name: "valid argument-hint",
			data: map[string]any{
				"name":          "test",
				"argument-hint": "PR number or URL",
			},
			contents:     "---\nname: test\nargument-hint: PR number or URL\n---\nContent",
			wantErrCount: 0,
			wantWarnings: 0,
		},
		{
			name: "name matches parent directory - no error",
			data: map[string]any{
				"name": "test",
			},
			contents:     "---\nname: test\n---\nContent",
			wantErrCount: 0,
		},
		{
			name: "argument-hint non-string type",
			data: map[string]any{
				"name":          "test",
				"argument-hint": 42,
			},
			contents:     "---\nname: test\nargument-hint: 42\n---\nContent",
			wantErrCount: 1,
		},
		{
			name: "argument-hint empty string",
			data: map[string]any{
				"name":          "test",
				"argument-hint": "",
			},
			contents:     "---\nname: test\nargument-hint: \"\"\n---\nContent",
			wantErrCount: 0,
			wantWarnings: 1,
		},
		{
			name: "argument-hint whitespace only",
			data: map[string]any{
				"name":          "test",
				"argument-hint": "   ",
			},
			contents:     "---\nname: test\nargument-hint: \"   \"\n---\nContent",
			wantErrCount: 0,
			wantWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := linter.ValidateSpecific(tt.data, "skills/test/SKILL.md", tt.contents)

			errCount := 0
			warnCount := 0
			suggCount := 0
			for _, e := range errors {
				switch e.Severity {
				case "error":
					errCount++
				case "warning":
					warnCount++
				case "suggestion":
					suggCount++
				}
			}

			if errCount != tt.wantErrCount {
				t.Errorf("ValidateSpecific() errors = %d, want %d", errCount, tt.wantErrCount)
				for _, e := range errors {
					if e.Severity == "error" {
						t.Logf("  Error: %s", e.Message)
					}
				}
			}

			if warnCount != tt.wantWarnings {
				t.Errorf("ValidateSpecific() warnings = %d, want %d", warnCount, tt.wantWarnings)
				for _, e := range errors {
					if e.Severity == "warning" {
						t.Logf("  Warning: %s", e.Message)
					}
				}
			}

			if suggCount < tt.wantSuggestions {
				t.Errorf("ValidateSpecific() suggestions = %d, want at least %d", suggCount, tt.wantSuggestions)
			}
		})
	}
}

func TestValidateSkillNameDirectoryMatch(t *testing.T) {
	tests := []struct {
		name         string
		skillName    string
		filePath     string
		wantErrCount int
		wantMsg      string
	}{
		{
			name:         "name matches parent directory",
			skillName:    "my-skill",
			filePath:     "skills/my-skill/SKILL.md",
			wantErrCount: 0,
		},
		{
			name:         "name does not match parent directory",
			skillName:    "other-name",
			filePath:     "skills/my-skill/SKILL.md",
			wantErrCount: 1,
			wantMsg:      "agentskills.io spec",
		},
		{
			name:         "parent is skills directory - skip check",
			skillName:    "anything",
			filePath:     "skills/SKILL.md",
			wantErrCount: 0,
		},
		{
			name:         "parent is .claude directory - skip check",
			skillName:    "anything",
			filePath:     ".claude/SKILL.md",
			wantErrCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents := "---\nname: " + tt.skillName + "\n---\n"
			errs := validateSkillName(tt.skillName, tt.filePath, contents)

			errCount := 0
			for _, e := range errs {
				if e.Severity == "error" {
					errCount++
					if tt.wantMsg != "" && !strings.Contains(e.Message, tt.wantMsg) {
						t.Errorf("error message %q does not contain %q", e.Message, tt.wantMsg)
					}
				}
			}

			if errCount != tt.wantErrCount {
				t.Errorf("validateSkillName() errors = %d, want %d", errCount, tt.wantErrCount)
				for _, e := range errs {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestSkillLinterType(t *testing.T) {
	linter := NewSkillLinter()
	if linter.Type() != "skill" {
		t.Errorf("SkillLinter.Type() = %q, want %q", linter.Type(), "skill")
	}
}

func TestSkillLinterParseContent(t *testing.T) {
	linter := NewSkillLinter()

	tests := []struct {
		name        string
		contents    string
		wantDataNil bool
	}{
		{
			name:        "with frontmatter",
			contents:    "---\nname: test\n---\nBody",
			wantDataNil: false,
		},
		{
			name:        "without frontmatter",
			contents:    "Just content",
			wantDataNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, body, err := linter.ParseContent(tt.contents)

			if err != nil {
				t.Errorf("ParseContent() unexpected error: %v", err)
			}

			if (data == nil) != tt.wantDataNil {
				t.Errorf("ParseContent() data nil = %v, want %v", data == nil, tt.wantDataNil)
			}

			if body == "" {
				t.Error("ParseContent() body is empty")
			}
		})
	}
}
