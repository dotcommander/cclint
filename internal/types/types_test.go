package types_test

import (
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/types"
)

// TestComponentTypeConstantsMatchDiscoveryVocabulary pins the shared string
// constants (aliased by internal/cue and compared across lint, crossfile, and
// textutil) to discovery's canonical component-name vocabulary.
func TestComponentTypeConstantsMatchDiscoveryVocabulary(t *testing.T) {
	t.Parallel()

	for constant, want := range map[string]string{
		types.TypeAgent:   discovery.FileTypeAgent.String(),
		types.TypeCommand: discovery.FileTypeCommand.String(),
		types.TypeSkill:   discovery.FileTypeSkill.String(),
		types.TypeRule:    discovery.FileTypeRule.String(),
	} {
		if constant != want {
			t.Errorf("component type constant %q does not match discovery vocabulary %q", constant, want)
		}
	}
}
