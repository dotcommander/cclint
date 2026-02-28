package lint

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/scoring"
	"github.com/dotcommander/cclint/internal/textutil"
)

// bodyToolNegativePattern matches lines that explicitly disclaim a tool (e.g. "do not use Bash").
var bodyToolNegativePattern = regexp.MustCompile(`(?i)\b(do not use|don't use|never use|avoid using|not use)\b`)

// validModelPattern matches known Claude Code model values.
// Bare names: haiku, sonnet, opus, inherit, opusplan.
// Optional version suffix in brackets: sonnet[1m], haiku[2].
var validModelPattern = regexp.MustCompile(`^(haiku|sonnet|opus|inherit|opusplan)(\[\w+\])?$`)

// LintResult represents a single linting result
type LintResult struct {
	File         string
	Type         string
	Errors       []cue.ValidationError
	Warnings     []cue.ValidationError
	Suggestions  []cue.ValidationError
	Improvements []textutil.ImprovementRecommendation
	Success      bool
	Duration     int64
	Quality      *scoring.QualityScore
}

// LintSummary summarizes all linting results
type LintSummary struct {
	ProjectRoot      string
	ComponentType    string // e.g., "agents", "commands", "skills"
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

// LintAgents runs linting on agent files using the generic linter.
func LintAgents(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewAgentLinter()), nil
}

// knownAgentFields lists valid frontmatter fields per Anthropic docs
// Source: https://code.claude.com/docs/en/sub-agents
var knownAgentFields = map[string]bool{
	"name":            true, // Required: unique identifier
	"description":     true, // Required: what the agent does
	"model":           true, // Optional: sonnet, opus, haiku, inherit
	"color":           true, // Optional: display color in UI (set via /agents wizard)
	"tools":           true, // Optional: tool access allowlist
	"disallowedTools": true, // Optional: tool access denylist
	"permissionMode":  true, // Optional: default, acceptEdits, delegate, dontAsk, bypassPermissions, plan
	"maxTurns":        true, // Optional: max conversation turns (positive integer)
	"skills":          true, // Optional: skills to preload into context
	"hooks":           true, // Optional: agent-level hooks (PreToolUse, PostToolUse, Stop)
	"memory":          true, // Optional: persistent memory scope (user, project, local) (v2.1.33+)
	"mcpServers":      true, // Optional: MCP server names available to agent
	"isolation":       true, // Optional: subagent isolation mode (worktree) (v2.1.49+)
	"background":      true, // Optional: always run as background task (v2.1.49+)
}

// validateAgentSpecific implements agent-specific validation rules.
// Orchestrates validation by delegating to focused check functions.
func validateAgentSpecific(data map[string]any, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Frontmatter field validation
	errors = append(errors, validateUnknownFields(data, filePath, contents)...)
	errors = append(errors, validateRequiredFields(data, filePath, contents)...)

	// Individual field validation
	errors = append(errors, validateAgentColor(data, filePath)...)
	errors = append(errors, validateAgentMemory(data, filePath, contents)...)
	errors = append(errors, validateAgentModel(data, filePath, contents)...)
	errors = append(errors, validateAgentMCPServersField(data, filePath, contents)...)
	errors = append(errors, validateAgentPermissionMode(data, filePath, contents)...)
	errors = append(errors, validateAgentMaxTurns(data, filePath, contents)...)
	errors = append(errors, validateAgentAutonomousPattern(data, filePath, contents)...)

	// Cross-field validation
	errors = append(errors, textutil.ValidateToolFieldName(data, filePath, contents, "agent")...)
	errors = append(errors, validateAgentHooks(data, filePath)...)
	errors = append(errors, validateAgentBestPractices(filePath, contents, data)...)
	errors = append(errors, validateBodyToolMismatch(data, filePath, contents)...)

	return errors
}

// validateUnknownFields checks for unsupported frontmatter fields.
func validateUnknownFields(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError
	for key := range data {
		if !knownAgentFields[key] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown frontmatter field '%s'. Valid fields: name, description, model, color, tools, disallowedTools, permissionMode, maxTurns, skills, hooks, memory, mcpServers", key),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
				Line:     textutil.FindFrontmatterFieldLine(contents, key),
			})
		}
	}
	return errors
}

// validateRequiredFields validates name and description requirements.
func validateRequiredFields(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	if name, ok := data["name"].(string); !ok || name == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'name' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	} else {
		errors = append(errors, validateAgentName(name, filePath, contents)...)
	}

	if description, ok := data["description"].(string); !ok || description == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'description' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "description"),
		})
	} else if !strings.Contains(strings.ToUpper(description), "PROACTIVELY") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding 'Use PROACTIVELY when...' pattern in description for agent discoverability",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     textutil.FindFrontmatterFieldLine(contents, "description"),
		})
	}

	return errors
}

// validateAgentColor validates the color field.
func validateAgentColor(data map[string]any, filePath string) []cue.ValidationError {
	color, ok := data["color"].(string)
	if !ok {
		return nil
	}

	validColors := map[string]bool{
		"red": true, "blue": true, "green": true, "yellow": true,
		"purple": true, "orange": true, "pink": true, "cyan": true,
		"gray": true, "magenta": true, "white": true,
	}
	if validColors[color] {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("Invalid color '%s'. Valid colors are: red, blue, green, yellow, purple, orange, pink, cyan, gray, magenta, white", color),
		Severity: "suggestion",
		Source:   cue.SourceCClintObserve,
	}}
}

// validateAgentMemory validates the memory scope field.
func validateAgentMemory(data map[string]any, filePath, contents string) []cue.ValidationError {
	memory, ok := data["memory"].(string)
	if !ok {
		return nil
	}

	validScopes := map[string]bool{"user": true, "project": true, "local": true}
	if validScopes[memory] {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("Invalid memory scope '%s'. Valid scopes: user, project, local", memory),
		Severity: "error",
		Source:   cue.SourceAnthropicDocs,
		Line:     textutil.FindFrontmatterFieldLine(contents, "memory"),
	}}
}

// validateAgentModel validates the model field.
func validateAgentModel(data map[string]any, filePath, contents string) []cue.ValidationError {
	model, ok := data["model"].(string)
	if !ok {
		return nil
	}

	if validModelPattern.MatchString(model) {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("Unknown model %q. Valid models: haiku, sonnet, opus, inherit, opusplan (with optional version suffix like sonnet[1m])", model),
		Severity: "warning",
		Source:   cue.SourceCClintObserve,
		Line:     textutil.FindFrontmatterFieldLine(contents, "model"),
	}}
}

// validateAgentMCPServersField validates the mcpServers field.
func validateAgentMCPServersField(data map[string]any, filePath, contents string) []cue.ValidationError {
	mcpServers, ok := data["mcpServers"]
	if !ok {
		return nil
	}
	return validateAgentMCPServers(mcpServers, filePath, contents)
}

// validateAgentPermissionMode validates the permissionMode field.
func validateAgentPermissionMode(data map[string]any, filePath, contents string) []cue.ValidationError {
	permMode, ok := data["permissionMode"].(string)
	if !ok {
		return nil
	}

	validModes := map[string]bool{
		"default": true, "acceptEdits": true, "delegate": true,
		"dontAsk": true, "bypassPermissions": true, "plan": true,
	}
	if validModes[permMode] {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("Invalid permissionMode value %q; must be one of: default, acceptEdits, delegate, dontAsk, bypassPermissions, plan", permMode),
		Severity: "error",
		Source:   cue.SourceAnthropicDocs,
		Line:     textutil.FindFrontmatterFieldLine(contents, "permissionMode"),
	}}
}

// validateAgentMaxTurns validates the maxTurns field is a positive integer.
func validateAgentMaxTurns(data map[string]any, filePath, contents string) []cue.ValidationError {
	maxTurns, ok := data["maxTurns"]
	if !ok {
		return nil
	}

	switch v := maxTurns.(type) {
	case int:
		if v > 0 {
			return nil
		}
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Invalid maxTurns value %d; must be a positive integer", v),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "maxTurns"),
		}}
	case float64:
		if v > 0 && v == float64(int(v)) {
			return nil
		}
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Invalid maxTurns value %v; must be a positive integer", v),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "maxTurns"),
		}}
	default:
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Invalid maxTurns value %v; must be a positive integer", maxTurns),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "maxTurns"),
		}}
	}
}

// validateAgentAutonomousPattern checks for maxTurns + dontAsk pattern.
func validateAgentAutonomousPattern(data map[string]any, filePath, contents string) []cue.ValidationError {
	_, hasMaxTurns := data["maxTurns"]
	if !hasMaxTurns {
		return nil
	}

	permMode, ok := data["permissionMode"].(string)
	if !ok || permMode != "dontAsk" {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  "Agent uses maxTurns with permissionMode 'dontAsk' - this is a common pattern for autonomous sub-agents.",
		Severity: "info",
		Source:   cue.SourceCClintObserve,
		Line:     textutil.FindFrontmatterFieldLine(contents, "maxTurns"),
	}}
}

// validateAgentHooks validates the hooks field.
func validateAgentHooks(data map[string]any, filePath string) []cue.ValidationError {
	hooks, ok := data["hooks"]
	if !ok {
		return nil
	}
	return ValidateComponentHooks(hooks, filePath)
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
		// Check comma-separated string
		for _, tool := range editingTools {
			if strings.Contains(v, tool) {
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

// inlineMethodologyPatterns detect methodology that should be in skills.
var inlineMethodologyPatterns = []struct {
	pattern string
	message string
}{
	{`score\s*=\s*\([^)]{20,}`, "Inline scoring formula detected - should be 'See skill for scoring'"},
	{`\|\s*CRITICAL\s*\|[^|]*\|\s*HIGH\s*\|`, "Inline priority matrix detected - move to skill"},
	{`(?i)tier\s*(bonus|1|2|3|4)[^|]*\+\s*\d+`, "Tier scoring details inline - move to skill"},
	{`regexp\.(MustCompile|Compile)\s*\(`, "Detection patterns inline - move to skill"},
}

// checkAgentInlineMethodology detects methodology patterns that belong in skills.
func checkAgentInlineMethodology(contents, filePath string) []cue.ValidationError {
	var suggestions []cue.ValidationError
	for _, ip := range inlineMethodologyPatterns {
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

// countTools returns the number of tool entries in a frontmatter tools value.
// It handles both a comma-separated string and a []any slice.
func countTools(tools any) int {
	switch v := tools.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return 0
		}
		return len(strings.Split(v, ","))
	case []any:
		return len(v)
	}
	return 0
}

// validateBodyToolMismatch checks whether tool names declared in frontmatter
// are referenced in the agent body. Declared tools not mentioned in the body
// may indicate a mismatch between the agent's declared capabilities and its
// documented workflow.
//
// Agents that declare 8 or more tools are treated as general-purpose agents
// where tools represent capability scope rather than per-instruction references,
// so the check is skipped entirely for those agents.
func validateBodyToolMismatch(data map[string]any, filePath, contents string) []cue.ValidationError {
	if countTools(data["tools"]) >= 8 {
		return nil
	}

	declaredTools := extractDeclaredTools(data["tools"])
	if declaredTools == nil {
		return nil
	}

	body := extractBody(contents)
	reported := make(map[string]bool)
	var suggestions []cue.ValidationError

	for toolName := range declaredTools {
		if toolName == "*" {
			return nil // wildcard — no mismatch possible
		}
		if reported[toolName] {
			continue
		}

		lines := strings.Split(body, "\n")
		foundPositive := false
		for _, line := range lines {
			if bodyToolNegativePattern.MatchString(line) {
				continue
			}
			if containsToolReference(line, toolName) {
				foundPositive = true
				break
			}
		}

		if !foundPositive {
			reported[toolName] = true
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Tool %q declared in frontmatter but not referenced in agent body — verify the tool is actually used", toolName),
				Severity: cue.SeveritySuggestion,
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return suggestions
}

// extractDeclaredTools parses the frontmatter tools field into a name set.
// Returns nil if the field is absent.
func extractDeclaredTools(tools any) map[string]bool {
	if tools == nil {
		return nil
	}

	result := make(map[string]bool)

	switch v := tools.(type) {
	case string:
		for _, part := range strings.Split(v, ",") {
			name := strings.TrimSpace(part)
			if name != "" {
				result[name] = true
			}
		}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				result[s] = true
			}
		}
	default:
		return nil
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// containsToolReference reports whether line contains a reference to toolName
// using word-boundary logic: the preceding char must not be a letter and the
// following char must not be a lowercase letter (allows camelCase boundaries
// like "WebSearch" vs "Search").
func containsToolReference(line, toolName string) bool {
	idx := 0
	for {
		pos := strings.Index(line[idx:], toolName)
		if pos < 0 {
			return false
		}
		abs := idx + pos

		// Check preceding character.
		if abs > 0 {
			prev := rune(line[abs-1])
			if (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z') {
				idx = abs + 1
				continue
			}
		}

		// Check following character.
		after := abs + len(toolName)
		if after < len(line) {
			next := rune(line[after])
			if next >= 'a' && next <= 'z' {
				idx = abs + 1
				continue
			}
		}

		return true
	}
}

// validateAgentName checks name format, reserved words, and filename match.
func validateAgentName(name, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	if !isKebabCase(name) {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Name must contain only lowercase letters, numbers, and hyphens",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Reserved word check - FROM ANTHROPIC DOCS
	reservedWords := map[string]bool{"anthropic": true, "claude": true}
	if reservedWords[strings.ToLower(name)] {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Name '%s' is a reserved word and cannot be used", name),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Check if name matches filename - OUR OBSERVATION
	filename := extractBaseFilename(filePath)
	if name != filename {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Name %q doesn't match filename %q", name, filename),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	return errors
}

// isKebabCase returns true if the string contains only lowercase letters, digits, and hyphens.
func isKebabCase(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

// extractBaseFilename extracts the base name without extension from a file path.
// e.g., ".claude/agents/test-agent.md" -> "test-agent"
func extractBaseFilename(filePath string) string {
	filename := filePath
	if idx := strings.LastIndex(filename, "/"); idx != -1 {
		filename = filename[idx+1:]
	}
	filename = strings.TrimSuffix(filename, ".md")
	if idx := strings.LastIndex(filename, "/"); idx != -1 {
		filename = filename[idx+1:]
	}
	return strings.TrimSuffix(filename, ".md")
}

// validateAgentMCPServers validates the mcpServers field is an array of non-empty strings.
func validateAgentMCPServers(mcpServers any, filePath, contents string) []cue.ValidationError {
	arr, isArr := mcpServers.([]any)
	if !isArr {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "mcpServers must be an array of server name strings",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "mcpServers"),
		}}
	}

	var errors []cue.ValidationError
	for i, item := range arr {
		s, isStr := item.(string)
		if !isStr || s == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers[%d] must be a non-empty string", i),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     textutil.FindFrontmatterFieldLine(contents, "mcpServers"),
			})
		}
	}
	return errors
}
