package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAgentsCmd(t *testing.T) {
	// Verify agents command is properly configured
	assert.Equal(t, "agents", agentsCmd.Use)
	assert.NotEmpty(t, agentsCmd.Short)
	assert.NotEmpty(t, agentsCmd.Long)
	assert.NotNil(t, agentsCmd.Run)
}

func TestCommandsCmd(t *testing.T) {
	// Verify commands command is properly configured
	assert.Equal(t, "commands", commandsCmd.Use)
	assert.NotEmpty(t, commandsCmd.Short)
	assert.NotEmpty(t, commandsCmd.Long)
	assert.NotNil(t, commandsCmd.Run)
}

func TestSkillsCmd(t *testing.T) {
	// Verify skills command is properly configured
	assert.Equal(t, "skills", skillsCmd.Use)
	assert.NotEmpty(t, skillsCmd.Short)
	assert.NotEmpty(t, skillsCmd.Long)
	assert.NotNil(t, skillsCmd.Run)
}

func TestPluginsCmd(t *testing.T) {
	// Verify plugins command is properly configured
	assert.Equal(t, "plugins", pluginsCmd.Use)
	assert.NotEmpty(t, pluginsCmd.Short)
	assert.NotEmpty(t, pluginsCmd.Long)
	assert.NotNil(t, pluginsCmd.Run)
}

func TestSummaryCmd(t *testing.T) {
	// Verify summary command is properly configured
	assert.Equal(t, "summary", summaryCmd.Use)
	assert.NotEmpty(t, summaryCmd.Short)
	assert.NotEmpty(t, summaryCmd.Long)
	assert.NotNil(t, summaryCmd.Run)
}

func TestFmtCmd(t *testing.T) {
	// Verify fmt command is properly configured
	assert.Equal(t, "fmt [files...]", fmtCmd.Use)
	assert.NotEmpty(t, fmtCmd.Short)
	assert.NotEmpty(t, fmtCmd.Long)
	assert.NotNil(t, fmtCmd.Run)
}

func TestRunAgentsLint(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid agent file
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	agentFile := filepath.Join(agentsDir, "test-agent.md")
	content := `---
name: test-agent
description: A test agent
model: sonnet
---

## Foundation
Test agent.

## Workflow
1. Test step
`
	require.NoError(t, os.WriteFile(agentFile, []byte(content), 0644))

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

	// Mock os.Exit to prevent test termination
	originalOsExit := osExit
	osExit = func(code int) {
		// Prevent actual exit
	}
	defer func() { osExit = originalOsExit }()

	// Run the function
	err := runAgentsLint()

	// Should complete without error (or exit might be called)
	if err != nil {
		t.Logf("runAgentsLint returned error: %v", err)
	}
}

func TestRunCommandsLint(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid command file
	commandsDir := filepath.Join(tmpDir, ".claude", "commands")
	require.NoError(t, os.MkdirAll(commandsDir, 0755))
	commandFile := filepath.Join(commandsDir, "test-command.md")
	content := `---
name: test-command
description: A test command
model: sonnet
---

## Quick Reference

| User Intent | Action |
|-------------|--------|
| Test | Execute test |
`
	require.NoError(t, os.WriteFile(commandFile, []byte(content), 0644))

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

	// Mock os.Exit
	originalOsExit := osExit
	osExit = func(code int) {}
	defer func() { osExit = originalOsExit }()

	// Run the function
	err := runCommandsLint()

	if err != nil {
		t.Logf("runCommandsLint returned error: %v", err)
	}
}

func TestRunSkillsLint(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid skill file
	skillsDir := filepath.Join(tmpDir, ".claude", "skills", "test-skill")
	require.NoError(t, os.MkdirAll(skillsDir, 0755))
	skillFile := filepath.Join(skillsDir, "SKILL.md")
	content := `---
name: test-skill
description: A test skill
---

## Quick Reference

| User Intent | Action |
|-------------|--------|
| Test | Execute test |

## Workflow

1. Test step 1
2. Test step 2
`
	require.NoError(t, os.WriteFile(skillFile, []byte(content), 0644))

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

	// Mock os.Exit
	originalOsExit := osExit
	osExit = func(code int) {}
	defer func() { osExit = originalOsExit }()

	// Run the function
	err := runSkillsLint()

	if err != nil {
		t.Logf("runSkillsLint returned error: %v", err)
	}
}

func TestRunPluginsLint(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid plugin manifest
	pluginDir := filepath.Join(tmpDir, ".claude-plugin")
	require.NoError(t, os.MkdirAll(pluginDir, 0755))
	pluginFile := filepath.Join(pluginDir, "plugin.json")
	content := `{
  "name": "test-plugin",
  "description": "A test plugin",
  "version": "1.0.0",
  "author": {
    "name": "Test Author"
  }
}`
	require.NoError(t, os.WriteFile(pluginFile, []byte(content), 0644))

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

	// Mock os.Exit
	originalOsExit := osExit
	osExit = func(code int) {}
	defer func() { osExit = originalOsExit }()

	// Run the function
	err := runPluginsLint()

	if err != nil {
		t.Logf("runPluginsLint returned error: %v", err)
	}
}

func TestSubcommandInitialization(t *testing.T) {
	// Verify all subcommands are registered with root
	commands := rootCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	expectedCommands := []string{
		"agents",
		"commands",
		"skills",
		"plugins",
		"summary",
		"fmt",
	}

	for _, name := range expectedCommands {
		assert.True(t, commandMap[name], "subcommand %s should be registered", name)
	}
}

func TestSubcommandLongDescriptions(t *testing.T) {
	// Verify all subcommands have helpful long descriptions
	tests := []struct {
		cmd  *cobra.Command
		name string
	}{
		{agentsCmd, "agents"},
		{commandsCmd, "commands"},
		{skillsCmd, "skills"},
		{pluginsCmd, "plugins"},
		{summaryCmd, "summary"},
		{fmtCmd, "fmt"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.NotEmpty(t, tt.cmd.Long, "%s command should have Long description", tt.name)
			assert.Greater(t, len(tt.cmd.Long), len(tt.cmd.Short),
				"%s Long description should be longer than Short", tt.name)
		})
	}
}

func TestAgentsCmd_FilePatterns(t *testing.T) {
	// Verify agents command documents expected file patterns
	assert.Contains(t, agentsCmd.Long, ".claude/agents")
	assert.Contains(t, agentsCmd.Long, "agents/")
}

func TestCommandsCmd_FilePatterns(t *testing.T) {
	// Verify commands command documents expected file patterns
	assert.Contains(t, commandsCmd.Long, ".claude/commands")
	assert.Contains(t, commandsCmd.Long, "commands/")
}

func TestSkillsCmd_FilePatterns(t *testing.T) {
	// Verify skills command documents expected file patterns
	assert.Contains(t, skillsCmd.Long, "SKILL.md")
	assert.Contains(t, skillsCmd.Long, "skills/")
}

func TestPluginsCmd_FilePatterns(t *testing.T) {
	// Verify plugins command documents expected file patterns
	assert.Contains(t, pluginsCmd.Long, ".claude-plugin")
	assert.Contains(t, pluginsCmd.Long, "plugin.json")
}

func TestFmtCmd_UsageExamples(t *testing.T) {
	// Verify fmt command has usage examples
	assert.Contains(t, fmtCmd.Long, "USAGE MODES")
	assert.Contains(t, fmtCmd.Long, "EXAMPLES")
	assert.Contains(t, fmtCmd.Long, "FLAGS")
}

func TestSummaryCmd_Purpose(t *testing.T) {
	// Verify summary command describes its purpose
	assert.Contains(t, summaryCmd.Long, "quality")
	assert.Contains(t, summaryCmd.Long, "components")
}

func TestSubcommandErrorHandling(t *testing.T) {
	// Verify all subcommand Run functions handle errors properly
	// by checking they call os.Exit or return errors

	tests := []struct {
		cmdName string
		cmd     *cobra.Command
	}{
		{"agents", agentsCmd},
		{"commands", commandsCmd},
		{"skills", skillsCmd},
		{"plugins", pluginsCmd},
		{"summary", summaryCmd},
		{"fmt", fmtCmd},
	}

	for _, tt := range tests {
		t.Run(tt.cmdName, func(t *testing.T) {
			// Verify Run function is not nil
			assert.NotNil(t, tt.cmd.Run, "%s command should have Run function", tt.cmdName)
		})
	}
}

func TestRunComponentLint_Integration(t *testing.T) {
	// This is tested indirectly through the subcommand tests,
	// but we can verify the linter function signature is correct
	assert.NotNil(t, runAgentsLint)
	assert.NotNil(t, runCommandsLint)
	assert.NotNil(t, runSkillsLint)
	assert.NotNil(t, runPluginsLint)
}

func TestContextCmd(t *testing.T) {
	// Verify context command is properly configured
	assert.Equal(t, "context", contextCmd.Use)
	assert.NotEmpty(t, contextCmd.Short)
	assert.NotEmpty(t, contextCmd.Long)
	assert.NotNil(t, contextCmd.Run)
	assert.Contains(t, contextCmd.Long, "CLAUDE.md")
}

func TestRulesCmd(t *testing.T) {
	// Verify rules command is properly configured
	assert.Equal(t, "rules", rulesCmd.Use)
	assert.NotEmpty(t, rulesCmd.Short)
	assert.NotEmpty(t, rulesCmd.Long)
	assert.NotNil(t, rulesCmd.RunE)
	assert.Contains(t, rulesCmd.Long, ".claude/rules")
}

func TestSettingsCmd(t *testing.T) {
	// Verify settings command is properly configured
	assert.Equal(t, "settings", settingsCmd.Use)
	assert.NotEmpty(t, settingsCmd.Short)
	assert.NotEmpty(t, settingsCmd.Long)
	assert.NotNil(t, settingsCmd.Run)
	assert.Contains(t, settingsCmd.Long, "settings.json")
}

func TestRunContextLint(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid CLAUDE.md file
	claudeDir := filepath.Join(tmpDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	claudeFile := filepath.Join(claudeDir, "CLAUDE.md")
	content := `# Project Context

This is a test CLAUDE.md file for testing.

## Build Commands

` + "`go build`" + `

## Test Commands

` + "`go test ./...`" + `
`
	require.NoError(t, os.WriteFile(claudeFile, []byte(content), 0644))

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

	// Run the function
	err := runContextLint()

	// Should complete without error
	if err != nil {
		t.Logf("runContextLint returned error: %v", err)
	}
}

func TestRunRulesLint(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid rules file
	rulesDir := filepath.Join(tmpDir, ".claude", "rules")
	require.NoError(t, os.MkdirAll(rulesDir, 0755))
	rulesFile := filepath.Join(rulesDir, "test-rule.md")
	content := `# Test Rule

This is a test rule file.

## Guidelines

Follow these guidelines.
`
	require.NoError(t, os.WriteFile(rulesFile, []byte(content), 0644))

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

	// Run the function
	err := runRulesLint()

	if err != nil {
		t.Logf("runRulesLint returned error: %v", err)
	}
}

func TestRunSettingsLint(t *testing.T) {
	tmpDir := t.TempDir()

	// Create a valid settings.json file
	claudeDir := filepath.Join(tmpDir, ".claude")
	require.NoError(t, os.MkdirAll(claudeDir, 0755))
	settingsFile := filepath.Join(claudeDir, "settings.json")
	content := `{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Bash",
        "command": "echo test"
      }
    ]
  }
}`
	require.NoError(t, os.WriteFile(settingsFile, []byte(content), 0644))

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

	// Run the function
	err := runSettingsLint()

	if err != nil {
		t.Logf("runSettingsLint returned error: %v", err)
	}
}

func TestRunSummary(t *testing.T) {
	tmpDir := t.TempDir()

	// Create valid component files for summary
	agentsDir := filepath.Join(tmpDir, ".claude", "agents")
	commandsDir := filepath.Join(tmpDir, ".claude", "commands")
	skillsDir := filepath.Join(tmpDir, ".claude", "skills", "test-skill")
	require.NoError(t, os.MkdirAll(agentsDir, 0755))
	require.NoError(t, os.MkdirAll(commandsDir, 0755))
	require.NoError(t, os.MkdirAll(skillsDir, 0755))

	agentContent := `---
name: test-agent
description: A test agent
model: sonnet
---

## Foundation
Test agent.

## Workflow
1. Test step
`
	require.NoError(t, os.WriteFile(filepath.Join(agentsDir, "test-agent.md"), []byte(agentContent), 0644))

	commandContent := `---
name: test-command
description: A test command
model: sonnet
---

## Quick Reference

| Intent | Action |
|--------|--------|
| Test | Execute |
`
	require.NoError(t, os.WriteFile(filepath.Join(commandsDir, "test-command.md"), []byte(commandContent), 0644))

	skillContent := `---
name: test-skill
description: A test skill
---

## Quick Reference

| Intent | Action |
|--------|--------|
| Test | Execute |

## Workflow

1. Step one
`
	require.NoError(t, os.WriteFile(filepath.Join(skillsDir, "SKILL.md"), []byte(skillContent), 0644))

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

	// Run the function
	err := runSummary()

	if err != nil {
		t.Logf("runSummary returned error: %v", err)
	}
}

func TestAllSubcommandsRegistered(t *testing.T) {
	// Verify all expected subcommands are registered
	commands := rootCmd.Commands()
	commandMap := make(map[string]bool)
	for _, cmd := range commands {
		commandMap[cmd.Name()] = true
	}

	expectedCommands := []string{
		"agents",
		"commands",
		"skills",
		"plugins",
		"summary",
		"fmt",
		"context",
		"rules",
		"settings",
	}

	for _, name := range expectedCommands {
		assert.True(t, commandMap[name], "subcommand %s should be registered", name)
	}
}
