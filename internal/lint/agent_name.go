package lint

import (
	"fmt"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/textutil"
)

// validateAgentName checks name format, reserved words, and filename match.
func validateAgentName(name, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	if !isKebabCase(name) {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Name must contain only lowercase letters, numbers, and hyphens",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Reserved word check - FROM ANTHROPIC DOCS
	reservedWords := map[string]bool{"anthropic": true, "claude": true}
	if reservedWords[strings.ToLower(name)] {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Name '%s' is a reserved word and cannot be used", name),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Check if name matches filename - OUR OBSERVATION
	filename := extractBaseFilename(filePath)
	if name != filename {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Name %q doesn't match filename %q", name, filename),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	return errors
}

// isKebabCase returns true if the string contains only lowercase letters, digits, and hyphens.
func isKebabCase(s string) bool {
	for _, c := range s {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-') {
			return false
		}
	}
	return true
}

// extractBaseFilename extracts the base name without extension from a file path.
// e.g., ".claude/agents/test-agent.md" -> "test-agent"
func extractBaseFilename(filePath string) string {
	filename := filePath
	if idx := strings.LastIndex(filename, "/"); idx != -1 {
		filename = filename[idx+1:]
	}
	return strings.TrimSuffix(filename, ".md")
}
