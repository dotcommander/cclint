package cmd

import (
	"fmt"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/lint"
)

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
