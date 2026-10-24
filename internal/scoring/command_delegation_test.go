package scoring

import "testing"

func TestCommandScorerAgentDelegationMatchesTask(t *testing.T) {
	t.Parallel()
	frontmatter := map[string]any{
		"allowed-tools": "Agent, AskUserQuestion",
		"description":   "Review source changes.",
	}
	scorer := NewCommandScorer()
	legacy := scorer.Score("Task(helper)", frontmatter, "Task(helper)")
	current := scorer.Score("Agent(helper)", frontmatter, "Agent(helper)")
	if current.Structural != legacy.Structural || current.Practices != legacy.Practices {
		t.Fatalf("Agent structural/practices = %d/%d; Task = %d/%d", current.Structural, current.Practices, legacy.Structural, legacy.Practices)
	}
	absent := scorer.Score("Review changes.", frontmatter, "Review changes.")
	if current.Overall-absent.Overall != 25 {
		t.Fatalf("delegation score increment = %d, want 25", current.Overall-absent.Overall)
	}
}
