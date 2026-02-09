package format

import (
	"bytes"
	"fmt"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// Formatter formats component files canonically.
type Formatter interface {
	// Format takes raw file content and returns formatted content.
	// Returns the formatted content and nil error if successful.
	// Returns original content and error if formatting fails.
	Format(content string) (string, error)
}

// ComponentFormatter provides base formatting for all component types.
type ComponentFormatter struct{}

// NewComponentFormatter creates a formatter for a specific component type.
func NewComponentFormatter(componentType string) Formatter {
	switch componentType {
	case "agent":
		return &AgentFormatter{}
	case "command":
		return &CommandFormatter{}
	case "skill":
		return &SkillFormatter{}
	default:
		return &GenericFormatter{}
	}
}

// parseResult holds the result of parsing frontmatter from content.
type parseResult struct {
	frontmatter    string
	body           string
	hasFrontmatter bool
	err            error
}

// parseFrontmatterRaw extracts frontmatter and body without fully parsing YAML.
func parseFrontmatterRaw(content string) parseResult {
	trimmed := strings.TrimLeft(content, " \t")
	if !strings.HasPrefix(trimmed, "---") {
		return parseResult{body: content}
	}

	// Find the closing ---
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		return parseResult{body: content, err: fmt.Errorf("unclosed frontmatter (missing closing ---)")}
	}

	return parseResult{frontmatter: parts[1], body: parts[2], hasFrontmatter: true}
}

// normalizeFrontmatter reorders and normalizes YAML frontmatter fields.
// Priority fields come first, then others alphabetically.
func normalizeFrontmatter(yamlContent string, priorityFields []string) (string, error) {
	// Extract key-value pairs
	data := make(map[string]any)
	if err := yaml.Unmarshal([]byte(yamlContent), &data); err != nil {
		return "", err
	}

	// Build ordered list of keys
	var orderedKeys []string

	// Add priority fields first (if present)
	for _, key := range priorityFields {
		if _, exists := data[key]; exists {
			orderedKeys = append(orderedKeys, key)
		}
	}

	// Add remaining fields alphabetically
	var otherKeys []string
	for key := range data {
		if !slices.Contains(priorityFields, key) {
			otherKeys = append(otherKeys, key)
		}
	}
	sort.Strings(otherKeys)
	orderedKeys = append(orderedKeys, otherKeys...)

	// Manually serialize each field in order to preserve ordering
	var buf bytes.Buffer
	for _, key := range orderedKeys {
		value := data[key]

		// Serialize the key-value pair
		var fieldBuf bytes.Buffer
		fieldEncoder := yaml.NewEncoder(&fieldBuf)
		fieldEncoder.SetIndent(2)

		// Create single-entry map for this field
		singleField := map[string]any{key: value}
		if err := fieldEncoder.Encode(singleField); err != nil {
			return "", err
		}

		fieldStr := fieldBuf.String()
		// Remove trailing newline
		fieldStr = strings.TrimSuffix(fieldStr, "\n")

		buf.WriteString(fieldStr)
		buf.WriteString("\n")
	}

	result := buf.String()
	// Remove final trailing newline
	return strings.TrimSuffix(result, "\n"), nil
}

// normalizeMarkdown normalizes markdown body content.
// hasFrontmatter indicates if this body follows frontmatter.
func normalizeMarkdown(body string, hasFrontmatter bool) string {
	// Trim trailing whitespace from each line
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		lines[i] = strings.TrimRight(line, " \t")
	}

	// Join lines
	result := strings.Join(lines, "\n")

	if hasFrontmatter {
		// Ensure exactly one blank line after frontmatter
		result = strings.TrimLeft(result, "\n")
		result = "\n" + result
	}

	// Ensure file ends with exactly one newline
	result = strings.TrimRight(result, "\n") + "\n"

	return result
}

// AgentFormatter formats agent files.
type AgentFormatter struct{}

func (f *AgentFormatter) Format(content string) (string, error) {
	result := parseFrontmatterRaw(content)
	if result.err != nil {
		return content, result.err
	}

	if !result.hasFrontmatter {
		// No frontmatter - just normalize markdown
		return normalizeMarkdown(result.body, false), nil
	}

	// Normalize frontmatter with agent-specific field order
	priorityFields := []string{"name", "description", "model", "tools", "allowed-tools"}
	normalizedFM, err := normalizeFrontmatter(result.frontmatter, priorityFields)
	if err != nil {
		return content, err
	}

	// Normalize body
	normalizedBody := normalizeMarkdown(result.body, true)

	// Reassemble
	return "---\n" + normalizedFM + "\n---" + normalizedBody, nil
}

// CommandFormatter formats command files.
type CommandFormatter struct{}

func (f *CommandFormatter) Format(content string) (string, error) {
	result := parseFrontmatterRaw(content)
	if result.err != nil {
		return content, result.err
	}

	if !result.hasFrontmatter {
		return normalizeMarkdown(result.body, false), nil
	}

	// Normalize frontmatter with command-specific field order
	priorityFields := []string{"name", "description", "allowed-tools"}
	normalizedFM, err := normalizeFrontmatter(result.frontmatter, priorityFields)
	if err != nil {
		return content, err
	}

	normalizedBody := normalizeMarkdown(result.body, true)
	return "---\n" + normalizedFM + "\n---" + normalizedBody, nil
}

// SkillFormatter formats skill files.
type SkillFormatter struct{}

func (f *SkillFormatter) Format(content string) (string, error) {
	result := parseFrontmatterRaw(content)
	if result.err != nil {
		return content, result.err
	}

	if !result.hasFrontmatter {
		return normalizeMarkdown(result.body, false), nil
	}

	// Normalize frontmatter with skill-specific field order
	priorityFields := []string{"name", "description"}
	normalizedFM, err := normalizeFrontmatter(result.frontmatter, priorityFields)
	if err != nil {
		return content, err
	}

	normalizedBody := normalizeMarkdown(result.body, true)
	return "---\n" + normalizedFM + "\n---" + normalizedBody, nil
}

// GenericFormatter formats generic markdown files.
type GenericFormatter struct{}

func (f *GenericFormatter) Format(content string) (string, error) {
	result := parseFrontmatterRaw(content)
	if result.err != nil {
		return content, result.err
	}

	if !result.hasFrontmatter {
		return normalizeMarkdown(result.body, false), nil
	}

	// Generic alphabetical order
	priorityFields := []string{"name", "description"}
	normalizedFM, err := normalizeFrontmatter(result.frontmatter, priorityFields)
	if err != nil {
		return content, err
	}

	normalizedBody := normalizeMarkdown(result.body, true)
	return "---\n" + normalizedFM + "\n---" + normalizedBody, nil
}

// Diff computes a simple unified diff between original and formatted content.
// Returns empty string if contents are identical.
func Diff(original, formatted, filename string) string {
	if original == formatted {
		return ""
	}

	var buf bytes.Buffer
	fmt.Fprintf(&buf, "--- %s\n", filename)
	fmt.Fprintf(&buf, "+++ %s (formatted)\n", filename)

	origLines := strings.Split(original, "\n")
	fmtLines := strings.Split(formatted, "\n")

	// Simple line-by-line diff
	maxLen := max(len(origLines), len(fmtLines))

	for i := 0; i < maxLen; i++ {
		var origLine, fmtLine string
		if i < len(origLines) {
			origLine = origLines[i]
		}
		if i < len(fmtLines) {
			fmtLine = fmtLines[i]
		}

		if origLine != fmtLine {
			if origLine != "" {
				fmt.Fprintf(&buf, "- %s\n", origLine)
			}
			if fmtLine != "" {
				fmt.Fprintf(&buf, "+ %s\n", fmtLine)
			}
		}
	}

	return buf.String()
}
