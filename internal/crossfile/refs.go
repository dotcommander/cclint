// Package crossfile provides cross-file validation for cclint.
//
// This file contains reference extraction functions: FindSkillReferences,
// ParseAllowedTools, IsToolUsed, and name extraction helpers.
package crossfile

import (
	"path/filepath"
	"regexp"
	"strings"
)

// Pre-compiled regex patterns share the complete local or plugin-qualified name form.
const skillRefNamePattern = `[a-z0-9][a-z0-9-]*(?::[a-z0-9][a-z0-9-]*)?`

var (
	skillPlainPattern    = regexp.MustCompile(`(?m)^[^*\n]*\bSkill:[ \t]*(` + skillRefNamePattern + `)`)
	skillBoldPattern     = regexp.MustCompile(`(?m)\*\*Skill\*\*:[ \t]*(` + skillRefNamePattern + `)`)
	skillFuncPattern     = regexp.MustCompile(`(?m)Skill\([ \t]*["']?(` + skillRefNamePattern + `)["']?[ \t]*\)`)
	skillListPattern     = regexp.MustCompile(`(?m)Skills?:[ \t]*\r?\n((?:[ \t]*[-*][ \t]+[^\n]*(?:\n|$))+)`)
	skillListItemPattern = regexp.MustCompile("(?m)^[ \\t]*[-*][ \\t]+[\"'`]?(" + skillRefNamePattern + ")")
	skillPatterns        = []*regexp.Regexp{skillPlainPattern, skillBoldPattern, skillFuncPattern}
)

// Pre-compiled regex patterns for tool parsing and matching.
var (
	// taskToolPattern matches Task(xxx) patterns in allowed-tools strings.
	taskToolPattern = regexp.MustCompile(`Task\([^)]+\)`)
)

// FindSkillReferences finds all skill references in content using multiple patterns.
// Matches: Skill: X, **Skill**: X, Skill(X), Skills: list, and code block declarations.
func FindSkillReferences(content string) []string {
	var skills []string
	seen := make(map[string]bool)
	add := func(skill string) {
		if skill != "" && !seen[skill] {
			skills = append(skills, skill)
			seen[skill] = true
		}
	}
	for _, pattern := range skillPatterns {
		for _, match := range pattern.FindAllStringSubmatch(content, -1) {
			add(match[1])
		}
	}
	for _, list := range skillListPattern.FindAllStringSubmatch(content, -1) {
		for _, item := range skillListItemPattern.FindAllStringSubmatch(list[1], -1) {
			add(item[1])
		}
	}
	return skills
}

// ParseAllowedTools parses allowed-tools frontmatter into individual tool names.
// Handles both Task(agent-name) patterns and simple tool names.
func ParseAllowedTools(s string) []string {
	var tools []string
	seen := make(map[string]bool)

	// Split on comma but be careful with Task(xxx) patterns
	// Use regex to find all tool declarations
	tasks := taskToolPattern.FindAllString(s, -1)
	for _, t := range tasks {
		if !seen[t] {
			tools = append(tools, t)
			seen[t] = true
		}
	}

	// Remove Task patterns from string to find other tools
	remaining := taskToolPattern.ReplaceAllString(s, "")
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

// ToolPatterns maps standard tools to their usage detection functions.
var ToolPatterns = map[string]func(string) bool{
	"Task":  func(c string) bool { return strings.Contains(c, "Task(") },
	"Read":  func(c string) bool { return strings.Contains(c, "Read(") || strings.Contains(c, "Read tool") },
	"Write": func(c string) bool { return strings.Contains(c, "Write(") || strings.Contains(c, "Write tool") },
	"Edit":  func(c string) bool { return strings.Contains(c, "Edit(") || strings.Contains(c, "Edit tool") },
	"Bash":  func(c string) bool { return strings.Contains(c, "Bash(") || strings.Contains(c, "Bash tool") },
	"Glob":  func(c string) bool { return strings.Contains(c, "Glob(") || strings.Contains(c, "Glob tool") },
	"Grep":  func(c string) bool { return strings.Contains(c, "Grep(") || strings.Contains(c, "Grep tool") },
}

// IsToolUsed checks if a tool is referenced in the content body.
func IsToolUsed(tool string, contents string) bool {
	// For Task(specific-agent), check if that specific agent is called
	if strings.HasPrefix(tool, "Task(") && strings.HasSuffix(tool, ")") {
		// Extract agent name: Task(foo-specialist) -> foo-specialist
		agentName := tool[5 : len(tool)-1]
		// Check if Task(agentName) appears in body (with possible whitespace)
		pattern := regexp.MustCompile(`Task\(\s*` + regexp.QuoteMeta(agentName) + `\s*[,)]`)
		return pattern.MatchString(contents)
	}

	// Check standard tools using pattern map
	if check, ok := ToolPatterns[tool]; ok {
		return check(contents)
	}

	// Default: check if tool name appears anywhere
	return strings.Contains(contents, tool)
}

// Helper functions for cross-file validation - name extraction

func ExtractAgentName(path string) string {
	// agents/foo-specialist.md -> foo-specialist
	// .claude/agents/foo.md -> foo
	parts := strings.Split(normalizeReferencePath(path), "/")
	filename := parts[len(parts)-1]
	return strings.TrimSuffix(filename, ".md")
}

func ExtractSkillName(path string) string {
	// skills/foo-bar/SKILL.md -> foo-bar
	// .claude/skills/foo/SKILL.md -> foo
	parts := strings.Split(normalizeReferencePath(path), "/")
	for i, part := range parts {
		if part == "skills" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}

func ExtractCommandName(path string) string {
	// commands/foo.md -> foo
	parts := strings.Split(normalizeReferencePath(path), "/")
	filename := parts[len(parts)-1]
	return strings.TrimSuffix(filename, ".md")
}

// normalizeReferencePath also handles foreign-platform paths supplied on Unix.
func normalizeReferencePath(path string) string {
	return strings.ReplaceAll(filepath.ToSlash(path), `\`, "/")
}
