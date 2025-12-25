package cli

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
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

// LintAgents runs linting on agent files
func LintAgents(rootPath string, quiet bool, verbose bool) (*LintSummary, error) {
	// Initialize shared context
	ctx, err := NewLinterContext(rootPath, quiet, verbose)
	if err != nil {
		return nil, err
	}

	// Filter agent files
	agentFiles := ctx.FilterFilesByType(discovery.FileTypeAgent)
	summary := ctx.NewSummary(len(agentFiles))

	// Process each agent file
	for _, file := range agentFiles {
		result := LintResult{
			File:    file.RelPath,
			Type:    "agent",
			Success: true,
		}

		// Parse frontmatter
		fm, err := frontend.ParseYAMLFrontmatter(file.Contents)
		if err != nil {
			result.Errors = append(result.Errors, cue.ValidationError{
				File:     file.RelPath,
				Message:  fmt.Sprintf("Error parsing frontmatter: %v", err),
				Severity: "error",
			})
			result.Success = false
			summary.FailedFiles++
			summary.TotalErrors++
		} else {
			// Validate with CUE
			if true { // CUE schemas not loaded yet
				errors, err := ctx.Validator.ValidateAgent(fm.Data)
				if err != nil {
					result.Errors = append(result.Errors, cue.ValidationError{
						File:     file.RelPath,
						Message:  fmt.Sprintf("Validation error: %v", err),
						Severity: "error",
					})
				}
				result.Errors = append(result.Errors, errors...)
				summary.TotalErrors += len(errors)
			}

			// Additional validation rules - separate errors and suggestions
			allIssues := validateAgentSpecific(fm.Data, file.RelPath, file.Contents)
			for _, issue := range allIssues {
				if issue.Severity == "suggestion" {
					result.Suggestions = append(result.Suggestions, issue)
					summary.TotalSuggestions++
				} else {
					result.Errors = append(result.Errors, issue)
					summary.TotalErrors++
				}
			}

			// Cross-file validation (missing skills)
			crossErrors := ctx.CrossValidator.ValidateAgent(file.RelPath, file.Contents)
			result.Errors = append(result.Errors, crossErrors...)
			summary.TotalErrors += len(crossErrors)

			if len(result.Errors) == 0 {
				summary.SuccessfulFiles++
			} else {
				result.Success = false
				summary.FailedFiles++
			}

			// Score agent quality
			scorer := scoring.NewAgentScorer()
			score := scorer.Score(file.Contents, fm.Data, fm.Body)
			result.Quality = &score

			// Get improvement recommendations
			result.Improvements = GetAgentImprovements(file.Contents, fm.Data)
		}

		summary.Results = append(summary.Results, result)
		ctx.LogProcessed(file.RelPath, len(result.Errors))
	}

	return summary, nil
}

// validateAgentSpecific implements agent-specific validation rules
func validateAgentSpecific(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

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

	// Best practice checks
	errors = append(errors, validateAgentBestPractices(filePath, contents, data)...)

	return errors
}

// validateAgentBestPractices checks opinionated best practices for agents
func validateAgentBestPractices(filePath string, contents string, data map[string]interface{}) []cue.ValidationError {
	var suggestions []cue.ValidationError
	fmEndLine := GetFrontmatterEndLine(contents)

	// Count total lines (±10% tolerance: 200 base + 20 = 220)
	lines := strings.Count(contents, "\n")
	if lines > 220 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Agent is %d lines. Best practice: keep agents under ~220 lines (200±10%%) - move methodology to skills instead.", lines),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     1,
		})
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

	return suggestions
}