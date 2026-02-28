package lint

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

func TestNewLinterContext(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal project structure
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a CLAUDE.md to make it a valid project root
	claudeMdPath := filepath.Join(tmpDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMdPath, []byte("# Project"), 0644); err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		name         string
		rootPath     string
		quiet        bool
		verbose      bool
		noCycleCheck bool
		wantErr      bool
	}{
		{
			name:         "valid context with explicit root",
			rootPath:     tmpDir,
			quiet:        false,
			verbose:      false,
			noCycleCheck: false,
			wantErr:      false,
		},
		{
			name:         "quiet mode",
			rootPath:     tmpDir,
			quiet:        true,
			verbose:      false,
			noCycleCheck: false,
			wantErr:      false,
		},
		{
			name:         "verbose mode",
			rootPath:     tmpDir,
			quiet:        false,
			verbose:      true,
			noCycleCheck: false,
			wantErr:      false,
		},
		{
			name:         "no cycle check",
			rootPath:     tmpDir,
			quiet:        false,
			verbose:      false,
			noCycleCheck: true,
			wantErr:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx, err := NewLinterContext(tt.rootPath, tt.quiet, tt.verbose, tt.noCycleCheck)

			if tt.wantErr {
				if err == nil {
					t.Error("NewLinterContext() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("NewLinterContext() unexpected error: %v", err)
				return
			}

			if ctx.RootPath != tt.rootPath {
				t.Errorf("NewLinterContext() RootPath = %q, want %q", ctx.RootPath, tt.rootPath)
			}

			if ctx.Quiet != tt.quiet {
				t.Errorf("NewLinterContext() Quiet = %v, want %v", ctx.Quiet, tt.quiet)
			}

			if ctx.Verbose != tt.verbose {
				t.Errorf("NewLinterContext() Verbose = %v, want %v", ctx.Verbose, tt.verbose)
			}

			if ctx.NoCycleCheck != tt.noCycleCheck {
				t.Errorf("NewLinterContext() NoCycleCheck = %v, want %v", ctx.NoCycleCheck, tt.noCycleCheck)
			}

			if ctx.Validator == nil {
				t.Error("NewLinterContext() Validator is nil")
			}

			if ctx.Discoverer == nil {
				t.Error("NewLinterContext() Discoverer is nil")
			}

			if ctx.CrossValidator == nil {
				t.Error("NewLinterContext() CrossValidator is nil")
			}
		})
	}
}

func TestFilterFilesByType(t *testing.T) {
	ctx := &LinterContext{
		Files: []discovery.File{
			{RelPath: "agents/test.md", Type: discovery.FileTypeAgent},
			{RelPath: "commands/cmd.md", Type: discovery.FileTypeCommand},
			{RelPath: "skills/skill/SKILL.md", Type: discovery.FileTypeSkill},
			{RelPath: "agents/other.md", Type: discovery.FileTypeAgent},
		},
	}

	tests := []struct {
		name     string
		fileType discovery.FileType
		want     int
	}{
		{"filter agents", discovery.FileTypeAgent, 2},
		{"filter commands", discovery.FileTypeCommand, 1},
		{"filter skills", discovery.FileTypeSkill, 1},
		{"filter settings", discovery.FileTypeSettings, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			filtered := ctx.FilterFilesByType(tt.fileType)
			if len(filtered) != tt.want {
				t.Errorf("FilterFilesByType() returned %d files, want %d", len(filtered), tt.want)
			}
		})
	}
}

func TestNewSummary(t *testing.T) {
	ctx := &LinterContext{
		RootPath: "/test/path",
	}

	summary := ctx.NewSummary(10)

	if summary.ProjectRoot != ctx.RootPath {
		t.Errorf("NewSummary() ProjectRoot = %q, want %q", summary.ProjectRoot, ctx.RootPath)
	}

	if summary.TotalFiles != 10 {
		t.Errorf("NewSummary() TotalFiles = %d, want 10", summary.TotalFiles)
	}
}

func TestLogProcessed(t *testing.T) {
	ctx := &LinterContext{}
	// This is a no-op function, just verify it doesn't panic
	ctx.LogProcessed("test.md", 0)
	ctx.LogProcessed("test.md", 5)
}

func TestLogProcessedWithSuggestions(t *testing.T) {
	ctx := &LinterContext{}
	// This is a no-op function, just verify it doesn't panic
	ctx.LogProcessedWithSuggestions("test.md", 0, 0)
	ctx.LogProcessedWithSuggestions("test.md", 5, 3)
}

func TestLinterContextAutoDiscoverRoot(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a minimal project structure
	claudeDir := filepath.Join(tmpDir, ".claude")
	if err := os.MkdirAll(filepath.Join(claudeDir, "agents"), 0755); err != nil {
		t.Fatal(err)
	}

	// Create a CLAUDE.md
	claudeMdPath := filepath.Join(tmpDir, "CLAUDE.md")
	if err := os.WriteFile(claudeMdPath, []byte("# Project"), 0644); err != nil {
		t.Fatal(err)
	}

	// Change to tmp directory
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if chErr := os.Chdir(tmpDir); chErr != nil {
		t.Fatal(chErr)
	}

	// Test with empty rootPath - should auto-discover
	ctx, err := NewLinterContext("", false, false, false)
	if err != nil {
		t.Fatalf("NewLinterContext() with empty root failed: %v", err)
	}

	if ctx.RootPath == "" {
		t.Error("NewLinterContext() with empty root should discover root")
	}
}
