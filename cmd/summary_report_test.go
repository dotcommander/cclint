package cmd

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"unicode/utf8"

	"charm.land/lipgloss/v2"
)

var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*m`)

func TestSummaryRowAlignment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		text string
	}{
		{"short", "Components Analyzed: 2"},
		{"quality", "  D/F (<50): 2    (100.0%)  ██████████"},
		{"long", strings.Repeat("long component filename ", 8)},
		{"unicode", "  1. .claude/agents/説明.md D 47"},
		{"ansi", "\x1b[31mD/F (<50)\x1b[0m: 2"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := renderSummaryRow(tt.text)
			if width := lipgloss.Width(got); width != lipgloss.Width(summaryTopBorder) {
				t.Fatalf("row width = %d, border width = %d: %q", width, lipgloss.Width(summaryTopBorder), got)
			}
			if !strings.HasPrefix(got, "║ ") || !strings.HasSuffix(got, " ║") {
				t.Fatalf("missing row edges: %q", got)
			}
			if !utf8.ValidString(got) || strings.Contains(got, "\n") {
				t.Fatalf("invalid or multiline row: %q", got)
			}
			if lipgloss.Width(tt.text) <= lipgloss.Width(summaryTopBorder)-4 && !strings.Contains(got, tt.text) {
				t.Fatalf("lost row content: %q", got)
			}
		})
	}
}

// Serial: printTopIssues writes through the process-global os.Stdout swap.
func TestPrintTopIssuesLabelPunctuation(t *testing.T) {
	summary := &ComponentSummary{TopIssues: map[string]int{"Other issues": 3, "Missing description": 1}}
	output := captureFile(t, os.Stdout, func() {
		printTopIssues(summary, newPrintStyles())
	})
	if strings.Contains(output, ".:") {
		t.Fatalf("double punctuation in TOP ISSUES labels: %q", output)
	}
	if !strings.Contains(output, "Other issues") || !strings.Contains(output, "Missing description") {
		t.Fatalf("lost issue rows: %q", output)
	}
	var ordinal int
	for _, line := range strings.Split(output, "\n") {
		if !strings.Contains(line, "Other issues") && !strings.Contains(line, "Missing description") {
			continue
		}
		ordinal++
		plain := ansiPattern.ReplaceAllString(line, "")
		if !strings.Contains(plain, fmt.Sprintf("  %d. ", ordinal)) {
			t.Fatalf("issue %d label not in \"N. \" style: %q", ordinal, plain)
		}
	}
	if ordinal != 2 {
		t.Fatalf("expected 2 issue rows, got %d: %q", ordinal, output)
	}
}
