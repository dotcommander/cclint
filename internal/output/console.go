package output

import (
	"fmt"
	"os"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/textutil"
	"golang.org/x/term"
)

// ConsoleFormatter formats output for console display
type ConsoleFormatter struct {
	quiet            bool
	verbose          bool
	colorize         bool
	showScores       bool
	showImprovements bool
}

// NewConsoleFormatter creates a new ConsoleFormatter
func NewConsoleFormatter(quiet, verbose, showScores, showImprovements bool) *ConsoleFormatter {
	return &ConsoleFormatter{
		quiet:            quiet,
		verbose:          verbose,
		colorize:         true,
		showScores:       showScores,
		showImprovements: showImprovements,
	}
}

// Format formats the lint summary for console output
func (f *ConsoleFormatter) Format(summary *lint.LintSummary) error {
	if f.quiet {
		// Only show exit code in quiet mode
		return nil
	}

	// Show header
	f.printHeader(summary)

	// Show file results
	f.printFileResults(summary)

	// Show summary
	f.printSummary(summary)

	// Show conclusion
	f.printConclusion(summary)

	return nil
}

// printHeader prints the report header
func (f *ConsoleFormatter) printHeader(summary *lint.LintSummary) {
	// No header in simplified UX
}

// printFileResults prints results for each file
func (f *ConsoleFormatter) printFileResults(summary *lint.LintSummary) {
	for i := range summary.Results {
		if !f.shouldShowFile(&summary.Results[i]) {
			continue
		}

		f.printFileHeader(&summary.Results[i])
		f.printFileIssues(&summary.Results[i])
		f.printScoreDetails(&summary.Results[i])
		f.printImprovements(&summary.Results[i])
	}
}

// shouldShowFile determines if a file result should be displayed.
func (f *ConsoleFormatter) shouldShowFile(result *lint.LintResult) bool {
	hasIssues := len(result.Errors) > 0 || len(result.Warnings) > 0
	return hasIssues || f.verbose
}

// printFileHeader prints the file header with status icon and quality score.
func (f *ConsoleFormatter) printFileHeader(result *lint.LintResult) {
	status := f.getFileStatus(result)
	fileStyle := f.getFileStyle(result)
	scoreStr := f.formatScoreString(result)

	fmt.Printf("%s %s%s\n", fileStyle.Render(status), result.File, scoreStr)
}

// getFileStatus returns the status icon for a file result.
func (f *ConsoleFormatter) getFileStatus(result *lint.LintResult) string {
	if len(result.Errors) > 0 {
		return "✗"
	}
	if f.verbose && len(result.Suggestions) > 0 {
		return "💡"
	}
	return "✓"
}

// getFileStyle returns the lipgloss style for a file based on its status.
func (f *ConsoleFormatter) getFileStyle(result *lint.LintResult) lipgloss.Style {
	if !f.colorize {
		return lipgloss.NewStyle()
	}

	switch {
	case len(result.Errors) > 0:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // red
	case f.verbose && len(result.Suggestions) > 0:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // gray
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	}
}

// formatScoreString formats the quality score string for display.
func (f *ConsoleFormatter) formatScoreString(result *lint.LintResult) string {
	if !f.showScores || result.Quality == nil {
		return ""
	}

	scoreStyle := f.getScoreStyle(result.Quality.Tier)
	return fmt.Sprintf(" [%s %s]", scoreStyle.Render(result.Quality.Tier), scoreStyle.Render(fmt.Sprintf("%d", result.Quality.Overall)))
}

// getScoreStyle returns the lipgloss style for a quality tier.
func (f *ConsoleFormatter) getScoreStyle(tier string) lipgloss.Style {
	if !f.colorize {
		return lipgloss.NewStyle()
	}

	switch tier {
	case "A":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	case "B":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // blue
	case "C":
		return lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
	default:
		return lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // red
	}
}

// printFileIssues prints all errors, warnings, and suggestions for a file.
func (f *ConsoleFormatter) printFileIssues(result *lint.LintResult) {
	for _, err := range result.Errors {
		f.printValidationError(err, "error")
	}
	for _, warning := range result.Warnings {
		f.printValidationError(warning, "warning")
	}
	if f.verbose {
		for _, suggestion := range result.Suggestions {
			f.printValidationError(suggestion, "suggestion")
		}
	}
}

// printScoreDetails prints detailed score breakdown in verbose mode.
func (f *ConsoleFormatter) printScoreDetails(result *lint.LintResult) {
	if !f.verbose || !f.showScores || result.Quality == nil {
		return
	}

	fmt.Printf("    Score: %d/100 (%s)\n", result.Quality.Overall, result.Quality.Tier)
	fmt.Printf("      Structural: %d/40  Practices: %d/40  Composition: %d/10  Documentation: %d/10\n",
		result.Quality.Structural, result.Quality.Practices, result.Quality.Composition, result.Quality.Documentation)
}

// printImprovements prints improvement suggestions if enabled.
func (f *ConsoleFormatter) printImprovements(result *lint.LintResult) {
	if !f.showImprovements || len(result.Improvements) == 0 {
		return
	}

	impStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // cyan
	ptsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green

	fmt.Printf("    %s\n", impStyle.Render("Improvements:"))
	for _, imp := range result.Improvements {
		f.printImprovement(imp, ptsStyle)
	}
}

// printImprovement prints a single improvement line.
func (f *ConsoleFormatter) printImprovement(imp textutil.ImprovementRecommendation, ptsStyle lipgloss.Style) {
	lineRef := ""
	if imp.Line > 0 {
		lineRef = fmt.Sprintf(" (line %d)", imp.Line)
	}
	fmt.Printf("      %s %s%s\n",
		ptsStyle.Render(fmt.Sprintf("+%d pts:", imp.PointValue)),
		imp.Description,
		lineRef)
}

// printValidationError prints a validation error with appropriate styling
func (f *ConsoleFormatter) printValidationError(err cue.ValidationError, severity string) {
	var style lipgloss.Style
	var sourceStyle lipgloss.Style
	if !f.colorize {
		style = lipgloss.NewStyle()
		sourceStyle = lipgloss.NewStyle()
	} else {
		switch severity {
		case "error":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // red
		case "warning":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
		case "suggestion":
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // gray
		default:
			style = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // gray
		}
		sourceStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("8")).Italic(true) // dim
	}

	var prefix string
	switch severity {
	case "error":
		prefix = "    ✘ "
	case "warning":
		prefix = "    ⚠ "
	case "suggestion":
		prefix = "    💡 "
	default:
		prefix = "    "
	}

	// Format source tag (only show in verbose mode for suggestions)
	sourceTag := ""
	if f.verbose && err.Source != "" {
		if err.Source == cue.SourceAnthropicDocs {
			sourceTag = sourceStyle.Render(" [docs]")
		} else if err.Source == cue.SourceCClintObserve {
			sourceTag = sourceStyle.Render(" [cclint]")
		}
	}

	if err.Line > 0 {
		fmt.Printf("%s%s:%d: %s%s\n", prefix, style.Render(err.File), err.Line, err.Message, sourceTag)
	} else {
		fmt.Printf("%s%s: %s%s\n", prefix, style.Render(err.File), err.Message, sourceTag)
	}
}

// printSummary prints the summary statistics
func (f *ConsoleFormatter) printSummary(summary *lint.LintSummary) {
	if f.quiet {
		return
	}

	// Only show summary if there are errors/warnings, or suggestions in verbose mode
	suggestionsCount := 0
	if f.verbose {
		suggestionsCount = summary.TotalSuggestions
	}
	if summary.FailedFiles == 0 && suggestionsCount == 0 {
		return
	}

	duration := time.Since(summary.StartTime)
	if f.verbose {
		fmt.Printf("\n%d/%d passed, %d errors, %d suggestions (%v)\n",
			summary.SuccessfulFiles, summary.TotalFiles,
			summary.TotalErrors, summary.TotalSuggestions,
			duration.Round(time.Millisecond))
	} else {
		fmt.Printf("\n%d/%d passed, %d errors (%v)\n",
			summary.SuccessfulFiles, summary.TotalFiles,
			summary.TotalErrors,
			duration.Round(time.Millisecond))
	}
}

// printConclusion prints the conclusion message
func (f *ConsoleFormatter) printConclusion(summary *lint.LintSummary) {
	if f.quiet {
		return
	}

	// Check for success (only count suggestions in verbose mode)
	allPassed := summary.FailedFiles == 0
	if f.verbose {
		allPassed = allPassed && summary.TotalSuggestions == 0
	}

	if allPassed {
		componentType := summary.ComponentType + "s" // pluralize: agent -> agents
		if summary.ComponentType == "" {
			componentType = "files"
		}
		msg := fmt.Sprintf("✓ All %d %s passed", summary.TotalFiles, componentType)
		perfectSuccess := summary.TotalErrors == 0 && summary.TotalWarnings == 0 && summary.TotalSuggestions == 0

		switch {
		case f.colorize && perfectSuccess && f.isTTY():
			f.printCelebration(msg)
		case f.colorize:
			style := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
			fmt.Printf("%s\n", style.Render(msg))
		default:
			fmt.Println(msg)
		}
	}

	// Add blank line after each component group for better readability
	fmt.Println()
}

// isTTY returns true if stdout is a terminal
func (f *ConsoleFormatter) isTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

// printCelebration shows a sparkle animation for perfect success.
// Delegates to the package-level helper to avoid code duplication.
func (f *ConsoleFormatter) printCelebration(msg string) {
	printCelebration(msg)
}
