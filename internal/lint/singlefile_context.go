package lint

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

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
	return newSingleFileLinterContext(SingleFileRequest{
		FilePath:     filePath,
		RootPath:     rootPath,
		TypeOverride: typeOverride,
		Quiet:        quiet,
		Verbose:      verbose,
	})
}

func newSingleFileLinterContext(req SingleFileRequest) (*SingleFileLinterContext, error) {
	// Validate file path (checks existence, not dir, not binary, etc.)
	//
	// Resolve a relative file path against an explicitly-set --root (default is ""),
	// so `cclint --root /DIR ./file.md` finds the file under /DIR, not cwd.
	filePath := req.FilePath
	if req.RootPath != "" && !filepath.IsAbs(filePath) {
		filePath = filepath.Join(req.RootPath, filePath)
	}
	absPath, err := discovery.ValidateFilePath(filePath)
	if err != nil {
		return nil, err
	}

	// Find project root if not provided
	rootPath := req.RootPath
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
	if req.TypeOverride != "" {
		fileType, err = discovery.ParseFileType(req.TypeOverride)
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
	info, err := os.Stat(absPath)
	if err != nil {
		return nil, fmt.Errorf("cannot stat file %s: %w", absPath, err)
	}

	file := discovery.File{
		Path:     absPath,
		RelPath:  relPath,
		Size:     info.Size(),
		Type:     fileType,
		Contents: string(contents),
	}

	// Schema load is shared only within a LintFiles invocation. Each file still
	// receives its own warning when that load fails.
	var validator *cue.Validator
	var schemaErr error
	if req.invocation != nil {
		validator, schemaErr = req.invocation.cueValidator()
	} else {
		validator = cue.NewValidator()
		schemaErr = validator.LoadSchemas()
	}
	var warnings []cue.ValidationError
	if schemaErr != nil {
		warnings = append(warnings, cue.ValidationError{
			File:     relPath,
			Message:  fmt.Sprintf("CUE schemas not loaded, using Go validation: %v", schemaErr),
			Severity: cue.SeverityWarning,
			Source:   cue.SourceCClintObserve,
		})
	}

	return &SingleFileLinterContext{
		RootPath:       rootPath,
		File:           file,
		Quiet:          req.Quiet,
		Verbose:        req.Verbose,
		Validator:      validator,
		Warnings:       warnings,
		discoveryCache: req.DiscoveryCache,
		invocation:     req.invocation,
	}, nil
}
