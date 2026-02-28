package lint

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// validateRules validates the rules array in settings.json.
// Each entry must be a non-empty string containing a valid glob pattern.
// Warns on suspicious patterns like absolute paths.
func validateRules(rules any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	rulesArray, ok := rules.([]any)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "rules must be an array of glob pattern strings",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	for i, entry := range rulesArray {
		str, ok := entry.(string)
		if !ok || str == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("rules[%d]: each entry must be a non-empty string", i),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		// Validate glob syntax using the shared helper from rules.go
		if err := validateGlobPattern(str); err != nil {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("rules[%d]: invalid glob pattern %q: %v", i, str, err),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		// Warn on absolute paths (not portable across machines)
		if filepath.IsAbs(str) || strings.HasPrefix(str, "/") {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("rules[%d]: absolute path %q is not portable; use relative glob patterns", i, str),
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return errors
}

// validateMatcherToolName validates a toolName pattern from a hook matcher.
// Patterns look like "Bash(npm*)", "Edit", "mcp__server_tool", etc.
// Returns errors if the base tool name is unrecognized or the glob portion is invalid.
func validateMatcherToolName(toolNamePattern string, location string, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	if toolNamePattern == "" {
		return errors
	}

	// Extract the base tool name
	baseTool := extractToolName(toolNamePattern)
	if !isKnownTool(baseTool) {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("%s: unrecognized tool name '%s' in toolName pattern '%s'", location, baseTool, toolNamePattern),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Validate the parenthetical glob pattern if present
	errors = append(errors, validateToolNameGlob(toolNamePattern, location, filePath)...)

	return errors
}

// validateToolNameGlob validates the glob portion inside parentheses of a toolName pattern.
func validateToolNameGlob(toolNamePattern, location, filePath string) []cue.ValidationError {
	openIdx := strings.Index(toolNamePattern, "(")
	if openIdx <= 0 {
		return nil
	}

	closeIdx := strings.LastIndex(toolNamePattern, ")")
	if closeIdx <= openIdx {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("%s: unclosed parenthesis in toolName pattern '%s'", location, toolNamePattern),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	globPart := toolNamePattern[openIdx+1 : closeIdx]
	if globPart == "" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("%s: empty glob pattern in parentheses for toolName '%s'", location, toolNamePattern),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
		}}
	}

	if err := validateGlobPattern(globPart); err != nil {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("%s: invalid glob in toolName '%s': %v", location, toolNamePattern, err),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		}}
	}

	return nil
}
