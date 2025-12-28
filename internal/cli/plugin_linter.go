package cli

import (
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/scoring"
)

// PluginLinter implements ComponentLinter for plugin.json files.
type PluginLinter struct {
	BaseLinter
}

// NewPluginLinter creates a new PluginLinter.
func NewPluginLinter() *PluginLinter {
	return &PluginLinter{}
}

func (l *PluginLinter) Type() string {
	return "plugin"
}

func (l *PluginLinter) FileType() discovery.FileType {
	return discovery.FileTypePlugin
}

func (l *PluginLinter) ParseContent(contents string) (map[string]interface{}, string, error) {
	return parseJSONContent(contents)
}

func (l *PluginLinter) ValidateCUE(validator *cue.Validator, data map[string]interface{}) ([]cue.ValidationError, error) {
	// Plugins don't use CUE validation
	return nil, nil
}

func (l *PluginLinter) ValidateSpecific(data map[string]interface{}, filePath, contents string) []cue.ValidationError {
	return validatePluginSpecific(data, filePath, contents)
}

func (l *PluginLinter) Score(contents string, data map[string]interface{}, body string) *scoring.QualityScore {
	scorer := scoring.NewPluginScorer()
	score := scorer.Score(contents, data, body)
	return &score
}

func (l *PluginLinter) GetImprovements(contents string, data map[string]interface{}) []ImprovementRecommendation {
	return GetPluginImprovements(contents, data)
}

// PostProcess handles plugin-specific filtering (only show errors)
func (l *PluginLinter) PostProcess(result *LintResult) {
	// For batch mode, we may want to filter suggestions/warnings
	// This is kept for compatibility but the actual filtering happens in LintPlugins
}
