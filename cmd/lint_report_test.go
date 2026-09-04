package cmd

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/lint"
)

func TestLintOutcomeAdapters(t *testing.T) {
	summary := &lint.LintSummary{
		TotalErrors:      2,
		TotalWarnings:    3,
		TotalSuggestions: 4,
	}
	got := summaryLintOutcome(summary)
	if got != (lintOutcome{errors: 2, warnings: 3, suggestions: 4}) {
		t.Fatalf("summaryLintOutcome() = %+v", got)
	}

	result := &lint.Result{
		TotalErrors:        5,
		TotalWarnings:      6,
		TotalSuggestions:   7,
		BaselineIgnored:    8,
		ErrorsIgnored:      9,
		SuggestionsIgnored: 10,
	}
	got = fullLintOutcome(result)
	want := lintOutcome{
		errors:      5,
		warnings:    6,
		suggestions: 7,
		baseline:    lintBaselineOutcome{total: 8, errors: 9, suggestions: 10},
	}
	if got != want {
		t.Fatalf("fullLintOutcome() = %+v, want %+v", got, want)
	}

	componentSummary := &lint.LintSummary{TotalErrors: 11, TotalWarnings: 12, TotalSuggestions: 13}
	got = componentLintOutcome(result, componentSummary)
	want = lintOutcome{
		errors:      11,
		warnings:    12,
		suggestions: 13,
		baseline:    lintBaselineOutcome{total: 8, errors: 9, suggestions: 10},
	}
	if got != want {
		t.Fatalf("componentLintOutcome() = %+v, want %+v", got, want)
	}
}

func TestReportLintOutcomeAppliesPostLintClosureInOrder(t *testing.T) {
	restoreStderr := captureStderr(t)

	originalExit := exitFunc
	t.Cleanup(func() { exitFunc = originalExit })
	exitFunc = func(code int) {
		if code != 1 {
			t.Errorf("exit code = %d, want 1", code)
		}
		fmt.Fprintf(os.Stderr, "exit\n")
	}

	cfg := &config.Config{Verbose: true}
	outcome := lintOutcome{
		errors:   1,
		baseline: lintBaselineOutcome{total: 2, errors: 1, suggestions: 1},
	}
	err := reportLintOutcome(cfg, executionOptions{}, outcome, func() error {
		fmt.Fprintln(os.Stderr, "format")
		return nil
	})
	if err != nil {
		_ = restoreStderr()
		t.Fatalf("reportLintOutcome() error = %v", err)
	}

	got := restoreStderr()
	wantOrder := []string{
		"format\n",
		"2 baseline issues ignored (1 errors, 1 suggestions)",
		"Validate suggestions against docs.anthropic.com or docs.claude.com",
		"exit\n",
	}
	offset := 0
	for _, want := range wantOrder {
		index := strings.Index(got[offset:], want)
		if index < 0 {
			t.Fatalf("report closure output missing %q after offset %d: %q", want, offset, got)
		}
		offset += index + len(want)
	}
}

func TestReportLintOutcomeReturnsFormatErrorBeforeFailurePolicy(t *testing.T) {
	originalExit := exitFunc
	t.Cleanup(func() { exitFunc = originalExit })
	exitCalled := false
	exitFunc = func(int) { exitCalled = true }

	formatErr := errors.New("formatter failed")
	err := reportLintOutcome(
		&config.Config{Quiet: true},
		executionOptions{},
		lintOutcome{errors: 1},
		func() error { return formatErr },
	)
	if !errors.Is(err, formatErr) {
		t.Fatalf("reportLintOutcome() error = %v, want %v", err, formatErr)
	}
	if exitCalled {
		t.Fatal("failure policy ran after formatter error")
	}
}

func captureStderr(t *testing.T) func() string {
	t.Helper()

	original := os.Stderr
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stderr pipe: %v", err)
	}
	os.Stderr = write

	return func() string {
		os.Stderr = original
		if err := write.Close(); err != nil {
			t.Fatalf("close stderr pipe writer: %v", err)
		}
		got, err := io.ReadAll(read)
		if err != nil {
			t.Fatalf("read stderr pipe: %v", err)
		}
		if err := read.Close(); err != nil {
			t.Fatalf("close stderr pipe reader: %v", err)
		}
		return string(got)
	}
}
