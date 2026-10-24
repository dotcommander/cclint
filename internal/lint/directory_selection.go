package lint

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/project"
)

// fileWithHint pairs a file path with an optional type hint inferred during
// directory expansion (e.g., directory named "command" → "command").
type fileWithHint struct {
	Path     string
	TypeHint string // empty = auto-detect, non-empty = use as typeOverride
}

// expandDirectories expands any directory paths in the input list to their
// contained .md and .json files. Non-directory paths are kept as-is.
// Hidden child directories other than .claude and .claude-plugin are skipped.
//
// When typeOverride is empty, files are filtered through DetectFileType so
// that non-component files (references/, prompts/, usage-data/, etc.) are
// silently excluded. If DetectFileType fails, the walked directory's base
// name is checked via ParseFileType as a fallback — this allows singular
// directory names (e.g., "command/", "agent/") to work alongside the
// standard plural forms. When typeOverride is set, all .md/.json files are
// included since the user is explicitly declaring the component type.
func expandDirectories(paths []string, typeOverride string) ([]fileWithHint, error) {
	return expandDirectoriesAtRoot(paths, "", typeOverride)
}

// expandDirectoriesAtRoot applies the same explicit-root semantics as file
// validation before inspecting directories. Expanded paths are absolute when
// rooted, so single-file validation cannot prepend the root a second time.
func expandDirectoriesAtRoot(paths []string, rootPath, typeOverride string) ([]fileWithHint, error) {
	result := make([]fileWithHint, 0, len(paths))
	for _, p := range paths {
		directoryPath := p
		if rootPath != "" && !filepath.IsAbs(directoryPath) {
			directoryPath = filepath.Join(rootPath, directoryPath)
		}
		// Stat the lexically-resolved absolute path: a cwd-relative path with
		// ".." segments can disagree with the kernel's physical parent chain
		// when cwd contains symlinks (logical vs physical cwd divergence).
		absDir, err := filepath.Abs(directoryPath)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve directory %s: %w", p, err)
		}
		info, err := os.Stat(absDir)
		if err != nil || !info.IsDir() {
			result = append(result, fileWithHint{Path: p, TypeHint: typeOverride})
			continue
		}
		var detectionRoot string
		if rootPath != "" {
			detectionRoot, err = filepath.Abs(rootPath)
		} else {
			detectionRoot, err = project.FindProjectRoot(absDir)
		}
		if err != nil {
			return nil, fmt.Errorf("cannot resolve directory %s: %w", p, err)
		}

		// Pre-check: can the directory name itself hint at a component type?
		dirHint := ""
		if typeOverride == "" {
			if ft, parseErr := discovery.ParseFileType(filepath.Base(absDir)); parseErr == nil {
				dirHint = ft.String()
			}
		}

		if err := filepath.WalkDir(absDir, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				if path != absDir && strings.HasPrefix(d.Name(), ".") && d.Name() != ".claude" && d.Name() != ".claude-plugin" {
					return fs.SkipDir
				}
				return nil
			}
			ext := strings.ToLower(filepath.Ext(path))
			if ext != ".md" && ext != ".json" {
				return nil
			}
			if typeOverride != "" {
				result = append(result, fileWithHint{Path: path, TypeHint: typeOverride})
				return nil
			}
			absPath, err := filepath.Abs(path)
			if err != nil {
				return err
			}
			if _, err := discovery.DetectFileType(absPath, detectionRoot); err == nil {
				result = append(result, fileWithHint{Path: path})
			} else if dirHint != "" {
				// Skills require SKILL.md filename — don't apply dirHint to arbitrary .md files
				if dirHint == "skill" && !strings.EqualFold(filepath.Base(path), "SKILL.md") {
					return nil
				}
				result = append(result, fileWithHint{Path: path, TypeHint: dirHint})
			}
			return nil
		}); err != nil {
			return nil, fmt.Errorf("cannot walk directory %s: %w", p, err)
		}
	}
	return result, nil
}
