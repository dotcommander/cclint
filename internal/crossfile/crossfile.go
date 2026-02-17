// Package crossfile provides cross-file validation for cclint.
//
// This package contains the validation orchestration logic. Related functionality
// is split into:
//   - refs.go: Reference extraction (FindSkillReferences, ParseAllowedTools, etc.)
//   - graph.go: Cycle detection and chain tracing (DetectCycles, TraceChain, etc.)
package crossfile

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

// Pre-compiled regex patterns for cross-file validation.
var (
	// validateCommandTaskPattern matches Task(agent-name) in command contents.
	validateCommandTaskPattern = regexp.MustCompile(`Task\(([^,\)]+)`)

	// flagPattern matches --flag-name patterns in command contents.
	flagPattern = regexp.MustCompile(`--([a-z][a-z0-9-]*)`)

	// routingFlagPattern matches flags used as dispatch keys in command body.
	// Matches: `--flag` | (table row), --flag: (label), `--flag` (backtick-wrapped).
	routingFlagPattern = regexp.MustCompile("(?m)(?:`--([a-z][a-z0-9-]*)`\\s*\\||--([a-z][a-z0-9-]*)\\s*:)")

	// collectSkillRefPattern matches Skill: references in agent contents.
	collectSkillRefPattern = regexp.MustCompile(`(?m)^[^*]*\bSkill:\s*([a-z0-9][a-z0-9-]*)`)

	// extractTaskAgentRefsPattern matches Task(agent-name) in tools field.
	extractTaskAgentRefsPattern = regexp.MustCompile(`Task\(([a-z0-9][a-z0-9-]*)\)`)
)

// Pre-compiled regex patterns for ValidateSkill agent reference detection.
var skillAgentPatterns = []struct {
	pattern *regexp.Regexp
	example string
}{
	// Specialist patterns (existing, most specific)
	{regexp.MustCompile(`delegate to\s+([a-z0-9][a-z0-9-]*-specialist)`), "delegate to foo-specialist"},
	{regexp.MustCompile(`use\s+([a-z0-9][a-z0-9-]*-specialist)`), "use foo-specialist"},
	{regexp.MustCompile(`see\s+([a-z0-9][a-z0-9-]*-specialist)`), "see foo-specialist"},
	{regexp.MustCompile(`Task\(([a-z0-9][a-z0-9-]*-specialist)`), "Task(foo-specialist)"},

	// Generic Task() pattern - matches any agent name
	{regexp.MustCompile(`Task\(\s*["']?([a-z0-9][a-z0-9-]*)["']?\s*[,)]`), "Task(agent-name)"},

	// Narrative patterns for agents
	{regexp.MustCompile(`delegate via\s+([a-z0-9][a-z0-9-]*)`), "delegate via agent-name"},
	{regexp.MustCompile(`([a-z0-9][a-z0-9-]*-agent)\s+handles`), "foo-agent handles"},
}

// BuiltInSubagentTypes are Task() targets that exist in Claude Code's runtime,
// not as user-defined agent files. These should not trigger "missing agent" errors.
// Includes both built-in subagent types and model name references.
var BuiltInSubagentTypes = map[string]bool{
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
			name := ExtractAgentName(f.RelPath)
			v.agents[name] = f
		case discovery.FileTypeSkill:
			name := ExtractSkillName(f.RelPath)
			v.skills[name] = f
		case discovery.FileTypeCommand:
			name := ExtractCommandName(f.RelPath)
			v.commands[name] = f
		}
	}

	return v
}

// ValidateCommand checks command references to agents
func (v *CrossFileValidator) ValidateCommand(filePath string, contents string, frontmatter map[string]any) []cue.ValidationError {
	var errors []cue.ValidationError
	seenAgentErrors := make(map[string]bool)

	// Find all Task(X-specialist) or Task(X) patterns
	matches := validateCommandTaskPattern.FindAllStringSubmatch(contents, -1)

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
		if BuiltInSubagentTypes[agentRef] {
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
	primaryAgent, primaryAgentContents := v.findPrimaryAgent(taskMatches)
	if primaryAgent == "" {
		return errors
	}

	// Collect skill contents referenced by primary agent
	skillContents := v.collectAgentSkillContents(primaryAgentContents)

	// Build set of routing flags (used as dispatch keys in command body)
	routingFlags := make(map[string]bool)
	for _, match := range routingFlagPattern.FindAllStringSubmatch(contents, -1) {
		for _, group := range match[1:] {
			if group != "" {
				routingFlags[group] = true
			}
		}
	}

	// Find flags documented in command
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

		// Skip command-level routing flags (used in dispatch tables)
		if routingFlags[flag] {
			continue
		}

		// Check if flag is found in agent or skills
		if !v.isFlagInAgentOrSkills(flag, primaryAgentContents, skillContents) {
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

// findPrimaryAgent finds the primary agent that the command delegates to.
func (v *CrossFileValidator) findPrimaryAgent(taskMatches [][]string) (agentName, agentContents string) {
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
			return agentRef, agentFile.Contents
		}
	}
	return "", ""
}

// collectAgentSkillContents collects the contents of skills referenced by an agent.
func (v *CrossFileValidator) collectAgentSkillContents(agentContents string) []string {
	var skillContents []string
	skillMatches := collectSkillRefPattern.FindAllStringSubmatch(agentContents, -1)
	for _, sm := range skillMatches {
		if len(sm) >= 2 {
			if skillFile, exists := v.skills[sm[1]]; exists {
				skillContents = append(skillContents, skillFile.Contents)
			}
		}
	}
	return skillContents
}

// isFlagInAgentOrSkills checks if a flag is found in agent contents or skill contents.
// Only matches the --flag form to avoid false positives from bare substring matches
// (e.g., a flag named "file" matching the word "file" in prose).
func (v *CrossFileValidator) isFlagInAgentOrSkills(flag, agentContents string, skillContents []string) bool {
	prefixed := "--" + flag
	if strings.Contains(agentContents, prefixed) {
		return true
	}

	for _, sc := range skillContents {
		if strings.Contains(sc, prefixed) {
			return true
		}
	}
	return false
}

// checkUnusedAllowedTools detects tools declared in allowed-tools but never used.
func (v *CrossFileValidator) checkUnusedAllowedTools(filePath, contents string, frontmatter map[string]any) []cue.ValidationError {
	var errors []cue.ValidationError

	allowedTools, ok := frontmatter["allowed-tools"].(string)
	if !ok {
		return errors
	}

	tools := ParseAllowedTools(allowedTools)
	for _, tool := range tools {
		if IsToolUsed(tool, contents) {
			continue
		}

		isDeclarative := strings.Contains(strings.ToLower(contents), "saved as") ||
			strings.Contains(strings.ToLower(contents), "write to") ||
			strings.Contains(strings.ToLower(contents), "output to") ||
			strings.Contains(strings.ToLower(contents), "save to") ||
			strings.Contains(contents, ".md") ||
			strings.Contains(contents, ".json")

		var message string
		switch {
		case isDeclarative && (tool == "Write" || tool == "Read"):
			message = fmt.Sprintf("allowed-tools declares '%s' - consider making tool usage more explicit for LLM (e.g., 'Use Write tool to create...')", tool)
		case isDeclarative:
			message = fmt.Sprintf("allowed-tools declares '%s' without obvious invocation (consider making tool usage explicit)", tool)
		default:
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

	skillRefs := FindSkillReferences(contents)
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

// ValidateAgent checks agent references to skills and team agent references.
// It validates in-body Skill: references, frontmatter skills array, and
// Task() agent references in the frontmatter tools field (agent teams).
func (v *CrossFileValidator) ValidateAgent(filePath string, contents string, frontmatter map[string]any) []cue.ValidationError {
	var errors []cue.ValidationError

	skillRefs := FindSkillReferences(contents)
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

	// Validate Task() agent references in frontmatter tools field (agent teams)
	errors = append(errors, v.validateToolsAgentRefs(filePath, frontmatter)...)

	return errors
}

// validateToolsAgentRefs extracts Task(agent-name) patterns from the frontmatter
// tools field and validates that referenced agents exist.
// This validates agent team semantics where agents spawn sub-agents via Task().
func (v *CrossFileValidator) validateToolsAgentRefs(filePath string, frontmatter map[string]any) []cue.ValidationError {
	if frontmatter == nil {
		return nil
	}

	agentRefs := ExtractTaskAgentRefs(frontmatter["tools"])
	if len(agentRefs) == 0 {
		return nil
	}

	var errors []cue.ValidationError
	for _, agentRef := range agentRefs {
		if BuiltInSubagentTypes[agentRef] {
			continue
		}
		if _, exists := v.agents[agentRef]; !exists {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("tools field Task(%s) references non-existent agent. Create agents/%s.md", agentRef, agentRef),
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
			})
		}
	}
	return errors
}

// ExtractTaskAgentRefs extracts agent names from Task(agent-name) patterns
// in the tools field. Handles both string and array formats.
func ExtractTaskAgentRefs(tools any) []string {
	if tools == nil {
		return nil
	}

	var refs []string
	seen := make(map[string]bool)

	addMatches := func(s string) {
		matches := extractTaskAgentRefsPattern.FindAllStringSubmatch(s, -1)
		for _, m := range matches {
			if len(m) >= 2 && !seen[m[1]] {
				refs = append(refs, m[1])
				seen[m[1]] = true
			}
		}
	}

	switch v := tools.(type) {
	case string:
		addMatches(v)
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok {
				addMatches(s)
			}
		}
	}

	return refs
}

// validateFrontmatterSkills validates the skills array in agent frontmatter.
// Each entry should reference an existing skill.
func (v *CrossFileValidator) validateFrontmatterSkills(filePath string, frontmatter map[string]any) []cue.ValidationError {
	if frontmatter == nil {
		return nil
	}

	var errors []cue.ValidationError

	skillsVal, ok := frontmatter["skills"]
	if !ok {
		return nil
	}

	skillsList, ok := skillsVal.([]any)
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
func (v *CrossFileValidator) ValidateSkill(filePath string, contents string, frontmatter map[string]any) []cue.ValidationError {
	var errors []cue.ValidationError

	// Agent reference patterns - ordered from most specific to least specific
	seenAgents := make(map[string]bool)
	for _, agentPattern := range skillAgentPatterns {
		matches := agentPattern.pattern.FindAllStringSubmatch(contents, -1)
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
			if BuiltInSubagentTypes[agentRef] {
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
func (v *CrossFileValidator) validateFrontmatterAgent(filePath string, frontmatter map[string]any) []cue.ValidationError {
	if frontmatter == nil {
		return nil
	}

	agentName, ok := frontmatter["agent"].(string)
	if !ok || agentName == "" {
		return nil
	}

	// Skip built-in agent types
	if BuiltInSubagentTypes[agentName] {
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
	// Collect all references into a single map
	referencedSkills := v.getAllReferencedSkills()

	// Find orphans
	return v.findSkillOrphans(referencedSkills)
}

// getAllReferencedSkills returns a combined map of all referenced skills.
func (v *CrossFileValidator) getAllReferencedSkills() map[string]bool {
	referencedSkills := make(map[string]bool)

	// Collect from commands
	v.collectCommandReferences(referencedSkills)

	// Collect from agents
	v.collectAgentReferences(referencedSkills)

	// Collect from skills
	v.collectSkillToSkillReferencesMap(referencedSkills)

	return referencedSkills
}

// collectCommandReferences collects skill references from commands.
func (v *CrossFileValidator) collectCommandReferences(referencedSkills map[string]bool) {
	for _, cmd := range v.commands {
		// Check Task() pattern
		for _, match := range validateCommandTaskPattern.FindAllStringSubmatch(cmd.Contents, -1) {
			if len(match) >= 2 {
				agentRef := strings.TrimSpace(strings.Trim(match[1], `"'`))
				if !strings.HasSuffix(agentRef, "-specialist") {
					if _, exists := v.skills[agentRef]; exists {
						referencedSkills[agentRef] = true
					}
				}
			}
		}
		// Check Skill() and Skill: references
		for _, skillRef := range FindSkillReferences(cmd.Contents) {
			referencedSkills[skillRef] = true
		}
	}
}

// collectAgentReferences collects skill references from agents.
func (v *CrossFileValidator) collectAgentReferences(referencedSkills map[string]bool) {
	for _, agent := range v.agents {
		for _, skillRef := range FindSkillReferences(agent.Contents) {
			referencedSkills[skillRef] = true
		}
	}
}

// collectSkillToSkillReferencesMap collects skill references from other skills.
func (v *CrossFileValidator) collectSkillToSkillReferencesMap(referencedSkills map[string]bool) {
	for _, skill := range v.skills {
		currentSkillName := ExtractSkillName(skill.RelPath)
		for skillName := range v.skills {
			if skillName != currentSkillName && strings.Contains(skill.Contents, skillName) {
				referencedSkills[skillName] = true
			}
		}
	}
}

// findSkillOrphans returns validation errors for orphaned skills.
func (v *CrossFileValidator) findSkillOrphans(referencedSkills map[string]bool) []cue.ValidationError {
	var orphans []cue.ValidationError
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
