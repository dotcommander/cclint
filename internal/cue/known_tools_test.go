package cue

import (
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/textutil"
)

// quotedStringRe extracts a single quoted token's inner text. Used to parse the
// members of the generated #KnownTool definition line.
var quotedStringRe = regexp.MustCompile(`"([^"]+)"`)

// TestKnownToolUnionMatchesSource asserts that knownToolUnionCUE() — the single
// generator feeding the CUE #KnownTool union — reproduces exactly the union of
// textutil.KnownTools (minus the "*" wildcard) and schemaOnlyTools. Since the
// CUE schemas no longer hand-maintain #KnownTool (it is injected at load time
// from this generator), this is the drift guard: adding a tool to one side and
// not the other fails the build.
func TestKnownToolUnionMatchesSource(t *testing.T) {
	t.Parallel()

	// Expected: KnownTools minus "*", plus the schema-only tools.
	expected := make(map[string]struct{}, len(textutil.KnownTools)+len(schemaOnlyTools))
	for name := range textutil.KnownTools {
		if name == "*" {
			continue
		}
		expected[name] = struct{}{}
	}
	for _, name := range schemaOnlyTools {
		expected[name] = struct{}{}
	}

	// Actual: parse the quoted members out of the generated definition.
	got := make(map[string]struct{}, len(expected))
	for _, m := range quotedStringRe.FindAllStringSubmatch(knownToolUnionCUE(), -1) {
		got[m[1]] = struct{}{}
	}

	if len(got) == 0 {
		t.Fatal("knownToolUnionCUE() produced no members")
	}

	// Symmetric diff for a clear mismatch message.
	var missing, extra []string
	for name := range expected {
		if _, ok := got[name]; !ok {
			missing = append(missing, name)
		}
	}
	for name := range got {
		if _, ok := expected[name]; !ok {
			extra = append(extra, name)
		}
	}
	sort.Strings(missing)
	sort.Strings(extra)

	if len(missing) > 0 || len(extra) > 0 {
		var b strings.Builder
		if len(missing) > 0 {
			b.WriteString("missing from generated union: " + strings.Join(missing, ", ") + ". ")
		}
		if len(extra) > 0 {
			b.WriteString("unexpected in generated union: " + strings.Join(extra, ", ") + ". ")
		}
		t.Errorf("%supdate textutil.KnownTools or schemaOnlyTools — the CUE union is generated from them.", b.String())
	}
}

// TestKnownToolUnionExcludesWildcard asserts the "*" wildcard is never a union
// member (it is a sibling union option of #KnownTool, handled separately).
func TestKnownToolUnionExcludesWildcard(t *testing.T) {
	t.Parallel()

	union := knownToolUnionCUE()
	if !strings.Contains(union, `"Read"`) {
		t.Errorf("generated union missing a real tool: %s", union)
	}
	if strings.Contains(union, `"DBClient"`) == false {
		t.Errorf("generated union missing schema-only tool DBClient: %s", union)
	}
	if strings.Contains(union, `"*"`) {
		t.Errorf("generated union must not include the \"*\" wildcard: %s", union)
	}
}

// TestKnownToolUnionInjectedValidates exercises the end-to-end injection path:
// the generated #KnownTool union is wired into the loaded schemas at load time,
// so validation accepts a known tool and rejects an unknown one. This proves
// the injected union actually drives validation, not just the generator string.
func TestKnownToolUnionInjectedValidates(t *testing.T) {
	t.Parallel()

	v := NewValidator()
	if err := v.LoadSchemas(""); err != nil {
		t.Fatalf("LoadSchemas: %v", err)
	}

	// A real tool in the union must produce no validation errors.
	if errs, err := v.ValidateAgent(map[string]any{
		"name":        "test-agent",
		"description": "test",
		"tools":       []any{"Read"},
	}); err != nil {
		t.Fatalf("ValidateAgent(valid tool): unexpected error: %v", err)
	} else if len(errs) != 0 {
		t.Errorf("ValidateAgent(valid tool): expected no errors, got %d: %v", len(errs), errs)
	}

	// An unknown tool must produce a validation error.
	errs, err := v.ValidateAgent(map[string]any{
		"name":        "test-agent",
		"description": "test",
		"tools":       []any{"DefinitelyNotARealTool"},
	})
	if err != nil {
		t.Fatalf("ValidateAgent(unknown tool): unexpected error: %v", err)
	}
	if len(errs) == 0 {
		t.Errorf("ValidateAgent(unknown tool): expected a validation error for an unknown tool, got none")
	}
}
