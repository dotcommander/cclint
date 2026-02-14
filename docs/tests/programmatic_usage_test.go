package docs_test

import (
	"os"
	"strings"
	"testing"
)

// TestProgrammaticUsage_InternalScopeDocumented verifies internal API scope is documented
func TestProgrammaticUsage_InternalScopeDocumented(t *testing.T) {
	content, err := os.ReadFile("../../docs/guides/programmatic-usage.md")
	if err != nil {
		t.Fatalf("Failed to read programmatic-usage.md: %v", err)
	}

	strContent := string(content)

	// Check for explicit internal-only scope statement
	if !strings.Contains(strContent, "internal/...") || !strings.Contains(strContent, "not a supported external API surface") {
		t.Error("Missing internal API scope statement")
	}

	// Verify it's a code block (wrapped in backticks)
	if !strings.Contains(strContent, "```go") {
		t.Error("Import path not in Go code block")
	}
}

// TestProgrammaticUsage_KeyInternalAPIsListed verifies key internal APIs are documented
func TestProgrammaticUsage_KeyInternalAPIsListed(t *testing.T) {
	content, err := os.ReadFile("../../docs/guides/programmatic-usage.md")
	if err != nil {
		t.Fatalf("Failed to read programmatic-usage.md: %v", err)
	}

	strContent := string(content)

	// Check for Core Internal APIs section
	if !strings.Contains(strContent, "## Core Internal APIs") {
		t.Error("Missing '## Core Internal APIs' section")
	}

	// Verify key API packages are documented
	requiredAPIs := []string{
		"### Discovery (`internal/discovery`)",
		"### Validation (`internal/cue`)",
		"### Lint Context (`internal/cli`)",
		"### Scoring (`internal/scoring`)",
	}

	for _, api := range requiredAPIs {
		if !strings.Contains(strContent, api) {
			t.Errorf("Missing API section: %s", api)
		}
	}

	// Check for type documentation
	requiredTypes := []string{
		"`File`:",
		"`FileType`:",
		"`ValidationError`:",
		"`Frontmatter`:",
		"`LinterContext`:",
		"`QualityScore`:",
	}

	for _, typ := range requiredTypes {
		if !strings.Contains(strContent, typ) {
			t.Errorf("Missing type documentation: %s", typ)
		}
	}
}

// TestProgrammaticUsage_ValidationCallExample verifies validation code example is included
func TestProgrammaticUsage_ValidationCallExample(t *testing.T) {
	content, err := os.ReadFile("../../docs/guides/programmatic-usage.md")
	if err != nil {
		t.Fatalf("Failed to read programmatic-usage.md: %v", err)
	}

	strContent := string(content)

	// Check for example section
	if !strings.Contains(strContent, "## Example: Validate Single File") {
		t.Error("Missing '## Example: Validate Single File' section")
	}

	// Verify key validation patterns are shown
	requiredPatterns := []string{
		"cue.NewValidator()",
		"validator.LoadSchemas",
		"validator.ValidateFile(",
	}

	for _, pattern := range requiredPatterns {
		if !strings.Contains(strContent, pattern) {
			t.Errorf("Missing validation pattern: %s", pattern)
		}
	}

	// Check for proper Go example structure
	if !strings.Contains(strContent, "```go\npackage main") {
		t.Error("Validation example not in proper Go code block format")
	}

	// Verify the example shows error handling
	if !strings.Contains(strContent, "if err != nil") {
		t.Error("Example missing error handling")
	}
}
