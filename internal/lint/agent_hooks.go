package lint

import (
	"github.com/dotcommander/cclint/internal/cue"
)

// validateAgentHooks validates the hooks field.
func validateAgentHooks(data map[string]any, filePath string) []cue.ValidationError {
	hooks, ok := data["hooks"]
	if !ok {
		return nil
	}
	return ValidateComponentHooks(hooks, filePath)
}
