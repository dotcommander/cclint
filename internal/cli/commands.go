package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
	"github.com/dotcommander/cclint/internal/project"
	"github.com/dotcommander/cclint/internal/scoring"
)

// LintCommands runs linting on command files
func LintCommands(rootPath string, quiet bool, verbose bool) (*LintSummary, error) {
	summary := &LintSummary{}

	// Find project root first
	if rootPath == "" {
		var err error
		rootPath, err = project.FindProjectRoot(".")
		if err != nil {
			return nil, fmt.Errorf("error finding project root: %w", err)
		}
	}

	// Initialize components
	validator := cue.NewValidator()
	discoverer := discovery.NewFileDiscovery(rootPath, false)

	// Load schemas
	schemaDir := "schemas"
	if err := validator.LoadSchemas(schemaDir); err != nil {
		log.Printf("Error loading schemas: %v", err)
		// Continue with basic validation
	}

	// Discover files
	files, err := discoverer.DiscoverFiles()
	if err != nil {
		return nil, fmt.Errorf("error discovering files: %w", err)
	}

	// Filter command files
	var commandFiles []discovery.File
	for _, file := range files {
		if file.Type == discovery.FileTypeCommand {
			commandFiles = append(commandFiles, file)
		}
	}

	summary.TotalFiles = len(commandFiles)

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
				errors, err := validator.ValidateCommand(fm.Data)
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

		if verbose {
			log.Printf("Processed %s: %d errors", file.RelPath, len(result.Errors))
		}
	}

	return summary, nil
}

// validateCommandSpecific implements command-specific validation rules
func validateCommandSpecific(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Note: name is optional in frontmatter - it's derived from filename
	// Check name format if present
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
			})
		}
	}

	return errors
}

// validateCommandBestPractices checks opinionated best practices
func validateCommandBestPractices(filePath string, contents string) []cue.ValidationError {
	var suggestions []cue.ValidationError

	// Count total lines
	lines := strings.Count(contents, "\n")
	if lines > 50 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Command is %d lines. Best practice: keep commands under 50 lines - delegate to specialist agents instead of implementing logic directly.", lines),
			Severity: "suggestion",
		})
	}

	// Check for direct implementation patterns
	if strings.Contains(contents, "## Implementation") || strings.Contains(contents, "### Steps") {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Command contains implementation steps. Consider delegating to a specialist agent instead.",
			Severity: "suggestion",
		})
	}

	// Check for missing allowed-tools when Task tool is mentioned
	if strings.Contains(contents, "Task(") && !strings.Contains(contents, "allowed-tools:") {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Command uses Task() but lacks 'allowed-tools' permission. Add 'allowed-tools: Task' to frontmatter.",
			Severity: "suggestion",
		})
	}

	// Check for required sections
	if !strings.Contains(contents, "## Usage") && !strings.Contains(contents, "## Workflow") {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Command lacks '## Usage' or '## Workflow' section. Document how to use this command.",
			Severity: "suggestion",
		})
	}

	// Check for semantic routing table
	if !strings.Contains(contents, "| User") && !strings.Contains(contents, "User Question") && !strings.Contains(contents, "Quick Reference") {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding a semantic routing table for discoverability (| User Question | Action |).",
			Severity: "suggestion",
		})
	}

	// Check for Examples section
	if !strings.Contains(contents, "## Example") && !strings.Contains(contents, "## Examples") {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding an Examples section to show typical usage patterns.",
			Severity: "suggestion",
		})
	}

	return suggestions
}