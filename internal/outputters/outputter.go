package outputters

import (
	"fmt"
	"time"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/output"
)

// =============================================================================
// Dependency Inversion: Formatter interface for output formatters
// =============================================================================

// Formatter defines the interface for output formatters.
// This allows the Outputter to depend on an abstraction rather than
// concrete formatter implementations (DIP compliance).
type Formatter interface {
	// Format formats the lint summary and outputs it
	Format(summary *cli.LintSummary) error
}

// FormatterFactory creates formatters based on format type.
// This follows the Factory pattern for DIP compliance.
type FormatterFactory interface {
	// CreateFormatter creates a formatter for the given format type
	CreateFormatter(format string) (Formatter, error)
}

// =============================================================================
// Default FormatterFactory implementation
// =============================================================================

// DefaultFormatterFactory creates formatters using the output package.
type DefaultFormatterFactory struct {
	cfg *config.Config
}

// NewDefaultFormatterFactory creates a new DefaultFormatterFactory.
func NewDefaultFormatterFactory(cfg *config.Config) *DefaultFormatterFactory {
	return &DefaultFormatterFactory{cfg: cfg}
}

// CreateFormatter implements FormatterFactory interface.
func (f *DefaultFormatterFactory) CreateFormatter(format string) (Formatter, error) {
	switch format {
	case "console":
		return output.NewConsoleFormatter(f.cfg.Quiet, f.cfg.Verbose, f.cfg.ShowScores, f.cfg.ShowImprovements), nil
	case "json":
		return output.NewJSONFormatter(f.cfg.Quiet, true, f.cfg.Output), nil
	case "markdown":
		return output.NewMarkdownFormatter(f.cfg.Quiet, f.cfg.Verbose, f.cfg.Output), nil
	default:
		return nil, fmt.Errorf("unsupported format: %s", format)
	}
}

// =============================================================================
// Outputter with DIP-compliant design
// =============================================================================

// Outputter handles output formatting using injected formatter factory.
// It depends on the FormatterFactory abstraction, not concrete formatters.
type Outputter struct {
	config  *config.Config
	factory FormatterFactory
}

// NewOutputter creates a new Outputter with the default formatter factory.
func NewOutputter(config *config.Config) *Outputter {
	return &Outputter{
		config:  config,
		factory: NewDefaultFormatterFactory(config),
	}
}

// NewOutputterWithFactory creates a new Outputter with a custom formatter factory.
// This allows for dependency injection and easier testing.
func NewOutputterWithFactory(config *config.Config, factory FormatterFactory) *Outputter {
	return &Outputter{
		config:  config,
		factory: factory,
	}
}

// Format formats the lint summary using the configured format.
func (o *Outputter) Format(summary *cli.LintSummary, format string) error {
	// Set start time if not set
	if summary.StartTime.IsZero() {
		summary.StartTime = time.Now()
	}

	// Set project root in summary for display
	summary.ProjectRoot = o.config.Root

	// Create formatter via factory (DIP compliance)
	formatter, err := o.factory.CreateFormatter(format)
	if err != nil {
		return err
	}

	return formatter.Format(summary)
}