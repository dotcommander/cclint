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
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/project"
)

// DiscoveryCache caches the result of file discovery for a given root path.
// This prevents repeated full-discovery scans when linting multiple files
// in single-file mode (e.g., LintFiles iterating over N files).
//
// Thread-safe via sync.Once: the first call to Get() triggers discovery,
// subsequent calls return the cached result.
type DiscoveryCache struct {
	once  sync.Once
	files []discovery.File
	err   error
}

// Get returns cached discovery results, performing discovery on first call.
// The rootPath is used to create the FileDiscovery instance.
func (dc *DiscoveryCache) Get(rootPath string) ([]discovery.File, error) {
	dc.once.Do(func() {
		discoverer := discovery.NewFileDiscovery(rootPath, false)
		dc.files, dc.err = discoverer.DiscoverFiles()
	})
	return dc.files, dc.err
}

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

	// Lazy-loaded for cross-file validation
	crossValidator *crossfile.CrossFileValidator
	crossLoaded    bool

	// Shared discovery cache (optional, set by LintFiles for multi-file mode)
	discoveryCache *DiscoveryCache
}

// NewSingleFileLinterContext creates a context for linting a single file.
//
// This function performs comprehensive path validation and type detection:
//   - Resolves absolute path
//   - Validates file exists and is readable
//   - Detects project root
//   - Determines component type from path
//   - Reads file contents
//
// Parameters:
//   - filePath: Path to the file (absolute or relative)
//   - rootPath: Project root (empty to auto-detect)
//   - typeOverride: Force component type (empty to auto-detect)
//   - quiet: Suppress non-essential output
//   - verbose: Enable verbose output
//
// Returns an error with actionable message for any validation failure.
func NewSingleFileLinterContext(filePath, rootPath, typeOverride string, quiet, verbose bool) (*SingleFileLinterContext, error) {
	// Validate file path (checks existence, not dir, not binary, etc.)
	absPath, err := discovery.ValidateFilePath(filePath)
	if err != nil {
		return nil, err
	}

	// Find project root if not provided
	if rootPath == "" {
		rootPath, err = findProjectRootForFile(absPath)
		if err != nil {
			return nil, fmt.Errorf("cannot determine project root: %w", err)
		}
	} else {
		// Resolve provided root to absolute
		rootPath, err = filepath.Abs(rootPath)
		if err != nil {
			return nil, fmt.Errorf("invalid root path: %w", err)
		}
	}

	// Determine file type
	var fileType discovery.FileType
	if typeOverride != "" {
		fileType, err = discovery.ParseFileType(typeOverride)
		if err != nil {
			return nil, err
		}
	} else {
		fileType, err = discovery.DetectFileType(absPath, rootPath)
		if err != nil {
			return nil, err
		}
	}

	// Read file contents
	contents, err := os.ReadFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %w", absPath, err)
	}

	// Compute relative path for display
	relPath, err := filepath.Rel(rootPath, absPath)
	if err != nil {
		relPath = filepath.Base(absPath)
	}
	relPath = filepath.ToSlash(relPath) // Normalize for display

	// Get file info for size
	info, _ := os.Stat(absPath)
	var size int64
	if info != nil {
		size = info.Size()
	}

	file := discovery.File{
		Path:     absPath,
		RelPath:  relPath,
		Size:     size,
		Type:     fileType,
		Contents: string(contents),
	}

	// Initialize CUE validator
	validator := cue.NewValidator()
	_ = validator.LoadSchemas("") // Soft failure OK

	return &SingleFileLinterContext{
		RootPath:  rootPath,
		File:      file,
		Quiet:     quiet,
		Verbose:   verbose,
		Validator: validator,
	}, nil
}

// EnsureCrossFileValidator lazy-loads the cross-file validator.
//
// This triggers full file discovery to build the reference indexes needed
// for cross-file validation (agent→skill, command→agent, etc.).
// The result is cached for subsequent calls.
//
// When a DiscoveryCache is set (multi-file mode via LintFiles), discovery
// results are shared across all SingleFileLinterContext instances, avoiding
// redundant N+1 discovery scans.
//
// Returns nil if discovery fails (cross-file validation will be skipped).
func (ctx *SingleFileLinterContext) EnsureCrossFileValidator() *crossfile.CrossFileValidator {
	if ctx.crossLoaded {
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
		ctx.crossValidator = crossfile.NewCrossFileValidator(files)
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
// Exit codes (for CLI):
//   - 0: Success (no errors)
//   - 1: Lint errors found
//   - 2: Invalid invocation (returned as error)
func LintSingleFile(filePath, rootPath, typeOverride string, quiet, verbose bool) (*LintSummary, error) {
	return lintSingleFileWithCache(filePath, rootPath, typeOverride, quiet, verbose, nil)
}

// lintSingleFileWithCache is the internal implementation of LintSingleFile
// that accepts an optional DiscoveryCache for sharing discovery results
// across multiple single-file lint invocations (used by LintFiles).
func lintSingleFileWithCache(filePath, rootPath, typeOverride string, quiet, verbose bool, cache *DiscoveryCache) (*LintSummary, error) {
	ctx, err := NewSingleFileLinterContext(filePath, rootPath, typeOverride, quiet, verbose)
	if err != nil {
		return nil, err
	}

	// Attach shared discovery cache if provided
	ctx.discoveryCache = cache

	summary := &LintSummary{
		ProjectRoot: ctx.RootPath,
		TotalFiles:  1,
		StartTime:   time.Now(),
	}

	// Route to appropriate linter based on type
	var result LintResult
	switch ctx.File.Type {
	case discovery.FileTypeAgent:
		result = lintSingleAgent(ctx)
	case discovery.FileTypeCommand:
		result = lintSingleCommand(ctx)
	case discovery.FileTypeSkill:
		result = lintSingleSkill(ctx)
	case discovery.FileTypeSettings:
		result = lintSingleSettings(ctx)
	case discovery.FileTypeContext:
		result = lintSingleContext(ctx)
	case discovery.FileTypePlugin:
		result = lintSinglePlugin(ctx)
	case discovery.FileTypeOutputStyle:
		result = lintSingleOutputStyle(ctx)
	case discovery.FileTypeRule:
		result = lintSingleRule(ctx)
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ctx.File.Type.String())
	}

	// Update summary
	if result.Success {
		summary.SuccessfulFiles = 1
	} else {
		summary.FailedFiles = 1
	}
	summary.TotalErrors = len(result.Errors)
	summary.TotalWarnings = len(result.Warnings)
	summary.TotalSuggestions = len(result.Suggestions)
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

	summary := &LintSummary{
		StartTime: time.Now(),
	}

	// Shared discovery cache prevents N+1 full-discovery scans when
	// linting multiple files. The first file triggers discovery; all
	// subsequent files reuse the cached result.
	cache := &DiscoveryCache{}

	var firstRoot string

	for _, fp := range filePaths {
		result, err := lintSingleFileWithCache(fp, rootPath, typeOverride, quiet, verbose, cache)
		if err != nil {
			// Record as failed result with error message
			summary.Results = append(summary.Results, LintResult{
				File:    fp,
				Type:    "unknown",
				Success: false,
				Errors: []cue.ValidationError{{
					File:     fp,
					Message:  err.Error(),
					Severity: "error",
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

		// Merge results
		summary.TotalFiles += result.TotalFiles
		summary.SuccessfulFiles += result.SuccessfulFiles
		summary.FailedFiles += result.FailedFiles
		summary.TotalErrors += result.TotalErrors
		summary.TotalWarnings += result.TotalWarnings
		summary.TotalSuggestions += result.TotalSuggestions
		summary.Results = append(summary.Results, result.Results...)
	}

	summary.ProjectRoot = firstRoot
	summary.Duration = time.Since(summary.StartTime).Milliseconds()

	return summary, nil
}

// lintSingleAgent lints a single agent file using the generic linter.
func lintSingleAgent(ctx *SingleFileLinterContext) LintResult {
	return lintComponent(ctx, NewAgentLinter())
}

// lintSingleCommand lints a single command file using the generic linter.
func lintSingleCommand(ctx *SingleFileLinterContext) LintResult {
	return lintComponent(ctx, NewCommandLinter())
}

// lintSingleSkill lints a single skill file using the generic linter.
func lintSingleSkill(ctx *SingleFileLinterContext) LintResult {
	return lintComponent(ctx, NewSkillLinter())
}

// lintSingleSettings lints a single settings file using the generic linter.
func lintSingleSettings(ctx *SingleFileLinterContext) LintResult {
	return lintComponent(ctx, NewSettingsLinter())
}

// lintSingleContext lints a single CLAUDE.md context file using the generic linter.
func lintSingleContext(ctx *SingleFileLinterContext) LintResult {
	return lintComponent(ctx, NewContextLinter())
}

// lintSinglePlugin lints a single plugin.json file using the generic linter.
func lintSinglePlugin(ctx *SingleFileLinterContext) LintResult {
	return lintComponent(ctx, NewPluginLinter(ctx.RootPath))
}

// lintSingleOutputStyle lints a single output style file using the generic linter.
func lintSingleOutputStyle(ctx *SingleFileLinterContext) LintResult {
	return lintComponent(ctx, NewOutputStyleLinter())
}

// lintSingleRule lints a single rule file using the generic linter.
func lintSingleRule(ctx *SingleFileLinterContext) LintResult {
	return lintComponent(ctx, NewRuleLinter())
}

// LooksLikePath determines if an argument looks like a file path rather than a subcommand.
//
// This is used to distinguish between:
//   - cclint agents     (subcommand)
//   - cclint ./agents   (file path)
//
// The detection is conservative to avoid false positives:
//   - Absolute paths: /path, C:\path
//   - Relative paths: ./path, ../path
//   - Contains separator: path/to/file
//   - Has known extension: .md, .json
//
// Known subcommands (agents, commands, etc.) are NOT treated as paths
// even if they match these patterns, unless --file flag is used.
func LooksLikePath(arg string) bool {
	// Absolute Unix path
	if strings.HasPrefix(arg, "/") {
		return true
	}

	// Absolute Windows path (C:\, D:\, etc.)
	if len(arg) > 2 && arg[1] == ':' && (arg[2] == '\\' || arg[2] == '/') {
		return true
	}

	// Explicit relative paths
	if strings.HasPrefix(arg, "./") || strings.HasPrefix(arg, "../") {
		return true
	}
	if strings.HasPrefix(arg, ".\\") || strings.HasPrefix(arg, "..\\") {
		return true
	}

	// Contains path separator (not just a bare word)
	if strings.Contains(arg, "/") || strings.Contains(arg, "\\") {
		return true
	}

	// Has known file extension
	lower := strings.ToLower(arg)
	if strings.HasSuffix(lower, ".md") || strings.HasSuffix(lower, ".json") {
		return true
	}

	return false
}

// IsKnownSubcommand returns true if the argument is a known subcommand name.
// These are protected from being interpreted as file paths.
func IsKnownSubcommand(arg string) bool {
	subcommands := map[string]bool{
		"agents":        true,
		"commands":      true,
		"skills":        true,
		"settings":      true,
		"context":       true,
		"plugins":       true,
		"rules":         true,
		"output-styles": true,
		"help":          true,
		"version":       true,
	}
	return subcommands[strings.ToLower(arg)]
}
