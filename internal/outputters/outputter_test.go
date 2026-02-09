package outputters

import (
	"errors"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/config"
)

// =============================================================================
// Mock Formatter for testing
// =============================================================================

type mockFormatter struct {
	formatCalled bool
	formatError  error
	summary      *cli.LintSummary
}

func (m *mockFormatter) Format(summary *cli.LintSummary) error {
	m.formatCalled = true
	m.summary = summary
	return m.formatError
}

// =============================================================================
// Mock FormatterFactory for testing
// =============================================================================

type mockFormatterFactory struct {
	createCalled    bool
	requestedFormat string
	formatter       Formatter
	createError     error
}

func (m *mockFormatterFactory) CreateFormatter(format string) (Formatter, error) {
	m.createCalled = true
	m.requestedFormat = format
	if m.createError != nil {
		return nil, m.createError
	}
	return m.formatter, nil
}

// =============================================================================
// Test NewOutputter
// =============================================================================

func TestNewOutputter(t *testing.T) {
	cfg := &config.Config{
		Root:             "/test/root",
		Format:           "console",
		Quiet:            false,
		Verbose:          false,
		ShowScores:       false,
		ShowImprovements: false,
	}

	outputter := NewOutputter(cfg)

	if outputter == nil {
		t.Fatal("NewOutputter() returned nil")
	}

	if outputter.config != cfg {
		t.Errorf("NewOutputter() config = %v, want %v", outputter.config, cfg)
	}

	if outputter.factory == nil {
		t.Error("NewOutputter() factory is nil")
	}

	// Verify factory is DefaultFormatterFactory
	if _, ok := outputter.factory.(*DefaultFormatterFactory); !ok {
		t.Errorf("NewOutputter() factory type = %T, want *DefaultFormatterFactory", outputter.factory)
	}
}

// =============================================================================
// Test NewOutputterWithFactory
// =============================================================================

func TestNewOutputterWithFactory(t *testing.T) {
	cfg := &config.Config{
		Root:   "/test/root",
		Format: "json",
	}

	mockFactory := &mockFormatterFactory{}
	outputter := NewOutputterWithFactory(cfg, mockFactory)

	if outputter == nil {
		t.Fatal("NewOutputterWithFactory() returned nil")
	}

	if outputter.config != cfg {
		t.Errorf("NewOutputterWithFactory() config = %v, want %v", outputter.config, cfg)
	}

	if outputter.factory != mockFactory {
		t.Errorf("NewOutputterWithFactory() factory = %v, want %v", outputter.factory, mockFactory)
	}
}

// =============================================================================
// Test Format method
// =============================================================================

func TestOutputter_Format_Success(t *testing.T) {
	cfg := &config.Config{
		Root: "/test/root",
	}

	mockForm := &mockFormatter{}
	mockFactory := &mockFormatterFactory{
		formatter: mockForm,
	}

	outputter := NewOutputterWithFactory(cfg, mockFactory)

	summary := &cli.LintSummary{
		ComponentType:   "agents",
		TotalFiles:      10,
		SuccessfulFiles: 8,
		FailedFiles:     2,
	}

	err := outputter.Format(summary, "console")

	if err != nil {
		t.Errorf("Format() error = %v, want nil", err)
	}

	if !mockFactory.createCalled {
		t.Error("Format() did not call CreateFormatter")
	}

	if mockFactory.requestedFormat != "console" {
		t.Errorf("Format() requested format = %s, want 'console'", mockFactory.requestedFormat)
	}

	if !mockForm.formatCalled {
		t.Error("Format() did not call formatter.Format()")
	}

	if mockForm.summary != summary {
		t.Error("Format() passed wrong summary to formatter")
	}
}

func TestOutputter_Format_SetsStartTime(t *testing.T) {
	cfg := &config.Config{
		Root: "/test/root",
	}

	mockForm := &mockFormatter{}
	mockFactory := &mockFormatterFactory{
		formatter: mockForm,
	}

	outputter := NewOutputterWithFactory(cfg, mockFactory)

	summary := &cli.LintSummary{
		ComponentType: "commands",
		TotalFiles:    5,
		// StartTime not set (zero value)
	}

	beforeCall := time.Now()
	err := outputter.Format(summary, "json")
	afterCall := time.Now()

	if err != nil {
		t.Errorf("Format() error = %v, want nil", err)
	}

	// Verify StartTime was set
	if summary.StartTime.IsZero() {
		t.Error("Format() did not set StartTime")
	}

	// Verify StartTime is within reasonable bounds
	if summary.StartTime.Before(beforeCall) || summary.StartTime.After(afterCall) {
		t.Errorf("Format() StartTime = %v, want between %v and %v", summary.StartTime, beforeCall, afterCall)
	}
}

func TestOutputter_Format_PreservesExistingStartTime(t *testing.T) {
	cfg := &config.Config{
		Root: "/test/root",
	}

	mockForm := &mockFormatter{}
	mockFactory := &mockFormatterFactory{
		formatter: mockForm,
	}

	outputter := NewOutputterWithFactory(cfg, mockFactory)

	existingTime := time.Date(2024, 1, 1, 12, 0, 0, 0, time.UTC)
	summary := &cli.LintSummary{
		ComponentType: "skills",
		StartTime:     existingTime,
		TotalFiles:    3,
	}

	err := outputter.Format(summary, "markdown")

	if err != nil {
		t.Errorf("Format() error = %v, want nil", err)
	}

	// Verify StartTime was not changed
	if !summary.StartTime.Equal(existingTime) {
		t.Errorf("Format() changed StartTime from %v to %v", existingTime, summary.StartTime)
	}
}

func TestOutputter_Format_SetsProjectRoot(t *testing.T) {
	cfg := &config.Config{
		Root: "/custom/project/root",
	}

	mockForm := &mockFormatter{}
	mockFactory := &mockFormatterFactory{
		formatter: mockForm,
	}

	outputter := NewOutputterWithFactory(cfg, mockFactory)

	summary := &cli.LintSummary{
		ComponentType: "plugins",
		TotalFiles:    1,
	}

	err := outputter.Format(summary, "console")

	if err != nil {
		t.Errorf("Format() error = %v, want nil", err)
	}

	// Verify ProjectRoot was set from config
	if summary.ProjectRoot != "/custom/project/root" {
		t.Errorf("Format() ProjectRoot = %s, want '/custom/project/root'", summary.ProjectRoot)
	}
}

func TestOutputter_Format_CreateFormatterError(t *testing.T) {
	cfg := &config.Config{
		Root: "/test/root",
	}

	expectedErr := errors.New("unsupported format: invalid")
	mockFactory := &mockFormatterFactory{
		createError: expectedErr,
	}

	outputter := NewOutputterWithFactory(cfg, mockFactory)

	summary := &cli.LintSummary{
		ComponentType: "agents",
		TotalFiles:    10,
	}

	err := outputter.Format(summary, "invalid")

	if err == nil {
		t.Fatal("Format() error = nil, want error")
	}

	if err != expectedErr {
		t.Errorf("Format() error = %v, want %v", err, expectedErr)
	}
}

func TestOutputter_Format_FormatterError(t *testing.T) {
	cfg := &config.Config{
		Root: "/test/root",
	}

	expectedErr := errors.New("formatter failed")
	mockForm := &mockFormatter{
		formatError: expectedErr,
	}
	mockFactory := &mockFormatterFactory{
		formatter: mockForm,
	}

	outputter := NewOutputterWithFactory(cfg, mockFactory)

	summary := &cli.LintSummary{
		ComponentType: "commands",
		TotalFiles:    5,
	}

	err := outputter.Format(summary, "console")

	if err == nil {
		t.Fatal("Format() error = nil, want error")
	}

	if err != expectedErr {
		t.Errorf("Format() error = %v, want %v", err, expectedErr)
	}

	// Verify formatter was called despite error
	if !mockForm.formatCalled {
		t.Error("Format() did not call formatter.Format()")
	}
}

// =============================================================================
// Test DefaultFormatterFactory
// =============================================================================

func TestNewDefaultFormatterFactory(t *testing.T) {
	cfg := &config.Config{
		Root:   "/test/root",
		Format: "console",
	}

	factory := NewDefaultFormatterFactory(cfg)

	if factory == nil {
		t.Fatal("NewDefaultFormatterFactory() returned nil")
	}

	if factory.cfg != cfg {
		t.Errorf("NewDefaultFormatterFactory() cfg = %v, want %v", factory.cfg, cfg)
	}
}

func TestDefaultFormatterFactory_CreateFormatter_Console(t *testing.T) {
	cfg := &config.Config{
		Quiet:            true,
		Verbose:          false,
		ShowScores:       true,
		ShowImprovements: false,
	}

	factory := NewDefaultFormatterFactory(cfg)
	formatter, err := factory.CreateFormatter("console")

	if err != nil {
		t.Errorf("CreateFormatter('console') error = %v, want nil", err)
	}

	if formatter == nil {
		t.Fatal("CreateFormatter('console') returned nil formatter")
	}

	// Verify it's the correct type (console formatter from output package)
	// We can't inspect the internal fields without reflection, but we can verify it's not nil
}

func TestDefaultFormatterFactory_CreateFormatter_JSON(t *testing.T) {
	cfg := &config.Config{
		Quiet:  false,
		Output: "/tmp/output.json",
	}

	factory := NewDefaultFormatterFactory(cfg)
	formatter, err := factory.CreateFormatter("json")

	if err != nil {
		t.Errorf("CreateFormatter('json') error = %v, want nil", err)
	}

	if formatter == nil {
		t.Fatal("CreateFormatter('json') returned nil formatter")
	}
}

func TestDefaultFormatterFactory_CreateFormatter_Markdown(t *testing.T) {
	cfg := &config.Config{
		Quiet:   false,
		Verbose: true,
		Output:  "/tmp/output.md",
	}

	factory := NewDefaultFormatterFactory(cfg)
	formatter, err := factory.CreateFormatter("markdown")

	if err != nil {
		t.Errorf("CreateFormatter('markdown') error = %v, want nil", err)
	}

	if formatter == nil {
		t.Fatal("CreateFormatter('markdown') returned nil formatter")
	}
}

func TestDefaultFormatterFactory_CreateFormatter_Unsupported(t *testing.T) {
	cfg := &config.Config{}

	factory := NewDefaultFormatterFactory(cfg)
	formatter, err := factory.CreateFormatter("xml")

	if err == nil {
		t.Fatal("CreateFormatter('xml') error = nil, want error")
	}

	if formatter != nil {
		t.Errorf("CreateFormatter('xml') formatter = %v, want nil", formatter)
	}

	expectedErrMsg := "unsupported format: xml"
	if err.Error() != expectedErrMsg {
		t.Errorf("CreateFormatter('xml') error = %q, want %q", err.Error(), expectedErrMsg)
	}
}

func TestDefaultFormatterFactory_CreateFormatter_AllFormats(t *testing.T) {
	tests := []struct {
		name    string
		format  string
		wantErr bool
	}{
		{
			name:    "console format",
			format:  "console",
			wantErr: false,
		},
		{
			name:    "json format",
			format:  "json",
			wantErr: false,
		},
		{
			name:    "markdown format",
			format:  "markdown",
			wantErr: false,
		},
		{
			name:    "invalid format",
			format:  "yaml",
			wantErr: true,
		},
		{
			name:    "empty format",
			format:  "",
			wantErr: true,
		},
	}

	cfg := &config.Config{
		Output: "/tmp/output.txt",
	}
	factory := NewDefaultFormatterFactory(cfg)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formatter, err := factory.CreateFormatter(tt.format)

			if tt.wantErr && err == nil {
				t.Errorf("CreateFormatter(%q) error = nil, want error", tt.format)
			}
			if tt.wantErr && formatter != nil {
				t.Errorf("CreateFormatter(%q) returned formatter when error expected", tt.format)
			}
			if !tt.wantErr && err != nil {
				t.Errorf("CreateFormatter(%q) error = %v, want nil", tt.format, err)
			}
			if !tt.wantErr && formatter == nil {
				t.Errorf("CreateFormatter(%q) returned nil formatter", tt.format)
			}
		})
	}
}

// =============================================================================
// Integration tests
// =============================================================================

func TestOutputter_Format_Integration(t *testing.T) {
	tests := []struct {
		name        string
		config      *config.Config
		summary     *cli.LintSummary
		format      string
		wantErr     bool
		errContains string
	}{
		{
			name: "console format success",
			config: &config.Config{
				Root:       "/test/root",
				Quiet:      false,
				Verbose:    true,
				ShowScores: true,
			},
			summary: &cli.LintSummary{
				ComponentType:   "agents",
				TotalFiles:      10,
				SuccessfulFiles: 8,
				FailedFiles:     2,
			},
			format:  "console",
			wantErr: false,
		},
		{
			name: "json format success",
			config: &config.Config{
				Root:   "/test/root",
				Output: "/tmp/output.json",
			},
			summary: &cli.LintSummary{
				ComponentType:   "commands",
				TotalFiles:      5,
				SuccessfulFiles: 5,
				FailedFiles:     0,
			},
			format:  "json",
			wantErr: false,
		},
		{
			name: "markdown format success",
			config: &config.Config{
				Root:    "/test/root",
				Output:  "/tmp/output.md",
				Verbose: true,
			},
			summary: &cli.LintSummary{
				ComponentType:   "skills",
				TotalFiles:      3,
				SuccessfulFiles: 2,
				FailedFiles:     1,
			},
			format:  "markdown",
			wantErr: false,
		},
		{
			name: "unsupported format",
			config: &config.Config{
				Root: "/test/root",
			},
			summary: &cli.LintSummary{
				ComponentType: "plugins",
				TotalFiles:    1,
			},
			format:      "html",
			wantErr:     true,
			errContains: "unsupported format: html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			outputter := NewOutputter(tt.config)
			err := outputter.Format(tt.summary, tt.format)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Format() error = nil, want error containing %q", tt.errContains)
					return
				}
				if tt.errContains != "" && err.Error() != tt.errContains {
					t.Errorf("Format() error = %q, want error containing %q", err.Error(), tt.errContains)
				}
				return
			}
			if err != nil {
				t.Errorf("Format() error = %v, want nil", err)
			}

			// Verify summary was populated correctly
			if tt.summary.ProjectRoot != tt.config.Root {
				t.Errorf("Format() did not set ProjectRoot correctly: got %q, want %q", tt.summary.ProjectRoot, tt.config.Root)
			}

			if tt.summary.StartTime.IsZero() {
				t.Error("Format() did not set StartTime")
			}
		})
	}
}

// =============================================================================
// Edge cases and boundary tests
// =============================================================================

func TestOutputter_Format_NilSummary(t *testing.T) {
	cfg := &config.Config{
		Root: "/test/root",
	}

	mockForm := &mockFormatter{}
	mockFactory := &mockFormatterFactory{
		formatter: mockForm,
	}

	outputter := NewOutputterWithFactory(cfg, mockFactory)

	// This should panic or handle nil gracefully
	// Currently the code doesn't check for nil, which is a potential issue
	// but we're testing current behavior
	defer func() {
		if r := recover(); r != nil {
			// Expected panic when accessing nil summary
			t.Logf("Format() panicked with nil summary (expected): %v", r)
		}
	}()

	_ = outputter.Format(nil, "console")
}

func TestOutputter_Format_EmptySummary(t *testing.T) {
	cfg := &config.Config{
		Root: "/test/root",
	}

	mockForm := &mockFormatter{}
	mockFactory := &mockFormatterFactory{
		formatter: mockForm,
	}

	outputter := NewOutputterWithFactory(cfg, mockFactory)

	summary := &cli.LintSummary{}

	err := outputter.Format(summary, "console")

	if err != nil {
		t.Errorf("Format() with empty summary error = %v, want nil", err)
	}

	// Verify summary was populated
	if summary.ProjectRoot != cfg.Root {
		t.Errorf("Format() ProjectRoot = %s, want %s", summary.ProjectRoot, cfg.Root)
	}

	if summary.StartTime.IsZero() {
		t.Error("Format() did not set StartTime for empty summary")
	}
}

func TestOutputter_MultipleFormats(t *testing.T) {
	cfg := &config.Config{
		Root:   "/test/root",
		Output: "/tmp/output.txt",
	}

	outputter := NewOutputter(cfg)

	summary := &cli.LintSummary{
		ComponentType:   "agents",
		TotalFiles:      10,
		SuccessfulFiles: 8,
		FailedFiles:     2,
	}

	formats := []string{"console", "json", "markdown"}

	for _, format := range formats {
		t.Run(format, func(t *testing.T) {
			err := outputter.Format(summary, format)
			if err != nil {
				t.Errorf("Format(%q) error = %v, want nil", format, err)
			}
		})
	}
}
