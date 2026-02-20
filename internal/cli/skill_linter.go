package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/crossfile"
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
	return cue.TypeSkill
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
	errors = append(errors, checkUnknownSkillFields(data, filePath, contents)...)

	// Name validation (reserved words, format, directory match)
	if name, ok := data["name"].(string); ok {
		errors = append(errors, validateSkillName(name, filePath, contents)...)
	}

	// Validate context field: only valid value is "fork"
	errors = append(errors, validateSkillContextField(data, filePath, contents)...)

	// Validate agent field: non-empty string; warn if context is not "fork"
	if agentVal, ok := data["agent"]; ok {
		errors = append(errors, validateSkillAgentField(agentVal, data, filePath, contents)...)
	}

	// Validate boolean fields
	errors = append(errors, validateSkillBooleanFields(data, filePath, contents)...)

	// Validate argument-hint field
	errors = append(errors, validateSkillArgumentHint(data, filePath, contents)...)

	// Validate hooks (scoped to component events: PreToolUse, PostToolUse, Stop)
	if hooks, ok := data["hooks"]; ok {
		errors = append(errors, ValidateComponentHooks(hooks, filePath)...)
	}

	// Frontmatter suggestion
	errors = append(errors, checkSkillFrontmatter(filePath, contents)...)

	return errors
}

// checkUnknownSkillFields checks for unknown frontmatter fields in skill files.
func checkUnknownSkillFields(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError
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
	return errors
}

// validateSkillContextField validates the context field in skill frontmatter.
func validateSkillContextField(data map[string]any, filePath, contents string) []cue.ValidationError {
	if ctxVal, ok := data["context"]; ok {
		ctxStr, isStr := ctxVal.(string)
		if !isStr || ctxStr != "fork" {
			return []cue.ValidationError{{
				File:     filePath,
				Message:  fmt.Sprintf("context field must be 'fork' (got '%v')", ctxVal),
				Severity: "error",
				Source:   cue.SourceAnthropicDocs,
				Line:     FindFrontmatterFieldLine(contents, "context"),
			}}
		}
	}
	return nil
}

// validateSkillBooleanFields validates boolean fields in skill frontmatter.
func validateSkillBooleanFields(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError

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

	return errors
}

// validateSkillArgumentHint validates the argument-hint field in skill frontmatter.
func validateSkillArgumentHint(data map[string]any, filePath, contents string) []cue.ValidationError {
	var errors []cue.ValidationError
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
	return errors
}

// checkSkillFrontmatter checks if the skill has frontmatter and suggests adding it.
func checkSkillFrontmatter(filePath, contents string) []cue.ValidationError {
	if strings.HasPrefix(contents, "---") {
		return nil
	}

	skillName := extractSkillName(contents, filePath)
	suggestion := "Add YAML frontmatter with name and description (description is critical for skill discovery)"
	if skillName != "" {
		suggestion = fmt.Sprintf("Add frontmatter: ---\nname: %s\ndescription: Brief summary of what this skill does\n--- (description critical for discovery)", skillName)
	}

	return []cue.ValidationError{{
		File:     filePath,
		Message:  suggestion,
		Severity: "suggestion",
		Source:   cue.SourceAnthropicDocs,
	}}
}

// ValidateBestPractices implements BestPracticeValidator interface
func (l *SkillLinter) ValidateBestPractices(filePath, contents string, data map[string]any) []cue.ValidationError {
	return validateSkillBestPractices(filePath, contents, data)
}

// ValidateCrossFile implements CrossFileValidatable interface
func (l *SkillLinter) ValidateCrossFile(crossValidator *crossfile.CrossFileValidator, filePath, contents string, data map[string]any) []cue.ValidationError {
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

// PostProcessBatch implements BatchPostProcessor for orphan detection and ghost triggers.
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

	// Ghost trigger detection: validate skill/agent refs in trigger map tables
	ghostTriggers := ctx.CrossValidator.ValidateTriggerMaps(ctx.RootPath)
	for _, gt := range ghostTriggers {
		summary.TotalErrors++
		summary.FailedFiles++
		// Reference files are not in normal results, so append a new result entry.
		summary.Results = append(summary.Results, LintResult{
			File:    gt.File,
			Type:    "skill",
			Success: false,
			Errors:  []cue.ValidationError{gt},
		})
	}

	// Trigger conflict detection: same keyword routing to different targets across files
	triggerConflicts := ctx.CrossValidator.DetectTriggerConflicts(ctx.RootPath)
	for _, tc := range triggerConflicts {
		summary.TotalSuggestions++
		summary.Results = append(summary.Results, LintResult{
			File:        tc.File,
			Type:        "skill",
			Success:     true, // warnings do not fail the build
			Suggestions: []cue.ValidationError{tc},
		})
	}

	// Skill reference file validation: phantom refs (error) and orphaned refs (info).
	skillRefIssues := ctx.CrossValidator.ValidateSkillReferences(ctx.RootPath)
	for _, issue := range skillRefIssues {
		if issue.Severity == cue.SeverityError {
			summary.TotalErrors++
			// Attach phantom ref errors to the skill file result entry.
			attached := false
			for i, result := range summary.Results {
				if result.File == issue.File {
					summary.Results[i].Errors = append(summary.Results[i].Errors, issue)
					summary.Results[i].Success = false
					attached = true
					break
				}
			}
			if !attached {
				summary.FailedFiles++
				summary.Results = append(summary.Results, LintResult{
					File:    issue.File,
					Type:    "skill",
					Success: false,
					Errors:  []cue.ValidationError{issue},
				})
			}
		} else {
			summary.TotalSuggestions++
			attached := false
			for i, result := range summary.Results {
				if result.File == issue.File {
					summary.Results[i].Suggestions = append(summary.Results[i].Suggestions, issue)
					attached = true
					break
				}
			}
			if !attached {
				summary.Results = append(summary.Results, LintResult{
					File:        issue.File,
					Type:        "skill",
					Success:     true,
					Suggestions: []cue.ValidationError{issue},
				})
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
			Message:  fmt.Sprintf("Skill name '%s' must match parent directory name '%s' (agentskills.io spec: name field)", name, parentDir),
			Severity: "error",
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
