package cli

import (
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/scoring"
)

// SettingsLinter implements ComponentLinter for settings files.
type SettingsLinter struct {
	BaseLinter
}

// NewSettingsLinter creates a new SettingsLinter.
func NewSettingsLinter() *SettingsLinter {
	return &SettingsLinter{}
}

func (l *SettingsLinter) Type() string {
	return "settings"
}

func (l *SettingsLinter) FileType() discovery.FileType {
	return discovery.FileTypeSettings
}

func (l *SettingsLinter) ParseContent(contents string) (map[string]interface{}, string, error) {
	return parseJSONContent(contents)
}

func (l *SettingsLinter) ValidateCUE(validator *cue.Validator, data map[string]interface{}) ([]cue.ValidationError, error) {
	return validator.ValidateSettings(data)
}

func (l *SettingsLinter) ValidateSpecific(data map[string]interface{}, filePath, contents string) []cue.ValidationError {
	return validateSettingsSpecific(data, filePath)
}

// Settings don't have cross-file validation, scoring, or improvements
func (l *SettingsLinter) Score(contents string, data map[string]interface{}, body string) *scoring.QualityScore {
	return nil
}
