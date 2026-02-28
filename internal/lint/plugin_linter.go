package lint

import (
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/scoring"
	"github.com/dotcommander/cclint/internal/textutil"
)

// PluginLinter implements ComponentLinter for plugin.json files.
// It also implements Scorable and Improvable for quality scoring.
type PluginLinter struct {
	BaseLinter
	// RootPath is the project root directory, used to resolve relative paths
	// for filesystem existence checks. Empty string disables path existence validation.
	RootPath string
}

// Compile-time interface compliance checks
var (
	_ ComponentLinter = (*PluginLinter)(nil)
	_ Scorable        = (*PluginLinter)(nil)
	_ Improvable      = (*PluginLinter)(nil)
)

// NewPluginLinter creates a new PluginLinter.
// rootPath is the project root for resolving relative paths in plugin manifests.
// Pass empty string to skip path existence validation.
func NewPluginLinter(rootPath string) *PluginLinter {
	return &PluginLinter{RootPath: rootPath}
}

func (l *PluginLinter) Type() string {
	return "plugin"
}

func (l *PluginLinter) FileType() discovery.FileType {
	return discovery.FileTypePlugin
}

func (l *PluginLinter) ParseContent(contents string) (map[string]any, string, error) {
	return parseJSONContent(contents)
}

func (l *PluginLinter) ValidateCUE(validator *cue.Validator, data map[string]any) ([]cue.ValidationError, error) {
	// Plugins don't use CUE validation
	return nil, nil
}

func (l *PluginLinter) ValidateSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	errors := validatePluginSpecific(data, filePath, contents)
	errors = append(errors, validatePluginPathsExist(data, l.RootPath, filePath, contents)...)
	return errors
}

// Score implements Scorable interface
func (l *PluginLinter) Score(contents string, data map[string]any, body string) *scoring.QualityScore {
	scorer := scoring.NewPluginScorer()
	score := scorer.Score(contents, data, body)
	return &score
}

// GetImprovements implements Improvable interface
func (l *PluginLinter) GetImprovements(contents string, data map[string]any) []textutil.ImprovementRecommendation {
	return GetPluginImprovements(contents, data)
}
