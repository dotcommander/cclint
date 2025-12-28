package cli

import (
	"fmt"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/frontend"
)

// LintContext runs linting on CLAUDE.md files using the generic linter.
func LintContext(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewContextLinter()), nil
}

// parseMarkdownSections parses markdown content into sections
func parseMarkdownSections(content string) []interface{} {
	var sections []interface{}

	// This is a simplified parser - in practice, would use a proper markdown parser
	lines := []string{}
	if fm, err := frontend.ParseYAMLFrontmatter(content); err == nil {
		lines = append(lines, strings.Split(fm.Body, "\n")...)
	} else {
		lines = append(lines, strings.Split(content, "\n")...)
	}

	currentSection := map[string]interface{}{}
	inSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect markdown headers (# or ##)
		if strings.HasPrefix(line, "## ") {
			// New h2 section found
			if inSection {
				sections = append(sections, currentSection)
			}
			currentSection = map[string]interface{}{
				"heading": strings.TrimPrefix(line, "## "),
				"content": "",
			}
			inSection = true
		} else if strings.HasPrefix(line, "# ") {
			// New h1 section found
			if inSection {
				sections = append(sections, currentSection)
			}
			currentSection = map[string]interface{}{
				"heading": strings.TrimPrefix(line, "# "),
				"content": "",
			}
			inSection = true
		} else if inSection && line != "" {
			if contentStr, ok := currentSection["content"].(string); ok {
				currentSection["content"] = contentStr + line + "\n"
			} else {
				currentSection["content"] = line + "\n"
			}
		}
	}

	// Add the last section
	if inSection {
		sections = append(sections, currentSection)
	}

	return sections
}

// validateContextSpecific implements context-specific validation rules
func validateContextSpecific(data map[string]interface{}, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check if sections are present
	if sections, ok := data["sections"].([]interface{}); ok && len(sections) > 0 {
		for i, section := range sections {
			if sectionMap, ok := section.(map[string]interface{}); ok {
				// Check heading
				if heading, exists := sectionMap["heading"]; !exists || heading == "" {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Section %d: missing heading", i),
						Severity: "warning",
					})
				}

				// Check content
				if content, exists := sectionMap["content"]; !exists || content == "" {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Section %d: missing content", i),
						Severity: "warning",
					})
				}
			}
		}
	} else {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "No sections found in CLAUDE.md",
			Severity: "suggestion",
		})
	}

	return errors
}