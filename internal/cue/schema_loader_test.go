package cue

import (
	"errors"
	"io/fs"
	"maps"
	"strings"
	"testing"
	"testing/fstest"

	"cuelang.org/go/cue"
)

// deniedSchemaFS changes only one fixture's read behavior, never the embed.
type deniedSchemaFS struct {
	fs.FS
	denied string
}

func (f deniedSchemaFS) Open(name string) (fs.File, error) {
	if name == f.denied {
		return nil, &fs.PathError{Op: "open", Path: name, Err: fs.ErrPermission}
	}
	return f.FS.Open(name)
}

func schemaLoaderFixture(t *testing.T) fstest.MapFS {
	t.Helper()
	entries, err := fs.ReadDir(schemaFS, "schemas")
	if err != nil {
		t.Fatal(err)
	}
	fixture := fstest.MapFS{}
	for _, entry := range entries {
		name := "schemas/" + entry.Name()
		data, err := fs.ReadFile(schemaFS, name)
		if err != nil {
			t.Fatal(err)
		}
		fixture[name] = &fstest.MapFile{Data: data}
	}
	return fixture
}

func TestSchemaLoaderPublishesAllRootsOnly(t *testing.T) {
	t.Parallel()
	fixture := schemaLoaderFixture(t)
	fixture["schemas/custom.cue"] = &fstest.MapFile{Data: []byte("package schemas\n#Custom: string\n")}
	v := NewValidator()
	if err := v.loadSchemas(fixture); err != nil {
		t.Fatal(err)
	}
	if len(v.schemas) != len(fixture)-1 {
		t.Fatalf("loaded %d roots from %d files", len(v.schemas), len(fixture))
	}
	for _, name := range []string{"agent", "command", "skill", "settings", "claude_md", "custom"} {
		if _, ok := v.schemas[name]; !ok {
			t.Errorf("missing root %s", name)
		}
	}
	if _, ok := v.schemas["component_hooks"]; ok {
		t.Error("shared definitions were published as a root")
	}
	delete(fixture, "schemas/custom.cue")
	if err := v.loadSchemas(fixture); err != nil {
		t.Fatal(err)
	}
	if _, ok := v.schemas["custom"]; ok {
		t.Error("successful reload retained a removed root")
	}
}

func TestSchemaLoaderFailuresPreservePublishedMap(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name   string
		denied string
		broken string
		want   []string
	}{
		{"directory read", "schemas", "", []string{"schemas"}},
		{"root read", "schemas/agent.cue", "", []string{"schemas/agent.cue"}},
		{"root compile", "", "schemas/agent.cue", []string{"schemas/agent.cue"}},
		{"fragment read", "schemas/component_hooks.cue", "", []string{"schemas/component_hooks.cue"}},
		{"fragment compile", "", "schemas/component_hooks.cue", []string{"compile CUE schema", "component_hooks.cue"}},
		{"joined failures", "schemas/agent.cue", "schemas/command.cue", []string{"schemas/agent.cue", "schemas/command.cue"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			for _, reload := range []bool{false, true} {
				name := "initial"
				if reload {
					name = "reload"
				}
				t.Run(name, func(t *testing.T) {
					t.Parallel()
					v := NewValidator()
					if reload {
						if err := v.LoadSchemas(); err != nil {
							t.Fatal(err)
						}
					}
					before := maps.Clone(v.schemas)
					fixture := schemaLoaderFixture(t)
					if tc.broken != "" {
						fixture[tc.broken] = &fstest.MapFile{Data: []byte("package schemas\nbroken: [")}
					}
					err := v.loadSchemas(deniedSchemaFS{FS: fixture, denied: tc.denied})
					if err == nil {
						t.Fatal("load succeeded despite failed input")
					}
					for _, text := range tc.want {
						if !strings.Contains(err.Error(), text) {
							t.Errorf("error %q does not identify %s", err, text)
						}
					}
					if tc.denied != "" && !errors.Is(err, fs.ErrPermission) {
						t.Errorf("wrapped permission cause lost: %v", err)
					}
					if !maps.Equal(v.schemas, before) {
						t.Error("failed load changed the published schema map")
					}
				})
			}
		})
	}
}

func TestSchemaLoaderEmptyRootsPreservePublishedMap(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	before := maps.Clone(v.schemas)
	fixture := schemaLoaderFixture(t)
	for name := range fixture {
		if name != "schemas/component_hooks.cue" {
			delete(fixture, name)
		}
	}
	err := v.loadSchemas(fixture)
	if err == nil || !strings.Contains(err.Error(), "no CUE schemas loaded") {
		t.Fatalf("empty roots error = %v", err)
	}
	if !maps.Equal(v.schemas, before) {
		t.Error("empty reload replaced the previously loaded map")
	}
}

func TestComponentHookAliasesRemainEquivalent(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name  string
		hook  map[string]any
		valid bool
	}{
		{"shell", map[string]any{"type": "command", "command": "echo ok", "timeout": float64(3)}, true},
		{"exec", map[string]any{"type": "command", "args": []any{"echo", "ok"}, "onFailure": "block"}, true},
		{"wrong type", map[string]any{"type": "prompt", "prompt": "ok"}, false},
		{"bad failure policy", map[string]any{"type": "command", "command": "echo ok", "onFailure": "retry"}, false},
		{"bad timeout", map[string]any{"type": "command", "command": "echo ok", "timeout": "3"}, false},
		{"settings-only field", map[string]any{"type": "command", "command": "echo ok", "async": true}, false},
	}
	for _, kind := range []string{"agent", "command", "skill"} {
		t.Run(kind, func(t *testing.T) {
			t.Parallel()
			v := NewValidator()
			if err := v.LoadSchemas(); err != nil {
				t.Fatal(err)
			}
			for _, suffix := range []string{"HookCommand", "Hook", "Hooks"} {
				alias := "#" + strings.ToUpper(kind[:1]) + kind[1:] + suffix
				if !v.schemas[kind].LookupPath(cue.ParsePath(alias)).Exists() {
					t.Errorf("missing alias %s", alias)
				}
			}
			for _, tc := range cases {
				t.Run(tc.name, func(t *testing.T) {
					t.Parallel()
					data := map[string]any{"name": "sample", "description": "Sample.", "hooks": map[string]any{
						"PreToolUse": []any{map[string]any{"hooks": []any{tc.hook}}},
					}}
					diagnostics, err := v.validateSchema(kind, data)
					if err != nil {
						t.Fatal(err)
					}
					if valid := len(diagnostics) == 0; valid != tc.valid {
						t.Errorf("valid = %v, want %v: %v", valid, tc.valid, diagnostics)
					}
				})
			}
			t.Run("non-array handlers", func(t *testing.T) {
				t.Parallel()
				data := map[string]any{"name": "sample", "description": "Sample.", "hooks": map[string]any{
					"Stop": []any{map[string]any{"hooks": "not-an-array"}},
				}}
				diagnostics, err := v.validateSchema(kind, data)
				if err != nil || len(diagnostics) == 0 {
					t.Fatalf("non-array hooks diagnostics = %v, error = %v", diagnostics, err)
				}
			})
		})
	}
}

func TestSettingsHooksRemainDistinct(t *testing.T) {
	t.Parallel()
	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	data := map[string]any{"hooks": map[string]any{
		"PreToolUse": []any{map[string]any{"hooks": []any{
			map[string]any{"type": "prompt", "prompt": "Check.", "model": "haiku"},
			map[string]any{"type": "command", "command": "echo ok", "async": true, "statusMessage": "Checking"},
		}}},
	}}
	diagnostics, err := v.ValidateSettings(data)
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("richer settings hooks rejected: %v, %v", diagnostics, err)
	}
}
