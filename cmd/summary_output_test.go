package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/output"
)

func TestSummaryConfiguredOutput(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	for path, content := range map[string]string{
		"agents/helper.md":       "---\nname: helper\ndescription: Review source changes for correctness.\n---\n## Foundation\nRead source.\n## Workflow\nReview changes.\n",
		"commands/review.md":     "---\ndescription: Review source changes.\nallowed-tools: Agent\n---\nAgent(helper)\n",
		"skills/review/SKILL.md": "---\nname: review\ndescription: Review source changes for correctness.\n---\n## Quick Reference\nRead source.\n## Examples\nReview a parser change.\n",
	} {
		path = filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		requireWriteFile(t, path, content)
	}
	for _, format := range []string{"json", "markdown"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "report")
			if err := runSummary(executionOptions{root: &root, format: &format, output: &path}); err != nil {
				t.Fatal(err)
			}
			data, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			if format == "markdown" {
				if !strings.Contains(string(data), "# CCLint Report") || !strings.Contains(string(data), "agents/helper.md") {
					t.Fatalf("missing Markdown report content: %s", data)
				}
				return
			}
			var report output.JSONReport
			if err := json.Unmarshal(data, &report); err != nil {
				t.Fatal(err)
			}
			seen := make(map[string]bool)
			for _, result := range report.Results {
				seen[result.File] = true
				if result.Quality == nil {
					t.Errorf("missing quality for %s", result.File)
				}
			}
			if len(seen) != 3 || !seen["agents/helper.md"] || !seen["commands/review.md"] || !seen["skills/review/SKILL.md"] {
				t.Fatalf("wrong selected files: %v", seen)
			}
		})
	}
	t.Run("write failure", func(t *testing.T) {
		t.Parallel()
		format := "json"
		if err := runSummary(executionOptions{root: &root, format: &format, output: &root}); err == nil {
			t.Fatal("output failure swallowed")
		}
	})
}
