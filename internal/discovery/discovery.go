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

// FileTypeEntry defines the discovery configuration for a file type.
// This enables adding new component types without modifying DiscoverFiles().
type FileTypeEntry struct {
	Type                  FileType
	Patterns              []string
	DetectPatterns        []string
	FallbackBasenames     []string
	FallbackPathSubstring string
}

// DefaultFileTypes is the registry of file types and their discovery patterns.
// To add a new component type, update one entry here and detection/discovery stay aligned.
var DefaultFileTypes = []FileTypeEntry{
	{
		Type:                  FileTypeSkill,
		Patterns:              []string{".claude/skills/**/SKILL.md", "skills/**/SKILL.md", ".claude/skills/**/skill.md", "skills/**/skill.md"},
		FallbackBasenames:     []string{"SKILL.md", "skill.md"},
		FallbackPathSubstring: "",
	},
	{
		Type:                  FileTypeSettings,
		Patterns:              []string{".claude/settings.json", "claude/settings.json"},
		FallbackBasenames:     []string{"settings.json"},
		FallbackPathSubstring: "",
	},
	{
		Type:                  FileTypeContext,
		Patterns:              []string{".claude/CLAUDE.md", "CLAUDE.md"},
		FallbackBasenames:     []string{"CLAUDE.md"},
		FallbackPathSubstring: "",
	},
	{
		Type:                  FileTypePlugin,
		Patterns:              []string{"**/.claude-plugin/plugin.json"},
		FallbackBasenames:     []string{"plugin.json"},
		FallbackPathSubstring: ".claude-plugin",
	},
	{
		Type:                  FileTypeRule,
		Patterns:              []string{".claude/rules/**/*.md", "rules/**/*.md"},
		FallbackBasenames:     nil,
		FallbackPathSubstring: "",
	},
	{
		Type:                  FileTypeAgent,
		Patterns:              []string{".claude/agents/**/*.md", "agents/**/*.md"},
		FallbackBasenames:     nil,
		FallbackPathSubstring: "",
	},
	{
		Type:                  FileTypeOutputStyle,
		Patterns:              []string{".claude/output-styles/**/*.md", "output-styles/**/*.md"},
		FallbackBasenames:     nil,
		FallbackPathSubstring: "",
	},
	{
		Type:                  FileTypeCommand,
		Patterns:              []string{".claude/commands/**/*.md", "commands/**/*.md"},
		FallbackBasenames:     nil,
		FallbackPathSubstring: "",
	},
}

// typePatterns mirrors the canonical detection registry for tests and diagnostics.
var typePatterns = buildTypePatterns(DefaultFileTypes)

func buildTypePatterns(entries []FileTypeEntry) []TypePattern {
	patterns := make([]TypePattern, 0)
	for _, entry := range entries {
		for _, pattern := range entry.detectPatterns() {
			patterns = append(patterns, TypePattern{Pattern: pattern, FileType: entry.Type})
		}
	}
	return patterns
}

func (e FileTypeEntry) detectPatterns() []string {
	if len(e.DetectPatterns) > 0 {
		return e.DetectPatterns
	}
	return e.Patterns
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

	fileType, err := detectFileTypeFromRelativePath(relPath)
	if err != nil {
		return FileTypeUnknown, err
	}
	if fileType != FileTypeUnknown {
		return fileType, nil
	}

	// Fallback: match by basename for files outside standard directories.
	if fileType = detectFileTypeFromBasename(absPath); fileType != FileTypeUnknown {
		return fileType, nil
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
	defer func() { _ = f.Close() }()

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
	rootPath       string
	followSymlinks bool
	exclude        []string
}

// NewFileDiscovery creates a new FileDiscovery instance
func NewFileDiscovery(rootPath string, followSymlinks bool) *FileDiscovery {
	return &FileDiscovery{
		rootPath:       rootPath,
		followSymlinks: followSymlinks,
	}
}

// WithExclude sets glob patterns for files to exclude from discovery.
// Patterns are matched against relative paths using doublestar.Match.
func (fd *FileDiscovery) WithExclude(patterns []string) *FileDiscovery {
	fd.exclude = patterns
	return fd
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
		discovered, err := fd.findFilesByPattern(ftc.Patterns, ftc.Type)
		if err != nil {
			return nil, fmt.Errorf("error discovering %s files: %w", ftc.Type.String(), err)
		}
		files = append(files, discovered...)
	}

	return files, nil
}

// findFilesByPattern finds files matching the given glob patterns.
func (fd *FileDiscovery) findFilesByPattern(patterns []string, fileType FileType) ([]File, error) {
	var files []File

	for _, pattern := range patterns {
		// Use doublestar for glob matching with ** patterns
		matches, err := doublestar.Glob(os.DirFS(fd.rootPath), pattern)
		if err != nil {
			return nil, fmt.Errorf("error evaluating pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			f, ok := fd.processMatch(match, fileType)
			if ok {
				files = append(files, f)
			}
		}
	}

	return files, nil
}

// processMatch converts a glob match into a File, returning false if the match should be skipped.
func (fd *FileDiscovery) processMatch(match string, fileType FileType) (File, bool) {
	if fd.isExcluded(match) {
		return File{}, false
	}
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
		Type:     fileType,
		Contents: string(contents),
	}, true
}

// isExcluded checks if a relative path matches any exclude pattern.
func (fd *FileDiscovery) isExcluded(relPath string) bool {
	for _, pattern := range fd.exclude {
		if matched, err := doublestar.Match(pattern, relPath); err == nil && matched {
			return true
		}
	}
	return false
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
	normalizedPath := filepath.ToSlash(path)
	fileType, err := detectFileTypeFromRelativePath(normalizedPath)
	if err == nil && fileType != FileTypeUnknown {
		return fileType
	}
	return detectFileTypeFromBasename(normalizedPath)
}

func detectFileTypeFromRelativePath(relPath string) (FileType, error) {
	for _, tp := range typePatterns {
		matched, err := doublestar.Match(tp.Pattern, relPath)
		if err != nil {
			return FileTypeUnknown, fmt.Errorf("invalid detection pattern %q: %w", tp.Pattern, err)
		}
		if matched {
			return tp.FileType, nil
		}
	}

	return FileTypeUnknown, nil
}

func detectFileTypeFromBasename(path string) FileType {
	basename := filepath.Base(path)
	normalizedPath := filepath.ToSlash(path)

	for _, entry := range DefaultFileTypes {
		for _, candidate := range entry.FallbackBasenames {
			if !strings.EqualFold(basename, candidate) {
				continue
			}
			if entry.FallbackPathSubstring != "" && !strings.Contains(normalizedPath, entry.FallbackPathSubstring) {
				continue
			}
			return entry.Type
		}
	}

	return FileTypeUnknown
}

// ReadFileContents reads the contents of a file
func (fd *FileDiscovery) ReadFileContents(path string) (string, error) {
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(contents), nil
}
