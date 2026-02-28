package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dotcommander/cclint/internal/lint"
	"github.com/dotcommander/cclint/internal/cue"
)

// MarkdownFormatter formats output as Markdown
type MarkdownFormatter struct {
	quiet      bool
	verbose    bool
	outputFile string
}

// NewMarkdownFormatter creates a new MarkdownFormatter
func NewMarkdownFormatter(quiet, verbose bool, outputFile string) *MarkdownFormatter {
	return &MarkdownFormatter{
		quiet:      quiet,
		verbose:    verbose,
		outputFile: outputFile,
	}
}

// Format formats the lint summary as Markdown
func (f *MarkdownFormatter) Format(summary *lint.LintSummary) error {
	var builder strings.Builder

	f.writeHeader(&builder, summary)
	f.writeSummaryTable(&builder, summary)
	f.writeDetailedResults(&builder, summary)
	f.writeConclusion(&builder, summary)

	return f.writeOutput(builder.String())
}

func (f *MarkdownFormatter) writeHeader(builder *strings.Builder, summary *lint.LintSummary) {
	builder.WriteString("# CCLint Report\n\n")
	builder.WriteString(fmt.Sprintf("**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05")))
	builder.WriteString(fmt.Sprintf("**Project:** %s\n\n", detectProjectRootForMarkdown()))
	builder.WriteString(fmt.Sprintf("**Duration:** %v\n\n", time.Since(summary.StartTime).Round(time.Millisecond)))
	builder.WriteString(strings.Repeat("-", 50) + "\n\n")
}

func (f *MarkdownFormatter) writeSummaryTable(builder *strings.Builder, summary *lint.LintSummary) {
	builder.WriteString("## Summary\n\n")
	builder.WriteString("| Metric | Count |\n")
	builder.WriteString("|--------|-------|\n")
	builder.WriteString(fmt.Sprintf("| Files Scanned | %d |\n", summary.TotalFiles))
	builder.WriteString(fmt.Sprintf("| Successful | %d |\n", summary.SuccessfulFiles))
	builder.WriteString(fmt.Sprintf("| Failed | %d |\n", summary.FailedFiles))
	builder.WriteString(fmt.Sprintf("| Errors | %d |\n", summary.TotalErrors))
	builder.WriteString(fmt.Sprintf("| Warnings | %d |\n", summary.TotalWarnings))
	builder.WriteString("\n")
}

func (f *MarkdownFormatter) writeDetailedResults(builder *strings.Builder, summary *lint.LintSummary) {
	builder.WriteString("## Detailed Results\n\n")

	if summary.TotalFiles == 0 {
		builder.WriteString("*No files found to validate.*\n")
		return
	}

	f.writeTableOfContents(builder, summary)
	f.writeFileResults(builder, summary)
}

func (f *MarkdownFormatter) writeTableOfContents(builder *strings.Builder, summary *lint.LintSummary) {
	if summary.TotalFiles <= 1 {
		return
	}
	builder.WriteString("### Files\n\n")
	for _, result := range summary.Results {
		fileName := strings.TrimPrefix(result.File, "./")
		builder.WriteString(fmt.Sprintf("- [%s](#%s)\n", fileName, createAnchor(fileName)))
	}
	builder.WriteString("\n")
}

func (f *MarkdownFormatter) writeFileResults(builder *strings.Builder, summary *lint.LintSummary) {
	for _, result := range summary.Results {
		if !f.verbose && result.Success {
			continue
		}

		fileName := strings.TrimPrefix(result.File, "./")
		builder.WriteString(fmt.Sprintf("### %s\n\n", fileName))
		builder.WriteString(fmt.Sprintf("Status: %s\n\n", getStatusEmoji(result.Success)))
		builder.WriteString(fmt.Sprintf("Type: `%s`\n\n", result.Type))

		f.writeIssues(builder, result.Errors, "Errors")
		f.writeIssues(builder, result.Warnings, "Warnings")

		if !f.verbose {
			builder.WriteString("---\n\n")
		}
	}
}

func (f *MarkdownFormatter) writeIssues(builder *strings.Builder, issues []cue.ValidationError, title string) {
	if len(issues) == 0 {
		return
	}
	builder.WriteString(fmt.Sprintf("#### %s\n\n", title))
	for _, issue := range issues {
		builder.WriteString(fmt.Sprintf("- **%s** - %s", issue.File, issue.Message))
		if issue.Line > 0 {
			builder.WriteString(fmt.Sprintf(" (line %d)", issue.Line))
		}
		if issue.Source != "" {
			builder.WriteString(fmt.Sprintf(" `[%s]`", formatSourceTag(issue.Source)))
		}
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

func (f *MarkdownFormatter) writeConclusion(builder *strings.Builder, summary *lint.LintSummary) {
	builder.WriteString("## Conclusion\n\n")
	if summary.FailedFiles == 0 {
		builder.WriteString("✓ All files passed validation!\n")
	} else {
		builder.WriteString(fmt.Sprintf("✗ %d files failed validation\n", summary.FailedFiles))
	}
}

func (f *MarkdownFormatter) writeOutput(content string) error {
	if f.outputFile != "" {
		if err := os.WriteFile(f.outputFile, []byte(content), 0600); err != nil {
			return fmt.Errorf("error writing to file %s: %w", f.outputFile, err)
		}
		return nil
	}
	fmt.Print(content)
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
