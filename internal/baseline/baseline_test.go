package baseline

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/cue"
)

func TestCreateBaseline(t *testing.T) {
	issues := []cue.ValidationError{
		{
			File:     "agents/test-agent.md",
			Message:  "Name 'test-agent' doesn't match filename 'other-name'",
			Severity: "error",
			Source:   "cclint-observation",
		},
		{
			File:     "commands/test-command.md",
			Message:  "Missing required field 'name'",
			Severity: "error",
			Source:   "anthropic-docs",
		},
		// Duplicate issue - should be deduplicated
		{
			File:     "agents/test-agent.md",
			Message:  "Name 'test-agent' doesn't match filename 'other-name'",
			Severity: "error",
			Source:   "cclint-observation",
		},
	}

	baseline := CreateBaseline(issues)

	if baseline.Version != "1.0" {
		t.Errorf("Expected version 1.0, got %s", baseline.Version)
	}

	if len(baseline.Fingerprints) != 2 {
		t.Errorf("Expected 2 unique fingerprints, got %d", len(baseline.Fingerprints))
	}

	// Check index is built
	if len(baseline.index) != 2 {
		t.Errorf("Expected index with 2 entries, got %d", len(baseline.index))
	}
}

func TestIsKnown(t *testing.T) {
	issue1 := cue.ValidationError{
		File:     "agents/test.md",
		Message:  "Name 'test' doesn't match filename 'other'",
		Severity: "error",
		Source:   "cclint-observation",
	}

	issue2 := cue.ValidationError{
		File:     "commands/cmd.md",
		Message:  "Missing required field 'name'",
		Severity: "error",
		Source:   "anthropic-docs",
	}

	baseline := CreateBaseline([]cue.ValidationError{issue1})

	if !baseline.IsKnown(issue1) {
		t.Error("Expected issue1 to be known in baseline")
	}

	if baseline.IsKnown(issue2) {
		t.Error("Expected issue2 to not be known in baseline")
	}
}

func TestSaveAndLoadBaseline(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, ".cclintbaseline.json")

	issues := []cue.ValidationError{
		{
			File:     "agents/test.md",
			Message:  "Some error message",
			Severity: "error",
			Source:   "cclint-observation",
		},
	}

	// Create and save baseline
	original := CreateBaseline(issues)
	original.CreatedAt = time.Now().UTC().Format(time.RFC3339)

	if err := original.SaveBaseline(baselinePath); err != nil {
		t.Fatalf("Failed to save baseline: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(baselinePath); err != nil {
		t.Fatalf("Baseline file not created: %v", err)
	}

	// Load baseline
	loaded, err := LoadBaseline(baselinePath)
	if err != nil {
		t.Fatalf("Failed to load baseline: %v", err)
	}

	// Verify contents
	if loaded.Version != original.Version {
		t.Errorf("Version mismatch: expected %s, got %s", original.Version, loaded.Version)
	}

	if len(loaded.Fingerprints) != len(original.Fingerprints) {
		t.Errorf("Fingerprint count mismatch: expected %d, got %d",
			len(original.Fingerprints), len(loaded.Fingerprints))
	}

	// Verify index is rebuilt
	if len(loaded.index) != len(original.Fingerprints) {
		t.Errorf("Index not rebuilt: expected %d entries, got %d",
			len(original.Fingerprints), len(loaded.index))
	}

	// Verify we can check issues
	if !loaded.IsKnown(issues[0]) {
		t.Error("Expected loaded baseline to recognize original issue")
	}
}

func TestNormalizeMessage(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{
			input:    "Name 'test-agent' doesn't match filename 'other-name'",
			expected: "Name '*' doesn't match filename '*'",
		},
		{
			input:    "Agent is 250 lines. Best practice: keep under 220 lines",
			expected: "Agent is N lines. Best practice: keep under N lines",
		},
		{
			input:    `Missing required field "name"`,
			expected: `Missing required field "*"`,
		},
		{
			input:    "Extra   whitespace   here",
			expected: "Extra whitespace here",
		},
	}

	for _, tt := range tests {
		result := normalizeMessage(tt.input)
		if result != tt.expected {
			t.Errorf("normalizeMessage(%q)\nExpected: %q\nGot:      %q",
				tt.input, tt.expected, result)
		}
	}
}

func TestFingerprintStability(t *testing.T) {
	// Same issue should produce same fingerprint
	issue := cue.ValidationError{
		File:     "agents/test.md",
		Message:  "Name 'test' doesn't match filename 'other'",
		Severity: "error",
		Source:   "cclint-observation",
		Line:     10, // Line number shouldn't affect fingerprint
	}

	fp1 := fingerprint(issue)

	// Change line number
	issue.Line = 20
	fp2 := fingerprint(issue)

	if fp1 != fp2 {
		t.Error("Fingerprint changed when only line number changed")
	}

	// Change message content
	issue.Message = "Name 'different' doesn't match filename 'another'"
	fp3 := fingerprint(issue)

	if fp1 != fp3 {
		t.Error("Fingerprint changed when only specific values in message changed (should normalize)")
	}

	// Change message pattern
	issue.Message = "Completely different error"
	fp4 := fingerprint(issue)

	if fp1 == fp4 {
		t.Error("Fingerprint didn't change when message pattern changed")
	}
}

func TestLoadNonexistentBaseline(t *testing.T) {
	_, err := LoadBaseline("/nonexistent/path/.cclintbaseline.json")
	if err == nil {
		t.Error("Expected error when loading nonexistent baseline")
	}
}

func TestLoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	baselinePath := filepath.Join(tmpDir, ".cclintbaseline.json")

	// Write invalid JSON
	if err := os.WriteFile(baselinePath, []byte("invalid json"), 0644); err != nil {
		t.Fatalf("Failed to write test file: %v", err)
	}

	_, err := LoadBaseline(baselinePath)
	if err == nil {
		t.Error("Expected error when loading invalid JSON")
	}
}
