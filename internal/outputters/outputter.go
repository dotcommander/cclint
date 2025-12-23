package outputters

import (
	"fmt"
	"time"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/output"
)

// Outputter handles output formatting
type Outputter struct {
	config *config.Config
}

// NewOutputter creates a new Outputter
func NewOutputter(config *config.Config) *Outputter {
	return &Outputter{
		config: config,
	}
}

// Format formats the lint summary using the configured format
func (o *Outputter) Format(summary *cli.LintSummary, format string) error {
	// Set start time if not set
	if summary.StartTime.IsZero() {
		summary.StartTime = time.Now()
	}

	// Set project root in summary for display
	summary.ProjectRoot = o.config.Root

	// Create appropriate formatter based on format
	switch format {
	case "console":
		formatter := output.NewConsoleFormatter(o.config.Quiet, o.config.Verbose, o.config.ShowScores, o.config.ShowImprovements)
		return formatter.Format(summary)
	case "json":
		formatter := output.NewJSONFormatter(o.config.Quiet, true, o.config.Output)
		return formatter.Format(summary)
	case "markdown":
		formatter := output.NewMarkdownFormatter(o.config.Quiet, o.config.Verbose, o.config.Output)
		return formatter.Format(summary)
	default:
		return fmt.Errorf("unsupported format: %s", format)
	}

	return nil
}