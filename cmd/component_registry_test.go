package cmd

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/format"
	"github.com/dotcommander/cclint/internal/lint"
)

// Observe real owners rather than copying their current registration lists.
func componentRegistrySnapshot(t *testing.T) ([]discovery.FileType, map[string][]discovery.FileType) {
	t.Helper()
	registries := map[string][]discovery.FileType{}
	var vocabulary []discovery.FileType
	for _, name := range discovery.ValidTypeNames() {
		ft, err := discovery.ParseFileType(name)
		if err != nil {
			t.Fatal(err)
		}
		if ft == discovery.FileTypeUnknown || ft == discovery.FileTypeConfig {
			t.Fatalf("internal classification %d has a user-facing name", ft)
		}
		vocabulary = append(vocabulary, ft)
		if entry, ok := lint.LinterForType(ft); ok {
			if entry.New == nil {
				t.Fatalf("%s linter has no factory", name)
			}
			registries["lint"] = append(registries["lint"], ft)
		}
		if format.CanFormatComponent(name) {
			registries["format"] = append(registries["format"], ft)
		}
	}
	for _, entry := range discovery.DefaultFileTypes {
		registries["discovery"] = append(registries["discovery"], entry.Type)
	}
	for _, entry := range lint.DefaultLinters() {
		registries["default lint"] = append(registries["default lint"], entry.FileType)
	}
	entries, err := summaryLinterEntries()
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		registries["summary"] = append(registries["summary"], entry.FileType)
	}
	field, ok := reflect.TypeOf(lintCommand{}).FieldByName("Type")
	if !ok {
		t.Fatal("lintCommand.Type missing")
	}
	for _, name := range strings.Split(field.Tag.Get("enum"), ",") {
		ft, err := discovery.ParseFileType(name)
		if err != nil {
			t.Fatal(err)
		}
		registries["CLI"] = append(registries["CLI"], ft)
	}
	return vocabulary, registries
}

// These are intentional participation exceptions, not a second type catalog.
// New named types require bindings or an explicit decision to exclude them.
func componentRegistryIncludes(registry string, ft discovery.FileType) bool {
	switch registry {
	case "default lint":
		return ft != discovery.FileTypeContext
	case "format":
		return ft != discovery.FileTypeOutputStyle
	case "summary":
		switch ft {
		case discovery.FileTypeSettings, discovery.FileTypeContext, discovery.FileTypePlugin,
			discovery.FileTypeRule, discovery.FileTypeOutputStyle:
			return false
		}
	}
	return true
}

func componentRegistryProblems(vocabulary []discovery.FileType, registries map[string][]discovery.FileType) error {
	var problems error
	for _, name := range []string{"discovery", "lint", "default lint", "format", "CLI", "summary"} {
		counts := map[discovery.FileType]int{}
		for _, ft := range registries[name] {
			counts[ft]++
			if !slices.Contains(vocabulary, ft) {
				problems = errors.Join(problems, fmt.Errorf("%s: unexpected type %d", name, ft))
			}
		}
		for _, ft := range vocabulary {
			count := counts[ft]
			switch {
			case componentRegistryIncludes(name, ft) && count == 0:
				problems = errors.Join(problems, fmt.Errorf("%s: missing %s", name, ft))
			case !componentRegistryIncludes(name, ft) && count != 0:
				problems = errors.Join(problems, fmt.Errorf("%s: unexpected %s", name, ft))
			case count > 1:
				problems = errors.Join(problems, fmt.Errorf("%s: duplicate %s", name, ft))
			}
		}
	}
	return problems
}

func TestComponentRegistriesMatchVocabulary(t *testing.T) {
	t.Parallel()
	vocabulary, registries := componentRegistrySnapshot(t)
	if err := componentRegistryProblems(vocabulary, registries); err != nil {
		t.Fatal(err)
	}
}

func TestComponentRegistriesRejectMissingBindings(t *testing.T) {
	t.Parallel()
	vocabulary, registries := componentRegistrySnapshot(t)
	for name := range registries {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			fixture := maps.Clone(registries)
			fixture[name] = slices.DeleteFunc(slices.Clone(registries[name]), func(ft discovery.FileType) bool {
				return ft == discovery.FileTypeAgent
			})
			err := componentRegistryProblems(vocabulary, fixture)
			if err == nil || !strings.Contains(err.Error(), name+": missing agent") {
				t.Fatalf("omitted agent error = %v", err)
			}
		})
	}
}
