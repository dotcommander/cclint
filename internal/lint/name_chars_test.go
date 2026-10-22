package lint

import (
	"strings"
	"testing"
)

func TestNameCharacterProblem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "hyphenated ok", in: "my-skill", want: ""},
		{name: "regular space ok", in: "my skill", want: ""},
		{name: "digit suffix ok", in: "cmd-2", want: ""},
		{name: "hidden character", in: "bad\u200bname", want: "a hidden character"},
		{name: "direction-changing mark", in: "bad\u200fname", want: "a direction-changing character"},
		{name: "direction-changing override", in: "bad\u202ename", want: "a direction-changing character"},
		{name: "blank-like space", in: "bad\u00a0name", want: "a blank-like character"},
		{name: "less than", in: "a<b", want: "'<' or '>'"},
		{name: "greater than", in: "a>b", want: "'<' or '>'"},
		{name: "too long", in: strings.Repeat("a", 257), want: "more than 256 characters"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			if got := nameCharacterProblem(tt.in); got != tt.want {
				t.Errorf("nameCharacterProblem(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}
