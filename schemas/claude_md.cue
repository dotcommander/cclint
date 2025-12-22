package schemas

import (
	"lists"
)

// ============================================================================
// CLAUDE.md Schema
// ============================================================================

// Reference types that can be linked
#RefType :: "references" | "memory" | "hooks" | "commands" | "agents"

// Reference link
#RefLink :: {
	path:   string
	title?: string
}

// Section definition
#Section :: {
	heading: string
	content: string

	// Optional reference links
	references?: [...#RefLink]
}

// Required CLAUDE.md sections (based on AGENTS.md template)
#RequiredSection :: "navigating the codebase" |
                    "build & commands" |
                    "using subagents" |
                    "code style" |
                    "testing" |
                    "security" |
                    "configuration"

// Recommended CLAUDE.md sections
#RecommendedSection :: "git commit conventions" |
                       "architecture" |
                       "naming conventions" |
                       "cli tools reference" |
                       "navigating the codebase"

// Rule definition
#Rule :: {
	name:        string
	description: string
}

// CLAUDE.md frontmatter schema
#ClaudeMD :: {
	// Optional metadata
	title?:       string
	description?: string

	// Optional model specification
	model?: "sonnet" | "opus" | "haiku" | "sonnet[1m]" | "opusplan" | "inherit"

	// Optional tool permissions
	"allowed-tools"?: string | [...string]

	// Optional sections
	sections?: [...#Section]

	// Optional rules list
	rules?: [...#Rule]

	// All other fields allowed
	...
}

// Check if required sections are present
#hasRequiredSections: {
	sections: [...#Section]
	hasAll: lists.AllSections(lists.ForSection(sections, func(s #Section) bool {
		string == #RequiredSection
	}))
}

// Validation entry point for CLAUDE.md
validate: {
	input: #ClaudeMD
	result: true
}

// Suggestion level check for missing required sections
suggestRequired: {
	sections: [...#Section]
	missing: lists.Filter(lists.RequiredSections, func(required string) bool {
		!lists.Any(sections, func(s #Section) bool {
			s.heading == required
		})
	})
	suggestion: if len(missing) > 0 then {
		message:  "Consider adding these required sections: \(missing)"
		severity: "suggestion"
	}
}
