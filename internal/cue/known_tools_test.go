package cue

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/textutil"
)

// intentionalDivergence records #KnownTool union members and textutil.KnownTools
// keys that are deliberately not mirrored. Adjust this map only with a paired
// reason and a code comment explaining why the symmetry break is correct.
//
//	cueOnly: tool appears in CUE schema #KnownTool but not in Go map.
//	goOnly:  tool appears in Go map but not in CUE #KnownTool.
//
// Everything else must be in sync. A new tool added to one side and missing
// from the other (without an entry here) fails the test.
var intentionalDivergence = struct {
	cueOnly map[string]string
	goOnly  map[string]string
}{
	cueOnly: map[string]string{
		// Documented in Anthropic's tool list but not yet referenced by any
		// Go-side validator. Keep CUE-only until a Go consumer needs them.
		"BashOutput": "CUE-only: legacy/alternate Bash output tool name",
		"KillBash":   "CUE-only: legacy/alternate Bash kill tool name",
		"DBClient":   "CUE-only: database client tool not validated by Go",
	},
	goOnly: map[string]string{
		// The "*" wildcard is a sibling option of #KnownTool in CUE's
		// `allowed-tools?: "*" | string | [...#KnownTool]`, not a union
		// member. Go's flat map conflates both; the wildcard must remain
		// recognized by ValidateAllowedTools.
		"*": "Go-only: wildcard handled as separate CUE union option",
	},
}

// TestKnownToolsMatchCUESchema asserts that the CUE schema #KnownTool union
// and textutil.KnownTools agree, modulo intentionalDivergence. Removing a
// member from either side (without updating the divergence map) fails the
// build, preventing silent drift.
func TestKnownToolsMatchCUESchema(t *testing.T) {
	t.Parallel()

	cueTools, err := extractKnownToolUnionFromSchemas()
	if err != nil {
		t.Fatalf("extractKnownToolUnionFromSchemas: %v", err)
	}
	if len(cueTools) == 0 {
		t.Fatal("extracted CUE #KnownTool union is empty")
	}

	goTools := make(map[string]struct{}, len(textutil.KnownTools))
	for name := range textutil.KnownTools {
		goTools[name] = struct{}{}
	}

	// cueOnly: in CUE but not in Go.
	var unexpectedCueOnly []string
	for name := range cueTools {
		if _, ok := goTools[name]; ok {
			continue
		}
		if _, allowed := intentionalDivergence.cueOnly[name]; allowed {
			continue
		}
		unexpectedCueOnly = append(unexpectedCueOnly, name)
	}

	// goOnly: in Go but not in CUE.
	var unexpectedGoOnly []string
	for name := range goTools {
		if _, ok := cueTools[name]; ok {
			continue
		}
		if _, allowed := intentionalDivergence.goOnly[name]; allowed {
			continue
		}
		unexpectedGoOnly = append(unexpectedGoOnly, name)
	}

	sort.Strings(unexpectedCueOnly)
	sort.Strings(unexpectedGoOnly)

	if len(unexpectedCueOnly) > 0 {
		t.Errorf("tools present in CUE #KnownTool but missing from textutil.KnownTools (add to Go map or to intentionalDivergence.cueOnly): %v", unexpectedCueOnly)
	}
	if len(unexpectedGoOnly) > 0 {
		t.Errorf("tools present in textutil.KnownTools but missing from CUE #KnownTool (add to CUE schemas or to intentionalDivergence.goOnly): %v", unexpectedGoOnly)
	}
}

// knownToolBlockRe extracts the body of `#KnownTool: "A" | "B" | ...` blocks.
// The union may span multiple lines, terminated by a blank line, a closing
// brace, or any line that does not continue the alternation.
var knownToolBlockRe = regexp.MustCompile(`(?ms)^#KnownTool:\s*(.*?)(?:\n\s*\n|\z)`)
var quotedStringRe = regexp.MustCompile(`"([^"]+)"`)

// extractKnownToolUnionFromSchemas reads every embedded *.cue file, locates
// the `#KnownTool` union definition (if any), and returns the union of all
// quoted alternatives across files. Multiple identical definitions are merged.
func extractKnownToolUnionFromSchemas() (map[string]struct{}, error) {
	entries, err := schemaFS.ReadDir("schemas")
	if err != nil {
		return nil, err
	}

	out := make(map[string]struct{})
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".cue") {
			continue
		}
		data, err := schemaFS.ReadFile("schemas/" + e.Name())
		if err != nil {
			return nil, err
		}
		match := knownToolBlockRe.FindSubmatch(data)
		if match == nil {
			continue
		}
		for _, m := range quotedStringRe.FindAllSubmatch(match[1], -1) {
			out[string(m[1])] = struct{}{}
		}
	}
	return out, nil
}
