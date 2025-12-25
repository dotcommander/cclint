package cli

import (
	"fmt"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
	"github.com/dotcommander/cclint/internal/scoring"
)

// LintCommands runs linting on command files
func LintCommands(rootPath string, quiet bool, verbose bool) (*LintSummary, error) {
	// Initialize shared context
	ctx, err := NewLinterContext(rootPath, quiet, verbose)
	if err != nil {
		return nil, err
	}

	// Filter command files
	commandFiles := ctx.FilterFilesByType(discovery.FileTypeCommand)
	summary := ctx.NewSummary(len(commandFiles))

	// Process each command file
	for _, file := range commandFiles {
		result := LintResult{
			File:    file.RelPath,
			Type:    "command",
			Success: true,
		}

		// Parse frontmatter
		fm, err := frontend.ParseYAMLFrontmatter(file.Contents)
		if err != nil {
			result.Errors = append(result.Errors, cue.ValidationError{
				File:     file.RelPath,
				Message:  fmt.Sprintf("Error parsing frontmatter: %v", err),
				Severity: "error",
			})
			result.Success = false
			summary.FailedFiles++
			summary.TotalErrors++
		} else {
			// Validate with CUE
			if true { // CUE schemas not loaded yet
				errors, err := ctx.Validator.ValidateCommand(fm.Data)
				if err != nil {
					result.Errors = append(result.Errors, cue.ValidationError{
						File:     file.RelPath,
						Message:  fmt.Sprintf("Validation error: %v", err),
						Severity: "error",
					})
				}
				result.Errors = append(result.Errors, errors...)
				summary.TotalErrors += len(errors)
			}

			// Additional validation rules
			errors := validateCommandSpecific(fm.Data, file.RelPath, file.Contents)
			result.Errors = append(result.Errors, errors...)
			summary.TotalErrors += len(errors)

			// Cross-file validation (missing agents, unused tools)
			crossErrors := ctx.CrossValidator.ValidateCommand(file.RelPath, file.Contents, fm.Data)
			result.Errors = append(result.Errors, crossErrors...)
			summary.TotalErrors += len(crossErrors)

			// Best practice checks
			suggestions := validateCommandBestPractices(file.RelPath, file.Contents)
			result.Suggestions = append(result.Suggestions, suggestions...)
			summary.TotalSuggestions += len(suggestions)

			if len(result.Errors) == 0 {
				summary.SuccessfulFiles++
			} else {
				result.Success = false
				summary.FailedFiles++
			}

			// Score command quality
			scorer := scoring.NewCommandScorer()
			score := scorer.Score(file.Contents, fm.Data, fm.Body)
			result.Quality = &score

			// Get improvement recommendations
			result.Improvements = GetCommandImprovements(file.Contents, fm.Data)
		}

		summary.Results = append(summary.Results, result)
		ctx.LogProcessed(file.RelPath, len(result.Errors))
	}

	return summary, nil
}

// validateCommandSpecific implements command-specific validation rules
func validateCommandSpecific(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

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

	return errors
}

// validateCommandBestPractices checks opinionated best practices
func validateCommandBestPractices(filePath string, contents string) []cue.ValidationError {
	var suggestions []cue.ValidationError

	// Count total lines (±10% tolerance: 50 base + 5 = 55) - OUR OBSERVATION
	lines := strings.Count(contents, "\n")
	if lines > 55 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Command is %d lines. Best practice: keep commands under ~55 lines (50±10%%) - delegate to specialist agents instead of implementing logic directly.", lines),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
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