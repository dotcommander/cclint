package cli

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/scoring"
)

// LintPlugins runs linting on plugin manifest files
func LintPlugins(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	// Initialize shared context
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}

	// Filter plugin files
	pluginFiles := ctx.FilterFilesByType(discovery.FileTypePlugin)
	summary := ctx.NewSummary(len(pluginFiles))

	// Process each plugin file
	for _, file := range pluginFiles {
		result := LintResult{
			File:    file.RelPath,
			Type:    "plugin",
			Success: true,
		}

		// Parse JSON
		var data map[string]interface{}
		if err := json.Unmarshal([]byte(file.Contents), &data); err != nil {
			result.Errors = append(result.Errors, cue.ValidationError{
				File:     file.RelPath,
				Message:  fmt.Sprintf("Error parsing JSON: %v", err),
				Severity: "error",
			})
			result.Success = false
			summary.FailedFiles++
			summary.TotalErrors++
		} else {
			// Validate plugin-specific rules
			// NOTE: Only show errors for plugins until we can improve the suggestions
			allIssues := validatePluginSpecific(data, file.RelPath, file.Contents)
			for _, issue := range allIssues {
				if issue.Severity == "error" {
					result.Errors = append(result.Errors, issue)
					summary.TotalErrors++
				}
				// Skip suggestions and warnings for plugins
			}

			// Secrets detection - keep as errors only for plugins
			secretWarnings := detectSecrets(file.Contents, file.RelPath)
			for _, w := range secretWarnings {
				// Promote secrets to errors since they're important
				w.Severity = "error"
				result.Errors = append(result.Errors, w)
				summary.TotalErrors++
			}

			if len(result.Errors) == 0 {
				summary.SuccessfulFiles++
			} else {
				result.Success = false
				summary.FailedFiles++
			}

			// Score plugin quality
			scorer := scoring.NewPluginScorer()
			score := scorer.Score(file.Contents, data, "")
			result.Quality = &score

			// Get improvement recommendations
			result.Improvements = GetPluginImprovements(file.Contents, data)
		}

		summary.Results = append(summary.Results, result)
		ctx.LogProcessed(file.RelPath, len(result.Errors))
	}

	return summary, nil
}

// validatePluginSpecific implements plugin-specific validation rules
func validatePluginSpecific(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check required fields - FROM ANTHROPIC DOCS
	if name, ok := data["name"].(string); !ok || name == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'name' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "name"),
		})
	} else {
		// Reserved word check
		reservedWords := map[string]bool{"anthropic": true, "claude": true}
		if reservedWords[strings.ToLower(name)] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Name '%s' is a reserved word and cannot be used", name),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindJSONFieldLine(contents, "name"),
			})
		}

		// Character limit check
		if len(name) > 64 {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Name exceeds 64 character limit (%d chars)", len(name)),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindJSONFieldLine(contents, "name"),
			})
		}
	}

	if description, ok := data["description"].(string); !ok || description == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'description' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "description"),
		})
	} else {
		// Character limit check
		if len(description) > 1024 {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Description exceeds 1024 character limit (%d chars)", len(description)),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindJSONFieldLine(contents, "description"),
			})
		}
	}

	// Version field - recommended (Anthropic's own plugins omit this, using marketplace.json instead)
	if version, ok := data["version"].(string); !ok || version == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding 'version' field in semver format (e.g., 1.0.0)",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     FindJSONFieldLine(contents, "version"),
		})
	} else {
		// Validate semver format if present
		semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`)
		if !semverPattern.MatchString(version) {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Version '%s' should be in semver format (e.g., 1.0.0)", version),
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
				Line:     FindJSONFieldLine(contents, "version"),
			})
		}
	}

	// Author.name - required
	if author, ok := data["author"].(map[string]interface{}); !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'author' is missing",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "author"),
		})
	} else {
		if authorName, ok := author["name"].(string); !ok || authorName == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  "Required field 'author.name' is missing or empty",
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindJSONFieldLine(contents, "name"),
			})
		}
	}

	// Best practice checks
	errors = append(errors, validatePluginBestPractices(filePath, contents, data)...)

	return errors
}

// validatePluginBestPractices checks opinionated best practices for plugins
func validatePluginBestPractices(filePath string, contents string, data map[string]interface{}) []cue.ValidationError {
	var suggestions []cue.ValidationError

	// Check for homepage
	if _, ok := data["homepage"].(string); !ok {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding 'homepage' field with project URL",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Check for repository
	if _, ok := data["repository"].(string); !ok {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding 'repository' field with source code URL",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Check for license
	if _, ok := data["license"].(string); !ok {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding 'license' field (e.g., MIT, Apache-2.0)",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Check for keywords
	if keywords, ok := data["keywords"].([]interface{}); !ok || len(keywords) == 0 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding 'keywords' array for discoverability",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Check description length
	if desc, ok := data["description"].(string); ok && len(desc) < 50 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Description is only %d chars - consider expanding for clarity", len(desc)),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     FindJSONFieldLine(contents, "description"),
		})
	}

	return suggestions
}

// FindJSONFieldLine finds the line number of a JSON field
func FindJSONFieldLine(content string, fieldName string) int {
	lines := strings.Split(content, "\n")
	pattern := fmt.Sprintf(`"%s"\s*:`, fieldName)
	re := regexp.MustCompile(pattern)
	for i, line := range lines {
		if re.MatchString(line) {
			return i + 1
		}
	}
	return 0
}

// GetPluginImprovements returns specific improvement recommendations for plugins
func GetPluginImprovements(content string, data map[string]interface{}) []ImprovementRecommendation {
	var recs []ImprovementRecommendation

	// Check for missing optional but recommended fields
	if _, ok := data["homepage"]; !ok {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add 'homepage' field with project URL",
			PointValue:  5,
			Severity:    SeverityLow,
		})
	}

	if _, ok := data["repository"]; !ok {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add 'repository' field with source code URL",
			PointValue:  5,
			Severity:    SeverityLow,
		})
	}

	if _, ok := data["license"]; !ok {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add 'license' field (e.g., MIT, Apache-2.0)",
			PointValue:  5,
			Severity:    SeverityLow,
		})
	}

	if keywords, ok := data["keywords"].([]interface{}); !ok || len(keywords) == 0 {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add 'keywords' array for discoverability",
			PointValue:  5,
			Severity:    SeverityLow,
		})
	}

	if _, ok := data["readme"]; !ok {
		recs = append(recs, ImprovementRecommendation{
			Description: "Add 'readme' field pointing to README file",
			PointValue:  3,
			Severity:    SeverityLow,
		})
	}

	// Check description quality
	if desc, ok := data["description"].(string); ok {
		if len(desc) < 50 {
			recs = append(recs, ImprovementRecommendation{
				Description: "Expand description to at least 50 characters",
				PointValue:  5,
				Severity:    SeverityMedium,
			})
		}
	}

	return recs
}
