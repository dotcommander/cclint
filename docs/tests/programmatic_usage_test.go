package docs_test

import (
	"os"
	"strings"
	"testing"
)

// TestProgrammaticUsage_GoImportDocumented verifies the Go import path is documented
func TestProgrammaticUsage_GoImportDocumented(t *testing.T) {
	content, err := os.ReadFile("../../docs/guides/programmatic-usage.md")
	if err != nil {
		t.Fatalf("Failed to read programmatic-usage.md: %v", err)
	}

	strContent := string(content)

	// Check for import path section header
	if !strings.Contains(strContent, "## Import Path") {
		t.Error("Missing '## Import Path' section")
	}

	// Check for Go import statement with correct module path
	if !strings.Contains(strContent, `import "github.com/dotcommander/cclint/internal/...`) {
		t.Error("Missing or incorrect Go import path documentation")
	}

	// Verify it's a code block (wrapped in backticks)
	if !strings.Contains(strContent, "```go") {
		t.Error("Import path not in Go code block")
	}
}

// TestProgrammaticUsage_KeyPublicAPIsListed verifies key public APIs are documented
func TestProgrammaticUsage_KeyPublicAPIsListed(t *testing.T) {
	content, err := os.ReadFile("../../docs/guides/programmatic-usage.md")
	if err != nil {
		t.Fatalf("Failed to read programmatic-usage.md: %v", err)
	}

	strContent := string(content)

	// Check for Core Public APIs section
	if !strings.Contains(strContent, "## Core Public APIs") {
		t.Error("Missing '## Core Public APIs' section")
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
