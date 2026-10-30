package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/output"
)

func TestGitSelectionSkipsNonComponentFiles(t *testing.T) {
	t.Parallel()
	// A child owns the command exit boundary; parallel tests never replace exitFunc.
	if root := os.Getenv("CCLINT_GIT_SELECTION_TEST_ROOT"); root != "" {
		docsOnly := os.Getenv("CCLINT_GIT_SELECTION_TEST_MODE") == "docs-only"
		staged, diff := false, true
		format := "json"
		report := os.Getenv("CCLINT_GIT_SELECTION_TEST_REPORT")
		if err := runGitLint(executionOptions{root: &root, format: &format, output: &report}, &lintCommand{Staged: &staged, Diff: &diff}); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		if docsOnly {
			// The docs-only changeset ends at "No files to lint" before a report is written.
			if _, err := os.Stat(report); err == nil {
				fmt.Fprintf(os.Stderr, "unexpected report for docs-only changeset: %s\n", report)
				os.Exit(3)
			}
		}
		os.Exit(0)
	}
	if testing.Short() {
		t.Skip("subprocess Git workflow")
	}
	root := t.TempDir()
	agent := ".claude/agents/consumer.md"
	skill := ".claude/skills/review-source/SKILL.md"
	docs := ".claude/note.md"
	for path, content := range map[string]string{
		agent: "---\nname: consumer\ndescription: Use PROACTIVELY when reviewing source changes.\n---\n## Foundation\nSkill: review-source\n## Workflow\nRead the changed source.\n",
		skill: "---\nname: review-source\ndescription: Use when reviewing source changes for correctness.\n---\n## Quick Reference\nRead source.\n## Examples\nReview a parser change.\n",
		docs:  "# Notes\n\nInternal documentation.\n",
	} {
		path = filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		requireWriteFile(t, path, content)
	}
	commands := [][]string{{"init", "-q"}, {"add", "--", agent, skill, docs}, {"-c", "user.name=test", "-c", "user.email=test@example.com", "commit", "-q", "-m", "fixture"}}
	for _, args := range commands {
		command := exec.Command("git", append([]string{"-c", "core.excludesFile=/dev/null"}, args...)...)
		command.Dir = root
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("fixture Git %v: %v, %s", args, err, out)
		}
	}
	// Change both a component file and a non-component file.
	requireWriteFile(t, filepath.Join(root, agent), "---\nname: consumer\ndescription: Use PROACTIVELY when reviewing source changes.\n---\n## Foundation\nSkill: review-source\n## Workflow\nRead the changed source carefully.\n")
	requireWriteFile(t, filepath.Join(root, docs), "# Notes\n\nInternal documentation, revised.\n")

	// Sequential phases: both children share this worktree, so the agent
	// revert below must not race the mixed-changeset child.
	reportPath := runGitSelectionChild(t, root, "mixed")
	data, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	var report output.JSONReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Results) != 1 || report.Results[0].File != filepath.ToSlash(agent) {
		t.Fatalf("non-component file not skipped or component file lost: %+v", report.Results)
	}
	if report.Summary.TotalErrors != 0 {
		t.Fatalf("selected component should lint without errors: %+v", report.Summary)
	}

	command := exec.Command("git", "-c", "core.excludesFile=/dev/null", "checkout", "--", agent)
	command.Dir = root
	if out, err := command.CombinedOutput(); err != nil {
		t.Fatalf("fixture git checkout: %v, %s", err, out)
	}
	runGitSelectionChild(t, root, "docs-only")
}

func runGitSelectionChild(t *testing.T, root, mode string) string {
	t.Helper()
	reportPath := filepath.Join(t.TempDir(), "report.json")
	child := exec.Command(os.Args[0], "-test.run=^TestGitSelectionSkipsNonComponentFiles$")
	child.Env = append(os.Environ(), "CCLINT_GIT_SELECTION_TEST_ROOT="+root, "CCLINT_GIT_SELECTION_TEST_MODE="+mode, "CCLINT_GIT_SELECTION_TEST_REPORT="+reportPath)
	if out, err := child.CombinedOutput(); err != nil {
		t.Fatalf("Git %s child should pass: %v, %s", mode, err, out)
	}
	return reportPath
}
