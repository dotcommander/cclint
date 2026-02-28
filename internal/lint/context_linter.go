package lint

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/textutil"
)

// ContextLinter implements ComponentLinter for CLAUDE.md context files.
// It implements only the core ComponentLinter interface - no optional capabilities.
// Context files don't need scoring, improvements, or cross-file validation.
type ContextLinter struct {
	BaseLinter
}

// Compile-time interface compliance check
var _ ComponentLinter = (*ContextLinter)(nil)

// NewContextLinter creates a new ContextLinter.
func NewContextLinter() *ContextLinter {
	return &ContextLinter{}
}

func (l *ContextLinter) Type() string {
	return "context"
}

func (l *ContextLinter) FileType() discovery.FileType {
	return discovery.FileTypeContext
}

func (l *ContextLinter) ParseContent(contents string) (map[string]any, string, error) {
	// Context files have optional frontmatter
	fm, err := textutil.ParseYAMLFrontmatter(contents)

	var title, description any
	if err == nil && fm != nil && fm.Data != nil {
		title = fm.Data["title"]
		description = fm.Data["description"]
	}

	// Build the data structure expected by CUE validator
	data := map[string]any{
		"title":       title,
		"description": description,
		"sections":    parseMarkdownSections(contents),
	}

	body := contents
	if fm != nil {
		body = fm.Body
	}

	return data, body, nil
}

func (l *ContextLinter) ValidateCUE(validator *cue.Validator, data map[string]any) ([]cue.ValidationError, error) {
	return validator.ValidateClaudeMD(data)
}

func (l *ContextLinter) ValidateSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	return validateContextSpecific(data, filePath, contents)
}

// parseMarkdownSections parses markdown content into sections.
// Each section is a map with "heading" and "content" keys.
func parseMarkdownSections(content string) []any {
	var sections []any

	lines := []string{}
	if fm, err := textutil.ParseYAMLFrontmatter(content); err == nil {
		lines = append(lines, strings.Split(fm.Body, "\n")...)
	} else {
		lines = append(lines, strings.Split(content, "\n")...)
	}

	currentSection := map[string]any{}
	inSection := false

	for _, line := range lines {
		line = strings.TrimSpace(line)

		switch {
		case strings.HasPrefix(line, "## "):
			// New h2 section found
			if inSection {
				sections = append(sections, currentSection)
			}
			currentSection = map[string]any{
				"heading": strings.TrimPrefix(line, "## "),
				"content": "",
			}
			inSection = true
		case strings.HasPrefix(line, "# "):
			// New h1 section found
			if inSection {
				sections = append(sections, currentSection)
			}
			currentSection = map[string]any{
				"heading": strings.TrimPrefix(line, "# "),
				"content": "",
			}
			inSection = true
		case inSection && line != "":
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

// binaryExtensions lists file extensions that should not be included via @include.
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

// includePattern matches @include directives in CLAUDE.md files.
// Supports: @include path/to/file or @include ./relative/path
var includePattern = regexp.MustCompile(`(?m)@include\s+([^\s]+)`)

// validateContextSpecific implements context-specific validation rules.
func validateContextSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check if sections are present
	sections, ok := data["sections"].([]any)
	if !ok || len(sections) == 0 {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "No sections found in CLAUDE.md",
			Severity: cue.SeveritySuggestion,
		})
	} else {
		errors = append(errors, validateContextSections(sections, filePath)...)
	}

	// Check for binary file includes (Claude Code 2.1.2+ auto-skips these, but warn users)
	errors = append(errors, checkBinaryIncludes(contents, filePath)...)

	return errors
}

// validateContextSections validates individual sections have headings and content.
func validateContextSections(sections []any, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError
	for i, section := range sections {
		sectionMap, ok := section.(map[string]any)
		if !ok {
			continue
		}
		if heading, exists := sectionMap["heading"]; !exists || heading == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Section %d: missing heading", i),
				Severity: cue.SeverityWarning,
			})
		}
		if content, exists := sectionMap["content"]; !exists || content == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Section %d: missing content", i),
				Severity: cue.SeverityWarning,
			})
		}
	}
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
				Severity: cue.SeverityWarning,
				Source:   "binary-include",
			})
		}
	}

	return errors
}
