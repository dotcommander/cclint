package cli

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/scoring"
)

// LintResult represents a single linting result
type LintResult struct {
	File         string
	Type         string
	Errors       []cue.ValidationError
	Warnings     []cue.ValidationError
	Suggestions  []cue.ValidationError
	Improvements []ImprovementRecommendation
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
func LintAgents(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
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
}

// validateAgentSpecific implements agent-specific validation rules
func validateAgentSpecific(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check for unknown frontmatter fields - helps catch fabricated/deprecated fields
	for key := range data {
		if !knownAgentFields[key] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown frontmatter field '%s'. Valid fields: name, description, model, color, tools, disallowedTools, permissionMode, maxTurns, skills, hooks, memory, mcpServers", key),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
				Line:     FindFrontmatterFieldLine(contents, key),
			})
		}
	}

	// Check required fields - FROM ANTHROPIC DOCS: "name" and "description" are Required
	if name, ok := data["name"].(string); !ok || name == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'name' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "name"),
		})
	}

	if description, ok := data["description"].(string); !ok || description == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'description' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "description"),
		})
	} else if !strings.Contains(description, "PROACTIVELY") {
		// Check for PROACTIVELY pattern - REQUIRED for agent discoverability
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Description MUST include 'Use PROACTIVELY when...' pattern for agent discoverability.",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "description"),
		})
	}

	// Check name format - FROM ANTHROPIC DOCS: "Unique identifier using lowercase letters and hyphens"
	if name, ok := data["name"].(string); ok {
		valid := true
		for _, c := range name {
			if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
				valid = false
				break
			}
		}
		if !valid {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  "Name must contain only lowercase letters, numbers, and hyphens",
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "name"),
			})
		}

		// Reserved word check - FROM ANTHROPIC DOCS
		reservedWords := map[string]bool{"anthropic": true, "claude": true}
		if reservedWords[strings.ToLower(name)] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Name '%s' is a reserved word and cannot be used", name),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "name"),
			})
		}

		// Check if name matches filename - OUR OBSERVATION
		// Extract filename from path (e.g., ".claude/agents/test-agent.md" -> "test-agent")
		filename := filePath
		if idx := strings.LastIndex(filename, "/"); idx != -1 {
			filename = filename[idx+1:]
		}
		// Remove .md extension
		filename = strings.TrimSuffix(filename, ".md")
		// For nested paths, use the last component
		if idx := strings.LastIndex(filename, "/"); idx != -1 {
			filename = filename[idx+1:]
		}
		// Remove .md extension again if present
		filename = strings.TrimSuffix(filename, ".md")

		if name != filename {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Name %q doesn't match filename %q", name, filename),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	// Check valid colors - OUR OBSERVATION (not documented by Anthropic)
	if color, ok := data["color"].(string); ok {
		validColors := map[string]bool{
			"red":     true,
			"blue":    true,
			"green":   true,
			"yellow":  true,
			"purple":  true,
			"orange":  true,
			"pink":    true,
			"cyan":    true,
		}
		if !validColors[color] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Invalid color '%s'. Valid colors are: red, blue, green, yellow, purple, orange, pink, cyan", color),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	// Validate memory scope (v2.1.33+)
	if memory, ok := data["memory"].(string); ok {
		validMemoryScopes := map[string]bool{
			"user":    true,
			"project": true,
			"local":   true,
		}
		if !validMemoryScopes[memory] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Invalid memory scope '%s'. Valid scopes: user, project, local", memory),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "memory"),
			})
		}
	}

	// Validate mcpServers - must be an array of non-empty strings
	if mcpServers, ok := data["mcpServers"]; ok {
		if arr, isArr := mcpServers.([]interface{}); isArr {
			for i, item := range arr {
				s, isStr := item.(string)
				if !isStr || s == "" {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("mcpServers[%d] must be a non-empty string", i),
						Severity: "error",
						Source:   cue.SourceAnthropicDocs,
						Line:     FindFrontmatterFieldLine(contents, "mcpServers"),
					})
				}
			}
		} else {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  "mcpServers must be an array of server name strings",
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "mcpServers"),
			})
		}
	}

	// Validate permissionMode enum
	if permMode, ok := data["permissionMode"].(string); ok {
		validPermModes := map[string]bool{
			"default":           true,
			"acceptEdits":       true,
			"delegate":          true,
			"dontAsk":           true,
			"bypassPermissions": true,
			"plan":              true,
		}
		if !validPermModes[permMode] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Invalid permissionMode value %q; must be one of: default, acceptEdits, delegate, dontAsk, bypassPermissions, plan", permMode),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "permissionMode"),
			})
		}
	}

	// Validate maxTurns - must be a positive integer
	if maxTurns, ok := data["maxTurns"]; ok {
		switch v := maxTurns.(type) {
		case int:
			if v <= 0 {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Invalid maxTurns value %d; must be a positive integer", v),
					Severity: "error",
					Source:   cue.SourceAnthropicDocs,
					Line:     FindFrontmatterFieldLine(contents, "maxTurns"),
				})
			}
		case float64:
			if v <= 0 || v != float64(int(v)) {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Invalid maxTurns value %v; must be a positive integer", v),
					Severity: "error",
					Source:   cue.SourceAnthropicDocs,
					Line:     FindFrontmatterFieldLine(contents, "maxTurns"),
				})
			}
		default:
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Invalid maxTurns value %v; must be a positive integer", maxTurns),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "maxTurns"),
			})
		}
	}

	// Validate tool field naming (agents use 'tools:', not 'allowed-tools:')
	errors = append(errors, ValidateToolFieldName(data, filePath, contents, "agent")...)

	// Validate hooks (scoped to component events: PreToolUse, PostToolUse, Stop)
	if hooks, ok := data["hooks"]; ok {
		errors = append(errors, ValidateComponentHooks(hooks, filePath)...)
	}

	// Best practice checks
	errors = append(errors, validateAgentBestPractices(filePath, contents, data)...)

	return errors
}

// hasEditingTools checks if the tools field includes editing capabilities
func hasEditingTools(tools interface{}) bool {
	editingTools := []string{"Edit", "Write", "MultiEdit"}

	switch v := tools.(type) {
	case string:
		if v == "*" {
			return true
		}
		// Check comma-separated string
		for _, tool := range editingTools {
			if strings.Contains(v, tool) {
				return true
			}
		}
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				for _, tool := range editingTools {
					if s == tool {
						return true
					}
				}
			}
		}
	}
	return false
}

// validateAgentBestPractices checks opinionated best practices for agents.
// Aggregates results from focused validation functions for each concern.
func validateAgentBestPractices(filePath string, contents string, data map[string]interface{}) []cue.ValidationError {
	var suggestions []cue.ValidationError

	// Each check function handles one concern
	suggestions = append(suggestions, checkAgentXMLTags(data, filePath, contents)...)
	suggestions = append(suggestions, checkAgentSizeLimit(contents, filePath)...)
	suggestions = append(suggestions, checkAgentBloatSections(contents, filePath)...)
	suggestions = append(suggestions, checkAgentInlineMethodology(contents, filePath)...)
	suggestions = append(suggestions, checkAgentMissingFields(data, contents, filePath)...)

	return suggestions
}

// checkAgentXMLTags detects XML-like tags in description field.
// XML tags in agent descriptions can confuse Claude's parsing.
func checkAgentXMLTags(data map[string]interface{}, filePath, contents string) []cue.ValidationError {
	if description, ok := data["description"].(string); ok {
		if xmlErr := DetectXMLTags(description, "Description", filePath, contents); xmlErr != nil {
			return []cue.ValidationError{*xmlErr}
		}
	}
	return nil
}

// checkAgentSizeLimit ensures agents stay under recommended line count.
// Agents over 200 lines (+10% tolerance) should move methodology to skills.
func checkAgentSizeLimit(contents, filePath string) []cue.ValidationError {
	if sizeErr := CheckSizeLimit(contents, 200, 0.10, "agent", filePath); sizeErr != nil {
		return []cue.ValidationError{*sizeErr}
	}
	return nil
}

// bloatSectionPatterns are H2 headings that indicate content should be elsewhere.
var bloatSectionPatterns = []struct {
	regex   *regexp.Regexp
	message string
}{
	{regexp.MustCompile(`(?m)^## Quick Reference\s*$`), "Agent has '## Quick Reference' - belongs in skill, not agent"},
	{regexp.MustCompile(`(?m)^## When to Use\s*$`), "Agent has '## When to Use' - caller decides, use description triggers"},
	{regexp.MustCompile(`(?m)^## What it does\s*$`), "Agent has '## What it does' - belongs in description"},
	{regexp.MustCompile(`(?m)^## Usage\s*$`), "Agent has '## Usage' - belongs in skill or remove"},
}

// checkAgentBloatSections detects H2 sections that belong in skills, not agents.
func checkAgentBloatSections(contents, filePath string) []cue.ValidationError {
	var suggestions []cue.ValidationError
	for _, bp := range bloatSectionPatterns {
		if bp.regex.MatchString(contents) {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  bp.message,
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
			})
		}
	}
	return suggestions
}

// inlineMethodologyPatterns detect methodology that should be in skills.
var inlineMethodologyPatterns = []struct {
	pattern string
	message string
}{
	{`score\s*=\s*\([^)]{20,}`, "Inline scoring formula detected - should be 'See skill for scoring'"},
	{`\|\s*CRITICAL\s*\|[^|]*\|\s*HIGH\s*\|`, "Inline priority matrix detected - move to skill"},
	{`(?i)tier\s*(bonus|1|2|3|4)[^|]*\+\s*\d+`, "Tier scoring details inline - move to skill"},
	{`regexp\.(MustCompile|Compile)\s*\(`, "Detection patterns inline - move to skill"},
}

// checkAgentInlineMethodology detects methodology patterns that belong in skills.
func checkAgentInlineMethodology(contents, filePath string) []cue.ValidationError {
	var suggestions []cue.ValidationError
	for _, ip := range inlineMethodologyPatterns {
		matched, _ := regexp.MatchString(ip.pattern, contents)
		if matched {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  ip.message,
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
			})
		}
	}
	return suggestions
}

// checkAgentMissingFields checks for missing recommended fields and patterns.
func checkAgentMissingFields(data map[string]interface{}, contents, filePath string) []cue.ValidationError {
	var suggestions []cue.ValidationError
	fmEndLine := GetFrontmatterEndLine(contents)

	// Check for missing model specification
	if _, hasModel := data["model"]; !hasModel {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent lacks 'model' specification. Consider adding 'model: sonnet' or appropriate model for optimal performance.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     fmEndLine,
		})
	}

	// Check for Skill loading pattern (thin agent -> fat skill pattern)
	hasSkillRef := strings.Contains(contents, "Skill(") || strings.Contains(contents, "Skill:") || strings.Contains(contents, "Skills:")
	if !hasSkillRef && (strings.Contains(contents, "## Foundation") || strings.Contains(contents, "## Workflow")) {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "No skill reference found. If methodology is reusable, consider extracting to a skill.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     FindSectionLine(contents, "Foundation"),
		})
	}

	// Check for permissionMode when agent has editing tools
	if _, hasPermMode := data["permissionMode"]; !hasPermMode {
		if hasEditingTools(data["tools"]) {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  "Agent has editing tools but no permissionMode. Consider 'permissionMode: acceptEdits' for seamless file edits.",
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
				Line:     FindFrontmatterFieldLine(contents, "tools"),
			})
		}
	}

	return suggestions
}