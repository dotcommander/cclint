package crossfile

import (
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

func TestNewCrossFileValidator(t *testing.T) {
	files := []discovery.File{
		{RelPath: "agents/test-specialist.md", Type: discovery.FileTypeAgent, Contents: "test"},
		{RelPath: "skills/foo-bar/SKILL.md", Type: discovery.FileTypeSkill, Contents: "test"},
		{RelPath: "commands/cmd.md", Type: discovery.FileTypeCommand, Contents: "test"},
	}

	v := NewCrossFileValidator(files)

	if v == nil {
		t.Fatal("NewCrossFileValidator() returned nil")
	}

	if _, exists := v.agents["test-specialist"]; !exists {
		t.Error("agents index missing test-specialist")
	}

	if _, exists := v.skills["foo-bar"]; !exists {
		t.Error("skills index missing foo-bar")
	}

	if _, exists := v.commands["cmd"]; !exists {
		t.Error("commands index missing cmd")
	}
}

func TestValidateCommand(t *testing.T) {
	files := []discovery.File{
		{RelPath: "agents/test-specialist.md", Type: discovery.FileTypeAgent, Contents: "test"},
		{RelPath: "skills/foo/SKILL.md", Type: discovery.FileTypeSkill, Contents: "test"},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name         string
		filePath     string
		contents     string
		frontmatter  map[string]any
		wantErrCount int
	}{
		{
			name:     "valid Task() reference",
			filePath: "commands/test.md",
			contents: "Task(test-specialist): do something",
			wantErrCount: 0,
		},
		{
			name:     "missing agent reference",
			filePath: "commands/test.md",
			contents: "Task(nonexistent-agent): do something",
			wantErrCount: 1,
		},
		{
			name:     "built-in subagent type",
			filePath: "commands/test.md",
			contents: "Task(general-purpose): do something",
			wantErrCount: 0,
		},
		{
			name:     "unused allowed-tools",
			filePath: "commands/test.md",
			contents: "Just content",
			frontmatter: map[string]any{
				"allowed-tools": "Read",
			},
			wantErrCount: 1, // Info level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.frontmatter == nil {
				tt.frontmatter = map[string]any{}
			}
			errors := v.ValidateCommand(tt.filePath, tt.contents, tt.frontmatter)

			if len(errors) != tt.wantErrCount {
				t.Errorf("ValidateCommand() errors = %d, want %d", len(errors), tt.wantErrCount)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestValidateAgent(t *testing.T) {
	files := []discovery.File{
		{RelPath: "skills/foo/SKILL.md", Type: discovery.FileTypeSkill, Contents: "test"},
		{RelPath: "skills/bar/SKILL.md", Type: discovery.FileTypeSkill, Contents: "test"},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name         string
		filePath     string
		contents     string
		wantErrCount int
	}{
		{
			name:         "valid Skill: reference",
			filePath:     "agents/test.md",
			contents:     "Skill: foo",
			wantErrCount: 0,
		},
		{
			name:         "missing skill reference",
			filePath:     "agents/test.md",
			contents:     "Skill: nonexistent",
			wantErrCount: 1,
		},
		{
			name:         "Skill() format",
			filePath:     "agents/test.md",
			contents:     "Skill(bar)",
			wantErrCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateAgent(tt.filePath, tt.contents, nil)

			if len(errors) != tt.wantErrCount {
				t.Errorf("ValidateAgent() errors = %d, want %d", len(errors), tt.wantErrCount)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestValidateSkill(t *testing.T) {
	files := []discovery.File{
		{RelPath: "agents/test-specialist.md", Type: discovery.FileTypeAgent, Contents: "test"},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name         string
		filePath     string
		contents     string
		wantErrCount int
	}{
		{
			name:         "valid delegate to",
			filePath:     "skills/test/SKILL.md",
			contents:     "delegate to test-specialist",
			wantErrCount: 0,
		},
		{
			name:         "missing agent reference",
			filePath:     "skills/test/SKILL.md",
			contents:     "use nonexistent-specialist",
			wantErrCount: 1,
		},
		{
			name:         "Task() pattern",
			filePath:     "skills/test/SKILL.md",
			contents:     "Task(test-specialist)",
			wantErrCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateSkill(tt.filePath, tt.contents, nil)

			if len(errors) != tt.wantErrCount {
				t.Errorf("ValidateSkill() errors = %d, want %d", len(errors), tt.wantErrCount)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
		})
	}
}

func TestFindSkillReferences(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "Skill: format",
			content: "Skill: foo-bar",
			want:    []string{"foo-bar"},
		},
		{
			name:    "**Skill**: format",
			content: "**Skill**: baz-qux",
			want:    []string{"baz-qux"},
		},
		{
			name:    "Skill() format",
			content: "Skill(test-skill)",
			want:    []string{"test-skill"},
		},
		{
			name:    "multiple references",
			content: "Skill: foo\n**Skill**: bar\nSkill(baz)",
			want:    []string{"foo", "bar", "baz"},
		},
		{
			name:    "no duplicates",
			content: "Skill: foo\nSkill: foo",
			want:    []string{"foo"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindSkillReferences(tt.content)

			if len(got) != len(tt.want) {
				t.Errorf("FindSkillReferences() returned %d skills, want %d", len(got), len(tt.want))
				t.Logf("  Got: %v", got)
				t.Logf("  Want: %v", tt.want)
				return
			}

			gotMap := make(map[string]bool)
			for _, s := range got {
				gotMap[s] = true
			}

			for _, w := range tt.want {
				if !gotMap[w] {
					t.Errorf("FindSkillReferences() missing %q", w)
				}
			}
		})
	}
}

func TestParseAllowedTools(t *testing.T) {
	tests := []struct {
		name string
		s    string
		want []string
	}{
		{
			name: "simple tools",
			s:    "Read, Write, Bash",
			want: []string{"Read", "Write", "Bash"},
		},
		{
			name: "Task() pattern",
			s:    "Task(agent-specialist), Read",
			want: []string{"Task(agent-specialist)", "Read"},
		},
		{
			name: "multiple Task()",
			s:    "Task(a), Task(b), Write",
			want: []string{"Task(a)", "Task(b)", "Write"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseAllowedTools(tt.s)

			if len(got) != len(tt.want) {
				t.Errorf("ParseAllowedTools() returned %d tools, want %d", len(got), len(tt.want))
				t.Logf("  Got: %v", got)
				t.Logf("  Want: %v", tt.want)
				return
			}

			for i, w := range tt.want {
				if got[i] != w {
					t.Errorf("ParseAllowedTools()[%d] = %q, want %q", i, got[i], w)
				}
			}
		})
	}
}

func TestIsToolUsed(t *testing.T) {
	tests := []struct {
		name     string
		tool     string
		contents string
		want     bool
	}{
		{"Task", "Task", "Task(agent): do", true},
		{"Read", "Read", "Use Read() tool", true},
		{"Write", "Write", "Write tool to save", true},
		{"specific Task", "Task(foo)", "Task(foo): do it", true},
		{"unused", "Edit", "no editing here", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsToolUsed(tt.tool, tt.contents)
			if got != tt.want {
				t.Errorf("IsToolUsed(%q) = %v, want %v", tt.tool, got, tt.want)
			}
		})
	}
}

func TestTraceChain(t *testing.T) {
	files := []discovery.File{
		{
			RelPath:  "commands/test.md",
			Type:     discovery.FileTypeCommand,
			Contents: "Task(test-specialist): do something",
		},
		{
			RelPath:  "agents/test-specialist.md",
			Type:     discovery.FileTypeAgent,
			Contents: "Skill: foo-skill",
		},
		{
			RelPath:  "skills/foo-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "Skill methodology",
		},
	}
	v := NewCrossFileValidator(files)

	chain := v.TraceChain("command", "test")
	if chain == nil {
		t.Fatal("TraceChain() returned nil")
	}

	if chain.Type != "command" || chain.Name != "test" {
		t.Errorf("TraceChain() root = %s:%s, want command:test", chain.Type, chain.Name)
	}

	if len(chain.Children) != 1 {
		t.Fatalf("TraceChain() children = %d, want 1", len(chain.Children))
	}

	agentLink := chain.Children[0]
	if agentLink.Type != "agent" || agentLink.Name != "test-specialist" {
		t.Errorf("TraceChain() child = %s:%s, want agent:test-specialist", agentLink.Type, agentLink.Name)
	}

	if len(agentLink.Children) != 1 {
		t.Fatalf("TraceChain() grandchildren = %d, want 1", len(agentLink.Children))
	}

	skillLink := agentLink.Children[0]
	if skillLink.Type != "skill" || skillLink.Name != "foo-skill" {
		t.Errorf("TraceChain() grandchild = %s:%s, want skill:foo-skill", skillLink.Type, skillLink.Name)
	}
}

func TestFormatChain(t *testing.T) {
	link := &ChainLink{
		Type:  "command",
		Name:  "test",
		Lines: 10,
		Children: []ChainLink{
			{
				Type:  "agent",
				Name:  "test-specialist",
				Lines: 50,
			},
		},
	}

	output := FormatChain(link, "")
	if !strings.Contains(output, "test") {
		t.Error("FormatChain() should contain command name")
	}
	if !strings.Contains(output, "test-specialist") {
		t.Error("FormatChain() should contain agent name")
	}
}

func TestFindOrphanedSkills(t *testing.T) {
	files := []discovery.File{
		{
			RelPath:  "skills/used-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "Used skill",
		},
		{
			RelPath:  "skills/orphan-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "Orphaned skill",
		},
		{
			RelPath:  "agents/test.md",
			Type:     discovery.FileTypeAgent,
			Contents: "Skill: used-skill",
		},
	}
	v := NewCrossFileValidator(files)

	orphans := v.FindOrphanedSkills()

	if len(orphans) != 1 {
		t.Errorf("FindOrphanedSkills() returned %d orphans, want 1", len(orphans))
		for _, o := range orphans {
			t.Logf("  Orphan: %s", o.File)
		}
		return
	}

	if !strings.Contains(orphans[0].File, "orphan-skill") {
		t.Errorf("FindOrphanedSkills() = %q, want orphan-skill", orphans[0].File)
	}
}

func TestDetectCycles(t *testing.T) {
	// Create a simple cycle: agent A -> agent B -> agent A
	files := []discovery.File{
		{
			RelPath:  "agents/agent-a.md",
			Type:     discovery.FileTypeAgent,
			Contents: "Task(agent-b): do something",
		},
		{
			RelPath:  "agents/agent-b.md",
			Type:     discovery.FileTypeAgent,
			Contents: "Task(agent-a): do something back",
		},
	}
	v := NewCrossFileValidator(files)

	cycles := v.DetectCycles()

	if len(cycles) == 0 {
		t.Error("DetectCycles() should find cycle but found none")
		return
	}

	// Verify cycle contains both agents
	cycleStr := FormatCycle(cycles[0])
	if !strings.Contains(cycleStr, "agent-a") || !strings.Contains(cycleStr, "agent-b") {
		t.Errorf("DetectCycles() cycle = %q, should contain both agents", cycleStr)
	}
}

func TestCrossExtractFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string) string
		path     string
		expected string
	}{
		{"agent name", ExtractAgentName, "agents/test-specialist.md", "test-specialist"},
		{"skill name", ExtractSkillName, "skills/foo-bar/SKILL.md", "foo-bar"},
		{"command name", ExtractCommandName, "commands/test.md", "test"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fn(tt.path)
			if got != tt.expected {
				t.Errorf("%s(%q) = %q, want %q", tt.name, tt.path, got, tt.expected)
			}
		})
	}
}

func TestBuiltInSubagentTypes(t *testing.T) {
	expected := []string{
		// Built-in subagent types
		"general-purpose", "statusline-setup", "Explore", "Plan", "claude-code-guide",
		// Model names
		"haiku", "sonnet", "opus",
	}
	for _, name := range expected {
		if !BuiltInSubagentTypes[name] {
			t.Errorf("BuiltInSubagentTypes missing: %s", name)
		}
	}
}

func TestTraceChainEdgeCases(t *testing.T) {
	files := []discovery.File{
		{
			RelPath:  "agents/agent-a.md",
			Type:     discovery.FileTypeAgent,
			Contents: "Skill: skill-a\nSkill: skill-b",
		},
		{
			RelPath:  "skills/skill-a/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "Skill methodology",
		},
		{
			RelPath:  "skills/skill-b/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "Another skill",
		},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name          string
		componentType string
		componentName string
		wantNil       bool
		wantChildren  int
	}{
		{
			name:          "agent with multiple skills",
			componentType: "agent",
			componentName: "agent-a",
			wantNil:       false,
			wantChildren:  2,
		},
		{
			name:          "nonexistent component",
			componentType: "command",
			componentName: "nonexistent",
			wantNil:       true,
		},
		{
			name:          "skill trace",
			componentType: "skill",
			componentName: "skill-a",
			wantNil:       false,
			wantChildren:  0,
		},
		{
			name:          "invalid component type",
			componentType: "invalid",
			componentName: "test",
			wantNil:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			chain := v.TraceChain(tt.componentType, tt.componentName)
			if tt.wantNil && chain != nil {
				t.Errorf("TraceChain() = %v, want nil", chain)
			}
			if !tt.wantNil && chain == nil {
				t.Fatal("TraceChain() = nil, want non-nil")
			}
			if !tt.wantNil && len(chain.Children) != tt.wantChildren {
				t.Errorf("TraceChain() children = %d, want %d", len(chain.Children), tt.wantChildren)
			}
		})
	}
}

func TestFindOrphanedSkillsEdgeCases(t *testing.T) {
	tests := []struct {
		name         string
		files        []discovery.File
		wantOrphans  int
		wantContains []string
	}{
		{
			name: "skill referenced by command Task()",
			files: []discovery.File{
				{
					RelPath:  "commands/test.md",
					Type:     discovery.FileTypeCommand,
					Contents: "Task(my-skill)",
				},
				{
					RelPath:  "skills/my-skill/SKILL.md",
					Type:     discovery.FileTypeSkill,
					Contents: "Skill content",
				},
			},
			wantOrphans: 0,
		},
		{
			name: "skill referenced by another skill",
			files: []discovery.File{
				{
					RelPath:  "skills/skill-a/SKILL.md",
					Type:     discovery.FileTypeSkill,
					Contents: "See skill-b for details",
				},
				{
					RelPath:  "skills/skill-b/SKILL.md",
					Type:     discovery.FileTypeSkill,
					Contents: "Skill B content",
				},
			},
			wantOrphans:  1, // skill-a is orphaned
			wantContains: []string{"skill-a"},
		},
		{
			name: "all skills orphaned",
			files: []discovery.File{
				{
					RelPath:  "skills/orphan-1/SKILL.md",
					Type:     discovery.FileTypeSkill,
					Contents: "Orphan 1",
				},
				{
					RelPath:  "skills/orphan-2/SKILL.md",
					Type:     discovery.FileTypeSkill,
					Contents: "Orphan 2",
				},
			},
			wantOrphans:  2,
			wantContains: []string{"orphan-1", "orphan-2"},
		},
		{
			name: "self-reference doesn't count",
			files: []discovery.File{
				{
					RelPath:  "skills/self-ref/SKILL.md",
					Type:     discovery.FileTypeSkill,
					Contents: "This is skill self-ref doing self-ref things",
				},
			},
			wantOrphans:  1,
			wantContains: []string{"self-ref"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewCrossFileValidator(tt.files)
			orphans := v.FindOrphanedSkills()

			if len(orphans) != tt.wantOrphans {
				t.Errorf("FindOrphanedSkills() = %d orphans, want %d", len(orphans), tt.wantOrphans)
				for _, o := range orphans {
					t.Logf("  Orphan: %s", o.File)
				}
			}

			// Check that expected orphans are present
			orphanMap := make(map[string]bool)
			for _, o := range orphans {
				for _, want := range tt.wantContains {
					if strings.Contains(o.File, want) {
						orphanMap[want] = true
					}
				}
			}
			for _, want := range tt.wantContains {
				if !orphanMap[want] {
					t.Errorf("FindOrphanedSkills() missing expected orphan %q", want)
				}
			}
		})
	}
}

func TestDetectCyclesEdgeCases(t *testing.T) {
	tests := []struct {
		name       string
		files      []discovery.File
		wantCycles int
	}{
		{
			name: "no cycles",
			files: []discovery.File{
				{
					RelPath:  "commands/cmd.md",
					Type:     discovery.FileTypeCommand,
					Contents: "Task(agent-a)",
				},
				{
					RelPath:  "agents/agent-a.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Skill: skill-a",
				},
				{
					RelPath:  "skills/skill-a/SKILL.md",
					Type:     discovery.FileTypeSkill,
					Contents: "No cycles here",
				},
			},
			wantCycles: 0,
		},
		{
			name: "skill to skill cycle",
			files: []discovery.File{
				{
					RelPath:  "skills/skill-a/SKILL.md",
					Type:     discovery.FileTypeSkill,
					Contents: "delegate to skill-b-specialist",
				},
				{
					RelPath:  "agents/skill-b-specialist.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Skill: skill-a",
				},
			},
			wantCycles: 1,
		},
		{
			name: "agent self-reference filtered",
			files: []discovery.File{
				{
					RelPath:  "agents/agent-a.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Task(agent-a)", // Self-reference should be filtered
				},
			},
			wantCycles: 0,
		},
		{
			name: "three-way cycle",
			files: []discovery.File{
				{
					RelPath:  "agents/agent-a.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Task(agent-b)",
				},
				{
					RelPath:  "agents/agent-b.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Task(agent-c)",
				},
				{
					RelPath:  "agents/agent-c.md",
					Type:     discovery.FileTypeAgent,
					Contents: "Task(agent-a)",
				},
			},
			wantCycles: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewCrossFileValidator(tt.files)
			cycles := v.DetectCycles()

			if len(cycles) < tt.wantCycles {
				t.Errorf("DetectCycles() found %d cycles, want at least %d", len(cycles), tt.wantCycles)
			}
		})
	}
}

func TestValidateCommandEdgeCases(t *testing.T) {
	files := []discovery.File{
		{
			RelPath:  "agents/primary-agent.md",
			Type:     discovery.FileTypeAgent,
			Contents: "Skill: skill-a\n--verbose flag support",
		},
		{
			RelPath:  "skills/skill-a/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "Supports --verbose flag",
		},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name         string
		contents     string
		frontmatter  map[string]any
		wantErrCount int
		wantWarnings int
	}{
		{
			name:     "flag in agent",
			contents: "Task(primary-agent): do --verbose processing",
			frontmatter: map[string]any{
				"allowed-tools": "Task(primary-agent)",
			},
			wantErrCount: 0,
			wantWarnings: 0,
		},
		{
			name:     "flag in skill",
			contents: "Task(primary-agent): use --verbose",
			frontmatter: map[string]any{
				"allowed-tools": "Task(primary-agent)",
			},
			wantErrCount: 0,
		},
		{
			name:     "fake flag not in agent or skill",
			contents: "Task(primary-agent): use --fake-flag",
			frontmatter: map[string]any{
				"allowed-tools": "Task(primary-agent)",
			},
			wantWarnings: 1, // Should warn about fake flag
		},
		{
			name:     "declarative command with Write",
			contents: "This command will save to output.md",
			frontmatter: map[string]any{
				"allowed-tools": "Write",
			},
			wantErrCount: 0,
		},
		{
			name:     "unused tool non-declarative",
			contents: "Just a command",
			frontmatter: map[string]any{
				"allowed-tools": "Grep",
			},
			wantWarnings: 1, // Should have info about unused tool
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateCommand("commands/test.md", tt.contents, tt.frontmatter)
			errCount := 0
			warnCount := 0
			suggestionCount := 0
			infoCount := 0

			for _, e := range errors {
				switch e.Severity {
				case "error":
					errCount++
				case "warning":
					warnCount++
				case "suggestion":
					suggestionCount++
				case "info":
					infoCount++
				}
			}

			if tt.wantErrCount > 0 && errCount < tt.wantErrCount {
				t.Errorf("ValidateCommand() errors = %d, want at least %d", errCount, tt.wantErrCount)
			}
			if tt.wantWarnings > 0 && (warnCount+suggestionCount+infoCount) < tt.wantWarnings {
				t.Errorf("ValidateCommand() warnings+suggestions+info = %d, want at least %d", warnCount+suggestionCount+infoCount, tt.wantWarnings)
			}
		})
	}
}

func TestFormatChainNil(t *testing.T) {
	output := FormatChain(nil, "")
	if output != "" {
		t.Errorf("FormatChain(nil) = %q, want empty string", output)
	}
}

func TestFormatCycleEmpty(t *testing.T) {
	cycle := Cycle{Path: []string{}}
	output := FormatCycle(cycle)
	if output != "" {
		t.Errorf("FormatCycle(empty) = %q, want empty string", output)
	}
}

func TestDetermineCycleType(t *testing.T) {
	tests := []struct {
		name string
		path []string
		want string
	}{
		{
			name: "empty path",
			path: []string{},
			want: "unknown",
		},
		{
			name: "agent to agent",
			path: []string{"agent:a", "agent:b", "agent:a"},
			want: "agent \u2192 agent \u2192 agent",
		},
		{
			name: "mixed types",
			path: []string{"command:c", "agent:a", "skill:s", "command:c"},
			want: "command \u2192 agent \u2192 skill \u2192 command",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := determineCycleType(tt.path)
			if got != tt.want {
				t.Errorf("determineCycleType() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestValidateCommand_SkillReferences tests skill reference validation in commands
func TestValidateCommand_SkillReferences(t *testing.T) {
	files := []discovery.File{
		{RelPath: "skills/existing-skill/SKILL.md", Type: discovery.FileTypeSkill, Contents: "Skill content"},
		{RelPath: "agents/test-agent.md", Type: discovery.FileTypeAgent, Contents: "Agent content"},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name        string
		contents    string
		frontmatter map[string]any
		wantErrors  int
		wantMessage string
	}{
		{
			name:       "valid Skill: reference",
			contents:   "Use Skill: existing-skill",
			wantErrors: 0,
		},
		{
			name:        "missing Skill: reference",
			contents:    "Use Skill: nonexistent-skill",
			wantErrors:  1,
			wantMessage: "non-existent skill 'nonexistent-skill'",
		},
		{
			name:       "valid Skill() function call",
			contents:   "Load via Skill(existing-skill)",
			wantErrors: 0,
		},
		{
			name:        "missing Skill() function call",
			contents:    `Invoke Skill("missing-skill")`,
			wantErrors:  1,
			wantMessage: "non-existent skill 'missing-skill'",
		},
		{
			name:       "Skill() with single quotes",
			contents:   "Call Skill('existing-skill')",
			wantErrors: 0,
		},
		{
			name:       "Skill() with whitespace",
			contents:   "Call Skill( existing-skill )",
			wantErrors: 0,
		},
		{
			name:       "no skill references",
			contents:   "Just a command with Task(test-agent)",
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.frontmatter == nil {
				tt.frontmatter = map[string]any{}
			}
			errors := v.ValidateCommand("commands/test.md", tt.contents, tt.frontmatter)

			// Filter to only skill-related errors
			skillErrors := 0
			for _, e := range errors {
				if strings.Contains(e.Message, "skill") {
					skillErrors++
				}
			}

			if skillErrors != tt.wantErrors {
				t.Errorf("ValidateCommand() skill errors = %d, want %d", skillErrors, tt.wantErrors)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
			if tt.wantMessage != "" && skillErrors > 0 {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.wantMessage) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing %q, not found", tt.wantMessage)
				}
			}
		})
	}
}

// TestValidateSkill_BroadAgentPatterns tests expanded agent reference patterns in skills
func TestValidateSkill_BroadAgentPatterns(t *testing.T) {
	files := []discovery.File{
		{RelPath: "agents/test-agent.md", Type: discovery.FileTypeAgent, Contents: "Agent content"},
		{RelPath: "agents/helper-agent.md", Type: discovery.FileTypeAgent, Contents: "Helper content"},
		{RelPath: "agents/test-specialist.md", Type: discovery.FileTypeAgent, Contents: "Specialist content"},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name        string
		contents    string
		wantErrors  int
		wantMessage string
	}{
		{
			name:       "Task() without specialist suffix - exists",
			contents:   "Delegates via Task(test-agent)",
			wantErrors: 0,
		},
		{
			name:        "Task() without specialist suffix - missing",
			contents:    "Delegates via Task(missing-agent)",
			wantErrors:  1,
			wantMessage: "agent doesn't exist",
		},
		{
			name:       "built-in agent Explore - no error",
			contents:   "Use Task(Explore) to discover",
			wantErrors: 0,
		},
		{
			name:       "built-in agent Plan - no error",
			contents:   "Use Task(Plan) for planning",
			wantErrors: 0,
		},
		{
			name:       "built-in agent general-purpose - no error",
			contents:   "Use Task(general-purpose) for general tasks",
			wantErrors: 0,
		},
		{
			name:        "delegate via pattern - missing",
			contents:    "delegate via nonexistent-agent",
			wantErrors:  1,
			wantMessage: "agent doesn't exist",
		},
		{
			name:       "delegate via pattern - exists",
			contents:   "delegate via test-agent",
			wantErrors: 0,
		},
		{
			name:        "agent handles pattern - missing",
			contents:    "The missing-agent handles this task",
			wantErrors:  1,
			wantMessage: "agent doesn't exist",
		},
		{
			name:       "agent handles pattern - exists",
			contents:   "The helper-agent handles this task",
			wantErrors: 0,
		},
		{
			name:       "Task() with quotes - exists",
			contents:   `Use Task("test-agent") for work`,
			wantErrors: 0,
		},
		{
			name:        "Task() with quotes - missing",
			contents:    `Use Task("fake-agent") for work`,
			wantErrors:  1,
			wantMessage: "agent doesn't exist",
		},
		{
			name:       "dynamic reference - skip validation",
			contents:   "Task(args.agent): dynamic",
			wantErrors: 0, // Should be skipped due to dot in name
		},
		{
			name:       "subagent_type reference - skip validation",
			contents:   "Task(subagent_type): dynamic type",
			wantErrors: 0, // Should be skipped
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateSkill("skills/test/SKILL.md", tt.contents, nil)

			if len(errors) != tt.wantErrors {
				t.Errorf("ValidateSkill() errors = %d, want %d", len(errors), tt.wantErrors)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
			if tt.wantMessage != "" && len(errors) > 0 {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.wantMessage) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing %q, not found", tt.wantMessage)
				}
			}
		})
	}
}

// TestFindSkillReferences_FunctionCalls tests all Skill() function call formats
func TestFindSkillReferences_FunctionCalls(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected []string
	}{
		{
			name:     "Skill() without quotes",
			content:  "Load via Skill(foo-bar)",
			expected: []string{"foo-bar"},
		},
		{
			name:     "Skill() with double quotes",
			content:  `Use Skill("my-skill")`,
			expected: []string{"my-skill"},
		},
		{
			name:     "Skill() with single quotes",
			content:  `Invoke Skill('test-skill')`,
			expected: []string{"test-skill"},
		},
		{
			name:     "Skill() with whitespace",
			content:  "Call Skill( some-skill )",
			expected: []string{"some-skill"},
		},
		{
			name:     "multiple Skill() calls",
			content:  "Skill(first)\nSkill(second)\nSkill(third)",
			expected: []string{"first", "second", "third"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FindSkillReferences(tt.content)
			if len(got) != len(tt.expected) {
				t.Errorf("FindSkillReferences() = %v, want %v", got, tt.expected)
				return
			}
			gotMap := make(map[string]bool)
			for _, s := range got {
				gotMap[s] = true
			}
			for _, want := range tt.expected {
				if !gotMap[want] {
					t.Errorf("FindSkillReferences() missing %q in %v", want, got)
				}
			}
		})
	}
}

// TestCrossFileValidation_EndToEnd tests the full cross-file validation flow
func TestCrossFileValidation_EndToEnd(t *testing.T) {
	files := []discovery.File{
		// Agent referencing skills
		{
			RelPath:  "agents/test-agent.md",
			Type:     discovery.FileTypeAgent,
			Contents: "Skill: real-skill\nSkill(another-skill)\nTask(helper-agent)",
		},
		// Command referencing agent and skill
		{
			RelPath:  "commands/test-cmd.md",
			Type:     discovery.FileTypeCommand,
			Contents: "Task(test-agent)\nUse Skill: real-skill",
		},
		// Skill referencing agent
		{
			RelPath:  "skills/real-skill/SKILL.md",
			Type:     discovery.FileTypeSkill,
			Contents: "Delegate to helper-agent via Task(helper-agent)",
		},
		// Helper agent
		{
			RelPath:  "agents/helper-agent.md",
			Type:     discovery.FileTypeAgent,
			Contents: "Helper agent content",
		},
	}

	validator := NewCrossFileValidator(files)

	// Test agent validation - should error on 'another-skill' (missing)
	agentErrors := validator.ValidateAgent("agents/test-agent.md", files[0].Contents, nil)
	if len(agentErrors) != 1 {
		t.Errorf("ValidateAgent() errors = %d, want 1", len(agentErrors))
		for _, e := range agentErrors {
			t.Logf("  Error: %s", e.Message)
		}
	} else if !strings.Contains(agentErrors[0].Message, "another-skill") {
		t.Errorf("Expected error about 'another-skill', got: %s", agentErrors[0].Message)
	}

	// Test command validation - all references should be valid
	cmdErrors := validator.ValidateCommand("commands/test-cmd.md", files[1].Contents, map[string]any{})
	// Filter to non-info errors
	realCmdErrors := 0
	for _, e := range cmdErrors {
		if e.Severity == "error" {
			realCmdErrors++
		}
	}
	if realCmdErrors != 0 {
		t.Errorf("ValidateCommand() errors = %d, want 0", realCmdErrors)
		for _, e := range cmdErrors {
			t.Logf("  %s: %s", e.Severity, e.Message)
		}
	}

	// Test skill validation - helper-agent exists
	skillErrors := validator.ValidateSkill("skills/real-skill/SKILL.md", files[2].Contents, nil)
	if len(skillErrors) != 0 {
		t.Errorf("ValidateSkill() errors = %d, want 0", len(skillErrors))
		for _, e := range skillErrors {
			t.Logf("  Error: %s", e.Message)
		}
	}
}

// TestValidateSkill_FrontmatterAgent tests the frontmatter agent field validation
func TestValidateSkill_FrontmatterAgent(t *testing.T) {
	files := []discovery.File{
		{RelPath: "agents/my-agent.md", Type: discovery.FileTypeAgent, Contents: "Agent content"},
		{RelPath: "agents/helper-agent.md", Type: discovery.FileTypeAgent, Contents: "Helper content"},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name        string
		contents    string
		frontmatter map[string]any
		wantErrors  int
		wantMessage string
	}{
		{
			name:     "skill with existing agent in frontmatter",
			contents: "Skill methodology content",
			frontmatter: map[string]any{
				"agent": "my-agent",
			},
			wantErrors: 0,
		},
		{
			name:     "skill with non-existent agent in frontmatter",
			contents: "Skill methodology content",
			frontmatter: map[string]any{
				"agent": "ghost-agent",
			},
			wantErrors:  1,
			wantMessage: "non-existent agent 'ghost-agent'",
		},
		{
			name:     "skill with built-in agent Explore in frontmatter",
			contents: "Skill methodology content",
			frontmatter: map[string]any{
				"agent": "Explore",
			},
			wantErrors: 0,
		},
		{
			name:     "skill with built-in agent Plan in frontmatter",
			contents: "Skill methodology content",
			frontmatter: map[string]any{
				"agent": "Plan",
			},
			wantErrors: 0,
		},
		{
			name:     "skill with built-in agent general-purpose in frontmatter",
			contents: "Skill methodology content",
			frontmatter: map[string]any{
				"agent": "general-purpose",
			},
			wantErrors: 0,
		},
		{
			name:        "skill with no agent in frontmatter",
			contents:    "Skill methodology content",
			frontmatter: map[string]any{},
			wantErrors:  0,
		},
		{
			name:        "skill with nil frontmatter",
			contents:    "Skill methodology content",
			frontmatter: nil,
			wantErrors:  0,
		},
		{
			name:     "skill with empty agent string",
			contents: "Skill methodology content",
			frontmatter: map[string]any{
				"agent": "",
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateSkill("skills/test/SKILL.md", tt.contents, tt.frontmatter)

			// Filter to frontmatter-agent-specific errors
			fmErrors := 0
			for _, e := range errors {
				if strings.Contains(e.Message, "Frontmatter agent") {
					fmErrors++
				}
			}

			if fmErrors != tt.wantErrors {
				t.Errorf("ValidateSkill() frontmatter agent errors = %d, want %d", fmErrors, tt.wantErrors)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
			if tt.wantMessage != "" && fmErrors > 0 {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.wantMessage) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing %q, not found", tt.wantMessage)
				}
			}
		})
	}
}

// TestExtractTaskAgentRefs tests extraction of Task(agent-name) from tools field
func TestExtractTaskAgentRefs(t *testing.T) {
	tests := []struct {
		name  string
		tools any
		want  []string
	}{
		{
			name:  "nil tools",
			tools: nil,
			want:  nil,
		},
		{
			name:  "wildcard string",
			tools: "*",
			want:  nil,
		},
		{
			name:  "string with single Task ref",
			tools: "Read, Task(helper-agent), Write",
			want:  []string{"helper-agent"},
		},
		{
			name:  "string with multiple Task refs",
			tools: "Task(agent-a), Task(agent-b), Read",
			want:  []string{"agent-a", "agent-b"},
		},
		{
			name:  "array with Task refs",
			tools: []any{"Read", "Task(worker-agent)", "Write", "Task(reviewer-agent)"},
			want:  []string{"worker-agent", "reviewer-agent"},
		},
		{
			name:  "array with no Task refs",
			tools: []any{"Read", "Write", "Bash"},
			want:  nil,
		},
		{
			name:  "string with no Task refs",
			tools: "Read, Write, Bash",
			want:  nil,
		},
		{
			name:  "deduplicates repeated refs",
			tools: "Task(same-agent), Task(same-agent)",
			want:  []string{"same-agent"},
		},
		{
			name:  "non-string non-array type",
			tools: 42,
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractTaskAgentRefs(tt.tools)

			if tt.want == nil && got != nil {
				t.Errorf("ExtractTaskAgentRefs() = %v, want nil", got)
				return
			}
			if tt.want != nil && got == nil {
				t.Errorf("ExtractTaskAgentRefs() = nil, want %v", tt.want)
				return
			}
			if len(got) != len(tt.want) {
				t.Errorf("ExtractTaskAgentRefs() = %v (len %d), want %v (len %d)", got, len(got), tt.want, len(tt.want))
				return
			}

			gotMap := make(map[string]bool)
			for _, ref := range got {
				gotMap[ref] = true
			}
			for _, w := range tt.want {
				if !gotMap[w] {
					t.Errorf("ExtractTaskAgentRefs() missing %q in %v", w, got)
				}
			}
		})
	}
}

// TestValidateAgent_ToolsAgentRefs tests cross-file validation of Task() refs in tools field
func TestValidateAgent_ToolsAgentRefs(t *testing.T) {
	files := []discovery.File{
		{RelPath: "agents/helper-agent.md", Type: discovery.FileTypeAgent, Contents: "Agent content"},
		{RelPath: "agents/worker-agent.md", Type: discovery.FileTypeAgent, Contents: "Worker content"},
		{RelPath: "skills/real-skill/SKILL.md", Type: discovery.FileTypeSkill, Contents: "Skill content"},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name        string
		contents    string
		frontmatter map[string]any
		wantWarns   int
		wantMessage string
	}{
		{
			name:     "tools with existing agent Task ref",
			contents: "Agent content",
			frontmatter: map[string]any{
				"tools": []any{"Read", "Task(helper-agent)", "Write"},
			},
			wantWarns: 0,
		},
		{
			name:     "tools with missing agent Task ref",
			contents: "Agent content",
			frontmatter: map[string]any{
				"tools": []any{"Read", "Task(ghost-agent)", "Write"},
			},
			wantWarns:   1,
			wantMessage: "Task(ghost-agent) references non-existent agent",
		},
		{
			name:     "tools string with missing agent",
			contents: "Agent content",
			frontmatter: map[string]any{
				"tools": "Read, Task(missing-agent), Write",
			},
			wantWarns:   1,
			wantMessage: "Task(missing-agent) references non-existent agent",
		},
		{
			name:     "tools with built-in agent ref",
			contents: "Agent content",
			frontmatter: map[string]any{
				"tools": []any{"Task(sonnet)", "Task(Explore)"},
			},
			wantWarns: 0,
		},
		{
			name:     "tools with mixed existing and missing",
			contents: "Agent content",
			frontmatter: map[string]any{
				"tools": []any{"Task(helper-agent)", "Task(nonexistent-agent)"},
			},
			wantWarns:   1,
			wantMessage: "Task(nonexistent-agent) references non-existent agent",
		},
		{
			name:        "no tools field",
			contents:    "Agent content",
			frontmatter: map[string]any{},
			wantWarns:   0,
		},
		{
			name:        "nil frontmatter",
			contents:    "Agent content",
			frontmatter: nil,
			wantWarns:   0,
		},
		{
			name:     "tools wildcard no warnings",
			contents: "Agent content",
			frontmatter: map[string]any{
				"tools": "*",
			},
			wantWarns: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateAgent("agents/test.md", tt.contents, tt.frontmatter)

			// Filter to tools-agent-ref warnings
			toolsWarns := 0
			for _, e := range errors {
				if strings.Contains(e.Message, "tools field Task(") {
					toolsWarns++
				}
			}

			if toolsWarns != tt.wantWarns {
				t.Errorf("ValidateAgent() tools agent ref warnings = %d, want %d", toolsWarns, tt.wantWarns)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
			if tt.wantMessage != "" && toolsWarns > 0 {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.wantMessage) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected warning containing %q, not found", tt.wantMessage)
				}
			}
		})
	}
}

// TestValidateAgent_FrontmatterSkills tests the frontmatter skills array validation
func TestValidateAgent_FrontmatterSkills(t *testing.T) {
	files := []discovery.File{
		{RelPath: "skills/real-skill/SKILL.md", Type: discovery.FileTypeSkill, Contents: "Skill content"},
		{RelPath: "skills/another-skill/SKILL.md", Type: discovery.FileTypeSkill, Contents: "Another skill"},
	}
	v := NewCrossFileValidator(files)

	tests := []struct {
		name        string
		contents    string
		frontmatter map[string]any
		wantErrors  int
		wantMessage string
	}{
		{
			name:     "agent with existing skills in frontmatter",
			contents: "Agent content",
			frontmatter: map[string]any{
				"skills": []any{"real-skill", "another-skill"},
			},
			wantErrors: 0,
		},
		{
			name:     "agent with non-existent skill in frontmatter",
			contents: "Agent content",
			frontmatter: map[string]any{
				"skills": []any{"real-skill", "missing-skill"},
			},
			wantErrors:  1,
			wantMessage: "non-existent skill 'missing-skill'",
		},
		{
			name:     "agent with all non-existent skills in frontmatter",
			contents: "Agent content",
			frontmatter: map[string]any{
				"skills": []any{"ghost-a", "ghost-b"},
			},
			wantErrors: 2,
		},
		{
			name:        "agent with no skills in frontmatter",
			contents:    "Agent content",
			frontmatter: map[string]any{},
			wantErrors:  0,
		},
		{
			name:        "agent with nil frontmatter",
			contents:    "Agent content",
			frontmatter: nil,
			wantErrors:  0,
		},
		{
			name:     "agent with empty skills array",
			contents: "Agent content",
			frontmatter: map[string]any{
				"skills": []any{},
			},
			wantErrors: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := v.ValidateAgent("agents/test.md", tt.contents, tt.frontmatter)

			// Filter to frontmatter-skills-specific errors
			fmErrors := 0
			for _, e := range errors {
				if strings.Contains(e.Message, "Frontmatter skills") {
					fmErrors++
				}
			}

			if fmErrors != tt.wantErrors {
				t.Errorf("ValidateAgent() frontmatter skills errors = %d, want %d", fmErrors, tt.wantErrors)
				for _, e := range errors {
					t.Logf("  %s: %s", e.Severity, e.Message)
				}
			}
			if tt.wantMessage != "" && fmErrors > 0 {
				found := false
				for _, e := range errors {
					if strings.Contains(e.Message, tt.wantMessage) {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected error containing %q, not found", tt.wantMessage)
				}
			}
		})
	}
}
