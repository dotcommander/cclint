package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFmtDiscoverySkipsUnreadablePath(t *testing.T) {
	t.Parallel()
	if os.Getuid() == 0 {
		t.Skip("permissions are not enforced for root")
	}
	root := t.TempDir()
	path := filepath.Join(root, "commands", "unreadable.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("# Example\n"), 0o600); err != nil {
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
	for _, cmd := range []fmtCommand{{}, {Args: []string{"commands"}}} {
		selected, err := collectFilesToFormat(&cmd, root, root, nil)
		if err != nil || len(selected) != 0 {
			t.Fatalf("selection = %v, %v; want unreadable path omitted", selected, err)
		}
	}
}
