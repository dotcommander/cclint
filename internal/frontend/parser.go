package frontend

import (
	"strings"

	"gopkg.in/yaml.v3"
)

// Frontmatter represents parsed frontmatter data
type Frontmatter struct {
	Data  map[string]interface{}
	Body  string
}

// ParseYAMLFrontmatter extracts YAML frontmatter from markdown content.
// Frontmatter must start at the beginning of the file with "---".
func ParseYAMLFrontmatter(content string) (*Frontmatter, error) {
	// Frontmatter must start at the very beginning of the file
	trimmed := strings.TrimLeft(content, " \t")
	if !strings.HasPrefix(trimmed, "---") {
		// No frontmatter - return content as body
		return &Frontmatter{
			Data: make(map[string]interface{}),
			Body: content,
		}, nil
	}

	// Split content by ---
	parts := strings.SplitN(content, "---", 3)

	// If we have less than 3 parts, there's no closing ---
	if len(parts) < 3 {
		return &Frontmatter{
			Data: make(map[string]interface{}),
			Body: content,
		}, nil
	}

	// The frontmatter is the part between the first pair of ---
	frontmatterYAML := parts[1]
	body := parts[2]

	// Parse YAML content
	var data map[string]interface{}
	if err := yaml.Unmarshal([]byte(frontmatterYAML), &data); err != nil {
		return nil, err
	}

	return &Frontmatter{
		Data: data,
		Body: body,
	}, nil
}