package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/discovery"
	"github.com/dotcommander/cclint/internal/format"
	"github.com/spf13/cobra"
)

var (
	fmtCheck  bool
	fmtWrite  bool
	fmtDiff   bool
	fmtFiles  []string // Explicit file paths
	fmtType   string   // Force component type
)

var fmtCmd = &cobra.Command{
	Use:   "fmt [files...]",
	Short: "Format Claude Code component files canonically",
	Long: `Format Claude Code component files with canonical style.

FORMATTING RULES:

  Frontmatter:
  - Normalize field order: name, description, model, tools/allowed-tools, then alphabetical
  - Ensure exactly one blank line after frontmatter

  Markdown:
  - Trim trailing whitespace from lines
  - Ensure file ends with exactly one newline

USAGE MODES:

  Format all components (preview):
    cclint fmt                    # Print formatted to stdout
    cclint fmt agents             # Format only agents
    cclint fmt --write            # Write changes in place

  Format specific files:
    cclint fmt file.md            # Preview formatting
    cclint fmt -w file.md         # Write changes
    cclint fmt --diff file.md     # Show diff

  CI mode:
    cclint fmt --check            # Exit 1 if files need formatting

FLAGS:
  --check      Exit 1 if files would change (for CI)
  -w, --write  Write changes in place
  --diff       Show diff of what would change

EXAMPLES:

  # Preview formatting
  cclint fmt ./agents/my-agent.md

  # Format in place
  cclint fmt -w ./agents/my-agent.md

  # Check if formatting needed (CI)
  cclint fmt --check agents/

  # Format all components
  cclint fmt --write`,
	Args: cobra.ArbitraryArgs,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runFmt(args); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(fmtCmd)

	fmtCmd.Flags().BoolVar(&fmtCheck, "check", false, "Exit 1 if files would change (for CI)")
	fmtCmd.Flags().BoolVarP(&fmtWrite, "write", "w", false, "Write changes in place")
	fmtCmd.Flags().BoolVar(&fmtDiff, "diff", false, "Show diff of what would change")
	fmtCmd.Flags().StringArrayVar(&fmtFiles, "file", nil, "Explicit file path(s) to format")
	fmtCmd.Flags().StringVarP(&fmtType, "type", "t", "", "Force component type (agent|command|skill)")
}

func runFmt(args []string) error {
	// Load configuration
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Determine which files to format
	filesToFormat, err := collectFilesToFormat(args, cfg.Root)
	if err != nil {
		return err
	}

	if len(filesToFormat) == 0 {
		return fmt.Errorf("no files to format")
	}

	// Track if any files need changes
	var needsFormatting []string
	var totalFiles int

	// Format each file
	for _, filePath := range filesToFormat {
		totalFiles++

		// Validate file exists
		absPath, err := discovery.ValidateFilePath(filePath)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", filePath, err)
			}
			continue
		}

		// Detect component type
		var fileType discovery.FileType
		if fmtType != "" {
			fileType, err = discovery.ParseFileType(fmtType)
			if err != nil {
				return err
			}
		} else {
			fileType, err = discovery.DetectFileType(absPath, cfg.Root)
			if err != nil {
				if !quiet {
					fmt.Fprintf(os.Stderr, "Skipping %s: %v\n", filePath, err)
				}
				continue
			}
		}

		// Only format markdown files
		if !strings.HasSuffix(strings.ToLower(absPath), ".md") {
			if verbose {
				fmt.Fprintf(os.Stderr, "Skipping %s: not a markdown file\n", filePath)
			}
			continue
		}

		// Read file
		content, err := os.ReadFile(absPath)
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Error reading %s: %v\n", filePath, err)
			}
			continue
		}

		// Create formatter for component type
		formatter := format.NewComponentFormatter(fileType.String())

		// Format content
		formatted, err := formatter.Format(string(content))
		if err != nil {
			if !quiet {
				fmt.Fprintf(os.Stderr, "Error formatting %s: %v\n", filePath, err)
			}
			continue
		}

		// Check if content changed
		if string(content) != formatted {
			needsFormatting = append(needsFormatting, absPath)

			// Handle different output modes
			if fmtCheck {
				// Check mode: just track, will exit 1 later
				if !quiet {
					fmt.Printf("%s needs formatting\n", filePath)
				}
			} else if fmtDiff {
				// Diff mode: show differences
				diff := format.Diff(string(content), formatted, filePath)
				fmt.Print(diff)
			} else if fmtWrite {
				// Write mode: save changes
				if err := os.WriteFile(absPath, []byte(formatted), 0644); err != nil {
					return fmt.Errorf("error writing %s: %w", absPath, err)
				}
				if !quiet {
					fmt.Printf("Formatted %s\n", filePath)
				}
			} else {
				// Default mode: print to stdout
				fmt.Print(formatted)
			}
		} else {
			// File already formatted
			if verbose {
				fmt.Printf("%s already formatted\n", filePath)
			}
		}
	}

	// Summary
	if !quiet && totalFiles > 1 {
		if len(needsFormatting) == 0 {
			fmt.Printf("\nAll %d files already formatted\n", totalFiles)
		} else {
			if fmtWrite {
				fmt.Printf("\nFormatted %d of %d files\n", len(needsFormatting), totalFiles)
			} else {
				fmt.Printf("\n%d of %d files need formatting\n", len(needsFormatting), totalFiles)
			}
		}
	}

	// Check mode: exit 1 if files need formatting
	if fmtCheck && len(needsFormatting) > 0 {
		exitFunc(1)
	}

	return nil
}

// collectFilesToFormat determines which files to format based on args and flags.
func collectFilesToFormat(args []string, rootPath string) ([]string, error) {
	var files []string

	// 1. Explicit --file flag
	if len(fmtFiles) > 0 {
		return fmtFiles, nil
	}

	// 2. Check args for file paths
	var pathArgs []string
	var componentTypeArg string

	for _, arg := range args {
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
		for _, path := range pathArgs {
			// Check if it's a directory or file
			info, err := os.Stat(path)
			if err != nil {
				return nil, fmt.Errorf("cannot access %s: %w", path, err)
			}

			if info.IsDir() {
				// Discover files in directory
				dirFiles, err := discoverFilesInDir(path)
				if err != nil {
					return nil, err
				}
				files = append(files, dirFiles...)
			} else {
				// Single file
				files = append(files, path)
			}
		}
		return files, nil
	}

	// 4. If we have component type arg, discover those files
	if componentTypeArg != "" {
		return discoverFilesByType(rootPath, componentTypeArg)
	}

	// 5. No args: discover all component files
	return discoverAllFiles(rootPath)
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
