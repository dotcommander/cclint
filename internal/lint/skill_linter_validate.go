package lint

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/textutil"
)

// checkUnknownSkillFields checks for unknown frontmatter fields in skill files.
func checkUnknownSkillFields(data map[string]any, filePath, contents string) []cue.ValidationError {
	return checkUnknownFields(data, filePath, contents, unknownFieldCheck{
		known:    knownSkillFields,
		label:    "frontmatter field",
		suffix:   ". See https://agentskills.io/specification for valid fields",
		findLine: textutil.FindFrontmatterFieldLine,
	})
}

// validateSkillContextField validates the context field in skill frontmatter.
func validateSkillContextField(data map[string]any, filePath, contents string) []cue.ValidationError {
	if ctxVal, ok := data["context"]; ok {
		ctxStr, isStr := ctxVal.(string)
		if !isStr || ctxStr != "fork" {
			return []cue.ValidationError{{
				File:     filePath,
				Message:  fmt.Sprintf("context field must be 'fork' (got '%v')", ctxVal),
				Severity: cue.SeverityError,
				Source:   cue.SourceAnthropicDocs,
				Line:     textutil.FindFrontmatterFieldLine(contents, "context"),
			}}
		}
	}
	return nil
}

// validateSkillBooleanFields validates boolean fields in skill frontmatter.
// Claude Code also accepts yes/no/on/off/1/0 (case-insensitive) for these
// frontmatter booleans as of v2.1.218.
func validateSkillBooleanFields(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	for _, field := range []string{"user-invocable", "disable-model-invocation", "background"} {
		value, ok := data[field]
		if !ok || isClaudeBoolean(value) {
			continue
		}

		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("%s field must be a boolean or one of yes/no/on/off/1/0 (got '%v')", field, value),
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, field),
		})
	}

	return errors
}

func isClaudeBoolean(value any) bool {
	switch v := value.(type) {
	case bool:
		return true
	case int:
		return v == 0 || v == 1
	case int8:
		return v == 0 || v == 1
	case int16:
		return v == 0 || v == 1
	case int32:
		return v == 0 || v == 1
	case int64:
		return v == 0 || v == 1
	case uint:
		return v == 0 || v == 1
	case uint8:
		return v == 0 || v == 1
	case uint16:
		return v == 0 || v == 1
	case uint32:
		return v == 0 || v == 1
	case uint64:
		return v == 0 || v == 1
	case string:
		switch strings.ToLower(strings.TrimSpace(v)) {
		case "yes", "no", "on", "off", "1", "0":
			return true
		default:
			return false
		}
	default:
		return false
	}
}

// validateSkillArgumentHint validates the argument-hint field in skill frontmatter.
func validateSkillArgumentHint(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError
	if ahVal, ok := data["argument-hint"]; ok {
		ahStr, isStr := ahVal.(string)
		if !isStr {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("argument-hint field must be a string (got '%v')", ahVal),
				Severity: cue.SeverityError,
				Source:   cue.SourceAnthropicDocs,
				Line:     textutil.FindFrontmatterFieldLine(contents, "argument-hint"),
			})
		} else if strings.TrimSpace(ahStr) == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  "argument-hint field is empty - provide a hint for autocomplete (e.g., 'PR number or URL')",
				Severity: cue.SeverityWarning,
				Source:   cue.SourceAnthropicDocs,
				Line:     textutil.FindFrontmatterFieldLine(contents, "argument-hint"),
			})
		}
	}
	return errors
}

// checkSkillFrontmatter checks if the skill has frontmatter and suggests adding it.
func checkSkillFrontmatter(filePath, contents string) []cue.ValidationError {
	if strings.HasPrefix(contents, "---") {
		return nil
	}

	skillName := extractSkillName(contents, filePath)
	suggestion := "Add YAML frontmatter with name and description (description is critical for skill discovery)"
	if skillName != "" {
		suggestion = fmt.Sprintf("Add frontmatter: ---\nname: %s\ndescription: Brief summary of what this skill does\n--- (description critical for discovery)", skillName)
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  suggestion,
		Severity: cue.SeveritySuggestion,
		Source:   cue.SourceAnthropicDocs,
	}}
}

// validateSkillName checks reserved words, hyphen placement, consecutive hyphens, and directory match.
func validateSkillName(name, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	if reservedNames[strings.ToLower(name)] {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Name '%s' is a reserved word and cannot be used", name),
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Rule 048: Name cannot start or end with hyphen (agentskills.io spec)
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Skill name '%s' cannot start or end with a hyphen", name),
			Severity: cue.SeverityError,
			Source:   cue.SourceAgentSkillsIO,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Rule 049: Name cannot contain consecutive hyphens (agentskills.io spec)
	if strings.Contains(name, "--") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Skill name '%s' contains consecutive hyphens (--) which are not allowed", name),
			Severity: cue.SeverityError,
			Source:   cue.SourceAgentSkillsIO,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Rule 050: Name must match parent directory name (agentskills.io spec)
	parentDir := filepath.Base(filepath.Dir(filePath))
	isSpecialDir := parentDir == "." || parentDir == "skills" || parentDir == ".claude"
	if !isSpecialDir && name != parentDir {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Skill name '%s' must match parent directory name '%s' (agentskills.io spec: name field)", name, parentDir),
			Severity: cue.SeverityError,
			Source:   cue.SourceAgentSkillsIO,
			Line:     textutil.FindFrontmatterFieldLine(contents, "name"),
		})
	}

	return errors
}

// validateSkillAgentField validates the agent frontmatter field and its relationship with context.
func validateSkillAgentField(agentVal any, data map[string]any, filePath, contents string) []cue.ValidationError {
	agentStr, isStr := agentVal.(string)
	if !isStr || strings.TrimSpace(agentStr) == "" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "agent field must be a non-empty string",
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "agent"),
		}}
	}

	// Warn if agent is set but context is not "fork"
	ctxStr, _ := data["context"].(string)
	if ctxStr != "fork" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "agent field is set but context is not 'fork' - consider adding 'context: fork' for sub-agent execution",
			Severity: cue.SeverityWarning,
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(contents, "agent"),
		}}
	}

	return nil
}
