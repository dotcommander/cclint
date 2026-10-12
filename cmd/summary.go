package cmd

import (
	"fmt"
	"sort"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/lint"
)

const (
	componentTypeAgent   = "agent"
	componentTypeCommand = "command"
	componentTypeSkill   = "skill"

	issueOversized          = "Oversized component (fat)"
	issueMissingFoundation  = "Missing Foundation section"
	issueMissingWorkflow    = "Missing Workflow section"
	issueMissingAntiPattern = "Missing Anti-Patterns section"
	issueMissingRouting     = "Missing semantic routing"
	issueMissingTriggers    = "Missing or incomplete triggers"
	issueMissingModel       = "Missing model specification"
	issueOther              = "Other issues"
)

// ComponentSummary holds aggregated data for summary report
type ComponentSummary struct {
	TotalComponents int
	AgentCount      int
	CommandCount    int
	SkillCount      int
	TierCounts      map[string]int
	TopIssues       map[string]int
	LowestScoring   []ScoredComponent
	AllResults      []lint.LintResult
}

// ScoredComponent represents a component with its score for sorting
type ScoredComponent struct {
	File  string
	Type  string
	Score int
	Tier  string
}

func runSummary(opts executionOptions) error {
	cfg, err := loadCLIConfig(opts)
	if err != nil {
		return err
	}

	entries := make([]lint.LinterEntry, 0, 3)
	for _, fileType := range []discovery.FileType{
		discovery.FileTypeAgent,
		discovery.FileTypeCommand,
		discovery.FileTypeSkill,
	} {
		entry, ok := lint.LinterForType(fileType)
		if !ok {
			return fmt.Errorf("no linter for type %s", fileType)
		}
		entries = append(entries, entry)
	}

	result, err := runOrchestratedLint(cfg, opts, entries)
	if err != nil {
		return fmt.Errorf("error building summary: %w", err)
	}

	summary := &ComponentSummary{
		TierCounts: make(map[string]int),
		TopIssues:  make(map[string]int),
	}

	for _, componentSummary := range result.Summaries {
		switch componentSummary.ComponentType {
		case componentTypeAgent:
			summary.AgentCount = componentSummary.TotalFiles
		case componentTypeCommand:
			summary.CommandCount = componentSummary.TotalFiles
		case componentTypeSkill:
			summary.SkillCount = componentSummary.TotalFiles
		}
		aggregateResults(summary, componentSummary.Results)
	}

	summary.TotalComponents = summary.AgentCount + summary.CommandCount + summary.SkillCount

	// Sort lowest scoring
	sort.Slice(summary.LowestScoring, func(i, j int) bool {
		return summary.LowestScoring[i].Score < summary.LowestScoring[j].Score
	})

	// Print summary report
	printSummaryReport(summary)

	return nil
}

func aggregateResults(summary *ComponentSummary, results []lint.LintResult) {
	for _, result := range results {
		summary.AllResults = append(summary.AllResults, result)

		if result.Quality != nil {
			summary.TierCounts[result.Quality.Tier]++
			summary.LowestScoring = append(summary.LowestScoring, ScoredComponent{
				File:  result.File,
				Type:  result.Type,
				Score: result.Quality.Overall,
				Tier:  result.Quality.Tier,
			})
		}

		// Aggregate issues
		for _, err := range result.Errors {
			summary.TopIssues[categorizeIssue(err.Message)]++
		}
		for _, sug := range result.Suggestions {
			summary.TopIssues[categorizeIssue(sug.Message)]++
		}
	}
}

func categorizeIssue(message string) string {
	// Categorize issues into buckets for aggregation
	switch {
	case contains(message, "lines", "Best practice"):
		return issueOversized
	case contains(message, "Foundation"):
		return issueMissingFoundation
	case contains(message, "Workflow"):
		return issueMissingWorkflow
	case contains(message, "Anti-Pattern"):
		return issueMissingAntiPattern
	case contains(message, "Quick Reference", "semantic routing"):
		return issueMissingRouting
	case contains(message, "Success Criteria"):
		return "Missing Success Criteria"
	case contains(message, "Expected Output"):
		return "Missing Expected Output"
	case contains(message, "Skill()", "methodology"):
		return "Embedded methodology (should extract)"
	case contains(message, "triggers"):
		return issueMissingTriggers
	case contains(message, "PROACTIVELY"):
		return "Missing PROACTIVELY pattern"
	case contains(message, "model"):
		return issueMissingModel
	case contains(message, "description"):
		return "Missing or poor description"
	default:
		return issueOther
	}
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}
