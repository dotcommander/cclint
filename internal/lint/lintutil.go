// lintutil.go holds stateless lint helpers with no dependency on the linter
// interface types in linter_interfaces.go — extracted to keep that file under
// the size tripwire.
package lint

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/textutil"
)

// knownFrontmatterFields maps component types to their known field names.
// Used by DetectSwallowedFields to identify when a block scalar has absorbed
// what should be a sibling frontmatter field.
var knownFrontmatterFields = map[string]map[string]bool{
	"agent":   knownAgentFields,
	"command": knownCommandFields,
	"skill":   knownSkillFields,
}

// blockScalarPattern matches YAML block scalar indicators (| or >) with optional modifiers.
var blockScalarPattern = regexp.MustCompile(`^([a-zA-Z][a-zA-Z0-9_-]*)\s*:\s*[|>][-+]?\s*$`)

// DetectSwallowedFields checks for YAML block scalar fields (| or >) that have
// accidentally absorbed subsequent frontmatter fields as part of their text content.
//
// Example of the bug this catches:
//
//	---
//	description: |
//	  Some text here.
//	model: haiku        ← YAML treats this as part of description, not a separate field
//	---
//
// This happens when a new field is inserted after a block scalar without proper
// indentation awareness. The YAML parser silently absorbs the new field as text.
func DetectSwallowedFields(contents, filePath, componentType string) []cue.ValidationError {
	lines := strings.Split(contents, "\n")

	fmStart, fmEnd, ok := findFrontmatterBounds(lines)
	if !ok {
		return nil
	}

	// Only check component types with known field sets.
	// Types like settings, context, and plugin don't have block scalar
	// frontmatter fields, so checking them risks false positives.
	fields := knownFrontmatterFields[componentType]
	if fields == nil {
		return nil
	}

	fmLines := lines[fmStart+1 : fmEnd]
	return detectSwallowedBlockScalarFields(fmLines, fmStart, filePath, fields)
}

func findFrontmatterBounds(lines []string) (int, int, bool) {
	fmStart, fmEnd := -1, -1
	for i, line := range lines {
		if strings.TrimSpace(line) != "---" {
			continue
		}
		if fmStart == -1 {
			fmStart = i
			continue
		}
		fmEnd = i
		break
	}
	if fmStart == -1 || fmEnd == -1 {
		return 0, 0, false
	}
	return fmStart, fmEnd, true
}

func detectSwallowedBlockScalarFields(fmLines []string, fmStart int, filePath string, fields map[string]bool) []cue.ValidationError {
	var errors []cue.ValidationError

	for i, line := range fmLines {
		match := blockScalarPattern.FindStringSubmatch(line)
		if match == nil {
			continue
		}
		scalarField := match[1]
		errors = append(errors, findSwallowedFieldsForScalar(fmLines, fmStart, filePath, fields, scalarField, i)...)
	}

	return errors
}

func findSwallowedFieldsForScalar(fmLines []string, fmStart int, filePath string, fields map[string]bool, scalarField string, scalarIndex int) []cue.ValidationError {
	var errors []cue.ValidationError

	for j := scalarIndex + 1; j < len(fmLines); j++ {
		subsequent := fmLines[j]
		if strings.TrimSpace(subsequent) == "" {
			continue
		}
		if !isIndentedLine(subsequent) {
			break
		}

		candidateKey, ok := swallowedFieldCandidate(subsequent, fields)
		if !ok {
			continue
		}

		lineNum := fmStart + j + 2
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Field '%s' appears to be swallowed by block scalar '%s: |' above — it is parsed as text, not a separate field", candidateKey, scalarField),
			Severity: cue.SeverityError,
			Source:   cue.SourceCClintObserve,
			Line:     lineNum,
		})
	}

	return errors
}

func isIndentedLine(line string) bool {
	return len(line) > 0 && (line[0] == ' ' || line[0] == '\t')
}

func swallowedFieldCandidate(line string, fields map[string]bool) (string, bool) {
	trimmed := strings.TrimSpace(line)
	colonIdx := strings.Index(trimmed, ":")
	if colonIdx <= 0 {
		return "", false
	}

	candidateKey := trimmed[:colonIdx]
	if !fields[candidateKey] {
		return "", false
	}

	return candidateKey, true
}

// xmlTagPattern matches XML-like tags (e.g. <tag>, <br/>) — hoisted to avoid recompilation.
var xmlTagPattern = regexp.MustCompile(`<[a-zA-Z][^>]*>`)

// descriptionFieldError is the shared constructor for angle-bracket validation errors.
// Both DetectXMLTags and DetectBareAngleBrackets delegate here to stay in sync.
func descriptionFieldError(fieldName, message, filePath, fileContents string) *cue.ValidationError {
	return &cue.ValidationError{
		File:     filePath,
		Message:  fmt.Sprintf("%s %s", fieldName, message),
		Severity: cue.SeverityError,
		Source:   cue.SourceAnthropicDocs,
		Line:     textutil.FindFrontmatterFieldLine(fileContents, strings.ToLower(fieldName)),
	}
}

// DetectXMLTags checks for XML-like tags in a string and returns an error if found.
// Used by agents and commands. Skills should use DetectBareAngleBrackets instead,
// which covers a strict superset of inputs (all angle brackets, not just XML tags).
func DetectXMLTags(content, fieldName, filePath, fileContents string) *cue.ValidationError {
	if xmlTagPattern.MatchString(content) {
		return descriptionFieldError(fieldName, "contains XML-like tags which are not allowed", filePath, fileContents)
	}
	return nil
}

// DetectBareAngleBrackets checks for any bare < or > characters in a string.
// Anthropic's official validator (quick_validate.py) rejects all angle brackets
// in skill descriptions — not just XML tags. Covers bare <, >, <123>, x < y etc.
// For skills, call this instead of DetectXMLTags (it is a strict superset).
func DetectBareAngleBrackets(content, fieldName, filePath, fileContents string) *cue.ValidationError {
	if strings.ContainsAny(content, "<>") {
		return descriptionFieldError(fieldName, "contains angle bracket characters (< or >) which are not allowed by Anthropic's validator — use entities like &lt; and &gt; if needed", filePath, fileContents)
	}
	return nil
}

// CheckSizeLimit checks if content exceeds a line limit with tolerance.
// Returns a suggestion if the limit is exceeded.
func CheckSizeLimit(contents string, limit int, tolerance float64, componentType, filePath string) *cue.ValidationError {
	lines := strings.Count(contents, "\n")
	maxLines := int(float64(limit) * (1 + tolerance))

	if lines > maxLines {
		var message string
		switch componentType {
		case cue.TypeAgent:
			message = fmt.Sprintf("Agent is %d lines. Best practice: keep agents under ~%d lines (%d%s%d%%) - move methodology to skills instead.",
				lines, maxLines, limit, "±", int(tolerance*100))
		case cue.TypeCommand:
			message = fmt.Sprintf("Command is %d lines. Best practice: keep commands under ~%d lines (%d%s%d%%) - delegate to specialist agents instead of implementing logic directly.",
				lines, maxLines, limit, "±", int(tolerance*100))
		case cue.TypeSkill:
			message = fmt.Sprintf("Skill is %d lines. Best practice: keep skills under ~%d lines (%d%s%d%%) - move heavy docs to references/ subdirectory.",
				lines, maxLines, limit, "±", int(tolerance*100))
		default:
			message = fmt.Sprintf("%s is %d lines, exceeds recommended %d lines.",
				componentType, lines, maxLines)
		}

		return &cue.ValidationError{
			File:     filePath,
			Message:  message,
			Severity: "suggestion",
			Source:   cue.SourceCClintObserve,
			Line:     1,
		}
	}
	return nil
}

// sortedMapKeys returns the keys of a map[string]bool as a sorted, comma-separated string.
func sortedMapKeys(m map[string]bool) string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return strings.Join(keys, ", ")
}
