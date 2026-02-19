package discovery

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// TypePattern maps a glob pattern to a FileType for type detection.
// Patterns are matched in order; first match wins.
type TypePattern struct {
	Pattern  string
	FileType FileType
}

// typePatterns defines the canonical patterns for detecting component types.
// These patterns mirror those used in DiscoverFiles() for consistency.
// Order matters: more specific patterns should come first.
var typePatterns = []TypePattern{
	// Skills - most specific (requires SKILL.md filename)
	{".claude/skills/**/SKILL.md", FileTypeSkill},
	{"skills/**/SKILL.md", FileTypeSkill},
	// Catch misnamed lowercase skill files so PreValidate can report the casing error
	{".claude/skills/**/skill.md", FileTypeSkill},
	{"skills/**/skill.md", FileTypeSkill},

	// Settings - exact filename match
	{".claude/settings.json", FileTypeSettings},
	{"claude/settings.json", FileTypeSettings},

	// Context - exact filename match
	{".claude/CLAUDE.md", FileTypeContext},
	{"CLAUDE.md", FileTypeContext},

	// Plugins - specific directory structure
	{"**/.claude-plugin/plugin.json", FileTypePlugin},

	// Rules - .claude/rules/**/*.md (before agents/commands to avoid misdetection)
	{".claude/rules/**/*.md", FileTypeRule},
	{"rules/**/*.md", FileTypeRule},

	// Agents - directory-based (after more specific patterns)
	{".claude/agents/**/*.md", FileTypeAgent},
	{"agents/**/*.md", FileTypeAgent},

	// Output Styles - directory-based
	{".claude/output-styles/**/*.md", FileTypeOutputStyle},
	{"output-styles/**/*.md", FileTypeOutputStyle},

	// Commands - directory-based
	{".claude/commands/**/*.md", FileTypeCommand},
	{"commands/**/*.md", FileTypeCommand},
}

// FileTypeEntry defines the discovery configuration for a file type.
// This enables adding new component types without modifying DiscoverFiles().
type FileTypeEntry struct {
	Type     FileType
	Patterns []string
}

// DefaultFileTypes is the registry of file types and their discovery patterns.
// To add a new component type, simply add an entry here - no code changes needed.
var DefaultFileTypes = []FileTypeEntry{
	{Type: FileTypeAgent, Patterns: []string{".claude/agents/**/*.md", "agents/**/*.md"}},
	{Type: FileTypeCommand, Patterns: []string{".claude/commands/**/*.md", "commands/**/*.md"}},
	{Type: FileTypeSettings, Patterns: []string{".claude/settings.json", "claude/settings.json"}},
	{Type: FileTypeContext, Patterns: []string{".claude/CLAUDE.md", "CLAUDE.md"}},
	{Type: FileTypeSkill, Patterns: []string{".claude/skills/*/SKILL.md", "skills/*/SKILL.md", ".claude/skills/*/skill.md", "skills/*/skill.md"}},
	{Type: FileTypePlugin, Patterns: []string{"**/.claude-plugin/plugin.json"}},
	{Type: FileTypeRule, Patterns: []string{".claude/rules/**/*.md", "rules/**/*.md"}},
	{Type: FileTypeOutputStyle, Patterns: []string{".claude/output-styles/**/*.md", "output-styles/**/*.md"}},
}

// DetectFileType determines the component type from a file path using glob pattern matching.
//
// This function uses the same patterns as DiscoverFiles() to ensure consistency
// between file discovery and single-file linting.
//
// Parameters:
//   - absPath: Absolute path to the file (must exist)
//   - rootPath: Project root directory for computing relative paths
//
// Returns:
//   - FileType: The detected component type
//   - error: Descriptive error if type cannot be determined
//
// Type detection priority:
//  1. Glob pattern match against relative path (most reliable)
//  2. Basename match for special files (SKILL.md, CLAUDE.md, etc.)
//  3. Error with actionable message for ambiguous cases
//
// Example:
//
//	fileType, err := DetectFileType("/home/user/.claude/agents/my-agent.md", "/home/user/.claude")
//	// fileType == FileTypeAgent
func DetectFileType(absPath, rootPath string) (FileType, error) {
	// Compute relative path for pattern matching
	relPath, err := filepath.Rel(rootPath, absPath)
	if err != nil {
		return FileTypeUnknown, fmt.Errorf("cannot compute relative path from %s to %s: %w", rootPath, absPath, err)
	}

	// Normalize to forward slashes for cross-platform pattern matching
	relPath = filepath.ToSlash(relPath)

	// Reject paths that escape the root (start with ..)
	if strings.HasPrefix(relPath, "..") {
		return FileTypeUnknown, fmt.Errorf("file is outside project root: %s", absPath)
	}

	// Try each pattern in order (first match wins)
	for _, tp := range typePatterns {
		matched, err := doublestar.Match(tp.Pattern, relPath)
		if err != nil {
			continue // Pattern error, try next
		}
		if matched {
			return tp.FileType, nil
		}
	}

	// Fallback: match by basename for files outside standard directories
	basename := filepath.Base(absPath)
	switch {
	case strings.EqualFold(basename, "SKILL.md"):
		return FileTypeSkill, nil
	case strings.EqualFold(basename, "CLAUDE.md"):
		return FileTypeContext, nil
	case basename == "settings.json":
		return FileTypeSettings, nil
	case basename == "plugin.json" && strings.Contains(absPath, ".claude-plugin"):
		return FileTypePlugin, nil
	}

	// Cannot determine type - provide actionable error
	ext := strings.ToLower(filepath.Ext(absPath))
	switch ext {
	case ".md":
		return FileTypeUnknown, fmt.Errorf(
			"cannot determine type: %s is a .md file but not in agents/, commands/, skills/, or output-styles/ directory. "+
				"Use --type to specify (agent, command, skill, context, output-style)", relPath)
	case ".json":
		return FileTypeUnknown, fmt.Errorf(
			"cannot determine type: %s is a .json file but not settings.json or plugin.json. "+
				"Use --type to specify (settings, plugin)", relPath)
	case "":
		return FileTypeUnknown, fmt.Errorf(
			"unsupported file: %s has no extension. cclint validates .md and .json files only", filepath.Base(absPath))
	default:
		return FileTypeUnknown, fmt.Errorf(
			"unsupported file type: %s. cclint validates .md and .json files only", ext)
	}
}

// ValidateFilePath performs comprehensive validation of a file path for linting.
//
// This function checks all preconditions required before linting a file:
//   - File exists
//   - Path is a file (not directory)
//   - File is readable
//   - File is not empty
//   - File is not binary
//
// Returns descriptive errors for each failure mode to guide user action.
//
// Example:
//
//	absPath, err := ValidateFilePath("./agents/my-agent.md")
//	if err != nil {
//	    fmt.Fprintf(os.Stderr, "Error: %v\n", err)
//	    os.Exit(2)
//	}
func ValidateFilePath(path string) (absPath string, err error) {
	// Resolve to absolute path
	absPath, err = filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("invalid path %q: %w", path, err)
	}

	// Check file exists
	info, err := os.Lstat(absPath) // Lstat to detect symlinks
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("file not found: %s", absPath)
		}
		if os.IsPermission(err) {
			return "", fmt.Errorf("permission denied: %s", absPath)
		}
		return "", fmt.Errorf("cannot access file: %s: %w", absPath, err)
	}

	// Handle symlinks
	if info.Mode()&os.ModeSymlink != 0 {
		realPath, evalErr := filepath.EvalSymlinks(absPath)
		if evalErr != nil {
			return "", fmt.Errorf("cannot resolve symlink %s: %w", absPath, evalErr)
		}
		absPath = realPath
		info, err = os.Stat(absPath)
		if err != nil {
			return "", fmt.Errorf("symlink target inaccessible: %s: %w", absPath, err)
		}
	}

	// Reject directories
	if info.IsDir() {
		return "", fmt.Errorf("path is a directory, not a file: %s", absPath)
	}

	// Check not empty
	if info.Size() == 0 {
		return "", fmt.Errorf("file is empty: %s", absPath)
	}

	// Read first bytes to check for binary content
	f, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %s: %w", absPath, err)
	}
	defer f.Close()

	// Read first 512 bytes for binary detection
	buf := make([]byte, 512)
	n, err := f.Read(buf)
	if err != nil {
		return "", fmt.Errorf("cannot read file: %s: %w", absPath, err)
	}

	// Check for null bytes (binary indicator)
	if bytes.Contains(buf[:n], []byte{0}) {
		return "", fmt.Errorf("file appears to be binary, not text: %s", absPath)
	}

	return absPath, nil
}

// File represents a discovered file with its metadata
type File struct {
	Path     string
	RelPath  string
	Size     int64
	Type     FileType
	Contents string
}

// FileType categorizes discovered files
type FileType int

const (
	FileTypeUnknown FileType = iota
	FileTypeAgent
	FileTypeCommand
	FileTypeSettings
	FileTypeContext
	FileTypeSkill
	FileTypeConfig
	FileTypePlugin
	FileTypeRule
	FileTypeOutputStyle
)

// String returns the human-readable name of the file type.
func (ft FileType) String() string {
	switch ft {
	case FileTypeAgent:
		return "agent"
	case FileTypeCommand:
		return "command"
	case FileTypeSettings:
		return "settings"
	case FileTypeContext:
		return "context"
	case FileTypeSkill:
		return "skill"
	case FileTypePlugin:
		return "plugin"
	case FileTypeRule:
		return "rule"
	case FileTypeOutputStyle:
		return "output-style"
	default:
		return "unknown"
	}
}

// ParseFileType converts a string to a FileType.
// Valid values: agent, command, skill, settings, context, plugin, rule.
// Returns FileTypeUnknown and an error for invalid input.
func ParseFileType(s string) (FileType, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "agent", "agents":
		return FileTypeAgent, nil
	case "command", "commands":
		return FileTypeCommand, nil
	case "skill", "skills":
		return FileTypeSkill, nil
	case "settings":
		return FileTypeSettings, nil
	case "context":
		return FileTypeContext, nil
	case "plugin", "plugins":
		return FileTypePlugin, nil
	case "rule", "rules":
		return FileTypeRule, nil
	case "output-style", "output-styles":
		return FileTypeOutputStyle, nil
	default:
		return FileTypeUnknown, fmt.Errorf(
			"invalid type %q: valid types are agent, command, skill, settings, context, plugin, rule, output-style", s)
	}
}

// FileDiscovery manages file discovery operations
type FileDiscovery struct {
	rootPath   string
	followSymlinks bool
}

// NewFileDiscovery creates a new FileDiscovery instance
func NewFileDiscovery(rootPath string, followSymlinks bool) *FileDiscovery {
	return &FileDiscovery{
		rootPath:       rootPath,
		followSymlinks: followSymlinks,
	}
}

// DiscoverFiles finds all relevant files in the project.
// It iterates over the DefaultFileTypes registry, making it easy to add
// new component types without modifying this method.
func (fd *FileDiscovery) DiscoverFiles() ([]File, error) {
	return fd.DiscoverFilesWithRegistry(DefaultFileTypes)
}

// DiscoverFilesWithRegistry finds files using a custom registry.
// This allows filtering or extending the default file types.
func (fd *FileDiscovery) DiscoverFilesWithRegistry(registry []FileTypeEntry) ([]File, error) {
	var files []File

	for _, ftc := range registry {
		discovered, err := fd.findFilesByPattern(ftc.Patterns)
		if err != nil {
			return nil, fmt.Errorf("error discovering %s files: %w", ftc.Type.String(), err)
		}
		for _, f := range discovered {
			f.Type = ftc.Type
			files = append(files, f)
		}
	}

	return files, nil
}

// findFilesByPattern finds files matching the given glob patterns
func (fd *FileDiscovery) findFilesByPattern(patterns []string) ([]File, error) {
	var files []File

	for _, pattern := range patterns {
		// Use doublestar for glob matching with ** patterns
		matches, err := doublestar.Glob(os.DirFS(fd.rootPath), pattern)
		if err != nil {
			return nil, fmt.Errorf("error evaluating pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			f, ok := fd.processMatch(match)
			if ok {
				files = append(files, f)
			}
		}
	}

	return files, nil
}

// processMatch converts a glob match into a File, returning false if the match should be skipped.
func (fd *FileDiscovery) processMatch(match string) (File, bool) {
	fullPath := filepath.Join(fd.rootPath, match)

	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		return File{}, false
	}

	if info.Mode()&os.ModeSymlink != 0 {
		resolved, resolvedInfo, ok := fd.resolveSymlink(fullPath)
		if !ok {
			return File{}, false
		}
		match = resolved
		info = resolvedInfo
	}

	contents, err := os.ReadFile(fullPath)
	if err != nil {
		return File{}, false
	}

	relPath := match
	return File{
		Path:     fullPath,
		RelPath:  relPath,
		Size:     info.Size(),
		Type:     fd.determineFileType(relPath),
		Contents: string(contents),
	}, true
}

// resolveSymlink follows a symlink if configured, returning the resolved path and info.
// Returns false if the symlink should be skipped.
func (fd *FileDiscovery) resolveSymlink(fullPath string) (string, os.FileInfo, bool) {
	if !fd.followSymlinks {
		return "", nil, false
	}

	realPath, err := filepath.EvalSymlinks(fullPath)
	if err != nil {
		return "", nil, false
	}

	if !strings.HasPrefix(realPath, fd.rootPath) {
		return "", nil, false
	}

	info, err := os.Stat(realPath)
	if err != nil {
		return "", nil, false
	}

	return realPath, info, true
}

// determineFileType determines the file type based on its path.
// Uses path component matching (not substring) to avoid false positives
// like matching "/my-agents-backup/" when looking for "/agents/".
func (fd *FileDiscovery) determineFileType(path string) FileType {
	lowerPath := strings.ToLower(path)

	// Rules must be checked before agents/commands (rules/ is inside .claude/)
	if hasPathComponent(lowerPath, ".claude/rules") && strings.HasSuffix(lowerPath, ".md") {
		return FileTypeRule
	}
	// Output styles (before agents/commands to avoid misdetection)
	if (hasPathComponent(lowerPath, ".claude/output-styles") || hasPathComponent(lowerPath, "output-styles")) && strings.HasSuffix(lowerPath, ".md") {
		return FileTypeOutputStyle
	}
	if hasPathComponent(lowerPath, "agents") && strings.HasSuffix(lowerPath, ".md") {
		return FileTypeAgent
	}
	if hasPathComponent(lowerPath, "commands") && strings.HasSuffix(lowerPath, ".md") {
		return FileTypeCommand
	}
	if strings.HasSuffix(lowerPath, "settings.json") {
		return FileTypeSettings
	}
	if strings.HasSuffix(lowerPath, "claude.md") {
		return FileTypeContext
	}
	if hasPathComponent(lowerPath, ".claude-plugin") && strings.HasSuffix(lowerPath, "plugin.json") {
		return FileTypePlugin
	}

	return FileTypeUnknown
}

// hasPathComponent checks if a path contains a directory component.
// Unlike strings.Contains, this matches on path boundaries to avoid
// false positives (e.g., "agents" won't match "my-agents-backup").
func hasPathComponent(path, component string) bool {
	// Handle both forward and back slashes for cross-platform support
	normalizedPath := strings.ReplaceAll(path, "\\", "/")
	normalizedComponent := strings.ReplaceAll(component, "\\", "/")
	sep := "/"

	// Check: /component/, /component (end), component/ (start)
	if strings.Contains(normalizedPath, sep+normalizedComponent+sep) {
		return true
	}
	if strings.HasSuffix(normalizedPath, sep+normalizedComponent) {
		return true
	}
	if strings.HasPrefix(normalizedPath, normalizedComponent+sep) {
		return true
	}
	// Exact match (path == component)
	if normalizedPath == normalizedComponent {
		return true
	}
	return false
}

// ReadFileContents reads the contents of a file
func (fd *FileDiscovery) ReadFileContents(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}