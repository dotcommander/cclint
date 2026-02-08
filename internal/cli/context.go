package cli

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/frontend"
)

// binaryExtensions lists file extensions that should not be included via @include
var binaryExtensions = map[string]bool{
	".png": true, ".jpg": true, ".jpeg": true, ".gif": true, ".bmp": true,
	".ico": true, ".webp": true, ".svg": true, ".tiff": true,
	".pdf": true, ".doc": true, ".docx": true, ".xls": true, ".xlsx": true,
	".ppt": true, ".pptx": true, ".odt": true, ".ods": true,
	".zip": true, ".tar": true, ".gz": true, ".rar": true, ".7z": true,
	".exe": true, ".dll": true, ".so": true, ".dylib": true, ".bin": true,
	".mp3": true, ".mp4": true, ".wav": true, ".avi": true, ".mov": true,
	".ttf": true, ".otf": true, ".woff": true, ".woff2": true,
}

// includePattern matches @include directives in CLAUDE.md files
// Supports: @include path/to/file or @include ./relative/path
var includePattern = regexp.MustCompile(`(?m)@include\s+([^\s]+)`)

// LintContext runs linting on CLAUDE.md files using the generic linter.
func LintContext(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewContextLinter()), nil
}

// parseMarkdownSections parses markdown content into sections
func parseMarkdownSections(content string) []any {
	var sections []any

	// This is a simplified parser - in practice, would use a proper markdown parser
	lines := []string{}
	if fm, err := frontend.ParseYAMLFrontmatter(content); err == nil {
		lines = append(lines, strings.Split(fm.Body, "\n")...)
	} else {
		lines = append(lines, strings.Split(content, "\n")...)
	}

	currentSection := map[string]any{}
	inSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Detect markdown headers (# or ##)
		if strings.HasPrefix(line, "## ") {
			// New h2 section found
			if inSection {
				sections = append(sections, currentSection)
			}
			currentSection = map[string]any{
				"heading": strings.TrimPrefix(line, "## "),
				"content": "",
			}
			inSection = true
		} else if strings.HasPrefix(line, "# ") {
			// New h1 section found
			if inSection {
				sections = append(sections, currentSection)
			}
			currentSection = map[string]any{
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
func validateContextSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check if sections are present
	if sections, ok := data["sections"].([]any); ok && len(sections) > 0 {
		for i, section := range sections {
			if sectionMap, ok := section.(map[string]any); ok {
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

	// Check for binary file includes (Claude Code 2.1.2+ auto-skips these, but warn users)
	errors = append(errors, checkBinaryIncludes(contents, filePath)...)

	return errors
}

// checkBinaryIncludes detects @include directives referencing binary files.
// Claude Code 2.1.2 fixed a bug where binary files were accidentally included in memory.
// This check warns users about ineffective includes that will be silently skipped.
func checkBinaryIncludes(contents, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	matches := includePattern.FindAllStringSubmatch(contents, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		includePath := match[1]
		ext := strings.ToLower(filepath.Ext(includePath))

		if binaryExtensions[ext] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("@include references binary file '%s' which will be skipped by Claude Code", includePath),
				Severity: "warning",
				Source:   "binary-include",
			})
		}
	}

	return errors
}