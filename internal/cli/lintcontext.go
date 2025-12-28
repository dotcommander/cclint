package cli

import (
	"fmt"
	"os"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/project"
)

// LinterContext holds the shared context for all linting operations.
// This follows the Single Responsibility Principle by centralizing
// the initialization and discovery logic used by all linters.
type LinterContext struct {
	RootPath       string
	Quiet          bool
	Verbose        bool
	NoCycleCheck   bool
	Validator      *cue.Validator
	Discoverer     *discovery.FileDiscovery
	Files          []discovery.File
	CrossValidator *CrossFileValidator
}

// NewLinterContext creates a new LinterContext with all dependencies initialized.
// It handles project root detection, schema loading, file discovery, and
// cross-file validator setup.
func NewLinterContext(rootPath string, quiet, verbose, noCycleCheck bool) (*LinterContext, error) {
	// Find project root if not provided
	if rootPath == "" {
		var err error
		rootPath, err = project.FindProjectRoot(".")
		if err != nil {
			return nil, fmt.Errorf("error finding project root: %w", err)
		}
	}

	// Initialize validator
	validator := cue.NewValidator()

	// Load schemas (soft failure - continue with Go validation)
	if err := validator.LoadSchemas(""); err != nil {
		if !quiet {
			fmt.Fprintf(os.Stderr, "Warning: CUE schemas not loaded, using Go validation\n")
		}
	}

	// Initialize discoverer
	discoverer := discovery.NewFileDiscovery(rootPath, false)

	// Discover all files
	files, err := discoverer.DiscoverFiles()
	if err != nil {
		return nil, fmt.Errorf("error discovering files: %w", err)
	}

	// Initialize cross-file validator
	crossValidator := NewCrossFileValidator(files)

	return &LinterContext{
		RootPath:       rootPath,
		Quiet:          quiet,
		Verbose:        verbose,
		NoCycleCheck:   noCycleCheck,
		Validator:      validator,
		Discoverer:     discoverer,
		Files:          files,
		CrossValidator: crossValidator,
	}, nil
}

// FilterFilesByType returns files matching the specified type.
func (ctx *LinterContext) FilterFilesByType(fileType discovery.FileType) []discovery.File {
	var filtered []discovery.File
	for _, file := range ctx.Files {
		if file.Type == fileType {
			filtered = append(filtered, file)
		}
	}
	return filtered
}

// NewSummary creates an initialized LintSummary with the total file count.
func (ctx *LinterContext) NewSummary(totalFiles int) *LintSummary {
	return &LintSummary{
		ProjectRoot: ctx.RootPath,
		TotalFiles:  totalFiles,
	}
}

// LogProcessed is a no-op - console formatter handles file status display.
// Kept for API compatibility with callers.
func (ctx *LinterContext) LogProcessed(filePath string, errorCount int) {
	// No-op: console formatter shows ✓/✗ for each file
}

// LogProcessedWithSuggestions is a no-op - console formatter handles file status display.
// Kept for API compatibility with callers.
func (ctx *LinterContext) LogProcessedWithSuggestions(filePath string, errorCount, suggestionCount int) {
	// No-op: console formatter shows ✓/✗ for each file
}
