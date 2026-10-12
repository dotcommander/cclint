package cmd

import (
	"fmt"
	"sort"

	"github.com/charmbracelet/lipgloss"
)

const summaryTopBorder = "╔═══════════════════════════════════════════════════════════╗"

// renderSummaryRow pads by terminal cells, not bytes or ANSI sequence length.
// Its width follows the existing border; long content cannot move the edge.
func renderSummaryRow(content string) string {
	width := lipgloss.Width(summaryTopBorder) - 4
	return "║ " + lipgloss.NewStyle().Width(width).MaxWidth(width).MaxHeight(1).Render(content) + " ║"
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
	fmt.Println(styles.header.Render(summaryTopBorder))
	fmt.Println(renderSummaryRow(styles.header.Render(lipgloss.NewStyle().Align(lipgloss.Center).Width(lipgloss.Width(summaryTopBorder) - 4).Render("COMPONENT QUALITY SUMMARY"))))
	fmt.Println(styles.header.Render("╠═══════════════════════════════════════════════════════════╣"))
}

func printComponentCounts(summary *ComponentSummary, styles printStyles) {
	fmt.Println(renderSummaryRow(fmt.Sprintf("Components Analyzed: %d", summary.TotalComponents)))
	fmt.Println(renderSummaryRow(fmt.Sprintf("  Agents: %d │ Commands: %d │ Skills: %d",
		summary.AgentCount, summary.CommandCount, summary.SkillCount)))
}

func printQualityDistribution(summary *ComponentSummary, styles printStyles) {
	fmt.Println(styles.header.Render("╠───────────────────────────────────────────────────────────╣"))
	fmt.Println(renderSummaryRow("QUALITY DISTRIBUTION"))

	total := float64(summary.TotalComponents)
	if total == 0 {
		total = 1
	}

	aCount := summary.TierCounts["A"]
	bCount := summary.TierCounts["B"]
	cCount := summary.TierCounts["C"]
	dfCount := summary.TierCounts["D"] + summary.TierCounts["F"]

	fmt.Println(renderSummaryRow(fmt.Sprintf("  %s: %-4d (%5.1f%%)  %s",
		styles.tierA.Render("A (85-100)"), aCount, float64(aCount)/total*100,
		renderBar(aCount, summary.TotalComponents, "10"))))
	fmt.Println(renderSummaryRow(fmt.Sprintf("  %s: %-4d (%5.1f%%)  %s",
		styles.tierB.Render("B (70-84) "), bCount, float64(bCount)/total*100,
		renderBar(bCount, summary.TotalComponents, "12"))))
	fmt.Println(renderSummaryRow(fmt.Sprintf("  %s: %-4d (%5.1f%%)  %s",
		styles.tierC.Render("C (50-69) "), cCount, float64(cCount)/total*100,
		renderBar(cCount, summary.TotalComponents, "3"))))
	fmt.Println(renderSummaryRow(fmt.Sprintf("  %s: %-4d (%5.1f%%)  %s",
		styles.tierDF.Render("D/F (<50) "), dfCount, float64(dfCount)/total*100,
		renderBar(dfCount, summary.TotalComponents, "9"))))
}

type issueCount struct {
	issue string
	count int
}

func printTopIssues(summary *ComponentSummary, styles printStyles) {
	fmt.Println(styles.header.Render("╠───────────────────────────────────────────────────────────╣"))
	fmt.Println(renderSummaryRow("TOP ISSUES"))

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
		fmt.Println(renderSummaryRow(fmt.Sprintf("  %s: %-40s %3d", styles.dim.Render(fmt.Sprintf("%d.", i+1)), truncated, ic.count)))
	}
}

func printLowestScoring(summary *ComponentSummary, styles printStyles) {
	fmt.Println(styles.header.Render("╠───────────────────────────────────────────────────────────╣"))
	fmt.Println(renderSummaryRow("LOWEST SCORING COMPONENTS"))

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
		fmt.Println(renderSummaryRow(fmt.Sprintf("  %s %-35s %s %2d",
			styles.dim.Render(fmt.Sprintf("%d.", i+1)),
			truncated,
			tierStyle.Render(comp.Tier),
			comp.Score)))
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
