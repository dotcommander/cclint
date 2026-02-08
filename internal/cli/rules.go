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

// Compile-time interface compliance checks
var (
	_ ComponentLinter    = (*RuleLinter)(nil)
	_ PreValidator       = (*RuleLinter)(nil)
	_ BestPracticeValidator = (*RuleLinter)(nil)
	_ BatchPostProcessor = (*RuleLinter)(nil)
)

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
	errors = append(errors, validateSymlinkTarget(filePath)...)

	return errors
}

// validateSymlinkTarget checks if a file is a symlink and validates its target exists.
func validateSymlinkTarget(filePath string) []cue.ValidationError {
	info, err := os.Lstat(filePath)
	if err != nil {
		return nil
	}
	if info.Mode()&os.ModeSymlink == 0 {
		return nil
	}
	target, err := filepath.EvalSymlinks(filePath)
	if err != nil {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Symlink target does not exist or is inaccessible: %v", err),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
		}}
	}
	if _, err := os.Stat(target); err != nil {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  fmt.Sprintf("Symlink target not found: %s", target),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
		}}
	}
	return nil
}

func (l *RuleLinter) ParseContent(contents string) (map[string]any, string, error) {
	// Rules have optional frontmatter with 'paths' field
	fm, err := frontend.ParseYAMLFrontmatter(contents)

	var paths any
	if err == nil && fm != nil && fm.Data != nil {
		paths = fm.Data["paths"]
	}

	data := map[string]any{
		"paths": paths,
	}

	body := contents
	if fm != nil {
		body = fm.Body
	}

	return data, body, nil
}

func (l *RuleLinter) ValidateCUE(validator *cue.Validator, data map[string]any) ([]cue.ValidationError, error) {
	// Rules don't have a CUE schema yet - frontmatter is optional
	return nil, nil
}

func (l *RuleLinter) ValidateSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Validate paths: field if present
	if paths, ok := data["paths"]; ok && paths != nil {
		pathErrors := validatePathsGlob(paths, filePath, contents)
		errors = append(errors, pathErrors...)
	}

	return errors
}

func (l *RuleLinter) ValidateBestPractices(filePath, contents string, data map[string]any) []cue.ValidationError {
	var suggestions []cue.ValidationError

	// Check for @imports and validate they exist
	importErrors := validateImports(contents, filePath)
	suggestions = append(suggestions, importErrors...)

	return suggestions
}

// validatePathsGlob validates the paths: frontmatter field
func validatePathsGlob(paths any, filePath, contents string) []cue.ValidationError {
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
			if len(match) < 2 {
				continue
			}
			resolved := resolveRuleImportPath(match[1], filePath)
			if _, err := os.Stat(resolved); os.IsNotExist(err) {
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

	// Circular import detection is handled at batch level by PostProcessBatch,
	// which builds a full import graph and runs DFS cycle detection.

	return errors
}

// resolveRuleImportPath resolves an @import path to an absolute filesystem path.
// Handles ~ expansion and relative path resolution.
func resolveRuleImportPath(importPath, filePath string) string {
	if strings.HasPrefix(importPath, "~") {
		home, err := os.UserHomeDir()
		if err == nil {
			importPath = filepath.Join(home, importPath[1:])
		}
	}
	if !filepath.IsAbs(importPath) {
		dir := filepath.Dir(filePath)
		importPath = filepath.Join(dir, importPath)
	}
	return importPath
}

// PostProcessBatch implements BatchPostProcessor for import cycle detection.
// It builds an import graph across all rule files and reports circular @import chains.
func (l *RuleLinter) PostProcessBatch(ctx *LinterContext, summary *LintSummary) {
	if ctx.NoCycleCheck {
		return
	}

	// Build file map: absolute path -> contents
	fileMap := make(map[string]string)
	for _, file := range ctx.FilterFilesByType(discovery.FileTypeRule) {
		absPath := file.Path
		if absPath == "" {
			absPath = filepath.Join(ctx.RootPath, file.RelPath)
		}
		fileMap[absPath] = file.Contents
	}

	if len(fileMap) == 0 {
		return
	}

	cycleErrors := DetectImportCycles(fileMap)
	for _, cycleErr := range cycleErrors {
		// Find the matching result and attach the error
		for i, result := range summary.Results {
			absResult := result.File
			if !filepath.IsAbs(absResult) {
				absResult = filepath.Join(ctx.RootPath, absResult)
			}
			if absResult == cycleErr.File {
				summary.Results[i].Errors = append(summary.Results[i].Errors, cycleErr)
				summary.TotalErrors++
				if summary.Results[i].Success {
					summary.Results[i].Success = false
					summary.SuccessfulFiles--
					summary.FailedFiles++
				}
				break
			}
		}
	}
}
