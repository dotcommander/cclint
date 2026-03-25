package lint

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/textutil"
)

// validModelPattern matches known Claude Code model values.
// Bare names: haiku, sonnet, opus, inherit, opusplan.
// Optional version suffix in brackets: sonnet[1m], haiku[2].
// Full model IDs: claude-* (e.g. claude-opus-4-5, claude-sonnet-4-6).
var validModelPattern = regexp.MustCompile(`^(haiku|sonnet|opus|inherit|opusplan)(\[\w+\])?$|^claude-[a-z0-9-]+$`)

// validateUnknownFields checks for unsupported frontmatter fields.
func validateUnknownFields(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError
	validList := sortedMapKeys(knownAgentFields)
	for key := range data {
		if !knownAgentFields[key] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown frontmatter field '%s'. Valid fields: %s", key, validList),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
				Line:     textutil.FindFrontmatterFieldLine(contents, key),
			})
		}
	}
	return errors
}

// validateRequiredFields validates name and description requirements.
func validateRequiredFields(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	if name, ok := data["name"].(string); !ok || name == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'name' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	} else {
		errors = append(errors, validateAgentName(name, filePath, contents)...)
	}

	if description, ok := data["description"].(string); !ok || description == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'description' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "description"),
		})
	} else if !strings.Contains(strings.ToUpper(description), "PROACTIVELY") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding 'Use PROACTIVELY when...' pattern in description for agent discoverability",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     textutil.FindFrontmatterFieldLine(contents, "description"),
		})
	}

	return errors
}

// validateAgentColor validates the color field.
func validateAgentColor(data map[string]any, filePath string) []cue.ValidationError {
	color, ok := data["color"].(string)
	if !ok {
		return nil
	}

	validColors := map[string]bool{
		"red": true, "blue": true, "green": true, "yellow": true,
		"purple": true, "orange": true, "pink": true, "cyan": true,
		"gray": true, "magenta": true, "white": true,
	}
	if validColors[color] {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("Invalid color '%s'. Valid colors are: red, blue, green, yellow, purple, orange, pink, cyan, gray, magenta, white", color),
		Severity: "suggestion",
		Source:   cue.SourceCClintObserve,
	}}
}

// validateAgentMemory validates the memory scope field.
func validateAgentMemory(data map[string]any, filePath, contents string) []cue.ValidationError {
	memory, ok := data["memory"].(string)
	if !ok {
		return nil
	}

	validScopes := map[string]bool{"user": true, "project": true, "local": true}
	if validScopes[memory] {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("Invalid memory scope '%s'. Valid scopes: user, project, local", memory),
		Severity: "error",
		Source:   cue.SourceAnthropicDocs,
		Line:     textutil.FindFrontmatterFieldLine(contents, "memory"),
	}}
}

// validateAgentModel validates the model field.
func validateAgentModel(data map[string]any, filePath, contents string) []cue.ValidationError {
	model, ok := data["model"].(string)
	if !ok {
		return nil
	}

	if validModelPattern.MatchString(model) {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("Unknown model %q. Valid models: haiku, sonnet, opus, inherit, opusplan (with optional version suffix like sonnet[1m]), or full model ID (claude-*)", model),
		Severity: "warning",
		Source:   cue.SourceCClintObserve,
		Line:     textutil.FindFrontmatterFieldLine(contents, "model"),
	}}
}

// validateAgentMCPServersField validates the mcpServers field.
func validateAgentMCPServersField(data map[string]any, filePath, contents string) []cue.ValidationError {
	mcpServers, ok := data["mcpServers"]
	if !ok {
		return nil
	}
	return validateAgentMCPServers(mcpServers, filePath, contents)
}

// validateAgentMCPServers validates the mcpServers field is an array of non-empty strings.
func validateAgentMCPServers(mcpServers any, filePath, contents string) []cue.ValidationError {
	arr, isArr := mcpServers.([]any)
	if !isArr {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "mcpServers must be an array of server name strings",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "mcpServers"),
		}}
	}

	var errors []cue.ValidationError
	for i, item := range arr {
		s, isStr := item.(string)
		if !isStr || s == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("mcpServers[%d] must be a non-empty string", i),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     textutil.FindFrontmatterFieldLine(contents, "mcpServers"),
			})
		}
	}
	return errors
}

// validateAgentPermissionMode validates the permissionMode field.
func validateAgentPermissionMode(data map[string]any, filePath, contents string) []cue.ValidationError {
	permMode, ok := data["permissionMode"].(string)
	if !ok {
		return nil
	}

	validModes := map[string]bool{
		"default": true, "acceptEdits": true, "delegate": true,
		"dontAsk": true, "bypassPermissions": true, "plan": true,
	}
	if validModes[permMode] {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  fmt.Sprintf("Invalid permissionMode value %q; must be one of: default, acceptEdits, delegate, dontAsk, bypassPermissions, plan", permMode),
		Severity: "error",
		Source:   cue.SourceAnthropicDocs,
		Line:     textutil.FindFrontmatterFieldLine(contents, "permissionMode"),
	}}
}

// validateAgentMaxTurns validates the maxTurns field is a positive integer.
func validateAgentMaxTurns(data map[string]any, filePath, contents string) []cue.ValidationError {
	maxTurns, ok := data["maxTurns"]
	if !ok {
		return nil
	}

	switch v := maxTurns.(type) {
	case int:
		if v > 0 {
			return nil
		}
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Invalid maxTurns value %d; must be a positive integer", v),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "maxTurns"),
		}}
	case float64:
		if v > 0 && v == float64(int(v)) {
			return nil
		}
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Invalid maxTurns value %v; must be a positive integer", v),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "maxTurns"),
		}}
	default:
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Invalid maxTurns value %v; must be a positive integer", maxTurns),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "maxTurns"),
		}}
	}
}

// validateAgentAutonomousPattern checks for maxTurns + dontAsk pattern.
func validateAgentAutonomousPattern(data map[string]any, filePath, contents string) []cue.ValidationError {
	_, hasMaxTurns := data["maxTurns"]
	if !hasMaxTurns {
		return nil
	}

	permMode, ok := data["permissionMode"].(string)
	if !ok || permMode != "dontAsk" {
		return nil
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  "Agent uses maxTurns with permissionMode 'dontAsk' - this is a common pattern for autonomous sub-agents.",
		Severity: "info",
		Source:   cue.SourceCClintObserve,
		Line:     textutil.FindFrontmatterFieldLine(contents, "maxTurns"),
	}}
}
