package crossfile

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"github.com/dotcommander/cclint/internal/cue"
)

// Pre-compiled regex patterns for trigger map detection and parsing.
var (
	// triggerTableHeaderPattern matches a markdown table header row with a "Trigger" column.
	triggerTableHeaderPattern = regexp.MustCompile(`(?im)^\|\s*Trigger[^|]*\|`)

	// triggerSkillCellPattern extracts bare hyphenated names from a table cell.
	// Only matches names that contain a hyphen (all real skill/agent names are hyphenated).
	triggerSkillCellPattern = regexp.MustCompile(`\b([a-z][a-z0-9-]{2,})\b`)

	// triggerTaskPattern matches Task(agent-name) patterns in table cells.
	triggerTaskPattern = regexp.MustCompile(`Task\(\s*` + "`?" + `([a-z0-9][a-z0-9-]*)` + "`?" + `\s*\)`)
)

// referenceFileGlobs are the glob patterns used to discover reference files.
var referenceFileGlobs = []string{
	".claude/skills/*/references/*.md",
	"skills/*/references/*.md",
}

// TriggerRef holds a single skill or agent reference extracted from a trigger map table.
type TriggerRef struct {
	File    string // path to the references/*.md file
	RefType string // "skill" or "agent"
	RefName string // e.g. "arch-database-core"
}

// IsTriggerMap reports whether the file content looks like a trigger routing table.
// The heuristic: the file contains a markdown table with a "Trigger" column header.
func IsTriggerMap(contents string) bool {
	return triggerTableHeaderPattern.MatchString(contents)
}

// IsSeparatorRow reports whether a markdown table line is a separator row (e.g. |---|---|).
func IsSeparatorRow(line string) bool {
	trimmed := strings.TrimSpace(line)
	if !strings.HasPrefix(trimmed, "|") {
		return false
	}
	// A separator row consists only of |, -, :, and whitespace
	for _, ch := range trimmed {
		if ch != '|' && ch != '-' && ch != ':' && ch != ' ' && ch != '\t' {
			return false
		}
	}
	return strings.Contains(trimmed, "-")
}

// IsLikelySkillName returns true when the candidate string looks like a real skill or
// agent name. Real names are always hyphenated (e.g. "arch-database-core").
// Single words without hyphens are almost certainly prose and should be ignored.
func IsLikelySkillName(s string) bool {
	if len(s) < 4 {
		return false
	}
	return strings.Contains(s, "-")
}

// ExtractRefsFromRow extracts TriggerRef values from a single non-separator table data row.
// filePath is the source file for attribution. seen deduplicates refs within the table.
func ExtractRefsFromRow(filePath, row string, seen map[string]bool) []TriggerRef {
	// Split on | to get individual cells. The trigger column (first data column) is skipped.
	cells := strings.Split(row, "|")
	if len(cells) < 3 {
		return nil
	}
	// cells[0] is before the first |, cells[1] is the trigger keyword column — skip both.
	// cells[len-1] may be empty (trailing |), ignore it.
	dataCells := cells[2 : len(cells)-1]

	var refs []TriggerRef
	for _, cell := range dataCells {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			continue
		}

		// First pass: extract Task(agent-name) patterns.
		taskMatches := triggerTaskPattern.FindAllStringSubmatch(cell, -1)
		for _, m := range taskMatches {
			if len(m) < 2 {
				continue
			}
			name := strings.TrimSpace(m[1])
			key := "agent:" + name
			if seen[key] {
				continue
			}
			seen[key] = true
			refs = append(refs, TriggerRef{File: filePath, RefType: "agent", RefName: name})
		}

		// Second pass: extract bare skill names from cells that have no Task() pattern.
		if len(taskMatches) == 0 {
			nameMatches := triggerSkillCellPattern.FindAllStringSubmatch(cell, -1)
			for _, m := range nameMatches {
				if len(m) < 2 {
					continue
				}
				name := m[1]
				if !IsLikelySkillName(name) {
					continue
				}
				key := "skill:" + name
				if seen[key] {
					continue
				}
				seen[key] = true
				refs = append(refs, TriggerRef{File: filePath, RefType: "skill", RefName: name})
			}
		}
	}
	return refs
}

// ParseTriggerTable parses a trigger map file and returns all TriggerRef values found
// in routing columns (all columns except the first "Trigger" keyword column).
func ParseTriggerTable(filePath, contents string) []TriggerRef {
	var refs []TriggerRef
	seen := make(map[string]bool)

	inTable := false
	headerFound := false

	for _, rawLine := range strings.Split(contents, "\n") {
		line := strings.TrimSpace(rawLine)

		if !strings.HasPrefix(line, "|") {
			// Reset table state when we leave a table block
			if inTable {
				inTable = false
				headerFound = false
			}
			continue
		}

		// We are inside a table row.
		if !headerFound {
			// Look for the Trigger column header
			if triggerTableHeaderPattern.MatchString(line) {
				headerFound = true
				inTable = true
			}
			continue
		}

		// Skip separator rows (|---|---|)
		if IsSeparatorRow(line) {
			continue
		}

		refs = append(refs, ExtractRefsFromRow(filePath, line, seen)...)
	}

	return refs
}

// discoverReferenceFiles globs for all reference markdown files under rootPath.
func discoverReferenceFiles(rootPath string) []string {
	fsys := os.DirFS(rootPath)
	var results []string

	for _, pattern := range referenceFileGlobs {
		matches, err := doublestar.Glob(fsys, pattern)
		if err != nil {
			continue
		}
		results = append(results, matches...)
	}

	return results
}

// validateTriggerRef checks a single TriggerRef against the known skills/agents maps
// in the CrossFileValidator. Returns any ghost trigger errors.
func (v *CrossFileValidator) validateTriggerRef(ref TriggerRef) []cue.ValidationError {
	switch ref.RefType {
	case "skill":
		if _, exists := v.skills[ref.RefName]; !exists {
			return []cue.ValidationError{{
				File:     ref.File,
				Message:  fmt.Sprintf("Trigger map references non-existent skill '%s'. Create skills/%s/SKILL.md", ref.RefName, ref.RefName),
				Severity: cue.SeverityError,
				Source:   cue.SourceCClintObserve,
			}}
		}
	case "agent":
		if BuiltInSubagentTypes[ref.RefName] {
			return nil
		}
		if _, exists := v.agents[ref.RefName]; !exists {
			return []cue.ValidationError{{
				File:     ref.File,
				Message:  fmt.Sprintf("Trigger map references non-existent agent '%s'. Create agents/%s.md", ref.RefName, ref.RefName),
				Severity: cue.SeverityError,
				Source:   cue.SourceCClintObserve,
			}}
		}
	}
	return nil
}

// ValidateTriggerMaps discovers all reference files under rootPath, identifies those
// that contain trigger routing tables, and validates that every skill/agent ref in
// those tables resolves to a known file. Ghost trigger errors are returned as
// cue.ValidationError with SeverityError.
func (v *CrossFileValidator) ValidateTriggerMaps(rootPath string) []cue.ValidationError {
	relPaths := discoverReferenceFiles(rootPath)

	var errors []cue.ValidationError
	seen := make(map[string]bool)

	for _, relPath := range relPaths {
		fullPath := rootPath + "/" + relPath
		data, err := os.ReadFile(fullPath) //nolint:gosec // G304: path comes from controlled glob inside rootPath
		if err != nil {
			continue
		}
		contents := string(data)

		if !IsTriggerMap(contents) {
			continue
		}

		refs := ParseTriggerTable(relPath, contents)
		for _, ref := range refs {
			dedupeKey := ref.RefType + ":" + ref.RefName + "@" + ref.File
			if seen[dedupeKey] {
				continue
			}
			seen[dedupeKey] = true
			errors = append(errors, v.validateTriggerRef(ref)...)
		}
	}

	return errors
}
