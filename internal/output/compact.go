package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dotcommander/cclint/internal/cli"
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
func NewCompactFormatter(quiet, verbose, showScores, showImprovements bool) *CompactFormatter {
	return &CompactFormatter{
		quiet:            quiet,
		verbose:          verbose,
		colorize:         true,
		showScores:       showScores,
		showImprovements: showImprovements,
		startTime:        time.Now(),
	}
}

// FormatAll formats multiple lint summaries in compact style.
func (f *CompactFormatter) FormatAll(summaries []*cli.LintSummary) error {
	if f.quiet {
		return nil
	}

	// Define styles
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("9"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	boldStyle := lipgloss.NewStyle().Bold(true)

	// Track totals
	var totalFiles, totalErrors, totalSuggestions int
	var allErrors []errorEntry
	var allSuggestions []errorEntry

	// Calculate max component name length for alignment
	maxNameLen := 0
	maxCountLen := 0
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

	// Print component status table
	fmt.Println()
	for _, s := range summaries {
		if s.TotalFiles == 0 {
			continue
		}

		name := pluralize(s.ComponentType)
		padding := strings.Repeat(" ", maxNameLen-len(name))

		var statusIcon, statusText string
		var style lipgloss.Style

		if s.FailedFiles > 0 {
			statusIcon = "✗"
			statusText = fmt.Sprintf("%*d/%d passed", maxCountLen, s.SuccessfulFiles, s.TotalFiles)
			style = redStyle
		} else {
			statusIcon = "✓"
			statusText = fmt.Sprintf("%*d passed", maxCountLen, s.TotalFiles)
			style = greenStyle
		}

		if f.colorize {
			fmt.Printf("  %s %s%s  %s\n",
				style.Render(statusIcon),
				dimStyle.Render(name),
				padding,
				style.Render(statusText))
		} else {
			fmt.Printf("  %s %s%s  %s\n", statusIcon, name, padding, statusText)
		}

		// Collect totals and errors
		totalFiles += s.TotalFiles
		totalErrors += s.TotalErrors
		totalSuggestions += s.TotalSuggestions

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
	}

	// Print errors if any
	if len(allErrors) > 0 {
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

	// Print suggestions if verbose
	if f.verbose && len(allSuggestions) > 0 {
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

	// Print summary line
	duration := time.Since(f.startTime)
	fmt.Println()

	successCount := totalFiles - countFilesWithErrors(summaries)
	summaryText := fmt.Sprintf("%d/%d passed", successCount, totalFiles)

	if totalErrors > 0 {
		summaryText += fmt.Sprintf(", %d %s", totalErrors, pluralizeCount("error", totalErrors))
	}
	summaryText += fmt.Sprintf(" (%s)", formatDuration(duration))

	// Perfect success: celebrate!
	perfectSuccess := totalErrors == 0 && totalSuggestions == 0
	if f.colorize && perfectSuccess && f.isTTY() {
		f.printCelebration(summaryText)
	} else if f.colorize {
		if totalErrors > 0 {
			fmt.Printf("%s\n", redStyle.Render(summaryText))
		} else {
			fmt.Printf("%s\n", greenStyle.Render(summaryText))
		}
	} else {
		fmt.Println(summaryText)
	}

	return nil
}

// isTTY returns true if stdout is a terminal
func (f *CompactFormatter) isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// printCelebration shows a sparkle animation for perfect success
func (f *CompactFormatter) printCelebration(msg string) {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	bold := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)

	frames := []struct {
		text  string
		delay time.Duration
	}{
		{green.Render(msg), 200 * time.Millisecond},
		{yellow.Render("✨ " + msg + " ✨"), 300 * time.Millisecond},
		{bold.Render("🎉 " + msg + " 🎉"), 400 * time.Millisecond},
		{yellow.Render("✨ " + msg + " ✨"), 300 * time.Millisecond},
		{green.Render(msg), 0},
	}

	for i, frame := range frames {
		if i > 0 {
			fmt.Print("\r\033[K")
		}
		fmt.Print(frame.text)
		if frame.delay > 0 {
			time.Sleep(frame.delay)
		}
	}
	fmt.Println()
}

// Format implements Formatter interface for single summary (falls back to verbose style).
func (f *CompactFormatter) Format(summary *cli.LintSummary) error {
	return f.FormatAll([]*cli.LintSummary{summary})
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

// pluralize adds 's' to component type names.
func pluralize(s string) string {
	if s == "" {
		return "files"
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
func countFilesWithErrors(summaries []*cli.LintSummary) int {
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
