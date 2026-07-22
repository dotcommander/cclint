package output

import (
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/lint"
)

func TestFormatSummaryMutatesMetadataAndWritesJSON(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "report.json")
	cfg := &config.Config{
		Root:    "/test/root",
		Format:  "json",
		Output:  outputPath,
		Version: "v1.2.3",
	}
	summary := &lint.LintSummary{
		ComponentType:   "commands",
		TotalFiles:      1,
		SuccessfulFiles: 1,
	}

	before := time.Now()
	if err := FormatSummary(cfg, summary); err != nil {
		t.Fatalf("FormatSummary() error = %v", err)
	}
	if summary.StartTime.Before(before) || summary.StartTime.After(time.Now()) {
		t.Fatalf("StartTime = %v, want initialized during FormatSummary", summary.StartTime)
	}
	if summary.ProjectRoot != cfg.Root {
		t.Errorf("ProjectRoot = %q, want %q", summary.ProjectRoot, cfg.Root)
	}

	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read JSON output: %v", err)
	}
	var report JSONReport
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatalf("unmarshal JSON output: %v", err)
	}
	if report.Header.Version != cfg.Version {
		t.Errorf("JSON version = %q, want %q", report.Header.Version, cfg.Version)
	}
	if report.Summary.TotalFiles != summary.TotalFiles {
		t.Errorf("JSON total files = %d, want %d", report.Summary.TotalFiles, summary.TotalFiles)
	}
}

func TestFormatSummaryPreservesStartTimeAndWritesMarkdown(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "report.md")
	existing := time.Date(2024, time.January, 2, 3, 4, 5, 0, time.UTC)
	cfg := &config.Config{Root: "/project", Format: "markdown", Output: outputPath}
	summary := &lint.LintSummary{
		StartTime:       existing,
		ComponentType:   "skills",
		TotalFiles:      1,
		SuccessfulFiles: 1,
	}

	if err := FormatSummary(cfg, summary); err != nil {
		t.Fatalf("FormatSummary() error = %v", err)
	}
	if !summary.StartTime.Equal(existing) {
		t.Errorf("StartTime = %v, want preserved value %v", summary.StartTime, existing)
	}
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read Markdown output: %v", err)
	}
	if content := string(data); !strings.Contains(content, "# CCLint Report") || !strings.Contains(content, "| Files Scanned | 1 |") {
		t.Errorf("Markdown output missing report content:\n%s", content)
	}
}

func TestFormatSummaryConsole(t *testing.T) {
	cfg := &config.Config{Root: "/project", Format: "console"}
	summary := &lint.LintSummary{
		ComponentType:   "agent",
		TotalFiles:      1,
		SuccessfulFiles: 1,
	}

	stdout := captureBoundaryStdout(t, func() {
		if err := FormatSummary(cfg, summary); err != nil {
			t.Fatalf("FormatSummary() error = %v", err)
		}
	})
	if !strings.Contains(stdout, "All 1 agents passed") {
		t.Errorf("console output = %q, want passing summary", stdout)
	}
}

func TestFormatSummaryUnsupportedFormatStillMutatesSummary(t *testing.T) {
	cfg := &config.Config{Root: "/test/root", Format: "xml"}
	summary := &lint.LintSummary{}

	err := FormatSummary(cfg, summary)
	if err == nil || err.Error() != "unsupported format: xml" {
		t.Fatalf("FormatSummary() error = %v, want %q", err, "unsupported format: xml")
	}
	if summary.StartTime.IsZero() {
		t.Error("StartTime was not initialized before format selection")
	}
	if summary.ProjectRoot != cfg.Root {
		t.Errorf("ProjectRoot = %q, want %q", summary.ProjectRoot, cfg.Root)
	}
}

func TestFormatSummaryReturnsOutputFileError(t *testing.T) {
	cfg := &config.Config{
		Format: "json",
		Output: filepath.Join(t.TempDir(), "missing", "report.json"),
	}
	err := FormatSummary(cfg, &lint.LintSummary{})
	if err == nil || !strings.Contains(err.Error(), "error writing to file") {
		t.Fatalf("FormatSummary() error = %v, want output-file error", err)
	}
}

func TestFormatAllAlwaysUsesCompactOutput(t *testing.T) {
	outputPath := filepath.Join(t.TempDir(), "ignored.json")
	cfg := &config.Config{Format: "json", Output: outputPath}
	summaries := []*lint.LintSummary{{
		ComponentType:   "agent",
		TotalFiles:      2,
		SuccessfulFiles: 2,
	}}

	stdout := captureBoundaryStdout(t, func() {
		if err := FormatAll(cfg, summaries, time.Now()); err != nil {
			t.Fatalf("FormatAll() error = %v", err)
		}
	})
	if !strings.Contains(stdout, "PASS  2 files") {
		t.Errorf("full-run output = %q, want compact PASS output", stdout)
	}
	if _, err := os.Stat(outputPath); !os.IsNotExist(err) {
		t.Errorf("configured output file was used: stat error = %v", err)
	}
}

func TestFormatAllQuietWritesNothing(t *testing.T) {
	cfg := &config.Config{Quiet: true, Format: "json", Output: filepath.Join(t.TempDir(), "ignored.json")}
	stdout := captureBoundaryStdout(t, func() {
		if err := FormatAll(cfg, []*lint.LintSummary{{TotalFiles: 1}}, time.Now()); err != nil {
			t.Fatalf("FormatAll() error = %v", err)
		}
	})
	if stdout != "" {
		t.Errorf("quiet full-run output = %q, want empty", stdout)
	}
}

func captureBoundaryStdout(t *testing.T, fn func()) string {
	t.Helper()

	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	original := os.Stdout
	os.Stdout = writer
	t.Cleanup(func() { os.Stdout = original })

	fn()

	if err := writer.Close(); err != nil {
		t.Fatalf("close stdout writer: %v", err)
	}
	os.Stdout = original
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close stdout reader: %v", err)
	}
	return string(data)
}
