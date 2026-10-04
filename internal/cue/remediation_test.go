package cue

import (
	"strings"
	"testing"
)

func TestContextSchemaDefinitionAndInvalidMetadata(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	for _, data := range []map[string]any{{"model": "invalid-model"}, {"title": 17}, {"sections": []any{map[string]any{"heading": 12}}}} {
		diagnostics, err := v.ValidateClaudeMD(data)
		if err != nil || len(diagnostics) == 0 {
			t.Fatalf("invalid context metadata %#v: diagnostics=%v, err=%v", data, diagnostics, err)
		}
	}
	v.schemas["claude_md"] = v.ctx.CompileString("#Wrong: {...}")
	if _, err := v.ValidateClaudeMD(map[string]any{}); err == nil || !strings.Contains(err.Error(), "#ClaudeMD") {
		t.Fatalf("missing expected definition must fail visibly: %v", err)
	}
}

func TestSchemaDiagnosticsHaveDocumentIdentity(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	for _, file := range []string{"agents/one.md", "agents/two.md"} {
		diagnostics, err := v.ValidateFile(file, "---\nname: 123\ndescription: test\n---\nbody", "agent")
		if err != nil || len(diagnostics) == 0 {
			t.Fatalf("expected schema diagnostics: %v, %v", diagnostics, err)
		}
		for _, diagnostic := range diagnostics {
			if diagnostic.File != file || diagnostic.Line != 0 || diagnostic.Column != 0 {
				t.Errorf("unexpected document identity: %#v", diagnostic)
			}
			if strings.Contains(diagnostic.Message, "name: name:") {
				t.Errorf("duplicate field path: %s", diagnostic.Message)
			}
		}
	}
}

func TestSettingsLifecycleAndObjectMatchers(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name    string
		matcher any
		omit    bool
		invalid bool
	}{
		{name: "lifecycle omission", omit: true},
		{name: "string", matcher: "Bash"},
		{name: "tool object", matcher: map[string]any{"toolName": "Bash"}},
		{name: "empty object", matcher: map[string]any{}},
		{name: "number", matcher: 42, invalid: true},
		{name: "malformed tool name", matcher: map[string]any{"toolName": 42}, invalid: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry := map[string]any{"hooks": []any{map[string]any{"type": "command", "command": "echo hello"}}}
			if !tc.omit {
				entry["matcher"] = tc.matcher
			}
			diagnostics, err := v.ValidateSettings(map[string]any{"hooks": map[string]any{"SessionStart": []any{entry}}})
			if err != nil || (len(diagnostics) > 0) != tc.invalid {
				t.Fatalf("diagnostics=%v, err=%v; invalid=%v", diagnostics, err, tc.invalid)
			}
		})
	}
}

func TestAgentSkillsStringOrList(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		value   any
		invalid bool
	}{
		{"one, two", false}, {[]string{"one", "two"}, false}, {[]any{"one", 42}, true}, {42, true},
	} {
		diagnostics, err := v.ValidateAgent(map[string]any{"name": "tester", "description": "Tests things", "skills": tc.value})
		if err != nil || (len(diagnostics) > 0) != tc.invalid {
			t.Fatalf("skills=%#v diagnostics=%v err=%v", tc.value, diagnostics, err)
		}
	}
}

func TestMarketplaceWrapperFields(t *testing.T) {
	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"extraKnownMarketplaces", "additionalMarketplaces"} {
		for _, tc := range []struct {
			location any
			update   any
			invalid  bool
		}{
			{"/tmp/plugins", true, false}, {42, true, true}, {"/tmp/plugins", "yes", true},
		} {
			data := map[string]any{key: map[string]any{"example": map[string]any{"source": map[string]any{"source": "github", "repo": "example/plugins"}, "installLocation": tc.location, "autoUpdate": tc.update}}}
			diagnostics, err := v.ValidateSettings(data)
			if err != nil || (len(diagnostics) > 0) != tc.invalid {
				t.Fatalf("%s diagnostics=%v err=%v", key, diagnostics, err)
			}
		}
	}
}
