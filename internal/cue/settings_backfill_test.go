package cue

import (
	"testing"
)

func TestSettingsBackfillFields(t *testing.T) {
	t.Parallel()

	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name: "valid apiKeyHelper string",
			data: map[string]any{
				"apiKeyHelper": "/bin/helper.sh",
			},
			wantError: false,
		},
		{
			name: "invalid apiKeyHelper as number",
			data: map[string]any{
				"apiKeyHelper": 123,
			},
			wantError: true,
		},
		{
			name: "valid additionalDirectories array",
			data: map[string]any{
				"additionalDirectories": []any{"/a", "/b"},
			},
			wantError: false,
		},
		{
			name: "invalid additionalDirectories as string",
			data: map[string]any{
				"additionalDirectories": "/a",
			},
			wantError: true,
		},
		{
			name: "valid autoCompactWindow number",
			data: map[string]any{
				"autoCompactWindow": 5000,
			},
			wantError: false,
		},
		{
			name: "valid includeCoAuthoredBy bool",
			data: map[string]any{
				"includeCoAuthoredBy": false,
			},
			wantError: false,
		},
		{
			name: "invalid includeCoAuthoredBy as string",
			data: map[string]any{
				"includeCoAuthoredBy": "no",
			},
			wantError: true,
		},
		{
			name: "valid autoUpdatesChannel enum",
			data: map[string]any{
				"autoUpdatesChannel": "stable",
			},
			wantError: false,
		},
		{
			name: "invalid autoUpdatesChannel member",
			data: map[string]any{
				"autoUpdatesChannel": "weekly",
			},
			wantError: true,
		},
		{
			name: "valid forceLoginMethod enum",
			data: map[string]any{
				"forceLoginMethod": "console",
			},
			wantError: false,
		},
		{
			name: "invalid forceLoginMethod member",
			data: map[string]any{
				"forceLoginMethod": "sso",
			},
			wantError: true,
		},
		{
			name: "valid statusLine object",
			data: map[string]any{
				"statusLine": map[string]any{"type": "command", "command": "echo hi"},
			},
			wantError: false,
		},
		{
			name: "valid modelOverrides object",
			data: map[string]any{
				"modelOverrides": map[string]any{"plan": "opus"},
			},
			wantError: false,
		},
		{
			name: "valid defaultMode permissive string",
			data: map[string]any{
				"defaultMode": "acceptEdits",
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs, err := v.ValidateSettings(tt.data)
			if err != nil {
				t.Fatalf("ValidateSettings returned error: %v", err)
			}
			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateSettings() hasErrors = %v, want %v", hasErrors, tt.wantError)
				for _, e := range errs {
					t.Logf("  Error: %s", e.Message)
				}
			}
		})
	}
}

func TestSettingsBackfillUIFields(t *testing.T) {
	t.Parallel()

	v := NewValidator()
	if err := v.LoadSchemas("schemas"); err != nil {
		t.Fatalf("Failed to load schemas: %v", err)
	}

	tests := []struct {
		name      string
		data      map[string]any
		wantError bool
	}{
		{
			name: "valid terminalProgressBarEnabled bool",
			data: map[string]any{
				"terminalProgressBarEnabled": true,
			},
			wantError: false,
		},
		{
			name: "invalid terminalProgressBarEnabled as string",
			data: map[string]any{
				"terminalProgressBarEnabled": "yes",
			},
			wantError: true,
		},
		{
			name: "valid defaultShell enum bash",
			data: map[string]any{
				"defaultShell": "bash",
			},
			wantError: false,
		},
		{
			name: "valid defaultShell enum powershell",
			data: map[string]any{
				"defaultShell": "powershell",
			},
			wantError: false,
		},
		{
			name: "invalid defaultShell member",
			data: map[string]any{
				"defaultShell": "zsh",
			},
			wantError: true,
		},
		{
			name: "valid symlinkDirectories array",
			data: map[string]any{
				"symlinkDirectories": []any{"/x"},
			},
			wantError: false,
		},
		{
			name: "valid breakThresholdMinutes number",
			data: map[string]any{
				"breakThresholdMinutes": 30,
			},
			wantError: false,
		},
		{
			name: "valid voice object",
			data: map[string]any{
				"voice": map[string]any{"enabled": true},
			},
			wantError: false,
		},
		{
			name: "valid ultracode bool",
			data: map[string]any{
				"ultracode": true,
			},
			wantError: false,
		},
		{
			name: "valid quietHours object",
			data: map[string]any{
				"quietHours": map[string]any{"enabled": false},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs, err := v.ValidateSettings(tt.data)
			if err != nil {
				t.Fatalf("ValidateSettings returned error: %v", err)
			}
			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateSettings() hasErrors = %v, want %v", hasErrors, tt.wantError)
				for _, e := range errs {
					t.Logf("  Error: %s", e.Message)
				}
			}
		})
	}
}
