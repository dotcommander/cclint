package cli

import (
	"fmt"
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

	return suggestions
}