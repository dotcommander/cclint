package lint

import (
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

// TestRegistryCoversNamedComponentTypes pins the invariant that every named
// component type in the discovery vocabulary has a linter registry entry.
func TestRegistryCoversNamedComponentTypes(t *testing.T) {
	t.Parallel()
	covered := map[discovery.FileType]bool{}
	for _, entry := range linterRegistry {
		covered[entry.FileType] = true
	}
	for _, name := range discovery.ValidTypeNames() {
		ft, err := discovery.ParseFileType(name)
		if err != nil {
			t.Fatalf("ParseFileType(%q): %v", name, err)
		}
		if !covered[ft] {
			t.Errorf("named component type %q has no linter registry entry", name)
		}
	}
}

// TestRegistryLinterTypesMatchVocabulary pins each registry linter's Type()
// string to the discovery singular for its FileType, so the per-linter type
// strings cannot drift from the canonical name vocabulary.
func TestRegistryLinterTypesMatchVocabulary(t *testing.T) {
	t.Parallel()
	for _, entry := range linterRegistry {
		linter := entry.New("")
		if got, want := linter.Type(), entry.FileType.String(); got != want {
			t.Errorf("%s linter Type() = %q; discovery singular = %q", entry.Name, got, want)
		}
	}
}
