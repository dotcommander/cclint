package crossfile

import (
	"strings"
	"testing"

	"github.com/dotcommander/cclint/internal/discovery"
)

func TestAgentPreloadsCountOnlyResolvedSkillIdentities(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		agent       string
		skills      []string
		wantOrphans []string
		missing     string
	}{
		{
			name:        "inline list with unrelated skill",
			agent:       "---\nskills: [my-skill]\n---\nAgent body",
			skills:      []string{"my-skill", "unrelated"},
			wantOrphans: []string{"unrelated"},
		},
		{
			name:        "block list with duplicate preloads",
			agent:       "---\nskills:\n  - my-skill\n  - my-skill\n---\nAgent body",
			skills:      []string{"my-skill"},
			wantOrphans: nil,
		},
		{
			name:        "comma-separated string",
			agent:       "---\nskills: my-skill, second-skill\n---\nAgent body",
			skills:      []string{"my-skill", "second-skill"},
			wantOrphans: nil,
		},
		{
			name:        "body only",
			agent:       "Agent body\nSkill: my-skill",
			skills:      []string{"my-skill"},
			wantOrphans: nil,
		},
		{
			name:        "body and preload overlap",
			agent:       "---\nskills: [my-skill]\n---\nSkill: my-skill",
			skills:      []string{"my-skill"},
			wantOrphans: nil,
		},
		{
			name:        "missing preload cannot mark a substring",
			agent:       "---\nskills: [missing-skill]\n---\nAgent body",
			skills:      []string{"skill", "my-skill"},
			wantOrphans: []string{"skill", "my-skill"},
			missing:     "missing-skill",
		},
		{
			name:        "qualified preload does not mark local basename",
			agent:       "---\nskills: [alpha:shared]\n---\nAgent body",
			skills:      []string{"shared"},
			wantOrphans: []string{"shared"},
		},
		{
			name:        "qualified preload does not mark another plugin basename",
			agent:       "---\nskills: [alpha:shared]\n---\nAgent body",
			skills:      []string{"plugins/marketplaces/market/plugins/beta/skills/shared"},
			wantOrphans: []string{"plugins/marketplaces/market/plugins/beta/skills/shared"},
		},
		{
			name:        "qualified preload does not cross plugin namespaces with duplicate basenames",
			agent:       "---\nskills: [alpha:shared]\n---\nAgent body",
			skills:      []string{"plugins/marketplaces/market/plugins/alpha/skills/shared", "plugins/marketplaces/market/plugins/beta/skills/shared"},
			wantOrphans: []string{"plugins/marketplaces/market/plugins/beta/skills/shared"},
		},
		{
			name:        "malformed YAML does not create edges",
			agent:       "---\nskills: [my-skill\n---\nAgent body",
			skills:      []string{"my-skill"},
			wantOrphans: []string{"my-skill"},
		},
		{
			name:        "unclosed frontmatter does not create edges",
			agent:       "---\nskills: [my-skill]",
			skills:      []string{"my-skill"},
			wantOrphans: []string{"my-skill"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			files := []discovery.File{{RelPath: "agents/preloader.md", Type: discovery.FileTypeAgent, Contents: tc.agent}}
			for _, name := range tc.skills {
				path := name
				if !strings.HasPrefix(path, "plugins/") {
					path = "skills/" + path
				}
				files = append(files, discovery.File{
					RelPath: path + "/SKILL.md", Type: discovery.FileTypeSkill, Contents: "Documentation",
				})
			}
			v := NewCrossFileValidator(files)
			orphans := v.FindOrphanedSkills()
			got := make(map[string]bool, len(orphans))
			for _, orphan := range orphans {
				got[orphan.File] = true
			}
			if len(got) != len(tc.wantOrphans) {
				t.Errorf("orphan paths = %v; want skills %v", got, tc.wantOrphans)
			}
			for _, name := range tc.wantOrphans {
				path := name
				if !strings.HasPrefix(path, "plugins/") {
					path = "skills/" + path
				}
				path += "/SKILL.md"
				if !got[path] {
					t.Errorf("missing orphan %q in %v", path, got)
				}
			}
			if tc.missing != "" {
				errors := v.ValidateAgent("agents/preloader.md", tc.agent, map[string]any{"skills": []any{tc.missing}})
				if len(errors) != 1 || !strings.Contains(errors[0].Message, tc.missing) {
					t.Errorf("missing preload diagnostics = %v", errors)
				}
			}
		})
	}
}
