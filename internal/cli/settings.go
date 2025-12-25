package cli

import (
	"encoding/json"
	"fmt"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

// LintSettings runs linting on settings files
func LintSettings(rootPath string, quiet bool, verbose bool) (*LintSummary, error) {
	// Initialize shared context
	ctx, err := NewLinterContext(rootPath, quiet, verbose)
	if err != nil {
		return nil, err
	}

	// Filter settings files
	settingsFiles := ctx.FilterFilesByType(discovery.FileTypeSettings)
	summary := ctx.NewSummary(len(settingsFiles))

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
				errors, err := ctx.Validator.ValidateSettings(data)
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
		ctx.LogProcessed(file.RelPath, len(result.Errors))
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