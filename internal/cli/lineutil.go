package cli

import (
	"strings"
)

// FindLineNumber finds the line number (1-based) where a pattern first appears
func FindLineNumber(content, pattern string) int {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			return i + 1 // 1-based line numbers
		}
	}
	return 0
}

// FindSectionLine finds the line number of a markdown section header
func FindSectionLine(content, sectionName string) int {
	lines := strings.Split(content, "\n")
	patterns := []string{
		"## " + sectionName,
		"### " + sectionName,
		"# " + sectionName,
	}
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		for _, pattern := range patterns {
			if strings.HasPrefix(trimmed, pattern) {
				return i + 1
			}
		}
	}
	return 0
}

// FindFrontmatterFieldLine finds the line number of a YAML frontmatter field
func FindFrontmatterFieldLine(content, fieldName string) int {
	lines := strings.Split(content, "\n")
	inFrontmatter := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "---" {
			if !inFrontmatter {
				inFrontmatter = true
				continue
			} else {
				// End of frontmatter
				break
			}
		}
		if inFrontmatter {
			if strings.HasPrefix(trimmed, fieldName+":") {
				return i + 1
			}
		}
	}
	return 0
}

// GetFrontmatterEndLine returns the line number where frontmatter ends
func GetFrontmatterEndLine(content string) int {
	lines := strings.Split(content, "\n")
	frontmatterCount := 0
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			frontmatterCount++
			if frontmatterCount == 2 {
				return i + 1
			}
		}
	}
	return 1
}

// CountLines returns the total number of lines in content
func CountLines(content string) int {
	return strings.Count(content, "\n") + 1
}

// Severity levels
const (
	SeverityHigh   = "high"
	SeverityMedium = "medium"
	SeverityLow    = "low"
)

// DetermineSeverity determines issue severity based on the message
func DetermineSeverity(message string) string {
	// High severity - structural/critical issues
	highPatterns := []string{
		"Required field",
		"missing or empty",
		"Name must",
		"Invalid color",
		"fat agent",
		"fat command",
		"Error parsing",
		"Validation error",
	}
	for _, p := range highPatterns {
		if strings.Contains(message, p) {
			return SeverityHigh
		}
	}

	// Medium severity - best practice violations
	mediumPatterns := []string{
		"Best practice",
		"lines",
		"Foundation",
		"Workflow",
		"Anti-Pattern",
		"methodology",
		"Skill()",
	}
	for _, p := range mediumPatterns {
		if strings.Contains(message, p) {
			return SeverityMedium
		}
	}

	// Low severity - suggestions
	return SeverityLow
}

// ImprovementRecommendation represents a specific fix with point value
type ImprovementRecommendation struct {
	Description string
	PointValue  int
	Line        int
	Severity    string
}

// GetAgentImprovements returns specific improvement recommendations for agents
func GetAgentImprovements(content string, data map[string]interface{}) []ImprovementRecommendation {
	var recs []ImprovementRecommendation
	lines := CountLines(content)

	// Check for missing fields and sections
	if _, ok := data["model"]; !ok {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add 'model: sonnet' to frontmatter",
			PointValue:  5,
			Line:        GetFrontmatterEndLine(content),
			Severity:    SeverityMedium,
		})
	}


	if !strings.Contains(content, "## Foundation") {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add '## Foundation' section with Skill: reference",
			PointValue:  5,
			Line:        GetFrontmatterEndLine(content) + 2,
			Severity:    SeverityMedium,
		})
	}

	if !strings.Contains(content, "## Workflow") {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add '## Workflow' section with Phase 0-N structure",
			PointValue:  4,
			Line:        0,
			Severity:    SeverityMedium,
		})
	}

	if !strings.Contains(content, "## Success Criteria") {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add '## Success Criteria' section with checklist",
			PointValue:  3,
			Line:        0,
			Severity:    SeverityLow,
		})
	}

	if lines > 200 {
		recs = append(recs, ImprovementRecommendation{
			Description: "Extract methodology to skill - agent is over 200 lines",
			PointValue:  10,
			Line:        1,
			Severity:    SeverityHigh,
		})
	}

	return recs
}

// GetCommandImprovements returns specific improvement recommendations for commands
func GetCommandImprovements(content string, data map[string]interface{}) []ImprovementRecommendation {
	var recs []ImprovementRecommendation
	lines := CountLines(content)

	if _, ok := data["allowed-tools"]; !ok {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add 'allowed-tools: Task(X-specialist)' to frontmatter",
			PointValue:  10,
			Line:        GetFrontmatterEndLine(content),
			Severity:    SeverityHigh,
		})
	}

	if !strings.Contains(content, "## Quick Reference") && !strings.Contains(content, "Quick Reference") {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add '## Quick Reference' with semantic routing table",
			PointValue:  10,
			Line:        GetFrontmatterEndLine(content) + 2,
			Severity:    SeverityMedium,
		})
	}

	if !strings.Contains(content, "Task(") {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add Task(X-specialist) delegation - commands should delegate",
			PointValue:  10,
			Line:        0,
			Severity:    SeverityHigh,
		})
	}

	if lines > 50 {
		recs = append(recs, ImprovementRecommendation{
			Description: "Reduce command to <50 lines - move logic to agent",
			PointValue:  10,
			Line:        1,
			Severity:    SeverityHigh,
		})
	}

	return recs
}

// GetSkillImprovements returns specific improvement recommendations for skills
func GetSkillImprovements(content string, data map[string]interface{}) []ImprovementRecommendation {
	var recs []ImprovementRecommendation
	lines := CountLines(content)

	if !strings.Contains(content, "## Quick Reference") {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add '## Quick Reference' with semantic routing table",
			PointValue:  8,
			Line:        GetFrontmatterEndLine(content) + 2,
			Severity:    SeverityMedium,
		})
	}

	if !strings.Contains(content, "## Workflow") {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add '## Workflow' with Phase-based structure",
			PointValue:  6,
			Line:        0,
			Severity:    SeverityMedium,
		})
	}

	if !strings.Contains(content, "## Anti-Pattern") {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add '## Anti-Patterns' table with Problem + Fix columns",
			PointValue:  4,
			Line:        0,
			Severity:    SeverityMedium,
		})
	}

	if !strings.Contains(content, "## Success Criteria") {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add '## Success Criteria' with checkbox items",
			PointValue:  2,
			Line:        0,
			Severity:    SeverityLow,
		})
	}

	if lines > 500 {
		recs = append(recs, ImprovementRecommendation{
			Description: "Move heavy content to references/ subdirectory",
			PointValue:  10,
			Line:        1,
			Severity:    SeverityHigh,
		})
	}

	return recs
}
