package textutil

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents parsed frontmatter data
type Frontmatter struct {
	Data map[string]any
	Body string
}

// RawFrontmatter retains YAML bytes and the suffix after the closing three dashes.
type RawFrontmatter struct {
	YAML           string
	Body           string
	HasFrontmatter bool
}

// ExtractRawFrontmatter recognizes only complete, unindented delimiter lines.
// Body includes closing-line whitespace and its line ending for parser compatibility.
func ExtractRawFrontmatter(content string) (RawFrontmatter, error) {
	lines := strings.SplitAfter(content, "\n")
	if !isFrontmatterDelimiter(lines[0]) {
		return RawFrontmatter{Body: content}, nil
	}
	for i := 1; i < len(lines); i++ {
		if isFrontmatterDelimiter(lines[i]) {
			return RawFrontmatter{
				YAML:           strings.Join(lines[1:i], ""),
				Body:           strings.TrimPrefix(strings.Join(lines[i:], ""), "---"),
				HasFrontmatter: true,
			}, nil
		}
	}
	return RawFrontmatter{Body: content}, fmt.Errorf("unclosed frontmatter (missing closing ---)")
}

func isFrontmatterDelimiter(line string) bool {
	if withoutLF, ok := strings.CutSuffix(line, "\n"); ok {
		line = strings.TrimSuffix(withoutLF, "\r")
	}
	return strings.TrimRight(line, " \t") == "---"
}

func ParseYAMLFrontmatter(content string) (*Frontmatter, error) {
	raw, err := ExtractRawFrontmatter(content)
	if err != nil {
		return nil, err
	}
	if !raw.HasFrontmatter {
		return &Frontmatter{Data: make(map[string]any), Body: raw.Body}, nil
	}
	var data map[string]any
	if err := yaml.Unmarshal([]byte(raw.YAML), &data); err != nil {
		return nil, err
	}
	return &Frontmatter{Data: data, Body: raw.Body}, nil
}
