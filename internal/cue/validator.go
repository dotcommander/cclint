package cue

import (
	"fmt"
	"strings"
	"sync"

	"github.com/dotcommander/cclint/internal/textutil"
	"github.com/dotcommander/cclint/internal/types"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cuerrors "cuelang.org/go/cue/errors"
)

// Re-export types and constants from internal/types for backward compatibility.
type ValidationError = types.ValidationError

const (
	SourceAnthropicDocs = types.SourceAnthropicDocs
	SourceCClintObserve = types.SourceCClintObserve
	SourceAgentSkillsIO = types.SourceAgentSkillsIO
	SeverityError       = types.SeverityError
	SeverityWarning     = types.SeverityWarning
	SeveritySuggestion  = types.SeveritySuggestion
	SeverityInfo        = types.SeverityInfo
	TypeAgent           = types.TypeAgent
	TypeCommand         = types.TypeCommand
	TypeSkill           = types.TypeSkill
	TypeRule            = types.TypeRule
	TypeHTTP            = types.TypeHTTP
)

// Validator handles CUE validation
type Validator struct {
	mu      sync.Mutex
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

// ValidateAgent validates agent data against the agent schema
func (v *Validator) ValidateAgent(data map[string]any) ([]ValidationError, error) {
	return v.validateSchema("agent", data)
}

// ValidateCommand validates command data against the command schema
func (v *Validator) ValidateCommand(data map[string]any) ([]ValidationError, error) {
	return v.validateSchema("command", data)
}

// ValidateSettings validates settings data against the settings schema
func (v *Validator) ValidateSettings(data map[string]any) ([]ValidationError, error) {
	return v.validateSchema("settings", data)
}

// ValidateSkill validates skill data against the skill schema
func (v *Validator) ValidateSkill(data map[string]any) ([]ValidationError, error) {
	return v.validateSchema("skill", data)
}

// ValidateClaudeMD validates CLAUDE.md data against the schema
func (v *Validator) ValidateClaudeMD(data map[string]any) ([]ValidationError, error) {
	return v.validateSchema("claude_md", data)
}

func (v *Validator) validateSchema(schemaType string, data map[string]any) ([]ValidationError, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	schema, ok := v.schemas[schemaType]
	if !ok {
		return nil, nil
	}
	return v.validateAgainstSchemaLocked(schema, data, schemaType)
}

// validateAgainstSchema validates data against a CUE schema
func (v *Validator) validateAgainstSchema(schema cue.Value, data map[string]any, schemaType string) ([]ValidationError, error) {
	v.mu.Lock()
	defer v.mu.Unlock()

	return v.validateAgainstSchemaLocked(schema, data, schemaType)
}

func (v *Validator) validateAgainstSchemaLocked(schema cue.Value, data map[string]any, schemaType string) ([]ValidationError, error) {
	// Create a CUE value from the data. JSON-decoded numbers arrive as
	// float64, which CUE treats as floats: the integer literal 3 in a settings
	// file would fail every `int` schema constraint (e.g. hook `timeout?:
	// *30 | int`). Normalize integral floats to int64 first so integer-valued
	// JSON numbers unify with integer constraints.
	dataValue := v.ctx.Encode(normalizeJSONNumbers(data))
	if encErr := dataValue.Err(); encErr != nil {
		return nil, fmt.Errorf("error encoding data: %w", encErr)
	}

	// Use Unify to check if data conforms to schema
	// Definition names are explicit because CLAUDE.md does not follow title casing.
	definitions := map[string]string{
		"agent": "#Agent", "command": "#Command", "skill": "#Skill",
		"settings": "#Settings", "claude_md": "#ClaudeMD",
	}
	definition, ok := definitions[schemaType]
	if !ok {
		return nil, fmt.Errorf("unknown CUE schema type: %s", schemaType)
	}
	defPath := cue.ParsePath(definition)

	// Try to get the definition from schema
	def := schema.LookupPath(defPath)
	if !def.Exists() {
		return nil, fmt.Errorf("CUE schema %s is missing expected definition %s", schemaType, definition)
	}

	// Check if data unifies with schema (unify checks if both can be true simultaneously)
	unified := def.Unify(dataValue)
	if err := unified.Err(); err != nil {
		return v.extractErrorsFromCUE(err, schemaType), nil
	}

	// Validate concreteness - ensures required fields are present.
	// Optional fields (name?: string) are correctly skipped by CUE.
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return v.extractErrorsFromCUE(err, schemaType), nil
	}

	// Data validates successfully
	return nil, nil
}

// normalizeJSONNumbers returns a copy of v with every integral float64
// replaced by int64. encoding/json decodes all JSON numbers to float64;
// without this, `"timeout": 3` cannot satisfy CUE `int` constraints. The
// input is never mutated: callers may reuse the decoded map for Go-side
// checks that expect encoding/json's float64 representation.
func normalizeJSONNumbers(v any) any {
	switch x := v.(type) {
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[k] = normalizeJSONNumbers(val)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = normalizeJSONNumbers(val)
		}
		return out
	case float64:
		if i := int64(x); float64(i) == x {
			return i
		}
		return x
	default:
		return v
	}
}

// extractErrorsFromCUE flattens a CUE error into one ValidationError per
// underlying field issue. Encoded maps have no document positions.
func (v *Validator) extractErrorsFromCUE(err error, schemaType string) []ValidationError {
	var validationErrors []ValidationError

	for _, cueErr := range cuerrors.Errors(err) {
		msg := cueErr.Error()
		if path := cueErr.Path(); len(path) > 0 {
			prefix := strings.Join(path, ".") + ": "
			if !strings.HasPrefix(msg, prefix) {
				msg = prefix + msg
			}
		}
		validationErrors = append(validationErrors, ValidationError{
			File:     "",
			Message:  msg,
			Severity: types.SeverityError,
			Source:   SourceAnthropicDocs,
			Line:     0,
			Column:   0,
		})
	}

	// cuerrors.Errors can return nil for some wrapped errors; never drop
	// the diagnostic entirely.
	if len(validationErrors) == 0 {
		validationErrors = append(validationErrors, ValidationError{
			File:     "",
			Message:  err.Error(),
			Severity: types.SeverityError,
			Source:   SourceAnthropicDocs,
			Line:     0,
			Column:   0,
		})
	}

	return validationErrors
}

// Frontmatter represents parsed frontmatter
type Frontmatter struct {
	Data map[string]any
	Body string
}

// ParseFrontmatter parses YAML frontmatter from markdown content.
// Delegates to textutil.ParseYAMLFrontmatter to avoid duplicate implementations.
func ParseFrontmatter(content string) (*Frontmatter, error) {
	fm, err := textutil.ParseYAMLFrontmatter(content)
	if err != nil {
		return nil, fmt.Errorf("error parsing frontmatter: %w", err)
	}
	return &Frontmatter{
		Data: fm.Data,
		Body: fm.Body,
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
			Severity: types.SeverityError,
		}}, nil
	}

	// Map-only validation APIs cannot know the file; attribute at this boundary.
	var diagnostics []ValidationError
	switch fileType {
	case "agent":
		diagnostics, err = v.ValidateAgent(fm.Data)
	case "command":
		diagnostics, err = v.ValidateCommand(fm.Data)
	case "skill":
		diagnostics, err = v.ValidateSkill(fm.Data)
	case "settings":
		diagnostics, err = v.ValidateSettings(fm.Data)
	case "claude_md":
		diagnostics, err = v.ValidateClaudeMD(fm.Data)
	default:
		return nil, fmt.Errorf("unknown file type: %s", fileType)
	}
	for i := range diagnostics {
		diagnostics[i].File = path
	}
	return diagnostics, err
}
