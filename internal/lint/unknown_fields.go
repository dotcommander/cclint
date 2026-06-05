package lint

import (
	"fmt"

	"github.com/dotcommander/cclint/internal/cue"
)

// unknownFieldCheck models the per-component divergences of the unknown-field
// lint as data (not as branching code), so adding a component or changing a
// message is a one-site data edit. The three real asymmetries are: the field
// label, the message suffix style, and the line-finder (YAML vs JSON).
type unknownFieldCheck struct {
	known    map[string]bool // valid field set for this component
	label    string          // e.g. "frontmatter field" or "plugin field"
	suffix   string          // pre-rendered suffix appended after the field name (may be empty)
	findLine func(contents, key string) int
}

// checkUnknownFields emits a "suggestion" for every key in data not present in
// c.known. Message shape: "Unknown <label> '<key>'<suffix>". Every existing
// caller's exact message is preserved by constructing c.suffix at the call site.
func checkUnknownFields(data map[string]any, filePath, contents string, c unknownFieldCheck) []cue.ValidationError {
	var errors []cue.ValidationError
	for key := range data {
		if !c.known[key] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown %s '%s'%s", c.label, key, c.suffix),
				Severity: cue.SeveritySuggestion,
				Source:   cue.SourceCClintObserve,
				Line:     c.findLine(contents, key),
			})
		}
	}
	return errors
}
