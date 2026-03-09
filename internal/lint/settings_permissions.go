package lint

import (
	"fmt"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/textutil"
)

// validatePermissions validates the permissions section of settings.json.
// Expected structure: {"allow": ["Bash(npm*)", ...], "deny": ["Bash(rm*)", ...]}
func validatePermissions(perms any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	permsMap, ok := perms.(map[string]any)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "permissions must be an object with optional 'allow' and 'deny' arrays",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	for key, val := range permsMap {
		if key != "allow" && key != "deny" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("permissions: unknown key '%s'. Only 'allow' and 'deny' are supported", key),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		errors = append(errors, validatePermissionEntries(val, key, filePath)...)
	}

	return errors
}

// validatePermissionEntries validates a single permission list (allow or deny).
func validatePermissionEntries(entries any, listName string, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	arr, ok := entries.([]any)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("permissions.%s must be an array of tool permission strings", listName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
		return errors
	}

	for i, entry := range arr {
		str, ok := entry.(string)
		if !ok || str == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("permissions.%s[%d]: each entry must be a non-empty string", listName, i),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			continue
		}

		// Extract tool name from patterns like "Bash(npm*)" or plain "Read"
		toolName := canonicalToolName(str)
		if !isKnownTool(toolName) {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("permissions.%s[%d]: unrecognized tool name '%s' in '%s'", listName, i, toolName, str),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return errors
}

// canonicalToolName returns the base tool name used for validation.
// "Bash(npm*)" -> "Bash", "Read" -> "Read", "mcp__foo" -> "mcp__"
func canonicalToolName(entry string) string {
	base := textutil.ExtractBaseToolName(entry)
	if strings.HasPrefix(base, "mcp__") {
		return "mcp__"
	}
	return base
}

// isKnownTool checks whether a tool name is in the known tools set.
func isKnownTool(name string) bool {
	if name == "mcp__" {
		return true
	}
	return textutil.KnownTools[name]
}
