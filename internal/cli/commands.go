package cli

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// frontmatterDelimiter is the YAML frontmatter delimiter
const frontmatterDelimiter = "---"

// LintCommands runs linting on command files using the generic linter.
func LintCommands(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewCommandLinter()), nil
}

// knownCommandFields lists valid frontmatter fields per Anthropic docs
// Source: https://docs.anthropic.com/en/docs/claude-code/slash-commands
var knownCommandFields = map[string]bool{
	"name":                     true, // Optional: derived from filename if not set
	"description":              true, // Optional: command description
	"allowed-tools":            true, // Optional: tool access permissions
	"argument-hint":            true, // Optional: hint for command arguments
	"model":                    true, // Optional: model to use
	"disable-model-invocation": true, // Optional: prevent SlashCommand tool from calling
}

// validateCommandSpecific implements command-specific validation rules
func validateCommandSpecific(data map[string]any, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check for unknown frontmatter fields - helps catch fabricated/deprecated fields
	for key := range data {
		if !knownCommandFields[key] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown frontmatter field '%s'. Valid fields: name, description, allowed-tools, argument-hint, model, disable-model-invocation", key),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
				Line:     FindFrontmatterFieldLine(contents, key),
			})
		}
	}

	// Note: name is optional in frontmatter - it's derived from filename (per Anthropic docs)
	// Check name format if present - format rule from Anthropic docs
	if name, ok := data["name"].(string); ok && name != "" {
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
				Message:  "Name must be lowercase alphanumeric with hyphens only",
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
		}
	}

	// Validate tool field naming (commands use 'allowed-tools:', not 'tools:')
	errors = append(errors, ValidateToolFieldName(data, filePath, contents, "command")...)

	// Validate allowed-tools only contains permitted tools
	errors = append(errors, checkCommandToolAllowlist(data, filePath, contents)...)

	return errors
}

// commandAllowedTools is the set of tools commands are permitted to declare.
var commandAllowedTools = map[string]bool{
	"Task":            true,
	"Skill":           true,
	"AskUserQuestion": true,
}

// checkCommandToolAllowlist validates that allowed-tools only contains permitted tools.
func checkCommandToolAllowlist(data map[string]any, filePath string, contents string) []cue.ValidationError {
	tools, ok := data["allowed-tools"].(string)
	if !ok || tools == "" {
		return nil
	}

	line := FindFrontmatterFieldLine(contents, "allowed-tools")

	// Wildcard is not permitted
	if strings.TrimSpace(tools) == "*" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  `command declares wildcard "*" in allowed-tools — commands should only use Task, Skill, AskUserQuestion`,
			Severity: "error",
			Source:   cue.SourceCClintObserve,
			Line:     line,
		}}
	}

	var errors []cue.ValidationError
	for tool := range strings.SplitSeq(tools, ",") {
		tool = strings.TrimSpace(tool)
		if tool == "" {
			continue
		}
		base := extractBaseToolName(tool)
		if !commandAllowedTools[base] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("command declares tool %q in allowed-tools — commands should only use Task, Skill, AskUserQuestion", tool),
				Severity: "error",
				Source:   cue.SourceCClintObserve,
				Line:     line,
			})
		}
	}
	return errors
}

// checkSkillWithoutTaskDelegation warns when a command calls Skill() directly
// without going through Task() delegation.
func checkSkillWithoutTaskDelegation(filePath string, contents string, data map[string]any) []cue.ValidationError {
	body := extractBody(contents)

	if !strings.Contains(body, "Skill(") {
		return nil
	}
	if strings.Contains(body, "Task(") {
		return nil
	}

	// Also skip if allowed-tools declares both Skill and Task
	if tools, ok := data["allowed-tools"].(string); ok {
		hasSkill, hasTask := false, false
		for tool := range strings.SplitSeq(tools, ",") {
			base := extractBaseToolName(strings.TrimSpace(tool))
			if base == "Skill" {
				hasSkill = true
			}
			if base == "Task" {
				hasTask = true
			}
		}
		if hasSkill && hasTask {
			return nil
		}
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  "command dispatches to Skill() without Task() delegation — commands must delegate through agents",
		Severity: "error",
		Source:   cue.SourceCClintObserve,
	}}
}

// validateCommandBestPractices checks opinionated best practices
func validateCommandBestPractices(filePath string, contents string, data map[string]any) []cue.ValidationError {
	var suggestions []cue.ValidationError

	suggestions = append(suggestions, checkDescriptionXMLTags(data, filePath, contents)...)
	suggestions = append(suggestions, checkCommandSizeLimit(contents, filePath)...)
	suggestions = append(suggestions, checkImplementationPatterns(contents, filePath)...)
	suggestions = append(suggestions, checkAllowedToolsWithTask(contents, filePath)...)

	hasTaskDelegation := strings.Contains(contents, "Task(")
	suggestions = append(suggestions, checkBloatSections(contents, filePath, hasTaskDelegation)...)
	suggestions = append(suggestions, checkExcessiveExamples(contents, filePath)...)
	suggestions = append(suggestions, checkSuccessCriteriaFormat(contents, filePath)...)

	lines := strings.Count(contents, "\n")
	suggestions = append(suggestions, checkFatCommandUsage(contents, filePath, lines, hasTaskDelegation)...)

	suggestions = append(suggestions, validateCommandPreprocessing(filePath, contents)...)
	suggestions = append(suggestions, validateCommandSubstitution(filePath, contents, data)...)
	suggestions = append(suggestions, checkSkillWithoutTaskDelegation(filePath, contents, data)...)

	return suggestions
}

func checkDescriptionXMLTags(data map[string]any, filePath, contents string) []cue.ValidationError {
	if description, ok := data["description"].(string); ok {
		if xmlErr := DetectXMLTags(description, "Description", filePath, contents); xmlErr != nil {
			return []cue.ValidationError{*xmlErr}
		}
	}
	return nil
}

func checkCommandSizeLimit(contents, filePath string) []cue.ValidationError {
	if sizeErr := CheckSizeLimit(contents, 50, 0.10, "command", filePath); sizeErr != nil {
		return []cue.ValidationError{*sizeErr}
	}
	return nil
}

func checkImplementationPatterns(contents, filePath string) []cue.ValidationError {
	if strings.Contains(contents, "## Implementation") || strings.Contains(contents, "### Steps") {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Command contains implementation steps. Consider delegating to a specialist agent instead.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		}}
	}
	return nil
}

func checkAllowedToolsWithTask(contents, filePath string) []cue.ValidationError {
	if strings.Contains(contents, "Task(") && !strings.Contains(contents, "allowed-tools:") {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Command uses Task() but lacks 'allowed-tools' permission. Add 'allowed-tools: Task' to frontmatter.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		}}
	}
	return nil
}

func checkBloatSections(contents, filePath string, hasTaskDelegation bool) []cue.ValidationError {
	if !hasTaskDelegation {
		return nil
	}

	bloatSections := []struct {
		pattern string
		message string
	}{
		{"## Quick Reference", "Thin command has '## Quick Reference' - belongs in skill, not command"},
		{"## Usage", "Thin command has '## Usage' - agent has full context, remove"},
		{"## Workflow", "Thin command has '## Workflow' - duplicates agent content, remove"},
		{"## When to use", "Thin command has '## When to use' - belongs in description, remove"},
		{"## What it does", "Thin command has '## What it does' - belongs in description, remove"},
	}

	var suggestions []cue.ValidationError
	for _, section := range bloatSections {
		if strings.Contains(contents, section.pattern) {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  section.message,
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
			})
		}
	}
	return suggestions
}

func checkExcessiveExamples(contents, filePath string) []cue.ValidationError {
	exampleCount := strings.Count(contents, "```bash") + strings.Count(contents, "```shell")
	if exampleCount > 2 {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Command has %d code examples. Best practice: max 2 examples.", exampleCount),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		}}
	}
	return nil
}

func checkSuccessCriteriaFormat(contents, filePath string) []cue.ValidationError {
	hasSuccessSection := strings.Contains(contents, "## Success") || strings.Contains(contents, "Success criteria:")
	hasCheckboxes := strings.Contains(contents, "- [ ]")
	if hasSuccessSection && !hasCheckboxes {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Success criteria should use checkbox format '- [ ]' not prose",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		}}
	}
	return nil
}

func checkFatCommandUsage(contents, filePath string, lines int, hasTaskDelegation bool) []cue.ValidationError {
	if !hasTaskDelegation && lines > 40 && !strings.Contains(contents, "## Usage") && !strings.Contains(contents, "## Workflow") {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Fat command without Task delegation lacks '## Usage' section. Consider delegating to a specialist agent.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		}}
	}
	return nil
}

// dangerousCommandPatterns lists shell patterns that are obviously destructive.
// Each entry has a regex pattern and a human-readable description.
var dangerousCommandPatterns = []struct {
	pattern *regexp.Regexp
	message string
}{
	{regexp.MustCompile(`\brm\s+(-[a-zA-Z]*f[a-zA-Z]*\s+)?(-[a-zA-Z]*r[a-zA-Z]*\s+)?/(\s|$)`), "destructive 'rm' targeting root filesystem"},
	{regexp.MustCompile(`\brm\s+(-[a-zA-Z]*r[a-zA-Z]*\s+)?(-[a-zA-Z]*f[a-zA-Z]*\s+)?/(\s|$)`), "destructive 'rm' targeting root filesystem"},
	{regexp.MustCompile(`\bmkfs\b`), "'mkfs' formats a filesystem"},
	{regexp.MustCompile(`\bdd\b.*\bof=/dev/`), "'dd' writing to device"},
	{regexp.MustCompile(`:\(\)\{.*\|.*\}`), "fork bomb pattern"},
	{regexp.MustCompile(`>\s*/dev/sda`), "writing directly to disk device"},
	{regexp.MustCompile(`\bchmod\s+(-[a-zA-Z]*\s+)?777\s+/(\s|$)`), "'chmod 777 /' on root filesystem"},
}

// validateCommandPreprocessing checks !command preprocessing directives in command body.
// Claude Code commands can use !command syntax to execute a shell command and inject output.
func validateCommandPreprocessing(filePath string, contents string) []cue.ValidationError {
	var issues []cue.ValidationError

	lines := strings.Split(contents, "\n")
	inFrontmatter := false
	frontmatterDone := false
	inCodeBlock := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		// Update parsing state (frontmatter, code blocks)
		inFrontmatter, frontmatterDone, inCodeBlock = updateParsingState(
			trimmed, inFrontmatter, frontmatterDone, inCodeBlock,
		)

		// Skip lines inside frontmatter or code blocks
		if inFrontmatter || inCodeBlock {
			continue
		}

		// Check for preprocessing directives
		issues = appendPreprocessingIssues(issues, trimmed, filePath, i+1)
	}

	return issues
}

// updateParsingState updates the parsing state based on the current line.
// Returns the new state values.
func updateParsingState(trimmed string, inFrontmatter, frontmatterDone, inCodeBlock bool) (newInFrontmatter, newFrontmatterDone, newInCodeBlock bool) {
	// Track frontmatter boundaries
	if trimmed == frontmatterDelimiter {
		if !inFrontmatter && !frontmatterDone {
			return true, false, inCodeBlock
		}
		return false, true, inCodeBlock
	}

	// Track code blocks
	if strings.HasPrefix(trimmed, "```") {
		return inFrontmatter, frontmatterDone, !inCodeBlock
	}

	return inFrontmatter, frontmatterDone, inCodeBlock
}

// appendPreprocessingIssues checks for preprocessing directives and adds issues.
// Returns the updated issues slice.
func appendPreprocessingIssues(issues []cue.ValidationError, trimmed, filePath string, lineNum int) []cue.ValidationError {
	// Must start with ! to be a preprocessing directive
	if !strings.HasPrefix(trimmed, "!") {
		return issues
	}

	command := strings.TrimSpace(trimmed[1:])

	// Empty command after !
	if command == "" {
		return append(issues, cue.ValidationError{
			File:     filePath,
			Message:  "Empty preprocessing directive '!' with no command",
			Severity: "error",
			Source:   cue.SourceCClintObserve,
			Line:     lineNum,
		})
	}

	// Check for dangerous patterns
	return appendDangerousPatternIssue(issues, command, filePath, lineNum)
}

// appendDangerousPatternIssue checks for dangerous command patterns and adds an issue if found.
func appendDangerousPatternIssue(issues []cue.ValidationError, command, filePath string, lineNum int) []cue.ValidationError {
	for _, dp := range dangerousCommandPatterns {
		if dp.pattern.MatchString(command) {
			issues = append(issues, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Dangerous preprocessing command: %s", dp.message),
				Severity: "error",
				Source:   cue.SourceCClintObserve,
				Line:     lineNum,
			})
			break
		}
	}
	return issues
}
// positionalArgPattern matches $1, $2, ... $99 substitution variables.
var positionalArgPattern = regexp.MustCompile(`\$(\d+)`)

// argumentsPattern matches $ARGUMENTS substitution variable.
var argumentsPattern = regexp.MustCompile(`\$ARGUMENTS`)

// validateCommandSubstitution checks $ARGUMENTS and $N substitution variable usage.
// Commands using substitution should declare argument-hint for discoverability,
// positional args should be sequential, and high positional args are likely unintended.
func validateCommandSubstitution(filePath string, contents string, data map[string]any) []cue.ValidationError {
	var issues []cue.ValidationError

	// Extract body (skip frontmatter) for scanning
	body := extractBody(contents)

	hasArguments := argumentsPattern.MatchString(body)
	positionalMatches := positionalArgPattern.FindAllStringSubmatch(body, -1)

	// No substitution variables found - nothing to validate
	if !hasArguments && len(positionalMatches) == 0 {
		return nil
	}

	// Collect unique positional arg numbers
	positionalNums := collectPositionalArgs(positionalMatches)

	// Check: commands using substitution should have argument-hint for discoverability
	if _, hasHint := data["argument-hint"]; !hasHint {
		issues = append(issues, cue.ValidationError{
			File:     filePath,
			Message:  "Command uses substitution variables ($ARGUMENTS or $N) but lacks 'argument-hint' in frontmatter. Add argument-hint to describe expected arguments.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     GetFrontmatterEndLine(contents),
		})
	}

	// Check: sequential positional args (warn if $2 used without $1)
	if len(positionalNums) > 0 {
		sort.Ints(positionalNums)
		for idx, n := range positionalNums {
			if idx == 0 && n != 1 {
				issues = append(issues, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Positional argument $%d used without $1. Arguments should start at $1.", n),
					Severity: "warning",
					Source:   cue.SourceCClintObserve,
					Line:     findSubstitutionLine(contents, fmt.Sprintf("$%d", n)),
				})
				break
			}
			if idx > 0 && n > positionalNums[idx-1]+1 {
				issues = append(issues, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Positional argument gap: $%d used without $%d. Arguments should be sequential.", n, positionalNums[idx-1]+1),
					Severity: "warning",
					Source:   cue.SourceCClintObserve,
					Line:     findSubstitutionLine(contents, fmt.Sprintf("$%d", n)),
				})
				break
			}
		}

		// Check: warn on high positional args ($10+)
		maxArg := positionalNums[len(positionalNums)-1]
		if maxArg >= 10 {
			issues = append(issues, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("High positional argument $%d detected. Commands with 10+ arguments are likely unintended. Consider using $ARGUMENTS instead.", maxArg),
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
				Line:     findSubstitutionLine(contents, fmt.Sprintf("$%d", maxArg)),
			})
		}
	}

	return issues
}

// extractBody returns content after frontmatter delimiters.
func extractBody(contents string) string {
	parts := strings.SplitN(contents, frontmatterDelimiter, 3)
	if len(parts) >= 3 {
		return parts[2]
	}
	return contents
}

// collectPositionalArgs extracts unique positional argument numbers from regex matches.
func collectPositionalArgs(matches [][]string) []int {
	seen := make(map[int]bool)
	var nums []int
	for _, m := range matches {
		if len(m) < 2 {
			continue
		}
		n, err := strconv.Atoi(m[1])
		if err != nil || n == 0 {
			continue // $0 is not a user positional arg
		}
		if !seen[n] {
			seen[n] = true
			nums = append(nums, n)
		}
	}
	return nums
}

// findSubstitutionLine finds the first line number (1-based) containing a substitution pattern.
// Skips frontmatter and code blocks for accurate reporting.
func findSubstitutionLine(contents string, pattern string) int {
	lines := strings.Split(contents, "\n")
	inFrontmatter := false
	frontmatterDone := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == frontmatterDelimiter {
			if !inFrontmatter && !frontmatterDone {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				inFrontmatter = false
				frontmatterDone = true
				continue
			}
		}
		if inFrontmatter {
			continue
		}
		if strings.Contains(line, pattern) {
			return i + 1
		}
	}
	return 0
}
