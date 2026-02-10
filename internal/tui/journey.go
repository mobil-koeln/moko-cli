package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/mobil-koeln/moko-cli/internal/output"
)

// renderJourneyDetail renders journey stops with route symbols.
func (m Model) renderJourneyDetail(width, height int) string {
	title := "JOURNEY"
	if m.journey != nil {
		title += ": " + m.journey.Name
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

	currentIdx := output.FindCurrentStopIndex(stops, time.Now())

	var b strings.Builder
	b.WriteString(titleStr)
	b.WriteString("\n")

	maxVisible := height - 2
	if maxVisible < 1 {
		maxVisible = 1
	}
	start, end := visibleRange(m.journeyScroll, len(stops), maxVisible)

	for i := start; i < end; i++ {
		stop := stops[i]
		isFirst := i == 0
		isLast := i == len(stops)-1
		isCurrent := i == currentIdx

		// Route symbol
		symbol := "├"
		if isFirst {
			symbol = "┌"
		} else if isLast {
			symbol = "└"
		}

		// Current indicator
		indicator := " "
		if isCurrent {
			indicator = ">"
		}

		// Time
		timeStr := "     "
		if stop.Arr != nil && !isFirst {
			timeStr = stop.Arr.Format("15:04")
		} else if stop.Dep != nil && isFirst {
			timeStr = stop.Dep.Format("15:04")
		}

		// Delay
		delayStr := "    "
		if stop.Delay != 0 {
			delayStr = formatDelay(stop.Delay)
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

		// Station name
		name := stop.Name
		fixedWidth := 1 + 1 + 1 + 1 + 5 + 1 + 4 + 2 + 7 // indicator+sp+symbol+sp+time+sp+delay+sp+platform
		maxName := width - fixedWidth - 2
		if maxName > 0 && len(name) > maxName {
			name = name[:maxName]
		}

		if isCurrent && !stop.IsCancelled {
			b.WriteString(fmt.Sprintf("%s %s %s %s  %s %s",
				styleCanceled.Render(indicator),
				styleMuted.Render(symbol),
				styleCanceled.Render(timeStr),
				delayStr,
				styleCanceled.Render(platformStr),
				styleCanceled.Render(name),
			))
		} else if stop.IsCancelled {
			b.WriteString(fmt.Sprintf("%s %s %s %s  %s %s",
				indicator,
				styleMuted.Render(symbol),
				styleCanceled.Render(timeStr),
				delayStr,
				styleCanceled.Render(platformStr),
				styleCanceled.Render(name+" [X]"),
			))
		} else {
			b.WriteString(fmt.Sprintf("%s %s %s %s  %s %s",
				indicator,
				styleMuted.Render(symbol),
				styleTime.Render(timeStr),
				delayStr,
				stylePlatform.Render(platformStr),
				name,
			))
		}

		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
