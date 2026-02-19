package crossfile

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"

	"github.com/dotcommander/cclint/internal/cue"
)

// referenceFileMentionPattern matches references/filename.md in skill contents.
// Handles: prose mentions, markdown links (references/foo.md), and Read(references/foo.md).
// Does NOT match a bare "references/" without a filename.
var referenceFileMentionPattern = regexp.MustCompile(`references/([a-zA-Z0-9_-]+\.md)`)

// ValidateSkillReferences checks that all references/*.md files mentioned in each skill
// exist on disk (phantom ref = error), and that all files in references/ are mentioned
// by the skill (orphaned ref = info suggestion).
func (v *CrossFileValidator) ValidateSkillReferences(rootPath string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Sort skill names for deterministic output.
	skillNames := make([]string, 0, len(v.skills))
	for name := range v.skills {
		skillNames = append(skillNames, name)
	}
	sort.Strings(skillNames)

	for _, name := range skillNames {
		skill := v.skills[name]
		errors = append(errors, validateSingleSkillRefs(rootPath, skill.RelPath, skill.Contents)...)
	}

	return errors
}

// validateSingleSkillRefs checks a single skill file for phantom and orphaned refs.
func validateSingleSkillRefs(rootPath, relPath, contents string) []cue.ValidationError {
	skillDir := filepath.Join(rootPath, filepath.Dir(relPath))
	refsDir := filepath.Join(skillDir, "references")

	mentioned := extractMentionedRefs(contents)
	actual := listActualRefs(refsDir)

	var errors []cue.ValidationError
	errors = append(errors, phantomRefErrors(relPath, refsDir, mentioned, actual)...)
	errors = append(errors, orphanedRefErrors(relPath, refsDir, mentioned, actual)...)
	return errors
}

// extractMentionedRefs returns a sorted, deduplicated list of reference filenames
// mentioned in the skill content (e.g. "foo.md" from "references/foo.md").
func extractMentionedRefs(contents string) []string {
	seen := make(map[string]bool)
	matches := referenceFileMentionPattern.FindAllStringSubmatch(contents, -1)
	for _, m := range matches {
		if len(m) >= 2 {
			seen[m[1]] = true
		}
	}

	result := make([]string, 0, len(seen))
	for name := range seen {
		result = append(result, name)
	}
	sort.Strings(result)
	return result
}

// listActualRefs returns a sorted list of .md filenames present in refsDir.
// Returns nil when the directory does not exist.
func listActualRefs(refsDir string) []string {
	entries, err := os.ReadDir(refsDir) //nolint:gosec // G304: path derived from rootPath + controlled skill relpath
	if err != nil {
		return nil
	}

	var names []string
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".md" {
			names = append(names, entry.Name())
		}
	}
	sort.Strings(names)
	return names
}

// phantomRefErrors returns errors for references mentioned in the skill but missing on disk.
func phantomRefErrors(relPath, refsDir string, mentioned, actual []string) []cue.ValidationError {
	actualSet := make(map[string]bool, len(actual))
	for _, name := range actual {
		actualSet[name] = true
	}

	var errors []cue.ValidationError
	for _, name := range mentioned {
		if !actualSet[name] {
			errors = append(errors, cue.ValidationError{
				File:     relPath,
				Message:  fmt.Sprintf("references/%s is mentioned but does not exist on disk", name),
				Severity: cue.SeverityError,
				Source:   cue.SourceCClintObserve,
			})
		}
	}
	return errors
}

// orphanedRefErrors returns info suggestions for files in references/ not mentioned by the skill.
func orphanedRefErrors(relPath, refsDir string, mentioned, actual []string) []cue.ValidationError {
	mentionedSet := make(map[string]bool, len(mentioned))
	for _, name := range mentioned {
		mentionedSet[name] = true
	}

	var errors []cue.ValidationError
	for _, name := range actual {
		if !mentionedSet[name] {
			errors = append(errors, cue.ValidationError{
				File:     filepath.Join(filepath.Dir(relPath), "references", name),
				Message:  fmt.Sprintf("references/%s exists but is not mentioned in SKILL.md - add a reference or remove the file", name),
				Severity: cue.SeverityInfo,
				Source:   cue.SourceCClintObserve,
			})
		}
	}
	return errors
}
