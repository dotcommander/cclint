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

// ParseYAMLFrontmatter extracts YAML frontmatter from markdown content
func ParseYAMLFrontmatter(content string) (*Frontmatter, error) {
	// Split content by ---
	parts := strings.SplitN(content, "---", 3)

	// If we have less than 3 parts, there's no frontmatter
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