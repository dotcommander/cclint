package cue

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/textutil"
)

// TestKnownToolDriftVsGo fails when the Go textutil.KnownTools set and the
// CUE #KnownTool union (defined byte-identically in agent.cue/command.cue/
// skill.cue) drift apart beyond their documented intentional differences.
//
// Documented exceptions (must NOT fail the test):
//   - Go-only: "*" — the wildcard is a sibling union option in CUE
//     (`#Tools: "*" | string | [...#KnownTool]`), never a #KnownTool member.
//   - CUE-only: "BashOutput", "KillBash", "DBClient" — defined in the CUE
//     schema but not referenced by any Go-side validator.
//
// Any other drift fails: update both sets together, or add to the documented
// exception lists below if the asymmetry is intentional.
func TestKnownToolDriftVsGo(t *testing.T) {
	t.Parallel()

	// 1. Build the CUE #KnownTool member set from agent.cue's embedded schema.
	// schemaFS is the embed.FS declared in validator.go (//go:embed schemas/*.cue).
	cueBytes, err := schemaFS.ReadFile("schemas/agent.cue")
	if err != nil {
		t.Fatalf("read embedded schemas/agent.cue: %v", err)
	}

	// The #KnownTool union spans multiple lines and is terminated by a blank
	// line. Extract its body, then pull every quoted member literal.
	blockRe := regexp.MustCompile(`(?ms)^#KnownTool:\s*(.*?)(?:\n\s*\n|\z)`)
	quotedRe := regexp.MustCompile(`"([A-Za-z]+)"`)

	blockMatch := blockRe.FindSubmatch(cueBytes)
	if blockMatch == nil {
		t.Fatal("could not locate #KnownTool union in schemas/agent.cue")
	}
	cueSet := make(map[string]struct{})
	for _, m := range quotedRe.FindAllSubmatch(blockMatch[1], -1) {
		cueSet[string(m[1])] = struct{}{}
	}
	if len(cueSet) == 0 {
		t.Fatal("extracted CUE #KnownTool union has no members")
	}

	// 2. Build the Go set from textutil.KnownTools.
	goSet := make(map[string]struct{}, len(textutil.KnownTools))
	for name := range textutil.KnownTools {
		goSet[name] = struct{}{}
	}

	// 3. Documented intentional differences.
	goOnly := map[string]struct{}{"*": {}}
	cueOnly := map[string]struct{}{
		"BashOutput": {},
		"KillBash":   {},
		"DBClient":   {},
	}

	// 4. (Go \ CUE) minus goOnly and (CUE \ Go) minus cueOnly must both be empty.
	var goNotCue, cueNotGo []string
	for name := range goSet {
		if _, ok := cueSet[name]; ok {
			continue
		}
		if _, allowed := goOnly[name]; allowed {
			continue
		}
		goNotCue = append(goNotCue, name)
	}
	for name := range cueSet {
		if _, ok := goSet[name]; ok {
			continue
		}
		if _, allowed := cueOnly[name]; allowed {
			continue
		}
		cueNotGo = append(cueNotGo, name)
	}
	sort.Strings(goNotCue)
	sort.Strings(cueNotGo)

	hint := "hint: update both sets together, or add to the documented exception list in this test if intentional."
	if len(goNotCue) > 0 {
		t.Errorf("KnownTools has tools absent from CUE #KnownTool (outside documented goOnly exceptions): %s\n%s",
			strings.Join(goNotCue, ", "), hint)
	}
	if len(cueNotGo) > 0 {
		t.Errorf("CUE #KnownTool has tools absent from KnownTools (outside documented cueOnly exceptions): %s\n%s",
			strings.Join(cueNotGo, ", "), hint)
	}
}
