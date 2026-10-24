package cue

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"cuelang.org/go/cue"
)

//go:embed schemas/*.cue
var schemaFS embed.FS

// LoadSchemas loads the embedded roots, publishing only a complete schema set.
func (v *Validator) LoadSchemas() error {
	return v.loadSchemas(schemaFS)
}

func (v *Validator) loadSchemas(fsys fs.FS) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	entries, err := fs.ReadDir(fsys, "schemas")
	if err != nil {
		return fmt.Errorf("read CUE schema directory schemas: %w", err)
	}
	fragment, err := fs.ReadFile(fsys, "schemas/component_hooks.cue")
	if err != nil {
		return fmt.Errorf("read CUE fragment schemas/component_hooks.cue: %w", err)
	}

	candidate := make(map[string]cue.Value)
	var loadErr error
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cue") || entry.Name() == "component_hooks.cue" {
			continue
		}
		filename := path.Join("schemas", entry.Name())
		content, err := fs.ReadFile(fsys, filename)
		if err != nil {
			loadErr = errors.Join(loadErr, fmt.Errorf("read CUE schema %s: %w", filename, err))
			continue
		}
		schema, err := v.compileSchemaLocked(filename, content, fragment)
		if err != nil {
			loadErr = errors.Join(loadErr, err)
			continue
		}
		candidate[strings.TrimSuffix(entry.Name(), ".cue")] = schema
	}
	if loadErr != nil {
		return loadErr
	}
	if len(candidate) == 0 {
		return errors.New("warning: no CUE schemas loaded, using Go validation")
	}

	// Previously published values remain usable if any candidate root fails.
	v.schemas = candidate
	return nil
}

// compileSchemaLocked composes one root without changing either input. The
// caller holds v.mu because every compilation uses the validator's CUE context.
func (v *Validator) compileSchemaLocked(filename string, root, fragment []byte) (cue.Value, error) {
	var content bytes.Buffer
	content.Write(root)
	content.WriteByte('\n')
	content.Write(fragment)
	for _, injection := range []struct {
		token    string
		generate func() string
	}{
		{"#KnownTool", knownToolUnionCUE},
		{"#Model", modelUnionCUE},
	} {
		if bytes.Contains(root, []byte(injection.token)) {
			content.WriteByte('\n')
			content.WriteString(injection.generate())
			content.WriteByte('\n')
		}
	}
	compiled := v.ctx.CompileBytes(content.Bytes(), cue.Filename(filename))
	if err := compiled.Err(); err != nil {
		return cue.Value{}, fmt.Errorf("compile CUE schema %s with component_hooks.cue: %w", filename, err)
	}
	return compiled.Value(), nil
}
