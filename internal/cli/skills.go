package cli

import (
	"fmt"
	"log"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
	"github.com/dotcommander/cclint/internal/project"
	"github.com/dotcommander/cclint/internal/scoring"
)

// LintSkills runs linting on skill files
func LintSkills(rootPath string, quiet bool, verbose bool) (*LintSummary, error) {
	summary := &LintSummary{}

	// Find project root first
	if rootPath == "" {
		var err error
		rootPath, err = project.FindProjectRoot(".")
		if err != nil {
			return nil, fmt.Errorf("error finding project root: %w", err)
		}
	}

	// Initialize components
	discoverer := discovery.NewFileDiscovery(rootPath, false)

	// Discover files
	files, err := discoverer.DiscoverFiles()
	if err != nil {
		return nil, fmt.Errorf("error discovering files: %w", err)
	}

	// Filter skill files
	var skillFiles []discovery.File
	for _, file := range files {
		if file.Type == discovery.FileTypeSkill {
			skillFiles = append(skillFiles, file)
		}
	}

	summary.TotalFiles = len(skillFiles)

	// Process each skill file
	for _, file := range skillFiles {
		result := LintResult{
			File:    file.RelPath,
			Type:    "skill",
			Success: true,
		}

		// Check that file is named SKILL.md
		if !strings.HasSuffix(file.RelPath, "/SKILL.md") && !strings.HasSuffix(file.RelPath, "\\SKILL.md") {
			result.Errors = append(result.Errors, cue.ValidationError{
				File:     file.RelPath,
				Message:  "Skill file must be named SKILL.md",
				Severity: "error",
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
			})
			result.Success = false
			summary.FailedFiles++
			summary.TotalErrors++
		}

		// Check for markdown frontmatter (optional but recommended)
		if !strings.HasPrefix(file.Contents, "---") {
			// Try to extract skill name from first heading
			skillName := extractSkillName(file.Contents, file.RelPath)
			suggestion := "Add YAML frontmatter with name and description"
			if skillName != "" {
				suggestion = fmt.Sprintf("Add frontmatter: ---\nname: %s\ndescription: Brief summary of what this skill does\n---", skillName)
			}
			result.Suggestions = append(result.Suggestions, cue.ValidationError{
				File:     file.RelPath,
				Message:  suggestion,
				Severity: "suggestion",
			})
			summary.TotalSuggestions++
		}

		// Best practice checks
		suggestions := validateSkillBestPractices(file.RelPath, file.Contents)
		result.Suggestions = append(result.Suggestions, suggestions...)
		summary.TotalSuggestions += len(suggestions)

		if len(result.Errors) == 0 {
			summary.SuccessfulFiles++
		}

		// Score skill quality
		scorer := scoring.NewSkillScorer()
		fm, _ := frontend.ParseYAMLFrontmatter(file.Contents)
		var fmData map[string]interface{}
		var bodyContent string
		if fm != nil {
			fmData = fm.Data
			bodyContent = fm.Body
		} else {
			fmData = make(map[string]interface{})
			bodyContent = file.Contents
		}
		score := scorer.Score(file.Contents, fmData, bodyContent)
		result.Quality = &score

		// Get improvement recommendations
		result.Improvements = GetSkillImprovements(file.Contents, fmData)

		summary.Results = append(summary.Results, result)

		if verbose {
			log.Printf("Processed %s: %d errors, %d suggestions", file.RelPath, len(result.Errors), len(result.Suggestions))
		}
	}

	return summary, nil
}

// validateSkillBestPractices checks opinionated best practices for skills
func validateSkillBestPractices(filePath string, contents string) []cue.ValidationError {
	var suggestions []cue.ValidationError
	lowerContents := strings.ToLower(contents)

	// === MARKETING CLAIMS DETECTOR ===
	// Only flag vague/unverifiable marketing claims in main content
	// Skip: descriptions, tables, conditionals, code blocks, anti-pattern docs
	marketingPattern := `\d+(\.\d+)?-?\d*x\s*(speedup|faster|improvement|boost)`
	matched, _ := regexp.MatchString(marketingPattern, lowerContents)
	if matched {
		skipThis := false
		lines := strings.Split(contents, "\n")
		inCodeBlock := false
		for _, line := range lines {
			// Track code blocks
			if strings.HasPrefix(strings.TrimSpace(line), "```") {
				inCodeBlock = !inCodeBlock
				continue
			}
			lowerLine := strings.ToLower(line)
			if regexp.MustCompile(marketingPattern).MatchString(lowerLine) {
				// Skip if in code block
				if inCodeBlock {
					skipThis = true
					break
				}
				// Skip if in table, conditional, comparison, or anti-pattern documentation
				skipContexts := []string{
					"when", "if", "can be", "up to", "potential", "|",
					"remove", "must not", "don't", "avoid", "#",
					"(", ")", // Parenthetical comparisons like "(10x faster)"
					"than", "vs", "compared to", // Legitimate comparisons
				}
				for _, ctx := range skipContexts {
					if strings.Contains(lowerLine, ctx) {
						skipThis = true
						break
					}
				}
				if skipThis {
					break
				}
			}
		}
		// Also skip if in frontmatter description
		if !skipThis && strings.Contains(contents[:min(500, len(contents))], "description:") {
			re := regexp.MustCompile(`(?i)description:.*` + marketingPattern)
			if re.MatchString(contents[:min(500, len(contents))]) {
				skipThis = true
			}
		}
		if !skipThis {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  "Skill contains marketing claim (e.g., '2x speedup'). Remove unverifiable performance metrics.",
				Severity: "warning",
			})
		}
	}

	// Check for Quick Reference table (semantic routing)
	if !strings.Contains(contents, "Quick Reference") && !strings.Contains(contents, "| User") && !strings.Contains(contents, "User Question") {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Skill lacks 'Quick Reference' table. Add semantic routing for discoverability (| User Question | Action |).",
			Severity: "suggestion",
		})
	}

	// Check for Anti-Patterns section (or equivalent)
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
		})
	}

	// Check skill size - recommend references for large skills
	lines := strings.Count(contents, "\n")
	if lines > 500 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Skill is %d lines. Consider moving heavy documentation to references/ subdirectory.", lines),
			Severity: "suggestion",
		})
	}

	// Check for Examples section (or equivalent)
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
		})
	}

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
