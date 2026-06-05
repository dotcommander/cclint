package lint

import (
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/scoring"
	"github.com/dotcommander/cclint/internal/textutil"
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
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
		})
	}

	// Check not empty
	if len(strings.TrimSpace(contents)) == 0 {
		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  "Skill file is empty",
			Severity: cue.SeverityError,
			Source:   cue.SourceAnthropicDocs,
			Abort:    true,
		})
	}

	return errors
}

func (l *SkillLinter) ParseContent(contents string) (map[string]any, string, error) {
	// Skills have optional frontmatter
	fm, err := textutil.ParseYAMLFrontmatter(contents)
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
func (l *SkillLinter) GetImprovements(contents string, data map[string]any) []textutil.ImprovementRecommendation {
	return textutil.GetSkillImprovements(contents, data)
}

// PostProcessBatch implements BatchPostProcessor — thin orchestrator over four named helpers.
func (l *SkillLinter) PostProcessBatch(ctx *LinterContext, summary *LintSummary) {
	applyOrphanedSkills(ctx, summary)
	applyGhostTriggers(ctx, summary)
	applyTriggerConflicts(ctx, summary)
	applySkillRefIssues(ctx, summary)
}
