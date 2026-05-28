package cue

import (
	"testing"
)

// TestValidateCommand_v2152Fields tests fields added in v2.1.152–v2.1.154.
func TestValidateCommand_v2152Fields(t *testing.T) {
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
			name: "valid disallowed-tools as array (v2.1.152)",
			data: map[string]any{
				"name":             "test-cmd",
				"disallowed-tools": []string{"Bash"},
			},
			wantError: false,
		},
		{
			name: "valid disallowed-tools as string (v2.1.152)",
			data: map[string]any{
				"name":             "test-cmd",
				"disallowed-tools": "Bash,Write",
			},
			wantError: false,
		},
		{
			name: "valid allowed-tools with Workflow (v2.1.154)",
			data: map[string]any{
				"name":          "test-cmd",
				"allowed-tools": []string{"Workflow"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs, err := v.ValidateCommand(tt.data)
			if err != nil {
				t.Fatalf("ValidateCommand returned error: %v", err)
			}
			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateCommand() hasErrors = %v, want %v", hasErrors, tt.wantError)
				for _, e := range errs {
					t.Logf("  Error: %s", e.Message)
				}
			}
		})
	}
}

// TestValidateAgent_v2156Fields tests ScheduleWakeup, PushNotification, and REPL tools added in v2.1.156.
func TestValidateAgent_v2156Fields(t *testing.T) {
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
			name: "valid tools with ScheduleWakeup (v2.1.156)",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with ScheduleWakeup tool",
				"tools":       []string{"ScheduleWakeup"},
			},
			wantError: false,
		},
		{
			name: "valid tools with PushNotification (v2.1.156)",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with PushNotification tool",
				"tools":       []string{"PushNotification"},
			},
			wantError: false,
		},
		{
			name: "valid tools with REPL (v2.1.156)",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with REPL tool",
				"tools":       []string{"REPL"},
			},
			wantError: false,
		},
		{
			name: "valid tools array with all three new tools (v2.1.156)",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with all v2.1.156 tools",
				"tools":       []string{"ScheduleWakeup", "PushNotification", "REPL"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs, err := v.ValidateAgent(tt.data)
			if err != nil {
				t.Fatalf("ValidateAgent returned error: %v", err)
			}
			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateAgent() hasErrors = %v, want %v", hasErrors, tt.wantError)
				for _, e := range errs {
					t.Logf("  Error: %s", e.Message)
				}
			}
		})
	}
}

// TestValidateAgent_v2154Fields tests the Workflow tool added in v2.1.154.
func TestValidateAgent_v2154Fields(t *testing.T) {
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
			name: "valid tools with Workflow (v2.1.154)",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with Workflow tool",
				"tools":       []string{"Workflow"},
			},
			wantError: false,
		},
		{
			name: "valid tools array including Workflow alongside others (v2.1.154)",
			data: map[string]any{
				"name":        "test-agent",
				"description": "Agent with Workflow and other tools",
				"tools":       []string{"Read", "Write", "Workflow"},
			},
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errs, err := v.ValidateAgent(tt.data)
			if err != nil {
				t.Fatalf("ValidateAgent returned error: %v", err)
			}
			hasErrors := len(errs) > 0
			if hasErrors != tt.wantError {
				t.Errorf("ValidateAgent() hasErrors = %v, want %v", hasErrors, tt.wantError)
				for _, e := range errs {
					t.Logf("  Error: %s", e.Message)
				}
			}
		})
	}
}
