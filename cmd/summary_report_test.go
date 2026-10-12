package cmd

import (
	"strings"
	"testing"
	"unicode/utf8"

	"github.com/charmbracelet/lipgloss"
)

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
