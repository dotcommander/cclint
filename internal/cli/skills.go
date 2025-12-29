package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// knownSkillFields lists valid frontmatter fields per Anthropic docs
// Source: https://docs.anthropic.com/en/docs/claude-code/skills
var knownSkillFields = map[string]bool{
	"name":          true, // Required: skill identifier
	"description":   true, // Required: what the skill does (critical for discovery)
	"allowed-tools": true, // Optional: tool access permissions
	"model":         true, // Optional: model to use when skill is active
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
func validateSkillBestPractices(filePath string, contents string, fmData map[string]interface{}) []cue.ValidationError {
	var suggestions []cue.ValidationError
	var warnings []cue.ValidationError
	lowerContents := strings.ToLower(contents)

	// XML tag detection in text fields - FROM ANTHROPIC DOCS
	if description, ok := fmData["description"].(string); ok {
		if xmlErr := DetectXMLTags(description, "Description", filePath, contents); xmlErr != nil {
			suggestions = append(suggestions, *xmlErr)
		}

		// P3: Third-person description check - FROM ANTHROPIC DOCS
		firstPersonStarts := []string{"I ", "I'm ", "I'll ", "I've ", "My ", "We ", "We're "}
		for _, fp := range firstPersonStarts {
			if strings.HasPrefix(description, fp) {
				suggestions = append(suggestions, cue.ValidationError{
					File:     filePath,
					Message:  "Skill description should use third person (e.g., 'Analyzes...' not 'I analyze...')",
					Severity: "suggestion",
					Source:   cue.SourceAnthropicDocs,
					Line:     FindFrontmatterFieldLine(contents, "description"),
				})
				break
			}
		}

		// Also check for "You " which addresses the user incorrectly
		if strings.HasPrefix(description, "You ") {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  "Skill description should describe what it does, not address the user",
				Severity: "suggestion",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "description"),
			})
		}

		// P3: Description specificity check - FROM ANTHROPIC DOCS
		if len(description) < 50 {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Description is only %d chars. Aim for 50+ to help with skill discovery.", len(description)),
				Severity: "suggestion",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "description"),
			})
		}

		// Missing trigger phrases
		hasTrigger := strings.Contains(strings.ToLower(description), "use when") ||
			strings.Contains(strings.ToLower(description), "use for") ||
			strings.Contains(strings.ToLower(description), "use proactively")
		if !hasTrigger && len(description) > 0 {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  "Consider adding trigger phrases like 'Use when...' or 'Use for...' to help skill discovery",
				Severity: "suggestion",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "description"),
			})
		}
	}

	// P3: Semver validation for version field - OUR OBSERVATION
	if version, ok := fmData["version"].(string); ok && version != "" {
		semverPattern := regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`)
		if !semverPattern.MatchString(version) {
			warnings = append(warnings, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Version '%s' should follow semver format (e.g., '1.0.0')", version),
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
				Line:     FindFrontmatterFieldLine(contents, "version"),
			})
		}
	}

	// Check for Anti-Patterns section (or equivalent) - OUR OBSERVATION
	hasAntiPatterns := strings.Contains(contents, "## Anti-Pattern") ||
		strings.Contains(contents, "## Anti-Patterns") ||
		strings.Contains(contents, "### Anti-Pattern") ||
		(strings.Contains(contents, "## Best Practices") && strings.Contains(lowerContents, "### don't")) ||
		strings.Contains(contents, "| Anti-Pattern")
	if !hasAntiPatterns {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding '## Anti-Patterns' section to document common mistakes.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Check skill size - recommend references for large skills (±10% tolerance: 500 base) - OUR OBSERVATION
	if sizeErr := CheckSizeLimit(contents, 500, 0.10, "skill", filePath); sizeErr != nil {
		suggestions = append(suggestions, *sizeErr)
	}

	// Check for Examples section (or equivalent) - OUR OBSERVATION
	hasExamples := strings.Contains(contents, "## Example") ||
		strings.Contains(contents, "## Examples") ||
		strings.Contains(contents, "## Expected Output") ||
		strings.Contains(contents, "## Usage") ||
		strings.Contains(contents, "### Example")
	if !hasExamples {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Consider adding '## Examples' section to illustrate skill usage.",
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Validate tool field naming (skills use 'allowed-tools:', not 'tools:')
	suggestions = append(suggestions, ValidateToolFieldName(fmData, filePath, contents, "skill")...)

	// Merge warnings into suggestions for return
	suggestions = append(suggestions, warnings...)

	return suggestions
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
