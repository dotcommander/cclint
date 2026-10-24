package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/output"
)

func TestGitSelectionHonorsConfiguredExclusions(t *testing.T) {
	t.Parallel()
	// A child owns the command exit boundary; parallel tests never replace exitFunc.
	if root := os.Getenv("CCLINT_GIT_EXCLUDE_TEST_ROOT"); root != "" {
		staged := os.Getenv("CCLINT_GIT_EXCLUDE_TEST_MODE") == "staged"
		diff, useBaseline := !staged, true
		format, report, baselinePath := "json", os.Getenv("CCLINT_GIT_EXCLUDE_TEST_REPORT"), "accepted.json"
		err := runGitLint(executionOptions{root: &root, format: &format, output: &report, baseline: &useBaseline, baselinePath: &baselinePath}, &lintCommand{Staged: &staged, Diff: &diff})
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		os.Exit(0)
	}
	if testing.Short() {
		t.Skip("subprocess Git workflow")
	}
	root := t.TempDir()
	ignored := "plugins/marketplaces/acme/plugins/ignored/agents/broken.md"
	included := ".claude/agents/consumer.md"
	data, err := json.Marshal(map[string]any{"exclude": []string{"plugins/marketplaces/acme/plugins/ignored/**"}})
	if err != nil {
		t.Fatal(err)
	}
	requireWriteFile(t, filepath.Join(root, ".cclintrc.json"), string(data))
	for path, content := range map[string]string{
		included:                                "---\nname: consumer\ndescription: Use PROACTIVELY when reviewing source changes.\n---\n## Foundation\nSkill: review-source\n## Workflow\nRead the changed source.\n",
		".claude/skills/review-source/SKILL.md": "---\nname: review-source\ndescription: Use when reviewing source changes for correctness.\n---\n## Quick Reference\nRead source.\n## Examples\nReview a parser change.\n",
		ignored:                                 "---\nname: INVALID\ndescription: Excluded fixture.\n---\nSkill: missing-excluded-skill\n",
	} {
		path = filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		requireWriteFile(t, path, content)
	}
	quiet, create := true, true
	baselinePath := "accepted.json"
	if err := runRootCommand(executionOptions{root: &root, quiet: &quiet, baselineCreate: &create, baselinePath: &baselinePath}, &lintCommand{}); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"init", "-q"}, {"-c", "core.excludesFile=/dev/null", "add", "--", included, ignored}} {
		command := exec.Command("git", args...)
		command.Dir = root
		if out, err := command.CombinedOutput(); err != nil {
			t.Fatalf("fixture Git %v: %v, %s", args, err, out)
		}
	}
	for _, mode := range []string{"staged", "diff"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			reportPath := filepath.Join(t.TempDir(), "report.json")
			ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
			t.Cleanup(cancel)
			child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestGitSelectionHonorsConfiguredExclusions$")
			child.Env = append(os.Environ(), "CCLINT_GIT_EXCLUDE_TEST_ROOT="+root, "CCLINT_GIT_EXCLUDE_TEST_MODE="+mode, "CCLINT_GIT_EXCLUDE_TEST_REPORT="+reportPath)
			if out, err := child.CombinedOutput(); err != nil {
				t.Fatalf("Git %s should ignore excluded errors: %v, %s", mode, err, out)
			}
			data, err := os.ReadFile(reportPath)
			if err != nil {
				t.Fatal(err)
			}
			var report output.JSONReport
			if err := json.Unmarshal(data, &report); err != nil {
				t.Fatal(err)
			}
			if len(report.Results) != 1 || report.Results[0].File != included {
				t.Fatalf("wrong selected files: %+v", report.Results)
			}
			if report.Summary.TotalErrors != 0 || report.Summary.TotalSuggestions != 0 {
				t.Fatalf("baseline or unselected skill context lost: %+v", report.Summary)
			}
		})
	}
}
