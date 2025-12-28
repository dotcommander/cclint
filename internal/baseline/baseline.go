package baseline

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"sort"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// Baseline represents a snapshot of known issues that should be ignored
type Baseline struct {
	Version      string   `json:"version"`
	CreatedAt    string   `json:"created_at"`
	Fingerprints []string `json:"fingerprints"`
	index        map[string]bool // For fast lookup
}

// CreateBaseline creates a new baseline from a list of validation errors
func CreateBaseline(issues []cue.ValidationError) *Baseline {
	fingerprints := make([]string, 0, len(issues))
	index := make(map[string]bool)

	for _, issue := range issues {
		fp := fingerprint(issue)
		if !index[fp] {
			fingerprints = append(fingerprints, fp)
			index[fp] = true
		}
	}

	// Sort for deterministic output
	sort.Strings(fingerprints)

	return &Baseline{
		Version:      "1.0",
		Fingerprints: fingerprints,
		index:        index,
	}
}

// LoadBaseline loads a baseline from a JSON file
func LoadBaseline(path string) (*Baseline, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read baseline file: %w", err)
	}

	var b Baseline
	if err := json.Unmarshal(data, &b); err != nil {
		return nil, fmt.Errorf("failed to parse baseline file: %w", err)
	}

	// Build index for fast lookup
	b.index = make(map[string]bool, len(b.Fingerprints))
	for _, fp := range b.Fingerprints {
		b.index[fp] = true
	}

	return &b, nil
}

// SaveBaseline saves the baseline to a JSON file
func (b *Baseline) SaveBaseline(path string) error {
	data, err := json.MarshalIndent(b, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal baseline: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write baseline file: %w", err)
	}

	return nil
}

// IsKnown checks if an issue is in the baseline
func (b *Baseline) IsKnown(issue cue.ValidationError) bool {
	if b.index == nil {
		return false
	}
	fp := fingerprint(issue)
	return b.index[fp]
}

// fingerprint creates a stable hash of an issue for comparison
// Uses: file path + source + normalized message pattern
func fingerprint(issue cue.ValidationError) string {
	// Normalize the message to create a stable pattern
	// Remove specific values that might change while keeping the pattern
	msg := normalizeMessage(issue.Message)

	// Create fingerprint from file + source + normalized message
	// We don't include line numbers as they may shift
	data := fmt.Sprintf("%s|%s|%s", issue.File, issue.Source, msg)

	hash := sha256.Sum256([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// normalizeMessage normalizes error messages to create stable patterns
// Replaces specific values with placeholders to match similar issues
func normalizeMessage(msg string) string {
	// Replace double-quoted strings with placeholder
	msg = regexp.MustCompile(`"[^"]+"`).ReplaceAllString(msg, `"*"`)

	// Replace single-quoted strings with placeholder
	// Match only when surrounded by whitespace/start/end to avoid contractions
	msg = regexp.MustCompile(`(^|\s)'([^']+)'(\s|$)`).ReplaceAllString(msg, `$1'*'$3`)

	// Replace numbers with placeholder (except version-like patterns)
	msg = regexp.MustCompile(`\b\d+\b`).ReplaceAllString(msg, `N`)

	// Normalize whitespace
	msg = strings.Join(strings.Fields(msg), " ")

	return msg
}
