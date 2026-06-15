package cue

import "strings"

// modelAliases are the Claude Code model aliases accepted in a `model:` field.
// Single source for the generated CUE #Model union (see modelUnionCUE / validator.go injection).
var modelAliases = []string{"sonnet", "opus", "haiku", "fable", "best", "sonnet[1m]", "opus[1m]", "fable[1m]", "opusplan", "inherit"}

// claudeModelRegexCUE is the CUE-source regex branch matching full model IDs like
// claude-opus-4-5 and claude-fable-5[1m]. Written exactly as it must appear in CUE.
const claudeModelRegexCUE = `=~"^claude-[a-z0-9-]+(\\[[0-9a-z]+\\])?$"`

func modelUnionCUE() string {
	members := make([]string, 0, len(modelAliases))
	for _, alias := range modelAliases {
		members = append(members, `"`+alias+`"`)
	}

	return "#Model: " + strings.Join(members, " | ") + " | " + claudeModelRegexCUE
}
