package output

import (
	"fmt"
	"time"

	"github.com/carlrannaberg/cclint/internal/cli"
	"github.com/carlrannaberg/cclint/internal/cue"
	"github.com/charmbracelet/lipgloss"
)

// ConsoleFormatter formats output for console display
type ConsoleFormatter struct {
	quiet     bool
	verbose   bool
	colorize   bool
	startTime  time.Time
}

// NewConsoleFormatter creates a new ConsoleFormatter
func NewConsoleFormatter(quiet, verbose bool) *ConsoleFormatter {
	return &ConsoleFormatter{
		quiet:   quiet,
		verbose: verbose,
		colorize: true,
		startTime: time.Now(),
	}
}

// Format formats the lint summary for console output
func (f *ConsoleFormatter) Format(summary *cli.LintSummary) error {
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
func (f *ConsoleFormatter) printHeader(summary *cli.LintSummary) {
	// No header in simplified UX
}

// printFileResults prints results for each file
func (f *ConsoleFormatter) printFileResults(summary *cli.LintSummary) {
	for _, result := range summary.Results {
		// Show file if it has errors, warnings, or suggestions
		hasIssues := len(result.Errors) > 0 || len(result.Warnings) > 0 || len(result.Suggestions) > 0
		if !hasIssues && !f.verbose {
			continue
		}

		// Print file header
		status := "✓"
		if len(result.Errors) > 0 {
			status = "✗"
		} else if len(result.Suggestions) > 0 {
			status = "💡"
		}

		var fileStyle lipgloss.Style
		if f.colorize {
			if len(result.Errors) > 0 {
				fileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // red
			} else if len(result.Suggestions) > 0 {
				fileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("7")) // gray
			} else {
				fileStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
			}
		}

		fmt.Printf("%s %s\n", fileStyle.Render(status), result.File)

		// Print errors and warnings
		for _, err := range result.Errors {
			f.printValidationError(err, "error")
		}
		for _, warning := range result.Warnings {
			f.printValidationError(warning, "warning")
		}
		for _, suggestion := range result.Suggestions {
			f.printValidationError(suggestion, "suggestion")
		}
	}
}

// printValidationError prints a validation error with appropriate styling
func (f *ConsoleFormatter) printValidationError(err cue.ValidationError, severity string) {
	var style lipgloss.Style
	if !f.colorize {
		style = lipgloss.NewStyle()
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
	}

	prefix := "    "
	if severity == "error" {
		prefix = "    ✘ "
	} else if severity == "warning" {
		prefix = "    ⚠ "
	} else if severity == "suggestion" {
		prefix = "    💡 "
	}

	if err.Line > 0 {
		fmt.Printf("%s%s:%d: %s\n", prefix, style.Render(err.File), err.Line, err.Message)
	} else {
		fmt.Printf("%s%s: %s\n", prefix, style.Render(err.File), err.Message)
	}
}

// printSummary prints the summary statistics
func (f *ConsoleFormatter) printSummary(summary *cli.LintSummary) {
	if f.quiet {
		return
	}

	// Only show summary if there are issues
	if summary.FailedFiles == 0 && summary.TotalSuggestions == 0 {
		return
	}

	duration := time.Since(f.startTime)
	fmt.Printf("\n%d/%d passed, %d errors, %d suggestions (%v)\n",
		summary.SuccessfulFiles, summary.TotalFiles,
		summary.TotalErrors, summary.TotalSuggestions,
		duration.Round(time.Millisecond))
}

// printConclusion prints the conclusion message
func (f *ConsoleFormatter) printConclusion(summary *cli.LintSummary) {
	if f.quiet {
		return
	}

	// Print newline before conclusion if there were file results
	if len(summary.Results) > 0 {
		fmt.Println()
	}

	if summary.FailedFiles == 0 && summary.TotalSuggestions == 0 {
		if f.colorize {
			style := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
			fmt.Printf("%s\n", style.Render("✓ All passed"))
		} else {
			fmt.Println("✓ All passed")
		}
	}
}