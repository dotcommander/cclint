package lint

import (
	"strings"
	"errors"
	"testing"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

func TestDetectXMLTags(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"no XML", "Plain text content", false},
		{"XML tag", "Content with <tag>xml</tag>", true},
		{"self-closing", "Content with <br/>", true},
		{"HTML entities", "Content with &amp; entity", false},
		{"angle in math", "x < y and y > z", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := DetectXMLTags(tt.content, "Test", "test.md", "")
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("DetectXMLTags(%q) hasErr = %v, want %v", tt.content, hasErr, tt.wantErr)
			}
		})
	}
}

func TestCheckSizeLimit(t *testing.T) {
	tests := []struct {
		name          string
		contents      string
		limit         int
		tolerance     float64
		componentType string
		wantSugg      bool
	}{
		{
			name:          "under limit",
			contents:      strings.Repeat("line\n", 50),
			limit:         200,
			tolerance:     0.10,
			componentType: "agent",
			wantSugg:      false,
		},
		{
			name:          "over limit",
			contents:      strings.Repeat("line\n", 250),
			limit:         200,
			tolerance:     0.10,
			componentType: "agent",
			wantSugg:      true,
		},
		{
			name:          "command over limit",
			contents:      strings.Repeat("line\n", 60),
			limit:         50,
			tolerance:     0.10,
			componentType: "command",
			wantSugg:      true,
		},
		{
			name:          "skill over limit",
			contents:      strings.Repeat("line\n", 600),
			limit:         500,
			tolerance:     0.10,
			componentType: "skill",
			wantSugg:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := CheckSizeLimit(tt.contents, tt.limit, tt.tolerance, tt.componentType, "test.md")
			hasSugg := err != nil
			if hasSugg != tt.wantSugg {
				t.Errorf("CheckSizeLimit() hasSugg = %v, want %v", hasSugg, tt.wantSugg)
			}
			if hasSugg && err.Severity != "suggestion" {
				t.Errorf("CheckSizeLimit() severity = %q, want %q", err.Severity, "suggestion")
			}
		})
	}
}

func TestValidateSemver(t *testing.T) {
	tests := []struct {
		name    string
		version string
		wantErr bool
	}{
		{"valid 1.0.0", "1.0.0", false},
		{"valid 1.2.3", "1.2.3", false},
		{"valid with pre-release", "1.0.0-alpha", false},
		{"valid with build", "1.0.0+20130313144700", false},
		{"valid with both", "1.0.0-beta+exp.sha.5114f85", false},
		{"invalid 1.0", "1.0", true},
		{"invalid v1.0.0", "v1.0.0", true},
		{"invalid 1", "1", true},
		{"invalid text", "latest", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSemver(tt.version, "test.md", 1)
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("ValidateSemver(%q) hasErr = %v, want %v", tt.version, hasErr, tt.wantErr)
			}
		})
	}
}

func TestParseFrontmatter(t *testing.T) {
	tests := []struct {
		name        string
		contents    string
		wantErr     bool
		wantDataKey string
	}{
		{
			name:        "valid frontmatter",
			contents:    "---\nname: test\n---\nBody",
			wantErr:     false,
			wantDataKey: "name",
		},
		{
			name:     "no frontmatter",
			contents: "Just body content",
			wantErr:  false,
		},
		{
			name:     "invalid YAML",
			contents: "---\nname: [unclosed\n---\n",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _, err := parseFrontmatter(tt.contents)
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("parseFrontmatter() hasErr = %v, want %v", hasErr, tt.wantErr)
			}
			if !hasErr && tt.wantDataKey != "" {
				if _, ok := data[tt.wantDataKey]; !ok {
					t.Errorf("parseFrontmatter() missing expected key %q", tt.wantDataKey)
				}
			}
		})
	}
}

func TestParseJSONContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{
			name:    "valid JSON",
			content: `{"name": "test", "version": "1.0.0"}`,
			wantErr: false,
		},
		{
			name:    "invalid JSON",
			content: `{name: test}`,
			wantErr: true,
		},
		{
			name:    "empty object",
			content: `{}`,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, body, err := parseJSONContent(tt.content)
			hasErr := err != nil
			if hasErr != tt.wantErr {
				t.Errorf("parseJSONContent() hasErr = %v, want %v", hasErr, tt.wantErr)
			}
			if !hasErr {
				if data == nil {
					t.Error("parseJSONContent() data is nil")
				}
				if body != "" {
					t.Error("parseJSONContent() body should be empty for JSON")
				}
			}
		})
	}
}

// Mock linter for testing generic linter infrastructure
type mockLinter struct {
	BaseLinter
	typeStr       string
	fileType      discovery.FileType
	parseErr      error
	cueErrors     []cue.ValidationError
	specificErrs  []cue.ValidationError
	preValidErrs  []cue.ValidationError
	bestPracErrs  []cue.ValidationError
	postProcessed bool
}

func (m *mockLinter) Type() string { return m.typeStr }
func (m *mockLinter) FileType() discovery.FileType { return m.fileType }

func (m *mockLinter) ParseContent(contents string) (map[string]any, string, error) {
	if m.parseErr != nil {
		return nil, "", m.parseErr
	}
	return map[string]any{"test": "data"}, "body", nil
}

func (m *mockLinter) ValidateCUE(validator *cue.Validator, data map[string]any) ([]cue.ValidationError, error) {
	return m.cueErrors, nil
}

func (m *mockLinter) ValidateSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	return m.specificErrs
}

func (m *mockLinter) PreValidate(filePath, contents string) []cue.ValidationError {
	return m.preValidErrs
}

func (m *mockLinter) ValidateBestPractices(filePath, contents string, data map[string]any) []cue.ValidationError {
	return m.bestPracErrs
}

func (m *mockLinter) PostProcess(result *LintResult) {
	m.postProcessed = true
}

func TestLintComponent(t *testing.T) {
	tmpDir := t.TempDir()
	validator := cue.NewValidator()
	_ = validator.LoadSchemas("")

	tests := []struct {
		name         string
		linter       *mockLinter
		fileContents string
		wantSuccess  bool
		wantErrCount int
	}{
		{
			name: "successful lint",
			linter: &mockLinter{
				typeStr:  "test",
				fileType: discovery.FileTypeAgent,
			},
			fileContents: "test content",
			wantSuccess:  true,
			wantErrCount: 0,
		},
		{
			name: "parse error",
			linter: &mockLinter{
				typeStr:  "test",
				fileType: discovery.FileTypeAgent,
				parseErr: errors.New("parse failed"),
			},
			fileContents: "test content",
			wantSuccess:  false,
			wantErrCount: 1,
		},
		{
			name: "validation errors",
			linter: &mockLinter{
				typeStr:  "test",
				fileType: discovery.FileTypeAgent,
				specificErrs: []cue.ValidationError{
					{Message: "error 1", Severity: "error"},
					{Message: "error 2", Severity: "error"},
				},
			},
			fileContents: "test content",
			wantSuccess:  false,
			wantErrCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := &SingleFileLinterContext{
				RootPath: tmpDir,
				File: discovery.File{
					RelPath:  "test.md",
					Contents: tt.fileContents,
				},
				Validator: validator,
			}

			result := lintComponent(ctx, tt.linter)

			if result.Success != tt.wantSuccess {
				t.Errorf("lintComponent() success = %v, want %v", result.Success, tt.wantSuccess)
			}

			if len(result.Errors) != tt.wantErrCount {
				t.Errorf("lintComponent() errors = %d, want %d", len(result.Errors), tt.wantErrCount)
			}

			// Verify PostProcess was called for successful parsing
			if tt.wantSuccess && !tt.linter.postProcessed {
				t.Error("PostProcess was not called")
			}
		})
	}
}

func TestDetectSwallowedFields(t *testing.T) {
	tests := []struct {
		name      string
		contents  string
		compType  string
		wantCount int
		wantField string // first swallowed field name (if any)
	}{
		{
			name: "pipe swallows model",
			contents: "---\nname: test\ndescription: |\n  Some text.\n  model: haiku\n---\n# Body",
			compType: "skill",
			wantCount: 1,
			wantField: "model",
		},
		{
			name: "pipe swallows multiple fields",
			contents: "---\nname: test\ndescription: |\n  Some text.\n  model: haiku\n  context: fork\n---\n# Body",
			compType: "skill",
			wantCount: 2,
			wantField: "model",
		},
		{
			name: "folded scalar swallows model",
			contents: "---\nname: test\ndescription: >\n  Folded text.\n  model: sonnet\n---\n# Body",
			compType: "skill",
			wantCount: 1,
			wantField: "model",
		},
		{
			name: "clean inline description",
			contents: "---\nname: test\ndescription: A clean description\nmodel: haiku\n---\n# Body",
			compType: "skill",
			wantCount: 0,
		},
		{
			name: "pipe with no swallowed fields",
			contents: "---\nname: test\ndescription: |\n  Just normal text here.\n  Nothing special.\n---\n# Body",
			compType: "skill",
			wantCount: 0,
		},
		{
			name: "no frontmatter",
			contents: "# Just a markdown file\nNo frontmatter here.",
			compType: "skill",
			wantCount: 0,
		},
		{
			name: "agent pipe swallows model",
			contents: "---\nname: test-agent\ndescription: |\n  Agent description.\n  model: opus\n---\n# Body",
			compType: "agent",
			wantCount: 1,
			wantField: "model",
		},
		{
			name: "pipe with strip modifier",
			contents: "---\nname: test\ndescription: |-\n  Some text.\n  model: haiku\n---\n# Body",
			compType: "skill",
			wantCount: 1,
			wantField: "model",
		},
		{
			name: "non-field indented text ignored",
			contents: "---\nname: test\ndescription: |\n  model_training is important.\n  context_switching too.\n---\n# Body",
			compType: "skill",
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := DetectSwallowedFields(tt.contents, "test.md", tt.compType)
			if len(errs) != tt.wantCount {
				t.Errorf("DetectSwallowedFields() got %d errors, want %d", len(errs), tt.wantCount)
				for _, e := range errs {
					t.Logf("  error: %s", e.Message)
				}
			}
			if tt.wantCount > 0 && len(errs) > 0 && tt.wantField != "" {
				if !strings.Contains(errs[0].Message, "'"+tt.wantField+"'") {
					t.Errorf("first error should mention field '%s', got: %s", tt.wantField, errs[0].Message)
				}
			}
		})
	}
}

func TestValidateAllowedToolsShared(t *testing.T) {
	// This is a wrapper function, test it exists
	data := map[string]any{
		"allowed-tools": "Read, Write",
	}
	errors := ValidateAllowedToolsShared(data, "test.md", "")
	// Verify function exists and returns without panic
	_ = errors
}
