package cmd

import (
	"fmt"
	"os"
	"sort"

	"github.com/charmbracelet/lipgloss"
	"github.com/dotcommander/cclint/internal/cli"
	"github.com/dotcommander/cclint/internal/config"
	"github.com/spf13/cobra"
)

var summaryCmd = &cobra.Command{
	Use:   "summary",
	Short: "Show quality summary across all components",
	Long: `Aggregates quality scores across all Claude Code components (agents, commands, skills)
and displays a summary report with quality distribution, top issues, and lowest-scoring components.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := runSummary(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			exitFunc(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(summaryCmd)
}

// ComponentSummary holds aggregated data for summary report
type ComponentSummary struct {
	TotalComponents int
	AgentCount      int
	CommandCount    int
	SkillCount      int
	TierCounts      map[string]int
	TopIssues       map[string]int
	LowestScoring   []ScoredComponent
	AllResults      []cli.LintResult
}

// ScoredComponent represents a component with its score for sorting
type ScoredComponent struct {
	File  string
	Type  string
	Score int
	Tier  string
}

func runSummary() error {
	cfg, err := config.LoadConfig(rootPath)
	if err != nil {
		return fmt.Errorf("error loading configuration: %w", err)
	}

	summary := &ComponentSummary{
		TierCounts: make(map[string]int),
		TopIssues:  make(map[string]int),
	}

	// Run all linters
	agentSummary, err := cli.LintAgents(cfg.Root, true, false, cfg.NoCycleCheck)
	if err != nil {
		return fmt.Errorf("error linting agents: %w", err)
	}
	summary.AgentCount = agentSummary.TotalFiles
	aggregateResults(summary, agentSummary.Results)

	commandSummary, err := cli.LintCommands(cfg.Root, true, false, cfg.NoCycleCheck)
	if err != nil {
		return fmt.Errorf("error linting commands: %w", err)
	}
	summary.CommandCount = commandSummary.TotalFiles
	aggregateResults(summary, commandSummary.Results)

	skillSummary, err := cli.LintSkills(cfg.Root, true, false, cfg.NoCycleCheck)
	if err != nil {
		return fmt.Errorf("error linting skills: %w", err)
	}
	summary.SkillCount = skillSummary.TotalFiles
	aggregateResults(summary, skillSummary.Results)

	summary.TotalComponents = summary.AgentCount + summary.CommandCount + summary.SkillCount

	// Sort lowest scoring
	sort.Slice(summary.LowestScoring, func(i, j int) bool {
		return summary.LowestScoring[i].Score < summary.LowestScoring[j].Score
	})

	// Print summary report
	printSummaryReport(summary)

	return nil
}

func aggregateResults(summary *ComponentSummary, results []cli.LintResult) {
	for _, result := range results {
		summary.AllResults = append(summary.AllResults, result)

		if result.Quality != nil {
			summary.TierCounts[result.Quality.Tier]++
			summary.LowestScoring = append(summary.LowestScoring, ScoredComponent{
				File:  result.File,
				Type:  result.Type,
				Score: result.Quality.Overall,
				Tier:  result.Quality.Tier,
			})
		}

		// Aggregate issues
		for _, err := range result.Errors {
			summary.TopIssues[categorizeIssue(err.Message)]++
		}
		for _, sug := range result.Suggestions {
			summary.TopIssues[categorizeIssue(sug.Message)]++
		}
	}
}

func categorizeIssue(message string) string {
	// Categorize issues into buckets for aggregation
	switch {
	case contains(message, "lines", "Best practice"):
		return "Oversized component (fat)"
	case contains(message, "Foundation"):
		return "Missing Foundation section"
	case contains(message, "Workflow"):
		return "Missing Workflow section"
	case contains(message, "Anti-Pattern"):
		return "Missing Anti-Patterns section"
	case contains(message, "Quick Reference", "semantic routing"):
		return "Missing semantic routing"
	case contains(message, "Success Criteria"):
		return "Missing Success Criteria"
	case contains(message, "Expected Output"):
		return "Missing Expected Output"
	case contains(message, "Skill()", "methodology"):
		return "Embedded methodology (should extract)"
	case contains(message, "triggers"):
		return "Missing or incomplete triggers"
	case contains(message, "PROACTIVELY"):
		return "Missing PROACTIVELY pattern"
	case contains(message, "model"):
		return "Missing model specification"
	case contains(message, "description"):
		return "Missing or poor description"
	default:
		return "Other issues"
	}
}

func contains(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if len(sub) > 0 && len(s) >= len(sub) {
			for i := 0; i <= len(s)-len(sub); i++ {
				if s[i:i+len(sub)] == sub {
					return true
				}
			}
		}
	}
	return false
}

// printStyles holds all the styles used in the summary report.
type printStyles struct {
	header lipgloss.Style
	tierA  lipgloss.Style
	tierB  lipgloss.Style
	tierC  lipgloss.Style
	tierDF lipgloss.Style
	dim    lipgloss.Style
}

// newPrintStyles creates a new set of print styles.
func newPrintStyles() printStyles {
	return printStyles{
		header: lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12")),
		tierA:  lipgloss.NewStyle().Foreground(lipgloss.Color("10")),
		tierB:  lipgloss.NewStyle().Foreground(lipgloss.Color("12")),
		tierC:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
		tierDF: lipgloss.NewStyle().Foreground(lipgloss.Color("9")),
		dim:    lipgloss.NewStyle().Foreground(lipgloss.Color("8")),
	}
}

func printSummaryReport(summary *ComponentSummary) {
	styles := newPrintStyles()

	printReportHeader(styles)
	printComponentCounts(summary, styles)
	printQualityDistribution(summary, styles)
	printTopIssues(summary, styles)
	printLowestScoring(summary, styles)
	printReportFooter(styles)
}

func printReportHeader(styles printStyles) {
	fmt.Println()
	fmt.Println(styles.header.Render("╔═══════════════════════════════════════════════════════════╗"))
	fmt.Println(styles.header.Render("║              COMPONENT QUALITY SUMMARY                     ║"))
	fmt.Println(styles.header.Render("╠═══════════════════════════════════════════════════════════╣"))
}

func printComponentCounts(summary *ComponentSummary, styles printStyles) {
	fmt.Printf("║ Components Analyzed: %-37d ║\n", summary.TotalComponents)
	fmt.Printf("║   Agents: %-5d │ Commands: %-5d │ Skills: %-13d ║\n",
		summary.AgentCount, summary.CommandCount, summary.SkillCount)
}

func printQualityDistribution(summary *ComponentSummary, styles printStyles) {
	fmt.Println(styles.header.Render("╠───────────────────────────────────────────────────────────╣"))
	fmt.Println("║ QUALITY DISTRIBUTION                                      ║")

	total := float64(summary.TotalComponents)
	if total == 0 {
		total = 1
	}

	aCount := summary.TierCounts["A"]
	bCount := summary.TierCounts["B"]
	cCount := summary.TierCounts["C"]
	dfCount := summary.TierCounts["D"] + summary.TierCounts["F"]

	fmt.Printf("║   %s: %-4d (%5.1f%%)  %s                              ║\n",
		styles.tierA.Render("A (85-100)"), aCount, float64(aCount)/total*100,
		renderBar(aCount, summary.TotalComponents, "10"))
	fmt.Printf("║   %s: %-4d (%5.1f%%)  %s                              ║\n",
		styles.tierB.Render("B (70-84) "), bCount, float64(bCount)/total*100,
		renderBar(bCount, summary.TotalComponents, "12"))
	fmt.Printf("║   %s: %-4d (%5.1f%%)  %s                              ║\n",
		styles.tierC.Render("C (50-69) "), cCount, float64(cCount)/total*100,
		renderBar(cCount, summary.TotalComponents, "3"))
	fmt.Printf("║   %s: %-4d (%5.1f%%)  %s                              ║\n",
		styles.tierDF.Render("D/F (<50) "), dfCount, float64(dfCount)/total*100,
		renderBar(dfCount, summary.TotalComponents, "9"))
}

type issueCount struct {
	issue string
	count int
}

func printTopIssues(summary *ComponentSummary, styles printStyles) {
	fmt.Println(styles.header.Render("╠───────────────────────────────────────────────────────────╣"))
	fmt.Println("║ TOP ISSUES                                                ║")

	var issues []issueCount
	for issue, count := range summary.TopIssues {
		issues = append(issues, issueCount{issue, count})
	}
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].count > issues[j].count
	})

	for i, ic := range issues {
		if i >= 5 {
			break
		}
		truncated := ic.issue
		if len(truncated) > 40 {
			truncated = truncated[:37] + "..."
		}
		fmt.Printf("║   %s: %-40s %3d ║\n", styles.dim.Render(fmt.Sprintf("%d.", i+1)), truncated, ic.count)
	}
}

func printLowestScoring(summary *ComponentSummary, styles printStyles) {
	fmt.Println(styles.header.Render("╠───────────────────────────────────────────────────────────╣"))
	fmt.Println("║ LOWEST SCORING COMPONENTS                                 ║")

	for i, comp := range summary.LowestScoring {
		if i >= 5 {
			break
		}
		tierStyle := styles.tierDF
		if comp.Tier == "C" {
			tierStyle = styles.tierC
		}
		truncated := comp.File
		if len(truncated) > 35 {
			truncated = "..." + truncated[len(truncated)-32:]
		}
		fmt.Printf("║   %s %-35s %s %2d ║\n",
			styles.dim.Render(fmt.Sprintf("%d.", i+1)),
			truncated,
			tierStyle.Render(comp.Tier),
			comp.Score)
	}
}

func printReportFooter(styles printStyles) {
	fmt.Println(styles.header.Render("╚═══════════════════════════════════════════════════════════╝"))
	fmt.Println()
}

func renderBar(count, total int, color string) string {
	if total == 0 {
		return ""
	}
	barWidth := 10
	filled := (count * barWidth) / total
	if count > 0 && filled == 0 {
		filled = 1
	}
	bar := ""
	style := lipgloss.NewStyle().Foreground(lipgloss.Color(color))
	for i := 0; i < filled; i++ {
		bar += style.Render("█")
	}
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	for i := filled; i < barWidth; i++ {
		bar += dimStyle.Render("░")
	}
	return bar
}
