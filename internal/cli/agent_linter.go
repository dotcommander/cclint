package cli

import (
	"fmt"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/scoring"
)

// AgentLinter implements ComponentLinter for agent files.
// It also implements optional interfaces for cross-file validation,
// scoring, improvements, and batch post-processing (cycle detection).
type AgentLinter struct {
	BaseLinter
}

// Compile-time interface compliance checks
var (
	_ ComponentLinter      = (*AgentLinter)(nil)
	_ CrossFileValidatable = (*AgentLinter)(nil)
	_ Scorable             = (*AgentLinter)(nil)
	_ Improvable           = (*AgentLinter)(nil)
	_ BatchPostProcessor   = (*AgentLinter)(nil)
)

// NewAgentLinter creates a new AgentLinter.
func NewAgentLinter() *AgentLinter {
	return &AgentLinter{}
}

func (l *AgentLinter) Type() string {
	return cue.TypeAgent
}

func (l *AgentLinter) FileType() discovery.FileType {
	return discovery.FileTypeAgent
}

func (l *AgentLinter) ParseContent(contents string) (map[string]any, string, error) {
	return parseFrontmatter(contents)
}

func (l *AgentLinter) ValidateCUE(validator *cue.Validator, data map[string]any) ([]cue.ValidationError, error) {
	return validator.ValidateAgent(data)
}

func (l *AgentLinter) ValidateSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	errors := validateAgentSpecific(data, filePath, contents)

	// Validate allowed-tools
	toolWarnings := ValidateAllowedTools(data, filePath, contents)
	errors = append(errors, toolWarnings...)

	return errors
}

// ValidateCrossFile implements CrossFileValidatable interface
func (l *AgentLinter) ValidateCrossFile(crossValidator *CrossFileValidator, filePath, contents string, data map[string]any) []cue.ValidationError {
	if crossValidator == nil {
		return nil
	}
	return crossValidator.ValidateAgent(filePath, contents, data)
}

// Score implements Scorable interface
func (l *AgentLinter) Score(contents string, data map[string]any, body string) *scoring.QualityScore {
	scorer := scoring.NewAgentScorer()
	score := scorer.Score(contents, data, body)
	return &score
}

// GetImprovements implements Improvable interface
func (l *AgentLinter) GetImprovements(contents string, data map[string]any) []ImprovementRecommendation {
	return GetAgentImprovements(contents, data)
}

// PostProcessBatch implements BatchPostProcessor for cycle detection.
func (l *AgentLinter) PostProcessBatch(ctx *LinterContext, summary *LintSummary) {
	if ctx.NoCycleCheck {
		return
	}

	cycles := ctx.CrossValidator.DetectCycles()
	cyclesReported := make(map[string]bool)

	for _, cycle := range cycles {
		cycleDesc := FormatCycle(cycle)
		if cyclesReported[cycleDesc] {
			continue
		}
		cyclesReported[cycleDesc] = true

		// Report cycle errors to all agents in the cycle
		l.reportCycleError(summary, cycle, cycleDesc)
	}
}

// reportCycleError reports a cycle error to all agents involved in the cycle.
func (l *AgentLinter) reportCycleError(summary *LintSummary, cycle Cycle, cycleDesc string) {
	agentsInCycle := extractAgentsFromCycle(cycle.Path)

	for agentName := range agentsInCycle {
		addCycleToSummary(summary, agentName, cycleDesc)
	}
}

// extractAgentsFromCycle extracts agent names from cycle path nodes.
func extractAgentsFromCycle(path []string) map[string]bool {
	agentsInCycle := make(map[string]bool)
	for _, node := range path {
		parts := strings.SplitN(node, ":", 2)
		if len(parts) == 2 && parts[0] == cue.TypeAgent {
			agentsInCycle[parts[1]] = true
		}
	}
	return agentsInCycle
}

// addCycleToSummary adds a cycle error to the summary for a specific agent.
func addCycleToSummary(summary *LintSummary, agentName, cycleDesc string) {
	for i := range summary.Results {
		resultName := crossExtractAgentName(summary.Results[i].File)
		if resultName == agentName {
			summary.Results[i].Errors = append(summary.Results[i].Errors, cue.ValidationError{
				File:     summary.Results[i].File,
				Message:  fmt.Sprintf("Circular dependency detected: %s", cycleDesc),
				Severity: "error",
				Source:   cue.SourceCClintObserve,
			})
			summary.TotalErrors++
			if summary.Results[i].Success {
				summary.Results[i].Success = false
				summary.SuccessfulFiles--
				summary.FailedFiles++
			}
			break
		}
	}
}
