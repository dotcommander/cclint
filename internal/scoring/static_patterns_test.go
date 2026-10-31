package scoring

import "testing"

func TestStaticScoringPatterns(t *testing.T) {
	t.Parallel()
	agent := NewAgentScorer()
	if !agent.hasSkillReference("**Skill**: deploy") || !agentAntiPatternsPattern.MatchString("## anti-patterns") {
		t.Fatal("agent patterns missed supported case-insensitive or bold reference")
	}
	if !commandTaskPattern.MatchString("Agent(helper)") || commandSuccessPattern.MatchString("intro\n- [ ] nested checklist") {
		t.Fatal("command patterns changed delegation or start-of-string semantics")
	}
	if !commandSuccessPattern.MatchString("\n- [ ] leading whitespace checklist") {
		t.Fatal("command pattern no longer accepts leading whitespace, including newlines")
	}
	if !IsMethodologySkill("### Phase 2") || IsMethodologySkill("## Notes") {
		t.Fatal("methodology classification changed")
	}
	var details []Metric
	if got := addPracticeMetric(practiceMetricCheck{details: &details, name: "reference", regex: skillReferencePattern, points: 4, content: "references/guide.md"}); got != 4 {
		t.Fatalf("regex metric = %d, want 4", got)
	}
	if got := addPracticeMetric(practiceMetricCheck{details: &details, name: "literal", pattern: "- [ ]", points: 4, content: "- [ ] check"}); got != 4 {
		t.Fatalf("literal metric = %d, want 4", got)
	}
}
