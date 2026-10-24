package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/discovery"
)

type discoverySnapshotLinter struct {
	recordingComponentLinter
	afterDiscovery func(*LinterContext)
}

func (l *discoverySnapshotLinter) PostProcessBatch(ctx *LinterContext, _ *LintSummary) {
	l.afterDiscovery(ctx)
}

func TestOrchestratorDiscoverySnapshot(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("workspace integration fixture")
	}
	oversized := strings.Repeat("x", CombinedMemorySizeWarningKB*1024+1)
	for _, tt := range []struct {
		name           string
		initial        string
		afterDiscovery string
		wantWarnings   [2]int
		wantFirstFiles int
	}{
		{"truncated_after_discovery", oversized, "", [2]int{1, 0}, 1},
		{"created_after_empty_discovery", "", oversized, [2]int{0, 1}, 0},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			path := filepath.Join(root, "CLAUDE.md")
			if tt.initial != "" {
				if err := os.WriteFile(path, []byte(tt.initial), 0o600); err != nil {
					t.Fatal(err)
				}
			}
			var contexts []*LinterContext
			entry := LinterEntry{
				FileType: discovery.FileTypeAgent,
				Name:     "agents",
				New: func(string) ComponentLinter {
					return &discoverySnapshotLinter{
						recordingComponentLinter: recordingComponentLinter{
							fileType: discovery.FileTypeAgent,
							typeName: "agents",
						},
						afterDiscovery: func(ctx *LinterContext) {
							contexts = append(contexts, ctx)
							if err := os.WriteFile(path, []byte(tt.afterDiscovery), 0o600); err != nil {
								t.Fatal(err)
							}
						},
					}
				},
			}
			orch := NewOrchestrator(&config.Config{Root: root, Quiet: true}, OrchestratorConfig{}).
				WithLinters([]LinterEntry{entry})
			for run, want := range tt.wantWarnings {
				result, err := orch.Run()
				if err != nil {
					t.Fatalf("Run %d: %v", run, err)
				}
				combinedWarnings := 0
				for _, summary := range result.Summaries {
					for _, issue := range CollectAllIssues(summary) {
						if strings.HasPrefix(issue.Message, "Always-loaded memory (") {
							combinedWarnings++
						}
					}
				}
				if combinedWarnings != want || result.TotalWarnings != want || result.TotalFiles != 0 || result.TotalErrors != 0 {
					t.Fatalf("Run %d: combined warnings=%d, result=%+v; want %d diagnostic-only warnings", run, combinedWarnings, result, want)
				}
			}
			if len(contexts) != 2 || contexts[0] == contexts[1] {
				t.Fatalf("contexts = %p; want two distinct run-local contexts", contexts)
			}
			if got := len(contexts[0].Files); got != tt.wantFirstFiles {
				t.Fatalf("first snapshot has %d files; want %d", got, tt.wantFirstFiles)
			}
		})
	}
}
