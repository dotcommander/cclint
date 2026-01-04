// Package lint provides the core linting orchestration logic.
package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/output"
)

// LinterFunc is the function signature for component linters.
type LinterFunc func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error)

// ComponentLinter pairs a component name with its linter function.
type ComponentLinter struct {
	Name   string
	Linter LinterFunc
}

// DefaultLinters returns the standard set of component linters.
func DefaultLinters() []ComponentLinter {
	return []ComponentLinter{
		{Name: "agents", Linter: cli.LintAgents},
		{Name: "commands", Linter: cli.LintCommands},
		{Name: "skills", Linter: cli.LintSkills},
		{Name: "settings", Linter: cli.LintSettings},
		{Name: "context", Linter: cli.LintContext},
		{Name: "rules", Linter: cli.LintRules},
		// {Name: "plugins", Linter: cli.LintPlugins}, // TODO: re-enable when output is less overwhelming
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
	linters []ComponentLinter
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
func (o *Orchestrator) WithLinters(linters []ComponentLinter) *Orchestrator {
	o.linters = linters
	return o
}

// Result holds the outcome of a lint run.
type Result struct {
	TotalFiles        int
	TotalErrors       int
	TotalSuggestions  int
	HasErrors         bool
	BaselineIgnored   int
	ErrorsIgnored     int
	SuggestionsIgnored int
}

// Run executes the full lint workflow.
func (o *Orchestrator) Run() (*Result, error) {
	// Resolve baseline path relative to project root
	baselineFile := o.resolveBaselinePath()

	// Load baseline if requested
	b, err := o.loadBaseline(baselineFile)
	if err != nil && !o.cfg.Quiet {
		fmt.Fprintf(os.Stderr, "Warning: Failed to load baseline: %v\n", err)
	}

	// Track totals across all linters
	result := &Result{}
	var allIssues []cue.ValidationError
	var allSummaries []*cli.LintSummary

	// Run all linters and collect summaries
	for _, l := range o.linters {
		summary, err := l.Linter(o.cfg.Root, o.cfg.Quiet, o.cfg.Verbose, o.cfg.NoCycleCheck)
		if err != nil {
			return nil, fmt.Errorf("error running %s linter: %w", l.Name, err)
		}

		// Skip empty results (no files of this type)
		if summary.TotalFiles == 0 {
			continue
		}

		// Collect issues for baseline creation
		if o.opts.CreateBaseline {
			allIssues = append(allIssues, cli.CollectAllIssues(summary)...)
		}

		// Filter with baseline if active
		if o.opts.UseBaseline && b != nil {
			ignored, errIgnored, suggIgnored := cli.FilterResults(summary, b)
			result.BaselineIgnored += ignored
			result.ErrorsIgnored += errIgnored
			result.SuggestionsIgnored += suggIgnored
		}

		// Collect summary for compact output
		allSummaries = append(allSummaries, summary)

		// Accumulate totals
		result.TotalFiles += summary.TotalFiles
		result.TotalErrors += summary.TotalErrors
		result.TotalSuggestions += summary.TotalSuggestions
		if summary.TotalErrors > 0 {
			result.HasErrors = true
		}
	}

	// Output all results using compact formatter (for console)
	if o.cfg.Format == "console" && !o.cfg.Quiet {
		formatter := output.NewCompactFormatter(o.cfg.Quiet, o.cfg.Verbose, o.cfg.ShowScores, o.cfg.ShowImprovements)
		if err := formatter.FormatAll(allSummaries); err != nil {
			return nil, fmt.Errorf("error formatting output: %w", err)
		}
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

	// Print baseline filtering summary
	if o.opts.UseBaseline && b != nil && result.BaselineIgnored > 0 && !o.cfg.Quiet {
		fmt.Printf("\n%d baseline issues ignored (%d errors, %d suggestions)\n",
			result.BaselineIgnored, result.ErrorsIgnored, result.SuggestionsIgnored)
	}

	// Print validation reminder (unless quiet mode)
	if !o.cfg.Quiet {
		fmt.Println("\n  Validate suggestions against docs.anthropic.com or docs.claude.com")
	}

	return result, nil
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
	gitignoreWarnings := cli.CheckClaudeLocalGitignore(o.cfg.Root)
	for _, w := range gitignoreWarnings {
		fmt.Printf("warning: %s: %s\n", w.File, w.Message)
	}

	// Check combined memory size
	fd := discovery.NewFileDiscovery(o.cfg.Root, false)
	allFiles, _ := fd.DiscoverFiles()
	sizeWarnings := cli.CheckCombinedMemorySize(o.cfg.Root, allFiles)
	for _, w := range sizeWarnings {
		fmt.Printf("warning: %s\n", w.Message)
	}
}
