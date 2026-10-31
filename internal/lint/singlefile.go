// Package cli provides single-file linting capabilities for cclint.
//
// This file implements the single-file linting mode, which allows users to
// lint individual files directly (e.g., `cclint ./agents/my-agent.md`)
// rather than running discovery across all component types.
//
// # Design Principles
//
//   - Unambiguous CLI: File paths clearly distinguished from subcommands
//   - Pattern-based type detection: Uses glob matching for reliability
//   - Fail-fast with clear errors: Every edge case has an actionable message
//   - Cross-file aware: Validates outgoing refs, skips orphan detection
//   - Reuses existing validation: Same logic as batch mode
//
// # Usage
//
//	// Lint a single file
//	summary, err := LintSingleFile("./agents/my-agent.md", "", false, false)
//
//	// Lint multiple files
//	summary, err := LintFiles([]string{"a.md", "b.md"}, "", "", false, false)
package lint

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/project"
)

// SingleFileLinterContext holds state for single-file linting operations.
//
// Unlike LinterContext which discovers all files upfront, this context
// is optimized for linting individual files. Cross-file validation
// is lazy-loaded only when needed.
type SingleFileLinterContext struct {
	RootPath  string
	File      discovery.File
	Quiet     bool
	Verbose   bool
	Validator *cue.Validator
	Warnings  []cue.ValidationError

	// Lazy-loaded for cross-file validation
	crossValidator *crossfile.CrossFileValidator
	crossLoaded    bool
	crossLoadErr   error

	// Optional invocation state belongs to LintFiles, never standalone calls.
	discoveryCache *DiscoveryCache
	invocation     *singleFileInvocation
}

// SingleFileRequest groups the inputs needed for a single-file lint run.
type SingleFileRequest struct {
	FilePath       string
	RootPath       string
	TypeOverride   string
	Quiet          bool
	Verbose        bool
	DiscoveryCache *DiscoveryCache
	invocation     *singleFileInvocation
}

// EnsureCrossFileValidator lazy-loads the cross-file validator.
//
// This triggers full file discovery to build the reference indexes needed
// for cross-file validation (agent→skill, command→agent, etc.).
// The result is cached for subsequent calls.
//
// LintFiles shares discovery and reference indexes per root for one invocation.
// Standalone callers can optionally share discovery through DiscoveryCache.
//
// Returns nil if discovery fails (cross-file validation will be skipped).
func (ctx *SingleFileLinterContext) EnsureCrossFileValidator() *crossfile.CrossFileValidator {
	if ctx.crossLoaded {
		return ctx.crossValidator
	}

	if ctx.invocation != nil {
		ctx.crossValidator, ctx.crossLoadErr = ctx.invocation.crossFileValidator(ctx.RootPath)
		ctx.crossLoaded = true
		return ctx.crossValidator
	}

	// Use shared cache if available, otherwise discover directly
	var files []discovery.File
	var err error
	if ctx.discoveryCache != nil {
		files, err = ctx.discoveryCache.Get(ctx.RootPath)
	} else {
		discoverer := discovery.NewFileDiscovery(ctx.RootPath, false)
		files, err = discoverer.DiscoverFiles()
	}

	if err == nil && len(files) > 0 {
		ctx.crossValidator = crossfile.NewCrossFileValidator(files, ctx.RootPath)
	} else {
		ctx.crossLoadErr = err
	}
	ctx.crossLoaded = true

	return ctx.crossValidator
}

// findProjectRootForFile attempts to find the project root for a given file.
// Falls back to inferring from .claude directory structure.
func findProjectRootForFile(absPath string) (string, error) {
	// Try standard project root detection
	dir := filepath.Dir(absPath)
	root, err := project.FindProjectRoot(dir)
	if err == nil {
		return root, nil
	}

	// Fallback: infer from .claude directory structure
	// e.g., /foo/.claude/agents/bar.md → /foo/.claude
	// e.g., /foo/agents/bar.md → /foo
	pathStr := absPath
	if before, _, found := strings.Cut(pathStr, "/.claude/"); found {
		return before + "/.claude", nil
	}

	// Check for component directories
	for _, comp := range []string{"/agents/", "/commands/", "/skills/"} {
		if before, _, found := strings.Cut(pathStr, comp); found {
			return before, nil
		}
	}

	// Last resort: use file's parent directory
	return dir, nil
}

// LintSingleFile lints a single file and returns a summary.
//
// This is the main entry point for single-file linting. It:
//  1. Creates a SingleFileLinterContext with full validation
//  2. Routes to the appropriate type-specific linter
//  3. Returns a LintSummary compatible with existing output formatters
//
// Parameters:
//   - filePath: Path to the file (absolute or relative)
//   - rootPath: Project root (empty to auto-detect)
//   - typeOverride: Force component type (empty to auto-detect)
//   - quiet: Suppress non-essential output
//   - verbose: Enable verbose output
//
// Returns a summary and error; it does not exit. The CLI uses exit code 1 for
// invocation errors or results meeting its --fail-on threshold, and 0 otherwise.
func LintSingleFile(filePath, rootPath, typeOverride string, quiet, verbose bool) (*LintSummary, error) {
	return lintSingleFileRequest(SingleFileRequest{
		FilePath:     filePath,
		RootPath:     rootPath,
		TypeOverride: typeOverride,
		Quiet:        quiet,
		Verbose:      verbose,
	})
}

// lintSingleFileRequest is the internal implementation of LintSingleFile.
func lintSingleFileRequest(req SingleFileRequest) (*LintSummary, error) {
	ctx, err := newSingleFileLinterContext(req)
	if err != nil {
		return nil, err
	}

	summary := &LintSummary{
		ProjectRoot: ctx.RootPath,
		TotalFiles:  1,
		StartTime:   time.Now(),
	}

	entry, ok := LinterForType(ctx.File.Type)
	if !ok {
		return nil, fmt.Errorf("unsupported file type: %s", ctx.File.Type.String())
	}
	result := lintComponent(ctx, entry.New(ctx.RootPath))
	result.Warnings = append(result.Warnings, ctx.Warnings...)
	result.Success = len(result.Errors) == 0

	// Update summary
	applyResultToSummary(summary, result)
	summary.Results = []LintResult{result}
	summary.Duration = time.Since(summary.StartTime).Milliseconds()

	return summary, nil
}

// LintFiles lints multiple files and returns a combined summary.
//
// Each file is validated independently; failures for one file do not
// prevent linting of other files. The summary aggregates all results.
//
// Parameters:
//   - filePaths: Paths to files (absolute or relative)
//   - rootPath: Project root (empty to auto-detect per file)
//   - typeOverride: Force component type (empty to auto-detect)
//   - quiet: Suppress non-essential output
//   - verbose: Enable verbose output
func LintFiles(filePaths []string, rootPath, typeOverride string, quiet, verbose bool) (*LintSummary, error) {
	if len(filePaths) == 0 {
		return nil, fmt.Errorf("no files specified")
	}

	expanded, err := expandDirectoriesAtRoot(filePaths, rootPath, typeOverride)
	if err != nil {
		return nil, err
	}
	if len(expanded) == 0 {
		return nil, fmt.Errorf("no lintable files found")
	}

	summary := &LintSummary{
		StartTime: time.Now(),
	}

	// The state lives only for this call; discovery and reference indexes are
	// shared per root while CUE schemas are shared across all selected files.
	invocation := &singleFileInvocation{}

	var firstRoot string

	for _, fh := range expanded {
		// Use per-file type hint if available, otherwise the global typeOverride
		effectiveType := typeOverride
		if fh.TypeHint != "" {
			effectiveType = fh.TypeHint
		}
		result, err := lintSingleFileRequest(SingleFileRequest{
			FilePath:     fh.Path,
			RootPath:     rootPath,
			TypeOverride: effectiveType,
			Quiet:        quiet,
			Verbose:      verbose,
			invocation:   invocation,
		})
		if err != nil {
			// Record as failed result with error message
			summary.Results = append(summary.Results, LintResult{
				File:    fh.Path,
				Type:    "unknown",
				Success: false,
				Errors: []cue.ValidationError{{
					File:     fh.Path,
					Message:  err.Error(),
					Severity: cue.SeverityError,
				}},
			})
			summary.TotalFiles++
			summary.FailedFiles++
			summary.TotalErrors++
			continue
		}

		// Capture first root for summary
		if firstRoot == "" {
			firstRoot = result.ProjectRoot
		}

		// Merge results — lintSingleFileRequest always produces exactly one result
		summary.TotalFiles += result.TotalFiles
		applyResultToSummary(summary, result.Results[0])
		summary.Results = append(summary.Results, result.Results...)
	}

	summary.ProjectRoot = firstRoot
	summary.Duration = time.Since(summary.StartTime).Milliseconds()

	return summary, nil
}
