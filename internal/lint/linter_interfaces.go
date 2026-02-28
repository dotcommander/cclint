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
//   - ISP-compliant: Core interface is minimal; optional behaviors via separate interfaces
package lint

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/textutil"
	"github.com/dotcommander/cclint/internal/scoring"
)

// =============================================================================
// Interface Segregation: Core interface + optional capability interfaces
// =============================================================================

// ComponentLinter defines the minimal interface for type-specific linting logic.
// Each component type (agent, command, skill, etc.) implements this interface.
//
// For optional behaviors (scoring, improvements, cross-file validation), implement
// the corresponding capability interface (Scorable, Improvable, CrossFileValidatable).
type ComponentLinter interface {
	// Type returns the component type name (e.g., "agent", "command")
	Type() string

	// FileType returns the discovery.FileType for filtering
	FileType() discovery.FileType

	// ParseContent parses the file content and returns frontmatter data and body.
	// For JSON files, body may be empty.
	ParseContent(contents string) (data map[string]any, body string, err error)

	// ValidateCUE runs CUE schema validation if applicable.
	// Returns nil if CUE validation is not used for this type.
	ValidateCUE(validator *cue.Validator, data map[string]any) ([]cue.ValidationError, error)

	// ValidateSpecific runs component-specific validation rules.
	ValidateSpecific(data map[string]any, filePath, contents string) []cue.ValidationError
}

// PreValidator is an optional interface for linters that need pre-validation checks.
// Implement this for components with special filename requirements or empty content checks.
type PreValidator interface {
	// PreValidate runs any pre-validation checks (e.g., filename, empty content).
	// Returns errors that should abort further validation.
	PreValidate(filePath, contents string) []cue.ValidationError
}

// BestPracticeValidator is an optional interface for linters with best practice checks.
type BestPracticeValidator interface {
	// ValidateBestPractices runs best practice checks beyond basic validation.
	ValidateBestPractices(filePath, contents string, data map[string]any) []cue.ValidationError
}

// CrossFileValidatable is an optional interface for linters that validate cross-file references.
type CrossFileValidatable interface {
	// ValidateCrossFile runs cross-file validation (e.g., agent→skill references).
	ValidateCrossFile(crossValidator *crossfile.CrossFileValidator, filePath, contents string, data map[string]any) []cue.ValidationError
}

// Scorable is an optional interface for linters that provide quality scores.
type Scorable interface {
	// Score returns the quality score for the component.
	Score(contents string, data map[string]any, body string) *scoring.QualityScore
}

// Improvable is an optional interface for linters that provide improvement recommendations.
type Improvable interface {
	// GetImprovements returns improvement recommendations with point values.
	GetImprovements(contents string, data map[string]any) []textutil.ImprovementRecommendation
}

// PostProcessable is an optional interface for linters needing result post-processing.
type PostProcessable interface {
	// PostProcess allows type-specific post-processing of results.
	PostProcess(result *LintResult)
}

// lintFileCore contains the shared linting logic for both single-file and batch modes.
// It validates a file using the provided ComponentLinter and returns the result.
// crossValidator may be nil if cross-file validation should be skipped.
func lintFileCore(filePath, contents string, linter ComponentLinter, validator *cue.Validator, crossValidator *crossfile.CrossFileValidator) LintResult {
	result := LintResult{
		File:    filePath,
		Type:    linter.Type(),
		Success: true,
	}

	// Pre-validation checks (filename, empty content, etc.) - optional capability
	if shouldAbort := runPreValidation(&result, filePath, contents, linter); shouldAbort {
		return result
	}

	// Parse content
	data, body, parseErr := linter.ParseContent(contents)
	if parseErr != nil {
		result.Errors = append(result.Errors, cue.ValidationError{
			File:     filePath,
			Message:  parseErr.Error(),
			Severity: cue.SeverityError,
		})
		result.Success = false
		return result
	}

	// Check for swallowed frontmatter fields (block scalar absorbed siblings)
	swallowedWarnings := DetectSwallowedFields(contents, filePath, linter.Type())
	categorizeIssues(&result, swallowedWarnings)

	// Run all validation steps
	runCUEValidation(&result, filePath, linter, validator, data)
	runComponentSpecificValidation(&result, linter, data, filePath, contents)
	runBestPracticeValidation(&result, linter, filePath, contents, data)
	runCrossFileValidation(crossFileValidationParams{
		result:         &result,
		linter:         linter,
		crossValidator: crossValidator,
		filePath:       filePath,
		contents:       contents,
		data:           data,
	})

	// Secrets detection (common to all types)
	secretWarnings := textutil.DetectSecrets(contents, filePath)
	result.Warnings = append(result.Warnings, secretWarnings...)

	// Quality scoring - optional capability
	if sc, ok := linter.(Scorable); ok {
		if score := sc.Score(contents, data, body); score != nil {
			result.Quality = score
		}
	}

	// Improvement recommendations - optional capability
	if imp, ok := linter.(Improvable); ok {
		if improvements := imp.GetImprovements(contents, data); improvements != nil {
			result.Improvements = improvements
		}
	}

	// Allow type-specific post-processing - optional capability
	if pp, ok := linter.(PostProcessable); ok {
		pp.PostProcess(&result)
	}

	result.Success = len(result.Errors) == 0
	return result
}

// runPreValidation runs pre-validation checks and returns true if validation should abort.
func runPreValidation(result *LintResult, filePath, contents string, linter ComponentLinter) bool {
	pv, ok := linter.(PreValidator)
	if !ok {
		return false
	}

	preErrors := pv.PreValidate(filePath, contents)
	if len(preErrors) == 0 {
		return false
	}

	result.Errors = append(result.Errors, preErrors...)

	// Check if any are fatal (should abort further validation)
	for _, e := range preErrors {
		if e.Severity == cue.SeverityError && strings.Contains(e.Message, "is empty") {
			result.Success = len(result.Errors) == 0
			return true
		}
	}
	return false
}

// runCUEValidation runs CUE schema validation.
func runCUEValidation(result *LintResult, filePath string, linter ComponentLinter, validator *cue.Validator, data map[string]any) {
	cueErrors, cueErr := linter.ValidateCUE(validator, data)
	if cueErr != nil {
		result.Errors = append(result.Errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Validation error: %v", cueErr),
			Severity: cue.SeverityError,
		})
	} else if cueErrors != nil {
		result.Errors = append(result.Errors, cueErrors...)
	}
}

// runComponentSpecificValidation runs component-specific validation rules.
func runComponentSpecificValidation(result *LintResult, linter ComponentLinter, data map[string]any, filePath, contents string) {
	specificErrors := linter.ValidateSpecific(data, filePath, contents)
	categorizeIssues(result, specificErrors)
}

// runBestPracticeValidation runs best practice checks.
func runBestPracticeValidation(result *LintResult, linter ComponentLinter, filePath, contents string, data map[string]any) {
	bpv, ok := linter.(BestPracticeValidator)
	if !ok {
		return
	}

	if bpIssues := bpv.ValidateBestPractices(filePath, contents, data); bpIssues != nil {
		categorizeIssues(result, bpIssues)
	}
}

// crossFileValidationParams groups parameters for cross-file validation.
type crossFileValidationParams struct {
	result         *LintResult
	linter         ComponentLinter
	crossValidator *crossfile.CrossFileValidator
	filePath       string
	contents       string
	data           map[string]any
}

// runCrossFileValidation runs cross-file validation checks.
func runCrossFileValidation(params crossFileValidationParams) {
	if params.crossValidator == nil {
		return
	}

	cfv, ok := params.linter.(CrossFileValidatable)
	if !ok {
		return
	}

	if crossErrors := cfv.ValidateCrossFile(params.crossValidator, params.filePath, params.contents, params.data); crossErrors != nil {
		categorizeIssues(params.result, crossErrors)
	}
}

// categorizeIssues distributes validation issues into errors, warnings, or suggestions.
func categorizeIssues(result *LintResult, issues []cue.ValidationError) {
	for _, issue := range issues {
		switch issue.Severity {
		case cue.SeveritySuggestion:
			result.Suggestions = append(result.Suggestions, issue)
		case cue.SeverityInfo:
			result.Suggestions = append(result.Suggestions, issue)
		case cue.SeverityWarning:
			result.Warnings = append(result.Warnings, issue)
		default:
			result.Errors = append(result.Errors, issue)
		}
	}
}

// lintComponent is the generic single-file linting function.
// It orchestrates the linting pipeline using a ComponentLinter.
func lintComponent(ctx *SingleFileLinterContext, linter ComponentLinter) LintResult {
	crossValidator := ctx.EnsureCrossFileValidator()
	result := lintFileCore(ctx.File.RelPath, ctx.File.Contents, linter, ctx.Validator, crossValidator)

	// Add info message if cross-file validation was skipped
	if crossValidator == nil && !ctx.Quiet {
		result.Suggestions = append(result.Suggestions, cue.ValidationError{
			File:     ctx.File.RelPath,
			Message:  "Cross-file validation skipped (could not discover project files)",
			Severity: "info",
			Source:   cue.SourceCClintObserve,
		})
	}

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
	summary.ComponentType = linter.Type()

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
// Delegates to lintFileCore for the actual validation logic.
func lintBatchFile(ctx *LinterContext, file discovery.File, linter ComponentLinter) LintResult {
	return lintFileCore(file.RelPath, file.Contents, linter, ctx.Validator, ctx.CrossValidator)
}

// =============================================================================
// Shared Utility Functions
// =============================================================================

// knownFrontmatterFields maps component types to their known field names.
// Used by DetectSwallowedFields to identify when a block scalar has absorbed
// what should be a sibling frontmatter field.
var knownFrontmatterFields = map[string]map[string]bool{
	"agent": {
		"name": true, "description": true, "model": true, "color": true,
		"tools": true, "disallowedTools": true, "permissionMode": true,
		"maxTurns": true, "skills": true, "hooks": true, "memory": true,
		"mcpServers": true,
	},
	"command": {
		"name": true, "description": true, "allowed-tools": true,
		"argument-hint": true, "model": true, "disable-model-invocation": true,
	},
	"skill": {
		"name": true, "description": true, "argument-hint": true,
		"disable-model-invocation": true, "user-invocable": true,
		"allowed-tools": true, "model": true, "context": true, "agent": true,
		"hooks": true, "license": true, "compatibility": true, "metadata": true,
		"version": true,
	},
}

// blockScalarPattern matches YAML block scalar indicators (| or >) with optional modifiers.
var blockScalarPattern = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_-]*)\s*:\s*[|>][-+]?\s*$`)

// DetectSwallowedFields checks for YAML block scalar fields (| or >) that have
// accidentally absorbed subsequent frontmatter fields as part of their text content.
//
// Example of the bug this catches:
//
//	---
//	description: |
//	  Some text here.
//	model: haiku        ← YAML treats this as part of description, not a separate field
//	---
//
// This happens when a new field is inserted after a block scalar without proper
// indentation awareness. The YAML parser silently absorbs the new field as text.
func DetectSwallowedFields(contents, filePath, componentType string) []cue.ValidationError {
	lines := strings.Split(contents, "\n")

	// Find frontmatter boundaries
	fmStart, fmEnd := -1, -1
	for i, line := range lines {
		if strings.TrimSpace(line) == "---" {
			if fmStart == -1 {
				fmStart = i
			} else {
				fmEnd = i
				break
			}
		}
	}
	if fmStart == -1 || fmEnd == -1 {
		return nil
	}

	// Only check component types with known field sets.
	// Types like settings, context, and plugin don't have block scalar
	// frontmatter fields, so checking them risks false positives.
	fields := knownFrontmatterFields[componentType]
	if fields == nil {
		return nil
	}

	var errors []cue.ValidationError

	// Scan frontmatter lines for block scalar indicators
	fmLines := lines[fmStart+1 : fmEnd]
	for i, line := range fmLines {
		match := blockScalarPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		scalarField := match[1]

		// Scan subsequent lines that are part of this block scalar's content.
		// Block scalar content is indented relative to the field; the first
		// non-empty line after the indicator sets the indentation level.
		// A line at column 0 with "key: value" terminates the block — but
		// YAML only does that if the line is NOT indented. If the line IS
		// indented (even by one space), it remains part of the scalar.
		//
		// We flag any indented line inside the block scalar that looks like
		// a known frontmatter field (e.g., "  model: haiku").
		for j := i + 1; j < len(fmLines); j++ {
			subsequent := fmLines[j]

			// Empty lines are valid block scalar content
			if strings.TrimSpace(subsequent) == "" {
				continue
			}

			// Non-indented line = end of block scalar
			if len(subsequent) > 0 && subsequent[0] != ' ' && subsequent[0] != '\t' {
				break
			}

			// Check if indented line looks like a swallowed field
			trimmed := strings.TrimSpace(subsequent)
			colonIdx := strings.Index(trimmed, ":")
			if colonIdx <= 0 {
				continue
			}
			candidateKey := trimmed[:colonIdx]

			if !fields[candidateKey] {
				continue
			}

			// This indented line matches a known frontmatter field name —
			// it was almost certainly swallowed by the block scalar above.
			lineNum := fmStart + 1 + j + 1 // 1-based
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Field '%s' appears to be swallowed by block scalar '%s: |' above — it is parsed as text, not a separate field", candidateKey, scalarField),
				Severity: cue.SeverityError,
				Source:   cue.SourceCClintObserve,
				Line:     lineNum,
			})
		}
	}

	return errors
}

// DetectXMLTags checks for XML-like tags in a string and returns an error if found.
// This is used by agents, commands, and skills to validate the description field.
func DetectXMLTags(content, fieldName, filePath, fileContents string) *cue.ValidationError {
	xmlTagPattern := regexp.MustCompile(`<[a-zA-Z][^>]*>`)
	if xmlTagPattern.MatchString(content) {
		return &cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("%s contains XML-like tags which are not allowed", fieldName),
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     textutil.FindFrontmatterFieldLine(fileContents, strings.ToLower(fieldName)),
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
		case cue.TypeAgent:
			message = fmt.Sprintf("Agent is %d lines. Best practice: keep agents under ~%d lines (%d%s%d%%) - move methodology to skills instead.",
				lines, maxLines, limit, "\u00B1", int(tolerance*100))
		case cue.TypeCommand:
			message = fmt.Sprintf("Command is %d lines. Best practice: keep commands under ~%d lines (%d%s%d%%) - delegate to specialist agents instead of implementing logic directly.",
				lines, maxLines, limit, "\u00B1", int(tolerance*100))
		case cue.TypeSkill:
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

// semverPattern is the compiled regex for semver validation.
// Extracted to avoid recompilation on each call.
var semverPattern = regexp.MustCompile(`^\d+\.\d+\.\d+(-[a-zA-Z0-9.]+)?(\+[a-zA-Z0-9.]+)?$`)

// ValidateSemver checks if a version string follows semver format.
// Returns nil if valid, or a ValidationError if invalid.
func ValidateSemver(version, filePath string, line int) *cue.ValidationError {
	if !semverPattern.MatchString(version) {
		return &cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Version '%s' should follow semver format (e.g., '1.0.0')", version),
			Severity: "warning",
			Source:   cue.SourceCClintObserve,
			Line:     line,
		}
	}
	return nil
}

// parseFrontmatter parses YAML frontmatter from markdown content.
// Returns (data, body, error).
func parseFrontmatter(contents string) (map[string]any, string, error) {
	fm, err := textutil.ParseYAMLFrontmatter(contents)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing frontmatter: %v", err)
	}
	return fm.Data, fm.Body, nil
}

// parseJSONContent parses JSON content into a map.
// Returns (data, "", error) - body is empty for JSON.
func parseJSONContent(contents string) (map[string]any, string, error) {
	var data map[string]any
	if err := json.Unmarshal([]byte(contents), &data); err != nil {
		return nil, "", fmt.Errorf("invalid JSON: %v", err)
	}
	return data, "", nil
}

// =============================================================================
// Base Linter Implementation
// =============================================================================

// BaseLinter is an empty struct that component-specific linters can embed.
// With the ISP refactoring, optional capabilities are now separate interfaces
// that linters implement only when needed. This struct remains for embedding
// to signal "I'm a component linter" but provides no default methods.
//
// Optional interfaces a linter can implement:
//   - PreValidator: for filename/empty content checks
//   - BestPracticeValidator: for best practice checks
//   - CrossFileValidatable: for cross-file reference validation
//   - Scorable: for quality scoring
//   - Improvable: for improvement recommendations
//   - PostProcessable: for result post-processing
//   - BatchPostProcessor: for batch-level post-processing (cycle detection)
type BaseLinter struct{}

// ValidateAllowedToolsShared is a shared helper for tool validation.
func ValidateAllowedToolsShared(data map[string]any, filePath, contents string) []cue.ValidationError {
	return textutil.ValidateAllowedTools(data, filePath, contents)
}
