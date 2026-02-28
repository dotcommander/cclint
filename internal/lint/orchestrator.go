// Package lint provides the core linting orchestration logic.
package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

// LinterFunc is the function signature for component linters.
type LinterFunc func(rootPath string, quiet, verbose, noCycleCheck bool) (*LintSummary, error)

// DefaultLinters returns the standard set of component linters.
func DefaultLinters() []LinterEntry {
	return []LinterEntry{
		{Name: "agents", Linter: LintAgents},
		{Name: "commands", Linter: LintCommands},
		{Name: "skills", Linter: LintSkills},
		{Name: "settings", Linter: LintSettings},
		{Name: "rules", Linter: LintRules},
		{Name: "output-styles", Linter: LintOutputStyles},
		// {Name: "plugins", Linter: LintPlugins}, // TODO: re-enable when output is less overwhelming
	}
}

// OrchestratorConfig holds configuration for the lint orchestrator.
type OrchestratorConfig struct {
	RootPath       string
	UseBaseline    bool
	CreateBaseline bool
	BaselinePath   string
}

// Orchestrator coordinates the linting process across all component types.
type Orchestrator struct {
	cfg     *config.Config
	opts    OrchestratorConfig
	linters []LinterEntry
}

// LinterEntry pairs a component name with its linter function.
type LinterEntry struct {
	Name   string
	Linter LinterFunc
}

// NewOrchestrator creates a new lint orchestrator.
func NewOrchestrator(cfg *config.Config, opts OrchestratorConfig) *Orchestrator {
	return &Orchestrator{
		cfg:     cfg,
		opts:    opts,
		linters: DefaultLinters(),
	}
}

// WithLinters allows customizing which linters to run.
func (o *Orchestrator) WithLinters(linters []LinterEntry) *Orchestrator {
	o.linters = linters
	return o
}

// Result holds the outcome of a lint run.
type Result struct {
	StartTime          time.Time
	TotalFiles         int
	TotalErrors        int
	TotalWarnings      int
	TotalSuggestions   int
	HasErrors          bool
	BaselineIgnored    int
	ErrorsIgnored      int
	SuggestionsIgnored int
	Summaries          []*LintSummary
}

// Run executes the full lint workflow.
func (o *Orchestrator) Run() (*Result, error) {
	startTime := time.Now()

	// Resolve baseline path relative to project root
	baselineFile := o.resolveBaselinePath()

	// Load baseline if requested
	b, err := o.loadBaseline(baselineFile)
	if err != nil && !o.cfg.Quiet {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load baseline: %v\n", err)
	}

	// Track totals across all linters
	result := &Result{StartTime: startTime}

	// Run all linters and collect summaries
	allIssues, _, errs := o.runAllLinters(b, result)
	if errs != nil {
		return nil, errs
	}

	// Run project-wide memory checks
	o.runMemoryChecks()

	// Create/update baseline if requested
	if o.opts.CreateBaseline {
		if err := o.saveBaseline(allIssues, baselineFile); err != nil {
			return nil, err
		}
		// When creating baseline, exit successfully to accept current state
		result.HasErrors = false
		return result, nil
	}

	// Note: Output formatting and summary printing is handled by cmd layer
	// to avoid import cycles with the output package

	return result, nil
}

// runAllLinters runs all configured linters and collects results.
func (o *Orchestrator) runAllLinters(b *baseline.Baseline, result *Result) ([]cue.ValidationError, []*LintSummary, error) {
	var allIssues []cue.ValidationError
	var allSummaries []*LintSummary

	for _, l := range o.linters {
		summary, err := l.Linter(o.cfg.Root, o.cfg.Quiet, o.cfg.Verbose, o.cfg.NoCycleCheck)
		if err != nil {
			return nil, nil, fmt.Errorf("error running %s linter: %w", l.Name, err)
		}

		// Skip empty results (no files of this type)
		if summary.TotalFiles == 0 {
			continue
		}

		// Collect issues for baseline creation
		if o.opts.CreateBaseline {
			allIssues = append(allIssues, CollectAllIssues(summary)...)
		}

		// Filter with baseline if active
		if o.opts.UseBaseline && b != nil {
			ignored, errIgnored, suggIgnored := FilterResults(summary, b)
			result.BaselineIgnored += ignored
			result.ErrorsIgnored += errIgnored
			result.SuggestionsIgnored += suggIgnored
		}

		// Collect summary for compact output
		allSummaries = append(allSummaries, summary)

		// Accumulate totals
		result.TotalFiles += summary.TotalFiles
		result.TotalErrors += summary.TotalErrors
		result.TotalWarnings += summary.TotalWarnings
		result.TotalSuggestions += summary.TotalSuggestions
		if summary.TotalErrors > 0 {
			result.HasErrors = true
		}
		result.Summaries = allSummaries

		// Progressive output in verbose mode
		if o.cfg.Verbose && !o.cfg.Quiet {
			status := "✓"
			if summary.TotalErrors > 0 {
				status = "✗"
			}
			fmt.Fprintf(os.Stderr, "  %s %s: %d files\n", status, l.Name, summary.TotalFiles)
		}
	}

	return allIssues, allSummaries, nil
}

// resolveBaselinePath returns the absolute path to the baseline file.
func (o *Orchestrator) resolveBaselinePath() string {
	baselineFile := o.opts.BaselinePath
	if !filepath.IsAbs(baselineFile) {
		baselineFile = filepath.Join(o.cfg.Root, baselineFile)
	}
	return baselineFile
}

// loadBaseline loads the baseline file if baseline mode is enabled.
func (o *Orchestrator) loadBaseline(baselineFile string) (*baseline.Baseline, error) {
	if !o.opts.UseBaseline && !o.opts.CreateBaseline {
		return nil, nil
	}

	if _, err := os.Stat(baselineFile); err != nil {
		return nil, nil // File doesn't exist, not an error
	}

	return baseline.LoadBaseline(baselineFile)
}

// saveBaseline creates and saves a new baseline from the collected issues.
func (o *Orchestrator) saveBaseline(issues []cue.ValidationError, baselineFile string) error {
	b := baseline.CreateBaseline(issues)
	b.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := b.SaveBaseline(baselineFile); err != nil {
		return fmt.Errorf("failed to save baseline: %w", err)
	}

	if !o.cfg.Quiet {
		fmt.Printf("\nBaseline created: %s (%d issues)\n", baselineFile, len(b.Fingerprints))
	}

	return nil
}

// runMemoryChecks performs project-wide memory checks.
func (o *Orchestrator) runMemoryChecks() {
	if o.cfg.Quiet {
		return
	}

	// Check CLAUDE.local.md gitignore
	// Output to stderr to avoid corrupting JSON/markdown stdout output
	gitignoreWarnings := CheckClaudeLocalGitignore(o.cfg.Root)
	for _, w := range gitignoreWarnings {
		fmt.Fprintf(os.Stderr, "warning: %s: %s\n", w.File, w.Message)
	}

	// Check combined memory size
	fd := discovery.NewFileDiscovery(o.cfg.Root, false)
	allFiles, _ := fd.DiscoverFiles()
	sizeWarnings := CheckCombinedMemorySize(o.cfg.Root, allFiles)
	for _, w := range sizeWarnings {
		fmt.Fprintf(os.Stderr, "warning: %s\n", w.Message)
	}
}
