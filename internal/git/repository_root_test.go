package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"testing"
)

func TestGitSelectionsFromChildDirectory(t *testing.T) {
	root := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %s: %v", args, out, err)
		}
	}
	run("init")
	for _, path := range []string{"rules/changed.md", "output-styles/deleted.md", "nested/child/.keep"} {
		abs := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(abs), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(abs, []byte("before"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	run("add", ".")
	run("-c", "user.name=Test", "-c", "user.email=test@example.invalid", "commit", "-m", "fixture")
	if err := os.WriteFile(filepath.Join(root, "rules/changed.md"), []byte("after"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(root, "output-styles/deleted.md")); err != nil {
		t.Fatal(err)
	}
	run("add", "-A")
	if err := os.WriteFile(filepath.Join(root, "output-styles/new.md"), []byte("new"), 0644); err != nil {
		t.Fatal(err)
	}
	child := filepath.Join(root, "nested/child")
	top, err := RepositoryRoot(child)
	if err != nil {
		t.Fatal(err)
	}
	// Temp directories can have a linked system ancestor; compare canonical paths.
	canonical, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	if top != canonical {
		t.Fatalf("top=%s, want %s", top, canonical)
	}
	for _, staged := range []bool{false, true} {
		var fromRoot, fromChild FileChanges
		if staged {
			fromRoot, err = GetStagedChanges(root)
			if err == nil {
				fromChild, err = GetStagedChanges(child)
			}
		} else {
			fromRoot, err = GetChangedFileChanges(root)
			if err == nil {
				fromChild, err = GetChangedFileChanges(child)
			}
		}
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(fromRoot, fromChild) {
			t.Fatalf("staged=%v: root=%+v child=%+v", staged, fromRoot, fromChild)
		}
		want := 1
		if !staged {
			want = 2
		}
		if len(fromChild.Files) != want || !reflect.DeepEqual(fromChild.DeletedFiles, []string{"output-styles/deleted.md"}) {
			t.Fatalf("selection=%+v", fromChild)
		}
	}
}
