package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// knownSkillFields lists valid frontmatter fields per Anthropic docs and agentskills.io spec
// Sources:
//   - https://code.claude.com/docs/en/skills
//   - https://agentskills.io/specification
var knownSkillFields = map[string]bool{
	// Required fields
	"name":        true, // Optional: skill identifier (defaults to directory name)
	"description": true, // Recommended: what the skill does (critical for discovery)
	// Optional Claude Code fields
	"argument-hint":            true, // Optional: hint shown during autocomplete (e.g., "[issue-number]")
	"disable-model-invocation": true, // Optional: prevent Claude from auto-loading this skill
	"user-invocable":           true, // Optional: show in slash command menu (default true)
	"allowed-tools":            true, // Optional: tool access permissions
	"model":                    true, // Optional: model to use when skill is active
	"context":                  true, // Optional: "fork" for sub-agent context
	"agent":                    true, // Optional: agent type for execution
	"hooks":                    true, // Optional: skill-level hooks (PreToolUse, PostToolUse, Stop)
	// Optional agentskills.io fields
	"license":       true, // Optional: SPDX identifier or license file reference
	"compatibility": true, // Optional: environment requirements (max 500 chars)
	"metadata":      true, // Optional: arbitrary key-value mapping
	// Legacy/common fields
	"version": true, // Optional: semver version string
}

// LintSkills runs linting on skill files using the generic linter.
func LintSkills(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewSkillLinter()), nil
}

// validateSkillBestPractices checks opinionated best practices for skills
func validateSkillBestPractices(filePath string, contents string, fmData map[string]any) []cue.ValidationError {
	suggestions := validateBasicSkillFields(fmData, filePath, contents)
	suggestions = append(suggestions, validateSkillContentSections(filePath, contents)...)
	suggestions = append(suggestions, ValidateToolFieldName(fmData, filePath, contents, "skill")...)
	suggestions = append(suggestions, validateAgentSkillsOSpecFields(fmData, filePath, contents)...)
	suggestions = append(suggestions, ValidateSkillDirectory(filePath, contents)...)
	return suggestions
}

// validateBasicSkillFields validates basic skill frontmatter fields.
func validateBasicSkillFields(fmData map[string]any, filePath, contents string) []cue.ValidationError {
	var suggestions []cue.ValidationError
	var warnings []cue.ValidationError

	// XML tag detection and description checks - FROM ANTHROPIC DOCS
	if description, ok := fmData["description"].(string); ok {
		suggestions = append(suggestions, validateSkillDescription(description, filePath, contents)...)
	}

	// argument-hint length check - hints should be concise for autocomplete display
	if ah, ok := fmData["argument-hint"].(string); ok && len(ah) > 80 {
		warnings = append(warnings, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("argument-hint is %d chars - keep under 80 for readability in autocomplete", len(ah)),
			Severity: "warning",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "argument-hint"),
		})
	}

	// P3: Semver validation for version field - OUR OBSERVATION
	if version, ok := fmData["version"].(string); ok && version != "" {
		if err := ValidateSemver(version, filePath, FindFrontmatterFieldLine(contents, "version")); err != nil {
			warnings = append(warnings, *err)
		}
	}

	return append(suggestions, warnings...)
}

// validateSkillContentSections validates content sections of skills.
func validateSkillContentSections(filePath, contents string) []cue.ValidationError {
	var suggestions []cue.ValidationError

	// Check skill size - recommend references for large skills (±10% tolerance: 500 base) - OUR OBSERVATION
	if sizeErr := CheckSizeLimit(contents, 500, 0.10, "skill", filePath); sizeErr != nil {
		suggestions = append(suggestions, *sizeErr)
	}

	// Check for Anti-Patterns section
	suggestions = append(suggestions, checkSkillAntiPatternsSection(filePath, contents)...)

	// Check for Examples section
	suggestions = append(suggestions, checkSkillExamplesSection(filePath, contents)...)

	return suggestions
}

// validateAgentSkillsOSpecFields validates fields per agentskills.io spec.
func validateAgentSkillsOSpecFields(fmData map[string]any, filePath, contents string) []cue.ValidationError {
	var suggestions []cue.ValidationError
	var warnings []cue.ValidationError

	// Rule 052: Validate allowed-tools format
	if allowedTools, ok := fmData["allowed-tools"].(string); ok && allowedTools != "*" {
		toolPattern := regexp.MustCompile(`^[A-Z][a-zA-Z]+(\([^)]+\))?$`)
		tokens := strings.Fields(allowedTools)
		for _, token := range tokens {
			if !toolPattern.MatchString(token) {
				warnings = append(warnings, cue.ValidationError{
					File:     filePath,
					Message:  "allowed-tools format should be space-delimited tool names (e.g., 'Bash(git:*) Read Write')",
					Severity: "warning",
					Source:   cue.SourceAgentSkillsIO,
					Line:     FindFrontmatterFieldLine(contents, "allowed-tools"),
				})
				break
			}
		}
	}

	// Rule 053: License field validation
	suggestions = append(suggestions, validateSkillLicenseField(fmData, filePath, contents)...)

	// Rule 054: Compatibility field length
	warnings = append(warnings, validateSkillCompatibilityField(fmData, filePath, contents)...)

	// Rule 055: Metadata field structure
	suggestions = append(suggestions, validateSkillMetadataField(fmData, filePath, contents)...)

	return append(suggestions, warnings...)
}

// checkSkillAntiPatternsSection checks for Anti-Patterns section.
func checkSkillAntiPatternsSection(filePath, contents string) []cue.ValidationError {
	lowerContents := strings.ToLower(contents)
	hasAntiPatterns := strings.Contains(contents, "## Anti-Pattern") ||
		strings.Contains(contents, "## Anti-Patterns") ||
		strings.Contains(contents, "### Anti-Pattern") ||
		(strings.Contains(contents, "## Best Practices") && strings.Contains(lowerContents, "### don't")) ||
		strings.Contains(contents, "| Anti-Pattern")

	if !hasAntiPatterns {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Consider adding '## Anti-Patterns' section to document common mistakes.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		}}
	}
	return nil
}

// checkSkillExamplesSection checks for Examples section.
func checkSkillExamplesSection(filePath, contents string) []cue.ValidationError {
	hasExamples := strings.Contains(contents, "## Example") ||
		strings.Contains(contents, "## Examples") ||
		strings.Contains(contents, "## Expected Output") ||
		strings.Contains(contents, "## Usage") ||
		strings.Contains(contents, "### Example")

	if !hasExamples {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "Consider adding '## Examples' section to illustrate skill usage.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		}}
	}
	return nil
}

// validateSkillLicenseField validates the license field per agentskills.io spec.
func validateSkillLicenseField(fmData map[string]any, filePath, contents string) []cue.ValidationError {
	if license, ok := fmData["license"].(string); ok {
		if strings.TrimSpace(license) == "" {
			return []cue.ValidationError{{
				File:     filePath,
				Message:  "license field is empty - provide SPDX identifier (e.g., 'MIT', 'Apache-2.0') or license file reference",
				Severity: "suggestion",
				Source:   cue.SourceAgentSkillsIO,
				Line:     FindFrontmatterFieldLine(contents, "license"),
			}}
		}
	}
	return nil
}

// validateSkillCompatibilityField validates the compatibility field per agentskills.io spec.
func validateSkillCompatibilityField(fmData map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError
	if compat, ok := fmData["compatibility"].(string); ok {
		if len(compat) > 500 {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("compatibility field is %d chars (max 500 per agentskills.io spec)", len(compat)),
				Severity: "warning",
				Source:   cue.SourceAgentSkillsIO,
				Line:     FindFrontmatterFieldLine(contents, "compatibility"),
			})
		}
	}
	return errors
}

// validateSkillMetadataField validates the metadata field per agentskills.io spec.
func validateSkillMetadataField(fmData map[string]any, filePath, contents string) []cue.ValidationError {
	if metadata, ok := fmData["metadata"]; ok {
		metaMap, isMap := metadata.(map[string]any)
		if !isMap {
			return []cue.ValidationError{{
				File:     filePath,
				Message:  "metadata field should be key-value mapping (e.g., metadata:\\n  author: example-org\\n  version: \"1.0\")",
				Severity: "suggestion",
				Source:   cue.SourceAgentSkillsIO,
				Line:     FindFrontmatterFieldLine(contents, "metadata"),
			}}
		}

		// Validate value types are primitives
		for key, val := range metaMap {
			switch val.(type) {
			case string, int, int64, float64, bool:
				// Valid primitive types
			default:
				return []cue.ValidationError{{
					File:     filePath,
					Message:  fmt.Sprintf("metadata.%s should be primitive value (string, number, or boolean)", key),
					Severity: "suggestion",
					Source:   cue.SourceAgentSkillsIO,
					Line:     FindFrontmatterFieldLine(contents, "metadata"),
				}}
			}
		}
	}
	return nil
}

// validateSkillDescription checks description field for style and content issues.
func validateSkillDescription(description, filePath, contents string) []cue.ValidationError {
	var out []cue.ValidationError

	if xmlErr := DetectXMLTags(description, "Description", filePath, contents); xmlErr != nil {
		out = append(out, *xmlErr)
	}

	// P3: Third-person description check
	firstPersonStarts := []string{"I ", "I'm ", "I'll ", "I've ", "My ", "We ", "We're "}
	for _, fp := range firstPersonStarts {
		if strings.HasPrefix(description, fp) {
			out = append(out, cue.ValidationError{
				File:     filePath,
				Message:  "Skill description should use third person (e.g., 'Analyzes...' not 'I analyze...')",
				Severity: "suggestion",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "description"),
			})
			break
		}
	}

	if strings.HasPrefix(description, "You ") {
		out = append(out, cue.ValidationError{
			File:     filePath,
			Message:  "Skill description should describe what it does, not address the user",
			Severity: "suggestion",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "description"),
		})
	}

	if len(description) < 50 {
		out = append(out, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Description is only %d chars. Aim for 50+ to help with skill discovery.", len(description)),
			Severity: "suggestion",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "description"),
		})
	}

	hasTrigger := strings.Contains(strings.ToLower(description), "use when") ||
		strings.Contains(strings.ToLower(description), "use for") ||
		strings.Contains(strings.ToLower(description), "use proactively")
	if !hasTrigger && len(description) > 0 {
		out = append(out, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding trigger phrases like 'Use when...' or 'Use for...' to help skill discovery",
			Severity: "suggestion",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "description"),
		})
	}

	return out
}

// extractSkillName extracts the skill name from the first heading
func extractSkillName(content, filePath string) string {
	// Try to match "# Heading" pattern
	re := regexp.MustCompile(`^#\s+([^\n]+)`)
	matches := re.FindStringSubmatch(content)
	if len(matches) > 1 {
		name := strings.TrimSpace(matches[1])
		// Clean up common patterns
		name = strings.TrimPrefix(name, "Agent ")
		name = strings.TrimPrefix(name, "Command ")
		name = strings.TrimPrefix(name, "Skill ")
		name = strings.TrimSuffix(name, " Patterns")
		name = strings.ToLower(strings.ReplaceAll(name, " ", "-"))
		return name
	}

	// Fallback to directory name
	parts := strings.Split(filePath, "/")
	if len(parts) > 0 {
		dirName := parts[len(parts)-1]
		return dirName
	}

	return ""
}
