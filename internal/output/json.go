package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/dotcommander/cclint/internal/cli"
)

// JSONFormatter formats output as JSON
type JSONFormatter struct {
	quiet    bool
	indent   bool
	outputFile string
}

// NewJSONFormatter creates a new JSONFormatter
func NewJSONFormatter(quiet bool, indent bool, outputFile string) *JSONFormatter {
	return &JSONFormatter{
		quiet:     quiet,
		indent:    indent,
		outputFile: outputFile,
	}
}

// Format formats the lint summary as JSON
func (f *JSONFormatter) Format(summary *cli.LintSummary) error {
	// Create JSON report
	report := JSONReport{
		Header: JSONHeader{
			Tool:      "cclint",
			Version:   "1.0.0",
			Timestamp: time.Now().Format(time.RFC3339),
		},
		Summary: JSONSummary{
			TotalFiles:     summary.TotalFiles,
			SuccessfulFiles: summary.SuccessfulFiles,
			FailedFiles:     summary.FailedFiles,
			TotalErrors:     summary.TotalErrors,
			TotalWarnings:   summary.TotalWarnings,
			Duration:        time.Since(summary.StartTime).Round(time.Millisecond).String(),
		},
		Results: make([]JSONResult, len(summary.Results)),
	}

	// Convert results
	for i, result := range summary.Results {
		jsonResult := JSONResult{
			File:     result.File,
			Type:     result.Type,
			Success:  result.Success,
			Duration: result.Duration,
		}

		for _, err := range result.Errors {
			jsonResult.Errors = append(jsonResult.Errors, JSONValidationError{
				File:     err.File,
				Message:  err.Message,
				Severity: err.Severity,
				Source:   err.Source,
				Line:     err.Line,
				Column:   err.Column,
			})
		}

		for _, warning := range result.Warnings {
			jsonResult.Warnings = append(jsonResult.Warnings, JSONValidationError{
				File:     warning.File,
				Message:  warning.Message,
				Severity: warning.Severity,
				Source:   warning.Source,
				Line:     warning.Line,
				Column:   warning.Column,
			})
		}

		// Add quality score if available
		if result.Quality != nil {
			jsonResult.Quality = &JSONQualityScore{
				Overall:       result.Quality.Overall,
				Tier:          result.Quality.Tier,
				Structural:    result.Quality.Structural,
				Practices:     result.Quality.Practices,
				Composition:   result.Quality.Composition,
				Documentation: result.Quality.Documentation,
			}
		}

		report.Results[i] = jsonResult
	}

	// Marshal JSON
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

	// Write to file or stdout
	if f.outputFile != "" {
		err = os.WriteFile(f.outputFile, jsonBytes, 0644)
		if err != nil {
			return fmt.Errorf("error writing to file %s: %w", f.outputFile, err)
		}
	} else {
		fmt.Println(string(jsonBytes))
	}

	return nil
}

// JSONReport represents the complete JSON report structure
type JSONReport struct {
	Header  JSONHeader `json:"header"`
	Summary JSONSummary `json:"summary"`
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
	TotalFiles     int           `json:"total_files"`
	SuccessfulFiles int         `json:"successful_files"`
	FailedFiles     int           `json:"failed_files"`
	TotalErrors     int           `json:"total_errors"`
	TotalWarnings   int           `json:"total_warnings"`
	Duration        string        `json:"duration"`
}

// JSONResult represents a single file's linting result
type JSONResult struct {
	File     string               `json:"file"`
	Type     string               `json:"type"`
	Success  bool                 `json:"success"`
	Duration int64                `json:"duration_ms,omitempty"`
	Errors   []JSONValidationError `json:"errors,omitempty"`
	Warnings []JSONValidationError `json:"warnings,omitempty"`
	Quality  *JSONQualityScore    `json:"quality,omitempty"`
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