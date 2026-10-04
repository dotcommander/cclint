package cmd

import (
	"encoding/json"
	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/output"
	"os"
	"path/filepath"
	"testing"
)

func TestMultiTypeOneReportAndExit(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"agents", "commands"} {
		if err := os.Mkdir(filepath.Join(root, dir), 0700); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "agents", "bad.md"), []byte("---\nname: BAD\ndescription: test\n---\nbody\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "commands", "check.md"), []byte("---\ndescription: check\n---\nTask(check)\n"), 0600); err != nil {
		t.Fatal(err)
	}
	report := filepath.Join(root, "report.json")
	format := "json"
	exits := 0
	old := exitFunc
	exitFunc = func(code int) { exits++ }
	defer func() { exitFunc = old }()
	err := runRootCommand(executionOptions{root: &root, format: &format, output: &report}, &lintCommand{Paths: []string{"agents", "commands"}})
	if err != nil {
		t.Fatal(err)
	}
	if exits != 1 {
		t.Fatalf("exit calls=%d", exits)
	}
	data, err := os.ReadFile(report)
	if err != nil {
		t.Fatal(err)
	}
	var envelope map[string]json.RawMessage
	if err = json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("one report: %v", err)
	}
	if !json.Valid(data) {
		t.Fatal("invalid JSON")
	}
	var reportData output.JSONReport
	if err = json.Unmarshal(data, &reportData); err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	for _, result := range reportData.Results {
		seen[result.Type] = true
	}
	if !seen["agent"] || !seen["command"] {
		t.Fatalf("later type skipped: %v", seen)
	}
	if err = runRootCommand(executionOptions{root: &root, format: &format, output: &root}, &lintCommand{Paths: []string{"agents", "commands"}}); err == nil {
		t.Fatal("output failure swallowed")
	}
}
func TestSelectedFileBaselineLifecycle(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "agents"), 0700); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "agents", "bad.md")
	if err := os.WriteFile(file, []byte("---\nname: BAD\ndescription: test\n---\nbody\n"), 0600); err != nil {
		t.Fatal(err)
	}
	create := true
	quiet := true
	base := filepath.Join(root, "selected.json")
	if err := runRootCommand(executionOptions{root: &root, quiet: &quiet, baselineCreate: &create, baselinePath: &base}, &lintCommand{Paths: []string{file}}); err != nil {
		t.Fatal(err)
	}
	b, err := baseline.LoadBaseline(base)
	if err != nil || len(b.Fingerprints) == 0 {
		t.Fatalf("selected snapshot: %v %+v", err, b)
	}
	use := true
	old := exitFunc
	exitFunc = func(int) { t.Error("known selected issue caused failure") }
	defer func() { exitFunc = old }()
	if err := runRootCommand(executionOptions{root: &root, quiet: &quiet, baseline: &use, baselinePath: &base}, &lintCommand{Paths: []string{file}}); err != nil {
		t.Fatal(err)
	}
}

func TestWarningOnlyFailurePolicy(t *testing.T) {
	if shouldFail(&config.Config{FailOn: "error"}, 0, 1, 0) {
		t.Fatal("warning escalated to error")
	}
	if !shouldFail(&config.Config{FailOn: "warning"}, 0, 1, 0) {
		t.Fatal("warning failure policy ignored")
	}
}
