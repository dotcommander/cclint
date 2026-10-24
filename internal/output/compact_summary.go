package output

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dotcommander/cclint/internal/lint"
	"golang.org/x/term"
)

// printMinimalResult prints a single PASS/FAIL line plus errors for the default (non-verbose) path.
func (f *CompactFormatter) printMinimalResult(p summaryLineParams, allErrors []errorEntry, boldStyle lipgloss.Style) {
	duration := time.Since(f.startTime)

	warnings := ""
	if p.totalWarnings > 0 {
		warnings = fmt.Sprintf("  %d %s", p.totalWarnings, pluralizeCount("warning", p.totalWarnings))
	}

	fmt.Println()
	if p.totalErrors == 0 {
		line := fmt.Sprintf("✓ PASS  %d %s%s  %s", p.totalFiles, pluralizeCount("file", p.totalFiles), warnings, formatDuration(duration))
		if f.colorize {
			fmt.Println(p.greenStyle.Render(line))
		} else {
			fmt.Println(line)
		}
	} else {
		successCount := p.totalFiles - countErrorFiles(allErrors)
		line := fmt.Sprintf("✗ FAIL  %d/%d files  %d %s%s  %s",
			successCount, p.totalFiles, p.totalErrors, pluralizeCount("error", p.totalErrors), warnings, formatDuration(duration))
		if f.colorize {
			fmt.Println(p.redStyle.Render(line))
		} else {
			fmt.Println(line)
		}
		f.printAllErrors(allErrors, boldStyle, p.redStyle)
	}
}

// countErrorFiles counts the number of unique files that have at least one error.
func countErrorFiles(errors []errorEntry) int {
	seen := make(map[string]bool)
	for _, e := range errors {
		seen[e.file] = true
	}
	return len(seen)
}

// summaryLineParams groups parameters for printing the summary line.
type summaryLineParams struct {
	totalFiles       int
	totalErrors      int
	totalWarnings    int
	totalSuggestions int
	summaries        []*lint.LintSummary
	greenStyle       lipgloss.Style
	redStyle         lipgloss.Style
}

// printSummaryLine prints the final summary line with celebration for perfect success.
func (f *CompactFormatter) printSummaryLine(p summaryLineParams) {
	duration := time.Since(f.startTime)
	fmt.Println()

	successCount := p.totalFiles - countFilesWithErrors(p.summaries)
	summaryText := fmt.Sprintf("%d/%d passed", successCount, p.totalFiles)

	if p.totalErrors > 0 {
		summaryText += fmt.Sprintf(", %d %s", p.totalErrors, pluralizeCount("error", p.totalErrors))
	}
	if p.totalWarnings > 0 {
		summaryText += fmt.Sprintf(", %d %s", p.totalWarnings, pluralizeCount("warning", p.totalWarnings))
	}
	summaryText += fmt.Sprintf(" (%s)", formatDuration(duration))

	// Perfect success: celebrate!
	perfectSuccess := p.totalErrors == 0 && p.totalWarnings == 0 && p.totalSuggestions == 0
	switch {
	case f.colorize && perfectSuccess && f.isTTY():
		printCelebration(summaryText)
	case f.colorize && p.totalErrors > 0:
		fmt.Printf("%s\n", p.redStyle.Render(summaryText))
	case f.colorize:
		fmt.Printf("%s\n", p.greenStyle.Render(summaryText))
	default:
		fmt.Println(summaryText)
	}
}

// isTTY returns true if stdout is a terminal
func (f *CompactFormatter) isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// countFilesWithErrors counts files that have at least one error.
func countFilesWithErrors(summaries []*lint.LintSummary) int {
	count := 0
	for _, s := range summaries {
		count += s.FailedFiles
	}
	return count
}
