package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dotcommander/cclint/internal/cli"
)

// MarkdownFormatter formats output as Markdown
type MarkdownFormatter struct {
	quiet     bool
	verbose   bool
	outputFile string
}

// NewMarkdownFormatter creates a new MarkdownFormatter
func NewMarkdownFormatter(quiet, verbose bool, outputFile string) *MarkdownFormatter {
	return &MarkdownFormatter{
		quiet:     quiet,
		verbose:   verbose,
		outputFile: outputFile,
	}
}

// Format formats the lint summary as Markdown
func (f *MarkdownFormatter) Format(summary *cli.LintSummary) error {
	var builder strings.Builder

	// Header
	builder.WriteString("# CCLint Report\n\n")
	builder.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("**Project:** %s\n\n", detectProjectRootForMarkdown()))
	builder.WriteString(fmt.Sprintf("**Duration:** %v\n\n", time.Since(summary.StartTime).Round(time.Millisecond)))
	builder.WriteString(strings.Repeat("-", 50) + "\n\n")

	// Summary Table
	builder.WriteString("## Summary\n\n")
	builder.WriteString("| Metric | Count |\n")
	builder.WriteString("|--------|-------|\n")
	builder.WriteString(fmt.Sprintf("| Files Scanned | %d |\n", summary.TotalFiles))
	builder.WriteString(fmt.Sprintf("| Successful | %d |\n", summary.SuccessfulFiles))
	builder.WriteString(fmt.Sprintf("| Failed | %d |\n", summary.FailedFiles))
	builder.WriteString(fmt.Sprintf("| Errors | %d |\n", summary.TotalErrors))
	builder.WriteString(fmt.Sprintf("| Warnings | %d |\n", summary.TotalWarnings))
	builder.WriteString("\n")

	// Detailed Results
	builder.WriteString("## Detailed Results\n\n")

	if summary.TotalFiles == 0 {
		builder.WriteString("*No files found to validate.*\n")
	} else {
		// Table of contents for multiple files
		if summary.TotalFiles > 1 {
			builder.WriteString("### Files\n\n")
			for _, result := range summary.Results {
				fileName := strings.TrimPrefix(result.File, "./")
				builder.WriteString(fmt.Sprintf("- [%s](#%s)\n", fileName, createAnchor(fileName)))
			}
			builder.WriteString("\n")
		}

		// Individual file results
		for _, result := range summary.Results {
			if !f.verbose && result.Success {
				continue // Skip successful files unless verbose
			}

			fileName := strings.TrimPrefix(result.File, "./")
			builder.WriteString(fmt.Sprintf("### %s\n\n", fileName))
			builder.WriteString(fmt.Sprintf("Status: %s\n\n", getStatusEmoji(result.Success)))
			builder.WriteString(fmt.Sprintf("Type: `%s`\n\n", result.Type))

			// Errors
			if len(result.Errors) > 0 {
				builder.WriteString("#### Errors\n\n")
				for _, err := range result.Errors {
					builder.WriteString(fmt.Sprintf("- **%s** - %s", err.File, err.Message))
					if err.Line > 0 {
						builder.WriteString(fmt.Sprintf(" (line %d)", err.Line))
					}
					if err.Source != "" {
						builder.WriteString(fmt.Sprintf(" `[%s]`", formatSourceTag(err.Source)))
					}
					builder.WriteString("\n")
				}
				builder.WriteString("\n")
			}

			// Warnings
			if len(result.Warnings) > 0 {
				builder.WriteString("#### Warnings\n\n")
				for _, warning := range result.Warnings {
					builder.WriteString(fmt.Sprintf("- **%s** - %s", warning.File, warning.Message))
					if warning.Line > 0 {
						builder.WriteString(fmt.Sprintf(" (line %d)", warning.Line))
					}
					if warning.Source != "" {
						builder.WriteString(fmt.Sprintf(" `[%s]`", formatSourceTag(warning.Source)))
					}
					builder.WriteString("\n")
				}
				builder.WriteString("\n")
			}

			if !f.verbose {
				builder.WriteString("---\n\n")
			}
		}
	}

	// Conclusion
	builder.WriteString("## Conclusion\n\n")
	if summary.FailedFiles == 0 {
		builder.WriteString("✓ All files passed validation!\n")
	} else {
		builder.WriteString(fmt.Sprintf("✗ %d files failed validation\n", summary.FailedFiles))
	}

	// Write to file or stdout
	content := builder.String()
	if f.outputFile != "" {
		err := os.WriteFile(f.outputFile, []byte(content), 0644)
		if err != nil {
			return fmt.Errorf("error writing to file %s: %w", f.outputFile, err)
		}
	} else {
		fmt.Print(content)
	}

	return nil
}

// getStatusEmoji returns an emoji for the status
func getStatusEmoji(success bool) string {
	if success {
		return "✅"
	}
	return "❌"
}

// createAnchor creates a markdown-safe anchor
func createAnchor(text string) string {
	// Simple implementation - replace spaces and special chars
	anchor := strings.ToLower(text)
	anchor = strings.ReplaceAll(anchor, " ", "-")
	anchor = strings.ReplaceAll(anchor, ".", "")
	anchor = strings.ReplaceAll(anchor, "/", "-")
	return anchor
}

// detectProjectRootForMarkdown detects and returns the project root
func detectProjectRootForMarkdown() string {
	// This is a simplified version - in practice, would use the project detector
	return "./"
}

// formatSourceTag formats the source tag for display
func formatSourceTag(source string) string {
	switch source {
	case "anthropic-docs":
		return "docs"
	case "cclint-observation":
		return "cclint"
	default:
		return source
	}
}