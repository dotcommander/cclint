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
// Source: https://docs.anthropic.com/en/docs/claude-code/sub-agents
var knownAgentFields = map[string]bool{
	"name":           true, // Required: unique identifier
	"description":    true, // Required: what the agent does
	"model":          true, // Optional: sonnet, opus, haiku, etc.
	"color":          true, // Optional: display color in UI
	"tools":          true, // Optional: tool access
	"permissionMode": true, // Optional: permission handling
	"skills":         true, // Optional: comma-separated skill names
}

// validateAgentSpecific implements agent-specific validation rules
func validateAgentSpecific(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check for unknown frontmatter fields - helps catch fabricated/deprecated fields
	for key := range data {
		if !knownAgentFields[key] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown frontmatter field '%s'. Valid fields: name, description, model, color, tools, permissionMode, skills", key),
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

	// Validate tool field naming (agents use 'tools:', not 'allowed-tools:')
	errors = append(errors, ValidateToolFieldName(data, filePath, contents, "agent")...)

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

// validateAgentBestPractices checks opinionated best practices for agents
func validateAgentBestPractices(filePath string, contents string, data map[string]interface{}) []cue.ValidationError {
	var suggestions []cue.ValidationError
	fmEndLine := GetFrontmatterEndLine(contents)

	// XML tag detection in text fields - FROM ANTHROPIC DOCS
	if description, ok := data["description"].(string); ok {
		if xmlErr := DetectXMLTags(description, "Description", filePath, contents); xmlErr != nil {
			suggestions = append(suggestions, *xmlErr)
		}
	}

	// Count total lines (±10% tolerance: 200 base)
	if sizeErr := CheckSizeLimit(contents, 200, 0.10, "agent", filePath); sizeErr != nil {
		suggestions = append(suggestions, *sizeErr)
	}

	// === BLOAT SECTIONS DETECTOR ===
	// Check for bloat sections (only match exact h2 headings, not substrings)
	bloatPatterns := []struct {
		regex   *regexp.Regexp
		message string
	}{
		{regexp.MustCompile(`(?m)^## Quick Reference\s*$`), "Agent has '## Quick Reference' - belongs in skill, not agent"},
		{regexp.MustCompile(`(?m)^## When to Use\s*$`), "Agent has '## When to Use' - caller decides, use description triggers"},
		{regexp.MustCompile(`(?m)^## What it does\s*$`), "Agent has '## What it does' - belongs in description"},
		{regexp.MustCompile(`(?m)^## Usage\s*$`), "Agent has '## Usage' - belongs in skill or remove"},
	}
	for _, bp := range bloatPatterns {
		if bp.regex.MatchString(contents) {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  bp.message,
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	// === INLINE METHODOLOGY DETECTOR ===
	inlinePatterns := []struct {
		pattern string
		message string
	}{
		{`score\s*=\s*\([^)]{20,}`, "Inline scoring formula detected - should be 'See skill for scoring'"},
		{`\|\s*CRITICAL\s*\|[^|]*\|\s*HIGH\s*\|`, "Inline priority matrix detected - move to skill"},
		{`(?i)tier\s*(bonus|1|2|3|4)[^|]*\+\s*\d+`, "Tier scoring details inline - move to skill"},
		{`regexp\.(MustCompile|Compile)\s*\(`, "Detection patterns inline - move to skill"},
	}
	for _, ip := range inlinePatterns {
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

	// Check for missing model specification - OUR OBSERVATION (Anthropic docs say model is optional)
	if _, hasModel := data["model"]; !hasModel {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent lacks 'model' specification. Consider adding 'model: sonnet' or appropriate model for optimal performance.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     fmEndLine,
		})
	}

	// Check for Skill loading pattern - OUR OBSERVATION (thin agent -> fat skill pattern)
	// Accept Skill() tool calls, Skill: references, and Skills: (plural) references
	hasSkillRef := strings.Contains(contents, "Skill(") || strings.Contains(contents, "Skill:") || strings.Contains(contents, "Skills:")
	if !hasSkillRef && (strings.Contains(contents, "## Foundation") || strings.Contains(contents, "## Workflow")) {
		// Has structure but no skill reference - gentle reminder
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "No skill reference found. If methodology is reusable, consider extracting to a skill.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     FindSectionLine(contents, "Foundation"),
		})
	}

	// Check for Use PROACTIVELY pattern in description - Anthropic docs mention this as a tip
	if desc, hasDesc := data["description"].(string); hasDesc {
		if !strings.Contains(desc, "PROACTIVELY") {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  "Description lacks 'Use PROACTIVELY when...' pattern. Add to clarify activation scenarios.",
				Severity: "suggestion",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "description"),
			})
		}
	}

	// Check for permissionMode when agent has editing tools - OUR OBSERVATION
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