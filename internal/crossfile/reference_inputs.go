package crossfile

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/cue"
	"github.com/dotcommander/cclint/internal/discovery"
)

// CrossFileValidator validates references between components
type CrossFileValidator struct {
	agents            map[string]discovery.File
	skills            map[string]discovery.File
	commands          map[string]discovery.File
	rootPath          string
	userScopeAgentDir string
	userScopeSkillDir string
}

// NewCrossFileValidator creates a validator with indexed files.
// rootPath is optional; if provided it enables trigger map scanning in orphan detection.
func NewCrossFileValidator(files []discovery.File, rootPath ...string) *CrossFileValidator {
	v := &CrossFileValidator{
		agents:   make(map[string]discovery.File),
		skills:   make(map[string]discovery.File),
		commands: make(map[string]discovery.File),
	}
	if len(rootPath) > 0 {
		v.rootPath = rootPath[0]
	}
	if homeDir, err := os.UserHomeDir(); err == nil {
		v.userScopeAgentDir = filepath.Join(homeDir, ".claude", "agents")
		v.userScopeSkillDir = filepath.Join(homeDir, ".claude", "skills")
	}

	// Index files. User-space agents (agents/**/*.md) are processed first so they
	// always win over plugin-shipped agents of the same bare name.
	// Two passes: (1) user-space agents/skills/commands, (2) plugin agents (fill gaps only).
	for _, f := range files {
		switch f.Type {
		case discovery.FileTypeAgent:
			if isPluginAgentRelPath(f.RelPath) {
				continue // handled in second pass
			}
			name := ExtractAgentName(f.RelPath)
			v.agents[name] = f
		case discovery.FileTypeSkill:
			name := ExtractSkillName(f.RelPath)
			v.skills[name] = f
		case discovery.FileTypeCommand:
			name := ExtractCommandName(f.RelPath)
			v.commands[name] = f
		}
	}
	// Second pass: plugin agents fill gaps — never overwrite a user-space entry.
	for _, f := range files {
		if f.Type == discovery.FileTypeAgent && isPluginAgentRelPath(f.RelPath) {
			name := ExtractAgentName(f.RelPath)
			if _, exists := v.agents[name]; !exists {
				v.agents[name] = f
			}
		}
	}

	return v
}

// checkUnusedAllowedTools detects tools declared in allowed-tools but never used.
func (v *CrossFileValidator) checkUnusedAllowedTools(filePath, contents string, frontmatter map[string]any) []cue.ValidationError {
	var errors []cue.ValidationError

	tools := ParseAllowedTools(allowedToolsText(frontmatter["allowed-tools"]))
	for _, tool := range tools {
		if IsToolUsed(tool, contents) {
			continue
		}

		isDeclarative := strings.Contains(strings.ToLower(contents), "saved as") ||
			strings.Contains(strings.ToLower(contents), "write to") ||
			strings.Contains(strings.ToLower(contents), "output to") ||
			strings.Contains(strings.ToLower(contents), "save to") ||
			strings.Contains(contents, ".md") ||
			strings.Contains(contents, ".json")

		var message string
		switch {
		case isDeclarative && (tool == "Write" || tool == "Read"):
			message = fmt.Sprintf("allowed-tools declares '%s' - consider making tool usage more explicit for LLM (e.g., 'Use Write tool to create...')", tool)
		case isDeclarative:
			message = fmt.Sprintf("allowed-tools declares '%s' without obvious invocation (consider making tool usage explicit)", tool)
		default:
			message = fmt.Sprintf("allowed-tools declares '%s' but it's never used in command body", tool)
		}

		errors = append(errors, cue.ValidationError{
			File:     filePath,
			Message:  message,
			Severity: cue.SeverityInfo,
			Source:   cue.SourceCClintObserve,
		})
	}

	return errors
}

func allowedToolsText(value any) string {
	switch value := value.(type) {
	case string:
		return value
	case []string:
		return strings.Join(value, ",")
	case []any:
		var entries []string
		for _, entry := range value {
			if entry, ok := entry.(string); ok {
				entries = append(entries, entry)
			}
		}
		return strings.Join(entries, ",")
	default:
		return ""
	}
}

// hasResolvableSkill keeps runtime-owned references out of local dangling checks.
func (v *CrossFileValidator) hasResolvableSkill(name string) bool {
	if BuiltInSkillNames[name] || IsPluginNamespacedRef(name) {
		return true
	}
	if _, exists := v.skills[name]; exists {
		return true
	}
	return v.hasUserSkillFile(name)
}

func (v *CrossFileValidator) hasUserSkillFile(name string) bool {
	if v.userScopeSkillDir == "" || name == "" || name == "." || name == ".." ||
		filepath.IsAbs(name) || strings.ContainsAny(name, "/\\") {
		return false
	}
	for _, filename := range []string{"SKILL.md", "skill.md"} {
		info, err := os.Stat(filepath.Join(v.userScopeSkillDir, name, filename))
		if err == nil && !info.IsDir() {
			return true
		}
	}
	return false
}
