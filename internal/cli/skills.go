package cli

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
	"github.com/dotcommander/cclint/internal/scoring"
)

// LintSkills runs linting on skill files
func LintSkills(rootPath string, quiet bool, verbose bool) (*LintSummary, error) {
	// Initialize shared context
	ctx, err := NewLinterContext(rootPath, quiet, verbose)
	if err != nil {
		return nil, err
	}

	// Filter skill files
	skillFiles := ctx.FilterFilesByType(discovery.FileTypeSkill)
	summary := ctx.NewSummary(len(skillFiles))

	// Process each skill file
	for _, file := range skillFiles {
		result := LintResult{
			File:    file.RelPath,
			Type:    "skill",
			Success: true,
		}

		// Check that file is named SKILL.md - FROM ANTHROPIC DOCS
		if !strings.HasSuffix(file.RelPath, "/SKILL.md") && !strings.HasSuffix(file.RelPath, "\\SKILL.md") {
			result.Errors = append(result.Errors, cue.ValidationError{
				File:     file.RelPath,
				Message:  "Skill file must be named SKILL.md",
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			result.Success = false
			summary.FailedFiles++
			summary.TotalErrors++
		}

		// Check that file has content
		if len(strings.TrimSpace(file.Contents)) == 0 {
			result.Errors = append(result.Errors, cue.ValidationError{
				File:     file.RelPath,
				Message:  "Skill file is empty",
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
			})
			result.Success = false
			summary.FailedFiles++
			summary.TotalErrors++
		}

		// Parse frontmatter if present
		fm, _ := frontend.ParseYAMLFrontmatter(file.Contents)
		var fmData map[string]interface{}
		if fm != nil {
			fmData = fm.Data

			// Reserved word check for skill names - FROM ANTHROPIC DOCS
			if name, ok := fmData["name"].(string); ok {
				reservedWords := map[string]bool{"anthropic": true, "claude": true}
				if reservedWords[strings.ToLower(name)] {
					result.Errors = append(result.Errors, cue.ValidationError{
						File:     file.RelPath,
						Message:  fmt.Sprintf("Name '%s' is a reserved word and cannot be used", name),
						Severity: "error",
						Source:   cue.SourceAnthropicDocs,
						Line:     FindFrontmatterFieldLine(file.Contents, "name"),
					})
					result.Success = false
					summary.FailedFiles++
					summary.TotalErrors++
				}
			}
		}

		// Check for markdown frontmatter - Anthropic docs: "You can add frontmatter", description "critical for discovery"
		if !strings.HasPrefix(file.Contents, "---") {
			// Try to extract skill name from first heading
			skillName := extractSkillName(file.Contents, file.RelPath)
			suggestion := "Add YAML frontmatter with name and description (description is critical for skill discovery)"
			if skillName != "" {
				suggestion = fmt.Sprintf("Add frontmatter: ---\nname: %s\ndescription: Brief summary of what this skill does\n--- (description critical for discovery)", skillName)
			}
			result.Suggestions = append(result.Suggestions, cue.ValidationError{
				File:     file.RelPath,
				Message:  suggestion,
				Severity: "suggestion",
				Source:   cue.SourceAnthropicDocs,
			})
			summary.TotalSuggestions++
		}

		// Best practice checks (pass fmData if available)
		if fmData == nil {
			fmData = make(map[string]interface{})
		}
		suggestions := validateSkillBestPractices(file.RelPath, file.Contents, fmData)
		result.Suggestions = append(result.Suggestions, suggestions...)
		summary.TotalSuggestions += len(suggestions)

		// Cross-file validation (missing agents)
		crossErrors := ctx.CrossValidator.ValidateSkill(file.RelPath, file.Contents)
		result.Errors = append(result.Errors, crossErrors...)
		summary.TotalErrors += len(crossErrors)

		// Secrets detection
		secretWarnings := detectSecrets(file.Contents, file.RelPath)
		result.Warnings = append(result.Warnings, secretWarnings...)
		summary.TotalWarnings += len(secretWarnings)

		if len(result.Errors) == 0 {
			summary.SuccessfulFiles++
		}

		// Score skill quality
		scorer := scoring.NewSkillScorer()
		var bodyContent string
		if fm != nil {
			bodyContent = fm.Body
		} else {
			if fmData == nil {
				fmData = make(map[string]interface{})
			}
			bodyContent = file.Contents
		}
		score := scorer.Score(file.Contents, fmData, bodyContent)
		result.Quality = &score

		// Get improvement recommendations
		result.Improvements = GetSkillImprovements(file.Contents, fmData)

		summary.Results = append(summary.Results, result)
		ctx.LogProcessedWithSuggestions(file.RelPath, len(result.Errors), len(result.Suggestions))
	}

	// Find orphaned skills (no incoming references)
	orphanedSkills := ctx.CrossValidator.FindOrphanedSkills()
	for _, orphan := range orphanedSkills {
		// Add as suggestions to the summary
		summary.TotalSuggestions++
		// Also add to individual file results for display
		for i, result := range summary.Results {
			if result.File == orphan.File {
				summary.Results[i].Suggestions = append(summary.Results[i].Suggestions, orphan)
				break
			}
		}
	}

	return summary, nil
}

// validateSkillBestPractices checks opinionated best practices for skills
func validateSkillBestPractices(filePath string, contents string, fmData map[string]interface{}) []cue.ValidationError {
	var suggestions []cue.ValidationError
	var warnings []cue.ValidationError
	lowerContents := strings.ToLower(contents)

	// XML tag detection in text fields - FROM ANTHROPIC DOCS
	xmlTagPattern := regexp.MustCompile(`<[a-zA-Z][^>]*>`)
	if description, ok := fmData["description"].(string); ok {
		if xmlTagPattern.MatchString(description) {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  "Description contains XML-like tags which are not allowed",
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "description"),
			})
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

	// Check skill size - recommend references for large skills (±10% tolerance: 500 base + 50 = 550) - OUR OBSERVATION
	lines := strings.Count(contents, "\n")
	if lines > 550 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Skill is %d lines. Best practice: keep skills under ~550 lines (500±10%%) - move heavy docs to references/ subdirectory.", lines),
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
		})
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
