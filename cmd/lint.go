package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/outputters"
)

// LinterFunc is the function signature for component linters.
// It takes root path, quiet mode, verbose mode, noCycleCheck and returns a summary.
type LinterFunc func(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*lint.LintSummary, error)

// runComponentLint is the generic function that handles config loading,
// linter execution, and output formatting for any component type.
// This follows the Single Responsibility Principle by separating
// orchestration from component-specific linting logic.
func runComponentLint(linterName string, linter LinterFunc) error {
	// Load configuration
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Determine and load baseline if requested
	baselineFile := getBaselinePath(cfg.Root)
	b := loadBaselineIfRequested(baselineFile, cfg.Quiet)

	// Run component-specific linter
	summary, err := linter(cfg.Root, cfg.Quiet, cfg.Verbose, cfg.NoCycleCheck)
	if err != nil {
		return fmt.Errorf("error running %s linter: %w", linterName, err)
	}

	// Collect issues for baseline creation
	allIssues := collectIssuesIfCreatingBaseline(summary)

	// Filter with baseline if active
	totalIgnored, errorsIgnored, suggestionsIgnored := filterWithBaseline(summary, b)

	// Format and output results
	if err := formatAndOutputResults(cfg, summary); err != nil {
		return err
	}

	// Handle baseline creation
	if createBaseline {
		return handleBaselineCreation(baselineFile, allIssues, cfg.Quiet)
	}

	// Print baseline filtering summary
	printBaselineSummary(totalIgnored, errorsIgnored, suggestionsIgnored, cfg.Quiet)

	return nil
}

// getBaselinePath returns the absolute path to the baseline file.
func getBaselinePath(projectRoot string) string {
	if filepath.IsAbs(baselinePath) {
		return baselinePath
	}
	return filepath.Join(projectRoot, baselinePath)
}

// loadBaselineIfRequested loads the baseline file if baseline mode is enabled.
func loadBaselineIfRequested(baselineFile string, quiet bool) *baseline.Baseline {
	if !useBaseline && !createBaseline {
		return nil
	}

	if _, err := os.Stat(baselineFile); err != nil {
		return nil
	}

	b, err := baseline.LoadBaseline(baselineFile)
	if err != nil && !quiet {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load baseline: %v\n", err)
		return nil
	}

	return b
}

// collectIssuesIfCreatingBaseline collects all issues if baseline creation is requested.
func collectIssuesIfCreatingBaseline(summary *lint.LintSummary) []cue.ValidationError {
	if !createBaseline {
		return nil
	}
	return lint.CollectAllIssues(summary)
}

// filterWithBaseline filters results using baseline if active.
// Returns counts of ignored issues.
func filterWithBaseline(summary *lint.LintSummary, b *baseline.Baseline) (total, errors, suggestions int) {
	if !useBaseline || b == nil {
		return 0, 0, 0
	}
	return lint.FilterResults(summary, b)
}

// formatAndOutputResults formats and outputs the lint results.
func formatAndOutputResults(cfg *config.Config, summary *lint.LintSummary) error {
	outputter := outputters.NewOutputter(cfg)
	if err := outputter.Format(summary, cfg.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}
	return nil
}

// handleBaselineCreation creates and saves the baseline file.
func handleBaselineCreation(baselineFile string, allIssues []cue.ValidationError, quiet bool) error {
	b := baseline.CreateBaseline(allIssues)
	b.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := b.SaveBaseline(baselineFile); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	if !quiet {
		fmt.Printf("\nBaseline created: %s (%d issues)\n", baselineFile, len(b.Fingerprints))
	}

	// When creating baseline, exit 0 (success) to accept current state
	return nil
}

// printBaselineSummary prints the baseline filtering summary if there are ignored issues.
func printBaselineSummary(total, errors, suggestions int, quiet bool) {
	if total == 0 || quiet {
		return
	}
	fmt.Printf("\n%d baseline issues ignored (%d errors, %d suggestions)\n",
		total, errors, suggestions)
}
