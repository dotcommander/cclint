package lint

import (
	"slices"

	"github.com/dotcommander/cclint/internal/cue"
)

// LintPlugins runs linting on plugin manifest files using the generic linter.
func LintPlugins(rootPath string, quiet bool, verbose bool, noCycleCheck bool, exclude []string) (*LintSummary, error) {
	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck, exclude)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, NewPluginLinter(ctx.RootPath)), nil
}

// validatePluginSpecific implements plugin-specific validation rules.
// External plugins (marketplace/cache) only get error-level checks — suggestions are suppressed
// since their metadata is third-party and not user-controlled.
func validatePluginSpecific(data map[string]any, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	errors = append(errors, validateUnknownPluginFields(data, filePath, contents)...)
	errors = append(errors, validatePluginName(data, filePath, contents)...)
	errors = append(errors, validatePluginDescription(data, filePath, contents)...)
	errors = append(errors, validatePluginVersion(data, filePath, contents)...)
	errors = append(errors, validatePluginAuthor(data, filePath, contents)...)
	errors = append(errors, validatePluginPaths(data, filePath, contents)...)
	errors = append(errors, validatePluginBestPractices(filePath, contents, data)...)

	if isExternalPlugin(filePath) {
		return slices.DeleteFunc(errors, func(e cue.ValidationError) bool {
			return e.Severity != "error" && e.Severity != "warning"
		})
	}

	return errors
}
