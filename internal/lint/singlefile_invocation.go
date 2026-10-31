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
	schemaErr error
	cross     map[string]crossFileIndex
}

type crossFileIndex struct {
	validator *crossfile.CrossFileValidator
	err       error
}

func (run *singleFileInvocation) cueValidator() (*cue.Validator, error) {
	if run.validator == nil {
		run.validator = cue.NewValidator()
		run.schemaErr = run.validator.LoadSchemas()
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
