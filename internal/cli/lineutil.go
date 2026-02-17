package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// Pre-compiled secret detection patterns (used by detectSecrets)
var secretPatterns = []struct {
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

// Pre-compiled placeholder detection patterns (used by isPlaceholderSecret)
var (
	placeholderFakeRe   = regexp.MustCompile(`(?i)(x{4,}|0{6,})`)
	placeholderBracketRe = regexp.MustCompile(`<[^>]*(?:key|token|secret|password|api)[^>]*>`)
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
func GetAgentImprovements(content string, data map[string]any) []ImprovementRecommendation {
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
func GetCommandImprovements(content string, data map[string]any) []ImprovementRecommendation {
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
func GetSkillImprovements(content string, data map[string]any) []ImprovementRecommendation {
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
func ValidateAllowedTools(data map[string]any, filePath string, contents string) []cue.ValidationError {
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
		tools, ok := data[field].(string)
		if !ok || tools == "" {
			continue
		}
		// Parse comma-separated tools
		for tool := range strings.SplitSeq(tools, ",") {
			tool = strings.TrimSpace(tool)
			if tool == "" {
				continue
			}
			baseTool := extractBaseToolName(tool)
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

	return warnings
}

// extractBaseToolName returns the base tool name from patterns like "Task(name)" or "Bash(npm:*)".
func extractBaseToolName(tool string) string {
	baseTool := tool
	if idx := strings.Index(tool, "("); idx > 0 {
		baseTool = tool[:idx]
	}
	if colonIdx := strings.Index(baseTool, ":"); colonIdx > 0 {
		baseTool = baseTool[:colonIdx]
	}
	return baseTool
}

// ValidateToolFieldName ensures components use correct field name (tools vs allowed-tools)
// - Agents MUST use 'tools:', not 'allowed-tools:'
// - Commands and Skills MUST use 'allowed-tools:', not 'tools:'
func ValidateToolFieldName(data map[string]any, filePath string, contents string, componentType string) []cue.ValidationError {
	var errors []cue.ValidationError

	switch componentType {
	case cue.TypeAgent:
		// Agents must use 'tools:', not 'allowed-tools:'
		if _, hasAllowedTools := data["allowed-tools"]; hasAllowedTools {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  "Agents must use 'tools:', not 'allowed-tools:'. Rename the field.",
				Severity: "error",
				Source:   cue.SourceCClintObserve,
				Line:     FindFrontmatterFieldLine(contents, "allowed-tools"),
			})
		}
	case cue.TypeCommand, cue.TypeSkill:
		// Commands and skills must use 'allowed-tools:', not 'tools:'
		if _, hasTools := data["tools"]; hasTools {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("%ss must use 'allowed-tools:', not 'tools:'. Rename the field.", cases.Title(language.English).String(componentType)),
				Severity: "error",
				Source:   cue.SourceCClintObserve,
				Line:     FindFrontmatterFieldLine(contents, "tools"),
			})
		}
	}

	return errors
}

// isPlaceholderSecret checks if a matched "secret" is actually a placeholder/example.
// Returns true if the value should be ignored (not flagged as a real secret).
func isPlaceholderSecret(line string) bool {
	lowerLine := strings.ToLower(line)

	// Common placeholder patterns in the value
	placeholderPatterns := []string{
		"example",
		"placeholder",
		"your-",
		"your_",
		"<your",
		"[your",
		"{your",
		"replace",
		"insert",
		"change-me",
		"changeme",
		"xxx",
		"***",
		"dummy",
		"sample",
		"fake",
		"test-key",
		"testkey",
		"test_key",
		"my-key",
		"mykey",
		"my_key",
		"here>",
		"here]",
		"here}",
		"-here",
		"_here",
		"file-value", // Documentation example
		"env-value",  // Documentation example
		"some-value", // Generic example
		"value-here", // Generic example
	}

	// Skip lines that are clearly documentation/comments showing examples
	trimmedLine := strings.TrimSpace(line)
	if strings.HasPrefix(trimmedLine, "//") || strings.HasPrefix(trimmedLine, "#") {
		// Comment lines showing examples are not real secrets
		return true
	}

	for _, p := range placeholderPatterns {
		if strings.Contains(lowerLine, p) {
			return true
		}
	}

	// AWS example keys (from AWS documentation)
	awsExamplePatterns := []string{
		"akiaiosfodnn7example",  // AWS example access key
		"wjalrxutnfemi/k7mdeng", // AWS example secret key prefix
		"examplekey",            // Generic example
	}

	for _, p := range awsExamplePatterns {
		if strings.Contains(lowerLine, p) {
			return true
		}
	}

	// Check for obviously fake patterns like repeated x's or zeros
	// sk-xxxxxxxx or sk-00000000
	if placeholderFakeRe.MatchString(line) {
		return true
	}

	// Check for angle bracket placeholders: <API_KEY>, <token>, etc.
	if placeholderBracketRe.MatchString(lowerLine) {
		return true
	}

	return false
}

// findSecretLineNumber returns the 1-based line number of the first non-placeholder match.
// Returns 0 if all matches are placeholders.
func findSecretLineNumber(pattern *regexp.Regexp, contents string) int {
	lines := strings.Split(contents, "\n")
	for i, line := range lines {
		if !pattern.MatchString(line) {
			continue
		}
		if isPlaceholderSecret(line) {
			continue
		}
		return i + 1
	}
	return 0
}

// detectSecrets checks for hardcoded secrets in content
func detectSecrets(contents string, filePath string) []cue.ValidationError {
	var warnings []cue.ValidationError

	for _, sp := range secretPatterns {
		if !sp.pattern.MatchString(contents) {
			continue
		}
		lineNum := findSecretLineNumber(sp.pattern, contents)
		if lineNum == 0 {
			continue // all matches were placeholders
		}
		warnings = append(warnings, cue.ValidationError{
			File:     filePath,
			Message:  sp.message,
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
			Line:     lineNum,
		})
	}

	return warnings
}
