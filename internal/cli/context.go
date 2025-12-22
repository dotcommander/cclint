package cli

import (
	"fmt"
	"log"
	"strings"

	"github.com/carlrannaberg/cclint/internal/cue"
	"github.com/carlrannaberg/cclint/internal/discovery"
	"github.com/carlrannaberg/cclint/internal/frontend"
	"github.com/carlrannaberg/cclint/internal/project"
)

// LintContext runs linting on CLAUDE.md files
func LintContext(rootPath string, quiet bool, verbose bool) (*LintSummary, error) {
	summary := &LintSummary{}

	// Initialize components
	// projectDetector := &project.Detector{} // Not needed
	validator := cue.NewValidator()
	discoverer := discovery.NewFileDiscovery(rootPath, false)

	// Load schemas
	schemaDir := "schemas"
	if err := validator.LoadSchemas(schemaDir); err != nil {
		log.Printf("Error loading schemas: %v", err)
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

	// Filter context files
	var contextFiles []discovery.File
	for _, file := range files {
		if file.Type == discovery.FileTypeContext {
			contextFiles = append(contextFiles, file)
		}
	}

	summary.TotalFiles = len(contextFiles)

	// Process each context file
	for _, file := range contextFiles {
		result := LintResult{
			File:    file.RelPath,
			Type:    "context",
			Success: true,
		}

		// Parse frontmatter (CLAUDE.md may have frontmatter)
		fm, err := frontend.ParseYAMLFrontmatter(file.Contents)
		if err != nil {
			result.Errors = append(result.Errors, cue.ValidationError{
				File:     file.RelPath,
				Message:  fmt.Sprintf("Error parsing frontmatter: %v", err),
				Severity: "error",
			})
		}

		// Parse markdown content
		data := map[string]interface{}{
			"title":       fm.Data["title"],
			"description": fm.Data["description"],
			"sections":    parseMarkdownSections(file.Contents),
		}

		// Validate with CUE
		if true { // CUE schemas not loaded yet
			errors, err := validator.ValidateClaudeMD(data)
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
		errors := validateContextSpecific(data, file.RelPath)
		result.Errors = append(result.Errors, errors...)
		summary.TotalErrors += len(errors)

		if len(result.Errors) == 0 {
			summary.SuccessfulFiles++
		} else {
			result.Success = false
			summary.FailedFiles++
		}

		summary.Results = append(summary.Results, result)

		if verbose {
			log.Printf("Processed %s: %d errors", file.RelPath, len(result.Errors))
		}
	}

	return summary, nil
}

// parseMarkdownSections parses markdown content into sections
func parseMarkdownSections(content string) []map[string]interface{} {
	var sections []map[string]interface{}

	// This is a simplified parser - in practice, would use a proper markdown parser
	lines := []string{}
	content = content
	if fm, err := frontend.ParseYAMLFrontmatter(content); err == nil {
		lines = append(lines, strings.Split(fm.Body, "\n")...)
	} else {
		lines = append(lines, strings.Split(content, "\n")...)
	}

	currentSection := map[string]interface{}{}
	inSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "# ") {
			// New section found
			if inSection {
				sections = append(sections, currentSection)
			}
			currentSection = map[string]interface{}{
				"heading": strings.TrimPrefix(line, "# "),
				"content": "",
			}
			inSection = true
		} else if inSection && line != "" {
			if contentStr, ok := currentSection["content"].(string); ok {
				currentSection["content"] = contentStr + line + "\n"
			} else {
				currentSection["content"] = line + "\n"
			}
		}
	}

	// Add the last section
	if inSection {
		sections = append(sections, currentSection)
	}

	return sections
}

// validateContextSpecific implements context-specific validation rules
func validateContextSpecific(data map[string]interface{}, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check if sections are present
	if sections, ok := data["sections"].([]interface{}); ok && len(sections) > 0 {
		for i, section := range sections {
			if sectionMap, ok := section.(map[string]interface{}); ok {
				// Check heading
				if heading, exists := sectionMap["heading"]; !exists || heading == "" {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Section %d: missing heading", i),
						Severity: "warning",
					})
				}

				// Check content
				if content, exists := sectionMap["content"]; !exists || content == "" {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Section %d: missing content", i),
						Severity: "warning",
					})
				}
			}
		}
	} else {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "No sections found in CLAUDE.md",
			Severity: "suggestion",
		})
	}

	return errors
}