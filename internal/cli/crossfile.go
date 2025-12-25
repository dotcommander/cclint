package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

// CrossFileValidator validates references between components
type CrossFileValidator struct {
	agents   map[string]discovery.File
	skills   map[string]discovery.File
	commands map[string]discovery.File
}

// NewCrossFileValidator creates a validator with indexed files
func NewCrossFileValidator(files []discovery.File) *CrossFileValidator {
	v := &CrossFileValidator{
		agents:   make(map[string]discovery.File),
		skills:   make(map[string]discovery.File),
		commands: make(map[string]discovery.File),
	}

	for _, f := range files {
		switch f.Type {
		case discovery.FileTypeAgent:
			// Extract agent name from path: agents/foo-specialist.md -> foo-specialist
			name := crossExtractAgentName(f.RelPath)
			v.agents[name] = f
		case discovery.FileTypeSkill:
			// Extract skill name from path: skills/foo-bar/SKILL.md -> foo-bar
			name := crossExtractSkillName(f.RelPath)
			v.skills[name] = f
		case discovery.FileTypeCommand:
			// Extract command name from path: commands/foo.md -> foo
			name := crossExtractCommandName(f.RelPath)
			v.commands[name] = f
		}
	}

	return v
}

// ValidateCommand checks command references to agents
func (v *CrossFileValidator) ValidateCommand(filePath string, contents string, frontmatter map[string]interface{}) []cue.ValidationError {
	var errors []cue.ValidationError
	seenAgentErrors := make(map[string]bool)

	// === MISSING AGENT REFERENCE ===
	// Find all Task(X-specialist) or Task(X) patterns
	taskPattern := regexp.MustCompile(`Task\(([^,\)]+)`)
	matches := taskPattern.FindAllStringSubmatch(contents, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		agentRef := strings.TrimSpace(match[1])
		// Remove quotes if present
		agentRef = strings.Trim(agentRef, `"'`)

		// Skip subagent_type patterns (those are handled differently)
		if strings.Contains(agentRef, "subagent_type") {
			continue
		}

		// Skip if already reported this agent
		if seenAgentErrors[agentRef] {
			continue
		}

		// Check if agent exists - OUR OBSERVATION (consistency check)
		if _, exists := v.agents[agentRef]; !exists {
			seenAgentErrors[agentRef] = true
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Task(%s) references non-existent agent. Create agents/%s.md", agentRef, agentRef),
				Severity: "error",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	// === FAKE FLAGS DETECTOR ===
	// Find the PRIMARY agent this command delegates to (first Task() call)
	var primaryAgent string
	var primaryAgentContents string
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		agentRef := strings.TrimSpace(match[1])
		agentRef = strings.Trim(agentRef, `"'`)
		if strings.Contains(agentRef, "subagent_type") {
			continue
		}
		if agentFile, exists := v.agents[agentRef]; exists {
			primaryAgent = agentRef
			primaryAgentContents = agentFile.Contents
			break // Use first valid agent as primary
		}
	}

	// Find flags documented in command but not in primary agent or its skills
	if primaryAgent != "" {
		flagPattern := regexp.MustCompile(`--([a-z][a-z0-9-]*)`)
		flagMatches := flagPattern.FindAllStringSubmatch(contents, -1)
		seenFlags := make(map[string]bool)

		// Collect all skill contents referenced by primary agent
		var skillContents []string
		skillRefPattern := regexp.MustCompile(`(?m)^[^*]*\bSkill:\s*([a-z0-9][a-z0-9-]*)`)
		skillMatches := skillRefPattern.FindAllStringSubmatch(primaryAgentContents, -1)
		for _, sm := range skillMatches {
			if len(sm) >= 2 {
				if skillFile, exists := v.skills[sm[1]]; exists {
					skillContents = append(skillContents, skillFile.Contents)
				}
			}
		}

		for _, match := range flagMatches {
			if len(match) < 2 {
				continue
			}
			flag := match[1]
			if seenFlags[flag] {
				continue
			}
			seenFlags[flag] = true

			// Check if flag appears in primary agent
			foundInAgent := strings.Contains(primaryAgentContents, "--"+flag) || strings.Contains(primaryAgentContents, flag)

			// Check if flag appears in any referenced skill
			foundInSkill := false
			for _, sc := range skillContents {
				if strings.Contains(sc, "--"+flag) || strings.Contains(sc, flag) {
					foundInSkill = true
					break
				}
			}

			if !foundInAgent && !foundInSkill {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Flag '--%s' documented but not found in agent '%s' or its skills - may be fake", flag, primaryAgent),
					Severity: "warning",
					Source:   cue.SourceCClintObserve,
				})
			}
		}
	}

	// === UNUSED ALLOWED-TOOLS ===
	// Check if allowed-tools declares tools that aren't used
	if allowedTools, ok := frontmatter["allowed-tools"].(string); ok {
		tools := parseAllowedTools(allowedTools)
		for _, tool := range tools {
			if !isToolUsed(tool, contents) {
				// Check if it's a declarative command (mentions file output in description/body)
				isDeclarative := strings.Contains(strings.ToLower(contents), "saved as") ||
					strings.Contains(strings.ToLower(contents), "write to") ||
					strings.Contains(strings.ToLower(contents), "output to") ||
					strings.Contains(strings.ToLower(contents), "save to") ||
					strings.Contains(contents, ".md") ||
					strings.Contains(contents, ".json")

				var message string
				if isDeclarative && (tool == "Write" || tool == "Read") {
					// Declarative command with file I/O - reminder to be explicit
					message = fmt.Sprintf("allowed-tools declares '%s' - consider making tool usage more explicit for LLM (e.g., 'Use Write tool to create...')", tool)
				} else if isDeclarative {
					message = fmt.Sprintf("allowed-tools declares '%s' without obvious invocation (consider making tool usage explicit)", tool)
				} else {
					message = fmt.Sprintf("allowed-tools declares '%s' but it's never used in command body", tool)
				}

				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  message,
					Severity: "info",
					Source:   cue.SourceCClintObserve,
				})
			}
		}
	}

	return errors
}

// ValidateAgent checks agent references to skills
func (v *CrossFileValidator) ValidateAgent(filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// === MISSING SKILL REFERENCE ===
	// Use comprehensive skill reference detection (Skill:, **Skill**:, Skill(), etc.)
	skillRefs := findSkillReferences(contents)

	for _, skillRef := range skillRefs {
		// Check if skill exists - OUR OBSERVATION (consistency check)
		if _, exists := v.skills[skillRef]; !exists {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Skill: %s references non-existent skill. Create skills/%s/SKILL.md", skillRef, skillRef),
				Severity: "error",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return errors
}

// ValidateSkill checks skill references to agents
func (v *CrossFileValidator) ValidateSkill(filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// === MISSING AGENT REFERENCE ===
	// Skills often mention "delegate to X-specialist" or "use Y agent"
	// Look for patterns like:
	// - "delegate to foo-specialist"
	// - "use bar-specialist"
	// - "see baz agent"
	// - "Task(qux-specialist)"
	agentPatterns := []struct {
		pattern string
		example string
	}{
		{`delegate to\s+([a-z0-9][a-z0-9-]*-specialist)`, "delegate to foo-specialist"},
		{`use\s+([a-z0-9][a-z0-9-]*-specialist)`, "use foo-specialist"},
		{`see\s+([a-z0-9][a-z0-9-]*-specialist)`, "see foo-specialist"},
		{`Task\(([a-z0-9][a-z0-9-]*-specialist)`, "Task(foo-specialist)"},
	}

	seenAgents := make(map[string]bool)
	for _, agentPattern := range agentPatterns {
		re := regexp.MustCompile(agentPattern.pattern)
		matches := re.FindAllStringSubmatch(contents, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			agentRef := strings.TrimSpace(match[1])
			if seenAgents[agentRef] {
				continue
			}
			seenAgents[agentRef] = true

			// Check if agent exists - OUR OBSERVATION (consistency check)
			if _, exists := v.agents[agentRef]; !exists {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Skill references '%s' but agent doesn't exist. Create agents/%s.md", agentRef, agentRef),
					Severity: "error",
					Source:   cue.SourceCClintObserve,
				})
			}
		}
	}

	return errors
}

// Helper functions for cross-file validation

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

func parseAllowedTools(s string) []string {
	// Parse "Task(X-specialist), Task(Y-specialist), Read" -> ["Task(X-specialist)", "Task(Y-specialist)", "Read"]
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

// ChainLink represents a component in the delegation chain
type ChainLink struct {
	Type     string // "command", "agent", "skill"
	Name     string
	Path     string
	Lines    int
	Children []ChainLink
}

// TraceChain traces the full delegation chain starting from a component
func (v *CrossFileValidator) TraceChain(componentType string, name string) *ChainLink {
	switch componentType {
	case "command":
		return v.traceFromCommand(name)
	case "agent":
		return v.traceFromAgent(name)
	case "skill":
		return v.traceFromSkill(name)
	}
	return nil
}

func (v *CrossFileValidator) traceFromCommand(name string) *ChainLink {
	file, exists := v.commands[name]
	if !exists {
		return nil
	}

	link := &ChainLink{
		Type:  "command",
		Name:  name,
		Path:  file.RelPath,
		Lines: strings.Count(file.Contents, "\n") + 1,
	}

	// Find Task() delegations
	taskPattern := regexp.MustCompile(`Task\(([^,\)]+)`)
	matches := taskPattern.FindAllStringSubmatch(file.Contents, -1)
	for _, match := range matches {
		if len(match) >= 2 {
			agentRef := strings.TrimSpace(match[1])
			agentRef = strings.Trim(agentRef, `"'`)
			if !strings.Contains(agentRef, "subagent_type") {
				if child := v.traceFromAgent(agentRef); child != nil {
					link.Children = append(link.Children, *child)
				}
			}
		}
	}

	return link
}

func (v *CrossFileValidator) traceFromAgent(name string) *ChainLink {
	file, exists := v.agents[name]
	if !exists {
		return nil
	}

	link := &ChainLink{
		Type:  "agent",
		Name:  name,
		Path:  file.RelPath,
		Lines: strings.Count(file.Contents, "\n") + 1,
	}

	// Find Skill references using comprehensive pattern matching
	skillRefs := findSkillReferences(file.Contents)
	for _, skillRef := range skillRefs {
		if child := v.traceFromSkill(skillRef); child != nil {
			link.Children = append(link.Children, *child)
		}
	}

	return link
}

func (v *CrossFileValidator) traceFromSkill(name string) *ChainLink {
	file, exists := v.skills[name]
	if !exists {
		return nil
	}

	return &ChainLink{
		Type:  "skill",
		Name:  name,
		Path:  file.RelPath,
		Lines: strings.Count(file.Contents, "\n") + 1,
	}
}

// FormatChain formats a chain link as a tree string
func FormatChain(link *ChainLink, indent string) string {
	if link == nil {
		return ""
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s%s (%s, %d lines)\n", indent, link.Name, link.Type, link.Lines))

	for i, child := range link.Children {
		prefix := "├── "
		childIndent := indent + "│   "
		if i == len(link.Children)-1 {
			prefix = "└── "
			childIndent = indent + "    "
		}
		sb.WriteString(fmt.Sprintf("%s%s", indent, prefix))
		sb.WriteString(FormatChain(&child, childIndent)[len(childIndent):])
	}

	return sb.String()
}

// FindOrphanedSkills finds skills that aren't referenced by any command, agent, or other skill
func (v *CrossFileValidator) FindOrphanedSkills() []cue.ValidationError {
	var orphans []cue.ValidationError

	// Track which skills are referenced
	referencedSkills := make(map[string]bool)

	// Check all commands for Task() delegation to skills
	for _, cmd := range v.commands {
		taskPattern := regexp.MustCompile(`Task\(([^,\)]+)`)
		matches := taskPattern.FindAllStringSubmatch(cmd.Contents, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			agentRef := strings.TrimSpace(match[1])
			agentRef = strings.Trim(agentRef, `"'`)
			// Skip if not a skill reference (skills usually don't have -specialist suffix)
			if strings.HasSuffix(agentRef, "-specialist") {
				continue
			}
			// Mark as referenced if it exists as a skill
			if _, exists := v.skills[agentRef]; exists {
				referencedSkills[agentRef] = true
			}
		}
	}

	// Check all agents for skill declarations using comprehensive pattern matching
	for _, agent := range v.agents {
		skillRefs := findSkillReferences(agent.Contents)
		for _, skillRef := range skillRefs {
			referencedSkills[skillRef] = true
		}
	}

	// Check all skills for references to other skills
	for _, skill := range v.skills {
		// Skills might mention "see X skill" or "use Y pattern"
		for skillName := range v.skills {
			if skillName == crossExtractSkillName(skill.RelPath) {
				continue // Don't count self-reference
			}
			// Look for skill name mentions
			if strings.Contains(skill.Contents, skillName) {
				referencedSkills[skillName] = true
			}
		}
	}

	// Find orphaned skills - OUR OBSERVATION
	for skillName, skillFile := range v.skills {
		if !referencedSkills[skillName] {
			orphans = append(orphans, cue.ValidationError{
				File:     skillFile.RelPath,
				Message:  fmt.Sprintf("Skill '%s' has no incoming references - consider adding crossrefs from commands/agents/skills", skillName),
				Severity: "info",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return orphans
}
