package cue

// schemaDefinition is the single mapping from component keys to their CUE
// root definitions. Unknown keys must not silently select a definition.
func schemaDefinition(component string) (string, bool) {
	switch component {
	case "agent":
		return "#Agent", true
	case "command":
		return "#Command", true
	case "skill":
		return "#Skill", true
	case "settings":
		return "#Settings", true
	case "claude_md":
		return "#ClaudeMD", true
	default:
		return "", false
	}
}
