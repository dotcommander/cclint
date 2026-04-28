package lint

import (
	"time"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/scoring"
	"github.com/dotcommander/cclint/internal/textutil"
)

// LintResult represents a single linting result
type LintResult struct {
	File         string
	Type         string
	Errors       []cue.ValidationError
	Warnings     []cue.ValidationError
	Suggestions  []cue.ValidationError
	Improvements []textutil.ImprovementRecommendation
	Success      bool
	Duration     int64
	Quality      *scoring.QualityScore
}

// LintSummary summarizes all linting results
type LintSummary struct {
	ProjectRoot      string
	ComponentType    string // e.g., "agents", "commands", "skills"
	StartTime        time.Time
	TotalFiles       int
	SuccessfulFiles  int
	FailedFiles      int
	TotalErrors      int
	TotalWarnings    int
	TotalSuggestions int
	Duration         int64
	Results          []LintResult
}

// LintAgents runs linting on agent files using the generic linter.
func LintAgents(rootPath string, quiet bool, verbose bool, noCycleCheck bool, exclude []string) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck, exclude)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewAgentLinter()), nil
}

// knownAgentFields lists valid frontmatter fields per Anthropic docs
// Source: https://code.claude.com/docs/en/sub-agents
var knownAgentFields = map[string]bool{
	"name":            true, // Required: unique identifier
	"description":     true, // Required: what the agent does
	"model":           true, // Optional: sonnet, opus, haiku, inherit
	"color":           true, // Optional: display color in UI (set via /agents wizard)
	"tools":           true, // Optional: tool access allowlist
	"disallowedTools": true, // Optional: tool access denylist
	"permissionMode":  true, // Optional: default, acceptEdits, delegate, dontAsk, bypassPermissions, plan
	"maxTurns":        true, // Optional: max conversation turns (positive integer)
	"skills":          true, // Optional: skills to preload into context
	"hooks":           true, // Optional: agent-level hooks (PreToolUse, PostToolUse, Stop)
	"memory":          true, // Optional: persistent memory scope (user, project, local) (v2.1.33+)
	"mcpServers":      true, // Optional: MCP server names available to agent
	"isolation":       true, // Optional: subagent isolation mode (worktree) (v2.1.49+)
	"background":      true, // Optional: always run as background task (v2.1.49+)
	"effort":          true, // Optional: reasoning effort level (v2.1.78+)
	"initialPrompt":   true, // Optional: auto-submit first turn (v2.1.83+)
	"triggers":        true,
}

// validateAgentSpecific implements agent-specific validation rules.
// Orchestrates validation by delegating to focused check functions.
func validateAgentSpecific(data map[string]any, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Frontmatter field validation
	errors = append(errors, validateUnknownFields(data, filePath, contents)...)
	errors = append(errors, validateRequiredFields(data, filePath, contents)...)

	// Individual field validation
	errors = append(errors, validateAgentColor(data, filePath)...)
	errors = append(errors, validateAgentMemory(data, filePath, contents)...)
	errors = append(errors, validateAgentModel(data, filePath, contents)...)
	errors = append(errors, validateAgentMCPServersField(data, filePath, contents)...)
	errors = append(errors, validateAgentPermissionMode(data, filePath, contents)...)
	errors = append(errors, validateAgentMaxTurns(data, filePath, contents)...)
	errors = append(errors, validateAgentAutonomousPattern(data, filePath, contents)...)

	// Cross-field validation
	errors = append(errors, textutil.ValidateToolFieldName(data, filePath, contents, "agent")...)
	errors = append(errors, validateAgentHooks(data, filePath)...)
	errors = append(errors, validateAgentBestPractices(filePath, contents, data)...)
	errors = append(errors, validateBodyToolMismatch(data, filePath, contents)...)

	return errors
}
