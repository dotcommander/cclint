package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/format"
)

// collectFilesToFormat determines which files to format based on args and flags.
func collectFilesToFormat(cmd *fmtCommand, rootPath, selectionRoot string, exclude []string) ([]string, error) {
	// 1. Explicit --file flag
	if len(cmd.Files) > 0 {
		// Preserve --file's historical cwd/tilde expansion without allowing
		// Kong's path mapper to anchor root-relative flags before config loads.
		if selectionRoot == "" {
			selectionRoot = "."
		}
		paths := make([]string, len(cmd.Files))
		for i, path := range cmd.Files {
			if strings.HasPrefix(path, "~/") {
				path = kong.ExpandPath(path)
			}
			paths[i] = path
		}
		return resolveFmtSelectionPaths(paths, selectionRoot)
	}

	// 2. Check args for file paths
	var pathArgs []string
	var componentTypeArgs []string

	for _, arg := range cmd.Args {
		if format.CanFormatComponent(arg) {
			componentTypeArgs = append(componentTypeArgs, arg)
			continue
		}

		// Otherwise treat as file/directory path
		pathArgs = append(pathArgs, arg)
	}

	// 3. If we have path args, use them
	if len(pathArgs) > 0 {
		paths, err := resolveFmtSelectionPaths(pathArgs, selectionRoot)
		if err != nil {
			return nil, err
		}
		resolved, err := resolvePathArgs(paths)
		if err != nil {
			return nil, err
		}
		return resolved, nil
	}

	// 4. Discover the union of all requested component types. Discovery-based
	// selection honors the exclude config, matching lint discovery (explicit
	// --file/arg selections are deliberate and stay unfiltered).
	if len(componentTypeArgs) > 0 {
		return discoverFilesByType(rootPath, exclude, componentTypeArgs...)
	}

	// 5. No args: discover all component files
	return discoverAllFiles(rootPath, exclude)
}

// resolveFmtSelectionPaths applies only an explicit project root to selections.
// Absolute results prevent later validation from interpreting them against cwd.
func resolveFmtSelectionPaths(paths []string, rootPath string) ([]string, error) {
	if rootPath == "" {
		return paths, nil
	}
	resolved := make([]string, 0, len(paths))
	for _, path := range paths {
		if !filepath.IsAbs(path) {
			absolutePath, err := filepath.Abs(filepath.Join(rootPath, path))
			if err != nil {
				return nil, fmt.Errorf("cannot resolve %s: %w", path, err)
			}
			path = absolutePath
		}
		resolved = append(resolved, path)
	}
	return resolved, nil
}

// resolvePathArgs expands a list of file/directory paths into individual file paths.
func resolvePathArgs(paths []string) ([]string, error) {
	var files []string
	for _, path := range paths {
		// Stat the lexically-resolved absolute path so cwd-relative paths with
		// ".." segments survive symlinked-cwd divergence (see expandDirectoriesAtRoot).
		absolutePath, err := filepath.Abs(path)
		if err != nil {
			return nil, fmt.Errorf("cannot resolve %s: %w", path, err)
		}
		info, err := os.Stat(absolutePath)
		if err != nil {
			return nil, fmt.Errorf("cannot access %s: %w", path, err)
		}

		if !info.IsDir() {
			files = append(files, path)
			continue
		}

		dirFiles, err := discoverFilesInDir(absolutePath)
		if err != nil {
			return nil, err
		}
		files = append(files, dirFiles...)
	}
	return files, nil
}

// discoverFilesInDir finds all .md files in a directory.
func discoverFilesInDir(dirPath string) ([]string, error) {
	var files []string
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(path), ".md") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// discoverFilesByType selects requested component types in one discovery pass.
func discoverFilesByType(rootPath string, exclude []string, componentTypes ...string) ([]string, error) {
	targetTypes := make(map[discovery.FileType]bool, len(componentTypes))
	for _, componentType := range componentTypes {
		targetType, err := discovery.ParseFileType(componentType)
		if err != nil || !format.CanFormatComponent(componentType) {
			return nil, fmt.Errorf("unknown component type: %s", componentType)
		}
		targetTypes[targetType] = true
	}

	discoverer := discovery.NewFileDiscovery(rootPath, false).WithExclude(exclude)
	allFiles, err := discoverer.DiscoverFiles()
	if err != nil {
		return nil, err
	}

	var files []string
	for _, f := range allFiles {
		if targetTypes[f.Type] {
			files = append(files, f.Path)
		}
	}

	return files, nil
}

// discoverAllFiles discovers all component files.
func discoverAllFiles(rootPath string, exclude []string) ([]string, error) {
	discoverer := discovery.NewFileDiscovery(rootPath, false).WithExclude(exclude)
	allFiles, err := discoverer.DiscoverFiles()
	if err != nil {
		return nil, err
	}

	var files []string
	for _, f := range allFiles {
		// Only format markdown files
		if strings.HasSuffix(strings.ToLower(f.Path), ".md") {
			files = append(files, f.Path)
		}
	}

	return files, nil
}
