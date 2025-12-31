package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
)

// LintRules runs linting on .claude/rules/*.md files using the generic linter.
func LintRules(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewRuleLinter()), nil
}

// RuleLinter implements ComponentLinter for .claude/rules/*.md files.
type RuleLinter struct {
	BaseLinter
}

// Compile-time interface compliance check
var _ ComponentLinter = (*RuleLinter)(nil)
var _ PreValidator = (*RuleLinter)(nil)
var _ BestPracticeValidator = (*RuleLinter)(nil)

// NewRuleLinter creates a new RuleLinter.
func NewRuleLinter() *RuleLinter {
	return &RuleLinter{}
}

func (l *RuleLinter) Type() string {
	return "rule"
}

func (l *RuleLinter) FileType() discovery.FileType {
	return discovery.FileTypeRule
}

func (l *RuleLinter) PreValidate(filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".md") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Rule files must have .md extension",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	// Check empty content
	if strings.TrimSpace(contents) == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Rule file is empty",
			Severity: "error",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Check for symlink and validate target exists
	if info, err := os.Lstat(filePath); err == nil {
		if info.Mode()&os.ModeSymlink != 0 {
			target, err := filepath.EvalSymlinks(filePath)
			if err != nil {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Symlink target does not exist or is inaccessible: %v", err),
					Severity: "warning",
					Source:   cue.SourceCClintObserve,
				})
			} else if _, err := os.Stat(target); err != nil {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Symlink target not found: %s", target),
					Severity: "warning",
					Source:   cue.SourceCClintObserve,
				})
			}
		}
	}

	return errors
}

func (l *RuleLinter) ParseContent(contents string) (map[string]interface{}, string, error) {
	// Rules have optional frontmatter with 'paths' field
	fm, err := frontend.ParseYAMLFrontmatter(contents)

	var paths interface{}
	if err == nil && fm != nil && fm.Data != nil {
		paths = fm.Data["paths"]
	}

	data := map[string]interface{}{
		"paths": paths,
	}

	body := contents
	if fm != nil {
		body = fm.Body
	}

	return data, body, nil
}

func (l *RuleLinter) ValidateCUE(validator *cue.Validator, data map[string]interface{}) ([]cue.ValidationError, error) {
	// Rules don't have a CUE schema yet - frontmatter is optional
	return nil, nil
}

func (l *RuleLinter) ValidateSpecific(data map[string]interface{}, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Validate paths: field if present
	if paths, ok := data["paths"]; ok && paths != nil {
		pathErrors := validatePathsGlob(paths, filePath, contents)
		errors = append(errors, pathErrors...)
	}

	return errors
}

func (l *RuleLinter) ValidateBestPractices(filePath, contents string, data map[string]interface{}) []cue.ValidationError {
	var suggestions []cue.ValidationError

	// Check for @imports and validate they exist
	importErrors := validateImports(contents, filePath)
	suggestions = append(suggestions, importErrors...)

	return suggestions
}

// validatePathsGlob validates the paths: frontmatter field
func validatePathsGlob(paths interface{}, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	pathStr, ok := paths.(string)
	if !ok {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "paths: field must be a string",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "paths"),
		})
		return errors
	}

	// Split by comma, but only commas outside of braces
	// e.g., "**/*.{ts,tsx}, src/**/*.js" -> ["**/*.{ts,tsx}", "src/**/*.js"]
	patterns := splitPathPatterns(pathStr)
	for _, pattern := range patterns {
		pattern = strings.TrimSpace(pattern)
		if pattern == "" {
			continue
		}

		// Validate glob syntax
		if err := validateGlobPattern(pattern); err != nil {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Invalid glob pattern %q: %v", pattern, err),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "paths"),
			})
		}
	}

	return errors
}

// splitPathPatterns splits a paths string by commas, respecting brace expansion.
// Commas inside braces are NOT treated as separators.
// e.g., "**/*.{ts,tsx}, src/**/*.js" -> ["**/*.{ts,tsx}", "src/**/*.js"]
func splitPathPatterns(s string) []string {
	var patterns []string
	var current strings.Builder
	braceDepth := 0

	for _, ch := range s {
		switch ch {
		case '{':
			braceDepth++
			current.WriteRune(ch)
		case '}':
			braceDepth--
			current.WriteRune(ch)
		case ',':
			if braceDepth == 0 {
				// Separator comma - start new pattern
				patterns = append(patterns, current.String())
				current.Reset()
			} else {
				// Comma inside braces - part of pattern
				current.WriteRune(ch)
			}
		default:
			current.WriteRune(ch)
		}
	}

	// Add final pattern
	if current.Len() > 0 {
		patterns = append(patterns, current.String())
	}

	return patterns
}

// validateGlobPattern validates a glob pattern for syntax errors
func validateGlobPattern(pattern string) error {
	// Check for balanced braces
	if err := validateBalancedBraces(pattern); err != nil {
		return err
	}

	// Try to compile the pattern using doublestar
	_, err := doublestar.Match(pattern, "test.txt")
	if err != nil {
		return fmt.Errorf("invalid glob syntax: %v", err)
	}

	return nil
}

// validateBalancedBraces checks that braces are balanced in a pattern
func validateBalancedBraces(pattern string) error {
	depth := 0
	for i, ch := range pattern {
		switch ch {
		case '{':
			depth++
		case '}':
			depth--
			if depth < 0 {
				return fmt.Errorf("unmatched closing brace at position %d", i)
			}
		}
	}
	if depth > 0 {
		return fmt.Errorf("unclosed brace (missing %d closing braces)", depth)
	}
	return nil
}

// validateImports checks @import references in content
func validateImports(contents, filePath string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Pattern to match @imports (not in code blocks)
	// Per docs: @path/to/import syntax, not evaluated in code spans/blocks
	importPattern := regexp.MustCompile(`(?m)^[^` + "`" + `]*@([~./][^\s]+)`)

	// Find imports not in code blocks
	lines := strings.Split(contents, "\n")
	inCodeBlock := false

	for lineNum, line := range lines {
		// Track code block state
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		if inCodeBlock {
			continue
		}

		// Find @imports in this line
		matches := importPattern.FindAllStringSubmatch(line, -1)
		for _, match := range matches {
			if len(match) >= 2 {
				importPath := match[1]

				// Expand ~ to home directory
				if strings.HasPrefix(importPath, "~") {
					home, err := os.UserHomeDir()
					if err == nil {
						importPath = filepath.Join(home, importPath[1:])
					}
				}

				// Make relative paths absolute based on file location
				if !filepath.IsAbs(importPath) {
					dir := filepath.Dir(filePath)
					importPath = filepath.Join(dir, importPath)
				}

				// Check if import target exists
				if _, err := os.Stat(importPath); os.IsNotExist(err) {
					errors = append(errors, cue.ValidationError{
						File:     filePath,
						Message:  fmt.Sprintf("Import target does not exist: @%s", match[1]),
						Severity: "warning",
						Source:   cue.SourceCClintObserve,
						Line:     lineNum + 1,
					})
				}
			}
		}
	}

	// Check for circular imports (simplified - just detect self-reference)
	// Full circular detection would require building an import graph
	absPath, _ := filepath.Abs(filePath)
	for lineNum, line := range lines {
		if strings.Contains(line, "@") && !inCodeBlock {
			if strings.Contains(line, filepath.Base(filePath)) {
				errors = append(errors, cue.ValidationError{
					File:     filePath,
					Message:  "Possible self-import detected",
					Severity: "warning",
					Source:   cue.SourceCClintObserve,
					Line:     lineNum + 1,
				})
			}
			_ = absPath // suppress unused warning
		}
	}

	return errors
}
