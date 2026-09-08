package output

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/lint"
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
	fmt.Fprintf(builder, "**Generated:** %s\n\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Fprintf(builder, "**Project:** %s\n\n", detectProjectRootForMarkdown())
	fmt.Fprintf(builder, "**Duration:** %v\n\n", time.Since(summary.StartTime).Round(time.Millisecond))
	builder.WriteString(strings.Repeat("-", 50) + "\n\n")
}

func (f *MarkdownFormatter) writeSummaryTable(builder *strings.Builder, summary *lint.LintSummary) {
	builder.WriteString("## Summary\n\n")
	builder.WriteString("| Metric | Count |\n")
	builder.WriteString("|--------|-------|\n")
	fmt.Fprintf(builder, "| Files Scanned | %d |\n", summary.TotalFiles)
	fmt.Fprintf(builder, "| Successful | %d |\n", summary.SuccessfulFiles)
	fmt.Fprintf(builder, "| Failed | %d |\n", summary.FailedFiles)
	fmt.Fprintf(builder, "| Errors | %d |\n", summary.TotalErrors)
	fmt.Fprintf(builder, "| Warnings | %d |\n", summary.TotalWarnings)
	fmt.Fprintf(builder, "| Suggestions | %d |\n", summary.TotalSuggestions)
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
	issues := BuildFlatIssues(summary)
	for i := range summary.Results {
		result := summary.Results[i]
		if !f.shouldRenderResult(result, issuesForResult(issues, i)) {
			continue
		}
		fileName := strings.TrimPrefix(result.File, "./")
		fmt.Fprintf(builder, "- [%s](#%s)\n", fileName, createAnchor(fileName))
	}
	builder.WriteString("\n")
}

func (f *MarkdownFormatter) writeFileResults(builder *strings.Builder, summary *lint.LintSummary) {
	issues := BuildFlatIssues(summary)
	for i := range summary.Results {
		result := summary.Results[i]
		fileIssues := issuesForResult(issues, i)
		if !f.shouldRenderResult(result, fileIssues) {
			continue
		}

		fileName := strings.TrimPrefix(result.File, "./")
		fmt.Fprintf(builder, "### %s\n\n", fileName)
		fmt.Fprintf(builder, "Status: %s\n\n", getStatusEmoji(result.Success))
		fmt.Fprintf(builder, "Type: `%s`\n\n", result.Type)

		f.writeIssues(builder, severityErrors(fileIssues, SeverityError), "Errors")
		f.writeIssues(builder, severityErrors(fileIssues, SeverityWarning), "Warnings")
		f.writeIssues(builder, severityErrors(fileIssues, SeveritySuggestion), "Suggestions")

		if !f.verbose {
			builder.WriteString("---\n\n")
		}
	}
}

func (f *MarkdownFormatter) shouldRenderResult(result lint.LintResult, fileIssues []FlatIssue) bool {
	if f.verbose {
		return true
	}
	return !result.Success || len(fileIssues) > 0
}

func (f *MarkdownFormatter) writeIssues(builder *strings.Builder, issues []cue.ValidationError, title string) {
	if len(issues) == 0 {
		return
	}
	fmt.Fprintf(builder, "#### %s\n\n", title)
	for _, issue := range issues {
		fmt.Fprintf(builder, "- **%s** - %s", issue.File, issue.Message)
		if issue.Line > 0 {
			fmt.Fprintf(builder, " (line %d)", issue.Line)
		}
		if issue.Source != "" {
			fmt.Fprintf(builder, " `[%s]`", formatSourceTag(issue.Source))
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
		fmt.Fprintf(builder, "✗ %d files failed validation\n", summary.FailedFiles)
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
