package lint

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidatePathsGlob(t *testing.T) {
	tests := []struct {
		name         string
		paths        any
		wantErrCount int
	}{
		{
			name:         "valid pattern",
			paths:        "**/*.go",
			wantErrCount: 0,
		},
		{
			name:         "multiple patterns",
			paths:        "**/*.{ts,tsx}, src/**/*.js",
			wantErrCount: 0,
		},
		{
			name:         "invalid type",
			paths:        123,
			wantErrCount: 1,
		},
		{
			name:         "unbalanced braces",
			paths:        "**/*.{ts,tsx",
			wantErrCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validatePathsGlob(tt.paths, "test.md", "")

			if len(errors) != tt.wantErrCount {
				t.Errorf("validatePathsGlob() errors = %d, want %d", len(errors), tt.wantErrCount)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestSplitPathPatterns(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{
			name: "single pattern",
			s:    "**/*.go",
			want: []string{"**/*.go"},
		},
		{
			name: "multiple patterns",
			s:    "**/*.go, **/*.ts",
			want: []string{"**/*.go", " **/*.ts"},
		},
		{
			name: "brace expansion",
			s:    "**/*.{ts,tsx}",
			want: []string{"**/*.{ts,tsx}"},
		},
		{
			name: "mixed",
			s:    "**/*.{ts,tsx}, src/**/*.js, lib/**/*.go",
			want: []string{"**/*.{ts,tsx}", " src/**/*.js", " lib/**/*.go"},
		},
		{
			name: "nested braces",
			s:    "**/{foo,{bar,baz}}",
			want: []string{"**/{foo,{bar,baz}}"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitPathPatterns(tt.s)

			if len(got) != len(tt.want) {
				t.Errorf("splitPathPatterns() returned %d patterns, want %d", len(got), len(tt.want))
				t.Logf("  Got: %v", got)
				t.Logf("  Want: %v", tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitPathPatterns()[%d] = %q, want %q", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestValidateGlobPattern(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"valid pattern", "**/*.go", false},
		{"valid with braces", "**/*.{ts,tsx}", false},
		{"unbalanced braces open", "**/*.{ts,tsx", true},
		{"unbalanced braces close", "**/*.ts,tsx}", true},
		{"valid nested", "**/{foo,bar}/*.go", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateGlobPattern(tt.pattern)
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("validateGlobPattern(%q) hasErr = %v, want %v", tt.pattern, hasErr, tt.wantErr)
				if err != nil {
					t.Logf("  Error: %v", err)
				}
			}
		})
	}
}

func TestValidateBalancedBraces(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		wantErr bool
	}{
		{"balanced", "{foo,bar}", false},
		{"nested balanced", "{foo,{bar,baz}}", false},
		{"no braces", "foobar", false},
		{"unclosed", "{foo,bar", true},
		{"unopened", "foo,bar}", true},
		{"multiple unclosed", "{{foo", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateBalancedBraces(tt.pattern)
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("validateBalancedBraces(%q) hasErr = %v, want %v", tt.pattern, hasErr, tt.wantErr)
			}
		})
	}
}

func TestValidateImports(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a test import target
	importTarget := filepath.Join(tmpDir, "imported.md")
	if err := os.WriteFile(importTarget, []byte("imported content"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		contents     string
		filePath     string
		wantWarnings int
	}{
		{
			name:         "no imports",
			contents:     "Just regular content",
			filePath:     filepath.Join(tmpDir, "test.md"),
			wantWarnings: 0,
		},
		{
			name:         "valid import",
			contents:     "@./imported.md",
			filePath:     filepath.Join(tmpDir, "test.md"),
			wantWarnings: 0,
		},
		{
			name:         "nonexistent import",
			contents:     "@./nonexistent.md",
			filePath:     filepath.Join(tmpDir, "test.md"),
			wantWarnings: 1,
		},
		{
			name:         "import in code block",
			contents:     "```\n@./imported.md\n```",
			filePath:     filepath.Join(tmpDir, "test.md"),
			wantWarnings: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateImports(tt.contents, tt.filePath)

			warnCount := 0
			for _, e := range errors {
				if e.Severity == "warning" {
					warnCount++
				}
			}

			if warnCount != tt.wantWarnings {
				t.Errorf("validateImports() warnings = %d, want %d", warnCount, tt.wantWarnings)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestRuleLinterType(t *testing.T) {
	linter := NewRuleLinter()
	if linter.Type() != "rule" {
		t.Errorf("RuleLinter.Type() = %q, want %q", linter.Type(), "rule")
	}
}

func TestRuleLinterFileType(t *testing.T) {
	linter := NewRuleLinter()
	expected := "rule"
	got := linter.Type()
	if got != expected {
		t.Errorf("RuleLinter.Type() = %q, want %q", got, expected)
	}
}

func TestRuleLinterPreValidate(t *testing.T) {
	linter := NewRuleLinter()

	tests := []struct {
		name         string
		filePath     string
		contents     string
		wantErrCount int
	}{
		{
			name:         "valid md file",
			filePath:     "rules/test.md",
			contents:     "content",
			wantErrCount: 0,
		},
		{
			name:         "non-md extension",
			filePath:     "rules/test.txt",
			contents:     "content",
			wantErrCount: 1,
		},
		{
			name:         "empty file",
			filePath:     "rules/test.md",
			contents:     "   \n  \n",
			wantErrCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := linter.PreValidate(tt.filePath, tt.contents)

			errCount := 0
			for _, e := range errors {
				if e.Severity == "error" {
					errCount++
				}
			}

			if errCount != tt.wantErrCount {
				t.Errorf("PreValidate() errors = %d, want %d", errCount, tt.wantErrCount)
			}
		})
	}
}

func TestRuleLinterParseContent(t *testing.T) {
	linter := NewRuleLinter()

	tests := []struct {
		name     string
		contents string
		wantData bool
	}{
		{
			name:     "with paths frontmatter",
			contents: "---\npaths: \"**/*.go\"\n---\nContent",
			wantData: true,
		},
		{
			name:     "without frontmatter",
			contents: "Just content",
			wantData: true,
		},
		{
			name:     "with other frontmatter",
			contents: "---\nname: test\n---\nContent",
			wantData: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, body, err := linter.ParseContent(tt.contents)

			if err != nil {
				t.Errorf("ParseContent() unexpected error: %v", err)
				return
			}

			if tt.wantData && data == nil {
				t.Error("ParseContent() data is nil")
			}

			if !strings.Contains(tt.contents, "---") && body != tt.contents {
				t.Error("ParseContent() body should be full content when no frontmatter")
			}
		})
	}
}

func TestRuleLinterValidateCUE(t *testing.T) {
	linter := NewRuleLinter()
	errors, err := linter.ValidateCUE(nil, nil)

	if err != nil {
		t.Errorf("ValidateCUE() unexpected error: %v", err)
	}

	if errors != nil {
		t.Error("ValidateCUE() should return nil for rules (no CUE schema)")
	}
}

func TestRuleLinterValidateSpecific(t *testing.T) {
	linter := NewRuleLinter()

	tests := []struct {
		name         string
		data         map[string]any
		wantErrCount int
	}{
		{
			name:         "no paths field",
			data:         map[string]any{},
			wantErrCount: 0,
		},
		{
			name: "valid paths",
			data: map[string]any{
				"paths": "**/*.go",
			},
			wantErrCount: 0,
		},
		{
			name: "invalid paths type",
			data: map[string]any{
				"paths": 123,
			},
			wantErrCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := linter.ValidateSpecific(tt.data, "test.md", "")

			if len(errors) != tt.wantErrCount {
				t.Errorf("ValidateSpecific() errors = %d, want %d", len(errors), tt.wantErrCount)
			}
		})
	}
}

func TestRuleLinterPreValidateSymlinks(t *testing.T) {
	linter := NewRuleLinter()
	tmpDir := t.TempDir()

	// Create a target file
	targetFile := filepath.Join(tmpDir, "target.md")
	if err := os.WriteFile(targetFile, []byte("target content"), 0644); err != nil {
		t.Fatal(err)
	}

	// Create a symlink to the target
	symlinkFile := filepath.Join(tmpDir, "symlink.md")
	if err := os.Symlink(targetFile, symlinkFile); err != nil {
		t.Skip("Symlink creation not supported on this platform")
	}

	// Create a broken symlink
	brokenSymlink := filepath.Join(tmpDir, "broken.md")
	if err := os.Symlink("/nonexistent/path", brokenSymlink); err != nil {
		t.Skip("Symlink creation not supported on this platform")
	}

	tests := []struct {
		name         string
		filePath     string
		contents     string
		wantWarnings int
	}{
		{
			name:         "valid symlink",
			filePath:     symlinkFile,
			contents:     "content",
			wantWarnings: 0,
		},
		{
			name:         "broken symlink",
			filePath:     brokenSymlink,
			contents:     "content",
			wantWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := linter.PreValidate(tt.filePath, tt.contents)

			warnCount := 0
			for _, e := range errors {
				if e.Severity == "warning" {
					warnCount++
				}
			}

			if warnCount < tt.wantWarnings {
				t.Errorf("PreValidate() warnings = %d, want at least %d", warnCount, tt.wantWarnings)
			}
		})
	}
}

func TestValidateImportsEdgeCases(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	testFile := filepath.Join(tmpDir, "test.md")
	importedFile := filepath.Join(tmpDir, "imported.md")
	if err := os.WriteFile(importedFile, []byte("imported"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		contents     string
		wantWarnings int
	}{
		{
			name:         "import with tilde",
			contents:     "@~/nonexistent.md",
			wantWarnings: 1,
		},
		{
			name:         "multiple imports",
			contents:     "@./imported.md and @./nonexistent.md",
			wantWarnings: 1, // Only nonexistent should warn
		},
		{
			name: "import in multiline code block",
			contents: "```go\n" +
				"// @./imported.md\n" +
				"```",
			wantWarnings: 0, // Should be ignored in code block
		},
		{
			name:         "possible self-import",
			contents:     "@./test.md",
			wantWarnings: 1, // File does not exist on disk (circular detection at batch level)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validateImports(tt.contents, testFile)

			warnCount := 0
			for _, e := range errors {
				if e.Severity == "warning" {
					warnCount++
				}
			}

			if warnCount < tt.wantWarnings {
				t.Errorf("validateImports() warnings = %d, want at least %d", warnCount, tt.wantWarnings)
				for _, e := range errors {
					t.Logf("  %s (line %d): %s", e.Severity, e.Line, e.Message)
				}
			}
		})
	}
}

func TestRuleLinterValidateBestPractices(t *testing.T) {
	linter := NewRuleLinter()
	tmpDir := t.TempDir()

	// Create a valid import target
	importFile := filepath.Join(tmpDir, "import.md")
	if err := os.WriteFile(importFile, []byte("imported"), 0644); err != nil {
		t.Fatal(err)
	}

	testFile := filepath.Join(tmpDir, "test.md")

	tests := []struct {
		name         string
		contents     string
		wantWarnings int
	}{
		{
			name:         "no imports",
			contents:     "Just content",
			wantWarnings: 0,
		},
		{
			name:         "valid import",
			contents:     "@./import.md",
			wantWarnings: 0,
		},
		{
			name:         "invalid import",
			contents:     "@./nonexistent.md",
			wantWarnings: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := linter.ValidateBestPractices(testFile, tt.contents, nil)

			warnCount := 0
			for _, e := range errors {
				if e.Severity == "warning" {
					warnCount++
				}
			}

			if warnCount < tt.wantWarnings {
				t.Errorf("ValidateBestPractices() warnings = %d, want at least %d", warnCount, tt.wantWarnings)
			}
		})
	}
}
