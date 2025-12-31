package cli

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
		frontmatter  map[string]interface{}
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
			frontmatter: map[string]interface{}{
				"allowed-tools": "Read",
			},
			wantErrCount: 1, // Info level
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.frontmatter == nil {
				tt.frontmatter = map[string]interface{}{}
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
			errors := v.ValidateAgent(tt.filePath, tt.contents)

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
			errors := v.ValidateSkill(tt.filePath, tt.contents)

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
			got := findSkillReferences(tt.content)

			if len(got) != len(tt.want) {
				t.Errorf("findSkillReferences() returned %d skills, want %d", len(got), len(tt.want))
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
					t.Errorf("findSkillReferences() missing %q", w)
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
			got := parseAllowedTools(tt.s)

			if len(got) != len(tt.want) {
				t.Errorf("parseAllowedTools() returned %d tools, want %d", len(got), len(tt.want))
				t.Logf("  Got: %v", got)
				t.Logf("  Want: %v", tt.want)
				return
			}

			for i, w := range tt.want {
				if got[i] != w {
					t.Errorf("parseAllowedTools()[%d] = %q, want %q", i, got[i], w)
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
			got := isToolUsed(tt.tool, tt.contents)
			if got != tt.want {
				t.Errorf("isToolUsed(%q) = %v, want %v", tt.tool, got, tt.want)
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
		{"agent name", crossExtractAgentName, "agents/test-specialist.md", "test-specialist"},
		{"skill name", crossExtractSkillName, "skills/foo-bar/SKILL.md", "foo-bar"},
		{"command name", crossExtractCommandName, "commands/test.md", "test"},
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
	expected := []string{"general-purpose", "statusline-setup", "Explore", "Plan", "claude-code-guide"}
	for _, name := range expected {
		if !builtInSubagentTypes[name] {
			t.Errorf("builtInSubagentTypes missing: %s", name)
		}
	}
}
