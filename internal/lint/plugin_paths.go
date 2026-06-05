package lint

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// pluginPathFields lists plugin fields that contain file paths
var pluginPathFields = []string{
	"commands", "agents", "skills", "hooks", "mcpServers", "outputStyles", "lspServers", "experimental",
}

// allPluginPathFields includes all fields that contain filesystem paths,
// adding "readme" to the component path fields since it references a file too.
var allPluginPathFields = append([]string{"readme"}, pluginPathFields...)

// isExternalPlugin returns true for marketplace or cache plugins (third-party, not user-authored).
func isExternalPlugin(filePath string) bool {
	return strings.Contains(filePath, "marketplaces/") || strings.Contains(filePath, "cache/")
}

// isGlobPattern returns true if the path contains glob metacharacters.
// Paths with globs are skipped for existence checks since they represent
// patterns, not literal filesystem paths.
func isGlobPattern(path string) bool {
	return strings.ContainsAny(path, "*?[{")
}

// validatePluginPaths checks that path-bearing fields use relative paths
func validatePluginPaths(data map[string]any, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	for _, field := range pluginPathFields {
		value, ok := data[field]
		if !ok {
			continue
		}
		paths := extractPaths(value)
		for _, p := range paths {
			errors = append(errors, checkPath(p, field, filePath, contents)...)
		}
	}

	return errors
}

// extractPaths collects path strings from various JSON structures:
// string, []string, []object (extracts string values), or map (extracts string values).
func extractPaths(value any) []string {
	var paths []string
	switch v := value.(type) {
	case string:
		paths = append(paths, v)
	case []any:
		for _, item := range v {
			switch elem := item.(type) {
			case string:
				paths = append(paths, elem)
			case map[string]any:
				for _, mv := range elem {
					if s, ok := mv.(string); ok {
						paths = append(paths, s)
					}
				}
			}
		}
	case map[string]any:
		for _, mv := range v {
			if s, ok := mv.(string); ok {
				paths = append(paths, s)
			}
		}
	}
	return paths
}

// checkPath validates a single path string for relative path requirements.
func checkPath(path string, field string, filePath string, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	if strings.HasPrefix(path, "/") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Plugin paths must be relative (start with \"./\"): found \"%s\"", path),
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Line:     FindJSONFieldLine(contents, field),
		})
	}

	if strings.Contains(path, "..") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Path '%s' in '%s' contains '..', which risks traversal outside the plugin root", path, field),
			Severity: cue.SeverityWarning,
			Source:   cue.SourceCClintObserve,
			Line:     FindJSONFieldLine(contents, field),
		})
	}

	return errors
}

// validatePluginPathsExist checks that literal paths referenced in plugin
// manifests exist on disk relative to the plugin's directory.
//
// Reports missing paths as warnings (not errors) since referenced files may
// be generated at build time or installed later.
//
// Skips validation when rootPath is empty (e.g., in unit tests without a
// filesystem layout). Skips glob patterns, absolute paths, and traversal
// paths (the latter two are already reported by checkPath).
func validatePluginPathsExist(data map[string]any, rootPath, filePath, contents string) []cue.ValidationError {
	if rootPath == "" {
		return nil
	}

	// Resolve the plugin directory: filePath is relative to rootPath.
	// e.g., filePath="my-plugin/.claude-plugin/plugin.json"
	//   -> pluginDir = <rootPath>/my-plugin/.claude-plugin
	pluginDir := filepath.Join(rootPath, filepath.Dir(filePath))

	var warnings []cue.ValidationError

	for _, field := range allPluginPathFields {
		value, ok := data[field]
		if !ok {
			continue
		}
		paths := extractPaths(value)
		for _, p := range paths {
			if isGlobPattern(p) {
				continue
			}
			// Skip absolute paths and traversal (already reported by checkPath)
			if strings.HasPrefix(p, "/") || strings.Contains(p, "..") {
				continue
			}
			// Skip empty paths and URLs
			if p == "" || strings.HasPrefix(p, "http://") || strings.HasPrefix(p, "https://") {
				continue
			}

			resolved := filepath.Join(pluginDir, p)
			if _, err := os.Stat(resolved); os.IsNotExist(err) {
				warnings = append(warnings, cue.ValidationError{
					File:     filePath,
					Message:  fmt.Sprintf("Path '%s' in '%s' does not exist relative to plugin directory", p, field),
					Severity: cue.SeverityWarning,
					Source:   cue.SourceCClintObserve,
					Line:     FindJSONFieldLine(contents, field),
				})
			}
		}
	}

	return warnings
}
