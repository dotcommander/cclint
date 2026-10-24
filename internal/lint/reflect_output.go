package lint

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

var reflectSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+){1,9}\.md$`)

// Non-entry markdown files that legitimately live under kb/ but are not
// /dc:reflect entries — exempt from slug/source/size rules.
var reflectSkipNames = map[string]bool{
	"readme.md": true,
	"index.md":  true,
	"skill.md":  true,
	"memory.md": true,
}

const (
	ReflectMinBodyLines = 10
	ReflectMaxBodyLines = 500
)

// CheckReflectOutput validates regular Markdown entries under <root>/kb/.
// It does not follow symlinks. Walks kb/ directly because discovery omits kb/.
func CheckReflectOutput(rootPath string) []cue.ValidationError {
	return checkReflectOutput(rootPath, nil)
}

// checkReflectOutput applies configured exclusions before diagnostics or reads.
func checkReflectOutput(rootPath string, exclude []string) []cue.ValidationError {
	var errors []cue.ValidationError

	kbRoot := filepath.Join(rootPath, "kb")
	if info, err := os.Lstat(kbRoot); err != nil || !info.IsDir() {
		return nil
	}

	_ = filepath.WalkDir(kbRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if strings.HasPrefix(d.Name(), ".") {
				// Hidden scratch trees (e.g. kb/.work/) are not entries.
				return filepath.SkipDir
			}
		} else {
			name := strings.ToLower(d.Name())
			if !d.Type().IsRegular() || !strings.HasSuffix(name, ".md") || reflectSkipNames[name] {
				return nil
			}
		}

		rel, relErr := filepath.Rel(rootPath, path)
		if relErr != nil {
			rel = path
		}
		if discovery.IsExcluded(filepath.ToSlash(rel), exclude) {
			if d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			return nil
		}

		if !reflectSlugPattern.MatchString(d.Name()) {
			errors = append(errors, cue.ValidationError{
				File:     rel,
				Message:  "KB filename '" + d.Name() + "' must be a 2-10 word lowercase-hyphenated slug (e.g. 'race-condition-in-channel-close.md')",
				Severity: cue.SeverityWarning,
				Source:   cue.SourceCClintObserve,
			})
		}

		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		content := string(data)

		if !hasReflectH1(content) {
			errors = append(errors, cue.ValidationError{
				File:     rel,
				Message:  "KB entry is missing an H1 heading ('# ...')",
				Severity: cue.SeverityWarning,
				Source:   cue.SourceCClintObserve,
			})
		}
		if !strings.Contains(content, "(source:") {
			// Advisory: curated kb trees (book/course-derived notes with no URL
			// provenance) never carry (source: — only /dc:reflect-mined entries
			// do, and nothing else distinguishes them. Surfaces with -v only.
			errors = append(errors, cue.ValidationError{
				File:     rel,
				Message:  "KB entry is missing a source attribution (a line containing '(source:')",
				Severity: cue.SeveritySuggestion,
				Source:   cue.SourceCClintObserve,
			})
		}

		n := countNonEmptyLines(content)
		// Advisory: entry length is a judgment call (the corpus deliberately
		// holds both one-liner principle entries and long references), not a
		// defect. Surfaces with -v only.
		if n < ReflectMinBodyLines {
			errors = append(errors, cue.ValidationError{
				File:     rel,
				Message:  "KB entry has only " + strconv.Itoa(n) + " non-empty lines (< " + strconv.Itoa(ReflectMinBodyLines) + ") — candidate to fold into another entry",
				Severity: cue.SeveritySuggestion,
				Source:   cue.SourceCClintObserve,
			})
		} else if n > ReflectMaxBodyLines {
			errors = append(errors, cue.ValidationError{
				File:     rel,
				Message:  "KB entry has " + strconv.Itoa(n) + " non-empty lines (> " + strconv.Itoa(ReflectMaxBodyLines) + ") — candidate to split",
				Severity: cue.SeveritySuggestion,
				Source:   cue.SourceCClintObserve,
			})
		}

		return nil
	})

	return errors
}

func hasReflectH1(content string) bool {
	for _, line := range strings.Split(content, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "# ") {
			return true
		}
	}
	return false
}

func countNonEmptyLines(content string) int {
	n := 0
	for _, line := range strings.Split(content, "\n") {
		if strings.TrimSpace(line) != "" {
			n++
		}
	}
	return n
}
