package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/viper"
)

// Config represents the cclint configuration
type Config struct {
	Root          string            `mapstructure:"root"`
	Exclude       []string          `mapstructure:"exclude"`
	FollowSymlinks bool              `mapstructure:"followSymlinks"`
	Format        string            `mapstructure:"format"`
	Output        string            `mapstructure:"output"`
	FailOn        string            `mapstructure:"failOn"`
	Quiet         bool              `mapstructure:"quiet"`
	Verbose       bool              `mapstructure:"verbose"`
	Rules         RulesConfig       `mapstructure:"rules"`
	Schemas       SchemaConfig      `mapstructure:"schemas"`
	Concurrency   int               `mapstructure:"concurrency"`
	Parallel      bool              `mapstructure:"parallel"`
}

// RulesConfig contains rule configuration
type RulesConfig struct {
	Strict bool `mapstructure:"strict"`
}

// SchemaConfig contains schema configuration
type SchemaConfig struct {
	Enabled       bool                   `mapstructure:"enabled"`
	Extensions   map[string]interface{} `mapstructure:"extensions"`
}

// LoadConfig loads configuration from various sources
func LoadConfig(rootPath string) (*Config, error) {
	// Set default values
	homeDir, _ := os.UserHomeDir()
	viper.SetDefault("root", filepath.Join(homeDir, ".claude"))
	viper.SetDefault("format", "console")
	viper.SetDefault("failOn", "error")
	viper.SetDefault("followSymlinks", false)
	viper.SetDefault("quiet", false)
	viper.SetDefault("verbose", false)
	viper.SetDefault("concurrency", 10)
	viper.SetDefault("parallel", true)
	viper.SetDefault("rules.strict", true)
	viper.SetDefault("schemas.enabled", true)

	// Config file locations
	configPaths := []string{".cclintrc.json", ".cclintrc.yaml", ".cclintrc.yml"}
	for _, path := range configPaths {
		viper.SetConfigFile(path)
		if err := viper.ReadInConfig(); err == nil {
			break
		}
	}

	// Environment variables
	viper.SetEnvPrefix("CCLINT")
	viper.AutomaticEnv()

	// Create config instance
	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

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

// validateConfig validates the configuration
func validateConfig(config *Config) error {
	// Validate format
	if config.Format != "console" && config.Format != "json" && config.Format != "markdown" {
		return fmt.Errorf("invalid format: %s. Must be 'console', 'json', or 'markdown'", config.Format)
	}

	// Validate failOn level
	if config.FailOn != "error" && config.FailOn != "warning" && config.FailOn != "suggestion" {
		return fmt.Errorf("invalid fail-on level: %s. Must be 'error', 'warning', or 'suggestion'", config.FailOn)
	}

	// Validate concurrency
	if config.Concurrency < 1 {
		return fmt.Errorf("concurrency must be at least 1")
	}

	// Validate output file if format is not console
	if config.Format != "console" && config.Output == "" {
		return fmt.Errorf("output file is required when format is not 'console'")
	}

	return nil
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
	if err := os.WriteFile(path, jsonData, 0644); err != nil {
		return fmt.Errorf("error writing config file: %w", err)
	}

	return nil
}