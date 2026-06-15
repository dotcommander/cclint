package cue

import (
	"bytes"
	"embed"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/dotcommander/cclint/internal/textutil"
	"github.com/dotcommander/cclint/internal/types"

	"cuelang.org/go/cue"
	"cuelang.org/go/cue/cuecontext"
	cuerrors "cuelang.org/go/cue/errors"
)

//go:embed schemas/*.cue
var schemaFS embed.FS

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

// LoadSchemas loads all CUE schema files from the embedded filesystem
func (v *Validator) LoadSchemas(schemaDir string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

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

			// Inject generated CUE unions (single source in Go) so schemas never hand-maintain these lists.
			for _, inj := range []struct {
				token string
				gen   func() string
			}{
				{"#KnownTool", knownToolUnionCUE},
				{"#Model", modelUnionCUE},
			} {
				if bytes.Contains(content, []byte(inj.token)) {
					content = append(content, []byte("\n"+inj.gen()+"\n")...)
				}
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

	// Validate concreteness - ensures required fields are present.
	// Optional fields (name?: string) are correctly skipped by CUE.
	if err := unified.Validate(cue.Concrete(true)); err != nil {
		return v.extractErrorsFromCUE(err, schemaType), nil
	}

	// Data validates successfully
	return nil, nil
}

// extractErrorsFromCUE flattens a CUE error into one ValidationError per
// underlying field issue, preserving each issue's path/position/message.
func (v *Validator) extractErrorsFromCUE(err error, schemaType string) []ValidationError {
	var validationErrors []ValidationError

	for _, cueErr := range cuerrors.Errors(err) {
		pos := cueErr.Position()
		msg := cueErr.Error()
		if path := cueErr.Path(); len(path) > 0 {
			msg = fmt.Sprintf("%s: %s", strings.Join(path, "."), msg)
		}
		validationErrors = append(validationErrors, ValidationError{
			File:     "",
			Message:  msg,
			Severity: types.SeverityError,
			Source:   SourceAnthropicDocs,
			Line:     pos.Line(),
			Column:   pos.Column(),
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
