// Package cli provides generic linting infrastructure for cclint.
//
// This file implements the Generic Single-File Linter and Generic Batch Linter
// patterns, extracting common logic from the six component-specific linters.
//
// # Design Principles
//
//   - Interface-driven: ComponentLinter interface for type-specific logic
//   - DRY: Common linting logic extracted to lintComponent()
//   - Extensible: New component types only need to implement ComponentLinter
package cli

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
	"github.com/dotcommander/cclint/internal/scoring"
)

// ComponentLinter defines the interface for type-specific linting logic.
// Each component type (agent, command, skill, etc.) implements this interface.
type ComponentLinter interface {
	// Type returns the component type name (e.g., "agent", "command")
	Type() string

	// FileType returns the discovery.FileType for filtering
	FileType() discovery.FileType

	// ParseContent parses the file content and returns frontmatter data and body.
	// For JSON files, body may be empty.
	ParseContent(contents string) (data map[string]interface{}, body string, err error)

	// ValidateCUE runs CUE schema validation if applicable.
	// Returns nil if CUE validation is not used for this type.
	ValidateCUE(validator *cue.Validator, data map[string]interface{}) ([]cue.ValidationError, error)

	// ValidateSpecific runs component-specific validation rules.
	ValidateSpecific(data map[string]interface{}, filePath, contents string) []cue.ValidationError

	// ValidateBestPractices runs best practice checks.
	// May return nil if no best practices are defined.
	ValidateBestPractices(filePath, contents string, data map[string]interface{}) []cue.ValidationError

	// ValidateCrossFile runs cross-file validation if applicable.
	// May return nil if cross-file validation is not used.
	ValidateCrossFile(crossValidator *CrossFileValidator, filePath, contents string, data map[string]interface{}) []cue.ValidationError

	// Score returns the quality score for the component.
	// May return nil if scoring is not supported.
	Score(contents string, data map[string]interface{}, body string) *scoring.QualityScore

	// GetImprovements returns improvement recommendations.
	// May return nil if improvements are not supported.
	GetImprovements(contents string, data map[string]interface{}) []ImprovementRecommendation

	// PreValidate runs any pre-validation checks (e.g., filename, empty content).
	// Returns errors that should abort further validation.
	PreValidate(filePath, contents string) []cue.ValidationError

	// PostProcess allows type-specific post-processing of results.
	// Used for things like separating suggestions from errors.
	PostProcess(result *LintResult)
}

// lintComponent is the generic single-file linting function.
// It orchestrates the linting pipeline using a ComponentLinter.
func lintComponent(ctx *SingleFileLinterContext, linter ComponentLinter) LintResult {
	result := LintResult{
		File:    ctx.File.RelPath,
		Type:    linter.Type(),
		Success: true,
	}

	// Pre-validation checks (filename, empty content, etc.)
	preErrors := linter.PreValidate(ctx.File.RelPath, ctx.File.Contents)
	if len(preErrors) > 0 {
		result.Errors = append(result.Errors, preErrors...)
		// Check if any are fatal (should abort further validation)
		for _, e := range preErrors {
			if e.Severity == "error" && strings.Contains(e.Message, "is empty") {
				result.Success = len(result.Errors) == 0
				return result
			}
		}
	}

	// Parse content
	data, body, err := linter.ParseContent(ctx.File.Contents)
	if err != nil {
		result.Errors = append(result.Errors, cue.ValidationError{
			File:     ctx.File.RelPath,
			Message:  err.Error(),
			Severity: "error",
		})
		result.Success = false
		return result
	}

	// CUE validation
	if cueErrors, err := linter.ValidateCUE(ctx.Validator, data); err != nil {
		result.Errors = append(result.Errors, cue.ValidationError{
			File:     ctx.File.RelPath,
			Message:  fmt.Sprintf("Validation error: %v", err),
			Severity: "error",
		})
	} else if cueErrors != nil {
		result.Errors = append(result.Errors, cueErrors...)
	}

	// Component-specific validation
	specificErrors := linter.ValidateSpecific(data, ctx.File.RelPath, ctx.File.Contents)
	for _, issue := range specificErrors {
		if issue.Severity == "suggestion" {
			result.Suggestions = append(result.Suggestions, issue)
		} else if issue.Severity == "warning" {
			result.Warnings = append(result.Warnings, issue)
		} else {
			result.Errors = append(result.Errors, issue)
		}
	}

	// Best practice checks
	if bpIssues := linter.ValidateBestPractices(ctx.File.RelPath, ctx.File.Contents, data); bpIssues != nil {
		for _, issue := range bpIssues {
			if issue.Severity == "suggestion" {
				result.Suggestions = append(result.Suggestions, issue)
			} else if issue.Severity == "warning" {
				result.Warnings = append(result.Warnings, issue)
			} else {
				result.Errors = append(result.Errors, issue)
			}
		}
	}

	// Cross-file validation
	crossValidator := ctx.EnsureCrossFileValidator()
	if crossValidator != nil {
		if crossErrors := linter.ValidateCrossFile(crossValidator, ctx.File.RelPath, ctx.File.Contents, data); crossErrors != nil {
			result.Errors = append(result.Errors, crossErrors...)
		}
	} else if ctx.Verbose {
		result.Suggestions = append(result.Suggestions, cue.ValidationError{
			File:     ctx.File.RelPath,
			Message:  "Cross-file validation skipped (could not discover project files)",
			Severity: "info",
			Source:   cue.SourceCClintObserve,
		})
	}

	// Secrets detection (common to all types)
	secretWarnings := detectSecrets(ctx.File.Contents, ctx.File.RelPath)
	result.Warnings = append(result.Warnings, secretWarnings...)

	// Quality scoring
	if score := linter.Score(ctx.File.Contents, data, body); score != nil {
		result.Quality = score
	}

	// Improvement recommendations
	if improvements := linter.GetImprovements(ctx.File.Contents, data); improvements != nil {
		result.Improvements = improvements
	}

	// Allow type-specific post-processing
	linter.PostProcess(&result)

	result.Success = len(result.Errors) == 0
	return result
}

// BatchPostProcessor is an optional interface for post-processing batch results.
// Component linters can implement this for operations like cycle detection.
type BatchPostProcessor interface {
	PostProcessBatch(ctx *LinterContext, summary *LintSummary)
}

// lintBatch is the generic batch linting function.
// It orchestrates batch linting using a ComponentLinter.
func lintBatch(ctx *LinterContext, linter ComponentLinter) *LintSummary {
	files := ctx.FilterFilesByType(linter.FileType())
	summary := ctx.NewSummary(len(files))

	for _, file := range files {
		result := lintBatchFile(ctx, file, linter)

		// Update summary counts
		if result.Success {
			summary.SuccessfulFiles++
		} else {
			summary.FailedFiles++
		}
		summary.TotalErrors += len(result.Errors)
		summary.TotalWarnings += len(result.Warnings)
		summary.TotalSuggestions += len(result.Suggestions)

		summary.Results = append(summary.Results, result)
		ctx.LogProcessed(file.RelPath, len(result.Errors))
	}

	// Call post-processor if the linter implements it
	if pp, ok := linter.(BatchPostProcessor); ok {
		pp.PostProcessBatch(ctx, summary)
	}

	return summary
}

// lintBatchFile lints a single file in batch mode.
func lintBatchFile(ctx *LinterContext, file discovery.File, linter ComponentLinter) LintResult {
	result := LintResult{
		File:    file.RelPath,
		Type:    linter.Type(),
		Success: true,
	}

	// Pre-validation checks
	preErrors := linter.PreValidate(file.RelPath, file.Contents)
	if len(preErrors) > 0 {
		result.Errors = append(result.Errors, preErrors...)
		for _, e := range preErrors {
			if e.Severity == "error" && strings.Contains(e.Message, "is empty") {
				result.Success = len(result.Errors) == 0
				return result
			}
		}
	}

	// Parse content
	data, body, err := linter.ParseContent(file.Contents)
	if err != nil {
		result.Errors = append(result.Errors, cue.ValidationError{
			File:     file.RelPath,
			Message:  err.Error(),
			Severity: "error",
		})
		result.Success = false
		return result
	}

	// CUE validation
	if cueErrors, cueErr := linter.ValidateCUE(ctx.Validator, data); cueErr != nil {
		result.Errors = append(result.Errors, cue.ValidationError{
			File:     file.RelPath,
			Message:  fmt.Sprintf("Validation error: %v", cueErr),
			Severity: "error",
		})
	} else if cueErrors != nil {
		result.Errors = append(result.Errors, cueErrors...)
	}

	// Component-specific validation
	specificErrors := linter.ValidateSpecific(data, file.RelPath, file.Contents)
	for _, issue := range specificErrors {
		if issue.Severity == "suggestion" {
			result.Suggestions = append(result.Suggestions, issue)
		} else if issue.Severity == "warning" {
			result.Warnings = append(result.Warnings, issue)
		} else {
			result.Errors = append(result.Errors, issue)
		}
	}

	// Best practice checks
	if bpIssues := linter.ValidateBestPractices(file.RelPath, file.Contents, data); bpIssues != nil {
		for _, issue := range bpIssues {
			if issue.Severity == "suggestion" {
				result.Suggestions = append(result.Suggestions, issue)
			} else if issue.Severity == "warning" {
				result.Warnings = append(result.Warnings, issue)
			} else {
				result.Errors = append(result.Errors, issue)
			}
		}
	}

	// Cross-file validation
	if crossErrors := linter.ValidateCrossFile(ctx.CrossValidator, file.RelPath, file.Contents, data); crossErrors != nil {
		result.Errors = append(result.Errors, crossErrors...)
	}

	// Secrets detection
	secretWarnings := detectSecrets(file.Contents, file.RelPath)
	result.Warnings = append(result.Warnings, secretWarnings...)

	// Quality scoring
	if score := linter.Score(file.Contents, data, body); score != nil {
		result.Quality = score
	}

	// Improvement recommendations
	if improvements := linter.GetImprovements(file.Contents, data); improvements != nil {
		result.Improvements = improvements
	}

	// Post-processing
	linter.PostProcess(&result)

	result.Success = len(result.Errors) == 0
	return result
}

// =============================================================================
// Shared Utility Functions
// =============================================================================

// DetectXMLTags checks for XML-like tags in a string and returns an error if found.
// This is used by agents, commands, and skills to validate the description field.
func DetectXMLTags(content, fieldName, filePath, fileContents string) *cue.ValidationError {
	xmlTagPattern := regexp.MustCompile(`<[a-zA-Z][^>]*>`)
	if xmlTagPattern.MatchString(content) {
		return &cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("%s contains XML-like tags which are not allowed", fieldName),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(fileContents, strings.ToLower(fieldName)),
		}
	}
	return nil
}

// CheckSizeLimit checks if content exceeds a line limit with tolerance.
// Returns a suggestion if the limit is exceeded.
func CheckSizeLimit(contents string, limit int, tolerance float64, componentType, filePath string) *cue.ValidationError {
	lines := strings.Count(contents, "\n")
	maxLines := int(float64(limit) * (1 + tolerance))

	if lines > maxLines {
		var message string
		switch componentType {
		case "agent":
			message = fmt.Sprintf("Agent is %d lines. Best practice: keep agents under ~%d lines (%d%s%d%%) - move methodology to skills instead.",
				lines, maxLines, limit, "\u00B1", int(tolerance*100))
		case "command":
			message = fmt.Sprintf("Command is %d lines. Best practice: keep commands under ~%d lines (%d%s%d%%) - delegate to specialist agents instead of implementing logic directly.",
				lines, maxLines, limit, "\u00B1", int(tolerance*100))
		case "skill":
			message = fmt.Sprintf("Skill is %d lines. Best practice: keep skills under ~%d lines (%d%s%d%%) - move heavy docs to references/ subdirectory.",
				lines, maxLines, limit, "\u00B1", int(tolerance*100))
		default:
			message = fmt.Sprintf("%s is %d lines, exceeds recommended %d lines.",
				componentType, lines, maxLines)
		}

		return &cue.ValidationError{
			File:     filePath,
			Message:  message,
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     1,
		}
	}
	return nil
}

// parseFrontmatter parses YAML frontmatter from markdown content.
// Returns (data, body, error).
func parseFrontmatter(contents string) (map[string]interface{}, string, error) {
	fm, err := frontend.ParseYAMLFrontmatter(contents)
	if err != nil {
		return nil, "", fmt.Errorf("Error parsing frontmatter: %v", err)
	}
	return fm.Data, fm.Body, nil
}

// parseJSONContent parses JSON content into a map.
// Returns (data, "", error) - body is empty for JSON.
func parseJSONContent(contents string) (map[string]interface{}, string, error) {
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(contents), &data); err != nil {
		return nil, "", fmt.Errorf("Invalid JSON: %v", err)
	}
	return data, "", nil
}

// =============================================================================
// Base Linter Implementation
// =============================================================================

// BaseLinter provides default implementations for ComponentLinter methods.
// Embed this in component-specific linters to inherit defaults.
type BaseLinter struct{}

func (b *BaseLinter) PreValidate(filePath, contents string) []cue.ValidationError {
	return nil
}

func (b *BaseLinter) ValidateBestPractices(filePath, contents string, data map[string]interface{}) []cue.ValidationError {
	return nil
}

func (b *BaseLinter) ValidateCrossFile(crossValidator *CrossFileValidator, filePath, contents string, data map[string]interface{}) []cue.ValidationError {
	return nil
}

func (b *BaseLinter) Score(contents string, data map[string]interface{}, body string) *scoring.QualityScore {
	return nil
}

func (b *BaseLinter) GetImprovements(contents string, data map[string]interface{}) []ImprovementRecommendation {
	return nil
}

func (b *BaseLinter) PostProcess(result *LintResult) {
	// Default: no post-processing
}

// ValidateAllowedToolsShared is a shared helper for tool validation.
func ValidateAllowedToolsShared(data map[string]interface{}, filePath, contents string) []cue.ValidationError {
	return ValidateAllowedTools(data, filePath, contents)
}
