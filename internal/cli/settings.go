package cli

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/carlrannaberg/cclint/internal/cue"
	"github.com/carlrannaberg/cclint/internal/discovery"
	"github.com/carlrannaberg/cclint/internal/project"
)

// LintSettings runs linting on settings files
func LintSettings(rootPath string, quiet bool, verbose bool) (*LintSummary, error) {
	summary := &LintSummary{}

	// Initialize components
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

	// Filter settings files
	var settingsFiles []discovery.File
	for _, file := range files {
		if file.Type == discovery.FileTypeSettings {
			settingsFiles = append(settingsFiles, file)
		}
	}

	summary.TotalFiles = len(settingsFiles)

	// Process each settings file
	for _, file := range settingsFiles {
		result := LintResult{
			File:    file.RelPath,
			Type:    "settings",
			Success: true,
		}

		// Parse JSON content
		var data map[string]interface{}
		if err := parseJSON(file.Contents, &data); err != nil {
			result.Errors = append(result.Errors, cue.ValidationError{
				File:     file.RelPath,
				Message:  fmt.Sprintf("Error parsing JSON: %v", err),
				Severity: "error",
			})
			result.Success = false
			summary.FailedFiles++
			summary.TotalErrors++
		} else {
			// Validate with CUE
			if true { // CUE schemas not loaded yet
				errors, err := validator.ValidateSettings(data)
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
			errors := validateSettingsSpecific(data, file.RelPath)
			result.Errors = append(result.Errors, errors...)
			summary.TotalErrors += len(errors)

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

// parseJSON parses JSON content into the provided interface
func parseJSON(content string, out interface{}) error {
	return json.Unmarshal([]byte(content), out)
}

// validateSettingsSpecific implements settings-specific validation rules
func validateSettingsSpecific(data map[string]interface{}, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check hooks structure if present
	if hooks, ok := data["hooks"]; ok {
		if hooksArray, ok := hooks.([]interface{}); ok {
			for i, hook := range hooksArray {
				if hookMap, ok := hook.(map[string]interface{}); ok {
					// Check matcher field
					if _, exists := hookMap["matcher"]; !exists {
						errors = append(errors, cue.ValidationError{
							File:     filePath,
							Message:  fmt.Sprintf("Hook %d: missing required field 'matcher'", i),
							Severity: "error",
						})
					}

					// Check hooks field
					if innerHooks, exists := hookMap["hooks"]; exists {
						if innerHooksArray, ok := innerHooks.([]interface{}); ok {
							for j, innerHook := range innerHooksArray {
								if innerHookMap, ok := innerHook.(map[string]interface{}); ok {
									if _, exists := innerHookMap["type"]; !exists {
										errors = append(errors, cue.ValidationError{
											File:     filePath,
											Message:  fmt.Sprintf("Hook %d inner hook %d: missing required field 'type'", i, j),
											Severity: "error",
										})
									}
									if _, exists := innerHookMap["command"]; !exists {
										errors = append(errors, cue.ValidationError{
											File:     filePath,
											Message:  fmt.Sprintf("Hook %d inner hook %d: missing required field 'command'", i, j),
											Severity: "error",
										})
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return errors
}