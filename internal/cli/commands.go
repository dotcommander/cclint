package cli

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

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
func validateCommandSpecific(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
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

	return errors
}

// validateCommandBestPractices checks opinionated best practices
func validateCommandBestPractices(filePath string, contents string, data map[string]interface{}) []cue.ValidationError {
	var suggestions []cue.ValidationError

	// XML tag detection in text fields - FROM ANTHROPIC DOCS
	if description, ok := data["description"].(string); ok {
		if xmlErr := DetectXMLTags(description, "Description", filePath, contents); xmlErr != nil {
			suggestions = append(suggestions, *xmlErr)
		}
	}

	// Count total lines (±10% tolerance: 50 base) - OUR OBSERVATION
	lines := strings.Count(contents, "\n")
	if sizeErr := CheckSizeLimit(contents, 50, 0.10, "command", filePath); sizeErr != nil {
		suggestions = append(suggestions, *sizeErr)
	}

	// Check for direct implementation patterns - OUR OBSERVATION (thin command pattern)
	if strings.Contains(contents, "## Implementation") || strings.Contains(contents, "### Steps") {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Command contains implementation steps. Consider delegating to a specialist agent instead.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Check for missing allowed-tools when Task tool is mentioned - OUR OBSERVATION
	if strings.Contains(contents, "Task(") && !strings.Contains(contents, "allowed-tools:") {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Command uses Task() but lacks 'allowed-tools' permission. Add 'allowed-tools: Task' to frontmatter.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Thin command pattern: Commands should delegate to agents, not contain methodology.
	hasTaskDelegation := strings.Contains(contents, "Task(")

	// === BLOAT SECTIONS DETECTOR (thin commands only) ===
	if hasTaskDelegation {
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
	}

	// === EXCESSIVE EXAMPLES DETECTOR === - OUR OBSERVATION
	exampleCount := strings.Count(contents, "```bash") + strings.Count(contents, "```shell")
	if exampleCount > 2 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Command has %d code examples. Best practice: max 2 examples.", exampleCount),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// === SUCCESS CRITERIA FORMAT DETECTOR === - OUR OBSERVATION
	// Success criteria should be checkboxes, not prose
	hasSuccessSection := strings.Contains(contents, "## Success") || strings.Contains(contents, "Success criteria:")
	hasCheckboxes := strings.Contains(contents, "- [ ]")
	if hasSuccessSection && !hasCheckboxes {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Success criteria should use checkbox format '- [ ]' not prose",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Only suggest Usage section for FAT commands (>40 lines without Task delegation) - OUR OBSERVATION
	if !hasTaskDelegation && lines > 40 && !strings.Contains(contents, "## Usage") && !strings.Contains(contents, "## Workflow") {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Fat command without Task delegation lacks '## Usage' section. Consider delegating to a specialist agent.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// === PREPROCESSING DIRECTIVE VALIDATION === - OUR OBSERVATION
	suggestions = append(suggestions, validateCommandPreprocessing(filePath, contents)...)

	// === SUBSTITUTION VARIABLE VALIDATION === - OUR OBSERVATION
	suggestions = append(suggestions, validateCommandSubstitution(filePath, contents, data)...)

	return suggestions
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

		// Track frontmatter boundaries
		if trimmed == "---" {
			if !inFrontmatter && !frontmatterDone {
				inFrontmatter = true
				continue
			} else if inFrontmatter {
				inFrontmatter = false
				frontmatterDone = true
				continue
			}
		}

		// Skip lines inside frontmatter
		if inFrontmatter {
			continue
		}

		// Track code blocks to avoid false positives on examples
		if strings.HasPrefix(trimmed, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		if inCodeBlock {
			continue
		}

		// Detect !command preprocessing directives
		if !strings.HasPrefix(trimmed, "!") {
			continue
		}

		command := strings.TrimSpace(trimmed[1:])

		// Empty command after !
		if command == "" {
			issues = append(issues, cue.ValidationError{
				File:     filePath,
				Message:  "Empty preprocessing directive '!' with no command",
				Severity: "error",
				Source:   cue.SourceCClintObserve,
				Line:     i + 1,
			})
			continue
		}

		// Check for dangerous patterns
		for _, dp := range dangerousCommandPatterns {
			if dp.pattern.MatchString(command) {
				issues = append(issues, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Dangerous preprocessing command: %s", dp.message),
					Severity: "error",
					Source:   cue.SourceCClintObserve,
					Line:     i + 1,
				})
				break
			}
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
func validateCommandSubstitution(filePath string, contents string, data map[string]interface{}) []cue.ValidationError {
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
	parts := strings.SplitN(contents, "---", 3)
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
		if trimmed == "---" {
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