package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Colors matching existing output/colors.go scheme
var (
	colorCyan    = lipgloss.Color("6")  // Cyan - lines
	colorYellow  = lipgloss.Color("3")  // Yellow - minor delays
	colorRed     = lipgloss.Color("1")  // Red - major delays, canceled
	colorGreen   = lipgloss.Color("2")  // Green - on time
	colorMagenta = lipgloss.Color("5")  // Magenta - platforms
	colorWhite   = lipgloss.Color("15") // White - times, text
	colorGray    = lipgloss.Color("8")  // Gray - muted text
)

// Text styles
var (
	styleTime      = lipgloss.NewStyle().Foreground(colorWhite).Bold(true)
	styleDelay     = lipgloss.NewStyle().Foreground(colorYellow)
	styleDelayHigh = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	styleOnTime    = lipgloss.NewStyle().Foreground(colorGreen)
	styleLine      = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)
	stylePlatform  = lipgloss.NewStyle().Foreground(colorMagenta)
	styleCanceled  = lipgloss.NewStyle().Foreground(colorRed).Bold(true)
	styleMuted     = lipgloss.NewStyle().Foreground(colorGray)
	styleHeader    = lipgloss.NewStyle().Foreground(colorWhite).Bold(true)
)

// Panel border styles
var (
	stylePanelFocused = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorCyan)

	stylePanelNormal = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorGray)
)

// Selected item in a list
var styleSelected = lipgloss.NewStyle().Foreground(colorCyan).Bold(true)

// Current stop highlight (red background)
var styleCurrentStop = lipgloss.NewStyle().
	Foreground(lipgloss.Color("0")). // Black text
	Background(colorRed).            // Red background
	Bold(true)

// Board station highlight (green background)
var styleBoardStation = lipgloss.NewStyle().
	Foreground(lipgloss.Color("0")). // Black text
	Background(colorGreen).          // Green background
	Bold(true)

// Focused chip cursor in the filter bar — reverse-video style
var styleChipCursor = lipgloss.NewStyle().
	Foreground(lipgloss.Color("0")).
	Background(colorCyan).
	Bold(true)

// Status bar at the bottom
var styleStatusBar = lipgloss.NewStyle().
	Foreground(colorGray).
	Background(lipgloss.Color("0"))

// Loading indicator
var styleLoading = lipgloss.NewStyle().Foreground(colorYellow).Italic(true)

// Error text
var styleError = lipgloss.NewStyle().Foreground(colorRed)

// Logo/brand style
var styleLogo = lipgloss.NewStyle().Foreground(colorRed).Bold(true)

// formatDelay returns a styled delay string (4-char width)
func formatDelay(delay int) string {
	if delay == 0 {
		return "    "
	}
	if delay > 0 {
		s := fmt.Sprintf("%+4d", delay)
		if delay >= 10 {
			return styleDelayHigh.Render(s)
		}
		return styleDelay.Render(s)
	}
	s := fmt.Sprintf("%4d", delay)
	return styleOnTime.Render(s)
}
