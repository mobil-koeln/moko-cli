package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mobil-koeln/moko-cli/internal/output"
)

// renderJourneyDetail renders journey stops with route symbols.
func (m Model) renderJourneyDetail(width, height int) string {
	title := "JOURNEY"
	if m.journey != nil {
		title += ": " + m.journey.Name
	}
	if m.focus == focusJourney {
		title = "▶ " + title // Add indicator when focused
	}
	titleStr := styleHeader.Render(title)

	if m.journeyLoading {
		return titleStr + "\n" + styleLoading.Render(" Loading journey...")
	}
	if m.journeyErr != nil {
		return titleStr + "\n" + styleError.Render(" Error: "+m.journeyErr.Error())
	}
	if m.journey == nil {
		return titleStr + "\n" + styleMuted.Render(" Select a departure to view journey")
	}

	stops := m.journey.Stops
	if len(stops) == 0 {
		return titleStr + "\n" + styleMuted.Render(" No stops")
	}

	// Reserve space for scrollbar
	contentWidth := width - 2

	currentIdx := output.FindCurrentStopIndex(stops, time.Now())

	maxVisible := height - 2
	if maxVisible < 1 {
		maxVisible = 1
	}
	start, end := visibleRange(m.journeyScroll, len(stops), maxVisible)

	// Build content lines
	var contentLines []string
	for i := start; i < end; i++ {
		stop := stops[i]
		isFirst := i == 0
		isLast := i == len(stops)-1
		isCurrent := i == currentIdx
		isScrolledTo := i == m.journeyScroll // User's scroll position

		// Route symbol
		symbol := "├"
		if isFirst {
			symbol = "┌"
		} else if isLast {
			symbol = "└"
		}

		// Indicator: show scroll position when journey is visible, current stop otherwise
		indicator := " "
		if isScrolledTo && m.showJourney {
			indicator = "►" // Show scroll position when journey is visible
		} else if isCurrent && !isScrolledTo {
			indicator = "●" // Show current time-based stop with different symbol
		}

		// Time
		timeStr := "     "
		if stop.Arr != nil && !isFirst {
			timeStr = stop.Arr.Format("15:04")
		} else if stop.Dep != nil && isFirst {
			timeStr = stop.Dep.Format("15:04")
		}

		// Delay - format as plain text for width calculation
		var delayPlain string
		if stop.Delay == 0 {
			delayPlain = "    "
		} else if stop.Delay > 0 {
			delayPlain = fmt.Sprintf("%+4d", stop.Delay)
		} else {
			delayPlain = fmt.Sprintf("%4d", stop.Delay)
		}

		// Platform
		platform := stop.EffectivePlatform()
		platformStr := "       "
		if platform != "" {
			if len(platform) > 3 {
				platform = platform[:3]
			}
			platformStr = fmt.Sprintf("Pl.%-3s ", platform)
		}

		// Station name - pad to fill full width for consistent highlighting
		name := stop.Name
		fixedWidth := 1 + 1 + 1 + 1 + 5 + 1 + 4 + 2 + 7 // indicator+sp+symbol+sp+time+sp+delay+sp+platform
		maxName := contentWidth - fixedWidth - 2

		// Reserve space for [X] if cancelled
		if stop.IsCancelled {
			maxName -= 4 // Reserve 4 chars for " [X]"
		}

		if maxName > 0 {
			if len(name) > maxName {
				name = name[:maxName]
			} else {
				// Pad with spaces to fill the full width
				name = name + strings.Repeat(" ", maxName-len(name))
			}
		}

		// Build the line content with PLAIN TEXT (no ANSI codes) for proper width calculation
		var lineContent string
		if stop.IsCancelled {
			lineContent = fmt.Sprintf("%s %s %s %s  %s %s",
				indicator,
				symbol,
				timeStr,
				delayPlain, // Use plain text delay
				platformStr,
				name+" [X]",
			)
		} else {
			lineContent = fmt.Sprintf("%s %s %s %s  %s %s",
				indicator,
				symbol,
				timeStr,
				delayPlain, // Use plain text delay
				platformStr,
				name,
			)
		}

		// Apply full-width highlight based on state
		var line string
		if isScrolledTo && m.showJourney && !isCurrent {
			// Cyan highlight for scroll position (full width)
			line = styleSelected.Width(contentWidth).Render(lineContent)
		} else if isCurrent && !stop.IsCancelled {
			// Red highlight for current stop (full width)
			line = styleCurrentStop.Width(contentWidth).Render(lineContent)
		} else if stop.IsCancelled {
			// No background, just colored text for cancelled
			// Get styled delay for non-highlighted rows
			delayStyled := "    "
			if stop.Delay != 0 {
				delayStyled = formatDelay(stop.Delay)
			}
			lineContent = fmt.Sprintf("%s %s %s %s  %s %s",
				indicator,
				styleMuted.Render(symbol),
				styleCanceled.Render(timeStr),
				delayStyled,
				styleCanceled.Render(platformStr),
				styleCanceled.Render(name+" [X]"),
			)
			line = lineContent
		} else {
			// Normal rendering with colored text
			// Get styled delay for non-highlighted rows
			delayStyled := "    "
			if stop.Delay != 0 {
				delayStyled = formatDelay(stop.Delay)
			}
			lineContent = fmt.Sprintf("%s %s %s %s  %s %s",
				indicator,
				styleMuted.Render(symbol),
				styleTime.Render(timeStr),
				delayStyled,
				stylePlatform.Render(platformStr),
				name,
			)
			line = lineContent
		}

		contentLines = append(contentLines, line)
	}

	// Pad content lines to match maxVisible height
	for len(contentLines) < maxVisible {
		contentLines = append(contentLines, "")
	}

	// Render scrollbar
	scrollbar := renderScrollbar(m.journeyScroll, len(stops), maxVisible)
	scrollbarLines := strings.Split(scrollbar, "\n")

	// Combine content and scrollbar
	var b strings.Builder
	for i := 0; i < len(contentLines); i++ {
		line := contentLines[i]
		// Pad line to contentWidth
		lineWidth := lipgloss.Width(line)
		if lineWidth < contentWidth {
			line += strings.Repeat(" ", contentWidth-lineWidth)
		}
		b.WriteString(line)

		if i < len(scrollbarLines) {
			b.WriteString(" ")
			b.WriteString(scrollbarLines[i])
		}

		if i < len(contentLines)-1 {
			b.WriteString("\n")
		}
	}

	return titleStr + "\n" + b.String()
}
