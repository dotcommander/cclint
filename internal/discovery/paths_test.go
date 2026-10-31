package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverPathsPreservesFullDiscoveryMetadata(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for _, rel := range []string{".claude/agents/b.md", ".claude/agents/a.md", "commands/c.md", "commands/skip.md"} {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("body"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	fd := NewFileDiscovery(root, false).WithExclude([]string{"commands/skip.md"})
	paths, err := fd.DiscoverPathsOfTypes(map[FileType]bool{FileTypeCommand: true, FileTypeAgent: true})
	if err != nil {
		t.Fatal(err)
	}
	full, err := fd.DiscoverFilesOfTypes(map[FileType]bool{FileTypeCommand: true, FileTypeAgent: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 3 || len(full) != len(paths) {
		t.Fatalf("paths %v; full %v; want three files", paths, full)
	}
	for i, path := range paths {
		if path.Path != full[i].Path || path.RelPath != full[i].RelPath || path.Type != full[i].Type || path.Size != full[i].Size {
			t.Fatalf("metadata/order differs at %d: path %+v full %+v", i, path, full[i])
		}
		if path.Contents != "" || full[i].Contents != "body" {
			t.Fatalf("content mode differs at %d: path %q full %q", i, path.Contents, full[i].Contents)
		}
	}
}

func TestDiscoverPathsSkipsUnreadableFile(t *testing.T) {
	t.Parallel()
	if os.Getuid() == 0 {
		t.Skip("permissions are not enforced for root")
	}
	root := t.TempDir()
	path := filepath.Join(root, "commands", "unreadable.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("body"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(path, 0); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(path, 0o600) })
	if f, err := os.Open(path); err == nil {
		_ = f.Close()
		t.Skip("permissions are not enforced on this filesystem")
	}
	fd := NewFileDiscovery(root, false)
	paths, err := fd.DiscoverPathsOfTypes(map[FileType]bool{FileTypeCommand: true})
	if err != nil || len(paths) != 0 {
		t.Fatalf("metadata-only selection = %v, %v; want unreadable path omitted", paths, err)
	}
	full, err := fd.DiscoverFilesOfTypes(map[FileType]bool{FileTypeCommand: true})
	if err != nil || len(full) != 0 {
		t.Fatalf("full discovery = %v, %v; want unreadable path omitted", full, err)
	}
}
