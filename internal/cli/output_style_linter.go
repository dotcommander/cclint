package cli

import (
	"fmt"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
	"github.com/dotcommander/cclint/internal/scoring"
)

// LintOutputStyles runs linting on output style files using the generic linter.
func LintOutputStyles(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewOutputStyleLinter()), nil
}

// OutputStyleLinter implements ComponentLinter for output style markdown files.
// It also implements Scorable for quality scoring.
type OutputStyleLinter struct {
	BaseLinter
}

// Compile-time interface compliance checks
var (
	_ ComponentLinter = (*OutputStyleLinter)(nil)
	_ PreValidator    = (*OutputStyleLinter)(nil)
	_ Scorable        = (*OutputStyleLinter)(nil)
)

// knownOutputStyleFields lists valid frontmatter fields for output style files.
var knownOutputStyleFields = map[string]bool{
	"name":                     true, // Required: output style identifier
	"description":              true, // Required: what this style does
	"keep-coding-instructions": true, // Optional: boolean, whether to keep default coding instructions
}

// NewOutputStyleLinter creates a new OutputStyleLinter.
func NewOutputStyleLinter() *OutputStyleLinter {
	return &OutputStyleLinter{}
}

func (l *OutputStyleLinter) Type() string {
	return "output-style"
}

func (l *OutputStyleLinter) FileType() discovery.FileType {
	return discovery.FileTypeOutputStyle
}

// PreValidate implements PreValidator interface.
func (l *OutputStyleLinter) PreValidate(filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check file extension
	if !strings.HasSuffix(strings.ToLower(filePath), ".md") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Output style files must have .md extension",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	// Check empty content
	if strings.TrimSpace(contents) == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Output style file is empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	return errors
}

func (l *OutputStyleLinter) ParseContent(contents string) (map[string]any, string, error) {
	fm, err := frontend.ParseYAMLFrontmatter(contents)
	if err != nil {
		return nil, "", fmt.Errorf("error parsing frontmatter: %v", err)
	}
	return fm.Data, fm.Body, nil
}

func (l *OutputStyleLinter) ValidateCUE(validator *cue.Validator, data map[string]any) ([]cue.ValidationError, error) {
	// Output styles don't use CUE validation
	return nil, nil
}

func (l *OutputStyleLinter) ValidateSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check for unknown frontmatter fields
	for key := range data {
		if !knownOutputStyleFields[key] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown frontmatter field '%s'. Valid fields: name, description, keep-coding-instructions", key),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
				Line:     FindFrontmatterFieldLine(contents, key),
			})
		}
	}

	// Required: name field
	if name, ok := data["name"].(string); !ok || strings.TrimSpace(name) == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'name' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "name"),
		})
	} else {
		errors = append(errors, validateOutputStyleName(name, filePath, contents)...)
	}

	// Required: description field
	if desc, ok := data["description"].(string); !ok || strings.TrimSpace(desc) == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Required field 'description' is missing or empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "description"),
		})
	}

	// Optional: keep-coding-instructions must be boolean if present
	if kci, ok := data["keep-coding-instructions"]; ok {
		if _, isBool := kci.(bool); !isBool {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("keep-coding-instructions must be a boolean (got '%v')", kci),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "keep-coding-instructions"),
			})
		}
	}

	// Body must not be empty - output styles need content for system prompt customization
	_, body, _ := l.ParseContent(contents)
	if strings.TrimSpace(body) == "" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Output style body is empty - add markdown content for system prompt customization",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	return errors
}

// validateOutputStyleName validates kebab-case format and hyphen placement for output style names.
func validateOutputStyleName(name, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	if !isKebabCase(name) {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Name must contain only lowercase letters, numbers, and hyphens (kebab-case)",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "name"),
		})
	}

	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Name '%s' cannot start or end with a hyphen", name),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "name"),
		})
	}

	return errors
}

// Score implements Scorable interface.
func (l *OutputStyleLinter) Score(contents string, data map[string]any, body string) *scoring.QualityScore {
	scorer := scoring.NewOutputStyleScorer()
	score := scorer.Score(contents, data, body)
	return &score
}
