package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

// TestLintTypeEnumMatchesDiscoveryVocabulary guards the kong --type enum tag
// (a struct-tag literal that cannot be derived at runtime) against the
// discovery name vocabulary that ParseFileType accepts: the CLI must not
// offer or reject a type the rest of the system names differently.
func TestLintTypeEnumMatchesDiscoveryVocabulary(t *testing.T) {
	t.Parallel()
	field, ok := reflect.TypeOf(lintCommand{}).FieldByName("Type")
	if !ok {
		t.Fatal("lintCommand.Type field not found")
	}
	want := strings.Join(discovery.ValidTypeNames(), ",")
	if got := field.Tag.Get("enum"); got != want {
		t.Fatalf("--type enum tag = %q; discovery vocabulary = %q", got, want)
	}
}
