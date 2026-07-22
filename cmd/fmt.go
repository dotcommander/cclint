package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/format"
)

type fmtCommand struct {
	Check bool     `help:"Exit 1 if files would change."`
	Write bool     `short:"w" help:"Write changes in place."`
	Diff  bool     `help:"Show diff of what would change."`
	Files []string `name:"file" type:"path" help:"Explicit file path to format."`
	Type  string   `short:"t" help:"Force component type."`
	Args  []string `arg:"" optional:"" name:"files"`
}

func (cmd *fmtCommand) Run(opts executionOptions) error { return runFmt(opts, cmd) }

func runFmt(opts executionOptions, cmd *fmtCommand) error {
	// Load configuration
	cfg, err := config.LoadConfig(stringValue(opts.root))
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Determine which files to format
	filesToFormat, err := collectFilesToFormat(cmd, cfg.Root)
	if err != nil {
		return err
	}

	if len(filesToFormat) == 0 {
		return fmt.Errorf("no files to format")
	}

	// Format each file, collecting results
	var needsFormatting []string
	totalFiles := len(filesToFormat)

	for _, filePath := range filesToFormat {
		changed, fmtErr := formatOneFile(opts, cmd, filePath, cfg.Root)
		if fmtErr != nil {
			return fmtErr
		}
		if changed {
			needsFormatting = append(needsFormatting, filePath)
		}
	}

	printFmtSummary(opts, cmd, totalFiles, len(needsFormatting))

	// Check mode: exit 1 if files need formatting
	if cmd.Check && len(needsFormatting) > 0 {
		exitFunc(1)
	}

	return nil
}

// formatOneFile validates, reads, formats, and outputs a single file.
// Returns true if the file needed formatting, or an error for fatal failures.
func formatOneFile(opts executionOptions, cmd *fmtCommand, filePath, root string) (bool, error) {
	absPath, err := discovery.ValidateFilePath(filePath)
	if err != nil {
		if !boolValue(opts.quiet) {
			fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", filePath, err)
		}
		return false, nil
	}

	fileType, skip, err := resolveFileType(opts, cmd, absPath, filePath, root)
	if err != nil {
		return false, err
	}
	if skip {
		return false, nil
	}

	if !strings.HasSuffix(strings.ToLower(absPath), ".md") {
		if boolValue(opts.verbose) {
			fmt.Fprintf(os.Stderr, "Skipping %s: not a markdown file\n", filePath)
		}
		return false, nil
	}

	content, err := os.ReadFile(absPath)
	if err != nil {
		if !boolValue(opts.quiet) {
			fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filePath, err)
		}
		return false, nil
	}

	formatter := format.NewComponentFormatter(fileType.String())
	formatted, err := formatter.Format(string(content))
	if err != nil {
		if !boolValue(opts.quiet) {
			fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", filePath, err)
		}
		return false, nil
	}

	if string(content) == formatted {
		if boolValue(opts.verbose) {
			fmt.Printf("%s already formatted\n", filePath)
		}
		return false, nil
	}

	return true, emitFormatted(opts, cmd, absPath, filePath, string(content), formatted)
}

// resolveFileType determines the component type for a file. If the type cannot
// be resolved (and is not a fatal error), skip is returned as true.
func resolveFileType(opts executionOptions, cmd *fmtCommand, absPath, displayPath, root string) (discovery.FileType, bool, error) {
	if cmd.Type != "" {
		ft, err := discovery.ParseFileType(cmd.Type)
		return ft, false, err
	}

	ft, err := discovery.DetectFileType(absPath, root)
	if err != nil {
		if !boolValue(opts.quiet) {
			fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", displayPath, err)
		}
		return 0, true, nil
	}
	return ft, false, nil
}

// emitFormatted writes or displays the formatted output based on the active mode.
func emitFormatted(opts executionOptions, cmd *fmtCommand, absPath, displayPath, original, formatted string) error {
	switch {
	case cmd.Check:
		if !boolValue(opts.quiet) {
			fmt.Printf("%s needs formatting\n", displayPath)
		}
	case cmd.Diff:
		fmt.Print(format.Diff(original, formatted, displayPath))
	case cmd.Write:
		if err := os.WriteFile(absPath, []byte(formatted), 0600); err != nil {
			return fmt.Errorf("error writing %s: %w", absPath, err)
		}
		if !boolValue(opts.quiet) {
			fmt.Printf("Formatted %s\n", displayPath)
		}
	default:
		fmt.Print(formatted)
	}
	return nil
}

// printFmtSummary prints the formatting summary when multiple files were processed.
func printFmtSummary(opts executionOptions, cmd *fmtCommand, totalFiles, changedCount int) {
	if boolValue(opts.quiet) || totalFiles <= 1 {
		return
	}

	if changedCount == 0 {
		fmt.Printf("\nAll %d files already formatted\n", totalFiles)
		return
	}

	if cmd.Write {
		fmt.Printf("\nFormatted %d of %d files\n", changedCount, totalFiles)
	} else {
		fmt.Printf("\n%d of %d files need formatting\n", changedCount, totalFiles)
	}
}

// collectFilesToFormat determines which files to format based on args and flags.
func collectFilesToFormat(cmd *fmtCommand, rootPath string) ([]string, error) {
	// 1. Explicit --file flag
	if len(cmd.Files) > 0 {
		return cmd.Files, nil
	}

	// 2. Check args for file paths
	var pathArgs []string
	var componentTypeArg string

	for _, arg := range cmd.Args {
		// Check if it's a component type (agents, commands, skills)
		if isComponentType(arg) {
			componentTypeArg = arg
			continue
		}

		// Otherwise treat as file/directory path
		pathArgs = append(pathArgs, arg)
	}

	// 3. If we have path args, use them
	if len(pathArgs) > 0 {
		resolved, err := resolvePathArgs(pathArgs)
		if err != nil {
			return nil, err
		}
		return resolved, nil
	}

	// 4. If we have component type arg, discover those files
	if componentTypeArg != "" {
		return discoverFilesByType(rootPath, componentTypeArg)
	}

	// 5. No args: discover all component files
	return discoverAllFiles(rootPath)
}

// resolvePathArgs expands a list of file/directory paths into individual file paths.
func resolvePathArgs(paths []string) ([]string, error) {
	var files []string
	for _, path := range paths {
		info, err := os.Stat(path)
		if err != nil {
			return nil, fmt.Errorf("cannot access %s: %w", path, err)
		}

		if !info.IsDir() {
			files = append(files, path)
			continue
		}

		dirFiles, err := discoverFilesInDir(path)
		if err != nil {
			return nil, err
		}
		files = append(files, dirFiles...)
	}
	return files, nil
}

// isComponentType checks if arg is a component type name.
func isComponentType(arg string) bool {
	types := map[string]bool{
		"agents":   true,
		"commands": true,
		"skills":   true,
		"settings": true,
		"context":  true,
		"plugins":  true,
		"rules":    true,
	}
	return types[strings.ToLower(arg)]
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

// discoverFilesByType discovers files of a specific component type.
func discoverFilesByType(rootPath, componentType string) ([]string, error) {
	discoverer := discovery.NewFileDiscovery(rootPath, false)
	allFiles, err := discoverer.DiscoverFiles()
	if err != nil {
		return nil, err
	}

	var files []string
	var targetType discovery.FileType

	switch componentType {
	case "agents":
		targetType = discovery.FileTypeAgent
	case "commands":
		targetType = discovery.FileTypeCommand
	case "skills":
		targetType = discovery.FileTypeSkill
	case "settings":
		targetType = discovery.FileTypeSettings
	case "context":
		targetType = discovery.FileTypeContext
	case "plugins":
		targetType = discovery.FileTypePlugin
	case "rules":
		targetType = discovery.FileTypeRule
	default:
		return nil, fmt.Errorf("unknown component type: %s", componentType)
	}

	for _, f := range allFiles {
		if f.Type == targetType {
			files = append(files, f.Path)
		}
	}

	return files, nil
}

// discoverAllFiles discovers all component files.
func discoverAllFiles(rootPath string) ([]string, error) {
	discoverer := discovery.NewFileDiscovery(rootPath, false)
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
