package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestFmtConfiguredQuiet(t *testing.T) {
	t.Parallel()
	// The child owns stdout and exits after the real formatter command.
	if root := os.Getenv("CCLINT_FMT_QUIET_TEST_ROOT"); root != "" {
		if err := runFmt(executionOptions{root: &root}, &fmtCommand{Check: true}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(0)
	}
	if testing.Short() {
		t.Skip("subprocess CLI check")
	}
	root := t.TempDir()
	data, err := json.Marshal(map[string]any{"quiet": true})
	if err != nil {
		t.Fatal(err)
	}
	requireWriteFile(t, filepath.Join(root, ".cclintrc.json"), string(data))
	if err := os.Mkdir(filepath.Join(root, "commands"), 0700); err != nil {
		t.Fatal(err)
	}
	requireWriteFile(t, filepath.Join(root, "commands/review.md"), "---\ndescription: Review source changes.\nallowed-tools: Agent\n---\nAgent(helper)\n")
	child := exec.Command(os.Args[0], "-test.run=^TestFmtConfiguredQuiet$")
	child.Env = append(os.Environ(), "CCLINT_FMT_QUIET_TEST_ROOT="+root)
	out, err := child.CombinedOutput()
	if exit, ok := err.(*exec.ExitError); !ok || exit.ExitCode() != 1 {
		t.Fatalf("check should still fail for unformatted file: %v, %s", err, out)
	}
	if len(out) != 0 {
		t.Fatalf("configured quiet emitted %q", out)
	}
}
