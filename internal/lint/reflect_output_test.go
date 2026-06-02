package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckReflectOutput(t *testing.T) {
	goodBody := "# Race condition in channel close\n\n" +
		"line two\nline three\nline four\nline five\nline six\n" +
		"line seven\nline eight\nline nine\n\n" +
		"(source: https://example.com/post)\n"

	tests := []struct {
		name         string
		fileName     string
		content      string
		wantWarnings int
	}{
		{"good slug and source", "race-condition-in-channel-close.md", goodBody, 0},
		{"bad slug too few words", "oops.md", goodBody, 1},
		{"bad slug uppercase", "Race-Condition-In-Close.md", goodBody, 1},
		{"missing source", "a-valid-four-word-slug.md", strings.Replace(goodBody, "(source: https://example.com/post)\n", "no attribution here\n", 1), 1},
		{"missing h1", "another-valid-four-word.md", strings.Replace(goodBody, "# Race condition in channel close\n", "Race condition (no H1)\n", 1), 1},
		{"too short body", "short-fold-candidate-entry.md", "# Title\n\n(source: x)\n", 1},
		{"too long body", "long-split-candidate-entry.md", "# Title\n\n(source: x)\n" + strings.Repeat("body line\n", 510), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			kbDir := filepath.Join(tmpDir, "kb")
			if err := os.MkdirAll(kbDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(kbDir, tt.fileName), []byte(tt.content), 0644); err != nil {
				t.Fatal(err)
			}
			errors := CheckReflectOutput(tmpDir)
			warnCount := 0
			for _, e := range errors {
				if e.Severity == "warning" {
					warnCount++
				}
			}
			if warnCount != tt.wantWarnings {
				t.Errorf("CheckReflectOutput() warnings = %d, want %d", warnCount, tt.wantWarnings)
				for _, e := range errors {
					t.Logf("  %s: %s — %s", e.Severity, e.File, e.Message)
				}
			}
		})
	}
}

func TestCheckReflectOutput_SkipNonEntryFiles(t *testing.T) {
	t.Parallel()
	// SKILL.md (and README.md etc.) must be exempt from slug/source/size rules
	// even when their content would otherwise trigger multiple warnings.
	badContent := "not a slug name, no h1, no source, too short\n"
	skipNames := []string{"SKILL.md", "README.md", "skill.md", "readme.md", "index.md", "MEMORY.md"}
	for _, name := range skipNames {
		name := name
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			kbDir := filepath.Join(tmpDir, "kb")
			if err := os.MkdirAll(kbDir, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(kbDir, name), []byte(badContent), 0644); err != nil {
				t.Fatal(err)
			}
			errors := CheckReflectOutput(tmpDir)
			warnCount := 0
			for _, e := range errors {
				if e.Severity == "warning" {
					warnCount++
				}
			}
			if warnCount != 0 {
				t.Errorf("CheckReflectOutput() for %s: got %d warnings, want 0", name, warnCount)
				for _, e := range errors {
					t.Logf("  %s: %s — %s", e.Severity, e.File, e.Message)
				}
			}
		})
	}
}

func TestCheckReflectOutput_NoKBDir(t *testing.T) {
	tmpDir := t.TempDir()
	if errors := CheckReflectOutput(tmpDir); len(errors) != 0 {
		t.Errorf("CheckReflectOutput() with no kb/ dir = %d errors, want 0", len(errors))
	}
}

func TestCheckReflectOutput_Recursive(t *testing.T) {
	tmpDir := t.TempDir()
	nested := filepath.Join(tmpDir, "kb", "go")
	if err := os.MkdirAll(nested, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(nested, "bad.md"), []byte("# T\n(source: x)\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if errors := CheckReflectOutput(tmpDir); len(errors) == 0 {
		t.Errorf("CheckReflectOutput() should flag nested kb/go/bad.md, got 0 errors")
	}
}
