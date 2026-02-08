package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/frontend"
	"github.com/dotcommander/cclint/internal/scoring"
)

// SkillLinter implements ComponentLinter for skill files.
// It also implements optional interfaces for pre-validation, best practices,
// cross-file validation, scoring, improvements, and batch post-processing.
type SkillLinter struct {
	BaseLinter
}

// Compile-time interface compliance checks
var (
	_ ComponentLinter       = (*SkillLinter)(nil)
	_ PreValidator          = (*SkillLinter)(nil)
	_ BestPracticeValidator = (*SkillLinter)(nil)
	_ CrossFileValidatable  = (*SkillLinter)(nil)
	_ Scorable              = (*SkillLinter)(nil)
	_ Improvable            = (*SkillLinter)(nil)
	_ BatchPostProcessor    = (*SkillLinter)(nil)
)

// NewSkillLinter creates a new SkillLinter.
func NewSkillLinter() *SkillLinter {
	return &SkillLinter{}
}

func (l *SkillLinter) Type() string {
	return "skill"
}

func (l *SkillLinter) FileType() discovery.FileType {
	return discovery.FileTypeSkill
}

// PreValidate implements PreValidator interface
func (l *SkillLinter) PreValidate(filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check filename is SKILL.md
	if !strings.HasSuffix(filePath, "/SKILL.md") && !strings.HasSuffix(filePath, "\\SKILL.md") && filepath.Base(filePath) != "SKILL.md" {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Skill file must be named SKILL.md",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	// Check not empty
	if len(strings.TrimSpace(contents)) == 0 {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Skill file is empty",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	return errors
}

func (l *SkillLinter) ParseContent(contents string) (map[string]any, string, error) {
	// Skills have optional frontmatter
	fm, err := frontend.ParseYAMLFrontmatter(contents)
	if err != nil {
		// No frontmatter is OK for skills
		return make(map[string]any), contents, nil
	}
	return fm.Data, fm.Body, nil
}

func (l *SkillLinter) ValidateCUE(validator *cue.Validator, data map[string]any) ([]cue.ValidationError, error) {
	// Skills don't use CUE validation currently
	return nil, nil
}

func (l *SkillLinter) ValidateSpecific(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	// Check for unknown frontmatter fields - helps catch fabricated/deprecated fields
	for key := range data {
		if !knownSkillFields[key] {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Unknown frontmatter field '%s'. See https://agentskills.io/specification for valid fields", key),
				Severity: "suggestion",
				Source:   cue.SourceCClintObserve,
				Line:     FindFrontmatterFieldLine(contents, key),
			})
		}
	}

	// Name validation (reserved words, format, directory match)
	if name, ok := data["name"].(string); ok {
		errors = append(errors, validateSkillName(name, filePath, contents)...)
	}

	// Validate context field: only valid value is "fork"
	if ctxVal, ok := data["context"]; ok {
		ctxStr, isStr := ctxVal.(string)
		if !isStr || ctxStr != "fork" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("context field must be 'fork' (got '%v')", ctxVal),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "context"),
			})
		}
	}

	// Validate agent field: non-empty string; warn if context is not "fork"
	if agentVal, ok := data["agent"]; ok {
		errors = append(errors, validateSkillAgentField(agentVal, data, filePath, contents)...)
	}

	// Validate user-invocable field: must be boolean
	if uiVal, ok := data["user-invocable"]; ok {
		if _, isBool := uiVal.(bool); !isBool {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("user-invocable field must be a boolean (got '%v')", uiVal),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "user-invocable"),
			})
		}
	}

	// Validate disable-model-invocation field: must be boolean
	if dmiVal, ok := data["disable-model-invocation"]; ok {
		if _, isBool := dmiVal.(bool); !isBool {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("disable-model-invocation field must be a boolean (got '%v')", dmiVal),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "disable-model-invocation"),
			})
		}
	}

	// Validate argument-hint field: must be a non-empty string if present
	if ahVal, ok := data["argument-hint"]; ok {
		ahStr, isStr := ahVal.(string)
		if !isStr {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("argument-hint field must be a string (got '%v')", ahVal),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "argument-hint"),
			})
		} else if strings.TrimSpace(ahStr) == "" {
			errors = append(errors, cue.ValidationError{
				File:     filePath,
				Message:  "argument-hint field is empty - provide a hint for autocomplete (e.g., 'PR number or URL')",
				Severity: "warning",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "argument-hint"),
			})
		}
	}

	// Validate hooks (scoped to component events: PreToolUse, PostToolUse, Stop)
	if hooks, ok := data["hooks"]; ok {
		errors = append(errors, ValidateComponentHooks(hooks, filePath)...)
	}

	// Frontmatter suggestion
	if !strings.HasPrefix(contents, "---") {
		skillName := extractSkillName(contents, filePath)
		suggestion := "Add YAML frontmatter with name and description (description is critical for skill discovery)"
		if skillName != "" {
			suggestion = fmt.Sprintf("Add frontmatter: ---\nname: %s\ndescription: Brief summary of what this skill does\n--- (description critical for discovery)", skillName)
		}
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  suggestion,
			Severity: "suggestion",
			Source:   cue.SourceAnthropicDocs,
		})
	}

	return errors
}

// ValidateBestPractices implements BestPracticeValidator interface
func (l *SkillLinter) ValidateBestPractices(filePath, contents string, data map[string]any) []cue.ValidationError {
	return validateSkillBestPractices(filePath, contents, data)
}

// ValidateCrossFile implements CrossFileValidatable interface
func (l *SkillLinter) ValidateCrossFile(crossValidator *CrossFileValidator, filePath, contents string, data map[string]any) []cue.ValidationError {
	if crossValidator == nil {
		return nil
	}
	return crossValidator.ValidateSkill(filePath, contents, data)
}

// Score implements Scorable interface
func (l *SkillLinter) Score(contents string, data map[string]any, body string) *scoring.QualityScore {
	scorer := scoring.NewSkillScorer()
	score := scorer.Score(contents, data, body)
	return &score
}

// GetImprovements implements Improvable interface
func (l *SkillLinter) GetImprovements(contents string, data map[string]any) []ImprovementRecommendation {
	return GetSkillImprovements(contents, data)
}

// PostProcessBatch implements BatchPostProcessor for orphan detection.
func (l *SkillLinter) PostProcessBatch(ctx *LinterContext, summary *LintSummary) {
	orphanedSkills := ctx.CrossValidator.FindOrphanedSkills()
	for _, orphan := range orphanedSkills {
		summary.TotalSuggestions++
		// Add to individual file results for display
		for i, result := range summary.Results {
			if result.File == orphan.File {
				summary.Results[i].Suggestions = append(summary.Results[i].Suggestions, orphan)
				break
			}
		}
	}
}

// validateSkillName checks reserved words, hyphen placement, consecutive hyphens, and directory match.
func validateSkillName(name, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

	reservedWords := map[string]bool{"anthropic": true, "claude": true}
	if reservedWords[strings.ToLower(name)] {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Name '%s' is a reserved word and cannot be used", name),
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Rule 048: Name cannot start or end with hyphen (agentskills.io spec)
	if strings.HasPrefix(name, "-") || strings.HasSuffix(name, "-") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Skill name '%s' cannot start or end with a hyphen", name),
			Severity: "error",
			Source:   cue.SourceAgentSkillsIO,
			Line:     FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Rule 049: Name cannot contain consecutive hyphens (agentskills.io spec)
	if strings.Contains(name, "--") {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Skill name '%s' contains consecutive hyphens (--) which are not allowed", name),
			Severity: "error",
			Source:   cue.SourceAgentSkillsIO,
			Line:     FindFrontmatterFieldLine(contents, "name"),
		})
	}

	// Rule 050: Name must match parent directory name (agentskills.io spec)
	parentDir := filepath.Base(filepath.Dir(filePath))
	isSpecialDir := parentDir == "." || parentDir == "skills" || parentDir == ".claude"
	if !isSpecialDir && name != parentDir {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Skill name '%s' should match parent directory name '%s'", name, parentDir),
			Severity: "warning",
			Source:   cue.SourceAgentSkillsIO,
			Line:     FindFrontmatterFieldLine(contents, "name"),
		})
	}

	return errors
}

// validateSkillAgentField validates the agent frontmatter field and its relationship with context.
func validateSkillAgentField(agentVal any, data map[string]any, filePath, contents string) []cue.ValidationError {
	agentStr, isStr := agentVal.(string)
	if !isStr || strings.TrimSpace(agentStr) == "" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "agent field must be a non-empty string",
			Severity: "error",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "agent"),
		}}
	}

	// Warn if agent is set but context is not "fork"
	ctxStr, _ := data["context"].(string)
	if ctxStr != "fork" {
		return []cue.ValidationError{{
			File:     filePath,
			Message:  "agent field is set but context is not 'fork' - consider adding 'context: fork' for sub-agent execution",
			Severity: "warning",
			Source:   cue.SourceAnthropicDocs,
			Line:     FindFrontmatterFieldLine(contents, "agent"),
		}}
	}

	return nil
}
