// Package cli provides cross-file validation for cclint.
//
// This file contains the validation orchestration logic. Related functionality
// is split into:
//   - crossfile_refs.go: Reference extraction (findSkillReferences, parseAllowedTools, etc.)
//   - crossfile_graph.go: Cycle detection and chain tracing (DetectCycles, TraceChain, etc.)
package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

// builtInSubagentTypes are Task() targets that exist in Claude Code's runtime,
// not as user-defined agent files. These should not trigger "missing agent" errors.
// Includes both built-in subagent types and model name references.
var builtInSubagentTypes = map[string]bool{
	// Built-in subagent types
	"general-purpose":   true,
	"statusline-setup":  true,
	"Explore":           true,
	"Plan":              true,
	"claude-code-guide": true,

	// Model names (used in Task() for model selection)
	"haiku":  true,
	"sonnet": true,
	"opus":   true,
}

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
			name := crossExtractAgentName(f.RelPath)
			v.agents[name] = f
		case discovery.FileTypeSkill:
			name := crossExtractSkillName(f.RelPath)
			v.skills[name] = f
		case discovery.FileTypeCommand:
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

	// Find all Task(X-specialist) or Task(X) patterns
	taskPattern := regexp.MustCompile(`Task\(([^,\)]+)`)
	matches := taskPattern.FindAllStringSubmatch(contents, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		agentRef := strings.TrimSpace(match[1])
		agentRef = strings.Trim(agentRef, `"'`)

		if strings.Contains(agentRef, "subagent_type") {
			continue
		}
		if seenAgentErrors[agentRef] {
			continue
		}
		if builtInSubagentTypes[agentRef] {
			continue
		}

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

	// Check for fake flags documented in command but not in agent or its skills
	errors = append(errors, v.checkFakeFlags(filePath, contents, matches)...)

	// Check for unused allowed-tools
	errors = append(errors, v.checkUnusedAllowedTools(filePath, contents, frontmatter)...)

	// Check for skill references (Skill: or Skill() patterns)
	errors = append(errors, v.checkSkillReferences(filePath, contents)...)

	return errors
}

// checkFakeFlags detects flags documented in command but not in agent or skills.
func (v *CrossFileValidator) checkFakeFlags(filePath, contents string, taskMatches [][]string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Find the primary agent this command delegates to
	var primaryAgent string
	var primaryAgentContents string
	for _, match := range taskMatches {
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
			break
		}
	}

	if primaryAgent == "" {
		return errors
	}

	// Collect skill contents referenced by primary agent
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

	// Find flags documented in command
	flagPattern := regexp.MustCompile(`--([a-z][a-z0-9-]*)`)
	flagMatches := flagPattern.FindAllStringSubmatch(contents, -1)
	seenFlags := make(map[string]bool)

	for _, match := range flagMatches {
		if len(match) < 2 {
			continue
		}
		flag := match[1]
		if seenFlags[flag] {
			continue
		}
		seenFlags[flag] = true

		foundInAgent := strings.Contains(primaryAgentContents, "--"+flag) || strings.Contains(primaryAgentContents, flag)
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
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return errors
}

// checkUnusedAllowedTools detects tools declared in allowed-tools but never used.
func (v *CrossFileValidator) checkUnusedAllowedTools(filePath, contents string, frontmatter map[string]interface{}) []cue.ValidationError {
	var errors []cue.ValidationError

	allowedTools, ok := frontmatter["allowed-tools"].(string)
	if !ok {
		return errors
	}

	tools := parseAllowedTools(allowedTools)
	for _, tool := range tools {
		if isToolUsed(tool, contents) {
			continue
		}

		isDeclarative := strings.Contains(strings.ToLower(contents), "saved as") ||
			strings.Contains(strings.ToLower(contents), "write to") ||
			strings.Contains(strings.ToLower(contents), "output to") ||
			strings.Contains(strings.ToLower(contents), "save to") ||
			strings.Contains(contents, ".md") ||
			strings.Contains(contents, ".json")

		var message string
		if isDeclarative && (tool == "Write" || tool == "Read") {
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

	return errors
}

// checkSkillReferences validates skill references in any component.
func (v *CrossFileValidator) checkSkillReferences(filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError
	seenSkillErrors := make(map[string]bool)

	skillRefs := findSkillReferences(contents)
	for _, skillRef := range skillRefs {
		if seenSkillErrors[skillRef] {
			continue
		}
		if _, exists := v.skills[skillRef]; !exists {
			seenSkillErrors[skillRef] = true
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("References non-existent skill '%s'. Create skills/%s/SKILL.md", skillRef, skillRef),
				Severity: "error",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return errors
}

// ValidateAgent checks agent references to skills.
// It validates both in-body Skill: references and frontmatter skills array.
func (v *CrossFileValidator) ValidateAgent(filePath string, contents string, frontmatter map[string]interface{}) []cue.ValidationError {
	var errors []cue.ValidationError

	skillRefs := findSkillReferences(contents)
	for _, skillRef := range skillRefs {
		if _, exists := v.skills[skillRef]; !exists {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Skill: %s references non-existent skill. Create skills/%s/SKILL.md", skillRef, skillRef),
				Severity: "error",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	// Validate frontmatter skills array (preloaded skills)
	errors = append(errors, v.validateFrontmatterSkills(filePath, frontmatter)...)

	return errors
}

// validateFrontmatterSkills validates the skills array in agent frontmatter.
// Each entry should reference an existing skill.
func (v *CrossFileValidator) validateFrontmatterSkills(filePath string, frontmatter map[string]interface{}) []cue.ValidationError {
	if frontmatter == nil {
		return nil
	}

	var errors []cue.ValidationError

	skillsVal, ok := frontmatter["skills"]
	if !ok {
		return nil
	}

	skillsList, ok := skillsVal.([]interface{})
	if !ok {
		return nil
	}

	for _, item := range skillsList {
		skillName, ok := item.(string)
		if !ok {
			continue
		}
		if _, exists := v.skills[skillName]; !exists {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Frontmatter skills references non-existent skill '%s'. Create skills/%s/SKILL.md", skillName, skillName),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		}
	}

	return errors
}

// ValidateSkill checks skill references to agents.
// It validates both in-body agent references and frontmatter agent field.
func (v *CrossFileValidator) ValidateSkill(filePath string, contents string, frontmatter map[string]interface{}) []cue.ValidationError {
	var errors []cue.ValidationError

	// Agent reference patterns - ordered from most specific to least specific
	agentPatterns := []struct {
		pattern string
		example string
	}{
		// Specialist patterns (existing, most specific)
		{`delegate to\s+([a-z0-9][a-z0-9-]*-specialist)`, "delegate to foo-specialist"},
		{`use\s+([a-z0-9][a-z0-9-]*-specialist)`, "use foo-specialist"},
		{`see\s+([a-z0-9][a-z0-9-]*-specialist)`, "see foo-specialist"},
		{`Task\(([a-z0-9][a-z0-9-]*-specialist)`, "Task(foo-specialist)"},

		// Generic Task() pattern - matches any agent name
		{`Task\(\s*["']?([a-z0-9][a-z0-9-]*)["']?\s*[,)]`, "Task(agent-name)"},

		// Narrative patterns for agents
		{`delegate via\s+([a-z0-9][a-z0-9-]*)`, "delegate via agent-name"},
		{`([a-z0-9][a-z0-9-]*-agent)\s+handles`, "foo-agent handles"},
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

			// Skip dynamic/variable references
			if strings.Contains(agentRef, "subagent_type") ||
				strings.Contains(agentRef, ".") ||
				strings.Contains(agentRef, "[") {
				continue
			}

			// Skip built-in agents
			if builtInSubagentTypes[agentRef] {
				continue
			}

			if seenAgents[agentRef] {
				continue
			}
			seenAgents[agentRef] = true

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

	// Validate frontmatter agent field
	errors = append(errors, v.validateFrontmatterAgent(filePath, frontmatter)...)

	return errors
}

// validateFrontmatterAgent validates the agent field in skill frontmatter.
// The agent field references a specific agent type for execution.
func (v *CrossFileValidator) validateFrontmatterAgent(filePath string, frontmatter map[string]interface{}) []cue.ValidationError {
	if frontmatter == nil {
		return nil
	}

	agentName, ok := frontmatter["agent"].(string)
	if !ok || agentName == "" {
		return nil
	}

	// Skip built-in agent types
	if builtInSubagentTypes[agentName] {
		return nil
	}

	if _, exists := v.agents[agentName]; !exists {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Frontmatter agent field references non-existent agent '%s'. Create agents/%s.md", agentName, agentName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	return nil
}

// FindOrphanedSkills finds skills that aren't referenced by any command, agent, or other skill
func (v *CrossFileValidator) FindOrphanedSkills() []cue.ValidationError {
	var orphans []cue.ValidationError
	referencedSkills := make(map[string]bool)

	// Check commands for skill references via Task() and Skill() patterns
	for _, cmd := range v.commands {
		taskPattern := regexp.MustCompile(`Task\(([^,\)]+)`)
		matches := taskPattern.FindAllStringSubmatch(cmd.Contents, -1)
		for _, match := range matches {
			if len(match) < 2 {
				continue
			}
			agentRef := strings.TrimSpace(match[1])
			agentRef = strings.Trim(agentRef, `"'`)
			if strings.HasSuffix(agentRef, "-specialist") {
				continue
			}
			if _, exists := v.skills[agentRef]; exists {
				referencedSkills[agentRef] = true
			}
		}
		// Also check Skill() and Skill: references in commands
		skillRefs := findSkillReferences(cmd.Contents)
		for _, skillRef := range skillRefs {
			referencedSkills[skillRef] = true
		}
	}

	// Check agents for skill declarations
	for _, agent := range v.agents {
		skillRefs := findSkillReferences(agent.Contents)
		for _, skillRef := range skillRefs {
			referencedSkills[skillRef] = true
		}
	}

	// Check skills for references to other skills
	for _, skill := range v.skills {
		for skillName := range v.skills {
			if skillName == crossExtractSkillName(skill.RelPath) {
				continue
			}
			if strings.Contains(skill.Contents, skillName) {
				referencedSkills[skillName] = true
			}
		}
	}

	// Find orphans
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
