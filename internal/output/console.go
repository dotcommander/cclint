package output

import (
	"fmt"
	"time"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/charmbracelet/lipgloss"
)

// ConsoleFormatter formats output for console display
type ConsoleFormatter struct {
	quiet            bool
	verbose          bool
	colorize         bool
	showScores       bool
	showImprovements bool
	startTime        time.Time
}

// NewConsoleFormatter creates a new ConsoleFormatter
func NewConsoleFormatter(quiet, verbose, showScores, showImprovements bool) *ConsoleFormatter {
	return &ConsoleFormatter{
		quiet:            quiet,
		verbose:          verbose,
		colorize:         true,
		showScores:       showScores,
		showImprovements: showImprovements,
		startTime:        time.Now(),
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

		// Add quality score if available and enabled
		scoreStr := ""
		if f.showScores && result.Quality != nil {
			scoreStyle := lipgloss.NewStyle()
			if f.colorize {
				switch result.Quality.Tier {
				case "A":
					scoreStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
				case "B":
					scoreStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("12")) // blue
				case "C":
					scoreStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("3")) // yellow
				default:
					scoreStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("9")) // red
				}
			}
			scoreStr = fmt.Sprintf(" [%s %s]", scoreStyle.Render(result.Quality.Tier), scoreStyle.Render(fmt.Sprintf("%d", result.Quality.Overall)))
		}

		fmt.Printf("%s %s%s\n", fileStyle.Render(status), result.File, scoreStr)

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

		// Print score details if verbose and scores enabled
		if f.verbose && f.showScores && result.Quality != nil {
			fmt.Printf("    Score: %d/100 (%s)\n", result.Quality.Overall, result.Quality.Tier)
			fmt.Printf("      Structural: %d/40  Practices: %d/40  Composition: %d/10  Documentation: %d/10\n",
				result.Quality.Structural, result.Quality.Practices, result.Quality.Composition, result.Quality.Documentation)
		}

		// Print improvements if enabled
		if f.showImprovements && len(result.Improvements) > 0 {
			impStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("14")) // cyan
			ptsStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
			fmt.Printf("    %s\n", impStyle.Render("Improvements:"))
			for _, imp := range result.Improvements {
				lineRef := ""
				if imp.Line > 0 {
					lineRef = fmt.Sprintf(" (line %d)", imp.Line)
				}
				fmt.Printf("      %s %s%s\n",
					ptsStyle.Render(fmt.Sprintf("+%d pts:", imp.PointValue)),
					imp.Description,
					lineRef)
			}
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