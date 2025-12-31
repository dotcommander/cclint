package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// resetViper resets viper to a clean state for each test
func resetViper() {
	viper.Reset()
}

// setupTestDir creates a temporary directory for testing
func setupTestDir(t *testing.T) string {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "cclint-config-test-*")
	require.NoError(t, err)
	t.Cleanup(func() {
		os.RemoveAll(tmpDir)
	})
	return tmpDir
}

// TestLoadConfigDefaults tests that default values are set correctly
func TestLoadConfigDefaults(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Change to temp dir so no config files are found
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	config, err := LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify default values
	homeDir, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(homeDir, ".claude"), config.Root)
	assert.Equal(t, "console", config.Format)
	assert.Equal(t, "error", config.FailOn)
	assert.False(t, config.FollowSymlinks)
	assert.False(t, config.Quiet)
	assert.False(t, config.Verbose)
	assert.False(t, config.ShowScores)
	assert.False(t, config.ShowImprovements)
	assert.False(t, config.NoCycleCheck)
	assert.Equal(t, 10, config.Concurrency)
	assert.True(t, config.Parallel)
	assert.True(t, config.Rules.Strict)
	assert.True(t, config.Schemas.Enabled)
}

// TestLoadConfigFromJSON tests loading configuration from JSON file
func TestLoadConfigFromJSON(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Create JSON config file
	configData := map[string]interface{}{
		"root":             "/custom/root",
		"exclude":          []string{"node_modules", "*.tmp"},
		"followSymlinks":   true,
		"format":           "json",
		"output":           "report.json",
		"failOn":           "warning",
		"quiet":            true,
		"verbose":          false,
		"showScores":       true,
		"showImprovements": true,
		"no-cycle-check":   true,
		"concurrency":      20,
		"parallel":         false,
		"rules": map[string]interface{}{
			"strict": false,
		},
		"schemas": map[string]interface{}{
			"enabled": false,
			"extensions": map[string]interface{}{
				"custom": "value",
			},
		},
	}

	configPath := filepath.Join(tmpDir, ".cclintrc.json")
	jsonData, err := json.MarshalIndent(configData, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, jsonData, 0644))

	// Change to temp dir
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	config, err := LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify values from JSON
	assert.Equal(t, "/custom/root", config.Root)
	assert.Equal(t, []string{"node_modules", "*.tmp"}, config.Exclude)
	assert.True(t, config.FollowSymlinks)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, "report.json", config.Output)
	assert.Equal(t, "warning", config.FailOn)
	assert.True(t, config.Quiet)
	assert.False(t, config.Verbose)
	assert.True(t, config.ShowScores)
	assert.True(t, config.ShowImprovements)
	assert.True(t, config.NoCycleCheck)
	assert.Equal(t, 20, config.Concurrency)
	assert.False(t, config.Parallel)
	assert.False(t, config.Rules.Strict)
	assert.False(t, config.Schemas.Enabled)
	assert.Equal(t, "value", config.Schemas.Extensions["custom"])
}

// TestLoadConfigFromYAML tests loading configuration from YAML file
func TestLoadConfigFromYAML(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Create YAML config file
	yamlContent := `
root: /yaml/root
exclude:
  - dist
  - build
followSymlinks: true
format: markdown
output: report.md
failOn: suggestion
quiet: false
verbose: true
showScores: true
showImprovements: false
no-cycle-check: false
concurrency: 15
parallel: true
rules:
  strict: true
schemas:
  enabled: true
  extensions:
    yaml: test
`

	configPath := filepath.Join(tmpDir, ".cclintrc.yaml")
	require.NoError(t, os.WriteFile(configPath, []byte(yamlContent), 0644))

	// Change to temp dir
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	config, err := LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, config)

	// Verify values from YAML
	assert.Equal(t, "/yaml/root", config.Root)
	assert.Equal(t, []string{"dist", "build"}, config.Exclude)
	assert.True(t, config.FollowSymlinks)
	assert.Equal(t, "markdown", config.Format)
	assert.Equal(t, "report.md", config.Output)
	assert.Equal(t, "suggestion", config.FailOn)
	assert.False(t, config.Quiet)
	assert.True(t, config.Verbose)
	assert.True(t, config.ShowScores)
	assert.False(t, config.ShowImprovements)
	assert.False(t, config.NoCycleCheck)
	assert.Equal(t, 15, config.Concurrency)
	assert.True(t, config.Parallel)
	assert.True(t, config.Rules.Strict)
	assert.True(t, config.Schemas.Enabled)
	assert.Equal(t, "test", config.Schemas.Extensions["yaml"])
}

// TestLoadConfigYMLExtension tests .yml extension
func TestLoadConfigYMLExtension(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	yamlContent := `
root: /yml/root
format: json
output: report.json
`

	configPath := filepath.Join(tmpDir, ".cclintrc.yml")
	require.NoError(t, os.WriteFile(configPath, []byte(yamlContent), 0644))

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	config, err := LoadConfig("")
	require.NoError(t, err)
	assert.Equal(t, "/yml/root", config.Root)
	assert.Equal(t, "json", config.Format)
}

// TestLoadConfigRootPathOverride tests that provided rootPath overrides config
func TestLoadConfigRootPathOverride(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Create config with root set
	configData := map[string]interface{}{
		"root": "/config/root",
	}

	configPath := filepath.Join(tmpDir, ".cclintrc.json")
	jsonData, err := json.MarshalIndent(configData, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(configPath, jsonData, 0644))

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	// Load with override
	config, err := LoadConfig("/override/root")
	require.NoError(t, err)

	// Override should take precedence
	assert.Equal(t, "/override/root", config.Root)
}

// TestLoadConfigEnvironmentVariables tests environment variable overrides
func TestLoadConfigEnvironmentVariables(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Set environment variables
	// Note: Viper's AutomaticEnv() requires BindEnv for nested keys or SetEnvKeyReplacer
	envVars := map[string]string{
		"CCLINT_ROOT":        "/env/root",
		"CCLINT_FORMAT":      "console", // Use console to avoid output file requirement
		"CCLINT_FAILON":      "warning",
		"CCLINT_QUIET":       "true",
		"CCLINT_VERBOSE":     "true",
		"CCLINT_CONCURRENCY": "30",
		"CCLINT_PARALLEL":    "false",
	}

	for key, value := range envVars {
		t.Setenv(key, value)
	}

	oldWd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(tmpDir))
	t.Cleanup(func() {
		_ = os.Chdir(oldWd)
	})

	config, err := LoadConfig("")
	require.NoError(t, err)

	// Verify environment variables were applied
	assert.Equal(t, "/env/root", config.Root)
	assert.Equal(t, "console", config.Format)
	assert.Equal(t, "warning", config.FailOn)
	assert.True(t, config.Quiet)
	assert.True(t, config.Verbose)
	assert.Equal(t, 30, config.Concurrency)
	assert.False(t, config.Parallel)
}

// TestLoadConfigConfigFilePriority tests that first found config file is used
func TestLoadConfigConfigFilePriority(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Create multiple config files
	jsonConfig := map[string]interface{}{"root": "/json/root"}
	jsonData, _ := json.MarshalIndent(jsonConfig, "", "  ")
	_ = os.WriteFile(filepath.Join(tmpDir, ".cclintrc.json"), jsonData, 0644)

	yamlContent := "root: /yaml/root\n"
	_ = os.WriteFile(filepath.Join(tmpDir, ".cclintrc.yaml"), []byte(yamlContent), 0644)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	config, err := LoadConfig("")
	require.NoError(t, err)

	// .cclintrc.json should be loaded first
	assert.Equal(t, "/json/root", config.Root)
}

// TestValidateConfigInvalidFormat tests format validation
func TestValidateConfigInvalidFormat(t *testing.T) {
	config := &Config{
		Format: "invalid",
		FailOn: "error",
		Concurrency: 10,
	}

	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid format")
}

// TestValidateConfigInvalidFailOn tests failOn validation
func TestValidateConfigInvalidFailOn(t *testing.T) {
	config := &Config{
		Format: "console",
		FailOn: "invalid",
		Concurrency: 10,
	}

	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid fail-on level")
}

// TestValidateConfigInvalidConcurrency tests concurrency validation
func TestValidateConfigInvalidConcurrency(t *testing.T) {
	config := &Config{
		Format: "console",
		FailOn: "error",
		Concurrency: 0,
	}

	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "concurrency must be at least 1")
}

// TestValidateConfigMissingOutput tests output file requirement
func TestValidateConfigMissingOutput(t *testing.T) {
	config := &Config{
		Format: "json",
		FailOn: "error",
		Output: "",
		Concurrency: 10,
	}

	err := validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "output file is required")
}

// TestValidateConfigValid tests valid configuration
func TestValidateConfigValid(t *testing.T) {
	tests := []struct {
		name   string
		config *Config
	}{
		{
			name: "console format without output",
			config: &Config{
				Format: "console",
				FailOn: "error",
				Concurrency: 10,
			},
		},
		{
			name: "json format with output",
			config: &Config{
				Format: "json",
				Output: "report.json",
				FailOn: "warning",
				Concurrency: 5,
			},
		},
		{
			name: "markdown format with output",
			config: &Config{
				Format: "markdown",
				Output: "report.md",
				FailOn: "suggestion",
				Concurrency: 20,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateConfig(tt.config)
			assert.NoError(t, err)
		})
	}
}

// TestSaveConfig tests saving configuration to file
func TestSaveConfig(t *testing.T) {
	tmpDir := setupTestDir(t)

	config := &Config{
		Root:             "/test/root",
		Exclude:          []string{"*.tmp", "node_modules"},
		FollowSymlinks:   true,
		Format:           "json",
		Output:           "output.json",
		FailOn:           "warning",
		Quiet:            true,
		Verbose:          false,
		ShowScores:       true,
		ShowImprovements: false,
		NoCycleCheck:     true,
		Concurrency:      15,
		Parallel:         false,
		Rules: RulesConfig{
			Strict: false,
		},
		Schemas: SchemaConfig{
			Enabled: true,
			Extensions: map[string]interface{}{
				"test": "value",
			},
		},
	}

	savePath := filepath.Join(tmpDir, "config", "saved.json")
	err := SaveConfig(config, savePath)
	require.NoError(t, err)

	// Verify file was created
	assert.FileExists(t, savePath)

	// Read and verify content
	data, err := os.ReadFile(savePath)
	require.NoError(t, err)

	var loaded Config
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	// Verify all fields
	assert.Equal(t, config.Root, loaded.Root)
	assert.Equal(t, config.Exclude, loaded.Exclude)
	assert.Equal(t, config.FollowSymlinks, loaded.FollowSymlinks)
	assert.Equal(t, config.Format, loaded.Format)
	assert.Equal(t, config.Output, loaded.Output)
	assert.Equal(t, config.FailOn, loaded.FailOn)
	assert.Equal(t, config.Quiet, loaded.Quiet)
	assert.Equal(t, config.Verbose, loaded.Verbose)
	assert.Equal(t, config.ShowScores, loaded.ShowScores)
	assert.Equal(t, config.ShowImprovements, loaded.ShowImprovements)
	assert.Equal(t, config.NoCycleCheck, loaded.NoCycleCheck)
	assert.Equal(t, config.Concurrency, loaded.Concurrency)
	assert.Equal(t, config.Parallel, loaded.Parallel)
	assert.Equal(t, config.Rules.Strict, loaded.Rules.Strict)
	assert.Equal(t, config.Schemas.Enabled, loaded.Schemas.Enabled)
}

// TestSaveConfigCreatesDirectory tests that SaveConfig creates parent directories
func TestSaveConfigCreatesDirectory(t *testing.T) {
	tmpDir := setupTestDir(t)

	config := &Config{
		Format: "console",
		FailOn: "error",
		Concurrency: 10,
	}

	// Deep nested path that doesn't exist
	savePath := filepath.Join(tmpDir, "deep", "nested", "path", "config.json")
	err := SaveConfig(config, savePath)
	require.NoError(t, err)

	assert.FileExists(t, savePath)
}

// TestSaveConfigInvalidPath tests error handling for invalid paths
func TestSaveConfigInvalidPath(t *testing.T) {
	config := &Config{
		Format: "console",
		FailOn: "error",
		Concurrency: 10,
	}

	// Try to save to an invalid path (file as directory)
	tmpDir := setupTestDir(t)
	filePath := filepath.Join(tmpDir, "file")
	_ = os.WriteFile(filePath, []byte("test"), 0644)

	invalidPath := filepath.Join(filePath, "config.json")
	err := SaveConfig(config, invalidPath)
	assert.Error(t, err)
}

// TestLoadConfigUnmarshalError tests unmarshal error handling
func TestLoadConfigUnmarshalError(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Create invalid JSON config
	configPath := filepath.Join(tmpDir, ".cclintrc.json")
	invalidJSON := `{"concurrency": "not-a-number"}`
	require.NoError(t, os.WriteFile(configPath, []byte(invalidJSON), 0644))

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	config, err := LoadConfig("")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "error unmarshaling config")
}

// TestLoadConfigValidationError tests that LoadConfig returns validation errors
func TestLoadConfigValidationError(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Create config that will fail validation (invalid format)
	configData := map[string]interface{}{
		"format": "invalid-format",
	}

	configPath := filepath.Join(tmpDir, ".cclintrc.json")
	jsonData, _ := json.MarshalIndent(configData, "", "  ")
	_ = os.WriteFile(configPath, jsonData, 0644)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	config, err := LoadConfig("")
	assert.Error(t, err)
	assert.Nil(t, config)
	assert.Contains(t, err.Error(), "invalid configuration")
	assert.Contains(t, err.Error(), "invalid format")
}

// TestConfigStructFields tests that all Config struct fields are accessible
func TestConfigStructFields(t *testing.T) {
	config := Config{
		Root:             "/test",
		Exclude:          []string{"test"},
		FollowSymlinks:   true,
		Format:           "json",
		Output:           "out.json",
		FailOn:           "error",
		Quiet:            true,
		Verbose:          true,
		ShowScores:       true,
		ShowImprovements: true,
		NoCycleCheck:     true,
		Concurrency:      5,
		Parallel:         true,
		Rules: RulesConfig{
			Strict: true,
		},
		Schemas: SchemaConfig{
			Enabled: true,
			Extensions: map[string]interface{}{
				"key": "value",
			},
		},
	}

	// Verify all fields are accessible
	assert.Equal(t, "/test", config.Root)
	assert.Equal(t, []string{"test"}, config.Exclude)
	assert.True(t, config.FollowSymlinks)
	assert.Equal(t, "json", config.Format)
	assert.Equal(t, "out.json", config.Output)
	assert.Equal(t, "error", config.FailOn)
	assert.True(t, config.Quiet)
	assert.True(t, config.Verbose)
	assert.True(t, config.ShowScores)
	assert.True(t, config.ShowImprovements)
	assert.True(t, config.NoCycleCheck)
	assert.Equal(t, 5, config.Concurrency)
	assert.True(t, config.Parallel)
	assert.True(t, config.Rules.Strict)
	assert.True(t, config.Schemas.Enabled)
	assert.Equal(t, "value", config.Schemas.Extensions["key"])
}

// TestRulesConfig tests RulesConfig struct
func TestRulesConfig(t *testing.T) {
	rules := RulesConfig{
		Strict: true,
	}
	assert.True(t, rules.Strict)

	rules.Strict = false
	assert.False(t, rules.Strict)
}

// TestSchemaConfig tests SchemaConfig struct
func TestSchemaConfig(t *testing.T) {
	schemas := SchemaConfig{
		Enabled: true,
		Extensions: map[string]interface{}{
			"ext1": "value1",
			"ext2": 123,
			"ext3": true,
		},
	}

	assert.True(t, schemas.Enabled)
	assert.Equal(t, "value1", schemas.Extensions["ext1"])
	assert.Equal(t, 123, schemas.Extensions["ext2"])
	assert.Equal(t, true, schemas.Extensions["ext3"])
}

// TestLoadConfigWithEmptyExclude tests that empty exclude list works
func TestLoadConfigWithEmptyExclude(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	configData := map[string]interface{}{
		"exclude": []string{},
	}

	configPath := filepath.Join(tmpDir, ".cclintrc.json")
	jsonData, _ := json.MarshalIndent(configData, "", "  ")
	_ = os.WriteFile(configPath, jsonData, 0644)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	config, err := LoadConfig("")
	require.NoError(t, err)
	assert.Empty(t, config.Exclude)
}

// TestLoadConfigWithNullExtensions tests that null extensions map works
func TestLoadConfigWithNullExtensions(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	configData := map[string]interface{}{
		"schemas": map[string]interface{}{
			"enabled": true,
			"extensions": nil,
		},
	}

	configPath := filepath.Join(tmpDir, ".cclintrc.json")
	jsonData, _ := json.MarshalIndent(configData, "", "  ")
	_ = os.WriteFile(configPath, jsonData, 0644)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	config, err := LoadConfig("")
	require.NoError(t, err)
	assert.True(t, config.Schemas.Enabled)
	assert.Nil(t, config.Schemas.Extensions)
}

// TestLoadConfigEmptyRootPath tests that empty rootPath uses config value
func TestLoadConfigEmptyRootPath(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	configData := map[string]interface{}{
		"root": "/custom/path",
	}

	configPath := filepath.Join(tmpDir, ".cclintrc.json")
	jsonData, _ := json.MarshalIndent(configData, "", "  ")
	_ = os.WriteFile(configPath, jsonData, 0644)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	// Pass empty string for rootPath
	config, err := LoadConfig("")
	require.NoError(t, err)
	assert.Equal(t, "/custom/path", config.Root)
}

// TestValidateConfigAllFormats tests all valid format values
func TestValidateConfigAllFormats(t *testing.T) {
	formats := []struct {
		format      string
		requiresOut bool
	}{
		{"console", false},
		{"json", true},
		{"markdown", true},
	}

	for _, tc := range formats {
		t.Run(tc.format, func(t *testing.T) {
			config := &Config{
				Format:      tc.format,
				FailOn:      "error",
				Concurrency: 10,
			}

			if tc.requiresOut {
				config.Output = "output.txt"
			}

			err := validateConfig(config)
			assert.NoError(t, err)
		})
	}
}

// TestValidateConfigAllFailOnLevels tests all valid failOn values
func TestValidateConfigAllFailOnLevels(t *testing.T) {
	levels := []string{"error", "warning", "suggestion"}

	for _, level := range levels {
		t.Run(level, func(t *testing.T) {
			config := &Config{
				Format:      "console",
				FailOn:      level,
				Concurrency: 10,
			}

			err := validateConfig(config)
			assert.NoError(t, err)
		})
	}
}

// TestValidateConfigMinConcurrency tests boundary condition for concurrency
func TestValidateConfigMinConcurrency(t *testing.T) {
	// Test concurrency = 1 (minimum valid)
	config := &Config{
		Format:      "console",
		FailOn:      "error",
		Concurrency: 1,
	}
	err := validateConfig(config)
	assert.NoError(t, err)

	// Test concurrency = -1 (invalid)
	config.Concurrency = -1
	err = validateConfig(config)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "concurrency must be at least 1")
}

// TestSaveConfigWithComplexExtensions tests saving config with complex extension types
func TestSaveConfigWithComplexExtensions(t *testing.T) {
	tmpDir := setupTestDir(t)

	config := &Config{
		Format:      "console",
		FailOn:      "error",
		Concurrency: 10,
		Schemas: SchemaConfig{
			Enabled: true,
			Extensions: map[string]interface{}{
				"string":  "value",
				"number":  42,
				"boolean": true,
				"array":   []interface{}{"a", "b", "c"},
				"nested": map[string]interface{}{
					"key": "value",
				},
			},
		},
	}

	savePath := filepath.Join(tmpDir, "complex.json")
	err := SaveConfig(config, savePath)
	require.NoError(t, err)

	// Read back and verify
	data, err := os.ReadFile(savePath)
	require.NoError(t, err)

	var loaded Config
	err = json.Unmarshal(data, &loaded)
	require.NoError(t, err)

	assert.Equal(t, "value", loaded.Schemas.Extensions["string"])
	// JSON numbers are float64
	assert.Equal(t, float64(42), loaded.Schemas.Extensions["number"])
	assert.Equal(t, true, loaded.Schemas.Extensions["boolean"])
}

// TestLoadConfigNoConfigFiles tests behavior when no config files exist
func TestLoadConfigNoConfigFiles(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Ensure no config files exist
	_ = os.Remove(filepath.Join(tmpDir, ".cclintrc.json"))
	_ = os.Remove(filepath.Join(tmpDir, ".cclintrc.yaml"))
	_ = os.Remove(filepath.Join(tmpDir, ".cclintrc.yml"))

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	// Should succeed with defaults
	config, err := LoadConfig("")
	require.NoError(t, err)
	assert.NotNil(t, config)

	homeDir, _ := os.UserHomeDir()
	assert.Equal(t, filepath.Join(homeDir, ".claude"), config.Root)
}

// TestLoadConfigPartialConfig tests loading partial config (only some fields set)
func TestLoadConfigPartialConfig(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)

	// Only set a few fields, others should use defaults
	configData := map[string]interface{}{
		"quiet":   true,
		"verbose": true,
	}

	configPath := filepath.Join(tmpDir, ".cclintrc.json")
	jsonData, _ := json.MarshalIndent(configData, "", "  ")
	_ = os.WriteFile(configPath, jsonData, 0644)

	oldWd, _ := os.Getwd()
	_ = os.Chdir(tmpDir)
	defer func() { _ = os.Chdir(oldWd) }()

	config, err := LoadConfig("")
	require.NoError(t, err)

	// Verify set values
	assert.True(t, config.Quiet)
	assert.True(t, config.Verbose)

	// Verify defaults for unset values
	assert.Equal(t, "console", config.Format)
	assert.Equal(t, "error", config.FailOn)
	assert.Equal(t, 10, config.Concurrency)
}
