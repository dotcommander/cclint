package discovery

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestDiscoverFilesOfTypesPreservesRegistryOrderAndExcludes(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, name := range []string{"agents/keep.md", "agents/skip.md", "commands/run.md", "skills/demo/SKILL.md"} {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("content"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	fd := NewFileDiscovery(root, false).WithExclude([]string{"agents/skip.md"})
	full, err := fd.DiscoverFiles()
	if err != nil {
		t.Fatal(err)
	}
	selected := map[FileType]bool{FileTypeAgent: true, FileTypeSkill: true}
	got, err := fd.DiscoverFilesOfTypes(selected)
	if err != nil {
		t.Fatal(err)
	}
	var want []File
	for _, f := range full {
		if selected[f.Type] {
			want = append(want, f)
		}
	}
	if !reflect.DeepEqual(got, want) || len(got) != 2 || got[0].Type != FileTypeSkill {
		t.Fatalf("selected = %v, want registry-ordered filtered snapshot %v", got, want)
	}
}
