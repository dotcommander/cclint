package output

import (
	"bytes"
	"fmt"
	"os"
	"time"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/lint"
)

const (
	formatConsole  = "console"
	formatJSON     = "json"
	formatMarkdown = "markdown"
)

// summaryFormatter is the behavior required by the configured single-summary
// output path. Concrete formatter types remain owned by this package.
type summaryFormatter interface {
	Format(summary *lint.LintSummary) error
}

// FormatSummary formats one lint summary using the configured output format.
func FormatSummary(cfg *config.Config, summary *lint.LintSummary) error {
	if summary.StartTime.IsZero() {
		summary.StartTime = time.Now()
	}
	summary.ProjectRoot = cfg.Root

	formatter, err := newSummaryFormatter(cfg)
	if err != nil {
		return err
	}
	if console, ok := formatter.(*ConsoleFormatter); ok && cfg.Output != "" {
		var buffer bytes.Buffer
		console.writer = &buffer
		if err := console.Format(summary); err != nil {
			return err
		}
		return os.WriteFile(cfg.Output, buffer.Bytes(), 0600)
	}
	return formatter.Format(summary)
}

func newSummaryFormatter(cfg *config.Config) (summaryFormatter, error) {
	switch cfg.Format {
	case formatConsole:
		return NewConsoleFormatter(cfg.Quiet, cfg.Verbose, cfg.ShowScores, cfg.ShowImprovements), nil
	case formatJSON:
		return NewJSONFormatterWithVersion(cfg.Quiet, true, cfg.Output, cfg.Version), nil
	case formatMarkdown:
		return NewMarkdownFormatter(cfg.Quiet, cfg.Verbose, cfg.Output), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", cfg.Format)
	}
}

// FormatAll preserves compact console output and uses one report envelope for
// machine formats, with component types retained on each result.
func FormatAll(cfg *config.Config, summaries []*lint.LintSummary, startTime time.Time) error {
	if cfg.Format != formatConsole || cfg.Output != "" || cfg.ShowScores || cfg.ShowImprovements {
		combined := &lint.LintSummary{StartTime: startTime, ProjectRoot: cfg.Root}
		for _, summary := range summaries {
			combined.Results = append(combined.Results, summary.Results...)
			combined.TotalFiles += summary.TotalFiles
			combined.SuccessfulFiles += summary.SuccessfulFiles
			combined.FailedFiles += summary.FailedFiles
			combined.TotalErrors += summary.TotalErrors
			combined.TotalWarnings += summary.TotalWarnings
			combined.TotalSuggestions += summary.TotalSuggestions
		}
		return FormatSummary(cfg, combined)
	}
	return NewCompactFormatter(cfg.Quiet, cfg.Verbose, cfg.ShowScores, cfg.ShowImprovements, startTime).FormatAll(summaries)
}
