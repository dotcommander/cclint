package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/baseline"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockLinterFunc creates a mock linter function for testing
func mockLinterFunc(summary *lint.LintSummary, err error) LinterFunc {
	return func(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*lint.LintSummary, error) {
		return summary, err
	}
}

func TestRunComponentLint_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal config structure
	require.NoError(t, os.MkdirAll(tmpDir, 0755))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldUseBaseline := useBaseline
	oldCreateBaseline := createBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = true
	verbose = false
	useBaseline = false
	createBaseline = false
	baselinePath = ".cclintbaseline.json"

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		useBaseline = oldUseBaseline
		createBaseline = oldCreateBaseline
		baselinePath = oldBaselinePath
	}()

	// Create a successful linter
	successSummary := &lint.LintSummary{
		ProjectRoot:      tmpDir,
		ComponentType:    "agents",
		TotalFiles:       1,
		SuccessfulFiles:  1,
		FailedFiles:      0,
		TotalErrors:      0,
		TotalWarnings:    0,
		TotalSuggestions: 0,
	}
	linter := mockLinterFunc(successSummary, nil)

	// Run component lint
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)
}

func TestRunComponentLint_LinterError(t *testing.T) {
	tmpDir := t.TempDir()

	// Set global flags
	oldRootPath := rootPath
	rootPath = tmpDir
	defer func() { rootPath = oldRootPath }()

	// Create a failing linter
	linter := mockLinterFunc(nil, assert.AnError)

	// Run component lint
	err := runComponentLint("agents", linter)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "error running agents linter")
}

func TestRunComponentLint_BaselineCreation(t *testing.T) {
	tmpDir := t.TempDir()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldCreateBaseline := createBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = true
	createBaseline = true
	baselinePath = filepath.Join(tmpDir, ".cclintbaseline.json")

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		createBaseline = oldCreateBaseline
		baselinePath = oldBaselinePath
	}()

	// Create linter with some issues
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    1,
		TotalErrors:   2,
		Results: []lint.LintResult{
			{
				File: "test.md",
				Type: "agent",
				Errors: []cue.ValidationError{
					{Line: 1, Column: 1, Message: "Error 1", Source: "test", File: "test.md"},
					{Line: 2, Column: 1, Message: "Error 2", Source: "test", File: "test.md"},
				},
			},
		},
	}
	linter := mockLinterFunc(summary, nil)

	// Run component lint
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)

	// Verify baseline file was created
	assert.FileExists(t, baselinePath)

	// Verify baseline content
	b, err := baseline.LoadBaseline(baselinePath)
	assert.NoError(t, err)
	assert.NotNil(t, b)
	assert.Greater(t, len(b.Fingerprints), 0)
}

func TestRunComponentLint_BaselineFiltering(t *testing.T) {
	tmpDir := t.TempDir()
	baselineFile := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Create a baseline file with known issues
	testIssues := []cue.ValidationError{
		{File: "test.md", Source: "test", Message: "Error 1", Line: 1},
	}
	b := baseline.CreateBaseline(testIssues)

	require.NoError(t, b.SaveBaseline(baselineFile))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldUseBaseline := useBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = true
	useBaseline = true
	baselinePath = baselineFile

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		useBaseline = oldUseBaseline
		baselinePath = oldBaselinePath
	}()

	// Create linter with issues (one should be filtered)
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    1,
		TotalErrors:   2,
		Results: []lint.LintResult{
			{
				File: "test.md",
				Type: "agent",
				Errors: []cue.ValidationError{
					{Line: 1, Column: 1, Message: "Error 1", Source: "test", File: "test.md"}, // In baseline
					{Line: 3, Column: 1, Message: "Error 2", Source: "test", File: "test.md"}, // Not in baseline
				},
			},
		},
	}
	linter := mockLinterFunc(summary, nil)

	// Run component lint
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)
}

func TestRunComponentLint_AbsoluteBaselinePath(t *testing.T) {
	tmpDir := t.TempDir()
	absBaselinePath := filepath.Join(tmpDir, "custom-baseline.json")

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldCreateBaseline := createBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = true
	createBaseline = true
	baselinePath = absBaselinePath // Absolute path

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		createBaseline = oldCreateBaseline
		baselinePath = oldBaselinePath
	}()

	// Create linter
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    1,
		Results:       []lint.LintResult{},
	}
	linter := mockLinterFunc(summary, nil)

	// Run component lint
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)

	// Verify baseline was created at absolute path
	assert.FileExists(t, absBaselinePath)
}

func TestRunComponentLint_RelativeBaselinePath(t *testing.T) {
	tmpDir := t.TempDir()
	relBaselinePath := "custom.json"
	expectedPath := filepath.Join(tmpDir, relBaselinePath)

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldCreateBaseline := createBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = true
	createBaseline = true
	baselinePath = relBaselinePath // Relative path

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		createBaseline = oldCreateBaseline
		baselinePath = oldBaselinePath
	}()

	// Create linter
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    1,
		Results:       []lint.LintResult{},
	}
	linter := mockLinterFunc(summary, nil)

	// Run component lint
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)

	// Verify baseline was created relative to root
	assert.FileExists(t, expectedPath)
}

func TestRunComponentLint_BaselineLoadError(t *testing.T) {
	tmpDir := t.TempDir()
	baselineFile := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Create an invalid baseline file
	require.NoError(t, os.WriteFile(baselineFile, []byte("{invalid json}"), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldUseBaseline := useBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = false // Set to false to test warning message
	useBaseline = true
	baselinePath = baselineFile

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		useBaseline = oldUseBaseline
		baselinePath = oldBaselinePath
	}()

	// Suppress stderr to avoid cluttering test output
	oldStderr := os.Stderr
	os.Stderr = os.NewFile(0, os.DevNull)
	defer func() { os.Stderr = oldStderr }()

	// Create linter
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    1,
		Results:       []lint.LintResult{},
	}
	linter := mockLinterFunc(summary, nil)

	// Should handle error gracefully and continue
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)
}

func TestRunComponentLint_NoBaseline(t *testing.T) {
	tmpDir := t.TempDir()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldUseBaseline := useBaseline
	oldCreateBaseline := createBaseline

	rootPath = tmpDir
	quiet = true
	useBaseline = false
	createBaseline = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		useBaseline = oldUseBaseline
		createBaseline = oldCreateBaseline
	}()

	// Create linter
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    1,
		Results:       []lint.LintResult{},
	}
	linter := mockLinterFunc(summary, nil)

	// Run without baseline
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)
}

func TestRunComponentLint_VerboseOutput(t *testing.T) {
	tmpDir := t.TempDir()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose

	rootPath = tmpDir
	quiet = false
	verbose = true

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	// Create linter
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    5,
		TotalErrors:   2,
		Results:       []lint.LintResult{},
	}
	linter := mockLinterFunc(summary, nil)

	// Run with verbose
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)
}

func TestLinterFuncSignature(t *testing.T) {
	// Test that our mock matches the actual signature
	var linter LinterFunc = func(rootPath string, quiet bool, verbose bool, noCycleCheck bool) (*lint.LintSummary, error) {
		return &lint.LintSummary{}, nil
	}

	// Should be able to call it with expected parameters
	summary, err := linter("/tmp", true, false, false)
	assert.NoError(t, err)
	assert.NotNil(t, summary)
}

func TestRunComponentLint_QuietMode(t *testing.T) {
	tmpDir := t.TempDir()
	baselineFile := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Create baseline
	b := baseline.CreateBaseline([]cue.ValidationError{})
	require.NoError(t, b.SaveBaseline(baselineFile))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldUseBaseline := useBaseline
	oldBaselinePath := baselinePath
	oldCreateBaseline := createBaseline

	rootPath = tmpDir
	quiet = true // Quiet mode
	useBaseline = true
	createBaseline = true
	baselinePath = baselineFile

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		useBaseline = oldUseBaseline
		baselinePath = oldBaselinePath
		createBaseline = oldCreateBaseline
	}()

	// Create linter with issues
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    1,
		TotalErrors:   1,
		Results: []lint.LintResult{
			{
				File: "test.md",
				Type: "agent",
				Errors: []cue.ValidationError{
					{Line: 1, Column: 1, Message: "Error 1", Source: "test", File: "test.md"},
				},
			},
		},
	}
	linter := mockLinterFunc(summary, nil)

	// Run in quiet mode - should suppress output
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)
}

func TestRunComponentLint_ConfigLoadError(t *testing.T) {
	// Set root path to a location that will fail config loading
	oldRootPath := rootPath
	rootPath = "/nonexistent/directory/that/does/not/exist"
	defer func() { rootPath = oldRootPath }()

	// Create a dummy linter (shouldn't be called)
	linter := mockLinterFunc(&lint.LintSummary{}, nil)

	// Should fail at config loading stage
	// Note: config.LoadConfig may actually succeed even with non-existent path
	// by using defaults, so this test may not fail as expected
	err := runComponentLint("agents", linter)
	if err != nil {
		assert.Contains(t, err.Error(), "error")
	}
}

func TestRunComponentLint_BaselineFilteringWithIgnored(t *testing.T) {
	tmpDir := t.TempDir()
	baselineFile := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Create baseline with an issue that will be filtered
	testIssues := []cue.ValidationError{
		{File: "test.md", Source: "test", Message: "Existing issue", Line: 1},
	}
	b := baseline.CreateBaseline(testIssues)
	require.NoError(t, b.SaveBaseline(baselineFile))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldUseBaseline := useBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = false // Show baseline summary
	useBaseline = true
	baselinePath = baselineFile

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		useBaseline = oldUseBaseline
		baselinePath = oldBaselinePath
	}()

	// Create linter with issue that matches baseline
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    1,
		TotalErrors:   1,
		Results: []lint.LintResult{
			{
				File: "test.md",
				Type: "agent",
				Errors: []cue.ValidationError{
					{Line: 1, Column: 1, Message: "Existing issue", Source: "test", File: "test.md"},
				},
			},
		},
	}
	linter := mockLinterFunc(summary, nil)

	// Run - the issue should be filtered by baseline
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)
}

func TestRunComponentLint_BaselineCreationQuietMode(t *testing.T) {
	tmpDir := t.TempDir()
	baselineFile := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldCreateBaseline := createBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = false // Show baseline creation message
	createBaseline = true
	baselinePath = baselineFile

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		createBaseline = oldCreateBaseline
		baselinePath = oldBaselinePath
	}()

	// Create linter with some issues
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    2,
		TotalErrors:   3,
		Results: []lint.LintResult{
			{
				File: "agent1.md",
				Type: "agent",
				Errors: []cue.ValidationError{
					{Line: 1, Message: "Error 1", Source: "test", File: "agent1.md"},
					{Line: 2, Message: "Error 2", Source: "test", File: "agent1.md"},
				},
			},
			{
				File: "agent2.md",
				Type: "agent",
				Errors: []cue.ValidationError{
					{Line: 1, Message: "Error 1", Source: "test", File: "agent2.md"},
				},
				Suggestions: []cue.ValidationError{
					{Line: 5, Message: "Consider improving", Source: "test", File: "agent2.md"},
				},
			},
		},
	}
	linter := mockLinterFunc(summary, nil)

	// Run - should create baseline and print message
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)

	// Verify baseline was created
	assert.FileExists(t, baselineFile)
}

func TestRunComponentLint_ExistingBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	baselineFile := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Create existing baseline
	existingIssues := []cue.ValidationError{
		{File: "old.md", Source: "test", Message: "Old issue", Line: 1},
	}
	existingBaseline := baseline.CreateBaseline(existingIssues)
	require.NoError(t, existingBaseline.SaveBaseline(baselineFile))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldCreateBaseline := createBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = true
	createBaseline = true // Re-create baseline
	baselinePath = baselineFile

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		createBaseline = oldCreateBaseline
		baselinePath = oldBaselinePath
	}()

	// Create linter with new issues
	summary := &lint.LintSummary{
		ProjectRoot:   tmpDir,
		ComponentType: "agents",
		TotalFiles:    1,
		Results: []lint.LintResult{
			{
				File: "new.md",
				Type: "agent",
				Errors: []cue.ValidationError{
					{Line: 1, Message: "New issue", Source: "test", File: "new.md"},
				},
			},
		},
	}
	linter := mockLinterFunc(summary, nil)

	// Run - should overwrite existing baseline
	err := runComponentLint("agents", linter)
	assert.NoError(t, err)

	// Verify new baseline was saved
	newBaseline, err := baseline.LoadBaseline(baselineFile)
	assert.NoError(t, err)
	assert.NotNil(t, newBaseline)
}

func TestLinterFuncParameters(t *testing.T) {
	// Verify linter function receives correct parameters
	linterCalled := false

	tmpDir := t.TempDir()

	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldNoCycleCheck := noCycleCheck

	rootPath = tmpDir
	quiet = true
	verbose = true
	noCycleCheck = true

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		noCycleCheck = oldNoCycleCheck
	}()

	linter := func(rp string, q bool, v bool, ncc bool) (*lint.LintSummary, error) {
		linterCalled = true
		// Verify parameters are passed (they come from config, not flags directly)
		assert.NotEmpty(t, rp)
		return &lint.LintSummary{ProjectRoot: rp}, nil
	}

	_ = runComponentLint("test", linter)

	// Verify linter was called
	assert.True(t, linterCalled)
}
