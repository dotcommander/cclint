package output

import (
	"fmt"
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

// FormatAll formats a full, multi-component run in compact form. Full-run
// output intentionally ignores cfg.Format and cfg.Output.
func FormatAll(cfg *config.Config, summaries []*lint.LintSummary, startTime time.Time) error {
	if cfg.Quiet {
		return nil
	}

	formatter := NewCompactFormatter(cfg.Quiet, cfg.Verbose, cfg.ShowScores, cfg.ShowImprovements, startTime)
	return formatter.FormatAll(summaries)
}
