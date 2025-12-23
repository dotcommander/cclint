package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/outputters"
)

var (
	rootPath     string
	quiet        bool
	verbose      bool
	outputFormat string
	outputFile   string
	failOn       string
)

var rootCmd = &cobra.Command{
	Use:   "cclint",
	Short: "Claude Code Lint - A comprehensive linting tool for Claude Code projects",
	Long: `CCLint is a linting tool for Claude Code projects that validates agent files,
command files, settings, and documentation according to established patterns.

By default, cclint scans the entire project and reports on all validation issues.
Use specialized commands to focus on specific file types.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runLint(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVarP(&rootPath, "root", "r", "", "Project root directory (auto-detected if not specified)")
	rootCmd.PersistentFlags().BoolVarP(&quiet, "quiet", "q", false, "Suppress non-essential output")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose output")
	rootCmd.PersistentFlags().StringVarP(&outputFormat, "format", "f", "console", "Output format for reports (console|json|markdown)")
	rootCmd.PersistentFlags().StringVarP(&outputFile, "output", "o", "", "Output file for reports (requires --format)")
	rootCmd.PersistentFlags().StringVarP(&failOn, "fail-on", "", "error", "Fail build on specified level (error|warning|suggestion)")

	viper.BindPFlag("root", rootCmd.PersistentFlags().Lookup("root"))
	viper.BindPFlag("quiet", rootCmd.PersistentFlags().Lookup("quiet"))
	viper.BindPFlag("verbose", rootCmd.PersistentFlags().Lookup("verbose"))
	viper.BindPFlag("format", rootCmd.PersistentFlags().Lookup("format"))
	viper.BindPFlag("output", rootCmd.PersistentFlags().Lookup("output"))
	viper.BindPFlag("fail-on", rootCmd.PersistentFlags().Lookup("fail-on"))
}

func initConfig() {
	configPaths := []string{".cclintrc.json", ".cclintrc.yaml", ".cclintrc.yml"}
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			viper.SetConfigFile(path)
			if err := viper.ReadInConfig(); err != nil {
				fmt.Fprintf(os.Stderr, "Error reading config file: %v\n", err)
				os.Exit(1)
			}
			break
		}
	}
}

func runLint() error {
	// Load configuration
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	// Import specialized commands to register them
	_ = agentsCmd
	_ = commandsCmd
	_ = settingsCmd
	_ = contextCmd

	// For now, run all linters
	{
		summary, err := cli.LintAgents(cfg.Root, cfg.Quiet, cfg.Verbose)
		if err != nil {
			return fmt.Errorf("error running agents linter: %w", err)
		}

		// Format and output results
		outputter := outputters.NewOutputter(cfg)
		if err := outputter.Format(summary, cfg.Format); err != nil {
			return fmt.Errorf("error formatting output: %w", err)
		}
	}

	return nil
}

func validatePath(path string) error {
	// TODO: Implement path validation to ensure it's safe
	return nil
}