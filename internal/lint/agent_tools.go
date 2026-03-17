package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// bodyToolNegativePattern matches lines that explicitly disclaim a tool (e.g. "do not use Bash").
var bodyToolNegativePattern = regexp.MustCompile(`(?i)\b(do not use|don't use|never use|avoid using|not use)\b`)

// countTools returns the number of tool entries in a frontmatter tools value.
// It handles both a comma-separated string and a []any slice.
func countTools(tools any) int {
	switch v := tools.(type) {
	case string:
		if strings.TrimSpace(v) == "" {
			return 0
		}
		return len(strings.Split(v, ","))
	case []any:
		return len(v)
	}
	return 0
}

// validateBodyToolMismatch checks whether tool names declared in frontmatter
// are referenced in the agent body. Declared tools not mentioned in the body
// may indicate a mismatch between the agent's declared capabilities and its
// documented workflow.
//
// Agents that declare 8 or more tools are treated as general-purpose agents
// where tools represent capability scope rather than per-instruction references,
// so the check is skipped entirely for those agents.
func validateBodyToolMismatch(data map[string]any, filePath, contents string) []cue.ValidationError {
	if countTools(data["tools"]) >= 8 {
		return nil
	}

	declaredTools := extractDeclaredTools(data["tools"])
	if declaredTools == nil {
		return nil
	}

	body := extractBody(contents)
	lines := strings.Split(body, "\n")
	reported := make(map[string]bool)
	var suggestions []cue.ValidationError

	for toolName := range declaredTools {
		if toolName == "*" {
			return nil // wildcard — no mismatch possible
		}
		if reported[toolName] {
			continue
		}
		foundPositive := false
		for _, line := range lines {
			if bodyToolNegativePattern.MatchString(line) {
				continue
			}
			if containsToolReference(line, toolName) {
				foundPositive = true
				break
			}
		}

		if !foundPositive {
			reported[toolName] = true
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Tool %q declared in frontmatter but not referenced in agent body — verify the tool is actually used", toolName),
				Severity: cue.SeveritySuggestion,
				Source:   cue.SourceCClintObserve,
			})
		}
	}

	return suggestions
}

// extractDeclaredTools parses the frontmatter tools field into a name set.
// Returns nil if the field is absent.
func extractDeclaredTools(tools any) map[string]bool {
	if tools == nil {
		return nil
	}

	result := make(map[string]bool)

	switch v := tools.(type) {
	case string:
		for _, part := range strings.Split(v, ",") {
			name := strings.TrimSpace(part)
			if name != "" {
				result[name] = true
			}
		}
	case []any:
		for _, item := range v {
			if s, ok := item.(string); ok && s != "" {
				result[s] = true
			}
		}
	default:
		return nil
	}

	if len(result) == 0 {
		return nil
	}
	return result
}

// containsToolReference reports whether line contains a reference to toolName
// using word-boundary logic: the preceding char must not be a letter and the
// following char must not be a lowercase letter (allows camelCase boundaries
// like "WebSearch" vs "Search").
func containsToolReference(line, toolName string) bool {
	idx := 0
	for {
		pos := strings.Index(line[idx:], toolName)
		if pos < 0 {
			return false
		}
		abs := idx + pos

		// Check preceding character.
		if abs > 0 {
			prev := rune(line[abs-1])
			if (prev >= 'a' && prev <= 'z') || (prev >= 'A' && prev <= 'Z') {
				idx = abs + 1
				continue
			}
		}

		// Check following character.
		after := abs + len(toolName)
		if after < len(line) {
			next := rune(line[after])
			if next >= 'a' && next <= 'z' {
				idx = abs + 1
				continue
			}
		}

		return true
	}
}
