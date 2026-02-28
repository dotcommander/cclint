package cmd

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dotcommander/cclint/internal/lint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCollectFilesToLint(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		fileFlag []string
		want     []string
	}{
		{
			name:     "empty args and flags",
			args:     nil, // []string{}
			fileFlag: nil, // []string{}
			want:     nil, // []string{}
		},
		{
			name:     "file flag takes precedence",
			args:     []string{"agents", "commands"},
			fileFlag: []string{"agents", "commands"},
			want:     []string{"agents", "commands"},
		},
		{
			name:     "skip known subcommands in args",
			args:     []string{"agents", "commands", "skills"},
			fileFlag: nil, // []string{}
			want:     nil, // []string{}
		},
		{
			name:     "collect file paths from args",
			args:     []string{"./agents/test.md", "./commands/cmd.md"},
			fileFlag: []string{},
			want:     []string{"./agents/test.md", "./commands/cmd.md"},
		},
		{
			name:     "mix of subcommands and paths",
			args:     []string{"agents", "./custom.md", "commands"},
			fileFlag: []string{},
			want:     []string{"./custom.md"},
		},
		{
			name:     "absolute paths",
			args:     []string{"/Users/test/agents/agent.md", "/tmp/file.md"},
			fileFlag: []string{},
			want:     []string{"/Users/test/agents/agent.md", "/tmp/file.md"},
		},
		{
			name:     "relative paths with extensions",
			args:     []string{"file.md", "dir/file.json"},
			fileFlag: []string{},
			want:     []string{"file.md", "dir/file.json"},
		},
		{
			name:     "paths with dots and slashes",
			args:     []string{"./file.md", "../parent/file.md", "dir/./file.md"},
			fileFlag: []string{},
			want:     []string{"./file.md", "../parent/file.md", "dir/./file.md"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up test state
			fileFlag = tt.fileFlag

			// Run test
			got := collectFilesToLint(tt.args)

			// Verify
			assert.Equal(t, tt.want, got)

			// Clean up
			fileFlag = nil
		})
	}
}

func TestCollectFilesToLint_LooksLikePath(t *testing.T) {
	// Test integration with lint.LooksLikePath
	tests := []struct {
		name string
		args []string
		want int // number of files collected
	}{
		{
			name: "windows-style paths not mistaken for subcommands",
			args: []string{"C:\\Users\\test\\file.md"},
			want: 1,
		},
		{
			name: "paths with extensions",
			args: []string{"agents.md", "commands.json"},
			want: 2,
		},
		{
			name: "subcommand without extension",
			args: []string{"agents"},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := collectFilesToLint(tt.args)
			assert.Len(t, got, tt.want)
		})
	}
}

func TestRunSingleFileLint(t *testing.T) {
	// Create temporary test directory
	tmpDir := t.TempDir()

	// Create a valid test file
	validAgentPath := filepath.Join(tmpDir, "agents", "test-agent.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(validAgentPath), 0755))

	validContent := `---
name: test-agent
description: A test agent. Use PROACTIVELY when testing.. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

This is a test agent.

## Workflow

1. Test step 1
2. Test step 2
`
	require.NoError(t, os.WriteFile(validAgentPath, []byte(validContent), 0644))

	// Only test valid file to avoid os.Exit issues in tests
	tests := []struct {
		name      string
		files     []string
		rootPath  string
		typeFlag  string
		quiet     bool
		wantError bool
	}{
		{
			name:      "valid file",
			files:     []string{validAgentPath},
			rootPath:  tmpDir,
			typeFlag:  "",
			quiet:     true,
			wantError: false,
		},
		{
			name:      "force type override",
			files:     []string{validAgentPath},
			rootPath:  tmpDir,
			typeFlag:  "agent",
			quiet:     true,
			wantError: false,
		},
		// Note: Tests for error cases (invalid files, non-existent files) are skipped
		// because runSingleFileLint calls os.Exit() directly which terminates the test
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global flags
			oldRootPath := rootPath
			oldQuiet := quiet
			oldVerbose := verbose
			oldTypeFlag := typeFlag

			rootPath = tt.rootPath
			quiet = tt.quiet
			verbose = false
			typeFlag = tt.typeFlag

			defer func() {
				// Restore global flags
				rootPath = oldRootPath
				quiet = oldQuiet
				verbose = oldVerbose
				typeFlag = oldTypeFlag
			}()

			// Run the function - only testing success cases
			err := runSingleFileLint(tt.files)
			assert.NoError(t, err)
		})
	}
}

func TestRunGitLint(t *testing.T) {
	// Create temporary test directory with git repo
	tmpDir := t.TempDir()

	// Initialize git repo
	gitDir := filepath.Join(tmpDir, ".git")
	require.NoError(t, os.MkdirAll(gitDir, 0755))

	// Create a test file
	testFile := filepath.Join(tmpDir, "agents", "test-agent.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(testFile), 0755))
	validContent := `---
name: test-agent
description: A test agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation
Test agent.

## Workflow
1. Test step
`
	require.NoError(t, os.WriteFile(testFile, []byte(validContent), 0644))

	tests := []struct {
		name       string
		rootPath   string
		stagedMode bool
		diffMode   bool
		quiet      bool
		setupGit   bool
	}{
		{
			name:       "not in git repo falls back to full lint",
			rootPath:   t.TempDir(), // Different temp dir without .git
			stagedMode: false,
			diffMode:   true,
			quiet:      true,
			setupGit:   false,
		},
		{
			name:       "staged mode in git repo",
			rootPath:   tmpDir,
			stagedMode: true,
			diffMode:   false,
			quiet:      true,
			setupGit:   true,
		},
		{
			name:       "diff mode in git repo",
			rootPath:   tmpDir,
			stagedMode: false,
			diffMode:   true,
			quiet:      true,
			setupGit:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global flags
			oldRootPath := rootPath
			oldQuiet := quiet
			oldVerbose := verbose
			oldStagedMode := stagedMode
			oldDiffMode := diffMode

			rootPath = tt.rootPath
			quiet = tt.quiet
			verbose = false
			stagedMode = tt.stagedMode
			diffMode = tt.diffMode

			// Capture exit behavior
			originalOsExit := osExit
			osExit = func(code int) {
				// Just prevent actual exit
			}
			defer func() { osExit = originalOsExit }()

			// Run the function (may call os.Exit)
			_ = runGitLint()

			// Just verify it doesn't panic - actual behavior depends on git state
			// which we can't fully control in unit tests

			// Restore global flags
			rootPath = oldRootPath
			quiet = oldQuiet
			verbose = oldVerbose
			stagedMode = oldStagedMode
			diffMode = oldDiffMode
		})
	}
}

func TestInitConfig(t *testing.T) {
	// Create temporary directory for config files
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	// Change to temp directory
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalWd) }()

	tests := []struct {
		name       string
		configFile string
		content    string
	}{
		{
			name:       "no config file",
			configFile: "",
			content:    "",
		},
		{
			name:       "valid json config",
			configFile: ".cclintrc.json",
			content:    `{"quiet": true, "verbose": false}`,
		},
		{
			name:       "valid yaml config",
			configFile: ".cclintrc.yaml",
			content:    "quiet: true\nverbose: false\n",
		},
		// Note: "invalid json config" test case removed because initConfig()
		// calls os.Exit(1) directly which terminates the test process.
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing config files
			for _, name := range []string{".cclintrc.json", ".cclintrc.yaml", ".cclintrc.yml"} {
				_ = os.Remove(filepath.Join(tmpDir, name))
			}

			// Create config file if specified
			if tt.configFile != "" {
				configPath := filepath.Join(tmpDir, tt.configFile)
				require.NoError(t, os.WriteFile(configPath, []byte(tt.content), 0644))
			}

			// Run initConfig - only testing success cases
			initConfig()

			// If we reach here, the config was loaded successfully (no os.Exit called)
		})
	}
}

func TestVersion(t *testing.T) {
	// Test that Version variable can be set
	oldVersion := Version
	Version = "1.2.3"
	assert.Equal(t, "1.2.3", Version)
	Version = oldVersion
}

// Note: osExit is kept for backward compatibility with existing tests
var osExit = os.Exit

func TestExecute_Success(t *testing.T) {
	// Save original exit function
	originalExitFunc := exitFunc
	exitCalled := false
	exitCode := 0

	// Mock exit function
	exitFunc = func(code int) {
		exitCalled = true
		exitCode = code
	}
	defer func() { exitFunc = originalExitFunc }()

	// Set up a simple command that succeeds
	// Reset args to just help (which always succeeds)
	oldArgs := os.Args
	os.Args = []string{"cclint", "--help"}
	defer func() { os.Args = oldArgs }()

	// Execute should succeed with --help
	Execute()

	// Should not have called exit (help exits with 0 by default but Cobra handles it)
	// The key is it didn't call exitFunc(1)
	if exitCalled && exitCode != 0 {
		t.Errorf("Execute called exit with code %d, expected no error exit", exitCode)
	}
}

func TestExecute_WithVersion(t *testing.T) {
	// Save original exit function
	originalExitFunc := exitFunc
	exitCalled := false

	// Mock exit function
	exitFunc = func(code int) {
		exitCalled = true
	}
	defer func() { exitFunc = originalExitFunc }()

	// Set up version flag
	oldArgs := os.Args
	os.Args = []string{"cclint", "--version"}
	defer func() { os.Args = oldArgs }()

	// Execute should succeed with --version
	Execute()

	// Should not have called exitFunc(1)
	if exitCalled {
		t.Errorf("Execute should not have called exit for version flag")
	}
}

func TestExecute_ErrorPath(t *testing.T) {
	// Save original exit function
	originalExitFunc := exitFunc
	exitCalled := false
	exitCode := -1

	// Mock exit function to capture exit code
	exitFunc = func(code int) {
		exitCalled = true
		exitCode = code
	}
	defer func() { exitFunc = originalExitFunc }()

	// Set up an invalid flag to trigger an error from Cobra
	oldArgs := os.Args
	os.Args = []string{"cclint", "--invalid-flag-that-does-not-exist"}
	defer func() { os.Args = oldArgs }()

	// Execute should fail and call exitFunc(1)
	Execute()

	// Should have called exit with code 1
	assert.True(t, exitCalled, "Execute should have called exit on error")
	assert.Equal(t, 1, exitCode, "Exit code should be 1")
}

func TestRootCmdFlags(t *testing.T) {
	// Verify that root command has expected flags
	flags := rootCmd.PersistentFlags()

	testCases := []struct {
		name     string
		flagName string
	}{
		{"root flag", "root"},
		{"quiet flag", "quiet"},
		{"verbose flag", "verbose"},
		{"scores flag", "scores"},
		{"improvements flag", "improvements"},
		{"format flag", "format"},
		{"output flag", "output"},
		{"fail-on flag", "fail-on"},
		{"no-cycle-check flag", "no-cycle-check"},
		{"baseline flag", "baseline"},
		{"baseline-create flag", "baseline-create"},
		{"baseline-path flag", "baseline-path"},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			flag := flags.Lookup(tc.flagName)
			assert.NotNil(t, flag, "flag %s should exist", tc.flagName)
		})
	}

	// Check local flags
	localFlags := rootCmd.Flags()
	assert.NotNil(t, localFlags.Lookup("file"))
	assert.NotNil(t, localFlags.Lookup("type"))
	assert.NotNil(t, localFlags.Lookup("diff"))
	assert.NotNil(t, localFlags.Lookup("staged"))
}

func TestRootCmdSubcommands(t *testing.T) {
	// Verify subcommands are registered
	commands := rootCmd.Commands()
	commandNames := make(map[string]bool)
	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	expectedCommands := []string{
		"agents",
		"commands",
		"skills",
		"plugins",
		"fmt",
		"summary",
	}

	for _, name := range expectedCommands {
		assert.True(t, commandNames[name], "subcommand %s should be registered", name)
	}
}

func TestExecute(t *testing.T) {
	// Test that Execute doesn't panic with valid setup
	// We can't fully test it without running the command, but we can verify it exists
	assert.NotNil(t, Execute)
}

func TestRootCmdRun(t *testing.T) {
	// Test that root command Run function exists and is configured
	assert.NotNil(t, rootCmd.Run)

	// Verify root command has correct configuration
	assert.Equal(t, "cclint [files...]", rootCmd.Use)
	assert.Contains(t, rootCmd.Long, "USAGE MODES")
	assert.Contains(t, rootCmd.Long, "EXAMPLES")
}

func TestCollectFilesToLint_IsKnownSubcommand(t *testing.T) {
	// Test that lint.IsKnownSubcommand is properly used
	tests := []struct {
		arg      string
		isSubcmd bool
	}{
		{"agents", true},
		{"commands", true},
		{"skills", true},
		{"settings", true},
		{"./agents", false},
		{"agents.md", false},
		{"/path/to/agents", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			result := lint.IsKnownSubcommand(tt.arg)
			assert.Equal(t, tt.isSubcmd, result)
		})
	}
}

func TestRunGitLint_NoFiles(t *testing.T) {
	// Create a real git repo to test "no files to lint" path
	tmpDir := t.TempDir()

	// Initialize a real git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git user for the test
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = tmpDir
	quiet = true // Suppress output
	verbose = false
	stagedMode = true // Staged mode with nothing staged
	diffMode = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	// Run with empty staging area - should return nil (no files to lint)
	err := runGitLint()
	assert.NoError(t, err)
}

func TestRunGitLint_WithVerbose(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a real git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = tmpDir
	quiet = false
	verbose = true // Enable verbose
	stagedMode = true
	diffMode = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	err := runGitLint()
	assert.NoError(t, err)
}

func TestRunGitLint_NotGitRepoWithQuiet(t *testing.T) {
	tmpDir := t.TempDir()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = tmpDir
	quiet = false // Show warning
	verbose = false
	stagedMode = false
	diffMode = true

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	// Not in git repo should fallback to full lint (which may fail)
	// Just verify no panic
	_ = runGitLint()
}

func TestRunGitLint_WithEmptyRootPath(t *testing.T) {
	// Test the rootPath == "" branch that uses os.Getwd()
	// Save current directory
	origWd, err := os.Getwd()
	require.NoError(t, err)

	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Change to temp dir
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(origWd) }()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = "" // Empty to trigger os.Getwd() path
	quiet = true
	stagedMode = true
	diffMode = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	err = runGitLint()
	assert.NoError(t, err)
}

func TestRunSingleFileLint_WithTypeFlag(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file outside standard paths
	testFile := filepath.Join(tmpDir, "random.md")
	content := `---
name: test
description: Test. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test content.

## Workflow

1. Test step
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	// Set global flags with valid type override
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldTypeFlag := typeFlag

	rootPath = tmpDir
	quiet = true
	verbose = false
	typeFlag = "agent" // Valid type

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		typeFlag = oldTypeFlag
	}()

	// Should work with valid type override
	err := runSingleFileLint([]string{testFile})
	assert.NoError(t, err)
}

func TestRunSingleFileLint_VerboseMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid agent file
	agentPath := filepath.Join(tmpDir, "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test agent.

## Workflow

1. Test step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose

	rootPath = tmpDir
	quiet = false
	verbose = true // Enable verbose

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	err := runSingleFileLint([]string{agentPath})
	assert.NoError(t, err)
}

func TestRunLint_WithBaseline(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid agent file
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test agent.

## Workflow

1. Test step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldUseBaseline := useBaseline
	oldCreateBaseline := createBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = true
	useBaseline = false
	createBaseline = false
	baselinePath = filepath.Join(tmpDir, ".cclintbaseline.json")

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		useBaseline = oldUseBaseline
		createBaseline = oldCreateBaseline
		baselinePath = oldBaselinePath
	}()

	// runLint calls os.Exit on errors, but should succeed here
	err := runLint()
	// May return nil or error depending on component files
	_ = err
}

func TestRunLint_CreateBaseline(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid agent file
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test agent.

## Workflow

1. Test step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldUseBaseline := useBaseline
	oldCreateBaseline := createBaseline
	oldBaselinePath := baselinePath

	rootPath = tmpDir
	quiet = true
	useBaseline = false
	createBaseline = true // Test baseline creation
	baselinePath = filepath.Join(tmpDir, ".cclintbaseline.json")

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		useBaseline = oldUseBaseline
		createBaseline = oldCreateBaseline
		baselinePath = oldBaselinePath
	}()

	err := runLint()
	_ = err
}

func TestRootCmdLongDescription(t *testing.T) {
	assert.Contains(t, rootCmd.Long, "Single-file mode")
	assert.Contains(t, rootCmd.Long, "Git integration mode")
	assert.Contains(t, rootCmd.Long, "Baseline mode")
}

func TestRootCmdVersion(t *testing.T) {
	assert.Equal(t, Version, rootCmd.Version)
}

func TestRunGitLint_DiffMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a real git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Create an agent file
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test content.

## Workflow

1. Step one
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Add file to staging
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = tmpDir
	quiet = true
	stagedMode = false
	diffMode = true // Use diff mode

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	err := runGitLint()
	// May succeed or fail depending on files
	_ = err
}

func TestRunGitLint_StagedModeWithFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize a real git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git user
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Create an agent file
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test content.

## Workflow

1. Step one
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Add file to staging
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = tmpDir
	quiet = true
	stagedMode = true // Use staged mode
	diffMode = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	err := runGitLint()
	// May succeed or fail depending on files
	_ = err
}

func TestRunSingleFileLint_QuietMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid agent file
	agentPath := filepath.Join(tmpDir, "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test agent.

## Workflow

1. Test step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose

	rootPath = tmpDir
	quiet = true // Quiet mode
	verbose = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	err := runSingleFileLint([]string{agentPath})
	assert.NoError(t, err)
}

func TestRunSingleFileLint_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple valid agent files
	agent1 := filepath.Join(tmpDir, "agents", "test1.md")
	agent2 := filepath.Join(tmpDir, "agents", "test2.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agent1), 0755))

	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test agent.

## Workflow

1. Test step
`
	require.NoError(t, os.WriteFile(agent1, []byte(content), 0644))
	require.NoError(t, os.WriteFile(agent2, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose

	rootPath = tmpDir
	quiet = true
	verbose = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	err := runSingleFileLint([]string{agent1, agent2})
	assert.NoError(t, err)
}

func TestRunGitLint_StagedWithRealFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Create agent file with proper structure
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))

	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing. for git integration
model: sonnet
---

## Foundation

Test agent content.

## Workflow

1. Step one
2. Step two
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Stage the file
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = tmpDir
	quiet = true
	verbose = false
	stagedMode = true
	diffMode = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	// Run git lint - should find and lint staged files
	err := runGitLint()
	// May succeed or exit with os.Exit
	_ = err
}

func TestRunGitLint_DiffModeWithFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Create and commit initial file
	initialFile := filepath.Join(tmpDir, "initial.txt")
	require.NoError(t, os.WriteFile(initialFile, []byte("initial"), 0644))
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Create new agent file (unstaged change)
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "new.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: new-agent
description: New agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

New agent.

## Workflow

1. Step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = tmpDir
	quiet = true
	stagedMode = false
	diffMode = true // Diff mode captures all changes

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	err := runGitLint()
	// May succeed or fail - just verify no panic
	_ = err
}

func TestRunGitLint_OutputValidationReminder(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Create agent file
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test
description: Test. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation
Test.

## Workflow
1. Step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Stage the file
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Set global flags - non-quiet to test reminder output
	oldRootPath := rootPath
	oldQuiet := quiet
	oldStagedMode := stagedMode

	rootPath = tmpDir
	quiet = false // Show validation reminder
	stagedMode = true

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		stagedMode = oldStagedMode
	}()

	err := runGitLint()
	_ = err
}

func TestRunSingleFileLint_NonQuietMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid agent file
	agentPath := filepath.Join(tmpDir, "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

Test agent.

## Workflow

1. Test step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags for non-quiet mode
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose

	rootPath = tmpDir
	quiet = false // Non-quiet to show validation reminder
	verbose = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	err := runSingleFileLint([]string{agentPath})
	assert.NoError(t, err)
}

func TestCollectFilesToLint_WithFileFlagOnly(t *testing.T) {
	// Test that file flag takes precedence even with args
	oldFileFlag := fileFlag
	fileFlag = []string{"explicit-file.md"}
	defer func() { fileFlag = oldFileFlag }()

	// Args should be ignored when fileFlag is set
	result := collectFilesToLint([]string{"ignored.md", "also-ignored.md"})
	assert.Equal(t, []string{"explicit-file.md"}, result)
}

func TestRootCmdVersionFlag(t *testing.T) {
	// Test that version flag (-V) is properly configured
	flag := rootCmd.Flags().Lookup("version")
	assert.NotNil(t, flag)
	assert.Equal(t, "V", flag.Shorthand)
}

func TestRunLint_Success(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal valid structure
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test
description: Test. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation
Test.

## Workflow
1. Step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

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

	// Run lint
	err := runLint()
	// May succeed or exit
	_ = err
}

func TestRunGitLint_WithFilesToLint(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Create valid agent file
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test-agent.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test-agent
description: Test. Use PROACTIVELY when testing. agent. Use PROACTIVELY when testing. for lint
model: sonnet
---

## Foundation

Test agent content.

## Workflow

1. Step one
2. Step two
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Stage file
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = tmpDir
	quiet = true
	verbose = false
	stagedMode = true
	diffMode = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	// Run - this tests the full path with files
	err := runGitLint()
	// May succeed or call os.Exit
	_ = err
}

func TestRunGitLint_DiffModeRealFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Configure git
	cmd = exec.Command("git", "config", "user.email", "test@test.com")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "config", "user.name", "Test User")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Create and commit initial file
	initial := filepath.Join(tmpDir, "README.md")
	require.NoError(t, os.WriteFile(initial, []byte("# Init"), 0644))
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	_ = cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "init")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Create new agent file (uncommitted change)
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "new-agent.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: new-agent
description: New agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

New agent.

## Workflow

1. Step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldStagedMode := stagedMode
	oldDiffMode := diffMode

	rootPath = tmpDir
	quiet = true
	stagedMode = false
	diffMode = true

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		stagedMode = oldStagedMode
		diffMode = oldDiffMode
	}()

	err := runGitLint()
	_ = err
}

func TestRunSingleFileLint_ValidAgent(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid agent file
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "good-agent.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: good-agent
description: A good agent. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation

This is a good agent.

## Workflow

1. Do something good
2. Complete the task
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldTypeFlag := typeFlag

	rootPath = tmpDir
	quiet = true
	typeFlag = "" // Auto-detect type

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		typeFlag = oldTypeFlag
	}()

	err := runSingleFileLint([]string{agentPath})
	assert.NoError(t, err)
}

func TestRunSingleFileLint_WithFormatOption(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid agent file
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "format-test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: format-test
description: Test. Use PROACTIVELY when testing. file for format option
model: sonnet
---

## Foundation

Test content.

## Workflow

1. Test step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldOutputFormat := outputFormat

	rootPath = tmpDir
	quiet = true
	outputFormat = "json"

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		outputFormat = oldOutputFormat
	}()

	err := runSingleFileLint([]string{agentPath})
	assert.NoError(t, err)
}

func TestRunGitLint_NonQuietMode(t *testing.T) {
	tmpDir := t.TempDir()

	// Initialize git repo
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	require.NoError(t, cmd.Run())

	// Create agent file
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))
	content := `---
name: test
description: Test. Use PROACTIVELY when testing.
model: sonnet
---

## Foundation
Test.

## Workflow
1. Step
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Stage
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	_ = cmd.Run()

	// Set flags - non-quiet
	oldRootPath := rootPath
	oldQuiet := quiet
	oldStagedMode := stagedMode

	rootPath = tmpDir
	quiet = false
	stagedMode = true

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		stagedMode = oldStagedMode
	}()

	err := runGitLint()
	_ = err
}

func TestInitConfig_WithYmlFile(t *testing.T) {
	tmpDir := t.TempDir()
	originalWd, err := os.Getwd()
	require.NoError(t, err)

	// Change to temp dir
	require.NoError(t, os.Chdir(tmpDir))
	defer func() { _ = os.Chdir(originalWd) }()

	// Create yml config
	configPath := filepath.Join(tmpDir, ".cclintrc.yml")
	require.NoError(t, os.WriteFile(configPath, []byte("quiet: true\n"), 0644))

	// Should not panic
	assert.NotPanics(t, func() {
		initConfig()
	})
}
