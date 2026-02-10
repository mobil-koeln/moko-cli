package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/mobil-koeln/moko-cli/internal/models"
	"github.com/mobil-koeln/moko-cli/internal/output"
)

// View renders the entire TUI.
func (m Model) View() string {
	if m.width == 0 || m.height == 0 {
		return "Loading..."
	}

	// Layout: header + search bar + filter bar + panels + status bar
	header := renderHeader()
	searchBar := m.renderSearchBar()
	filterBar := m.renderFilterBar()
	statusBar := m.renderStatusBar()

	headerHeight := lipgloss.Height(header)
	searchHeight := lipgloss.Height(searchBar)
	filterHeight := lipgloss.Height(filterBar)
	statusHeight := lipgloss.Height(statusBar)
	panelHeight := m.height - headerHeight - searchHeight - filterHeight - statusHeight
	if panelHeight < 3 {
		panelHeight = 3
	}

	// Panel widths: ~35% left, ~65% right
	leftWidth := m.width*35/100 - 2 // subtract border
	rightWidth := m.width - leftWidth - 4
	if leftWidth < 20 {
		leftWidth = 20
	}
	if rightWidth < 20 {
		rightWidth = 20
	}

	leftPanel := m.renderStationList(leftWidth, panelHeight-2)
	rightPanel := m.renderRightPanel(rightWidth, panelHeight-2)

	// Apply borders
	leftBorder := stylePanelNormal
	if m.focus == focusStations {
		leftBorder = stylePanelFocused
	}
	leftPanel = leftBorder.
		Width(leftWidth).
		Height(panelHeight - 2).
		Render(leftPanel)

	rightBorder := stylePanelNormal
	if m.focus == focusDepartures || m.focus == focusJourney {
		rightBorder = stylePanelFocused
	}
	rightPanel = rightBorder.
		Width(rightWidth).
		Height(panelHeight - 2).
		Render(rightPanel)

	panels := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, rightPanel)

	return lipgloss.JoinVertical(lipgloss.Left, header, searchBar, filterBar, panels, statusBar)
}

// renderHeader renders the ASCII logo and brand name.
func renderHeader() string {
	logo := "" +
		" /\\  /\\ \n" +
		"/ O\\/ O\\\n" +
		"\\ |\\/ |/\n" +
		"  O  O  "

	title := "" +
		"          _    _ _   _            _      \n" +
		" _ __  ___| |__(_) | | |_____  ___| |_ _  \n" +
		"| '  \\/ _ \\ '_ \\ | |_| / / _ \\/ -_) | ' \\ \n" +
		"|_|_|_\\___/_.__/_|_(_)_\\_\\___/\\___|_|_||_|"

	styledLogo := styleLogo.Render(logo)
	styledTitle := styleLogo.Render(title)

	return lipgloss.JoinHorizontal(lipgloss.Bottom, styledLogo, "  ", styledTitle)
}

// renderSearchBar renders the search input at the top.
func (m Model) renderSearchBar() string {
	border := stylePanelNormal
	if m.focus == focusSearch {
		border = stylePanelFocused
	}

	label := styleHeader.Render("Search: ")
	input := m.searchInput.View()
	content := label + input

	return border.Width(m.width - 2).Render(content)
}

// renderStationList renders the left station panel.
func (m Model) renderStationList(width, height int) string {
	title := styleHeader.Render("STATIONS")

	if m.stationsLoading {
		return title + "\n" + styleLoading.Render(" Searching...")
	}
	if m.stationsErr != nil {
		return title + "\n" + styleError.Render(" Error: "+m.stationsErr.Error())
	}
	if len(m.stations) == 0 {
		return title + "\n" + styleMuted.Render(" Type a station name and press Enter")
	}

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n")

	// Calculate visible range to keep cursor in view
	maxVisible := height - 2 // account for title + spacing
	if maxVisible < 1 {
		maxVisible = 1
	}
	start, end := visibleRange(m.stationCursor, len(m.stations), maxVisible)

	for i := start; i < end; i++ {
		station := m.stations[i]
		name := truncate(station.Name, width-4)
		if i == m.stationCursor {
			b.WriteString(styleSelected.Render(" > " + name))
		} else {
			b.WriteString("   " + name)
		}
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderRightPanel renders departures and optionally journey details with route map.
func (m Model) renderRightPanel(width, height int) string {
	if m.showJourney && m.journey != nil {
		// Split: top 45% departures, bottom 55% journey+map side by side
		depHeight := height * 45 / 100
		bottomHeight := height - depHeight - 1 // -1 for separator
		if depHeight < 4 {
			depHeight = 4
		}
		if bottomHeight < 4 {
			bottomHeight = 4
		}

		depView := m.renderDepartureList(width, depHeight)
		separator := styleMuted.Render(strings.Repeat("─", width))

		// Split bottom area: journey on left ~55%, route map on right ~45%
		journeyWidth := width * 55 / 100
		mapWidth := width - journeyWidth - 1 // -1 for vertical separator
		if journeyWidth < 20 {
			journeyWidth = 20
		}
		if mapWidth < 10 {
			mapWidth = 10
		}

		journeyView := m.renderJourneyDetail(journeyWidth, bottomHeight)
		currentIdx := output.FindCurrentStopIndex(m.journey.Stops, time.Now())
		mapView := renderRouteMap(m.journey.Stops, currentIdx, mapWidth, bottomHeight)

		// Use lipgloss to enforce fixed-width columns for side-by-side layout.
		// This correctly handles ANSI escape codes in styled text.
		journeyBox := lipgloss.NewStyle().
			Width(journeyWidth).
			Height(bottomHeight).
			Render(journeyView)
		mapBox := lipgloss.NewStyle().
			Width(mapWidth).
			Height(bottomHeight).
			Render(mapView)

		vSep := styleMuted.Render(strings.Repeat("│\n", bottomHeight-1) + "│")

		bottomView := lipgloss.JoinHorizontal(lipgloss.Top, journeyBox, vSep, mapBox)

		return depView + "\n" + separator + "\n" + bottomView
	}

	return m.renderDepartureList(width, height)
}

// renderDepartureList renders the departure table.
func (m Model) renderDepartureList(width, height int) string {
	title := "DEPARTURES"
	if m.boardMode == boardArrival {
		title = "ARRIVALS"
	}
	if m.selectedStation != nil {
		title += " for " + truncate(m.selectedStation.Name, width-18)
	}
	titleStr := styleHeader.Render(title)

	if m.departuresLoading {
		return titleStr + "\n" + styleLoading.Render(" Loading departures...")
	}
	if m.departuresErr != nil {
		return titleStr + "\n" + styleError.Render(" Error: "+m.departuresErr.Error())
	}
	if m.selectedStation == nil {
		return titleStr + "\n" + styleMuted.Render(" Select a station to view departures")
	}
	if len(m.departures) == 0 {
		return titleStr + "\n" + styleMuted.Render(" No departures found")
	}

	var b strings.Builder
	b.WriteString(titleStr)
	b.WriteString("\n")

	maxVisible := height - 2
	if maxVisible < 1 {
		maxVisible = 1
	}
	start, end := visibleRange(m.departureCursor, len(m.departures), maxVisible)

	for i := start; i < end; i++ {
		dep := m.departures[i]
		line := renderDepartureLine(dep, width, i == m.departureCursor && m.focus == focusDepartures)
		b.WriteString(line)
		if i < end-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}

// renderDepartureLine renders a single departure entry.
func renderDepartureLine(dep models.Departure, width int, selected bool) string {
	// Time
	timeStr := "??:??"
	if dep.Dep != nil {
		timeStr = dep.Dep.Format("15:04")
	}

	// Delay
	delayStr := formatDelay(dep.Delay)

	// Line name (truncate to 10)
	line := dep.Line
	if line == "" {
		line = dep.TrainShort
	}
	if len(line) > 10 {
		line = line[:10]
	}
	lineStr := fmt.Sprintf("%-10s", line)

	// Platform
	platform := dep.EffectivePlatform()
	platformStr := "       "
	if platform != "" {
		if len(platform) > 3 {
			platform = platform[:3]
		}
		platformStr = fmt.Sprintf("Pl.%-3s ", platform)
	}

	// Destination
	dest := dep.Destination
	// Calculate remaining width for destination
	fixedWidth := 5 + 1 + 4 + 2 + 10 + 2 + 7 // time+sp+delay+sp+line+sp+platform
	maxDest := width - fixedWidth - 4        // 4 for cursor indicator + padding
	if maxDest > 0 && len(dest) > maxDest {
		dest = dest[:maxDest]
	}

	var entry string
	if dep.IsCancelled {
		entry = fmt.Sprintf("%s %s  %s  %s %s",
			styleTime.Render(timeStr),
			delayStr,
			styleCanceled.Render(lineStr),
			stylePlatform.Render(platformStr),
			styleCanceled.Render(dest+" [X]"),
		)
	} else {
		entry = fmt.Sprintf("%s %s  %s  %s %s",
			styleTime.Render(timeStr),
			delayStr,
			styleLine.Render(lineStr),
			stylePlatform.Render(platformStr),
			dest,
		)
	}

	if selected {
		return styleSelected.Render(">") + entry
	}
	return " " + entry
}

// renderStatusBar renders context-aware keyboard hints at the bottom.
func (m Model) renderStatusBar() string {
	var hints string
	switch m.focus {
	case focusSearch:
		hints = "Enter:search  Tab:filters  Esc:clear  Ctrl+C:quit"
	case focusFilters:
		hints = "h/l:move  Space:toggle  a:all  Tab:dep/arr  Esc:search  q:quit"
	case focusBoard:
		hints = "h/l:move  Space:select  Tab:auto-refresh  Esc:search  q:quit"
	case focusAutoRefresh:
		hints = "Space:toggle  Tab:stations  Esc:search  q:quit"
	case focusStations:
		hints = "j/k:navigate  Enter:select  Tab:departures  Esc:search  /:search  q:quit"
	case focusDepartures:
		hints = "j/k:navigate  Enter:journey  Tab:next  Esc:back  /:search  q:quit"
	case focusJourney:
		hints = "j/k:scroll  Tab:search  Esc:departures  q:quit"
	}

	return styleStatusBar.Width(m.width).Render(" " + hints)
}

// visibleRange calculates the start and end indices for a scrollable list.
func visibleRange(cursor, total, maxVisible int) (int, int) {
	if total <= maxVisible {
		return 0, total
	}

	start := cursor - maxVisible/2
	if start < 0 {
		start = 0
	}
	end := start + maxVisible
	if end > total {
		end = total
		start = end - maxVisible
		if start < 0 {
			start = 0
		}
	}
	return start, end
}

// truncate truncates a string to the given width.
func truncate(s string, width int) string {
	if width <= 0 {
		return ""
	}
	if len(s) <= width {
		return s
	}
	if width <= 3 {
		return s[:width]
	}
	return s[:width-1] + "~"
}
