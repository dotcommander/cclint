package cli

import (
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

// SettingsLinter implements ComponentLinter for settings files.
// It implements only the core ComponentLinter interface - no optional capabilities.
// Settings files don't need scoring, improvements, cross-file validation, etc.
type SettingsLinter struct {
	BaseLinter
}

// Compile-time interface compliance check
var _ ComponentLinter = (*SettingsLinter)(nil)

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

func (l *SettingsLinter) ParseContent(contents string) (map[string]any, string, error) {
	return parseJSONContent(contents)
}

func (l *SettingsLinter) ValidateCUE(validator *cue.Validator, data map[string]any) ([]cue.ValidationError, error) {
	return validator.ValidateSettings(data)
}

func (l *SettingsLinter) ValidateSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	return validateSettingsSpecific(data, filePath)
}
