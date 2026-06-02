package lint

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

var reflectSlugPattern = regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+){3,9}\.md$`)

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

// CheckReflectOutput validates every *.md under <root>/kb/ against deterministic
// /dc:reflect output rules: slug shape, H1+source presence, body-size bounds.
// Walks kb/ directly because DefaultFileTypes discovery does not cover kb/.
func CheckReflectOutput(rootPath string) []cue.ValidationError {
	var errors []cue.ValidationError

	kbRoot := filepath.Join(rootPath, "kb")
	if info, err := os.Stat(kbRoot); err != nil || !info.IsDir() {
		return nil
	}

	_ = filepath.WalkDir(kbRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() || !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		if reflectSkipNames[strings.ToLower(d.Name())] {
			return nil
		}

		rel, relErr := filepath.Rel(rootPath, path)
		if relErr != nil {
			rel = path
		}

		if !reflectSlugPattern.MatchString(d.Name()) {
			errors = append(errors, cue.ValidationError{
				File:     rel,
				Message:  "KB filename '" + d.Name() + "' must be a 4-10 word lowercase-hyphenated slug (e.g. 'race-condition-in-channel-close.md')",
				Severity: "warning",
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
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
			})
		}
		if !strings.Contains(content, "(source:") {
			errors = append(errors, cue.ValidationError{
				File:     rel,
				Message:  "KB entry is missing a source attribution (a line containing '(source:')",
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
			})
		}

		n := countNonEmptyLines(content)
		if n < ReflectMinBodyLines {
			errors = append(errors, cue.ValidationError{
				File:     rel,
				Message:  "KB entry has only " + strconv.Itoa(n) + " non-empty lines (< " + strconv.Itoa(ReflectMinBodyLines) + ") — candidate to fold into another entry",
				Severity: "warning",
				Source:   cue.SourceCClintObserve,
			})
		} else if n > ReflectMaxBodyLines {
			errors = append(errors, cue.ValidationError{
				File:     rel,
				Message:  "KB entry has " + strconv.Itoa(n) + " non-empty lines (> " + strconv.Itoa(ReflectMaxBodyLines) + ") — candidate to split",
				Severity: "warning",
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
