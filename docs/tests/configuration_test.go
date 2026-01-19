package docs_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigurationGuide_ConfigFileFormatsListed(t *testing.T) {
	// Given docs/guides/configuration.md exists
	path := filepath.Join("..", "guides", "configuration.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read configuration.md: %v", err)
	}

	contentStr := string(content)

	// When read, then supported config file formats are listed
	if !strings.Contains(contentStr, ".cclintrc.json") {
		t.Error("Missing .cclintrc.json format in documentation")
	}

	if !strings.Contains(contentStr, ".cclintrc.yaml") {
		t.Error("Missing .cclintrc.yaml format in documentation")
	}

	if !strings.Contains(contentStr, ".cclintrc.yml") {
		t.Error("Missing .cclintrc.yml format in documentation")
	}

	// Check that format order is mentioned
	if !strings.Contains(contentStr, "searched in order") && !strings.Contains(contentStr, "precedence") {
		t.Error("Missing information about config file search order/precedence")
	}
}

func TestConfigurationGuide_YAMLExampleShown(t *testing.T) {
	// Given configuration.md
	path := filepath.Join("..", "guides", "configuration.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read configuration.md: %v", err)
	}

	contentStr := string(content)

	// When scanning content, then example .cclintrc.yaml is shown
	if !strings.Contains(contentStr, "## Example Configuration") {
		t.Error("Missing Example Configuration section")
	}

	if !strings.Contains(contentStr, "### YAML Format") {
		t.Error("Missing YAML Format section heading")
	}

	// Check for key YAML configuration examples
	expectedKeys := []string{
		"root:",
		"exclude:",
		"followSymlinks:",
		"format:",
		"failOn:",
		"quiet:",
		"verbose:",
		"showScores:",
		"showImprovements:",
		"concurrency:",
		"rules:",
		"schemas:",
	}

	for _, key := range expectedKeys {
		if !strings.Contains(contentStr, key) {
			t.Errorf("Missing configuration key in YAML example: %s", key)
		}
	}

	// Check for recommended format indicator
	if !strings.Contains(contentStr, "recommended") {
		t.Error("Missing indication that YAML format is recommended")
	}
}

func TestConfigurationGuide_EnvironmentVariablesDocumented(t *testing.T) {
	// Given configuration.md
	path := filepath.Join("..", "guides", "configuration.md")
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read configuration.md: %v", err)
	}

	contentStr := string(content)

	// When checking sections, then environment variable overrides are documented
	if !strings.Contains(contentStr, "## Environment Variables") {
		t.Error("Missing Environment Variables section")
	}

	// Check for CCLINT_ prefix documentation
	if !strings.Contains(contentStr, "CCLINT_") {
		t.Error("Missing CCLINT_ prefix documentation")
	}

	// Check for specific environment variable examples
	expectedEnvVars := []string{
		"CCLINT_ROOT",
		"CCLINT_EXCLUDE",
		"CCLINT_FORMAT",
		"CCLINT_OUTPUT",
		"CCLINT_FAILON",
		"CCLINT_QUIET",
		"CCLINT_VERBOSE",
		"CCLINT_SHOWSCORES",
		"CCLINT_CONCURRENCY",
	}

	for _, envVar := range expectedEnvVars {
		if !strings.Contains(contentStr, envVar) {
			t.Errorf("Missing environment variable documentation: %s", envVar)
		}
	}

	// Check for priority order information
	if !strings.Contains(contentStr, "Priority Order") {
		t.Error("Missing Priority Order section")
	}

	expectedOrderSources := []string{
		"Default values",
		"Configuration file",
		"Environment variables",
		"Command-line flags",
	}

	for _, source := range expectedOrderSources {
		if !strings.Contains(contentStr, source) {
			t.Errorf("Missing priority order source: %s", source)
		}
	}

	// Check for export examples
	if !strings.Contains(contentStr, "export CCLINT_") {
		t.Error("Missing export command examples for environment variables")
	}
}
