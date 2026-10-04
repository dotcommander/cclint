package baseline

import (
	"github.com/dotcommander/cclint/internal/cue"
	"testing"
)

func TestCorrectedNormalizationAndLegacyIdentity(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{`invalid 'one', then 'two'.`, `invalid '*', then '*'.`},
		{`invalid "a\"b"`, `invalid "*"`},
		{`don't change unmatched 'text`, `don't change unmatched 'text`},
		{`version 1.2.3-beta-1+build-2 has 12 issues`, `version 1.2.3-beta-1+build-2 has N issues`},
	} {
		if got := normalizeMessage(tc.in); got != tc.want {
			t.Errorf("normalize %q: got %q want %q", tc.in, got, tc.want)
		}
	}
	issue := cue.ValidationError{File: "agents/a.md", Source: "schema", Message: `invalid 'x', 12 times`}
	fresh := CreateBaseline([]cue.ValidationError{issue})
	if fresh.Version != "1.1" || !fresh.IsKnown(issue) {
		t.Fatal("new baseline must match corrected fingerprint")
	}
	legacy := &Baseline{Version: "1.0", index: map[string]bool{legacyFingerprint(issue): true}}
	if !legacy.IsKnown(issue) {
		t.Fatal("same-file legacy fingerprint must match")
	}
	issue.File = "agents/b.md"
	if legacy.IsKnown(issue) || fresh.IsKnown(issue) {
		t.Fatal("different files must not match")
	}
	issue.File = ""
	if legacy.IsKnown(issue) {
		t.Fatal("must not fall back to blank file")
	}
}
