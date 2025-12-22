package cli

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/carlrannaberg/cclint/internal/cue"
	"github.com/carlrannaberg/cclint/internal/discovery"
	"github.com/carlrannaberg/cclint/internal/frontend"
	"github.com/carlrannaberg/cclint/internal/project"
)

// LintResult represents a single linting result
type LintResult struct {
	File       string
	Type       string
	Errors     []cue.ValidationError
	Warnings   []cue.ValidationError
	Suggestions []cue.ValidationError
	Success    bool
	Duration   int64
}

// LintSummary summarizes all linting results
type LintSummary struct {
	ProjectRoot      string
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

// LintAgents runs linting on agent files
func LintAgents(rootPath string, quiet bool, verbose bool) (*LintSummary, error) {
	summary := &LintSummary{}

	// Initialize components
	validator := cue.NewValidator()
	discoverer := discovery.NewFileDiscovery(rootPath, false)

	// Load schemas - try multiple paths
	schemaPaths := []string{
		"schemas",
		"./schemas",
		"/schemas",
	}
	var schemaLoaded bool
	for _, path := range schemaPaths {
		if err := validator.LoadSchemas(path); err == nil {
			schemaLoaded = true
			break
		}
	}
	if !schemaLoaded {
		log.Printf("Warning: CUE schemas not loaded, using Go validation")
		// Continue with basic validation
	}

	// Find project root
	if rootPath == "" {
		var err error
		rootPath, err = project.FindProjectRoot(".")
		if err != nil {
			return nil, fmt.Errorf("error finding project root: %w", err)
		}
	}

	// Discover files
	files, err := discoverer.DiscoverFiles()
	if err != nil {
		return nil, fmt.Errorf("error discovering files: %w", err)
	}

	// Filter agent files
	var agentFiles []discovery.File
	for _, file := range files {
		if file.Type == discovery.FileTypeAgent {
			agentFiles = append(agentFiles, file)
		}
	}

	summary.TotalFiles = len(agentFiles)

	// Process each agent file
	for _, file := range agentFiles {
		result := LintResult{
			File:    file.RelPath,
			Type:    "agent",
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
				errors, err := validator.ValidateAgent(fm.Data)
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

			// Additional validation rules - separate errors and suggestions
			allIssues := validateAgentSpecific(fm.Data, file.RelPath, file.Contents)
			for _, issue := range allIssues {
				if issue.Severity == "suggestion" {
					result.Suggestions = append(result.Suggestions, issue)
					summary.TotalSuggestions++
				} else {
					result.Errors = append(result.Errors, issue)
					summary.TotalErrors++
				}
			}

			if len(result.Errors) == 0 {
				summary.SuccessfulFiles++
			} else {
				result.Success = false
				summary.FailedFiles++
			}
		}

		summary.Results = append(summary.Results, result)

		if verbose {
			log.Printf("Processed %s: %d errors", file.RelPath, len(result.Errors))
		}
	}

	return summary, nil
}

// validateAgentSpecific implements agent-specific validation rules
func validateAgentSpecific(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check required fields
	if name, ok := data["name"].(string); !ok || name == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'name' is missing or empty",
			Severity: "error",
		})
	}

	if description, ok := data["description"].(string); !ok || description == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'description' is missing or empty",
			Severity: "error",
		})
	}

	// Check name format
	if name, ok := data["name"].(string); ok {
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
				Message:  "Name must contain only lowercase letters, numbers, and hyphens",
				Severity: "error",
			})
		}

		// Check if name matches filename
		// Extract filename from path (e.g., ".claude/agents/test-agent.md" -> "test-agent")
		filename := filePath
		if idx := strings.LastIndex(filename, "/"); idx != -1 {
			filename = filename[idx+1:]
		}
		// Remove .md extension
		if strings.HasSuffix(filename, ".md") {
			filename = filename[:len(filename)-3]
		}
		// For nested paths, use the last component
		if idx := strings.LastIndex(filename, "/"); idx != -1 {
			filename = filename[idx+1:]
		}
		// Remove .md extension again if present
		if strings.HasSuffix(filename, ".md") {
			filename = filename[:len(filename)-3]
		}

		if name != filename {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Name %q doesn't match filename %q", name, filename),
				Severity: "suggestion",
			})
		}
	}

	// Check valid colors
	if color, ok := data["color"].(string); ok {
		validColors := map[string]bool{
			"red":     true,
			"blue":    true,
			"green":   true,
			"yellow":  true,
			"purple":  true,
			"orange":  true,
			"pink":    true,
			"cyan":    true,
		}
		if !validColors[color] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Invalid color '%s'. Valid colors are: red, blue, green, yellow, purple, orange, pink, cyan", color),
				Severity: "error",
			})
		}
	}

	// Best practice checks
	errors = append(errors, validateAgentBestPractices(filePath, contents, data)...)

	return errors
}

// validateAgentBestPractices checks opinionated best practices for agents
func validateAgentBestPractices(filePath string, contents string, data map[string]interface{}) []cue.ValidationError {
	var suggestions []cue.ValidationError

	// Count total lines
	lines := strings.Count(contents, "\n")
	if lines > 200 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Agent is %d lines. Best practice: keep agents under 200 lines - move methodology to skills instead.", lines),
			Severity: "suggestion",
		})
	}

	// Check for missing model specification
	if _, hasModel := data["model"]; !hasModel {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent lacks 'model' specification. Consider adding 'model: sonnet' or appropriate model for optimal performance.",
			Severity: "suggestion",
		})
	}

	// Check for triggers array
	if _, hasTriggers := data["triggers"]; !hasTriggers {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent lacks 'triggers' array. Add keyword triggers for automatic activation.",
			Severity: "suggestion",
		})
	}

	// Check for proactive_triggers (complement to triggers)
	if _, hasProactiveTriggers := data["proactive_triggers"]; !hasProactiveTriggers {
		_, hasTriggers := data["triggers"]
		if hasTriggers {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  "Agent has 'triggers' but no 'proactive_triggers'. Consider adding proactive trigger phrases for better activation.",
				Severity: "suggestion",
			})
		}
	}

	// Check for performance optimization fields
	if _, hasContextIsolation := data["context_isolation"]; !hasContextIsolation {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent lacks 'context_isolation: true'. Consider adding for cleaner context management.",
			Severity: "suggestion",
		})
	}

	// Check for required sections
	hasFoundation := strings.Contains(contents, "## Foundation")
	hasWorkflow := strings.Contains(contents, "## Workflow")
	hasExpectedOutput := strings.Contains(contents, "## Expected Output")
	hasSuccessCriteria := strings.Contains(contents, "## Success Criteria")

	if !hasFoundation {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent lacks '## Foundation' section. Should define skill loading and initialization.",
			Severity: "suggestion",
		})
	}

	if !hasWorkflow {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent lacks '## Workflow' section. Should define phased execution plan.",
			Severity: "suggestion",
		})
	}

	if !hasExpectedOutput {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent lacks '## Expected Output' section. Should define what success looks like.",
			Severity: "suggestion",
		})
	}

	if !hasSuccessCriteria {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent lacks '## Success Criteria' checklist. Should define completion conditions.",
			Severity: "suggestion",
		})
	}

	// Check for Skill loading pattern
	if strings.Contains(contents, "Skill(") {
		// Has Skill calls - good practice
	} else if strings.Contains(contents, "## Foundation") || strings.Contains(contents, "## Workflow") {
		// Has structure but no explicit Skill loading
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Agent has methodology sections. Consider extracting to a skill and loading with Skill() tool for reusability.",
			Severity: "suggestion",
		})
	}

	// Check for Use PROACTIVELY pattern in description
	if desc, hasDesc := data["description"].(string); hasDesc {
		if !strings.Contains(strings.ToLower(desc), "proactively") && !strings.Contains(strings.ToLower(desc), "use proactively") {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  "Description lacks 'Use PROACTIVELY when...' pattern. Add to clarify activation scenarios.",
				Severity: "suggestion",
			})
		}
	}

	return suggestions
}