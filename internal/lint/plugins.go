package lint

import (
	"slices"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

// LintPlugins runs linting on plugin manifest files using the generic linter.
func LintPlugins(rootPath string, quiet bool, verbose bool, noCycleCheck bool, exclude []string) (*LintSummary, error) {
	return lintStandalone(rootPath, quiet, verbose, noCycleCheck, exclude, discovery.FileTypePlugin)
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
	errors = append(errors, validatePluginTypes(data, filePath, contents)...)
	errors = append(errors, validatePluginPaths(data, filePath, contents)...)
	errors = append(errors, validatePluginBestPractices(filePath, contents, data)...)

	if isExternalPlugin(filePath) {
		return slices.DeleteFunc(errors, func(e cue.ValidationError) bool {
			return e.Severity != "error" && e.Severity != "warning"
		})
	}

	return errors
}

// validatePluginTypes checks the optional "types" field, which names the
// plugin's TypeScript type contract (v2.1.299+). Claude Code accepts only
// string values ending in ".d.ts" that stay inside the plugin directory.
func validatePluginTypes(data map[string]any, filePath, contents string) []cue.ValidationError {
	raw, ok := data["types"]
	if !ok {
		return nil
	}

	name, ok := raw.(string)
	if !ok || !strings.HasSuffix(name, ".d.ts") {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "types must name a TypeScript declaration file ending in '.d.ts'",
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "types"),
		}}
	}

	if strings.Contains(name, "..") {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "types must stay inside the plugin directory (no '..' segment)",
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, "types"),
		}}
	}

	return nil
}
