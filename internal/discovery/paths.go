package discovery

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
)

// DiscoverPathsWithRegistry finds files without loading their contents. It uses
// the same registry order, exclusions and symlink policy as full discovery.
// Files that cannot be opened are omitted, as in full discovery. Failures
// encountered only while reading contents remain the consumer's responsibility.
func (fd *FileDiscovery) DiscoverPathsWithRegistry(registry []FileTypeEntry) ([]File, error) {
	return fd.discoverWithRegistry(registry, false)
}

// DiscoverPathsOfTypes scans selected types in canonical registry order,
// returning metadata only. Full lint discovery still loads Contents.
func (fd *FileDiscovery) DiscoverPathsOfTypes(types map[FileType]bool) ([]File, error) {
	registry := make([]FileTypeEntry, 0, len(types))
	for _, entry := range DefaultFileTypes {
		if types[entry.Type] {
			registry = append(registry, entry)
		}
	}
	return fd.DiscoverPathsWithRegistry(registry)
}

// DiscoverPaths scans all default types without reading file contents.
func (fd *FileDiscovery) DiscoverPaths() ([]File, error) {
	return fd.DiscoverPathsWithRegistry(DefaultFileTypes)
}

func (fd *FileDiscovery) discoverWithRegistry(registry []FileTypeEntry, loadContents bool) ([]File, error) {
	var files []File

	for _, ftc := range registry {
		discovered, err := fd.findFilesByPatternMode(ftc.Patterns, ftc.Type, loadContents)
		if err != nil {
			return nil, fmt.Errorf("error discovering %s files: %w", ftc.Type.String(), err)
		}
		files = append(files, discovered...)
	}

	return files, nil
}

func (fd *FileDiscovery) findFilesByPatternMode(patterns []string, fileType FileType, loadContents bool) ([]File, error) {
	var files []File

	canonicalRoot, err := filepath.EvalSymlinks(fd.rootPath)
	if err != nil {
		return nil, fmt.Errorf("cannot resolve discovery root %s: %w", fd.rootPath, err)
	}
	info, err := os.Stat(canonicalRoot)
	if err != nil || !info.IsDir() {
		return nil, fmt.Errorf("discovery root is not an accessible directory: %s", fd.rootPath)
	}
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "**/") {
			// A **-leading glob full-walks the root; bound it. All default
			// registry patterns are literal-prefixed, so this guard only
			// protects future registry entries from unbounded walks.
			files = append(files, fd.boundedStarWalkMode(pattern, fileType, loadContents)...)
			continue
		}
		// Use doublestar for glob matching with ** patterns
		options := []doublestar.GlobOption{}
		if !fd.followSymlinks {
			options = append(options, doublestar.WithNoFollow())
		}
		matches, err := doublestar.Glob(os.DirFS(fd.rootPath), pattern, options...)
		if err != nil {
			return nil, fmt.Errorf("error evaluating pattern %s: %w", pattern, err)
		}

		for _, match := range matches {
			f, ok := fd.processMatchMode(match, fileType, loadContents)
			if ok {
				files = append(files, f)
			}
		}
	}

	return files, nil
}

func (fd *FileDiscovery) boundedStarWalkMode(pattern string, fileType FileType, loadContents bool) []File {
	var files []File
	_ = fs.WalkDir(os.DirFS(fd.rootPath), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // skip unreadable entries, keep walking
		}
		if d.IsDir() {
			if pruneDirNames[d.Name()] || pathDepth(p) >= maxWalkDepth {
				return fs.SkipDir
			}
			return nil
		}
		if ok, matchErr := doublestar.Match(pattern, p); matchErr == nil && ok {
			if f, ok := fd.processMatchMode(p, fileType, loadContents); ok {
				files = append(files, f)
			}
		}
		return nil
	})
	return files
}

func (fd *FileDiscovery) processMatchMode(match string, fileType FileType, loadContents bool) (File, bool) {
	if IsExcluded(match, fd.exclude) {
		return File{}, false
	}
	fullPath := filepath.Join(fd.rootPath, match)

	if !fd.allowedPath(fullPath) {
		return File{}, false
	}
	info, err := os.Stat(fullPath)
	if err != nil || info.IsDir() {
		return File{}, false
	}

	contents := ""
	if loadContents {
		data, err := os.ReadFile(fullPath)
		if err != nil {
			return File{}, false
		}
		contents = string(data)
	} else {
		// Preserve discovery omission for files that cannot be opened, without
		// loading their contents merely to select them for formatting.
		file, err := os.Open(fullPath)
		if err != nil {
			return File{}, false
		}
		_ = file.Close()
	}

	relPath := match
	return File{
		Path:     fullPath,
		RelPath:  relPath,
		Size:     info.Size(),
		Type:     fileType,
		Contents: contents,
	}, true
}
