package cli

import (
	"regexp"
	"strings"
)

// findSkillReferences finds all skill references in content using multiple patterns.
// Matches: Skill: X, **Skill**: X, Skill(X), Skills: list, and code block declarations.
func findSkillReferences(content string) []string {
	var skills []string
	seen := make(map[string]bool)

	// Patterns to match skill references
	skillPatterns := []*regexp.Regexp{
		// Skill: foo-bar (plain format, not inside bold markers)
		// Note: [^*\n]* prevents matching across newlines (Go regex quirk)
		regexp.MustCompile(`(?m)^[^*\n]*\bSkill:\s*([a-z0-9][a-z0-9-]*)`),
		// **Skill**: foo-bar (bold format)
		regexp.MustCompile(`(?m)\*\*Skill\*\*:\s*([a-z0-9][a-z0-9-]*)`),
		// Skill("foo-bar") or Skill(foo-bar) (function call format)
		regexp.MustCompile(`(?m)Skill\(\s*["']?([a-z0-9][a-z0-9-]*)["']?\s*\)`),
		// Skills: followed by list items
		regexp.MustCompile(`(?m)Skills?:\s*\n\s*[-*]\s*([a-z0-9][a-z0-9-]*)`),
	}

	for _, pattern := range skillPatterns {
		matches := pattern.FindAllStringSubmatch(content, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				skill := strings.TrimSpace(match[1])
				if !seen[skill] && skill != "" {
					skills = append(skills, skill)
					seen[skill] = true
				}
			}
		}
	}

	return skills
}

// parseAllowedTools parses allowed-tools frontmatter into individual tool names.
// Handles both Task(agent-name) patterns and simple tool names.
func parseAllowedTools(s string) []string {
	var tools []string
	seen := make(map[string]bool)

	// Split on comma but be careful with Task(xxx) patterns
	// Use regex to find all tool declarations
	taskPattern := regexp.MustCompile(`Task\([^)]+\)`)
	tasks := taskPattern.FindAllString(s, -1)
	for _, t := range tasks {
		if !seen[t] {
			tools = append(tools, t)
			seen[t] = true
		}
	}

	// Remove Task patterns from string to find other tools
	remaining := taskPattern.ReplaceAllString(s, "")
	parts := strings.Split(remaining, ",")
	for _, p := range parts {
		tool := strings.TrimSpace(p)
		if tool != "" && !seen[tool] {
			tools = append(tools, tool)
			seen[tool] = true
		}
	}
	return tools
}

// isToolUsed checks if a tool is referenced in the content body.
func isToolUsed(tool string, contents string) bool {
	// For Task(specific-agent), check if that specific agent is called
	if strings.HasPrefix(tool, "Task(") && strings.HasSuffix(tool, ")") {
		// Extract agent name: Task(foo-specialist) -> foo-specialist
		agentName := tool[5 : len(tool)-1]
		// Check if Task(agentName) appears in body (with possible whitespace)
		pattern := regexp.MustCompile(`Task\(\s*` + regexp.QuoteMeta(agentName) + `\s*[,)]`)
		return pattern.MatchString(contents)
	}

	// Check standard tools
	switch tool {
	case "Task":
		return strings.Contains(contents, "Task(")
	case "Read":
		return strings.Contains(contents, "Read(") || strings.Contains(contents, "Read tool")
	case "Write":
		return strings.Contains(contents, "Write(") || strings.Contains(contents, "Write tool")
	case "Edit":
		return strings.Contains(contents, "Edit(") || strings.Contains(contents, "Edit tool")
	case "Bash":
		return strings.Contains(contents, "Bash(") || strings.Contains(contents, "Bash tool")
	case "Glob":
		return strings.Contains(contents, "Glob(") || strings.Contains(contents, "Glob tool")
	case "Grep":
		return strings.Contains(contents, "Grep(") || strings.Contains(contents, "Grep tool")
	default:
		return strings.Contains(contents, tool)
	}
}

// Helper functions for cross-file validation - name extraction

func crossExtractAgentName(path string) string {
	// agents/foo-specialist.md -> foo-specialist
	// .claude/agents/foo.md -> foo
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]
	return strings.TrimSuffix(filename, ".md")
}

func crossExtractSkillName(path string) string {
	// skills/foo-bar/SKILL.md -> foo-bar
	// .claude/skills/foo/SKILL.md -> foo
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "skills" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func crossExtractCommandName(path string) string {
	// commands/foo.md -> foo
	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]
	return strings.TrimSuffix(filename, ".md")
}
