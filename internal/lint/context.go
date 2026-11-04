package lint

import (
	"fmt"
	"os"
	"sync"

	"github.com/dotcommander/cclint/internal/crossfile"
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
	CrossValidator *crossfile.CrossFileValidator
	// Parallel enables concurrent per-file linting inside batch runs.
	// Zero values mean sequential execution for all pre-existing callers.
	Parallel    bool
	Concurrency int
	cycleOnce   sync.Once
	cycles      []crossfile.Cycle
}

// LinterOptions carries run-wide options for batch linting.
type LinterOptions struct {
	Quiet, Verbose, NoCycleCheck bool
	Exclude                      []string
	FollowSymlinks               bool
	Parallel                     bool
	Concurrency                  int
	SkipSchemaValidation         bool
	SchemaExtensions             map[string]any
}

// NewLinterContext creates a new LinterContext with all dependencies initialized.
// It handles project root detection, schema loading, file discovery, and
// cross-file validator setup.
func NewLinterContext(rootPath string, quiet, verbose, noCycleCheck bool, exclude []string, followSymlinks ...bool) (*LinterContext, error) {
	return NewLinterContextWithOptions(rootPath, LinterOptions{
		Quiet:          quiet,
		Verbose:        verbose,
		NoCycleCheck:   noCycleCheck,
		Exclude:        exclude,
		FollowSymlinks: len(followSymlinks) > 0 && followSymlinks[0],
	})
}

// NewLinterContextWithOptions creates a new LinterContext from LinterOptions.
// When SkipSchemaValidation is set no validator is constructed and CUE
// validation is skipped for the run. SchemaExtensions are loaded into the
// validator after the embedded schemas; a load error fails context creation.
func NewLinterContextWithOptions(rootPath string, opts LinterOptions) (*LinterContext, error) {
	// Find project root if not provided
	if rootPath == "" {
		var err error
		rootPath, err = project.FindProjectRoot(".")
		if err != nil {
			return nil, fmt.Errorf("error finding project root: %w", err)
		}
	}

	// Initialize validator unless schema validation is intentionally skipped.
	var validator *cue.Validator
	if !opts.SkipSchemaValidation {
		validator = cue.NewValidator()

		// Load schemas (soft failure - continue with Go validation)
		if err := validator.LoadSchemas(); err != nil {
			if !opts.Quiet {
				fmt.Fprintf(os.Stderr, "Warning: CUE schemas not loaded, using Go validation\n")
			}
		} else if len(opts.SchemaExtensions) > 0 {
			if err := validator.LoadExtensions(rootPath, opts.SchemaExtensions); err != nil {
				return nil, fmt.Errorf("load schema extensions: %w", err)
			}
		}
	}

	// Initialize discoverer
	discoverer := discovery.NewFileDiscovery(rootPath, opts.FollowSymlinks).WithExclude(opts.Exclude)

	// Discover all files
	files, err := discoverer.DiscoverFiles()
	if err != nil {
		return nil, fmt.Errorf("error discovering files: %w", err)
	}

	// Initialize cross-file validator
	crossValidator := crossfile.NewCrossFileValidator(files, rootPath)

	return &LinterContext{
		RootPath:       rootPath,
		Quiet:          opts.Quiet,
		Verbose:        opts.Verbose,
		NoCycleCheck:   opts.NoCycleCheck,
		Validator:      validator,
		Discoverer:     discoverer,
		Files:          files,
		CrossValidator: crossValidator,
		Parallel:       opts.Parallel,
		Concurrency:    opts.Concurrency,
	}, nil
}

// batchCycles computes cycles once per context, only when a postprocessor asks.
// Standalone lint calls construct a fresh context, so they never retain this result.
func (ctx *LinterContext) batchCycles() []crossfile.Cycle {
	if ctx.NoCycleCheck || ctx.CrossValidator == nil {
		return nil
	}
	ctx.cycleOnce.Do(func() { ctx.cycles = ctx.CrossValidator.DetectCycles() })
	return ctx.cycles
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

// LintContext runs linting on CLAUDE.md context files.
func LintContext(rootPath string, quiet bool, verbose bool, noCycleCheck bool, exclude []string) (*LintSummary, error) {
	return lintStandalone(rootPath, quiet, verbose, noCycleCheck, exclude, discovery.FileTypeContext)
}
