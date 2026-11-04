package cue

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"cuelang.org/go/cue"
)

// extensionComponentTypes is the closed set of component types that accept
// schema extensions. It matches the embedded schema roots keyed in loadSchemas.
var extensionComponentTypes = map[string]bool{
	"agent": true, "command": true, "skill": true, "settings": true, "claude_md": true,
}

// LoadExtensions unifies user-supplied CUE constraint files with the embedded
// schema for each component type. Keys: agent, command, skill, settings,
// claude_md. Values: a file path or a list of paths, resolved against rootPath
// when relative. The extension file must define the same definition it extends
// (e.g. #Agent); its constraints are AND-ed with the built-in schema.
//
// It must be called after LoadSchemas and before any validation runs.
func (v *Validator) LoadExtensions(rootPath string, exts map[string]any) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	// Deterministic iteration order for stable error reporting.
	keys := make([]string, 0, len(exts))
	for key := range exts {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		if !extensionComponentTypes[key] {
			return fmt.Errorf("schemas.extensions[%q]: unknown component type (must be one of agent, command, skill, settings, claude_md)", key)
		}
		base, ok := v.schemas[key]
		if !ok {
			return fmt.Errorf("schemas.extensions[%q]: no embedded CUE schema loaded for component type %q", key, key)
		}
		paths, err := extensionFilePaths(key, exts[key], rootPath)
		if err != nil {
			return err
		}
		for _, path := range paths {
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("schemas.extensions[%q]: read extension file %s: %w", key, path, err)
			}
			compiled := v.ctx.CompileBytes(content, cue.Filename(path))
			if err := compiled.Err(); err != nil {
				return fmt.Errorf("schemas.extensions[%q]: invalid CUE in extension file %s: %w", key, path, err)
			}
			unified := base.Unify(compiled)
			if err := unified.Err(); err != nil {
				return fmt.Errorf("schemas.extensions[%q]: unify extension file %s: %w", key, path, err)
			}
			v.schemas[key] = unified
			base = unified
		}
	}
	return nil
}

// extensionFilePaths converts one extension value (a path string or a list of
// path strings) into absolute file paths resolved against rootPath.
func extensionFilePaths(key string, value any, rootPath string) ([]string, error) {
	var raw []string
	switch v := value.(type) {
	case string:
		raw = []string{v}
	case []string:
		raw = v
	case []any:
		for i, elem := range v {
			s, ok := elem.(string)
			if !ok {
				return nil, fmt.Errorf("schemas.extensions[%q]: list element %d must be a string path, got %T", key, i, elem)
			}
			raw = append(raw, s)
		}
	default:
		return nil, fmt.Errorf("schemas.extensions[%q]: value must be a file path string or a list of paths, got %T", key, value)
	}

	paths := make([]string, 0, len(raw))
	for _, p := range raw {
		if !filepath.IsAbs(p) {
			p = filepath.Join(rootPath, p)
		}
		paths = append(paths, p)
	}
	return paths, nil
}
