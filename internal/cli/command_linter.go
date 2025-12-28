package cli

import (
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/scoring"
)

// CommandLinter implements ComponentLinter for command files.
type CommandLinter struct {
	BaseLinter
}

// NewCommandLinter creates a new CommandLinter.
func NewCommandLinter() *CommandLinter {
	return &CommandLinter{}
}

func (l *CommandLinter) Type() string {
	return "command"
}

func (l *CommandLinter) FileType() discovery.FileType {
	return discovery.FileTypeCommand
}

func (l *CommandLinter) ParseContent(contents string) (map[string]interface{}, string, error) {
	return parseFrontmatter(contents)
}

func (l *CommandLinter) ValidateCUE(validator *cue.Validator, data map[string]interface{}) ([]cue.ValidationError, error) {
	return validator.ValidateCommand(data)
}

func (l *CommandLinter) ValidateSpecific(data map[string]interface{}, filePath, contents string) []cue.ValidationError {
	errors := validateCommandSpecific(data, filePath, contents)

	// Validate allowed-tools
	toolWarnings := ValidateAllowedTools(data, filePath, contents)
	for _, w := range toolWarnings {
		errors = append(errors, w)
	}

	return errors
}

func (l *CommandLinter) ValidateBestPractices(filePath, contents string, data map[string]interface{}) []cue.ValidationError {
	return validateCommandBestPractices(filePath, contents, data)
}

func (l *CommandLinter) ValidateCrossFile(crossValidator *CrossFileValidator, filePath, contents string, data map[string]interface{}) []cue.ValidationError {
	if crossValidator == nil {
		return nil
	}
	return crossValidator.ValidateCommand(filePath, contents, data)
}

func (l *CommandLinter) Score(contents string, data map[string]interface{}, body string) *scoring.QualityScore {
	scorer := scoring.NewCommandScorer()
	score := scorer.Score(contents, data, body)
	return &score
}

func (l *CommandLinter) GetImprovements(contents string, data map[string]interface{}) []ImprovementRecommendation {
	return GetCommandImprovements(contents, data)
}
