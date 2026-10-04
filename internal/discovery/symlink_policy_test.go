package discovery

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDiscoverySymlinkPolicy(t *testing.T) {
	base := t.TempDir()
	root := filepath.Join(base, "project")
	sibling := root + "-sibling"
	for _, dir := range []string{filepath.Join(root, "agents"), filepath.Join(root, "targets"), sibling} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, path := range []string{filepath.Join(root, "targets", "real.md"), filepath.Join(sibling, "external.md")} {
		if err := os.WriteFile(path, []byte("content"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	links := map[string]string{
		"agents/file.md":   filepath.Join(root, "targets", "real.md"),
		"agents/linked":    filepath.Join(root, "targets"),
		"agents/escape.md": filepath.Join(sibling, "external.md"),
		"agents/escaped":   sibling,
		"agents/broken.md": filepath.Join(base, "missing"),
	}
	for name, target := range links {
		if err := os.Symlink(target, filepath.Join(root, name)); err != nil {
			t.Fatal(err)
		}
	}
	for _, follow := range []bool{false, true} {
		fd := NewFileDiscovery(root, follow)
		files, err := fd.findFilesByPattern([]string{"agents/**/*.md"}, FileTypeAgent)
		if err != nil {
			t.Fatal(err)
		}
		want := 0
		if follow {
			want = 2
		}
		if len(files) != want {
			t.Fatalf("follow=%v: got %v, want %d files", follow, files, want)
		}
		for _, f := range files {
			if f.RelPath != "agents/file.md" && f.RelPath != "agents/linked/real.md" {
				t.Fatalf("unexpected logical path %s", f.RelPath)
			}
			if f.Path != filepath.Join(root, filepath.FromSlash(f.RelPath)) {
				t.Fatalf("lost logical path: %v", f)
			}
		}
	}
	linkedRoot := filepath.Join(base, "selected")
	if err := os.Symlink(root, linkedRoot); err != nil {
		t.Fatal(err)
	}
	files, err := NewFileDiscovery(linkedRoot, true).WithExclude([]string{"agents/linked/**"}).findFilesByPattern([]string{"agents/**/*.md"}, FileTypeAgent)
	if err != nil || len(files) != 1 || files[0].RelPath != "agents/file.md" {
		t.Fatalf("linked root/exclusions: %v, %v", files, err)
	}
	brokenRoot := filepath.Join(base, "broken")
	if err := os.Symlink(filepath.Join(base, "missing"), brokenRoot); err != nil {
		t.Fatal(err)
	}
	if _, err := NewFileDiscovery(brokenRoot, true).DiscoverFiles(); err == nil {
		t.Fatal("broken root must fail")
	}
}
