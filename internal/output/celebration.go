package output

import (
	"fmt"
	"time"

	"charm.land/lipgloss/v2"
)

// printCelebration shows a sparkle animation for perfect success.
// This is a package-level helper to avoid duplication across formatters.
func printCelebration(msg string) {
	green := lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
	bold := lipgloss.NewStyle().Foreground(lipgloss.Color("10")).Bold(true)
	yellow := lipgloss.NewStyle().Foreground(lipgloss.Color("11")).Bold(true)

	frames := []struct {
		text  string
		delay time.Duration
	}{
		{green.Render(msg), 200 * time.Millisecond},
		{yellow.Render("✨ " + msg + " ✨"), 300 * time.Millisecond},
		{bold.Render("🎉 " + msg + " 🎉"), 400 * time.Millisecond},
		{yellow.Render("✨ " + msg + " ✨"), 300 * time.Millisecond},
		{green.Render(msg), 0},
	}

	for i, frame := range frames {
		if i > 0 {
			fmt.Print("\r\033[K")
		}
		fmt.Print(frame.text)
		if frame.delay > 0 {
			time.Sleep(frame.delay)
		}
	}
	fmt.Println()
}
