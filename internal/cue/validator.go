package cue

import (
	"embed"
	"fmt"
	"path/filepath"
	"strings"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	yamlv3 "gopkg.in/yaml.v3"
)

//go:embed schemas/*.cue
var schemaFS embed.FS

// Rule source constants
const (
	SourceAnthropicDocs = "anthropic-docs"     // Official Anthropic documentation
	SourceCClintObserve = "cclint-observation" // Our best practice observations
	SourceAgentSkillsIO = "agentskills-io"     // agentskills.io specification
)

// ValidationError represents a validation error
type ValidationError struct {
	File     string
	Message  string
	Severity string // error, warning, suggestion
	Source   string // anthropic-docs or cclint-observation
	Line     int
	Column   int
}

// Validator handles CUE validation
type Validator struct {
	ctx     *cue.Context
	schemas map[string]cue.Value
}

// NewValidator creates a new Validator instance
func NewValidator() *Validator {
	return &Validator{
		ctx:     cuecontext.New(),
		schemas: make(map[string]cue.Value),
	}
}

// LoadSchemas loads all CUE schema files from the embedded filesystem
func (v *Validator) LoadSchemas(schemaDir string) error {
	// List all files in the embedded schemas directory
	entries, err := schemaFS.ReadDir("schemas")
	if err != nil {
		return fmt.Errorf("warning: could not read embedded schemas: %w", err)
	}

	// Load each .cue file
	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".cue" {
			content, err := schemaFS.ReadFile(filepath.Join("schemas", entry.Name()))
			if err != nil {
				continue
			}

			// Compile the CUE schema
			inst := v.ctx.CompileBytes(content, cue.Filename(entry.Name()))
			if instErr := inst.Err(); instErr != nil {
				// Log but don't fail - schema files might have issues
				continue
			}

			// Store the compiled schema
			// Extract base name (agent.cue -> agent)
			schemaName := entry.Name()[:len(entry.Name())-4]
			v.schemas[schemaName] = inst.Value()
		}
	}

	if len(v.schemas) == 0 {
		return fmt.Errorf("warning: no CUE schemas loaded, using Go validation")
	}

	return nil
}

// ValidateAgent validates agent data against the agent schema
func (v *Validator) ValidateAgent(data map[string]any) ([]ValidationError, error) {
	schema, ok := v.schemas["agent"]
	if !ok {
		// Fallback to Go validation if CUE schema not loaded
		return nil, nil
	}
	return v.validateAgainstSchema(schema, data, "agent")
}

// ValidateCommand validates command data against the command schema
func (v *Validator) ValidateCommand(data map[string]any) ([]ValidationError, error) {
	schema, ok := v.schemas["command"]
	if !ok {
		return nil, nil
	}
	return v.validateAgainstSchema(schema, data, "command")
}

// ValidateSettings validates settings data against the settings schema
func (v *Validator) ValidateSettings(data map[string]any) ([]ValidationError, error) {
	schema, ok := v.schemas["settings"]
	if !ok {
		return nil, nil
	}
	return v.validateAgainstSchema(schema, data, "settings")
}

// ValidateSkill validates skill data against the skill schema
func (v *Validator) ValidateSkill(data map[string]any) ([]ValidationError, error) {
	schema, ok := v.schemas["skill"]
	if !ok {
		// Fallback to Go validation if CUE schema not loaded
		return nil, nil
	}
	return v.validateAgainstSchema(schema, data, "skill")
}

// ValidateClaudeMD validates CLAUDE.md data against the schema
func (v *Validator) ValidateClaudeMD(data map[string]any) ([]ValidationError, error) {
	schema, ok := v.schemas["claude_md"]
	if !ok {
		return nil, nil
	}
	return v.validateAgainstSchema(schema, data, "claude_md")
}

// validateAgainstSchema validates data against a CUE schema
func (v *Validator) validateAgainstSchema(schema cue.Value, data map[string]any, schemaType string) ([]ValidationError, error) {
	// Create a CUE value from the data
	dataValue := v.ctx.Encode(data)
	if encErr := dataValue.Err(); encErr != nil {
		return nil, fmt.Errorf("error encoding data: %w", encErr)
	}

	// Use Unify to check if data conforms to schema
	// We extract the #Agent, #Command, etc. definition from the schema
	defPath := cue.ParsePath(fmt.Sprintf("#%s", strings.ToUpper(schemaType[:1])+schemaType[1:]))

	// Try to get the definition from schema
	def := schema.LookupPath(defPath)
	if !def.Exists() {
		// Schema definition not found, this is OK - just return no errors
		return nil, nil
	}

	// Check if data unifies with schema (unify checks if both can be true simultaneously)
	unified := def.Unify(dataValue)
	if err := unified.Err(); err != nil {
		return v.extractErrorsFromCUE(err, schemaType), nil
	}

	// Validate concreteness - ensures required fields are present
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return v.extractErrorsFromCUE(err, schemaType), nil
	}

	// Data validates successfully
	return nil, nil
}

// extractErrorsFromCUE extracts user-friendly validation errors from CUE errors
func (v *Validator) extractErrorsFromCUE(err error, schemaType string) []ValidationError {
	var errors []ValidationError

	// CUE errors contain detailed information about what failed
	// For now, provide a simpler error message
	errors = append(errors, ValidationError{
		File:     "",
		Message:  fmt.Sprintf("Schema validation failed: %v", err),
		Severity: "error",
		Source:   SourceAnthropicDocs,
		Line:     0,
		Column:   0,
	})

	return errors
}

// Frontmatter represents parsed frontmatter
type Frontmatter struct {
	Data map[string]any
	Body string
}

// ParseFrontmatter parses YAML frontmatter from markdown content
func ParseFrontmatter(content string) (*Frontmatter, error) {
	// Split content by ---
	parts := strings.SplitN(content, "---", 3)
	if len(parts) < 3 {
		// No frontmatter found
		return &Frontmatter{
			Data: make(map[string]any),
			Body: content,
		}, nil
	}

	// Parse YAML frontmatter
	var data map[string]any
	if err := yamlv3.Unmarshal([]byte(parts[1]), &data); err != nil {
		return nil, fmt.Errorf("error parsing frontmatter: %w", err)
	}

	return &Frontmatter{
		Data: data,
		Body: parts[2],
	}, nil
}

// ValidateFile validates a file based on its type
func (v *Validator) ValidateFile(path string, content string, fileType string) ([]ValidationError, error) {
	// Parse frontmatter
	fm, err := ParseFrontmatter(content)
	if err != nil {
		return []ValidationError{{
			File:     path,
			Message:  err.Error(),
			Severity: "error",
		}}, nil
	}

	// Validate based on file type
	switch fileType {
	case "agent":
		return v.ValidateAgent(fm.Data)
	case "command":
		return v.ValidateCommand(fm.Data)
	case "skill":
		return v.ValidateSkill(fm.Data)
	case "settings":
		return v.ValidateSettings(fm.Data)
	case "claude_md":
		return v.ValidateClaudeMD(fm.Data)
	default:
		return nil, fmt.Errorf("unknown file type: %s", fileType)
	}
}
