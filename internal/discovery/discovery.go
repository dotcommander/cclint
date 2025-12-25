package discovery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

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
)

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

// DiscoverFiles finds all relevant files in the project
func (fd *FileDiscovery) DiscoverFiles() ([]File, error) {
	var files []File

	// Agent files
	agentFiles, err := fd.findFilesByPattern([]string{".claude/agents/**/*.md", "agents/**/*.md"})
	if err != nil {
		return nil, fmt.Errorf("error discovering agent files: %w", err)
	}
	for _, f := range agentFiles {
		f.Type = FileTypeAgent
		files = append(files, f)
	}

	// Command files
	commandFiles, err := fd.findFilesByPattern([]string{".claude/commands/**/*.md", "commands/**/*.md"})
	if err != nil {
		return nil, fmt.Errorf("error discovering command files: %w", err)
	}
	for _, f := range commandFiles {
		f.Type = FileTypeCommand
		files = append(files, f)
	}

	// Settings files
	settingsFiles, err := fd.findFilesByPattern([]string{".claude/settings.json", "claude/settings.json"})
	if err != nil {
		return nil, fmt.Errorf("error discovering settings files: %w", err)
	}
	for _, f := range settingsFiles {
		f.Type = FileTypeSettings
		files = append(files, f)
	}

	// Context files
	contextFiles, err := fd.findFilesByPattern([]string{".claude/CLAUDE.md", "CLAUDE.md"})
	if err != nil {
		return nil, fmt.Errorf("error discovering context files: %w", err)
	}
	for _, f := range contextFiles {
		f.Type = FileTypeContext
		files = append(files, f)
	}

	// Skill files
	skillFiles, err := fd.findFilesByPattern([]string{".claude/skills/**/SKILL.md", "skills/**/SKILL.md"})
	if err != nil {
		return nil, fmt.Errorf("error discovering skill files: %w", err)
	}
	for _, f := range skillFiles {
		f.Type = FileTypeSkill
		files = append(files, f)
	}

	// Plugin files
	pluginFiles, err := fd.findFilesByPattern([]string{"**/.claude-plugin/plugin.json"})
	if err != nil {
		return nil, fmt.Errorf("error discovering plugin files: %w", err)
	}
	for _, f := range pluginFiles {
		f.Type = FileTypePlugin
		files = append(files, f)
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

		// Process each match
		for _, match := range matches {
			// Get full path by joining with root
			fullPath := filepath.Join(fd.rootPath, match)

			// Get file info
			info, err := os.Stat(fullPath)
			if err != nil {
				continue // Skip files we can't stat
			}

			// Skip directories
			if info.IsDir() {
				continue
			}

			// For symlinks, follow if requested
			if info.Mode()&os.ModeSymlink != 0 {
				if fd.followSymlinks {
					realPath, err := filepath.EvalSymlinks(fullPath)
					if err != nil {
						continue
					}
					// Validate that symlink target is within project root
					if !strings.HasPrefix(realPath, fd.rootPath) {
						continue
					}
					match = realPath
					info, err = os.Stat(match)
					if err != nil {
						continue
					}
				} else {
					continue // Skip symlinks if not following
				}
			}

			// match is already relative to rootPath (from os.DirFS)
			relPath := match

			// Read file contents
			contents, err := os.ReadFile(fullPath)
			if err != nil {
				continue // Skip files we can't read
			}

			// Determine file type based on path
			fileType := fd.determineFileType(relPath)

			files = append(files, File{
				Path:     fullPath,
				RelPath:  relPath,
				Size:     info.Size(),
				Type:     fileType,
				Contents: string(contents),
			})
		}
	}

	return files, nil
}

// determineFileType determines the file type based on its path
func (fd *FileDiscovery) determineFileType(path string) FileType {
	lowerPath := strings.ToLower(path)

	if strings.Contains(lowerPath, "agents") && strings.HasSuffix(lowerPath, ".md") {
		return FileTypeAgent
	}
	if strings.Contains(lowerPath, "commands") && strings.HasSuffix(lowerPath, ".md") {
		return FileTypeCommand
	}
	if strings.HasSuffix(lowerPath, "settings.json") {
		return FileTypeSettings
	}
	if strings.HasSuffix(lowerPath, "claude.md") {
		return FileTypeContext
	}
	if strings.Contains(lowerPath, ".claude-plugin") && strings.HasSuffix(lowerPath, "plugin.json") {
		return FileTypePlugin
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