package lint

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cue"
)

// =============================================================================
// Test NewOrchestrator
// =============================================================================

func TestNewOrchestrator(t *testing.T) {
	cfg := &config.Config{
		Root:    "/test/root",
		Format:  "console",
		Quiet:   false,
		Verbose: true,
	}

	opts := OrchestratorConfig{
		RootPath:       "/test/root",
		UseBaseline:    false,
		CreateBaseline: false,
		BaselinePath:   ".cclintbaseline.json",
	}

	orch := NewOrchestrator(cfg, opts)

	if orch == nil {
		t.Fatal("NewOrchestrator() returned nil")
	}

	if orch.cfg != cfg {
		t.Error("Config not set correctly")
	}

	if orch.opts.RootPath != opts.RootPath {
		t.Errorf("RootPath = %s, want %s", orch.opts.RootPath, opts.RootPath)
	}

	// Should initialize with default linters
	if len(orch.linters) == 0 {
		t.Error("Default linters not initialized")
	}
}

// =============================================================================
// Test WithLinters
// =============================================================================

func TestWithLinters(t *testing.T) {
	cfg := &config.Config{Root: "/test", Format: "console"}
	opts := OrchestratorConfig{RootPath: "/test"}

	orch := NewOrchestrator(cfg, opts)

	// Create custom linters
	customLinters := []ComponentLinter{
		{
			Name: "test-linter",
			Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
				return &cli.LintSummary{}, nil
			},
		},
	}

	// Apply custom linters
	result := orch.WithLinters(customLinters)

	// Should return self for chaining
	if result != orch {
		t.Error("WithLinters() didn't return self for chaining")
	}

	if len(orch.linters) != 1 {
		t.Errorf("Linters count = %d, want 1", len(orch.linters))
	}

	if orch.linters[0].Name != "test-linter" {
		t.Errorf("Linter name = %s, want test-linter", orch.linters[0].Name)
	}
}

// =============================================================================
// Test DefaultLinters
// =============================================================================

func TestDefaultLinters(t *testing.T) {
	linters := DefaultLinters()

	if len(linters) == 0 {
		t.Fatal("DefaultLinters() returned empty slice")
	}

	// Check that expected linters are present
	expectedNames := []string{"agents", "commands", "skills", "settings", "context", "rules"}
	linterMap := make(map[string]bool)

	for _, l := range linters {
		linterMap[l.Name] = true
		if l.Linter == nil {
			t.Errorf("Linter %s has nil Linter function", l.Name)
		}
	}

	for _, name := range expectedNames {
		if !linterMap[name] {
			t.Errorf("Expected linter %s not found in DefaultLinters", name)
		}
	}
}

// =============================================================================
// Test resolveBaselinePath
// =============================================================================

func TestResolveBaselinePath(t *testing.T) {
	tests := []struct {
		name         string
		rootPath     string
		baselinePath string
		want         string
	}{
		{
			name:         "absolute path unchanged",
			rootPath:     "/project",
			baselinePath: "/absolute/path/baseline.json",
			want:         "/absolute/path/baseline.json",
		},
		{
			name:         "relative path joined to root",
			rootPath:     "/project",
			baselinePath: ".cclintbaseline.json",
			want:         "/project/.cclintbaseline.json",
		},
		{
			name:         "relative path with subdirectory",
			rootPath:     "/project",
			baselinePath: "baselines/custom.json",
			want:         "/project/baselines/custom.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Root: tt.rootPath, Format: "console"}
			opts := OrchestratorConfig{
				RootPath:     tt.rootPath,
				BaselinePath: tt.baselinePath,
			}

			orch := NewOrchestrator(cfg, opts)
			result := orch.resolveBaselinePath()

			if result != tt.want {
				t.Errorf("resolveBaselinePath() = %s, want %s", result, tt.want)
			}
		})
	}
}

// =============================================================================
// Test loadBaseline
// =============================================================================

func TestLoadBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Create a baseline file
	issues := []cue.ValidationError{
		{
			File:     "test.md",
			Message:  "Test error",
			Severity: "error",
			Source:   "test",
		},
	}
	b := baseline.CreateBaseline(issues)
	b.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := b.SaveBaseline(baselinePath); err != nil {
		t.Fatalf("Failed to create test baseline: %v", err)
	}

	tests := []struct {
		name             string
		useBaseline      bool
		createBaseline   bool
		baselineExists   bool
		wantNil          bool
		wantError        bool
	}{
		{
			name:           "baseline mode disabled",
			useBaseline:    false,
			createBaseline: false,
			baselineExists: true,
			wantNil:        true,
			wantError:      false,
		},
		{
			name:           "use baseline - file exists",
			useBaseline:    true,
			createBaseline: false,
			baselineExists: true,
			wantNil:        false,
			wantError:      false,
		},
		{
			name:           "use baseline - file missing",
			useBaseline:    true,
			createBaseline: false,
			baselineExists: false,
			wantNil:        true,
			wantError:      false, // Missing file is not an error
		},
		{
			name:           "create baseline mode",
			useBaseline:    false,
			createBaseline: true,
			baselineExists: true,
			wantNil:        false,
			wantError:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{Root: tmpDir, Format: "console"}

			testBaselinePath := baselinePath
			if !tt.baselineExists {
				testBaselinePath = filepath.Join(tmpDir, "nonexistent.json")
			}

			opts := OrchestratorConfig{
				RootPath:       tmpDir,
				UseBaseline:    tt.useBaseline,
				CreateBaseline: tt.createBaseline,
				BaselinePath:   testBaselinePath,
			}

			orch := NewOrchestrator(cfg, opts)
			result, err := orch.loadBaseline(testBaselinePath)

			if tt.wantError && err == nil {
				t.Error("loadBaseline() expected error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Errorf("loadBaseline() unexpected error: %v", err)
			}

			if tt.wantNil && result != nil {
				t.Error("loadBaseline() expected nil, got baseline")
			}

			if !tt.wantNil && !tt.wantError && result == nil {
				t.Error("loadBaseline() expected baseline, got nil")
			}
		})
	}
}

// =============================================================================
// Test saveBaseline
// =============================================================================

func TestSaveBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, "test_baseline.json")

	issues := []cue.ValidationError{
		{
			File:     "agents/test.md",
			Message:  "Error 1",
			Severity: "error",
			Source:   "test",
		},
		{
			File:     "commands/cmd.md",
			Message:  "Error 2",
			Severity: "error",
			Source:   "test",
		},
	}

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  true, // Quiet mode to avoid output
	}
	opts := OrchestratorConfig{
		RootPath:     tmpDir,
		BaselinePath: baselinePath,
	}

	orch := NewOrchestrator(cfg, opts)
	err := orch.saveBaseline(issues, baselinePath)

	if err != nil {
		t.Fatalf("saveBaseline() error: %v", err)
	}

	// Verify file was created
	if _, statErr := os.Stat(baselinePath); statErr != nil {
		t.Errorf("Baseline file not created: %v", statErr)
	}

	// Load and verify contents
	loaded, err := baseline.LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("Failed to load saved baseline: %v", err)
	}

	if len(loaded.Fingerprints) != 2 {
		t.Errorf("Baseline contains %d fingerprints, want 2", len(loaded.Fingerprints))
	}
}

// =============================================================================
// Test Run - happy path
// =============================================================================

func TestRun_Success(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Root:    tmpDir,
		Format:  "console",
		Quiet:   true,
		Verbose: false,
	}

	opts := OrchestratorConfig{
		RootPath:       tmpDir,
		UseBaseline:    false,
		CreateBaseline: false,
		BaselinePath:   ".cclintbaseline.json",
	}

	// Create orchestrator with mock linter that returns successful results
	orch := NewOrchestrator(cfg, opts)

	successLinter := ComponentLinter{
		Name: "test-linter",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return &cli.LintSummary{
				ProjectRoot:      rootPath,
				ComponentType:    "test",
				StartTime:        time.Now(),
				TotalFiles:       2,
				SuccessfulFiles:  2,
				FailedFiles:      0,
				TotalErrors:      0,
				TotalSuggestions: 0,
				Results: []cli.LintResult{
					{File: "test1.md", Success: true, Errors: nil},
					{File: "test2.md", Success: true, Errors: nil},
				},
			}, nil
		},
	}

	orch.WithLinters([]ComponentLinter{successLinter})

	result, err := orch.Run()

	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if result == nil {
		t.Fatal("Run() returned nil result")
	}

	if result.TotalFiles != 2 {
		t.Errorf("TotalFiles = %d, want 2", result.TotalFiles)
	}

	if result.TotalErrors != 0 {
		t.Errorf("TotalErrors = %d, want 0", result.TotalErrors)
	}

	if result.HasErrors {
		t.Error("HasErrors = true, want false")
	}
}

// =============================================================================
// Test Run - with errors
// =============================================================================

func TestRun_WithErrors(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Root:    tmpDir,
		Format:  "console",
		Quiet:   true,
		Verbose: false,
	}

	opts := OrchestratorConfig{
		RootPath:       tmpDir,
		UseBaseline:    false,
		CreateBaseline: false,
		BaselinePath:   ".cclintbaseline.json",
	}

	orch := NewOrchestrator(cfg, opts)

	errorLinter := ComponentLinter{
		Name: "error-linter",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return &cli.LintSummary{
				ProjectRoot:      rootPath,
				ComponentType:    "test",
				StartTime:        time.Now(),
				TotalFiles:       1,
				SuccessfulFiles:  0,
				FailedFiles:      1,
				TotalErrors:      2,
				TotalSuggestions: 1,
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Success: false,
						Errors: []cue.ValidationError{
							{File: "test.md", Message: "Error 1", Severity: "error"},
							{File: "test.md", Message: "Error 2", Severity: "error"},
						},
						Suggestions: []cue.ValidationError{
							{File: "test.md", Message: "Suggestion 1", Severity: "suggestion"},
						},
					},
				},
			}, nil
		},
	}

	orch.WithLinters([]ComponentLinter{errorLinter})

	result, err := orch.Run()

	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if !result.HasErrors {
		t.Error("HasErrors = false, want true")
	}

	if result.TotalErrors != 2 {
		t.Errorf("TotalErrors = %d, want 2", result.TotalErrors)
	}

	if result.TotalSuggestions != 1 {
		t.Errorf("TotalSuggestions = %d, want 1", result.TotalSuggestions)
	}
}

// =============================================================================
// Test Run - skip empty results
// =============================================================================

func TestRun_SkipEmptyResults(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  true,
	}

	opts := OrchestratorConfig{
		RootPath:     tmpDir,
		BaselinePath: ".cclintbaseline.json",
	}

	orch := NewOrchestrator(cfg, opts)

	emptyLinter := ComponentLinter{
		Name: "empty-linter",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return &cli.LintSummary{
				TotalFiles: 0, // No files found
				Results:    []cli.LintResult{},
			}, nil
		},
	}

	orch.WithLinters([]ComponentLinter{emptyLinter})

	result, err := orch.Run()

	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Empty results should be skipped
	if result.TotalFiles != 0 {
		t.Errorf("TotalFiles = %d, want 0 (empty results should be skipped)", result.TotalFiles)
	}
}

// =============================================================================
// Test Run - baseline creation
// =============================================================================

func TestRun_CreateBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, ".cclintbaseline.json")

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  true,
	}

	opts := OrchestratorConfig{
		RootPath:       tmpDir,
		UseBaseline:    false,
		CreateBaseline: true,
		BaselinePath:   baselinePath,
	}

	orch := NewOrchestrator(cfg, opts)

	linter := ComponentLinter{
		Name: "test-linter",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return &cli.LintSummary{
				TotalFiles:  1,
				TotalErrors: 1,
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Success: false,
						Errors: []cue.ValidationError{
							{File: "test.md", Message: "Error", Severity: "error", Source: "test"},
						},
					},
				},
			}, nil
		},
	}

	orch.WithLinters([]ComponentLinter{linter})

	result, err := orch.Run()

	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Baseline creation should clear errors
	if result.HasErrors {
		t.Error("HasErrors = true, want false (baseline creation accepts current state)")
	}

	// Verify baseline file was created
	if _, statErr := os.Stat(baselinePath); statErr != nil {
		t.Errorf("Baseline file not created: %v", statErr)
	}

	// Load and verify baseline contains the issue
	loaded, err := baseline.LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	if len(loaded.Fingerprints) != 1 {
		t.Errorf("Baseline contains %d fingerprints, want 1", len(loaded.Fingerprints))
	}
}

// =============================================================================
// Test Run - baseline filtering
// =============================================================================

func TestRun_WithBaselineFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Create baseline with known issues
	knownIssues := []cue.ValidationError{
		{File: "test.md", Message: "Known error", Severity: "error", Source: "test"},
	}
	b := baseline.CreateBaseline(knownIssues)
	b.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := b.SaveBaseline(baselinePath); err != nil {
		t.Fatalf("Failed to create baseline: %v", err)
	}

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  true,
	}

	opts := OrchestratorConfig{
		RootPath:       tmpDir,
		UseBaseline:    true,
		CreateBaseline: false,
		BaselinePath:   baselinePath,
	}

	orch := NewOrchestrator(cfg, opts)

	linter := ComponentLinter{
		Name: "test-linter",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return &cli.LintSummary{
				TotalFiles:  1,
				TotalErrors: 2,
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Success: false,
						Errors: []cue.ValidationError{
							{File: "test.md", Message: "Known error", Severity: "error", Source: "test"},    // Should be filtered
							{File: "test.md", Message: "New error", Severity: "error", Source: "test"},      // Should remain
						},
					},
				},
			}, nil
		},
	}

	orch.WithLinters([]ComponentLinter{linter})

	result, err := orch.Run()

	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// One error should be filtered by baseline
	if result.BaselineIgnored == 0 {
		t.Error("BaselineIgnored = 0, want > 0")
	}

	// Should still have one new error
	if result.TotalErrors == 0 {
		t.Error("TotalErrors = 0, want > 0 (new error should remain)")
	}
}

// =============================================================================
// Test Run - multiple linters
// =============================================================================

func TestRun_MultipleLinters(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  true,
	}

	opts := OrchestratorConfig{
		RootPath:     tmpDir,
		BaselinePath: ".cclintbaseline.json",
	}

	orch := NewOrchestrator(cfg, opts)

	linter1 := ComponentLinter{
		Name: "linter-1",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return &cli.LintSummary{
				TotalFiles:       2,
				TotalErrors:      1,
				TotalSuggestions: 0,
				Results: []cli.LintResult{
					{File: "file1.md", Success: true},
					{File: "file2.md", Success: false, Errors: []cue.ValidationError{
						{File: "file2.md", Message: "Error", Severity: "error"},
					}},
				},
			}, nil
		},
	}

	linter2 := ComponentLinter{
		Name: "linter-2",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return &cli.LintSummary{
				TotalFiles:       1,
				TotalErrors:      0,
				TotalSuggestions: 2,
				Results: []cli.LintResult{
					{File: "file3.md", Success: true, Suggestions: []cue.ValidationError{
						{File: "file3.md", Message: "Suggestion 1", Severity: "suggestion"},
						{File: "file3.md", Message: "Suggestion 2", Severity: "suggestion"},
					}},
				},
			}, nil
		},
	}

	orch.WithLinters([]ComponentLinter{linter1, linter2})

	result, err := orch.Run()

	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	// Verify aggregated results
	if result.TotalFiles != 3 {
		t.Errorf("TotalFiles = %d, want 3", result.TotalFiles)
	}

	if result.TotalErrors != 1 {
		t.Errorf("TotalErrors = %d, want 1", result.TotalErrors)
	}

	if result.TotalSuggestions != 2 {
		t.Errorf("TotalSuggestions = %d, want 2", result.TotalSuggestions)
	}

	if !result.HasErrors {
		t.Error("HasErrors = false, want true")
	}
}

// =============================================================================
// Test Run - linter error handling
// =============================================================================

func TestRun_LinterError(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  true,
	}

	opts := OrchestratorConfig{
		RootPath:     tmpDir,
		BaselinePath: ".cclintbaseline.json",
	}

	orch := NewOrchestrator(cfg, opts)

	failingLinter := ComponentLinter{
		Name: "failing-linter",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return nil, os.ErrNotExist // Return an error
		},
	}

	orch.WithLinters([]ComponentLinter{failingLinter})

	_, err := orch.Run()

	if err == nil {
		t.Fatal("Run() expected error from failing linter, got nil")
	}

	// Error message should mention the linter name
	if !containsString(err.Error(), "failing-linter") {
		t.Errorf("Error message should mention linter name, got: %s", err.Error())
	}
}

// =============================================================================
// Test Run - baseline load error (non-quiet mode)
// =============================================================================

func TestRun_BaselineLoadError_NotQuiet(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, "invalid.json")

	// Create invalid baseline file
	if err := os.WriteFile(baselinePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  false, // NOT quiet - should print warning
	}

	opts := OrchestratorConfig{
		RootPath:       tmpDir,
		UseBaseline:    true,
		CreateBaseline: false,
		BaselinePath:   baselinePath,
	}

	orch := NewOrchestrator(cfg, opts)

	// Empty linter - just want to test baseline loading
	orch.WithLinters([]ComponentLinter{})

	// Should not error even with invalid baseline
	result, err := orch.Run()

	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if result == nil {
		t.Fatal("Run() returned nil result")
	}

	// Warning should be printed to stderr but not cause failure
}

// =============================================================================
// Test Run - baseline filtering summary output (non-quiet)
// =============================================================================

func TestRun_BaselineFilteringSummary_NotQuiet(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Create baseline with known issues
	knownIssues := []cue.ValidationError{
		{File: "test.md", Message: "Known error", Severity: "error", Source: "test"},
	}
	b := baseline.CreateBaseline(knownIssues)
	b.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := b.SaveBaseline(baselinePath); err != nil {
		t.Fatalf("Failed to create baseline: %v", err)
	}

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  false, // NOT quiet - should print summary
	}

	opts := OrchestratorConfig{
		RootPath:       tmpDir,
		UseBaseline:    true,
		CreateBaseline: false,
		BaselinePath:   baselinePath,
	}

	orch := NewOrchestrator(cfg, opts)

	linter := ComponentLinter{
		Name: "test-linter",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return &cli.LintSummary{
				TotalFiles:  1,
				TotalErrors: 1,
				Results: []cli.LintResult{
					{
						File:    "test.md",
						Success: false,
						Errors: []cue.ValidationError{
							{File: "test.md", Message: "Known error", Severity: "error", Source: "test"}, // Should be filtered
						},
					},
				},
			}, nil
		},
	}

	orch.WithLinters([]ComponentLinter{linter})

	result, err := orch.Run()

	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if result.BaselineIgnored == 0 {
		t.Error("BaselineIgnored = 0, want > 0")
	}

	// Should print baseline filtering summary (verified by coverage)
}

// =============================================================================
// Test Run - validation reminder output (non-quiet)
// =============================================================================

func TestRun_ValidationReminder_NotQuiet(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  false, // NOT quiet - should print reminder
	}

	opts := OrchestratorConfig{
		RootPath:     tmpDir,
		BaselinePath: ".cclintbaseline.json",
	}

	orch := NewOrchestrator(cfg, opts)

	successLinter := ComponentLinter{
		Name: "test-linter",
		Linter: func(rootPath string, quiet, verbose, noCycleCheck bool) (*cli.LintSummary, error) {
			return &cli.LintSummary{
				TotalFiles: 1,
				Results:    []cli.LintResult{{File: "test.md", Success: true}},
			}, nil
		},
	}

	orch.WithLinters([]ComponentLinter{successLinter})

	result, err := orch.Run()

	if err != nil {
		t.Fatalf("Run() error: %v", err)
	}

	if result == nil {
		t.Fatal("Run() returned nil result")
	}

	// Validation reminder should be printed (verified by coverage)
}

// =============================================================================
// Test saveBaseline - non-quiet output
// =============================================================================

func TestSaveBaseline_NotQuiet(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, "test_baseline.json")

	issues := []cue.ValidationError{
		{File: "test.md", Message: "Error", Severity: "error", Source: "test"},
	}

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  false, // NOT quiet - should print output
	}
	opts := OrchestratorConfig{
		RootPath:     tmpDir,
		BaselinePath: baselinePath,
	}

	orch := NewOrchestrator(cfg, opts)
	err := orch.saveBaseline(issues, baselinePath)

	if err != nil {
		t.Fatalf("saveBaseline() error: %v", err)
	}

	// Verify file was created
	if _, err := os.Stat(baselinePath); err != nil {
		t.Errorf("Baseline file not created: %v", err)
	}

	// Output should be printed (verified by coverage)
}

// =============================================================================
// Test runMemoryChecks (integration test)
// =============================================================================

func TestRunMemoryChecks(t *testing.T) {
	tmpDir := t.TempDir()

	// Create CLAUDE.local.md but don't add to .gitignore
	localMdPath := filepath.Join(tmpDir, "CLAUDE.local.md")
	if err := os.WriteFile(localMdPath, []byte("test"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	cfg := &config.Config{
		Root:   tmpDir,
		Format: "console",
		Quiet:  false, // Not quiet so checks run
	}

	opts := OrchestratorConfig{
		RootPath:     tmpDir,
		BaselinePath: ".cclintbaseline.json",
	}

	orch := NewOrchestrator(cfg, opts)

	// This should not panic or error
	orch.runMemoryChecks()
}

// =============================================================================
// Test runMemoryChecks - quiet mode
// =============================================================================

func TestRunMemoryChecks_QuietMode(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.Config{
		Root:  tmpDir,
		Quiet: true, // Quiet mode - checks should be skipped
	}

	opts := OrchestratorConfig{
		RootPath:     tmpDir,
		BaselinePath: ".cclintbaseline.json",
	}

	orch := NewOrchestrator(cfg, opts)

	// Should return early without errors
	orch.runMemoryChecks()
}

// =============================================================================
// Helper functions
// =============================================================================

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
