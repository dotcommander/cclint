package format

import (
	"bytes"
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/dotcommander/cclint/internal/discovery"
	"gopkg.in/yaml.v3"
)

// Formatter formats component files canonically.
type Formatter interface {
	// Format takes raw file content and returns formatted content.
	// Returns the formatted content and nil error if successful.
	// Returns original content and error if formatting fails.
	Format(content string) (string, error)
}

// componentFormatterFactories is the formatter-owned capability registry.
// Generic Markdown component types retain the historical SkillFormatter.
// Output styles are intentionally absent: positional fmt arguments have never
// selected them.
var componentFormatterFactories = map[discovery.FileType]func() Formatter{ //nolint:gochecknoglobals // Immutable formatter capability registry.
	discovery.FileTypeAgent:    func() Formatter { return &AgentFormatter{} },
	discovery.FileTypeCommand:  func() Formatter { return &CommandFormatter{} },
	discovery.FileTypeSkill:    func() Formatter { return &SkillFormatter{} },
	discovery.FileTypeSettings: func() Formatter { return &SkillFormatter{} },
	discovery.FileTypeContext:  func() Formatter { return &SkillFormatter{} },
	discovery.FileTypePlugin:   func() Formatter { return &SkillFormatter{} },
	discovery.FileTypeRule:     func() Formatter { return &SkillFormatter{} },
}

// CanFormatComponent reports whether componentType names a component selected
// by positional fmt arguments. Singular and plural discovery aliases are valid.
func CanFormatComponent(componentType string) bool {
	_, ok := formatterFor(componentType)
	return ok
}

// formatterFor resolves the formatter factory registered for a component type.
func formatterFor(componentType string) (func() Formatter, bool) {
	fileType, err := discovery.ParseFileType(componentType)
	if err != nil {
		return nil, false
	}
	factory, ok := componentFormatterFactories[fileType]
	return factory, ok
}

// NewComponentFormatter creates a formatter for a specific component type.
func NewComponentFormatter(componentType string) Formatter {
	if factory, ok := formatterFor(componentType); ok {
		return factory()
	}
	return &SkillFormatter{}
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
	lines := strings.SplitAfter(content, "\n")
	if strings.TrimRight(lines[0], " \t\r\n") != "---" {
		return parseResult{body: content}
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], " \t\r\n") == "---" {
			return parseResult{
				frontmatter: strings.Join(lines[1:i], ""),
				body:        strings.Join(lines[i+1:], ""), hasFrontmatter: true,
			}
		}
	}
	return parseResult{body: content, err: fmt.Errorf("unclosed frontmatter (missing closing ---)")}
}

// normalizeFrontmatter reorders existing YAML nodes without discarding comments,
// aliases, tags, or scalar styles. Anchor dependencies take precedence over field
// order so that moving a field cannot leave an alias before its anchor.
func normalizeFrontmatter(yamlContent string, priorityFields []string) (string, error) {
	var document yaml.Node
	decoder := yaml.NewDecoder(strings.NewReader(yamlContent))
	if err := decoder.Decode(&document); err != nil && err != io.EOF {
		return "", err
	}
	var extra yaml.Node
	if err := decoder.Decode(&extra); err != io.EOF {
		if err != nil {
			return "", err
		}
		return "", fmt.Errorf("frontmatter must contain one YAML document")
	}
	if len(document.Content) == 0 {
		return strings.Trim(yamlContent, "\r\n"), nil
	}
	mapping := document.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return "", fmt.Errorf("frontmatter must be a YAML mapping")
	}
	// Node decoding alone does not reject duplicate mapping keys.
	var values map[string]any
	if err := mapping.Decode(&values); err != nil {
		return "", err
	}
	type field struct {
		key, value *yaml.Node
		rank       int
	}
	fields := make([]field, 0, len(mapping.Content)/2)
	owners := make(map[*yaml.Node]int)
	var visit func(*yaml.Node, func(*yaml.Node))
	visit = func(node *yaml.Node, action func(*yaml.Node)) {
		action(node)
		for _, child := range node.Content {
			visit(child, action)
		}
	}
	for i := 0; i < len(mapping.Content); i += 2 {
		key, value := mapping.Content[i], mapping.Content[i+1]
		rank := len(priorityFields)
		for index, name := range priorityFields {
			if key.Value == name {
				rank = index
				break
			}
		}
		fields = append(fields, field{key, value, rank})
		owner := i / 2
		if key.Anchor != "" {
			owners[key] = owner
		}
		visit(value, func(node *yaml.Node) {
			if node.Anchor != "" {
				owners[node] = owner
			}
			// Expand collection syntax as before, retaining scalar quoting/style.
			if node.Kind == yaml.MappingNode || node.Kind == yaml.SequenceNode {
				node.Style &^= yaml.FlowStyle
			}
		})
	}
	order := make([]int, len(fields))
	for i := range order {
		order[i] = i
	}
	sort.SliceStable(order, func(i, j int) bool {
		a, b := fields[order[i]], fields[order[j]]
		if a.rank != b.rank {
			return a.rank < b.rank
		}
		return a.key.Value < b.key.Value
	})
	dependencies := make([][]int, len(fields))
	for i, f := range fields {
		collectDependency := func(node *yaml.Node) {
			if node.Kind == yaml.AliasNode {
				if owner, ok := owners[node.Alias]; ok && owner != i {
					dependencies[i] = append(dependencies[i], owner)
				}
			}
		}
		visit(f.key, collectDependency)
		visit(f.value, collectDependency)
	}
	var ordered []*yaml.Node
	state := make([]uint8, len(fields))
	var appendField func(int) error
	appendField = func(i int) error {
		if state[i] == 2 {
			return nil
		}
		if state[i] == 1 {
			return fmt.Errorf("cyclic YAML anchor dependencies")
		}
		state[i] = 1
		for _, dependency := range dependencies[i] {
			if err := appendField(dependency); err != nil {
				return err
			}
		}
		ordered = append(ordered, fields[i].key, fields[i].value)
		state[i] = 2
		return nil
	}
	for _, i := range order {
		if err := appendField(i); err != nil {
			return "", err
		}
	}
	mapping.Content = ordered
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&document); err != nil {
		return "", err
	}
	if err := encoder.Close(); err != nil {
		return "", err
	}
	return strings.TrimSuffix(buf.String(), "\n"), nil
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

// formatComponent formats a component file with the given priority field ordering.
func formatComponent(content string, priorityFields []string) (string, error) {
	result := parseFrontmatterRaw(content)
	if result.err != nil {
		return content, result.err
	}

	if !result.hasFrontmatter {
		return normalizeMarkdown(result.body, false), nil
	}

	normalizedFM, err := normalizeFrontmatter(result.frontmatter, priorityFields)
	if err != nil {
		return content, err
	}

	normalizedBody := normalizeMarkdown(result.body, true)
	return "---\n" + normalizedFM + "\n---" + normalizedBody, nil
}

// AgentFormatter formats agent files.
type AgentFormatter struct{}

func (f *AgentFormatter) Format(content string) (string, error) {
	return formatComponent(content, []string{"name", "description", "model", "tools", "allowed-tools"})
}

// CommandFormatter formats command files.
type CommandFormatter struct{}

func (f *CommandFormatter) Format(content string) (string, error) {
	return formatComponent(content, []string{"name", "description", "allowed-tools"})
}

// SkillFormatter formats skill files.
type SkillFormatter struct{}

func (f *SkillFormatter) Format(content string) (string, error) {
	return formatComponent(content, []string{"name", "description"})
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
