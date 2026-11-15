package crossfile

import (
	"os"
	"path/filepath"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

func (v *CrossFileValidator) validateSingleSkillRefs(root, relPath, contents string) []cue.ValidationError {
	relDir := filepath.ToSlash(filepath.Join(filepath.Dir(relPath), "references"))
	refsDir := filepath.Join(root, relDir)
	selector := v.referenceDiscoverer(root)
	if !selector.AllowsPath(filepath.Dir(relDir)) {
		return nil
	}
	// A missing directory still permits genuine phantom-reference diagnostics;
	// an existing but policy-ineligible directory must not be inspected at all.
	if _, err := os.Lstat(refsDir); err == nil && !selector.AllowsPath(relDir) {
		return nil
	}
	var actual, mentioned []string
	for _, name := range listActualRefs(refsDir) {
		if selector.AllowsPath(filepath.Join(relDir, name)) {
			actual = append(actual, name)
		}
	}
	for _, name := range extractMentionedRefs(contents) {
		relative := filepath.ToSlash(filepath.Join(relDir, name))
		if discovery.IsExcluded(relative, v.referencePolicy.Exclude) {
			continue
		}
		if _, err := os.Lstat(filepath.Join(root, relative)); err == nil && !selector.AllowsPath(relative) {
			continue
		}
		mentioned = append(mentioned, name)
	}
	issues := phantomRefErrors(relPath, refsDir, mentioned, actual)
	return append(issues, orphanedRefErrors(relPath, refsDir, mentioned, actual)...)
}
