package cmd

import (
	"fmt"
	"os"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/outputters"
)

func loadCLIConfig() (*config.Config, error) {
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return nil, fmt.Errorf("error loading configuration: %w", err)
	}

	applyCLIOverrides(cfg)
	return cfg, nil
}

func applyCLIOverrides(cfg *config.Config) {
	if rootPath != "" {
		cfg.Root = rootPath
	}

	cfg.Quiet = quiet
	cfg.Verbose = verbose
	cfg.ShowScores = showScores
	cfg.ShowImprovements = showImprovements
	cfg.Format = outputFormat
	cfg.Output = outputFile
	cfg.FailOn = failOn
	cfg.NoCycleCheck = noCycleCheck
}

func runOrchestratedLint(cfg *config.Config, linters []lint.LinterEntry) (*lint.Result, error) {
	orchestrator := lint.NewOrchestrator(cfg, lint.OrchestratorConfig{
		RootPath:       rootPath,
		UseBaseline:    useBaseline,
		CreateBaseline: createBaseline,
		BaselinePath:   baselinePath,
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
	return outputters.NewOutputter(cfg).Format(summary, cfg.Format)
}

func formatFullRunOutput(cfg *config.Config, result *lint.Result) error {
	return outputters.NewOutputter(cfg).FormatAll(result.Summaries, result.StartTime)
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

func applyFailurePolicy(cfg *config.Config, errors, warnings, suggestions int) {
	if createBaseline {
		return
	}

	if shouldFail(cfg, errors, warnings, suggestions) {
		exitFunc(1)
	}
}
