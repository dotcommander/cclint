package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/crossfile"
)

func TestSkillPostProcessBatchDoesNotRetainReferenceSnapshot(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	path := filepath.Join(root, "skills", "routing", "references", "routes.md")
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	write := func(target string) {
		t.Helper()
		body := "| Trigger | Target |\n|---|---|\n| review | " + target + " |\n"
		if err := os.WriteFile(path, []byte(body), 0644); err != nil {
			t.Fatal(err)
		}
	}
	ctx := &LinterContext{RootPath: root, CrossValidator: crossfile.NewCrossFileValidator(nil), NoCycleCheck: true}
	linter := NewSkillLinter(root)
	write("missing-skill")
	first := &LintSummary{}
	linter.PostProcessBatch(ctx, first)
	if len(first.Results) != 1 || len(first.Results[0].Errors) != 1 {
		t.Fatalf("first batch did not report ghost: %+v", first.Results)
	}
	write("different-skill")
	second := &LintSummary{}
	linter.PostProcessBatch(ctx, second)
	if len(second.Results) != 1 || len(second.Results[0].Errors) != 1 ||
		first.Results[0].Errors[0].Message == second.Results[0].Errors[0].Message {
		t.Fatalf("later batch reused previous snapshot: %+v", second.Results)
	}
}
