package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// knownPluginFields lists valid plugin.json fields per Anthropic docs
var knownPluginFields = map[string]bool{
	"$schema":        true, // Optional: JSON Schema reference (v2.1.120+)
	"name":           true, // Required: plugin name
	"description":    true, // Required: plugin description
	"version":        true, // Recommended: semver version
	"author":         true, // Required: author object with name
	"homepage":       true, // Optional: project URL
	"repository":     true, // Optional: source code URL
	"license":        true, // Optional: SPDX license identifier
	"keywords":       true, // Optional: discoverability tags
	"readme":         true, // Optional: path to README
	"commands":       true, // Optional: command definitions
	"agents":         true, // Optional: agent definitions
	"skills":         true, // Optional: skill definitions
	"hooks":          true, // Optional: hook configurations
	"mcpServers":     true, // Optional: MCP server configurations
	"outputStyles":   true, // Optional: output style configurations
	"lspServers":     true, // Optional: LSP server configurations
	"monitors":       true, // Optional: background monitor configurations (deprecated top-level v2.1.129+ - prefer experimental.monitors)
	"themes":         true, // Optional: color theme components (deprecated top-level v2.1.129+ - prefer experimental.themes)
	"experimental":   true, // Optional: wrapper for experimental components (themes, monitors) (v2.1.129+)
	"defaultEnabled": true, // Optional: set false to disable plugin by default (v2.1.154+)
}

func validateUnknownPluginFields(data map[string]any, filePath, contents string) []cue.ValidationError {
	return checkUnknownFields(data, filePath, contents, unknownFieldCheck{
		known:    knownPluginFields,
		label:    "plugin field",
		suffix:   "",
		findLine: FindJSONFieldLine,
	})
}

func validatePluginName(data map[string]any, filePath, contents string) []cue.ValidationError {
	name, ok := data["name"].(string)
	if !ok || name == "" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Required field 'name' is missing or empty",
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "name"),
		}}
	}

	var errors []cue.ValidationError
	if reservedNames[strings.ToLower(name)] {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Name '%s' is a reserved word and cannot be used", name),
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "name"),
		})
	}

	if len(name) > 64 {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Name exceeds 64 character limit (%d chars)", len(name)),
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "name"),
		})
	}

	return errors
}

func validatePluginDescription(data map[string]any, filePath, contents string) []cue.ValidationError {
	desc, ok := data["description"].(string)
	if !ok || desc == "" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Required field 'description' is missing or empty",
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "description"),
		}}
	}

	if len(desc) > 1024 {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Description exceeds 1024 character limit (%d chars)", len(desc)),
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "description"),
		}}
	}

	return nil
}

func validatePluginVersion(data map[string]any, filePath, contents string) []cue.ValidationError {
	version, ok := data["version"].(string)
	if !ok || version == "" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Consider adding 'version' field in semver format (e.g., 1.0.0)",
			Severity: cue.SeveritySuggestion,
			Source:   cue.SourceCClintObserve,
			Line:     FindJSONFieldLine(contents, "version"),
		}}
	}

	if err := ValidateSemver(version, filePath, FindJSONFieldLine(contents, "version")); err != nil {
		return []cue.ValidationError{*err}
	}

	return nil
}

func validatePluginAuthor(data map[string]any, filePath, contents string) []cue.ValidationError {
	author, ok := data["author"].(map[string]any)
	if !ok {
		// External plugins (marketplaces, cache) may omit author — demote to warning
		severity := "error"
		if isExternalPlugin(filePath) {
			severity = "warning"
		}
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Required field 'author' is missing",
			Severity: severity,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "author"),
		}}
	}

	if authorName, ok := author["name"].(string); !ok || authorName == "" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Required field 'author.name' is missing or empty",
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "name"),
		}}
	}

	return nil
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
