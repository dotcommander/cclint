package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

const (
	formatConsole  = "console"
	formatJSON     = "json"
	formatMarkdown = "markdown"
	failOnError    = "error"
	failOnWarning  = "warning"
	failOnSuggest  = "suggestion"
)

// Config represents the cclint configuration
type Config struct {
	Root             string       `mapstructure:"root"`
	RootExplicit     bool         `mapstructure:"-" json:"-"`
	Version          string       `mapstructure:"-"`
	Exclude          []string     `mapstructure:"exclude"`
	FollowSymlinks   bool         `mapstructure:"followSymlinks"`
	Format           string       `mapstructure:"format"`
	Output           string       `mapstructure:"output"`
	FailOn           string       `mapstructure:"failOn"`
	Quiet            bool         `mapstructure:"quiet"`
	Verbose          bool         `mapstructure:"verbose"`
	ShowScores       bool         `mapstructure:"showScores"`
	ShowImprovements bool         `mapstructure:"showImprovements"`
	NoCycleCheck     bool         `mapstructure:"no-cycle-check"`
	Rules            RulesConfig  `mapstructure:"rules"`
	Schemas          SchemaConfig `mapstructure:"schemas"`
	Concurrency      int          `mapstructure:"concurrency"`
	Parallel         bool         `mapstructure:"parallel"`
}

// RulesConfig contains rule configuration
type RulesConfig struct {
	Strict bool `mapstructure:"strict"`
}

// SchemaConfig contains schema configuration
type SchemaConfig struct {
	Enabled    bool           `mapstructure:"enabled"`
	Extensions map[string]any `mapstructure:"extensions"`
}

// LoadConfig loads configuration from various sources
func LoadConfig(rootPath string) (*Config, error) {
	homeDir, _ := os.UserHomeDir()
	vp := viper.New()
	setDefaults(vp, homeDir)

	// Explicit roots restrict lookup; otherwise use the closest ancestor config.
	configDir := rootPath
	if configDir == "" {
		configDir, _ = os.Getwd()
	}
	for {
		found := false
		for _, name := range []string{".cclintrc.json", ".cclintrc.yaml", ".cclintrc.yml"} {
			path := filepath.Join(configDir, name)
			if _, err := os.Stat(path); err != nil {
				continue
			}
			vp.SetConfigFile(path)
			if err := vp.ReadInConfig(); err != nil {
				continue
			}
			found = true
			break
		}
		parent := filepath.Dir(configDir)
		if found || rootPath != "" || parent == configDir {
			break
		}
		configDir = parent
	}

	// Keep historical env spellings first, then conventional snake-case aliases.
	vp.SetEnvPrefix("CCLINT")
	vp.AutomaticEnv()
	for key, alias := range map[string]string{
		"root": "ROOT", "exclude": "EXCLUDE", "output": "OUTPUT", "format": "FORMAT",
		"failOn": "FAIL_ON", "followSymlinks": "FOLLOW_SYMLINKS",
		"showScores": "SHOW_SCORES", "showImprovements": "SHOW_IMPROVEMENTS",
		"no-cycle-check": "NO_CYCLE_CHECK", "quiet": "QUIET", "verbose": "VERBOSE",
		"concurrency": "CONCURRENCY", "parallel": "PARALLEL",
		"rules.strict": "RULES_STRICT", "schemas.enabled": "SCHEMAS_ENABLED",
		"schemas.extensions": "SCHEMAS_EXTENSIONS",
	} {
		if err := vp.BindEnv(key, "CCLINT_"+strings.ToUpper(key), "CCLINT_"+alias); err != nil {
			return nil, err
		}
	}

	// Create config instance
	var config Config
	if err := vp.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Record configured roots so Git selection can distinguish them from defaults.
	config.RootExplicit = rootPath != "" || vp.InConfig("root") || os.Getenv("CCLINT_ROOT") != ""

	// Override root if provided
	if rootPath != "" {
		config.Root = rootPath
	}

	// Validate configuration
	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

func setDefaults(vp *viper.Viper, homeDir string) {
	vp.SetDefault("root", defaultRoot(homeDir))
	vp.SetDefault("format", formatConsole)
	vp.SetDefault("failOn", failOnError)
	vp.SetDefault("followSymlinks", false)
	vp.SetDefault("quiet", false)
	vp.SetDefault("verbose", false)
	vp.SetDefault("showScores", false)
	vp.SetDefault("showImprovements", false)
	vp.SetDefault("no-cycle-check", false)
	vp.SetDefault("concurrency", 10)
	vp.SetDefault("parallel", true)
	vp.SetDefault("rules.strict", true)
	vp.SetDefault("schemas.enabled", true)
}

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate format
	if config.Format != formatConsole && config.Format != formatJSON && config.Format != formatMarkdown {
		return fmt.Errorf("invalid format: %s. Must be 'console', 'json', or 'markdown'", config.Format)
	}

	// Validate failOn level
	if config.FailOn != failOnError && config.FailOn != failOnWarning && config.FailOn != failOnSuggest {
		return fmt.Errorf("invalid fail-on level: %s. Must be 'error', 'warning', or 'suggestion'", config.FailOn)
	}

	// Validate concurrency
	if config.Concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1")
	}

	// Note: --format json/markdown without --output writes to stdout,
	// which is a valid use case (e.g., piping to jq).

	return nil
}

// defaultRoot chooses the default project root for cclint.
//
// When invoked from inside a Claude Code project the expected root is that
// project, not ~/.claude. This walks up from the current working directory
// looking for Claude-specific markers (a plugin manifest or a .claude
// directory). On the first match, that directory becomes the default root.
// If no project is found we fall back to ~/.claude to preserve the
// historical behaviour when cclint is run from unrelated directories.
//
// A dedicated walk is used here (rather than project.FindProjectRoot) so we
// only auto-switch on Claude-specific markers — stumbling into an unrelated
// git repo or go.mod should not hijack cclint's default root.
func defaultRoot(homeDir string) string {
	fallback := filepath.Join(homeDir, ".claude")

	cwd, err := os.Getwd()
	if err != nil {
		return fallback
	}

	dir := cwd
	for {
		if _, err := os.Stat(filepath.Join(dir, ".claude-plugin", "plugin.json")); err == nil {
			return dir
		}
		if info, err := os.Stat(filepath.Join(dir, ".claude")); err == nil && info.IsDir() {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return fallback
		}
		dir = parent
	}
}

// SaveConfig saves the current configuration to a file
func SaveConfig(config *Config, path string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Marshal config to JSON
	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling config: %w", err)
	}

	// Write to file
	if err := os.WriteFile(path, jsonData, 0600); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}
