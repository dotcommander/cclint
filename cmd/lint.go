package cmd

import (
	"fmt"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/lint"
)

// LinterFunc is the function signature for component linters.
// It takes root path, quiet mode, verbose mode, noCycleCheck and returns a summary.
type LinterFunc = lint.LinterFunc

// typeLinters maps file types to their linter name and function.
var typeLinters = map[discovery.FileType]struct {
	Name   string
	Linter LinterFunc
}{
	discovery.FileTypeAgent:       {"agents", lint.LintAgents},
	discovery.FileTypeCommand:     {"commands", lint.LintCommands},
	discovery.FileTypeSkill:       {"skills", lint.LintSkills},
	discovery.FileTypeSettings:    {"settings", lint.LintSettings},
	discovery.FileTypeContext:     {"context", lint.LintContext},
	discovery.FileTypePlugin:      {"plugins", lint.LintPlugins},
	discovery.FileTypeRule:        {"rules", lint.LintRules},
	discovery.FileTypeOutputStyle: {"output-styles", lint.LintOutputStyles},
}

// runTypeLint runs the linter for a specific file type.
func runTypeLint(ft discovery.FileType) error {
	entry, ok := typeLinters[ft]
	if !ok {
		return fmt.Errorf("no linter for type %s", ft)
	}
	return runComponentLint(entry.Name, entry.Linter)
}

// runComponentLint is the generic function that handles config loading,
// linter execution, and output formatting for any component type.
// This follows the Single Responsibility Principle by separating
// orchestration from component-specific linting logic.
func runComponentLint(linterName string, linter LinterFunc) error {
	cfg, err := loadCLIConfig()
	if err != nil {
		return err
	}

	result, err := runOrchestratedLint(cfg, []lint.LinterEntry{{
		Name:   linterName,
		Linter: linter,
	}})
	if err != nil {
		return fmt.Errorf("error running %s linter: %w", linterName, err)
	}

	summary := &lint.LintSummary{}
	if len(result.Summaries) > 0 {
		summary = result.Summaries[0]
	}

	if err := formatSummaryOutput(cfg, summary); err != nil {
		return err
	}

	printBaselineSummary(result.BaselineIgnored, result.ErrorsIgnored, result.SuggestionsIgnored, cfg.Quiet)
	printValidationReminder(cfg)
	applyFailurePolicy(cfg, summary.TotalErrors, summary.TotalWarnings, summary.TotalSuggestions)

	return nil
}
