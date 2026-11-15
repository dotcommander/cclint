package cue

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtensionDefinitionForEveryComponent(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		component  string
		definition string
	}{
		{"agent", "#Agent"},
		{"command", "#Command"},
		{"skill", "#Skill"},
		{"settings", "#Settings"},
		{"claude_md", "#ClaudeMD"},
	} {
		t.Run(tc.component, func(t *testing.T) {
			t.Parallel()
			if got, ok := schemaDefinition(tc.component); !ok || got != tc.definition {
				t.Fatalf("schemaDefinition(%q) = %q, %t; want %q", tc.component, got, ok, tc.definition)
			}

			root := t.TempDir()
			filename := filepath.Join(root, "extension.cue")
			content := fmt.Sprintf("%s: {r006Guard?: \"allowed\"}\n", tc.definition)
			if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
				t.Fatal(err)
			}
			v := NewValidator()
			if err := v.LoadSchemas(); err != nil {
				t.Fatal(err)
			}
			if err := v.LoadExtensions(root, map[string]any{tc.component: filename}); err != nil {
				t.Fatalf("LoadExtensions(%q): %v", tc.component, err)
			}
			data := map[string]any{"r006Guard": "denied"}
			if tc.component == "agent" || tc.component == "skill" {
				data["name"] = "valid-name"
				data["description"] = "valid description"
			}
			errs, err := v.validateSchema(tc.component, data)
			if err != nil {
				t.Fatal(err)
			}
			if len(errs) == 0 {
				t.Errorf("%s extension did not constrain validation", tc.definition)
			}
		})
	}
	if got, ok := schemaDefinition("unknown"); ok || got != "" {
		t.Errorf("unknown component mapped to %q, %t", got, ok)
	}
}

func TestLoadExtensionsRequiresDefinitionInEachFile(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name    string
		content string
	}{
		{"empty", ""},
		{"helper only", "#Helper: {maxTurns: int & >100}\n"},
		{"misspelled", "#Agnet: {maxTurns: int & >100}\n"},
		{"wrong root", "#Command: {maxTurns: int & >100}\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			root := t.TempDir()
			filename := filepath.Join(root, "invalid.cue")
			if err := os.WriteFile(filename, []byte(tc.content), 0o644); err != nil {
				t.Fatal(err)
			}
			v := NewValidator()
			if err := v.LoadSchemas(); err != nil {
				t.Fatal(err)
			}
			err := v.LoadExtensions(root, map[string]any{"agent": filename})
			if err == nil || !strings.Contains(err.Error(), "agent") ||
				!strings.Contains(err.Error(), filename) || !strings.Contains(err.Error(), "#Agent") {
				t.Errorf("LoadExtensions = %v, want component, file, and expected definition", err)
			}
		})
	}
}

func TestLoadExtensionsCannotBorrowDefinitionFromEarlierFile(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	first := filepath.Join(root, "first.cue")
	second := filepath.Join(root, "second.cue")
	if err := os.WriteFile(first, []byte("#Agent: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("#Helper: {}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	err := v.LoadExtensions(root, map[string]any{"agent": []string{first, second}})
	if err == nil || !strings.Contains(err.Error(), "agent") ||
		!strings.Contains(err.Error(), second) || !strings.Contains(err.Error(), "#Agent") {
		t.Errorf("LoadExtensions = %v, want second file's missing #Agent error", err)
	}
}

func TestLoadExtensionsAllowsAndEnforcesHelpers(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	filename := filepath.Join(root, "helpers.cue")
	content := "#Limit: {maxTurns?: int & >100, ...}\n#Agent: #Limit\n#Unused: string\n"
	if err := os.WriteFile(filename, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	if err := v.LoadExtensions(root, map[string]any{"agent": filename}); err != nil {
		t.Fatal(err)
	}
	data := map[string]any{"name": "valid-name", "description": "valid description", "maxTurns": 50}
	errs, err := v.ValidateAgent(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) == 0 {
		t.Error("helper constraint did not reject maxTurns of 50")
	}
	data["maxTurns"] = 200
	errs, err = v.ValidateAgent(data)
	if err != nil || len(errs) != 0 {
		t.Errorf("valid agent rejected: errs=%v err=%v", errs, err)
	}
}
