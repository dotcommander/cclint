package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// LintPlugins runs linting on plugin manifest files using the generic linter.
func LintPlugins(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewPluginLinter()), nil
}

// knownPluginFields lists valid plugin.json fields per Anthropic docs
var knownPluginFields = map[string]bool{
	"name":         true, // Required: plugin name
	"description":  true, // Required: plugin description
	"version":      true, // Recommended: semver version
	"author":       true, // Required: author object with name
	"homepage":     true, // Optional: project URL
	"repository":   true, // Optional: source code URL
	"license":      true, // Optional: SPDX license identifier
	"keywords":     true, // Optional: discoverability tags
	"readme":       true, // Optional: path to README
	"commands":     true, // Optional: command definitions
	"agents":       true, // Optional: agent definitions
	"skills":       true, // Optional: skill definitions
	"hooks":        true, // Optional: hook configurations
	"mcpServers":   true, // Optional: MCP server configurations
	"outputStyles": true, // Optional: output style configurations
	"lspServers":   true, // Optional: LSP server configurations
}

// pluginPathFields lists plugin fields that contain file paths
var pluginPathFields = []string{
	"commands", "agents", "skills", "hooks", "mcpServers", "outputStyles", "lspServers",
}

// validatePluginSpecific implements plugin-specific validation rules
func validatePluginSpecific(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check for unknown fields
	for key := range data {
		if !knownPluginFields[key] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown plugin field '%s'", key),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
				Line:     FindJSONFieldLine(contents, key),
			})
		}
	}

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
		if err := ValidateSemver(version, filePath, FindJSONFieldLine(contents, "version")); err != nil {
			errors = append(errors, *err)
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

	// Validate paths in path-bearing fields
	errors = append(errors, validatePluginPaths(data, filePath, contents)...)

	// Best practice checks
	errors = append(errors, validatePluginBestPractices(filePath, contents, data)...)

	return errors
}

// validatePluginPaths checks that path-bearing fields use relative paths
func validatePluginPaths(data map[string]interface{}, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	for _, field := range pluginPathFields {
		value, ok := data[field]
		if !ok {
			continue
		}
		paths := extractPaths(value)
		for _, p := range paths {
			errors = append(errors, checkPath(p, field, filePath, contents)...)
		}
	}

	return errors
}

// extractPaths collects path strings from various JSON structures:
// string, []string, []object (extracts string values), or map (extracts string values).
func extractPaths(value interface{}) []string {
	var paths []string
	switch v := value.(type) {
	case string:
		paths = append(paths, v)
	case []interface{}:
		for _, item := range v {
			switch elem := item.(type) {
			case string:
				paths = append(paths, elem)
			case map[string]interface{}:
				for _, mv := range elem {
					if s, ok := mv.(string); ok {
						paths = append(paths, s)
					}
				}
			}
		}
	case map[string]interface{}:
		for _, mv := range v {
			if s, ok := mv.(string); ok {
				paths = append(paths, s)
			}
		}
	}
	return paths
}

// checkPath validates a single path string for relative path requirements.
func checkPath(path string, field string, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	if strings.HasPrefix(path, "/") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Plugin paths must be relative (start with \"./\"): found \"%s\"", path),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, field),
		})
	}

	if strings.Contains(path, "..") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Path '%s' in '%s' contains '..', which risks traversal outside the plugin root", path, field),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
			Line:     FindJSONFieldLine(contents, field),
		})
	}

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
