package lint

import (
	"testing"
)

func TestLintContext(t *testing.T) {
	// Test with empty directory
	summary, err := LintContext("testdata/empty", false, false, true)
	if err != nil {
		t.Fatalf("LintContext() error = %v", err)
	}
	if summary == nil {
		t.Fatal("LintContext() returned nil summary")
	}
}

func TestParseMarkdownSections(t *testing.T) {
	// parseMarkdownSections is currently a stub that returns nil
	// This test verifies the stub behavior
	sections := parseMarkdownSections("")
	if sections != nil {
		t.Errorf("parseMarkdownSections() = %v, want nil", sections)
	}
}

