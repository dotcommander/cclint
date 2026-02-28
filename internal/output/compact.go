package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/cue"
	"golang.org/x/term"
)

// CompactFormatter formats output in a compact, summary-first style.
// It collects all component results and displays them together.
type CompactFormatter struct {
	quiet            bool
	verbose          bool
	colorize         bool
	showScores       bool
	showImprovements bool
	startTime        time.Time
}

// NewCompactFormatter creates a new CompactFormatter.
func NewCompactFormatter(quiet, verbose, showScores, showImprovements bool, startTime time.Time) *CompactFormatter {
	return &CompactFormatter{
		quiet:            quiet,
		verbose:          verbose,
		colorize:         true,
		showScores:       showScores,
		showImprovements: showImprovements,
		startTime:        startTime,
	}
}

// FormatAll formats multiple lint summaries in compact style.
func (f *CompactFormatter) FormatAll(summaries []*lint.LintSummary) error {
	if f.quiet {
		return nil
	}

	// Define styles
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	boldStyle := lipgloss.NewStyle().Bold(true)

	// Aggregate totals and collect errors/suggestions from all summaries
	var totalFiles, totalErrors, totalSuggestions int
	var allErrors []errorEntry
	var allSuggestions []errorEntry

	for _, s := range summaries {
		if s.TotalFiles == 0 {
			continue
		}
		totalFiles += s.TotalFiles
		totalErrors += s.TotalErrors
		totalSuggestions += s.TotalSuggestions
		allErrors, allSuggestions = f.collectErrorsAndSuggestions(s, allErrors, allSuggestions)
	}

	if f.verbose {
		// Verbose: print full component table + errors + suggestions + summary
		maxNameLen, maxCountLen := f.calculateColumnWidths(summaries)

		fmt.Println()
		for _, s := range summaries {
			if s.TotalFiles == 0 {
				continue
			}

			name := pluralize(s.ComponentType)
			padding := strings.Repeat(" ", maxNameLen-len(name))

			statusInfo := getStatusInfo(s, maxCountLen, greenStyle, redStyle)
			f.printStatusLine(statusLineParams{
				icon:     statusInfo.icon,
				name:     name,
				padding:  padding,
				text:     statusInfo.text,
				style:    statusInfo.style,
				dimStyle: dimStyle,
			})
		}

		f.printAllErrors(allErrors, boldStyle, redStyle)
		f.printAllSuggestions(allSuggestions, dimStyle)
		f.printSummaryLine(summaryLineParams{
			totalFiles:       totalFiles,
			totalErrors:      totalErrors,
			totalSuggestions: totalSuggestions,
			summaries:        summaries,
			greenStyle:       greenStyle,
			redStyle:         redStyle,
		})
	} else {
		// Default: minimal PASS/FAIL line + errors only
		f.printMinimalResult(totalFiles, totalErrors, allErrors, boldStyle, redStyle)
	}

	return nil
}

// printMinimalResult prints a single PASS/FAIL line plus errors for the default (non-verbose) path.
func (f *CompactFormatter) printMinimalResult(totalFiles, totalErrors int, allErrors []errorEntry, boldStyle, redStyle lipgloss.Style) {
	duration := time.Since(f.startTime)
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))

	fmt.Println()
	if totalErrors == 0 {
		line := fmt.Sprintf("✓ PASS  %d %s  %s", totalFiles, pluralizeCount("file", totalFiles), formatDuration(duration))
		if f.colorize {
			fmt.Println(greenStyle.Render(line))
		} else {
			fmt.Println(line)
		}
	} else {
		successCount := totalFiles - countErrorFiles(allErrors)
		line := fmt.Sprintf("✗ FAIL  %d/%d files  %d %s  %s",
			successCount, totalFiles, totalErrors, pluralizeCount("error", totalErrors), formatDuration(duration))
		if f.colorize {
			fmt.Println(redStyle.Render(line))
		} else {
			fmt.Println(line)
		}
		f.printAllErrors(allErrors, boldStyle, redStyle)
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

// calculateColumnWidths computes the maximum name and count column widths.
func (f *CompactFormatter) calculateColumnWidths(summaries []*lint.LintSummary) (maxNameLen, maxCountLen int) {
	for _, s := range summaries {
		name := pluralize(s.ComponentType)
		if len(name) > maxNameLen {
			maxNameLen = len(name)
		}
		countStr := fmt.Sprintf("%d", s.TotalFiles)
		if len(countStr) > maxCountLen {
			maxCountLen = len(countStr)
		}
	}
	return maxNameLen, maxCountLen
}

// statusInfo groups the status icon, text, and style for a summary.
type statusInfo struct {
	icon  string
	text  string
	style lipgloss.Style
}

// getStatusInfo returns the status icon, text, and style for a summary.
func getStatusInfo(s *lint.LintSummary, maxCountLen int, greenStyle, redStyle lipgloss.Style) statusInfo {
	if s.FailedFiles > 0 {
		return statusInfo{
			icon:  "✗",
			text:  fmt.Sprintf("%*d/%d passed", maxCountLen, s.SuccessfulFiles, s.TotalFiles),
			style: redStyle,
		}
	}
	return statusInfo{
		icon:  "✓",
		text:  fmt.Sprintf("%*d passed", maxCountLen, s.TotalFiles),
		style: greenStyle,
	}
}

// statusLineParams groups parameters for printing a status line.
type statusLineParams struct {
	icon      string
	name      string
	padding   string
	text      string
	style     lipgloss.Style
	dimStyle  lipgloss.Style
}

// printStatusLine prints a single status line with appropriate styling.
func (f *CompactFormatter) printStatusLine(p statusLineParams) {
	if f.colorize {
		fmt.Printf("  %s %s%s  %s\n",
			p.style.Render(p.icon),
			p.dimStyle.Render(p.name),
			p.padding,
			p.style.Render(p.text))
	} else {
		fmt.Printf("  %s %s%s  %s\n", p.icon, p.name, p.padding, p.text)
	}
}

// collectErrorsAndSuggestions aggregates errors and suggestions from a summary.
func (f *CompactFormatter) collectErrorsAndSuggestions(s *lint.LintSummary, allErrors, allSuggestions []errorEntry) ([]errorEntry, []errorEntry) {
	for _, result := range s.Results {
		for _, err := range result.Errors {
			allErrors = append(allErrors, errorEntry{
				componentType: s.ComponentType,
				file:          result.File,
				err:           err,
			})
		}
		if f.verbose {
			for _, sugg := range result.Suggestions {
				allSuggestions = append(allSuggestions, errorEntry{
					componentType: s.ComponentType,
					file:          result.File,
					err:           sugg,
				})
			}
		}
	}
	return allErrors, allSuggestions
}

// printAllErrors prints all errors grouped by file.
func (f *CompactFormatter) printAllErrors(allErrors []errorEntry, boldStyle, redStyle lipgloss.Style) {
	if len(allErrors) == 0 {
		return
	}

	fmt.Println()
	if f.colorize {
		fmt.Println(boldStyle.Render("Errors:"))
	} else {
		fmt.Println("Errors:")
	}

	// Group errors by file
	currentFile := ""
	for _, e := range allErrors {
		if e.file != currentFile {
			currentFile = e.file
			if f.colorize {
				fmt.Printf("  %s\n", redStyle.Render(e.file))
			} else {
				fmt.Printf("  %s\n", e.file)
			}
		}
		f.printError(e.err, "error")
	}
}

// printAllSuggestions prints all suggestions grouped by file.
func (f *CompactFormatter) printAllSuggestions(allSuggestions []errorEntry, dimStyle lipgloss.Style) {
	if !f.verbose || len(allSuggestions) == 0 {
		return
	}

	fmt.Println()
	if f.colorize {
		fmt.Println(dimStyle.Render("Suggestions:"))
	} else {
		fmt.Println("Suggestions:")
	}

	currentFile := ""
	for _, e := range allSuggestions {
		if e.file != currentFile {
			currentFile = e.file
			fmt.Printf("  %s\n", e.file)
		}
		f.printError(e.err, "suggestion")
	}
}

// summaryLineParams groups parameters for printing the summary line.
type summaryLineParams struct {
	totalFiles       int
	totalErrors      int
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
	summaryText += fmt.Sprintf(" (%s)", formatDuration(duration))

	// Perfect success: celebrate!
	perfectSuccess := p.totalErrors == 0 && p.totalSuggestions == 0
	switch {
	case f.colorize && perfectSuccess && f.isTTY():
		f.printCelebration(summaryText)
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

// printCelebration shows a sparkle animation for perfect success.
// Delegates to the package-level helper to avoid code duplication.
func (f *CompactFormatter) printCelebration(msg string) {
	printCelebration(msg)
}

// Format implements Formatter interface for single summary (falls back to verbose style).
func (f *CompactFormatter) Format(summary *lint.LintSummary) error {
	return f.FormatAll([]*lint.LintSummary{summary})
}

// printError prints a single error with indentation.
func (f *CompactFormatter) printError(err cue.ValidationError, severity string) {
	var style lipgloss.Style
	if f.colorize {
		switch severity {
		case "error":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
		case "suggestion":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
		default:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("7"))
		}
	}

	prefix := "    ✘ "
	if severity == "suggestion" {
		prefix = "    💡 "
	}

	msg := err.Message
	if f.colorize {
		fmt.Printf("%s%s\n", prefix, style.Render(msg))
	} else {
		fmt.Printf("%s%s\n", prefix, msg)
	}
}

// errorEntry groups an error with its source file and component type.
type errorEntry struct {
	componentType string
	file          string
	err           cue.ValidationError
}

// irregularPlurals maps component type names that don't pluralize by appending 's'.
var irregularPlurals = map[string]string{
	"": "files",
}

// pluralize returns the plural form of a component type name.
func pluralize(s string) string {
	if p, ok := irregularPlurals[s]; ok {
		return p
	}
	return s + "s"
}

// pluralizeCount returns singular or plural form based on count.
func pluralizeCount(s string, count int) string {
	if count == 1 {
		return s
	}
	return s + "s"
}

// countFilesWithErrors counts files that have at least one error.
func countFilesWithErrors(summaries []*lint.LintSummary) int {
	count := 0
	for _, s := range summaries {
		count += s.FailedFiles
	}
	return count
}

// formatDuration formats a duration in a human-readable way.
func formatDuration(d time.Duration) string {
	if d < time.Second {
		return fmt.Sprintf("%dms", d.Milliseconds())
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}
