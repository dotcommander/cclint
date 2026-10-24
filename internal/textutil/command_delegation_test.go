package textutil

import (
	"strings"
	"testing"
)

func TestCommandImprovementsDelegationAliases(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		body       string
		wantAdvice bool
	}{
		{"current", "Agent(helper)", false},
		{"legacy", "Task(helper)", false},
		{"absent", "Review changes.", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			gotAdvice := false
			for _, rec := range GetCommandImprovements(tt.body, map[string]any{"allowed-tools": "Agent, Task"}) {
				if strings.Contains(rec.Description, "commands should delegate") {
					gotAdvice = true
				}
			}
			if gotAdvice != tt.wantAdvice {
				t.Fatalf("delegation advice = %v, want %v", gotAdvice, tt.wantAdvice)
			}
		})
	}
}
