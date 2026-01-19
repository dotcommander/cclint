package reference

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSchemasDocumentationExists verifies docs/reference/schemas.md exists
func TestSchemasDocumentationExists(t *testing.T) {
	// Use absolute path to ensure the test works from any working directory
	path := filepath.Join("..", "..", "docs", "reference", "schemas.md")
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Try relative path from project root
		path = filepath.Join("docs", "reference", "schemas.md")
		if _, err := os.Stat(path); os.IsNotExist(err) {
			t.Fatalf("docs/reference/schemas.md does not exist")
		}
	}
}

// TestSchemasPurposeExplained verifies purpose of CUE validation is explained
func TestSchemasPurposeExplained(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "reference", "schemas.md")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Try relative path from project root
		path = filepath.Join("docs", "reference", "schemas.md")
		content, err = os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read schemas.md: %v", err)
		}
	}

	s := string(content)

	// Check for purpose section header
	if !strings.Contains(s, "## Purpose of CUE Validation") {
		t.Error("Missing '## Purpose of CUE Validation' section")
	}

	// Check for key purpose concepts
	purposeKeywords := []string{
		"CUE",
		"validation",
		"type safety",
		"constraint",
		"structural",
	}

	for _, kw := range purposeKeywords {
		if !strings.Contains(strings.ToLower(s), strings.ToLower(kw)) {
			t.Errorf("Purpose section does not mention key concept: %s", kw)
		}
	}

	// Check for validation pipeline explanation
	if !strings.Contains(s, "validation pipeline") {
		t.Error("Missing validation pipeline explanation")
	}
}

// TestSchemasDocumented verifies agent, command, and settings schemas are documented
func TestSchemasDocumented(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "reference", "schemas.md")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Try relative path from project root
		path = filepath.Join("docs", "reference", "schemas.md")
		content, err = os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read schemas.md: %v", err)
		}
	}

	s := string(content)

	// Check for each schema section
	requiredSchemas := []struct {
		name   string
		header string
		file   string
	}{
		{"agent", "## Agent Schema", "agent.cue"},
		{"command", "## Command Schema", "command.cue"},
		{"settings", "## Settings Schema", "settings.cue"},
		{"skill", "## Skill Schema", "skill.cue"},
	}

	for _, schema := range requiredSchemas {
		// Check section header exists
		if !strings.Contains(s, schema.header) {
			t.Errorf("Missing section: %s", schema.header)
		}

		// Check file reference
		if !strings.Contains(s, schema.file) {
			t.Errorf("Schema %s does not reference file: %s", schema.name, schema.file)
		}
	}
}

// TestFrontmatterRequirementsListed verifies frontmatter requirements for each type are listed
func TestFrontmatterRequirementsListed(t *testing.T) {
	path := filepath.Join("..", "..", "docs", "reference", "schemas.md")
	content, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		// Try relative path from project root
		path = filepath.Join("docs", "reference", "schemas.md")
		content, err = os.ReadFile(path)
		if err != nil {
			t.Fatalf("Failed to read schemas.md: %v", err)
		}
	}

	s := string(content)

	// Check Agent frontmatter requirements
	agentRequiredFields := []string{"name", "description"}
	for _, field := range agentRequiredFields {
		if !strings.Contains(s, "`"+field+"`") {
			t.Errorf("Agent schema does not document required field: %s", field)
		}
	}
	if !strings.Contains(s, "### Required Fields") {
		t.Error("Missing '### Required Fields' subsection")
	}

	// Check Command frontmatter (all optional)
	if !strings.Contains(s, "**None**") || !strings.Contains(s, "All fields are optional") {
		t.Error("Command schema should indicate all fields are optional")
	}

	// Check Skill frontmatter requirements
	skillRequiredFields := []string{"name", "description"}
	for _, field := range skillRequiredFields {
		if !strings.Contains(s, "`"+field+"`") && !strings.Contains(s, field) {
			t.Errorf("Skill schema does not document required field: %s", field)
		}
	}

	// Check for field type documentation
	if !strings.Contains(s, "Type") && !strings.Contains(s, "Constraints") {
		t.Error("Frontmatter requirements should include types and constraints")
	}
}