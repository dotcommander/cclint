package lint

import (
	"fmt"

	"github.com/dotcommander/cclint/internal/discovery"
)

// ComponentLinterFactory creates a component linter for a project root.
// Most linters do not need the root; plugin linting uses it to validate paths.
type ComponentLinterFactory func(rootPath string) ComponentLinter

// LinterEntry describes one supported component type.
type LinterEntry struct {
	FileType discovery.FileType
	Name     string
	Default  bool
	New      ComponentLinterFactory
}

var linterRegistry = []LinterEntry{ //nolint:gochecknoglobals // Immutable canonical component registry.
	{FileType: discovery.FileTypeAgent, Name: "agents", Default: true, New: func(string) ComponentLinter { return NewAgentLinter() }},
	{FileType: discovery.FileTypeCommand, Name: "commands", Default: true, New: func(string) ComponentLinter { return NewCommandLinter() }},
	{FileType: discovery.FileTypeSkill, Name: "skills", Default: true, New: func(string) ComponentLinter { return NewSkillLinter() }},
	{FileType: discovery.FileTypeSettings, Name: "settings", Default: true, New: func(string) ComponentLinter { return NewSettingsLinter() }},
	{FileType: discovery.FileTypeContext, Name: "context", Default: false, New: func(string) ComponentLinter { return NewContextLinter() }},
	{FileType: discovery.FileTypeRule, Name: "rules", Default: true, New: func(string) ComponentLinter { return NewRuleLinter() }},
	{FileType: discovery.FileTypeOutputStyle, Name: "output-styles", Default: true, New: func(string) ComponentLinter { return NewOutputStyleLinter() }},
	{FileType: discovery.FileTypePlugin, Name: "plugins", Default: true, New: func(rootPath string) ComponentLinter { return NewPluginLinter(rootPath) }},
}

// DefaultLinters returns the component linters included in a full lint run.
func DefaultLinters() []LinterEntry {
	linters := make([]LinterEntry, 0, len(linterRegistry))
	for _, entry := range linterRegistry {
		if entry.Default {
			linters = append(linters, entry)
		}
	}
	return linters
}

// LinterForType returns the registered linter for a discovered file type.
func LinterForType(fileType discovery.FileType) (LinterEntry, bool) {
	for _, entry := range linterRegistry {
		if entry.FileType == fileType {
			return entry, true
		}
	}
	return LinterEntry{}, false
}

func lintStandalone(rootPath string, quiet, verbose, noCycleCheck bool, exclude []string, fileType discovery.FileType) (*LintSummary, error) {
	entry, ok := LinterForType(fileType)
	if !ok {
		return nil, fmt.Errorf("no linter for type %s", fileType)
	}

	ctx, err := NewLinterContext(rootPath, quiet, verbose, noCycleCheck, exclude)
	if err != nil {
		return nil, err
	}
	return lintBatch(ctx, entry.New(ctx.RootPath)), nil
}
