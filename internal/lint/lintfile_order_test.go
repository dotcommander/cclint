package lint

import (
	"errors"
	"reflect"
	"testing"

	"github.com/dotcommander/cclint/internal/crossfile"
	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/scoring"
	"github.com/dotcommander/cclint/internal/textutil"
)

const (
	pipelineStepPreValidate        = "pre-validate"
	pipelineStepParseContent       = "parse-content"
	pipelineStepCUEValidation      = "cue-validation"
	pipelineStepSpecificValidation = "specific-validation"
	pipelineStepBestPractice       = "best-practice"
	pipelineStepCrossFile          = "cross-file"
	pipelineStepScore              = "score"
	pipelineStepImprovement        = "improvement"
	pipelineStepPostProcess        = "post-process"
	pipelineTestComponentType      = "agent"
	pipelineTestFilePath           = "agents/test.md"
	pipelineTestDataName           = "name"
	pipelineTestContents           = "---\ndescription: |\n  text\n  model: haiku\n---\napi_key: \"sk-AbCdEfGhIjKlMnOpQrStUvWxYz\""
	pipelineTestBody               = "parsed body"
	pipelineTestDuration           = 4242
)

type pipelineTestLinter struct {
	events              []string
	parseErr            error
	preIssues           []cue.ValidationError
	scoreContents       string
	scoreData           map[string]any
	scoreBody           string
	improvementContents string
	improvementData     map[string]any
}

func (l *pipelineTestLinter) record(event string) {
	l.events = append(l.events, event)
}

func (l *pipelineTestLinter) Type() string {
	return pipelineTestComponentType
}

func (l *pipelineTestLinter) FileType() discovery.FileType {
	return discovery.FileTypeAgent
}

func (l *pipelineTestLinter) PreValidate(string, string) []cue.ValidationError {
	l.record(pipelineStepPreValidate)
	return l.preIssues
}

func (l *pipelineTestLinter) ParseContent(string) (map[string]any, string, error) {
	l.record(pipelineStepParseContent)
	return map[string]any{pipelineTestDataName: pipelineTestComponentType}, pipelineTestBody, l.parseErr
}

func (l *pipelineTestLinter) ValidateCUE(*cue.Validator, map[string]any) ([]cue.ValidationError, error) {
	l.record(pipelineStepCUEValidation)
	return []cue.ValidationError{{Message: "cue", Severity: cue.SeverityError}}, nil
}

func (l *pipelineTestLinter) ValidateSpecific(map[string]any, string, string) []cue.ValidationError {
	l.record(pipelineStepSpecificValidation)
	return []cue.ValidationError{
		{Message: "specific error", Severity: cue.SeverityError, Line: 11},
		{Message: "specific warning", Severity: cue.SeverityWarning, Line: 12},
		{Message: "specific suggestion", Severity: cue.SeveritySuggestion, Line: 13},
	}
}

func (l *pipelineTestLinter) ValidateBestPractices(string, string, map[string]any) []cue.ValidationError {
	l.record(pipelineStepBestPractice)
	return []cue.ValidationError{
		{Message: "best-practice error", Severity: cue.SeverityError, Line: 21},
		{Message: "best-practice warning", Severity: cue.SeverityWarning, Line: 22},
		{Message: "best-practice suggestion", Severity: cue.SeveritySuggestion, Line: 23},
	}
}

func (l *pipelineTestLinter) ValidateCrossFile(*crossfile.CrossFileValidator, string, string, map[string]any) []cue.ValidationError {
	l.record(pipelineStepCrossFile)
	return []cue.ValidationError{
		{Message: "cross-file error", Severity: cue.SeverityError, Line: 31},
		{Message: "cross-file warning", Severity: cue.SeverityWarning, Line: 32},
		{Message: "cross-file suggestion", Severity: cue.SeveritySuggestion, Line: 33},
	}
}

func (l *pipelineTestLinter) Score(contents string, data map[string]any, body string) *scoring.QualityScore {
	l.record(pipelineStepScore)
	l.scoreContents = contents
	l.scoreData = data
	l.scoreBody = body
	return &scoring.QualityScore{Overall: 87, Tier: "A", Structural: 31, Practices: 37, Composition: 9, Documentation: 10}
}

func (l *pipelineTestLinter) GetImprovements(contents string, data map[string]any) []textutil.ImprovementRecommendation {
	l.record(pipelineStepImprovement)
	l.improvementContents = contents
	l.improvementData = data
	return []textutil.ImprovementRecommendation{{
		Description: "sentinel improvement",
		PointValue:  7,
		Line:        41,
		Severity:    textutil.SeverityMedium,
	}}
}

func (l *pipelineTestLinter) PostProcess(result *LintResult) {
	l.record(pipelineStepPostProcess)
	result.Duration = pipelineTestDuration
}

func TestLintFileCorePreservesStageAndIssueOrder(t *testing.T) {
	linter := &pipelineTestLinter{}
	result := lintFileCore(pipelineTestFilePath, pipelineTestContents, linter, nil, &crossfile.CrossFileValidator{})

	wantEvents := []string{
		pipelineStepPreValidate,
		pipelineStepParseContent,
		pipelineStepCUEValidation,
		pipelineStepSpecificValidation,
		pipelineStepBestPractice,
		pipelineStepCrossFile,
		pipelineStepScore,
		pipelineStepImprovement,
		pipelineStepPostProcess,
	}
	if !reflect.DeepEqual(linter.events, wantEvents) {
		t.Errorf("linter events = %v, want %v", linter.events, wantEvents)
	}
	if result.File != pipelineTestFilePath {
		t.Errorf("result file = %q, want %q", result.File, pipelineTestFilePath)
	}
	if result.Type != pipelineTestComponentType {
		t.Errorf("result type = %q, want %q", result.Type, pipelineTestComponentType)
	}

	wantIssues := []string{
		"Field 'model' appears to be swallowed by block scalar 'description: |' above — it is parsed as text, not a separate field",
		"cue",
		"specific error",
		"best-practice error",
		"cross-file error",
	}
	gotIssues := make([]string, 0, len(result.Errors))
	for _, issue := range result.Errors {
		gotIssues = append(gotIssues, issue.Message)
	}
	if !reflect.DeepEqual(gotIssues, wantIssues) {
		t.Errorf("error order = %v, want %v", gotIssues, wantIssues)
	}
	if result.Success {
		t.Error("pipeline result succeeded with validation errors")
	}
	wantWarnings := []cue.ValidationError{
		{Message: "specific warning", Severity: cue.SeverityWarning, Line: 12},
		{Message: "best-practice warning", Severity: cue.SeverityWarning, Line: 22},
		{Message: "cross-file warning", Severity: cue.SeverityWarning, Line: 32},
		{
			File:     pipelineTestFilePath,
			Message:  "Possible hardcoded API key detected - use environment variables",
			Severity: cue.SeverityWarning,
			Source:   cue.SourceCClintObserve,
			Line:     6,
		},
		{
			File:     pipelineTestFilePath,
			Message:  "OpenAI API key pattern detected - never commit API keys",
			Severity: cue.SeverityWarning,
			Source:   cue.SourceCClintObserve,
			Line:     6,
		},
	}
	if !reflect.DeepEqual(result.Warnings, wantWarnings) {
		t.Errorf("warnings = %#v, want %#v", result.Warnings, wantWarnings)
	}
	wantSuggestions := []cue.ValidationError{
		{Message: "specific suggestion", Severity: cue.SeveritySuggestion, Line: 13},
		{Message: "best-practice suggestion", Severity: cue.SeveritySuggestion, Line: 23},
		{Message: "cross-file suggestion", Severity: cue.SeveritySuggestion, Line: 33},
	}
	if !reflect.DeepEqual(result.Suggestions, wantSuggestions) {
		t.Errorf("suggestions = %#v, want %#v", result.Suggestions, wantSuggestions)
	}

	wantData := map[string]any{pipelineTestDataName: pipelineTestComponentType}
	if linter.scoreContents != pipelineTestContents || linter.scoreBody != pipelineTestBody || !reflect.DeepEqual(linter.scoreData, wantData) {
		t.Errorf("score inputs = (%q, %#v, %q), want (%q, %#v, %q)", linter.scoreContents, linter.scoreData, linter.scoreBody, pipelineTestContents, wantData, pipelineTestBody)
	}
	if linter.improvementContents != pipelineTestContents || !reflect.DeepEqual(linter.improvementData, wantData) {
		t.Errorf("improvement inputs = (%q, %#v), want (%q, %#v)", linter.improvementContents, linter.improvementData, pipelineTestContents, wantData)
	}

	wantQuality := &scoring.QualityScore{Overall: 87, Tier: "A", Structural: 31, Practices: 37, Composition: 9, Documentation: 10}
	if !reflect.DeepEqual(result.Quality, wantQuality) {
		t.Errorf("quality = %#v, want %#v", result.Quality, wantQuality)
	}
	wantImprovements := []textutil.ImprovementRecommendation{{
		Description: "sentinel improvement",
		PointValue:  7,
		Line:        41,
		Severity:    textutil.SeverityMedium,
	}}
	if !reflect.DeepEqual(result.Improvements, wantImprovements) {
		t.Errorf("improvements = %#v, want %#v", result.Improvements, wantImprovements)
	}
	if result.Duration != pipelineTestDuration {
		t.Errorf("post-processed duration = %d, want %d", result.Duration, pipelineTestDuration)
	}
}

func TestLintFileCoreStopsAtFatalPreValidation(t *testing.T) {
	linter := &pipelineTestLinter{preIssues: []cue.ValidationError{{
		Message:  "fatal",
		Severity: cue.SeverityError,
		Abort:    true,
	}}}
	result := lintFileCore(pipelineTestFilePath, "plain content", linter, nil, nil)

	if !reflect.DeepEqual(linter.events, []string{pipelineStepPreValidate}) {
		t.Errorf("linter events = %v, want pre-validation only", linter.events)
	}
	if result.Success {
		t.Error("fatal pre-validation result succeeded")
	}
}

func TestLintFileCoreStopsAtParseError(t *testing.T) {
	linter := &pipelineTestLinter{parseErr: errors.New("parse failed")}
	result := lintFileCore(pipelineTestFilePath, "plain content", linter, nil, nil)

	wantSteps := []string{pipelineStepPreValidate, pipelineStepParseContent}
	if !reflect.DeepEqual(linter.events, wantSteps) {
		t.Errorf("linter events = %v, want %v", linter.events, wantSteps)
	}
	if result.Success {
		t.Error("parse-error result succeeded")
	}
	if len(result.Errors) != 1 || result.Errors[0].Message != "parse failed" {
		t.Errorf("parse errors = %v, want one parse failed error", result.Errors)
	}
}
