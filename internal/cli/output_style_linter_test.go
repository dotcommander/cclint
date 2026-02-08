package cli

import (
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

func TestOutputStyleLinterType(t *testing.T) {
	linter := NewOutputStyleLinter()
	if linter.Type() != "output-style" {
		t.Errorf("Type() = %q, want %q", linter.Type(), "output-style")
	}
	if linter.FileType() != discovery.FileTypeOutputStyle {
		t.Errorf("FileType() = %v, want %v", linter.FileType(), discovery.FileTypeOutputStyle)
	}
}

func TestOutputStyleLinterValidOutput(t *testing.T) {
	linter := NewOutputStyleLinter()
	contents := "---\nname: concise-technical\ndescription: A concise technical writing style\nkeep-coding-instructions: true\n---\n\nWrite in a concise, technical style.\n"

	data, body, err := linter.ParseContent(contents)
	if err != nil {
		t.Fatalf("ParseContent() error: %v", err)
	}

	preErrors := linter.PreValidate("output-styles/concise-technical.md", contents)
	if len(preErrors) != 0 {
		t.Errorf("PreValidate() returned %d errors, want 0: %v", len(preErrors), preErrors)
	}

	specificErrors := linter.ValidateSpecific(data, "output-styles/concise-technical.md", contents)
	for _, e := range specificErrors {
		if e.Severity == "error" {
			t.Errorf("ValidateSpecific() returned unexpected error: %s", e.Message)
		}
	}

	score := linter.Score(contents, data, body)
	if score == nil {
		t.Fatal("Score() returned nil")
	}
	if score.Overall == 0 {
		t.Error("Score() returned 0 overall for valid output style")
	}
}

func TestOutputStyleLinterMissingName(t *testing.T) {
	linter := NewOutputStyleLinter()
	contents := "---\ndescription: A style without a name\n---\n\nSome body content.\n"

	data, _, err := linter.ParseContent(contents)
	if err != nil {
		t.Fatalf("ParseContent() error: %v", err)
	}

	errors := linter.ValidateSpecific(data, "output-styles/no-name.md", contents)
	foundNameError := false
	for _, e := range errors {
		if e.Severity == "error" && containsSubstring(e.Message, "name") {
			foundNameError = true
			break
		}
	}
	if !foundNameError {
		t.Error("ValidateSpecific() should report missing name as error")
	}
}

func TestOutputStyleLinterMissingDescription(t *testing.T) {
	linter := NewOutputStyleLinter()
	contents := "---\nname: test-style\n---\n\nSome body content.\n"

	data, _, err := linter.ParseContent(contents)
	if err != nil {
		t.Fatalf("ParseContent() error: %v", err)
	}

	errors := linter.ValidateSpecific(data, "output-styles/test-style.md", contents)
	foundDescError := false
	for _, e := range errors {
		if e.Severity == "error" && containsSubstring(e.Message, "description") {
			foundDescError = true
			break
		}
	}
	if !foundDescError {
		t.Error("ValidateSpecific() should report missing description as error")
	}
}

func TestOutputStyleLinterNonBooleanKeepCodingInstructions(t *testing.T) {
	linter := NewOutputStyleLinter()
	contents := "---\nname: test-style\ndescription: A test style\nkeep-coding-instructions: \"yes\"\n---\n\nSome body content.\n"

	data, _, err := linter.ParseContent(contents)
	if err != nil {
		t.Fatalf("ParseContent() error: %v", err)
	}

	errors := linter.ValidateSpecific(data, "output-styles/test-style.md", contents)
	foundBoolError := false
	for _, e := range errors {
		if e.Severity == "error" && containsSubstring(e.Message, "keep-coding-instructions") && containsSubstring(e.Message, "boolean") {
			foundBoolError = true
			break
		}
	}
	if !foundBoolError {
		t.Errorf("ValidateSpecific() should report non-boolean keep-coding-instructions as error, got: %v", errors)
	}
}

func TestOutputStyleLinterEmptyBody(t *testing.T) {
	linter := NewOutputStyleLinter()
	contents := "---\nname: test-style\ndescription: A test style\n---\n"

	data, _, err := linter.ParseContent(contents)
	if err != nil {
		t.Fatalf("ParseContent() error: %v", err)
	}

	errors := linter.ValidateSpecific(data, "output-styles/test-style.md", contents)
	foundBodyError := false
	for _, e := range errors {
		if e.Severity == "error" && containsSubstring(e.Message, "body") {
			foundBodyError = true
			break
		}
	}
	if !foundBodyError {
		t.Error("ValidateSpecific() should report empty body as error")
	}
}

func TestOutputStyleLinterInvalidNameFormat(t *testing.T) {
	linter := NewOutputStyleLinter()
	contents := "---\nname: My Style\ndescription: A test style\n---\n\nBody content.\n"

	data, _, err := linter.ParseContent(contents)
	if err != nil {
		t.Fatalf("ParseContent() error: %v", err)
	}

	errors := linter.ValidateSpecific(data, "output-styles/my-style.md", contents)
	foundFormatError := false
	for _, e := range errors {
		if e.Severity == "error" && containsSubstring(e.Message, "kebab-case") {
			foundFormatError = true
			break
		}
	}
	if !foundFormatError {
		t.Error("ValidateSpecific() should report invalid name format as error")
	}
}

func TestOutputStyleLinterUnknownField(t *testing.T) {
	linter := NewOutputStyleLinter()
	contents := "---\nname: test-style\ndescription: A test style\nunknown-field: value\n---\n\nBody content.\n"

	data, _, err := linter.ParseContent(contents)
	if err != nil {
		t.Fatalf("ParseContent() error: %v", err)
	}

	errors := linter.ValidateSpecific(data, "output-styles/test-style.md", contents)
	foundSuggestion := false
	for _, e := range errors {
		if e.Severity == "suggestion" && containsSubstring(e.Message, "unknown-field") {
			foundSuggestion = true
			break
		}
	}
	if !foundSuggestion {
		t.Error("ValidateSpecific() should suggest removing unknown frontmatter fields")
	}
}

func TestOutputStyleLinterNameStartsWithHyphen(t *testing.T) {
	linter := NewOutputStyleLinter()
	contents := "---\nname: -bad-name\ndescription: A test style\n---\n\nBody content.\n"

	data, _, err := linter.ParseContent(contents)
	if err != nil {
		t.Fatalf("ParseContent() error: %v", err)
	}

	errors := linter.ValidateSpecific(data, "output-styles/bad-name.md", contents)
	foundHyphenError := false
	for _, e := range errors {
		if e.Severity == "error" && containsSubstring(e.Message, "hyphen") {
			foundHyphenError = true
			break
		}
	}
	if !foundHyphenError {
		t.Error("ValidateSpecific() should report name starting with hyphen as error")
	}
}

func TestOutputStyleLinterCompileTimeChecks(t *testing.T) {
	// Verify interface compliance at compile time (already done via var _ declarations,
	// but this test ensures the linter satisfies the expected interfaces).
	var linter any = NewOutputStyleLinter()

	if _, ok := linter.(ComponentLinter); !ok {
		t.Error("OutputStyleLinter does not implement ComponentLinter")
	}
	if _, ok := linter.(PreValidator); !ok {
		t.Error("OutputStyleLinter does not implement PreValidator")
	}
	if _, ok := linter.(Scorable); !ok {
		t.Error("OutputStyleLinter does not implement Scorable")
	}
}

func TestOutputStyleLinterIntegration(t *testing.T) {
	// Test the full linting pipeline via lintFileCore
	linter := NewOutputStyleLinter()
	validator := cue.NewValidator()
	_ = validator.LoadSchemas("")

	tests := []struct {
		name        string
		contents    string
		wantSuccess bool
		wantErrors  int
	}{
		{
			name:        "valid output style",
			contents:    "---\nname: concise\ndescription: Write concisely\n---\n\nBe brief and clear.\n",
			wantSuccess: true,
			wantErrors:  0,
		},
		{
			name:        "missing all frontmatter fields",
			contents:    "---\n---\n\nSome body.\n",
			wantSuccess: false,
			wantErrors:  2, // missing name + missing description
		},
		{
			name:        "empty body",
			contents:    "---\nname: test\ndescription: Test\n---\n",
			wantSuccess: false,
			wantErrors:  1, // empty body
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := lintFileCore("output-styles/test.md", tt.contents, linter, validator, nil)

			if result.Success != tt.wantSuccess {
				t.Errorf("lintFileCore() success = %v, want %v, errors: %v", result.Success, tt.wantSuccess, result.Errors)
			}
			if len(result.Errors) != tt.wantErrors {
				t.Errorf("lintFileCore() errors = %d, want %d, got: %v", len(result.Errors), tt.wantErrors, result.Errors)
			}
		})
	}
}

// containsSubstring is defined in singlefile_test.go (shared test helper)
