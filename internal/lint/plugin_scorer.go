package lint

import (
	"fmt"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/textutil"
)

type pluginMetadataGap struct {
	Improvement string
	PointValue  int
	Severity    string
}

func collectPluginMetadataGaps(data map[string]any) []pluginMetadataGap {
	var gaps []pluginMetadataGap

	if _, ok := data["homepage"].(string); !ok {
		gaps = append(gaps, pluginMetadataGap{
			Improvement: "Add 'homepage' field with project URL",
			PointValue:  5,
			Severity:    textutil.SeverityLow,
		})
	}

	if _, ok := data["repository"].(string); !ok {
		gaps = append(gaps, pluginMetadataGap{
			Improvement: "Add 'repository' field with source code URL",
			PointValue:  5,
			Severity:    textutil.SeverityLow,
		})
	}

	if _, ok := data["license"].(string); !ok {
		gaps = append(gaps, pluginMetadataGap{
			Improvement: "Add 'license' field (e.g., MIT, Apache-2.0)",
			PointValue:  5,
			Severity:    textutil.SeverityLow,
		})
	}

	if keywords, ok := data["keywords"].([]any); !ok || len(keywords) == 0 {
		gaps = append(gaps, pluginMetadataGap{
			Improvement: "Add 'keywords' array for discoverability",
			PointValue:  5,
			Severity:    textutil.SeverityLow,
		})
	}

	return gaps
}

// validatePluginBestPractices checks opinionated best practices for plugins.
func validatePluginBestPractices(filePath string, contents string, data map[string]any) []cue.ValidationError {
	var suggestions []cue.ValidationError

	for _, gap := range collectPluginMetadataGaps(data) {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  "Consider " + strings.ToLower(gap.Improvement[:1]) + gap.Improvement[1:],
			Severity: cue.SeveritySuggestion,
			Source:   cue.SourceCClintObserve,
		})
	}

	// Suggest moving experimental components under the experimental wrapper (v2.1.129+)
	for _, field := range []string{"themes", "monitors"} {
		if _, ok := data[field]; ok {
			suggestions = append(suggestions, cue.ValidationError{
				File:     filePath,
				Message:  fmt.Sprintf("Top-level '%s' is deprecated - move under 'experimental.%s' (v2.1.129+; top-level still works but `claude plugin validate` warns)", field, field),
				Severity: cue.SeveritySuggestion,
				Source:   cue.SourceAnthropicDocs,
				Line:     FindJSONFieldLine(contents, field),
			})
		}
	}

	// Check description length
	if desc, ok := data["description"].(string); ok && len(desc) < 50 {
		suggestions = append(suggestions, cue.ValidationError{
			File:     filePath,
			Message:  fmt.Sprintf("Description is only %d chars - consider expanding for clarity", len(desc)),
			Severity: cue.SeveritySuggestion,
			Source:   cue.SourceCClintObserve,
			Line:     FindJSONFieldLine(contents, "description"),
		})
	}

	return suggestions
}

// GetPluginImprovements returns specific improvement recommendations for plugins
func GetPluginImprovements(content string, data map[string]any) []textutil.ImprovementRecommendation {
	var recs []textutil.ImprovementRecommendation

	for _, gap := range collectPluginMetadataGaps(data) {
		recs = append(recs, textutil.ImprovementRecommendation{
			Description: gap.Improvement,
			PointValue:  gap.PointValue,
			Severity:    gap.Severity,
		})
	}

	if _, ok := data["readme"]; !ok {
		recs = append(recs, textutil.ImprovementRecommendation{
			Description: "Add 'readme' field pointing to README file",
			PointValue:  3,
			Severity:    textutil.SeverityLow,
		})
	}

	// Check description quality
	if desc, ok := data["description"].(string); ok {
		if len(desc) < 50 {
			recs = append(recs, textutil.ImprovementRecommendation{
				Description: "Expand description to at least 50 characters",
				PointValue:  5,
				Severity:    textutil.SeverityMedium,
			})
		}
	}

	return recs
}
