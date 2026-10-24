package output

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/lint"
)

// Self-test subprocesses exercise stdout without replacing process-global I/O.
func TestCompactWarningTotals(t *testing.T) {
	t.Parallel()
	if mode := os.Getenv("CCLINT_TEST_COMPACT_WARNINGS"); mode != "" {
		summaries := []*lint.LintSummary{
			{ComponentType: "agent", TotalFiles: 1, SuccessfulFiles: 1, TotalWarnings: 1,
				Results: []lint.LintResult{{File: "agent.md", Warnings: []cue.ValidationError{{
					File: "agent.md", Message: "first warning", Severity: cue.SeverityWarning,
				}}}}},
			{ComponentType: "command", TotalFiles: 1, SuccessfulFiles: 1, TotalWarnings: 1,
				Results: []lint.LintResult{{File: "command.md", Warnings: []cue.ValidationError{{
					File: "command.md", Message: "second warning", Severity: cue.SeverityWarning,
				}}}}},
		}
		f := NewCompactFormatter(mode == "quiet", mode == "verbose", false, false, time.Now())
		f.colorize = false
		if err := f.FormatAll(summaries); err != nil {
			t.Fatal(err)
		}
		return
	}
	if testing.Short() {
		t.Skip("subprocess stdout integration")
	}
	for _, mode := range []string{"minimal", "verbose", "quiet"} {
		t.Run(mode, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithTimeout(t.Context(), 4*time.Second)
			t.Cleanup(cancel)
			child := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCompactWarningTotals$")
			child.Env = append(os.Environ(), "CCLINT_TEST_COMPACT_WARNINGS="+mode)
			data, err := child.CombinedOutput()
			if err != nil {
				t.Fatalf("compact subprocess: %v\n%s", err, data)
			}
			got := string(data)
			if mode == "quiet" {
				if strings.Contains(got, "warning") || strings.Contains(got, "✓ PASS") {
					t.Fatalf("quiet compact output leaked: %s", got)
				}
				return
			}
			if !strings.Contains(got, "2 warnings") || strings.Count(got, "warning: ") != 2 ||
				!strings.Contains(got, "warning: agent.md: first warning") ||
				!strings.Contains(got, "warning: command.md: second warning") ||
				strings.Contains(got, "FAIL") {
				t.Fatalf("warning totals/details or status incorrect: %s", got)
			}
			if mode == "minimal" && !strings.Contains(got, "✓ PASS  2 files  2 warnings") {
				t.Fatalf("warning-only run must retain PASS: %s", got)
			}
			if mode == "verbose" && !strings.Contains(got, "2/2 passed, 2 warnings") {
				t.Fatalf("verbose warning aggregate missing: %s", got)
			}
		})
	}
}
