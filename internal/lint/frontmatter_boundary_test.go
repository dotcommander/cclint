package lint

import (
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
)

func TestComponentMalformedFrontmatterStopsPipeline(t *testing.T) {
	t.Parallel()
	components := []struct {
		name, path string
		linter     ComponentLinter
	}{
		{"agent", "agents/test.md", NewAgentLinter()},
		{"command", "commands/test.md", NewCommandLinter()},
		{"skill", "skills/test/SKILL.md", NewSkillLinter()},
		{"rule", "rules/test.md", NewRuleLinter()},
		{"context", "CLAUDE.md", NewContextLinter()},
		{"output style", "output-styles/test.md", NewOutputStyleLinter()},
	}
	for _, component := range components {
		t.Run(component.name, func(t *testing.T) {
			t.Parallel()
			for _, malformed := range []struct{ name, content string }{
				{"unclosed", "---\nname: test\n"},
				{"invalid YAML", "---\nname: [\n---\nBody"},
			} {
				t.Run(malformed.name, func(t *testing.T) {
					t.Parallel()
					result := lintFileCore(component.path, malformed.content, component.linter, nil, nil)
					if result.Success || result.Quality != nil || len(result.Improvements) != 0 || len(result.Warnings) != 0 || len(result.Suggestions) != 0 {
						t.Fatalf("malformed YAML reached downstream work: %#v", result)
					}
					if len(result.Errors) != 1 {
						t.Fatalf("expected only the parse error, got %#v", result.Errors)
					}
					got := result.Errors[0]
					if got.File != component.path || got.Severity != cue.SeverityError || got.Message == "" {
						t.Errorf("parse error was not file-scoped: %#v", got)
					}
				})
			}
		})
	}
}

func TestOptionalComponentPlainMarkdownRemainsValid(t *testing.T) {
	t.Parallel()
	for _, component := range []struct {
		name   string
		linter ComponentLinter
	}{
		{"skill", NewSkillLinter()},
		{"rule", NewRuleLinter()},
		{"context", NewContextLinter()},
		{"output style", NewOutputStyleLinter()},
	} {
		t.Run(component.name, func(t *testing.T) {
			t.Parallel()
			input := "Plain Markdown\n"
			data, body, err := component.linter.ParseContent(input)
			if err != nil || body != input || data == nil {
				t.Errorf("plain Markdown rejected: data %#v, body %q, error %v", data, body, err)
			}
		})
	}
}
