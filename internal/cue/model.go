package cue

import (
	"regexp"
	"strings"
)

// modelAliases are the Claude Code model aliases accepted in a `model:` field.
// Single source for the generated CUE #Model union (see modelUnionCUE / validator.go injection).
var modelAliases = []string{"sonnet", "opus", "haiku", "fable", "best", "sonnet[1m]", "opus[1m]", "fable[1m]", "haiku[1m]", "best[1m]", "opusplan", "opusplan[1m]", "inherit", "claude-opus-5"}

// claudeModelRegexCUE is the CUE-source regex branch matching full model IDs like
// claude-opus-4-5 and claude-fable-5[1m]. Written exactly as it must appear in CUE.
const claudeModelRegexCUE = `=~"^claude-[a-z0-9-]+(\\[[0-9a-z]+\\])?$"`

// fullModelIDPattern is the Go mirror of claudeModelRegexCUE: full model IDs
// like claude-opus-4-5 and claude-fable-5[1m].
var fullModelIDPattern = regexp.MustCompile(`^claude-[a-z0-9-]+(\[[0-9a-z]+\])?$`)

// modelAliasSet is the lookup form of modelAliases.
var modelAliasSet = func() map[string]bool {
	set := make(map[string]bool, len(modelAliases))
	for _, alias := range modelAliases {
		set[alias] = true
	}
	return set
}()

// IsValidModelValue reports whether model is accepted by the #Model union:
// a modelAliases literal or a full claude-* model ID with an optional
// bracket suffix. Single source shared with the generated CUE schema; the
// Go-side agent model warning consumes it instead of a parallel regex.
func IsValidModelValue(model string) bool {
	return modelAliasSet[model] || fullModelIDPattern.MatchString(model)
}

func modelUnionCUE() string {
	members := make([]string, 0, len(modelAliases))
	for _, alias := range modelAliases {
		members = append(members, `"`+alias+`"`)
	}

	return "#Model: " + strings.Join(members, " | ") + " | " + claudeModelRegexCUE
}
