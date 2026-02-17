package cmd

import (
	"fmt"
	"os"

	"github.com/dotcommander/cclint/internal/cli"
	"github.com/spf13/cobra"
)

var pluginsCmd = &cobra.Command{
	Use:   "plugins",
	Short: "Lint only plugin manifest files",
	Long: `The plugins command scans for and validates plugin manifest files in your project.

Supported file patterns:
- **/.claude-plugin/plugin.json

Validation checks:
- Required fields: name, description, version, author.name
- Version format: semver (major.minor.patch)
- Plugin structure: plugin.json in .claude-plugin directory`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runPluginsLint(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			exitFunc(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(pluginsCmd)
}

func runPluginsLint() error {
	return runComponentLint("plugins", cli.LintPlugins)
}
