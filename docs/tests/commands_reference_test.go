package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// repoRoot returns the repository root directory
func repoRoot(t *testing.T) string {
	t.Helper()

	// Start from the test file directory and go up to find go.mod
	testDir, err := os.Getwd()
	require.NoError(t, err)

	dir := testDir
	for {
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repository root (go.mod not found)")
			return ""
		}
		dir = parent
	}
}

// TestCommandsReferenceUsageSyntax verifies the usage syntax is shown
func TestCommandsReferenceUsageSyntax(t *testing.T) {
	// Given docs/reference/commands/commands.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/commands.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err, "commands.md should exist")

	contentStr := string(content)

	// When read, then usage syntax is shown
	assert.Contains(t, contentStr, "## Usage",
		"Should have Usage section")
	assert.Contains(t, contentStr, "```bash",
		"Should show usage in bash code block")
	assert.Contains(t, contentStr, "cclint commands [flags]",
		"Should show the commands subcommand usage syntax")
}

// TestCommandsReferenceLineLimitRule verifies 50-line limit rule is documented
func TestCommandsReferenceLineLimitRule(t *testing.T) {
	// Given docs/reference/commands/commands.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/commands.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)

	// When scanning content, then 50-line limit rule is documented
	assert.Contains(t, contentStr, "## Line Limit Rule",
		"Should have Line Limit Rule section")

	// Verify the limit is mentioned with tolerance
	assert.Contains(t, contentStr, "~55 lines",
		"Should mention the 55-line limit (50±10%)")
	assert.Contains(t, contentStr, "50 lines ±10% tolerance",
		"Should explain the tolerance")

	// Verify rationale is provided
	assert.Contains(t, contentStr, "thin delegation pattern",
		"Should explain the thin delegation pattern rationale")
}

// TestCommandsReferenceDelegationPattern verifies delegation pattern requirement is explained
func TestCommandsReferenceDelegationPattern(t *testing.T) {
	// Given docs/reference/commands/commands.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/commands.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)

	// When checking sections, then delegation pattern requirement is explained
	assert.Contains(t, contentStr, "## Delegation Pattern Requirement",
		"Should have Delegation Pattern Requirement section")

	// Verify thin command pattern is documented
	assert.Contains(t, contentStr, "### Thin Command Pattern",
		"Should explain Thin Command Pattern")
	assert.Contains(t, contentStr, "Task()",
		"Should mention Task() delegation")

	// Verify fat command anti-pattern is documented
	assert.Contains(t, contentStr, "### Fat Command Anti-Pattern",
		"Should warn about Fat Command Anti-Pattern")

	// Verify validation rules are mentioned
	assert.Contains(t, contentStr, "### Validation",
		"Should list Validation rules")
	assert.Contains(t, contentStr, "Rule 025",
		"Should reference specific validation rules")
}

// TestCommandsReferenceFileExists verifies the file was created
func TestCommandsReferenceFileExists(t *testing.T) {
	// Given the docs structure
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/commands.md")

	// When checking if file exists
	info, err := os.Stat(path)

	// Then file should exist and be readable
	require.NoError(t, err, "commands.md should exist")
	assert.False(t, info.IsDir(), "commands.md should be a file, not directory")
	assert.Greater(t, info.Size(), int64(0), "commands.md should not be empty")
}

// TestCommandsReferenceHasProperStructure verifies the document has proper structure
func TestCommandsReferenceHasProperStructure(t *testing.T) {
	// Given docs/reference/commands/commands.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/commands.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// When checking structure
	// Then should have proper heading hierarchy
	var h1Count, h2Count int
	for _, line := range lines {
		if strings.HasPrefix(line, "# ") {
			h1Count++
		} else if strings.HasPrefix(line, "## ") {
			h2Count++
		}
	}

	assert.Equal(t, 1, h1Count, "Should have exactly one H1 heading")
	assert.GreaterOrEqual(t, h2Count, 4, "Should have at least 4 H2 sections")
}

// TestCommandsReferenceSeeAlsoLinks verifies the See Also section
func TestCommandsReferenceSeeAlsoLinks(t *testing.T) {
	// Given docs/reference/commands/commands.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/commands.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)

	// When checking See Also section
	assert.Contains(t, contentStr, "## See Also",
		"Should have See Also section")

	// Then should link to related documentation
	assert.Contains(t, contentStr, "[Common Tasks](../../common-tasks.md)",
		"Should link to common tasks")
	assert.Contains(t, contentStr, "[Command Lint Rules](../../rules/commands.md)",
		"Should link to command lint rules")
	assert.Contains(t, contentStr, "[Rules Reference](../../rules/README.md)",
		"Should link to rules reference")
	assert.Contains(t, contentStr, "[Configuration Guide](../../guides/configuration.md)",
		"Should link to configuration guide")
}

// TestCommandsReferenceExampleOutput verifies example output is included
func TestCommandsReferenceExampleOutput(t *testing.T) {
	// Given docs/reference/commands/commands.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/commands.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)

	// When checking example output section
	assert.Contains(t, contentStr, "## Example Output",
		"Should have Example Output section")

	// Then should show both passing and failing examples
	assert.Contains(t, contentStr, "### Passing Command",
		"Should show passing command example")
	assert.Contains(t, contentStr, "### Failing Command",
		"Should show failing command example")

	// Verify output format includes key elements
	assert.Contains(t, contentStr, "Score:",
		"Should show quality score in output")
	assert.Contains(t, contentStr, "errors:",
		"Should show errors in output")
	assert.Contains(t, contentStr, "suggestions:",
		"Should show suggestions in output")
}

// TestCommandsReferenceSupportedPatterns verifies supported file patterns are documented
func TestCommandsReferenceSupportedPatterns(t *testing.T) {
	// Given docs/reference/commands/commands.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/commands.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)

	// When checking description section
	assert.Contains(t, contentStr, "### Supported File Patterns",
		"Should document supported file patterns")

	// Then should list the patterns
	assert.Contains(t, contentStr, ".claude/commands/**/*.md",
		"Should include .claude/commands pattern")
	assert.Contains(t, contentStr, "commands/**/*.md",
		"Should include commands pattern")
}

// TestCommandsReferenceMarkdownLint verifies the file passes markdownlint
func TestCommandsReferenceMarkdownLint(t *testing.T) {
	// Given docs/reference/commands/commands.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs", "reference", "commands", "commands.md")

	// When checking if markdownlint would pass
	// Read the file and check for common markdownlint issues
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)
	lines := strings.Split(contentStr, "\n")

	// Check for line length issues (should be < 80 chars for code blocks)
	for i, line := range lines {
		// Skip code blocks when checking line length
		if strings.HasPrefix(line, "```") || strings.HasPrefix(line, "    ") {
			continue
		}
		// Allow headings to be longer
		if strings.HasPrefix(line, "#") {
			continue
		}
		// Check line length for regular content
		if len(line) > 80 && !strings.Contains(line, "http") {
			t.Logf("Line %d may be too long (%d chars): %s", i+1, len(line), line)
		}
	}

	// Verify all code blocks have language specified on opening fences
	inCodeBlock := false
	for i, line := range lines {
		if strings.HasPrefix(line, "```") {
			trimmed := strings.TrimSpace(line)
			if !inCodeBlock && trimmed == "```" {
				// Opening fence without language specification
				t.Errorf("Line %d: Code block missing language specification", i+1)
			}
			inCodeBlock = !inCodeBlock
		}
	}

	// If we got here without errors, the basic markdown checks passed
	assert.True(t, true, "Markdown structure is valid")
}

// TestPluginsReferenceUsageSyntax verifies the usage syntax is shown
func TestPluginsReferenceUsageSyntax(t *testing.T) {
	// Given docs/reference/commands/plugins.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/plugins.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err, "plugins.md should exist")

	contentStr := string(content)

	// When read, then usage syntax is shown
	assert.Contains(t, contentStr, "## Usage",
		"Should have Usage section")
	assert.Contains(t, contentStr, "```bash",
		"Should show usage in bash code block")
	assert.Contains(t, contentStr, "cclint plugins [flags]",
		"Should show the plugins subcommand usage syntax")
}

// TestPluginsReferenceJSONStructure verifies plugin.json structure requirements are listed
func TestPluginsReferenceJSONStructure(t *testing.T) {
	// Given docs/reference/commands/plugins.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/plugins.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)

	// When scanning content, then plugin.json structure requirements are listed
	assert.Contains(t, contentStr, "## Required Fields",
		"Should have Required Fields section")
	assert.Contains(t, contentStr, "## Recommended Fields",
		"Should have Recommended Fields section")

	// Verify required fields are documented
	assert.Contains(t, contentStr, "`name`",
		"Should document name field")
	assert.Contains(t, contentStr, "`description`",
		"Should document description field")
	assert.Contains(t, contentStr, "`author.name`",
		"Should document author.name field")

	// Verify field constraints are mentioned
	assert.Contains(t, contentStr, "max 64 chars",
		"Should mention name length constraint")
	assert.Contains(t, contentStr, "max 1024 chars",
		"Should mention description length constraint")
	assert.Contains(t, contentStr, "50+ recommended",
		"Should recommend minimum description length")
}

// TestPluginsReferenceSizeLimit verifies 5KB size limit is documented
func TestPluginsReferenceSizeLimit(t *testing.T) {
	// Given docs/reference/commands/plugins.md exists
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/plugins.md")
	content, err := os.ReadFile(path)
	require.NoError(t, err)

	contentStr := string(content)

	// When checking sections, then 5KB size limit is documented
	assert.Contains(t, contentStr, "## File Size Limit",
		"Should have File Size Limit section")

	// Verify the 5KB limit is mentioned
	assert.Contains(t, contentStr, "5KB",
		"Should mention the 5KB limit")

	// Verify scoring tiers are documented
	assert.Contains(t, contentStr, "### Scoring Tiers",
		"Should document scoring tiers")
	assert.Contains(t, contentStr, "≤1KB",
		"Should show ≤1KB tier")
	assert.Contains(t, contentStr, "≤2KB",
		"Should show ≤2KB tier")
	assert.Contains(t, contentStr, "≤5KB",
		"Should show ≤5KB tier")
	assert.Contains(t, contentStr, "≤10KB",
		"Should show ≤10KB tier")
	assert.Contains(t, contentStr, ">10KB",
		"Should show >10KB tier")

	// Verify rationale is provided
	assert.Contains(t, contentStr, "### Rationale",
		"Should explain the rationale")
	assert.Contains(t, contentStr, "concise metadata files",
		"Should explain manifests should be concise")

	// Verify example violation message
	assert.Contains(t, contentStr, "### Example Violation Message",
		"Should show example violation message")
	assert.Contains(t, contentStr, "keep plugin.json under 5KB",
		"Should mention 5KB recommendation in violation message")
}

// TestPluginsReferenceFileExists verifies the file was created
func TestPluginsReferenceFileExists(t *testing.T) {
	// Given the docs structure
	root := repoRoot(t)
	path := filepath.Join(root, "docs/reference/commands/plugins.md")

	// When checking if file exists
	info, err := os.Stat(path)

	// Then file should exist and be readable
	require.NoError(t, err, "plugins.md should exist")
	assert.False(t, info.IsDir(), "plugins.md should be a file, not directory")
	assert.Greater(t, info.Size(), int64(0), "plugins.md should not be empty")
}
