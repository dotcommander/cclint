package lint

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
)

func agentResultIndex(summary *LintSummary) map[string]int {
	byName := make(map[string]int, len(summary.Results))
	for i, result := range summary.Results {
		name := crossfile.ExtractAgentName(result.File)
		if _, exists := byName[name]; !exists {
			byName[name] = i
		}
	}
	return byName
}

func appendAgentCycle(summary *LintSummary, i int, cycleDesc string) {
	result := &summary.Results[i]
	result.Errors = append(result.Errors, cue.ValidationError{
		File:     result.File,
		Message:  fmt.Sprintf("Circular dependency detected: %s", cycleDesc),
		Severity: cue.SeverityError,
		Source:   cue.SourceCClintObserve,
	})
	summary.TotalErrors++
	if result.Success {
		result.Success = false
		summary.SuccessfulFiles--
		summary.FailedFiles++
	}
}

// Skill cycle attachment visits every matching path (unlike agent cycles,
// whose first matching name wins) and deduplicates by existing error message.
func attachSkillCycles(summary *LintSummary, cycles []crossfile.Cycle) {
	byName := make(map[string][]int, len(summary.Results))
	for i, result := range summary.Results {
		name := filepath.Base(filepath.Dir(result.File))
		byName[name] = append(byName[name], i)
	}
	seen := make(map[int]map[string]bool)
	for _, cycle := range cycles {
		message := "Circular dependency detected: " + crossfile.FormatCycle(cycle)
		for _, node := range cycle.Path {
			parts := strings.SplitN(node, ":", 2)
			if len(parts) != 2 || parts[0] != cue.TypeSkill {
				continue
			}
			for _, i := range byName[parts[1]] {
				if seen[i] == nil {
					seen[i] = make(map[string]bool, len(summary.Results[i].Errors))
					for _, existing := range summary.Results[i].Errors {
						seen[i][existing.Message] = true
					}
				}
				if seen[i][message] {
					continue
				}
				result := &summary.Results[i]
				categorizeIssues(result, []cue.ValidationError{{File: result.File, Message: message, Severity: cue.SeverityError, Source: cue.SourceCClintObserve}})
				seen[i][message] = true
			}
		}
	}
}
