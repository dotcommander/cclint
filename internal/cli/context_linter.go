package cli

import (
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
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
	fm, err := frontend.ParseYAMLFrontmatter(contents)

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
