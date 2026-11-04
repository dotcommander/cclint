package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateConfigExtensionsShape(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		extensions map[string]any
		wantErr    string
	}{
		{
			name:       "valid string value",
			extensions: map[string]any{"agent": "schemas/agent.cue"},
		},
		{
			name:       "valid string list value",
			extensions: map[string]any{"skill": []any{"a.cue", "b.cue"}},
		},
		{
			name:       "unknown key",
			extensions: map[string]any{"workflow": "x.cue"},
			wantErr:    `schemas.extensions["workflow"]: unknown component type`,
		},
		{
			name:       "non-string value",
			extensions: map[string]any{"agent": 42},
			wantErr:    `schemas.extensions["agent"]`,
		},
		{
			name:       "non-string list element",
			extensions: map[string]any{"claude_md": []any{"ok.cue", true}},
			wantErr:    `schemas.extensions["claude_md"]`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := &Config{
				Root:        t.TempDir(),
				Format:      formatConsole,
				FailOn:      failOnError,
				Concurrency: 10,
				Schemas:     SchemaConfig{Enabled: true, Extensions: tt.extensions},
			}
			err := validateConfig(cfg)
			if tt.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestSchemasExtensionsDefaultIsEmptyMap(t *testing.T) {
	resetViper()
	tmpDir := setupTestDir(t)
	t.Chdir(tmpDir)

	cfg, err := LoadConfig("")
	require.NoError(t, err)
	require.NotNil(t, cfg.Schemas.Extensions)
	assert.Empty(t, cfg.Schemas.Extensions, "untouched default should be an empty map, not nil")

	// A config file value overrides the default.
	configPath := filepath.Join(tmpDir, ".cclintrc.json")
	writeConfigFile(t, configPath, `{"schemas": {"extensions": {"agent": "ext.cue"}}}`)

	cfg, err = LoadConfig("")
	require.NoError(t, err)
	assert.Equal(t, map[string]any{"agent": "ext.cue"}, cfg.Schemas.Extensions)
}

func writeConfigFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatal(err)
	}
}
