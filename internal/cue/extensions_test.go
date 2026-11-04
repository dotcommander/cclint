package cue

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadExtensionsEnforcesExtraConstraint(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	ext := filepath.Join(root, "agent-ext.cue")
	// Embedded schema allows any positive maxTurns; the extension tightens it.
	if err := os.WriteFile(ext, []byte("#Agent: {maxTurns: int & >100}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	if err := v.LoadExtensions(root, map[string]any{"agent": "agent-ext.cue"}); err != nil {
		t.Fatalf("LoadExtensions: %v", err)
	}

	okData := map[string]any{"name": "x", "description": "d", "maxTurns": float64(200)}
	if errs, err := v.ValidateAgent(okData); err != nil || len(errs) != 0 {
		t.Errorf("compliant data rejected: errs=%v err=%v", errs, err)
	}
	badData := map[string]any{"name": "x", "description": "d", "maxTurns": float64(50)}
	errs, err := v.ValidateAgent(badData)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) == 0 {
		t.Error("violating data accepted despite extension constraint maxTurns > 100")
	} else if !strings.Contains(errs[0].Message, "maxTurns") {
		t.Errorf("error does not name maxTurns: %s", errs[0].Message)
	}
}

func TestLoadExtensionsListValue(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	first := filepath.Join(root, "first.cue")
	second := filepath.Join(root, "second.cue")
	if err := os.WriteFile(first, []byte("#Skill: {maxTurns: int & >10}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(second, []byte("#Skill: {maxTurns: int & <=100}\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	v := NewValidator()
	if err := v.LoadSchemas(); err != nil {
		t.Fatal(err)
	}
	if err := v.LoadExtensions(root, map[string]any{"skill": []string{"first.cue", "second.cue"}}); err != nil {
		t.Fatalf("LoadExtensions list value: %v", err)
	}
	bad := map[string]any{"name": "s", "description": "d", "maxTurns": float64(5)}
	errs, err := v.ValidateSkill(bad)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) == 0 {
		t.Error("list-form extensions were not applied")
	}
}

func TestLoadExtensionsErrors(t *testing.T) {
	t.Parallel()
	root := t.TempDir()

	t.Run("unknown key", func(t *testing.T) {
		t.Parallel()
		v := NewValidator()
		if err := v.LoadSchemas(); err != nil {
			t.Fatal(err)
		}
		err := v.LoadExtensions(root, map[string]any{"workflow": "x.cue"})
		if err == nil || !strings.Contains(err.Error(), "workflow") {
			t.Errorf("unknown key error = %v, want error naming the key", err)
		}
	})

	t.Run("missing file", func(t *testing.T) {
		t.Parallel()
		v := NewValidator()
		if err := v.LoadSchemas(); err != nil {
			t.Fatal(err)
		}
		err := v.LoadExtensions(root, map[string]any{"agent": "does-not-exist.cue"})
		if err == nil || !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), "does-not-exist.cue") {
			t.Errorf("missing file error = %v, want wrapped os.ErrNotExist naming the file", err)
		}
	})

	t.Run("invalid CUE", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		bad := filepath.Join(dir, "broken.cue")
		if err := os.WriteFile(bad, []byte("#Agent: {maxTurns: ???not cue}\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		v := NewValidator()
		if err := v.LoadSchemas(); err != nil {
			t.Fatal(err)
		}
		err := v.LoadExtensions(dir, map[string]any{"agent": bad})
		if err == nil || !strings.Contains(err.Error(), "broken.cue") {
			t.Errorf("invalid CUE error = %v, want error naming the file", err)
		}
	})

	t.Run("non-string value", func(t *testing.T) {
		t.Parallel()
		v := NewValidator()
		if err := v.LoadSchemas(); err != nil {
			t.Fatal(err)
		}
		err := v.LoadExtensions(root, map[string]any{"agent": 42})
		if err == nil || !strings.Contains(err.Error(), "agent") {
			t.Errorf("non-string value error = %v, want error naming the key", err)
		}
	})

	t.Run("non-string list element", func(t *testing.T) {
		t.Parallel()
		v := NewValidator()
		if err := v.LoadSchemas(); err != nil {
			t.Fatal(err)
		}
		err := v.LoadExtensions(root, map[string]any{"agent": []any{"ok.cue", true}})
		if err == nil || !strings.Contains(err.Error(), "agent") {
			t.Errorf("non-string list element error = %v, want error naming the key", err)
		}
	})
}
