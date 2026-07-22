package cmd

import (
	"fmt"
	"os"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/output"
)

func loadCLIConfig(opts executionOptions) (*config.Config, error) {
	cfg, err := config.LoadConfig(stringValue(opts.root))
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	applyCLIOverrides(cfg, opts)
	return cfg, nil
}

func applyCLIOverrides(cfg *config.Config, opts executionOptions) {
	cfg.Version = Version
	if opts.root != nil {
		cfg.Root = *opts.root
	}
	if opts.quiet != nil {
		cfg.Quiet = *opts.quiet
	}
	if opts.verbose != nil {
		cfg.Verbose = *opts.verbose
	}
	if opts.scores != nil {
		cfg.ShowScores = *opts.scores
	}
	if opts.improvements != nil {
		cfg.ShowImprovements = *opts.improvements
	}
	if opts.format != nil {
		cfg.Format = *opts.format
	}
	if opts.output != nil {
		cfg.Output = *opts.output
	}
	if opts.failOn != nil {
		cfg.FailOn = *opts.failOn
	}
	if opts.noCycleCheck != nil {
		cfg.NoCycleCheck = *opts.noCycleCheck
	}
}

func runOrchestratedLint(cfg *config.Config, opts executionOptions, linters []lint.LinterEntry) (*lint.Result, error) {
	orchestrator := lint.NewOrchestrator(cfg, lint.OrchestratorConfig{
		UseBaseline:    boolValue(opts.baseline),
		CreateBaseline: boolValue(opts.baselineCreate),
		BaselinePath:   stringValueOr(opts.baselinePath, ".cclintbaseline.json"),
	})
	if linters != nil {
		orchestrator.WithLinters(linters)
	}

	stop := startSpinner(cfg)
	result, err := orchestrator.Run()
	stop()
	if err != nil {
		return nil, err
	}

	return result, nil
}

func formatSummaryOutput(cfg *config.Config, summary *lint.LintSummary) error {
	return output.FormatSummary(cfg, summary)
}

func formatFullRunOutput(cfg *config.Config, result *lint.Result) error {
	return output.FormatAll(cfg, result.Summaries, result.StartTime)
}

func printBaselineSummary(total, errors, suggestions int, quiet bool) {
	if total == 0 || quiet {
		return
	}
	fmt.Fprintf(os.Stderr, "\n%d baseline issues ignored (%d errors, %d suggestions)\n",
		total, errors, suggestions)
}

func printValidationReminder(cfg *config.Config) {
	if cfg.Quiet || !cfg.Verbose {
		return
	}

	fmt.Fprintln(os.Stderr, "\n  Validate suggestions against docs.anthropic.com or docs.claude.com")
}

func applyFailurePolicy(cfg *config.Config, opts executionOptions, errors, warnings, suggestions int) {
	if boolValue(opts.baselineCreate) {
		return
	}

	if shouldFail(cfg, errors, warnings, suggestions) {
		exitFunc(1)
	}
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func stringValueOr(value *string, fallback string) string {
	if value == nil {
		return fallback
	}
	return *value
}
