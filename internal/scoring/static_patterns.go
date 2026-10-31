package scoring

import (
	"regexp"
	"strings"
)

// These fixed scoring patterns are shared across files; compile them once rather
// than recompiling each time a component is scored.
var (
	agentAntiPatternsPattern   = regexp.MustCompile(`(?i)## Anti-Patterns`)
	agentExpectedOutputPattern = regexp.MustCompile(`(?i)## Expected Output`)
	agentHardGatesPattern      = regexp.MustCompile(`(?i)HARD GATE`)
	agentSkillPatterns         = []*regexp.Regexp{
		regexp.MustCompile(`(?i)Skill:\s*\S+`),
		regexp.MustCompile(`(?i)\*\*Skill\*\*:\s*\S+`),
		regexp.MustCompile(`(?i)Skill\(\s*["']?[a-z0-9-]+`),
		regexp.MustCompile(`(?i)Skills:\s*\n`),
	}
	commandTaskPattern    = regexp.MustCompile(`(?:Task|Agent)\([^)]+\)`)
	commandSuccessPattern = regexp.MustCompile(`(?i)Success criteria|^\s*- \[ \]`)
	commandFlagsPattern   = regexp.MustCompile(`(?i)## Flags|--\w+`)
	methodologyPatterns   = []*regexp.Regexp{
		regexp.MustCompile(`(?i)## Workflow`),
		regexp.MustCompile(`(?i)### Phase \d`),
		regexp.MustCompile(`(?i)## Algorithm`),
		regexp.MustCompile(`(?i)## Process`),
		regexp.MustCompile(`(?i)### Step \d`),
	}
	skillRoutingTablePattern      = regexp.MustCompile(`\|.*User Question.*\|.*Action.*\|`)
	skillPhasePattern             = regexp.MustCompile(`(?i)### Phase \d`)
	skillAntiPatternsTablePattern = regexp.MustCompile(`\|.*Anti-Pattern.*\|.*Problem.*\|.*Fix.*\|`)
	skillHardGatePattern          = regexp.MustCompile(`(?i)HARD GATE`)
	skillReferencePattern         = regexp.MustCompile(`references/\w+\.md`)
	skillScoringFormulaPattern    = regexp.MustCompile(`(?i)(score\s*=|scoring formula)`)
)

// practiceMetricCheck encapsulates parameters for a practice metric check.
type practiceMetricCheck struct {
	details *[]Metric
	name    string
	pattern string
	regex   *regexp.Regexp
	points  int
	content string
}

func addPracticeMetric(check practiceMetricCheck) int {
	matched := false
	if check.regex != nil {
		matched = check.regex.MatchString(check.content)
	} else {
		matched = strings.Contains(check.content, check.pattern)
	}
	return recordMetric(check.details, "practices", check.name, matched, check.points)
}
