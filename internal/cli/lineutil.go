package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
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

// ValidateAllowedTools validates that allowed-tools and tools fields contain known tool names
func ValidateAllowedTools(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var warnings []cue.ValidationError

	// Known tools from Claude Code documentation
	knownTools := map[string]bool{
		// File operations
		"Read": true, "Write": true, "Edit": true, "MultiEdit": true,
		"Glob": true, "Grep": true, "LS": true,
		// Execution
		"Bash": true, "Task": true,
		// Web
		"WebFetch": true, "WebSearch": true,
		// Interactive
		"AskUserQuestion": true, "TodoWrite": true,
		// Special
		"Skill": true, "LSP": true, "NotebookEdit": true,
		"EnterPlanMode": true, "ExitPlanMode": true,
		"KillShell": true, "TaskOutput": true,
		// Wildcards
		"*": true,
	}

	// Check both "tools" and "allowed-tools" fields
	toolsFields := []string{"tools", "allowed-tools"}

	for _, field := range toolsFields {
		if tools, ok := data[field].(string); ok && tools != "" {
			// Parse comma-separated tools
			toolList := strings.Split(tools, ",")
			for _, tool := range toolList {
				tool = strings.TrimSpace(tool)
				if tool == "" {
					continue
				}
				// Handle patterns like "Task(specialist-name)"
				baseTool := tool
				if idx := strings.Index(tool, "("); idx > 0 {
					baseTool = tool[:idx]
				}
				// Handle patterns like "Bash(npm:*)" - ignore suffix after colon
				if colonIdx := strings.Index(baseTool, ":"); colonIdx > 0 {
					baseTool = baseTool[:colonIdx]
				}
				// Validate base tool
				if !knownTools[baseTool] {
					warnings = append(warnings, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Unknown tool '%s' in %s. Check spelling or verify it's a valid tool.", tool, field),
						Severity: "warning",
						Source:   cue.SourceCClintObserve,
						Line:     FindFrontmatterFieldLine(contents, field),
					})
				}
			}
		}
	}

	return warnings
}

// detectSecrets checks for hardcoded secrets in content
func detectSecrets(contents string, filePath string) []cue.ValidationError {
	var warnings []cue.ValidationError

	secretPatterns := []struct {
		pattern *regexp.Regexp
		message string
	}{
		{regexp.MustCompile(`(?i)(api[_-]?key|apikey)\s*[:=]\s*["'][^"']{10,}["']`),
			"Possible hardcoded API key detected - use environment variables"},
		{regexp.MustCompile(`(?i)(password|passwd|pwd)\s*[:=]\s*["'][^"']+["']`),
			"Possible hardcoded password detected - use secrets management"},
		{regexp.MustCompile(`(?i)(secret|token)\s*[:=]\s*["'][^"']{10,}["']`),
			"Possible hardcoded secret/token detected - use environment variables"},
		{regexp.MustCompile(`sk-[a-zA-Z0-9]{20,}`),
			"OpenAI API key pattern detected - never commit API keys"},
		{regexp.MustCompile(`xoxb-[a-zA-Z0-9-]+`),
			"Slack bot token pattern detected - use environment variables"},
		{regexp.MustCompile(`ghp_[a-zA-Z0-9]{36}`),
			"GitHub personal access token pattern detected"},
		{regexp.MustCompile(`AIza[0-9A-Za-z\-_]{35}`),
			"Google API key pattern detected - use environment variables"},
		{regexp.MustCompile(`[0-9]+-[0-9A-Za-z_]{32}\.apps\.googleusercontent\.com`),
			"Google OAuth client ID pattern detected"},
		{regexp.MustCompile(`-----BEGIN (RSA |DSA )?PRIVATE KEY-----`),
			"Private key detected - never commit private keys"},
		{regexp.MustCompile(`aws_access_key_id\s*[:=]\s*["']?[A-Z0-9]{20}["']?`),
			"AWS access key ID detected - use environment variables"},
		{regexp.MustCompile(`aws_secret_access_key\s*[:=]\s*["']?[A-Za-z0-9/+=]{40}["']?`),
			"AWS secret access key detected - use environment variables"},
	}

	for _, sp := range secretPatterns {
		if sp.pattern.MatchString(contents) {
			// Find line number for better error reporting
			lineNum := 0
			lines := strings.Split(contents, "\n")
			for i, line := range lines {
				if sp.pattern.MatchString(line) {
					lineNum = i + 1
					break
				}
			}

			warnings = append(warnings, cue.ValidationError{
				File:     filePath,
				Message:  sp.message,
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
				Line:     lineNum,
			})
		}
	}

	return warnings
}
