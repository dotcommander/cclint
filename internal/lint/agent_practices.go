package lint

import (
	"regexp"
	"slices"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/textutil"
)

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

// inlineMethodologyPatterns detect methodology that should be in skills.
var inlineMethodologyPatterns = []struct {
	regex   *regexp.Regexp
	message string
}{
	{regexp.MustCompile(`score\s*=\s*\([^)]{20,}`), "Inline scoring formula detected - should be 'See skill for scoring'"},
	{regexp.MustCompile(`\|\s*CRITICAL\s*\|[^|]*\|\s*HIGH\s*\|`), "Inline priority matrix detected - move to skill"},
	{regexp.MustCompile(`(?i)tier\s*(bonus|1|2|3|4)[^|]*\+\s*\d+`), "Tier scoring details inline - move to skill"},
	{regexp.MustCompile(`regexp\.(MustCompile|Compile)\s*\(`), "Detection patterns inline - move to skill"},
}

// validateAgentBestPractices checks opinionated best practices for agents.
// Aggregates results from focused validation functions for each concern.
func validateAgentBestPractices(filePath string, contents string, data map[string]any) []cue.ValidationError {
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
func checkAgentXMLTags(data map[string]any, filePath, contents string) []cue.ValidationError {
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

// checkAgentInlineMethodology detects methodology patterns that belong in skills.
func checkAgentInlineMethodology(contents, filePath string) []cue.ValidationError {
	var suggestions []cue.ValidationError
	for _, ip := range inlineMethodologyPatterns {
		if ip.regex.MatchString(contents) {
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
func checkAgentMissingFields(data map[string]any, contents, filePath string) []cue.ValidationError {
	var suggestions []cue.ValidationError
	fmEndLine := textutil.GetFrontmatterEndLine(contents)

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
	// Only emit this suggestion if the agent has the Skill tool available.
	hasSkillRef := strings.Contains(contents, "Skill(") || strings.Contains(contents, "Skill:") || strings.Contains(contents, "Skills:")
	if !hasSkillRef && hasSkillTool(data["tools"]) && (strings.Contains(contents, "## Foundation") || strings.Contains(contents, "## Workflow")) {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "No skill reference found. If methodology is reusable, consider extracting to a skill.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     textutil.FindSectionLine(contents, "Foundation"),
		})
	}

	// Check for permissionMode when agent has editing tools
	if _, hasPermMode := data["permissionMode"]; !hasPermMode && hasEditingTools(data["tools"]) {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent has editing tools but no permissionMode. Consider 'permissionMode: acceptEdits' for seamless file edits.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     textutil.FindFrontmatterFieldLine(contents, "tools"),
		})
	}

	return suggestions
}

// hasSkillTool checks if the tools field includes the Skill tool or is "*".
func hasSkillTool(tools any) bool {
	switch v := tools.(type) {
	case string:
		if v == "*" {
			return true
		}
		for _, part := range strings.Split(v, ",") {
			if strings.TrimSpace(part) == "Skill" {
				return true
			}
		}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s == "Skill" {
				return true
			}
		}
	}
	return false
}

// hasEditingTools checks if the tools field includes editing capabilities
func hasEditingTools(tools any) bool {
	editingTools := []string{"Edit", "Write", "MultiEdit"}

	switch v := tools.(type) {
	case string:
		if v == "*" {
			return true
		}
		for _, part := range strings.Split(v, ",") {
			if slices.Contains(editingTools, strings.TrimSpace(part)) {
				return true
			}
		}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && slices.Contains(editingTools, s) {
				return true
			}
		}
	}
	return false
}
