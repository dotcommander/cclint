package cmd

import (
	"fmt"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/lint"
)

// runTypeLint runs the linter for a specific file type.
func runTypeLint(opts executionOptions, ft discovery.FileType) error {
	entry, ok := lint.LinterForType(ft)
	if !ok {
		return fmt.Errorf("no linter for type %s", ft)
	}
	return runComponentLint(opts, entry)
}

// runComponentLint is the generic function that handles config loading,
// linter execution, and output formatting for any component type.
// This follows the Single Responsibility Principle by separating
// orchestration from component-specific linting logic.
func runComponentLint(opts executionOptions, entry lint.LinterEntry) error {
	cfg, err := loadCLIConfig(opts)
	if err != nil {
		return err
	}

	result, err := runOrchestratedLint(cfg, opts, []lint.LinterEntry{entry})
	if err != nil {
		return fmt.Errorf("error running %s linter: %w", entry.Name, err)
	}

	if err := reportLintOutcome(cfg, opts, fullLintOutcome(result), func() error { return formatFullRunOutput(cfg, result) }); err != nil {
		return err
	}
	return nil
}
func runTypesLint(opts executionOptions, types []discovery.FileType) error {
	cfg, err := loadCLIConfig(opts)
	if err != nil {
		return err
	}
	var entries []lint.LinterEntry
	seen := map[discovery.FileType]bool{}
	for _, ft := range types {
		if seen[ft] {
			continue
		}
		seen[ft] = true
		entry, ok := lint.LinterForType(ft)
		if !ok {
			return fmt.Errorf("no linter for type %s", ft)
		}
		entries = append(entries, entry)
	}
	result, err := runOrchestratedLint(cfg, opts, entries)
	if err != nil {
		return err
	}
	return reportLintOutcome(cfg, opts, fullLintOutcome(result), func() error { return formatFullRunOutput(cfg, result) })
}
