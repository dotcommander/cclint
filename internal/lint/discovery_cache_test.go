package lint

import (
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

func TestDiscoveryCacheIsolatesRootsAndRetainsSnapshots(t *testing.T) {
	t.Parallel()
	first, second := selectionTestRoot(t), selectionTestRoot(t)
	firstFile := writeSelectionFile(t, first, "agents/first.md", "first snapshot\n")
	secondFile := writeSelectionFile(t, second, "agents/second.md", "second snapshot\n")
	var cache DiscoveryCache
	assertDiscoveryPaths(t, &cache, first, []string{firstFile})
	assertDiscoveryPaths(t, &cache, second, []string{secondFile})
	writeSelectionFile(t, first, "agents/later.md", "later\n")
	writeSelectionFile(t, first, "agents/first.md", "changed contents\n")
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(cwd, first)
	if err != nil {
		t.Fatal(err)
	}
	for _, alias := range []string{first, first + string(filepath.Separator) + ".", relative} {
		files := assertDiscoveryPaths(t, &cache, alias, []string{firstFile})
		if files[0].Contents != "first snapshot\n" {
			t.Fatalf("snapshot contents changed through %q: %q", alias, files[0].Contents)
		}
	}
	assertDiscoveryPaths(t, &cache, second, []string{secondFile})
}

func TestDiscoveryCacheIsolatesAndRetainsErrors(t *testing.T) {
	t.Parallel()
	for _, errorFirst := range []bool{true, false} {
		name := "success first"
		if errorFirst {
			name = "error first"
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			root := selectionTestRoot(t)
			missing := filepath.Join(root, "missing")
			good := writeSelectionFile(t, root, "agents/good.md", "fixture\n")
			var cache DiscoveryCache
			if !errorFirst {
				assertDiscoveryPaths(t, &cache, root, []string{good})
			}
			_, initialErr := cache.Get(missing)
			if initialErr == nil {
				t.Fatal("missing root unexpectedly discovered")
			}
			writeSelectionFile(t, missing, "agents/repaired.md", "fixture\n")
			if _, err := cache.Get(missing); err != initialErr {
				t.Fatalf("cached error = %v, want original %v", err, initialErr)
			}
			assertDiscoveryPaths(t, &cache, root, []string{good})
		})
	}
}

func TestDiscoveryCacheConcurrentRoots(t *testing.T) {
	t.Parallel()
	roots := []string{selectionTestRoot(t), selectionTestRoot(t)}
	files := []string{
		writeSelectionFile(t, roots[0], "agents/first.md", "fixture\n"),
		writeSelectionFile(t, roots[1], "agents/second.md", "fixture\n"),
	}
	var cache DiscoveryCache
	for _, name := range []string{"first", "second", "third", "fourth"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			for index, root := range roots {
				assertDiscoveryPaths(t, &cache, root, []string{files[index]})
			}
		})
	}
}

func TestSharedDiscoveryCacheKeepsCrossFileIndexesLocal(t *testing.T) {
	t.Parallel()
	roots := []string{selectionTestRoot(t), selectionTestRoot(t)}
	names := []string{"cache-fixture-alpha-only", "cache-fixture-beta-only"}
	agents := make([]string, len(roots))
	for index, root := range roots {
		name := names[index]
		writeSelectionFile(t, root, ".claude/skills/"+name+"/SKILL.md", "---\nname: "+name+"\ndescription: Fixture\n---\nFixture.\n")
		agents[index] = writeSelectionFile(t, root, ".claude/agents/example.md", "---\nname: example\ndescription: Fixture\n---\nSkill: "+name+"\n")
	}
	var cache DiscoveryCache
	for index, root := range roots {
		ctx := &SingleFileLinterContext{RootPath: root, discoveryCache: &cache}
		validator := ctx.EnsureCrossFileValidator()
		if validator == nil {
			t.Fatalf("root %s has no cross-file validator: %v", root, ctx.crossLoadErr)
		}
		if errors := validator.ValidateAgent(agents[index], "Skill: "+names[index]+"\n", nil); len(errors) != 0 {
			t.Fatalf("own skill unresolved in %s: %+v", root, errors)
		}
		if errors := validator.ValidateAgent(agents[index], "Skill: "+names[1-index]+"\n", nil); len(errors) != 1 {
			t.Fatalf("other project's skill resolved in %s: %+v", root, errors)
		}
	}
	// Exercise the actual selected-file caller with independently inferred roots.
	summary, err := LintFiles(agents, "", "", true, false)
	if err != nil {
		t.Fatal(err)
	}
	if summary.TotalFiles != len(agents) || len(summary.Results) != len(agents) {
		t.Fatalf("multi-root summary lost files: %+v", summary)
	}
	for _, result := range summary.Results {
		if result.Type == "unknown" {
			t.Fatalf("root inference failed: %+v", result)
		}
		for _, issue := range result.Errors {
			if strings.Contains(issue.Message, "non-existent skill") {
				t.Fatalf("local skill lost in multi-root run: %+v", issue)
			}
		}
	}
}

func assertDiscoveryPaths(t *testing.T, cache *DiscoveryCache, root string, want []string) []discovery.File {
	t.Helper()
	files, err := cache.Get(root)
	if err != nil {
		t.Fatal(err)
	}
	got := make([]string, 0, len(files))
	for _, file := range files {
		got = append(got, file.Path)
	}
	slices.Sort(got)
	want = slices.Clone(want)
	slices.Sort(want)
	if !slices.Equal(got, want) {
		t.Fatalf("root %s discovered %v, want %v", root, got, want)
	}
	return files
}
