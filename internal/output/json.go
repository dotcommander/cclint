package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/cue"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	quiet      bool
	indent     bool
	outputFile string
}

// NewJSONFormatter creates a new JSONFormatter
func NewJSONFormatter(quiet bool, indent bool, outputFile string) *JSONFormatter {
	return &JSONFormatter{
		quiet:      quiet,
		indent:     indent,
		outputFile: outputFile,
	}
}

// Format formats the lint summary as JSON
func (f *JSONFormatter) Format(summary *cli.LintSummary) error {
	report := JSONReport{
		Header: JSONHeader{
			Tool:      "cclint",
			Version:   "1.0.0",
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Summary: JSONSummary{
			TotalFiles:      summary.TotalFiles,
			SuccessfulFiles: summary.SuccessfulFiles,
			FailedFiles:     summary.FailedFiles,
			TotalErrors:     summary.TotalErrors,
			TotalWarnings:   summary.TotalWarnings,
			Duration:        time.Since(summary.StartTime).Round(time.Millisecond).String(),
		},
		Results: convertResults(summary.Results),
	}

	return f.writeJSON(report)
}

// convertResults maps lint results to JSON-serializable form.
func convertResults(results []cli.LintResult) []JSONResult {
	out := make([]JSONResult, len(results))
	for i, r := range results {
		out[i] = convertResult(r)
	}
	return out
}

// convertResult maps a single lint result to its JSON representation.
func convertResult(r cli.LintResult) JSONResult {
	jr := JSONResult{
		File:     r.File,
		Type:     r.Type,
		Success:  r.Success,
		Duration: r.Duration,
		Errors:   convertValidationErrors(r.Errors),
		Warnings: convertValidationErrors(r.Warnings),
	}
	if r.Quality != nil {
		jr.Quality = &JSONQualityScore{
			Overall:       r.Quality.Overall,
			Tier:          r.Quality.Tier,
			Structural:    r.Quality.Structural,
			Practices:     r.Quality.Practices,
			Composition:   r.Quality.Composition,
			Documentation: r.Quality.Documentation,
		}
	}
	return jr
}

// convertValidationErrors maps validation errors to JSON form.
func convertValidationErrors(errs []cue.ValidationError) []JSONValidationError {
	if len(errs) == 0 {
		return nil
	}
	out := make([]JSONValidationError, len(errs))
	for i, e := range errs {
		out[i] = JSONValidationError{
			File:     e.File,
			Message:  e.Message,
			Severity: e.Severity,
			Source:   e.Source,
			Line:     e.Line,
			Column:   e.Column,
		}
	}
	return out
}

// writeJSON marshals the report and writes it to file or stdout.
func (f *JSONFormatter) writeJSON(report JSONReport) error {
	var jsonBytes []byte
	var err error

	if f.indent {
		jsonBytes, err = json.MarshalIndent(report, "", "  ")
	} else {
		jsonBytes, err = json.Marshal(report)
	}
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %w", err)
	}

	if f.outputFile != "" {
		if writeErr := os.WriteFile(f.outputFile, jsonBytes, 0600); writeErr != nil {
			return fmt.Errorf("error writing to file %s: %w", f.outputFile, writeErr)
		}
		return nil
	}

	fmt.Println(string(jsonBytes))
	return nil
}

// JSONReport represents the complete JSON report structure
type JSONReport struct {
	Header  JSONHeader   `json:"header"`
	Summary JSONSummary  `json:"summary"`
	Results []JSONResult `json:"results"`
}

// JSONHeader contains report metadata
type JSONHeader struct {
	Tool      string `json:"tool"`
	Version   string `json:"version"`
	Timestamp string `json:"timestamp"`
}

// JSONSummary contains summary statistics
type JSONSummary struct {
	TotalFiles      int    `json:"total_files"`
	SuccessfulFiles int    `json:"successful_files"`
	FailedFiles     int    `json:"failed_files"`
	TotalErrors     int    `json:"total_errors"`
	TotalWarnings   int    `json:"total_warnings"`
	Duration        string `json:"duration"`
}

// JSONResult represents a single file's linting result
type JSONResult struct {
	File     string                `json:"file"`
	Type     string                `json:"type"`
	Success  bool                  `json:"success"`
	Duration int64                 `json:"duration_ms,omitempty"`
	Errors   []JSONValidationError `json:"errors,omitempty"`
	Warnings []JSONValidationError `json:"warnings,omitempty"`
	Quality  *JSONQualityScore     `json:"quality,omitempty"`
}

// JSONQualityScore represents the quality score for a component
type JSONQualityScore struct {
	Overall       int    `json:"overall"`
	Tier          string `json:"tier"`
	Structural    int    `json:"structural"`
	Practices     int    `json:"practices"`
	Composition   int    `json:"composition"`
	Documentation int    `json:"documentation"`
}

// JSONValidationError represents a validation error
type JSONValidationError struct {
	File     string `json:"file"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
	Source   string `json:"source,omitempty"`
	Line     int    `json:"line,omitempty"`
	Column   int    `json:"column,omitempty"`
}
