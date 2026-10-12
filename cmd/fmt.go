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
	Files []string `name:"file" help:"Explicit file path to format (relative to an explicit project root, otherwise cwd)."`
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

	// Discovery and type detection must use the same absolute root as selections.
	cfg.Root, err = filepath.Abs(cfg.Root)
	if err != nil {
		return fmt.Errorf("cannot resolve project root: %w", err)
	}

	// Determine which files to format
	selectionRoot := ""
	if cfg.RootExplicit {
		selectionRoot = cfg.Root
	}
	filesToFormat, err := collectFilesToFormat(cmd, cfg.Root, selectionRoot, cfg.Exclude)
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
		return false, fmt.Errorf("cannot format %s: %w", filePath, err)
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
		return false, fmt.Errorf("error reading %s: %w", filePath, err)
	}

	formatter := format.NewComponentFormatter(fileType.String())
	formatted, err := formatter.Format(string(content))
	if err != nil {
		return false, fmt.Errorf("error formatting %s: %w", filePath, err)
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
		if err := replaceFormattedFile(absPath, formatted, os.Rename); err != nil {
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

// replaceFormattedFile prepares the complete replacement beside the original.
// All fallible writes finish before rename, so a failed operation preserves the
// original bytes. Resolve links first to preserve an explicitly selected link.
func replaceFormattedFile(path, content string, rename func(string, string) error) error {
	target, err := filepath.EvalSymlinks(path)
	if err != nil {
		return err
	}
	info, err := os.Stat(target)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("cannot replace non-regular file %s", path)
	}
	temporary, err := os.CreateTemp(filepath.Dir(target), ".cclint-fmt-*")
	if err != nil {
		return err
	}
	defer os.Remove(temporary.Name())
	defer temporary.Close()
	if _, err := temporary.WriteString(content); err != nil {
		return err
	}
	if err := temporary.Chmod(info.Mode()); err != nil {
		return err
	}
	if err := temporary.Sync(); err != nil {
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	return rename(temporary.Name(), target)
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
