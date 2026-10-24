package output

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/lint"
)

func TestConfiguredQuietSuppressesReports(t *testing.T) {
	t.Parallel()
	for _, format := range []string{"console", "json", "markdown"} {
		t.Run(format, func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(t.TempDir(), "report")
			cfg := &config.Config{Quiet: true, Format: format, Output: path}
			summary := &lint.LintSummary{TotalFiles: 1, SuccessfulFiles: 1, StartTime: time.Now()}
			if err := FormatSummary(cfg, summary); err != nil {
				t.Fatal(err)
			}
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Fatalf("quiet report unexpectedly wrote a file: %v", err)
			}
		})
	}
}
