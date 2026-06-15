package cue

import (
	"sort"
	"strings"

	"github.com/dotcommander/cclint/internal/textutil"
)

// schemaOnlyTools are accepted by the CUE #KnownTool union but intentionally
// absent from textutil.KnownTools (Go's body-scan set). They are the documented
// CUE-only exceptions; keep this list and KnownTools as the two inputs to the
// generated union.
var schemaOnlyTools = []string{"BashOutput", "KillBash", "DBClient"}

// knownToolUnionCUE generates the CUE #KnownTool definition from the single
// source of truth: textutil.KnownTools (minus the "*" wildcard, which is a
// sibling union option, not a tool member) plus schemaOnlyTools. Members are
// deduped and sorted alphabetically for deterministic output. The returned
// string is a single CUE definition line, e.g.:
//
//	#KnownTool: "A" | "B" | "C"
//
// It is injected into agent/command/skill schemas at load time (see
// validator.go) so those schemas reference #KnownTool without hand-maintaining
// the tool list.
func knownToolUnionCUE() string {
	seen := make(map[string]struct{}, len(textutil.KnownTools)+len(schemaOnlyTools))
	for name := range textutil.KnownTools {
		if name == "*" {
			continue
		}
		seen[name] = struct{}{}
	}
	for _, name := range schemaOnlyTools {
		seen[name] = struct{}{}
	}

	members := make([]string, 0, len(seen))
	for name := range seen {
		members = append(members, `"`+name+`"`)
	}
	sort.Strings(members)

	return "#KnownTool: " + strings.Join(members, " | ")
}
