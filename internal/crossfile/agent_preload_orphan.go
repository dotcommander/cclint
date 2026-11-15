package crossfile

import "github.com/dotcommander/cclint/internal/cue"

// collectAgentPreloadReferences adds only locally resolved preload identities.
// Runtime-owned plugin-qualified and built-in names do not identify indexed skills.
func (v *CrossFileValidator) collectAgentPreloadReferences(contents string, referencedSkills map[string]bool) {
	frontmatter, err := cue.ParseFrontmatter(contents)
	if err != nil {
		return // Parsing diagnostics belong to the frontmatter validator.
	}
	for _, name := range frontmatterSkillNames(frontmatter.Data["skills"]) {
		if BuiltInSkillNames[name] || IsPluginNamespacedRef(name) {
			continue
		}
		if _, exists := v.skills[name]; exists {
			referencedSkills[name] = true
		}
	}
}
