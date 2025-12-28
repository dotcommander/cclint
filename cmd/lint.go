package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/outputters"
)

// LinterFunc is the function signature for component linters.
// It takes root path, quiet mode, verbose mode, noCycleCheck and returns a summary.
type LinterFunc func(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*cli.LintSummary, error)

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

	// Determine baseline path (relative to project root)
	baselineFile := baselinePath
	if !filepath.IsAbs(baselineFile) {
		baselineFile = filepath.Join(cfg.Root, baselineFile)
	}

	// Load baseline if requested
	var b *baseline.Baseline
	if useBaseline || createBaseline {
		if _, err := os.Stat(baselineFile); err == nil {
			b, err = baseline.LoadBaseline(baselineFile)
			if err != nil && !cfg.Quiet {
				fmt.Fprintf(os.Stderr, "Warning: Failed to load baseline: %v\n", err)
				b = nil
			}
		}
	}

	// Run component-specific linter
	summary, err := linter(cfg.Root, cfg.Quiet, cfg.Verbose, cfg.NoCycleCheck)
	if err != nil {
		return fmt.Errorf("error running %s linter: %w", linterName, err)
	}

	// Collect issues for baseline creation
	var allIssues []cue.ValidationError
	if createBaseline {
		allIssues = cli.CollectAllIssues(summary)
	}

	// Filter with baseline if active
	var totalIgnored, errorsIgnored, suggestionsIgnored int
	if useBaseline && b != nil {
		totalIgnored, errorsIgnored, suggestionsIgnored = cli.FilterResults(summary, b)
	}

	// Format and output results
	outputter := outputters.NewOutputter(cfg)
	if err := outputter.Format(summary, cfg.Format); err != nil {
		return fmt.Errorf("error formatting output: %w", err)
	}

	// Create/update baseline if requested (do this BEFORE exiting on errors)
	if createBaseline {
		b = baseline.CreateBaseline(allIssues)
		b.CreatedAt = time.Now().UTC().Format(time.RFC3339)

		if err := b.SaveBaseline(baselineFile); err != nil {
			return fmt.Errorf("failed to save baseline: %w", err)
		}

		if !cfg.Quiet {
			fmt.Printf("\nBaseline created: %s (%d issues)\n", baselineFile, len(b.Fingerprints))
		}

		// When creating baseline, exit 0 (success) to accept current state
		return nil
	}

	// Print baseline filtering summary
	if useBaseline && b != nil && totalIgnored > 0 && !cfg.Quiet {
		fmt.Printf("\n%d baseline issues ignored (%d errors, %d suggestions)\n",
			totalIgnored, errorsIgnored, suggestionsIgnored)
	}

	return nil
}
