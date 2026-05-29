package lint

import (
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
)

// Guards the consolidated checkUnknownFields path (internal/lint/unknown_fields.go) — all 5 components must keep their exact message + severity.

// findUnknownErr returns the first error whose message begins with "Unknown ".
func findUnknownErr(errs []cue.ValidationError) (cue.ValidationError, bool) {
	for _, e := range errs {
		if strings.HasPrefix(e.Message, "Unknown ") {
			return e, true
		}
	}
	return cue.ValidationError{}, false
}

func TestUnknownFieldMessagesPreserved(t *testing.T) {
	t.Parallel()

	const filePath = "test.md"
	data := func() map[string]any { return map[string]any{"zzz": true} }

	cases := []struct {
		name string
		// run exercises the component's real unknown-field path.
		run func() []cue.ValidationError
		// check asserts the message shape for this component.
		check func(t *testing.T, msg string)
	}{
		{
			name: "agent",
			run:  func() []cue.ValidationError { return validateUnknownFields(data(), filePath, "") },
			check: func(t *testing.T, msg string) {
				const want = "Unknown frontmatter field 'zzz'. Valid fields: "
				if !strings.HasPrefix(msg, want) {
					t.Errorf("agent message = %q, want prefix %q", msg, want)
				}
			},
		},
		{
			name: "command",
			run:  func() []cue.ValidationError { return validateCommandSpecific(data(), filePath, "") },
			check: func(t *testing.T, msg string) {
				const want = "Unknown frontmatter field 'zzz'. Valid fields: "
				if !strings.HasPrefix(msg, want) {
					t.Errorf("command message = %q, want prefix %q", msg, want)
				}
			},
		},
		{
			name: "output-style",
			run:  func() []cue.ValidationError { return NewOutputStyleLinter().ValidateSpecific(data(), filePath, "") },
			check: func(t *testing.T, msg string) {
				const want = "Unknown frontmatter field 'zzz'. Valid fields: "
				if !strings.HasPrefix(msg, want) {
					t.Errorf("output-style message = %q, want prefix %q", msg, want)
				}
			},
		},
		{
			name: "skill",
			run:  func() []cue.ValidationError { return checkUnknownSkillFields(data(), filePath, "") },
			check: func(t *testing.T, msg string) {
				const want = "Unknown frontmatter field 'zzz'. See https://agentskills.io/specification for valid fields"
				if msg != want {
					t.Errorf("skill message = %q, want %q", msg, want)
				}
			},
		},
		{
			name: "plugin",
			run:  func() []cue.ValidationError { return validateUnknownPluginFields(data(), filePath, "") },
			check: func(t *testing.T, msg string) {
				const want = "Unknown plugin field 'zzz'"
				if msg != want {
					t.Errorf("plugin message = %q, want %q", msg, want)
				}
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			errs := tc.run()
			e, ok := findUnknownErr(errs)
			if !ok {
				t.Fatalf("%s: no error with prefix %q in %d returned errors", tc.name, "Unknown ", len(errs))
			}
			tc.check(t, e.Message)
			if e.Severity != "suggestion" {
				t.Errorf("%s: Severity = %q, want %q", tc.name, e.Severity, "suggestion")
			}
		})
	}
}
