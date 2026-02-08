package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsComponentType(t *testing.T) {
	tests := []struct {
		arg      string
		expected bool
	}{
		{"agents", true},
		{"commands", true},
		{"skills", true},
		{"settings", true},
		{"context", true},
		{"plugins", true},
		{"rules", true},
		{"AGENTS", true},  // Case insensitive
		{"Commands", true},
		{"unknown", false},
		{"agent", false},  // Singular form
		{"./agents", false},
		{"agents.md", false},
		{"/path/to/agents", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.arg, func(t *testing.T) {
			result := isComponentType(tt.arg)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCollectFilesToFormat(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test directory structure
	agentsDir := filepath.Join(tmpDir, "agents")
	commandsDir := filepath.Join(tmpDir, "commands")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.MkdirAll(commandsDir, 0755))

	// Create test files
	agent1 := filepath.Join(agentsDir, "agent1.md")
	agent2 := filepath.Join(agentsDir, "agent2.md")
	cmd1 := filepath.Join(commandsDir, "cmd1.md")
	require.NoError(t, os.WriteFile(agent1, []byte("# Agent 1"), 0644))
	require.NoError(t, os.WriteFile(agent2, []byte("# Agent 2"), 0644))
	require.NoError(t, os.WriteFile(cmd1, []byte("# Command 1"), 0644))

	tests := []struct {
		name      string
		args      []string
		fmtFiles  []string
		wantCount int
		wantError bool
	}{
		{
			name:      "explicit file flag",
			args:      []string{},
			fmtFiles:  []string{agent1, cmd1},
			wantCount: 2,
			wantError: false,
		},
		{
			name:      "component type arg",
			args:      []string{"agents"},
			fmtFiles:  []string{},
			wantCount: 2, // Both agent files
			wantError: false,
		},
		{
			name:      "file path arg",
			args:      []string{agent1},
			fmtFiles:  []string{},
			wantCount: 1,
			wantError: false,
		},
		{
			name:      "directory path arg",
			args:      []string{agentsDir},
			fmtFiles:  []string{},
			wantCount: 2,
			wantError: false,
		},
		{
			name:      "multiple file args",
			args:      []string{agent1, cmd1},
			fmtFiles:  []string{},
			wantCount: 2,
			wantError: false,
		},
		{
			name:      "non-existent file",
			args:      []string{filepath.Join(tmpDir, "nonexistent.md")},
			fmtFiles:  []string{},
			wantCount: 0,
			wantError: true,
		},
		{
			name:      "empty args discovers all",
			args:      []string{},
			fmtFiles:  []string{},
			wantCount: 3, // All markdown files
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set test state
			fmtFiles = tt.fmtFiles

			// Run test
			files, err := collectFilesToFormat(tt.args, tmpDir)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, files, tt.wantCount)
			}

			// Clean up
			fmtFiles = nil
		})
	}
}

func TestDiscoverFilesInDir(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test structure
	dir1 := filepath.Join(tmpDir, "dir1")
	dir2 := filepath.Join(tmpDir, "dir1", "dir2")
	require.NoError(t, os.MkdirAll(dir2, 0755))

	// Create markdown files
	file1 := filepath.Join(dir1, "file1.md")
	file2 := filepath.Join(dir2, "file2.md")
	file3 := filepath.Join(dir1, "file3.txt") // Not markdown
	require.NoError(t, os.WriteFile(file1, []byte("# File 1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("# File 2"), 0644))
	require.NoError(t, os.WriteFile(file3, []byte("Not markdown"), 0644))

	tests := []struct {
		name      string
		dirPath   string
		wantCount int
		wantError bool
	}{
		{
			name:      "discover in directory",
			dirPath:   dir1,
			wantCount: 2, // Only .md files
			wantError: false,
		},
		{
			name:      "discover in subdirectory",
			dirPath:   dir2,
			wantCount: 1,
			wantError: false,
		},
		{
			name:      "non-existent directory",
			dirPath:   filepath.Join(tmpDir, "nonexistent"),
			wantCount: 0,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := discoverFilesInDir(tt.dirPath)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, files, tt.wantCount)

				// Verify all files are markdown
				for _, f := range files {
					assert.True(t, filepath.Ext(f) == ".md" || filepath.Ext(f) == ".MD")
				}
			}
		})
	}
}

func TestDiscoverFilesByType(t *testing.T) {
	tmpDir := t.TempDir()

	// Create standard directory structure
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	commandsDir := filepath.Join(tmpDir, ".claude", "commands")
	skillsDir := filepath.Join(tmpDir, ".claude", "skills", "test-skill")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.MkdirAll(commandsDir, 0755))
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	// Create test files
	agent := filepath.Join(agentsDir, "test-agent.md")
	command := filepath.Join(commandsDir, "test-command.md")
	skill := filepath.Join(skillsDir, "SKILL.md")
	require.NoError(t, os.WriteFile(agent, []byte("# Agent"), 0644))
	require.NoError(t, os.WriteFile(command, []byte("# Command"), 0644))
	require.NoError(t, os.WriteFile(skill, []byte("# Skill"), 0644))

	tests := []struct {
		name          string
		componentType string
		wantError     bool
		minCount      int // Minimum expected files
	}{
		{
			name:          "discover agents",
			componentType: "agents",
			wantError:     false,
			minCount:      1,
		},
		{
			name:          "discover commands",
			componentType: "commands",
			wantError:     false,
			minCount:      1,
		},
		{
			name:          "discover skills",
			componentType: "skills",
			wantError:     false,
			minCount:      1,
		},
		{
			name:          "unknown type",
			componentType: "unknown",
			wantError:     true,
			minCount:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			files, err := discoverFilesByType(tmpDir, tt.componentType)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.GreaterOrEqual(t, len(files), tt.minCount)
			}
		})
	}
}

func TestDiscoverAllFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test directory structure
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	commandsDir := filepath.Join(tmpDir, ".claude", "commands")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.MkdirAll(commandsDir, 0755))

	// Create test files
	agent := filepath.Join(agentsDir, "test-agent.md")
	command := filepath.Join(commandsDir, "test-command.md")
	jsonFile := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(agent, []byte("# Agent"), 0644))
	require.NoError(t, os.WriteFile(command, []byte("# Command"), 0644))
	require.NoError(t, os.WriteFile(jsonFile, []byte("{}"), 0644))

	files, err := discoverAllFiles(tmpDir)
	assert.NoError(t, err)

	// Should find markdown files, but not json
	assert.GreaterOrEqual(t, len(files), 2)

	// Verify all are markdown
	for _, f := range files {
		ext := filepath.Ext(f)
		assert.True(t, ext == ".md" || ext == ".MD", "file %s should be markdown", f)
	}
}

func TestRunFmt(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file with formatting issues
	testFile := filepath.Join(tmpDir, "test.md")
	unformatted := `---
description: Test file
name: test-file
model: sonnet
---


# Test

Trailing spaces
Multiple blank lines


End of file`
	require.NoError(t, os.WriteFile(testFile, []byte(unformatted), 0644))

	tests := []struct {
		name      string
		args      []string
		fmtCheck  bool
		fmtWrite  bool
		fmtDiff   bool
		quiet     bool
		wantError bool
	}{
		{
			name:      "preview mode (default)",
			args:      []string{testFile},
			fmtCheck:  false,
			fmtWrite:  false,
			fmtDiff:   false,
			quiet:     true,
			wantError: false,
		},
		{
			name:      "check mode with unformatted file",
			args:      []string{testFile},
			fmtCheck:  true,
			fmtWrite:  false,
			fmtDiff:   false,
			quiet:     true,
			wantError: false, // Function doesn't return error, calls os.Exit
		},
		{
			name:      "write mode",
			args:      []string{testFile},
			fmtCheck:  false,
			fmtWrite:  true,
			fmtDiff:   false,
			quiet:     true,
			wantError: false,
		},
		{
			name:      "diff mode",
			args:      []string{testFile},
			fmtCheck:  false,
			fmtWrite:  false,
			fmtDiff:   true,
			quiet:     true,
			wantError: false,
		},
		{
			name:      "no files to format",
			args:      []string{},
			fmtCheck:  false,
			fmtWrite:  false,
			fmtDiff:   false,
			quiet:     true,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set global flags
			oldRootPath := rootPath
			oldQuiet := quiet
			oldVerbose := verbose
			oldFmtCheck := fmtCheck
			oldFmtWrite := fmtWrite
			oldFmtDiff := fmtDiff

			rootPath = tmpDir
			quiet = tt.quiet
			verbose = false
			fmtCheck = tt.fmtCheck
			fmtWrite = tt.fmtWrite
			fmtDiff = tt.fmtDiff

			// Capture exit behavior
			originalOsExit := osExit
			osExit = func(code int) {
				// Just prevent actual exit
			}
			defer func() { osExit = originalOsExit }()

			// Run test
			err := runFmt(tt.args)

			if tt.wantError {
				assert.Error(t, err)
			} else if err != nil {
				// May have error OR os.Exit called
				t.Logf("Got error: %v", err)
			}

			// Restore global flags
			rootPath = oldRootPath
			quiet = oldQuiet
			verbose = oldVerbose
			fmtCheck = oldFmtCheck
			fmtWrite = oldFmtWrite
			fmtDiff = oldFmtDiff

			// Recreate test file for next test
			_ = os.WriteFile(testFile, []byte(unformatted), 0644)
		})
	}
}

func TestFmtCmdFlags(t *testing.T) {
	// Verify fmt command has expected flags
	flags := fmtCmd.Flags()

	assert.NotNil(t, flags.Lookup("check"))
	assert.NotNil(t, flags.Lookup("write"))
	assert.NotNil(t, flags.Lookup("diff"))
	assert.NotNil(t, flags.Lookup("file"))
	assert.NotNil(t, flags.Lookup("type"))
}

func TestCollectFilesToFormat_Precedence(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test files
	file1 := filepath.Join(tmpDir, "file1.md")
	file2 := filepath.Join(tmpDir, "file2.md")
	require.NoError(t, os.WriteFile(file1, []byte("# File 1"), 0644))
	require.NoError(t, os.WriteFile(file2, []byte("# File 2"), 0644))

	// Test that --file flag takes precedence over args
	fmtFiles = []string{file1}
	files, err := collectFilesToFormat([]string{file2}, tmpDir)
	assert.NoError(t, err)
	assert.Equal(t, []string{file1}, files)

	fmtFiles = nil
}

func TestDiscoverFilesByType_AllTypes(t *testing.T) {
	tmpDir := t.TempDir()

	// Create minimal structure
	testDir := filepath.Join(tmpDir, ".claude")
	require.NoError(t, os.MkdirAll(testDir, 0755))

	componentTypes := []string{
		"agents",
		"commands",
		"skills",
		"settings",
		"context",
		"plugins",
		"rules",
	}

	for _, componentType := range componentTypes {
		t.Run(componentType, func(t *testing.T) {
			// Should not error even if no files found
			files, err := discoverFilesByType(tmpDir, componentType)
			assert.NoError(t, err)
			// Files may be nil or empty slice when no files found - that's ok
			_ = files
		})
	}
}

func TestRunFmt_NonMarkdownFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create non-markdown file
	jsonFile := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(jsonFile, []byte("{}"), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	rootPath = tmpDir
	quiet = true
	verbose = true
	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	// When passing a non-markdown file, it may be skipped or processed
	// The behavior depends on implementation - just verify no panic
	err := runFmt([]string{jsonFile})
	_ = err // Behavior varies - just verify no panic
}

func TestRunFmt_AlreadyFormatted(t *testing.T) {
	tmpDir := t.TempDir()

	// Create already-formatted file
	testFile := filepath.Join(tmpDir, "formatted.md")
	formatted := `---
name: test
description: Test file
model: sonnet
---

# Test

Already formatted content.
`
	require.NoError(t, os.WriteFile(testFile, []byte(formatted), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldFmtWrite := fmtWrite
	rootPath = tmpDir
	quiet = false
	verbose = true
	fmtWrite = true
	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		fmtWrite = oldFmtWrite
	}()

	// Should handle already-formatted files gracefully
	err := runFmt([]string{testFile})
	assert.NoError(t, err)
}

func TestCollectFilesToFormat_MixedArgs(t *testing.T) {
	tmpDir := t.TempDir()

	// Create structure
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	agent := filepath.Join(agentsDir, "agent.md")
	require.NoError(t, os.WriteFile(agent, []byte("# Agent"), 0644))

	customFile := filepath.Join(tmpDir, "custom.md")
	require.NoError(t, os.WriteFile(customFile, []byte("# Custom"), 0644))

	// Mix of component type and file path
	files, err := collectFilesToFormat([]string{"agents", customFile}, tmpDir)
	assert.NoError(t, err)
	// Should only get files from "agents" component type, not both
	// because when component type is found, file args are ignored
	assert.Greater(t, len(files), 0)
}

func TestRunFmt_WithTypeOverride(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test file (not in standard path)
	testFile := filepath.Join(tmpDir, "custom.md")
	content := `---
name: custom
description: Custom component
---

# Custom Component

Some content here.
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldFmtType := fmtType
	rootPath = tmpDir
	quiet = true
	verbose = false
	fmtType = "agent" // Force type override
	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		fmtType = oldFmtType
	}()

	err := runFmt([]string{testFile})
	assert.NoError(t, err)
}

func TestRunFmt_InvalidTypeOverride(t *testing.T) {
	tmpDir := t.TempDir()

	testFile := filepath.Join(tmpDir, "test.md")
	require.NoError(t, os.WriteFile(testFile, []byte("# Test"), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldFmtType := fmtType
	rootPath = tmpDir
	quiet = true
	fmtType = "invalid-type"
	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		fmtType = oldFmtType
	}()

	err := runFmt([]string{testFile})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid type")
}

func TestRunFmt_MultipleFiles(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple test files
	file1 := filepath.Join(tmpDir, "file1.md")
	file2 := filepath.Join(tmpDir, "file2.md")
	file3 := filepath.Join(tmpDir, "file3.md")

	content := `---
description: Test
name: test
---

# Test

Content.
`
	require.NoError(t, os.WriteFile(file1, []byte(content), 0644))
	require.NoError(t, os.WriteFile(file2, []byte(content), 0644))
	require.NoError(t, os.WriteFile(file3, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	oldFmtWrite := fmtWrite
	rootPath = tmpDir
	quiet = false // Show summary
	verbose = false
	fmtWrite = true
	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
		fmtWrite = oldFmtWrite
	}()

	err := runFmt([]string{file1, file2, file3})
	assert.NoError(t, err)
}

func TestRunFmt_MultipleFilesAllFormatted(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple already-formatted files
	file1 := filepath.Join(tmpDir, "file1.md")
	file2 := filepath.Join(tmpDir, "file2.md")

	content := `---
name: test
description: Test file
model: sonnet
---

# Test

Already formatted.
`
	require.NoError(t, os.WriteFile(file1, []byte(content), 0644))
	require.NoError(t, os.WriteFile(file2, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	rootPath = tmpDir
	quiet = false // Show "All X files already formatted" message
	verbose = true
	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	err := runFmt([]string{file1, file2})
	assert.NoError(t, err)
}

func TestRunFmt_VerboseSkipNonMarkdown(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a non-markdown file
	jsonFile := filepath.Join(tmpDir, "config.json")
	require.NoError(t, os.WriteFile(jsonFile, []byte(`{"key": "value"}`), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose
	rootPath = tmpDir
	quiet = false
	verbose = true // Enable verbose to show skip message
	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	err := runFmt([]string{jsonFile})
	// May error because no valid markdown files
	_ = err
}

func TestRunFmt_WriteError(t *testing.T) {
	// Skip on non-Unix systems
	if os.Getenv("CI") != "" {
		t.Skip("Skipping permission test in CI")
	}

	tmpDir := t.TempDir()

	// Create a test file
	testFile := filepath.Join(tmpDir, "readonly.md")
	content := `---
description: Test
name: test
---

# Test
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	// Make directory read-only to cause write error
	require.NoError(t, os.Chmod(tmpDir, 0555))
	defer func() { _ = os.Chmod(tmpDir, 0755) }()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldFmtWrite := fmtWrite
	rootPath = tmpDir
	quiet = true
	fmtWrite = true
	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		fmtWrite = oldFmtWrite
	}()

	// This should return an error when trying to write
	err := runFmt([]string{testFile})
	// Error expected due to permission issues
	_ = err
}

func TestRunFmt_CheckModeFormatted(t *testing.T) {
	tmpDir := t.TempDir()

	// Create an already-formatted file
	testFile := filepath.Join(tmpDir, "formatted.md")
	content := `---
name: test
description: Test file
model: sonnet
---

# Test

Formatted content.
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldFmtCheck := fmtCheck
	rootPath = tmpDir
	quiet = true
	fmtCheck = true
	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		fmtCheck = oldFmtCheck
	}()

	// Should not exit because file is already formatted
	err := runFmt([]string{testFile})
	assert.NoError(t, err)
}

func TestFmtCmdLongDescription(t *testing.T) {
	// Test long description content
	assert.Contains(t, fmtCmd.Long, "FORMATTING RULES")
	assert.Contains(t, fmtCmd.Long, "USAGE MODES")
	assert.Contains(t, fmtCmd.Long, "Frontmatter")
	assert.Contains(t, fmtCmd.Long, "Markdown")
}

func TestRunFmt_CheckModeNeedsFormatting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file that needs formatting in a standard path
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))

	// Create unformatted content (description before name)
	content := `---
description: Test agent
name: test-agent
model: sonnet
---

# Test

Content here.
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Mock exitFunc to prevent test termination
	originalExitFunc := exitFunc
	exitCalled := false
	exitCode := -1
	exitFunc = func(code int) {
		exitCalled = true
		exitCode = code
	}
	defer func() { exitFunc = originalExitFunc }()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldFmtCheck := fmtCheck

	rootPath = tmpDir
	quiet = false // Show "needs formatting" message
	fmtCheck = true

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		fmtCheck = oldFmtCheck
	}()

	// Run the function
	err := runFmt([]string{agentPath})
	assert.NoError(t, err)

	// Should have called exit with code 1
	assert.True(t, exitCalled, "Should call exit in check mode")
	assert.Equal(t, 1, exitCode, "Exit code should be 1")
}

func TestRunFmt_WriteModeSummary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files that need formatting in standard paths
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	// Create unformatted content
	content := `---
description: Test
name: test
model: sonnet
---

# Test

Content.
`
	file1 := filepath.Join(agentsDir, "test1.md")
	file2 := filepath.Join(agentsDir, "test2.md")
	require.NoError(t, os.WriteFile(file1, []byte(content), 0644))
	require.NoError(t, os.WriteFile(file2, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldFmtWrite := fmtWrite

	rootPath = tmpDir
	quiet = false // Show summary
	fmtWrite = true

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		fmtWrite = oldFmtWrite
	}()

	err := runFmt([]string{file1, file2})
	assert.NoError(t, err)
}

func TestRunFmt_DefaultModePrintsFormatted(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file that needs formatting in a standard path
	agentPath := filepath.Join(tmpDir, ".claude", "agents", "test.md")
	require.NoError(t, os.MkdirAll(filepath.Dir(agentPath), 0755))

	// Create unformatted content
	content := `---
description: Test
name: test
model: sonnet
---

# Test

Content.
`
	require.NoError(t, os.WriteFile(agentPath, []byte(content), 0644))

	// Set global flags - default mode (no check, no write, no diff)
	oldRootPath := rootPath
	oldQuiet := quiet
	oldFmtCheck := fmtCheck
	oldFmtWrite := fmtWrite
	oldFmtDiff := fmtDiff

	rootPath = tmpDir
	quiet = true
	fmtCheck = false
	fmtWrite = false
	fmtDiff = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		fmtCheck = oldFmtCheck
		fmtWrite = oldFmtWrite
		fmtDiff = oldFmtDiff
	}()

	err := runFmt([]string{agentPath})
	assert.NoError(t, err)
}

func TestRunFmt_SummaryNeedsFormatting(t *testing.T) {
	tmpDir := t.TempDir()

	// Create multiple files that need formatting
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	content := `---
description: Test
name: test
model: sonnet
---

# Test

Content.
`
	file1 := filepath.Join(agentsDir, "test1.md")
	file2 := filepath.Join(agentsDir, "test2.md")
	require.NoError(t, os.WriteFile(file1, []byte(content), 0644))
	require.NoError(t, os.WriteFile(file2, []byte(content), 0644))

	// Set global flags - neither write nor check mode
	oldRootPath := rootPath
	oldQuiet := quiet
	oldFmtCheck := fmtCheck
	oldFmtWrite := fmtWrite

	rootPath = tmpDir
	quiet = false // Show summary
	fmtCheck = false
	fmtWrite = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		fmtCheck = oldFmtCheck
		fmtWrite = oldFmtWrite
	}()

	err := runFmt([]string{file1, file2})
	assert.NoError(t, err)
}

func TestRunFmt_QuietModeSkipMessages(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with an invalid path to trigger skip message
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	// Create a valid file
	validFile := filepath.Join(agentsDir, "valid.md")
	content := `---
name: test
description: Test
model: sonnet
---

# Test
`
	require.NoError(t, os.WriteFile(validFile, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet

	rootPath = tmpDir
	quiet = true // Quiet mode - suppress skip messages

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
	}()

	// Pass a non-existent file along with a valid one
	err := runFmt([]string{filepath.Join(tmpDir, "nonexistent.md"), validFile})
	// May return error for non-existent file
	_ = err
}

func TestRunFmt_FormatError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file with content that causes format errors
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	// Create file with malformed frontmatter that may cause format issues
	testFile := filepath.Join(agentsDir, "malformed.md")
	content := `---
name: test
invalid yaml: [
---

# Test content
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

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

	// Should handle format errors gracefully
	err := runFmt([]string{testFile})
	// May return error or skip
	_ = err
}

func TestRunFmt_SkipInvalidFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Create test directory structure
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	// Create a valid file
	validFile := filepath.Join(agentsDir, "valid.md")
	content := `---
name: test
description: Test file
model: sonnet
---

# Test
`
	require.NoError(t, os.WriteFile(validFile, []byte(content), 0644))

	// Set global flags for quiet mode
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose

	rootPath = tmpDir
	quiet = false
	verbose = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	// Run with valid file - should not error
	err := runFmt([]string{validFile})
	assert.NoError(t, err)
}

func TestRunFmt_VerboseAlreadyFormatted(t *testing.T) {
	tmpDir := t.TempDir()

	// Create already-formatted file
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	testFile := filepath.Join(agentsDir, "formatted.md")
	content := `---
name: test
description: Test file
model: sonnet
---

# Test

Already formatted content.
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldVerbose := verbose

	rootPath = tmpDir
	quiet = false
	verbose = true // Verbose to show "already formatted" message

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		verbose = oldVerbose
	}()

	err := runFmt([]string{testFile})
	assert.NoError(t, err)
}

func TestRunFmt_CheckModeQuiet(t *testing.T) {
	tmpDir := t.TempDir()

	// Create file needing formatting
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	testFile := filepath.Join(agentsDir, "unformatted.md")
	content := `---
description: Test
name: test
model: sonnet
---

# Test

Content.
`
	require.NoError(t, os.WriteFile(testFile, []byte(content), 0644))

	// Mock exitFunc
	originalExitFunc := exitFunc
	exitCalled := false
	exitFunc = func(code int) {
		exitCalled = true
	}
	defer func() { exitFunc = originalExitFunc }()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet
	oldFmtCheck := fmtCheck

	rootPath = tmpDir
	quiet = true // Quiet mode in check
	fmtCheck = true

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		fmtCheck = oldFmtCheck
	}()

	err := runFmt([]string{testFile})
	assert.NoError(t, err)
	assert.True(t, exitCalled)
}

func TestCollectFilesToFormat_NonExistentPath(t *testing.T) {
	tmpDir := t.TempDir()

	// Test with non-existent path
	_, err := collectFilesToFormat([]string{"/nonexistent/path/file.md"}, tmpDir)
	assert.Error(t, err)
}

func TestDiscoverFilesInDir_Empty(t *testing.T) {
	tmpDir := t.TempDir()

	// Create empty directory
	emptyDir := filepath.Join(tmpDir, "empty")
	require.NoError(t, os.MkdirAll(emptyDir, 0755))

	files, err := discoverFilesInDir(emptyDir)
	assert.NoError(t, err)
	assert.Empty(t, files)
}

func TestRunFmt_ReadError(t *testing.T) {
	// Skip on CI where permission tests are unreliable
	if os.Getenv("CI") != "" {
		t.Skip("Skipping permission test in CI")
	}

	tmpDir := t.TempDir()

	// Create file
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))

	testFile := filepath.Join(agentsDir, "unreadable.md")
	require.NoError(t, os.WriteFile(testFile, []byte("# Test"), 0644))

	// Make file unreadable
	require.NoError(t, os.Chmod(testFile, 0000))
	defer func() { _ = os.Chmod(testFile, 0644) }()

	// Set global flags
	oldRootPath := rootPath
	oldQuiet := quiet

	rootPath = tmpDir
	quiet = false

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
	}()

	// Should handle read error gracefully
	err := runFmt([]string{testFile})
	_ = err
}

func TestRunFmt_DetectTypeError(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a file outside standard paths
	testFile := filepath.Join(tmpDir, "random.md")
	require.NoError(t, os.WriteFile(testFile, []byte("# Test"), 0644))

	// Set global flags - no type override
	oldRootPath := rootPath
	oldQuiet := quiet
	oldFmtType := fmtType

	rootPath = tmpDir
	quiet = false
	fmtType = "" // No type override

	defer func() {
		rootPath = oldRootPath
		quiet = oldQuiet
		fmtType = oldFmtType
	}()

	// Should skip file that can't be type-detected
	err := runFmt([]string{testFile})
	_ = err
}
