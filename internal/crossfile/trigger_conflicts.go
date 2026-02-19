package crossfile

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
)

// TriggerMapping holds a single trigger keyword paired with its routing target,
// extracted from a trigger map table.
type TriggerMapping struct {
	File    string // source file (relative path)
	Keyword string // trigger keyword (column 1, lowercased)
	Target  string // skill or agent name (from columns 2+)
	RefType string // "skill" or "agent"
}

// triggerTargetEntry holds a resolved target with its source file.
type triggerTargetEntry struct {
	target  string
	refType string
	file    string
}

// ParseTriggerMappings parses a trigger map file and returns all TriggerMapping values,
// pairing each trigger keyword with its routing targets.
// Unlike ParseTriggerTable (which collects all refs regardless of keyword),
// this function preserves the keyword→target relationship.
func ParseTriggerMappings(filePath, contents string) []TriggerMapping {
	var mappings []TriggerMapping

	inTable := false
	headerFound := false
	var routingCols []int

	for _, rawLine := range strings.Split(contents, "\n") {
		line := strings.TrimSpace(rawLine)

		if !strings.HasPrefix(line, "|") {
			if inTable {
				inTable = false
				headerFound = false
				routingCols = nil
			}
			continue
		}

		if !headerFound {
			if triggerTableHeaderPattern.MatchString(line) {
				headerFound = true
				inTable = true
				routingCols = identifyRoutingColumns(line)
			}
			continue
		}

		if IsSeparatorRow(line) {
			continue
		}

		mappings = append(mappings, extractMappingsFromRow(filePath, line, routingCols)...)
	}

	return mappings
}

// extractMappingsFromRow extracts TriggerMapping values from a single table data row.
// cells[1] is the keyword column; routingCols specifies which cell indices to inspect
// as routing targets. When routingCols is nil, all cells after the keyword column are used.
func extractMappingsFromRow(filePath, row string, routingCols []int) []TriggerMapping {
	cells := strings.Split(row, "|")
	if len(cells) < 3 {
		return nil
	}

	// cells[0] is before the first |, cells[1] is the trigger keyword
	keyword := strings.ToLower(strings.TrimSpace(cells[1]))
	if keyword == "" {
		return nil
	}

	// Build the list of cells to inspect as routing targets.
	var targetCells []string
	if len(routingCols) == 0 {
		// Backwards compat: all cells except leading empty and trigger keyword.
		// cells[len-1] may be empty (trailing |), ignore it.
		targetCells = cells[2 : len(cells)-1]
	} else {
		for _, idx := range routingCols {
			if idx < len(cells) {
				targetCells = append(targetCells, cells[idx])
			}
		}
	}

	seen := make(map[string]bool)
	var mappings []TriggerMapping

	for _, cell := range targetCells {
		cell = strings.TrimSpace(cell)
		if cell == "" {
			continue
		}

		// First pass: extract Task(agent-name) patterns
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
			mappings = append(mappings, TriggerMapping{
				File:    filePath,
				Keyword: keyword,
				Target:  name,
				RefType: "agent",
			})
		}

		// Second pass: bare skill names in cells without Task() patterns.
		// Strip file path references first to avoid matching path components as skill names.
		if len(taskMatches) == 0 {
			cleanCell := stripReferencePaths(cell)
			nameMatches := triggerSkillCellPattern.FindAllStringSubmatch(cleanCell, -1)
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
				mappings = append(mappings, TriggerMapping{
					File:    filePath,
					Keyword: keyword,
					Target:  name,
					RefType: "skill",
				})
			}
		}
	}

	return mappings
}

// DetectTriggerConflicts discovers all reference files under rootPath and finds cases
// where the same trigger keyword routes to different targets across files.
// Returns warning-level ValidationErrors for each conflicting keyword.
// Same keyword routing to the same target in multiple files is NOT a conflict.
func (v *CrossFileValidator) DetectTriggerConflicts(rootPath string) []cue.ValidationError {
	relPaths := discoverReferenceFiles(rootPath)

	keywordTargets := make(map[string][]triggerTargetEntry)

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

		mappings := ParseTriggerMappings(relPath, contents)
		for _, m := range mappings {
			keywordTargets[m.Keyword] = append(keywordTargets[m.Keyword], triggerTargetEntry{
				target:  m.Target,
				refType: m.RefType,
				file:    m.File,
			})
		}
	}

	return buildTriggerConflictErrors(keywordTargets)
}

// buildTriggerConflictErrors examines collected keyword→target entries and emits warnings
// for keywords that route to more than one distinct target.
func buildTriggerConflictErrors(keywordTargets map[string][]triggerTargetEntry) []cue.ValidationError {
	var errors []cue.ValidationError

	// Sort keywords for deterministic output
	keywords := make([]string, 0, len(keywordTargets))
	for kw := range keywordTargets {
		keywords = append(keywords, kw)
	}
	sort.Strings(keywords)

	for _, keyword := range keywords {
		entries := keywordTargets[keyword]

		// Sort entries by file for deterministic attribution
		sort.Slice(entries, func(i, j int) bool {
			return entries[i].file < entries[j].file
		})

		// Deduplicate: same target from multiple files is not a conflict.
		// Key includes refType so a skill and agent with the same name are distinct.
		uniqueTargets := make(map[string][]string) // "refType:target" → source files
		for _, e := range entries {
			key := e.refType + ":" + e.target
			uniqueTargets[key] = append(uniqueTargets[key], e.file)
		}

		if len(uniqueTargets) <= 1 {
			continue // no conflict
		}

		errors = append(errors, buildConflictError(keyword, entries[0].file, uniqueTargets))
	}

	return errors
}

// buildConflictError constructs a warning ValidationError for a conflicting keyword.
// uniqueTargets keys are "refType:target" strings.
func buildConflictError(keyword, firstFile string, uniqueTargets map[string][]string) cue.ValidationError {
	// Sort keys for deterministic message
	keys := make([]string, 0, len(uniqueTargets))
	for k := range uniqueTargets {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		refType, target, _ := strings.Cut(k, ":")
		files := uniqueTargets[k]
		sort.Strings(files)
		parts = append(parts, fmt.Sprintf("'%s' (%s, in %s)", target, refType, files[0]))
	}

	return cue.ValidationError{
		File:     firstFile,
		Message:  fmt.Sprintf("Trigger keyword '%s' routes to conflicting targets: %s", keyword, strings.Join(parts, ", ")),
		Severity: cue.SeverityWarning,
		Source:   cue.SourceCClintObserve,
	}
}
