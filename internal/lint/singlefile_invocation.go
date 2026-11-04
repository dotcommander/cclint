package lint

import (
	"path/filepath"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
)

// singleFileInvocation owns only resources reused during one LintFiles call.
// LintSingleFile and NewSingleFileLinterContext retain their independent state.
type singleFileInvocation struct {
	discovery DiscoveryCache
	validator *cue.Validator
	// schemaErr is a soft LoadSchemas failure (per-file warning); extensionErr
	// is fatal (extension files fail the run with the file named).
	schemaErr        error
	extensionErr     error
	schemaInit       bool
	skipSchema       bool
	schemaExtensions map[string]any
	cross            map[string]crossFileIndex
}

type crossFileIndex struct {
	validator *crossfile.CrossFileValidator
	err       error
}

// cueValidator returns the shared validator, nil when schema validation is
// skipped. The returned error is the soft schema-load failure; extension load
// failures are stored separately and fail the run via extensionErr.
func (run *singleFileInvocation) cueValidator(rootPath string) (*cue.Validator, error) {
	if run.skipSchema {
		return nil, nil
	}
	if !run.schemaInit {
		run.schemaInit = true
		run.validator = cue.NewValidator()
		run.schemaErr = run.validator.LoadSchemas()
		if run.schemaErr == nil && len(run.schemaExtensions) > 0 {
			run.extensionErr = run.validator.LoadExtensions(rootPath, run.schemaExtensions)
		}
	}
	return run.validator, run.schemaErr
}

func (run *singleFileInvocation) crossFileValidator(rootPath string) (*crossfile.CrossFileValidator, error) {
	root, err := filepath.Abs(rootPath)
	if err != nil {
		return nil, err
	}
	if index, ok := run.cross[root]; ok {
		return index.validator, index.err
	}

	files, err := run.discovery.Get(root)
	var validator *crossfile.CrossFileValidator
	if err == nil && len(files) > 0 {
		validator = crossfile.NewCrossFileValidator(files, root)
	}
	if run.cross == nil {
		run.cross = make(map[string]crossFileIndex)
	}
	run.cross[root] = crossFileIndex{validator: validator, err: err}
	return validator, err
}
