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

	// Panel widths: ~35% left, rest right
	leftWidth := m.width*35/100 - 2
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
	if m.focus == focusDepartures || m.focus == focusDestinations || m.focus == focusJourney {
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
	title := "STATIONS"
	if m.focus == focusStations {
		title = "▶ " + title // Add indicator when focused
	}
	title = styleHeader.Render(title)

	if m.stationsLoading {
		return title + "\n" + styleLoading.Render(" Searching...")
	}
	if m.stationsErr != nil {
		return title + "\n" + styleError.Render(" Error: "+m.stationsErr.Error())
	}
	if len(m.stations) == 0 {
		return title + "\n" + styleMuted.Render(" Type a station name and press Enter")
	}

	// Reserve space for scrollbar (2 chars: space + scrollbar)
	contentWidth := width - 2

	// Calculate visible range to keep cursor in view
	maxVisible := height - 2 // account for title + spacing
	if maxVisible < 1 {
		maxVisible = 1
	}
	start, end := visibleRange(m.stationCursor, len(m.stations), maxVisible)

	// Build content lines
	var contentLines []string
	for i := start; i < end; i++ {
		station := m.stations[i]
		name := truncate(station.Name, contentWidth-4)
		if i == m.stationCursor {
			contentLines = append(contentLines, styleSelected.Render(" > "+name))
		} else {
			contentLines = append(contentLines, "   "+name)
		}
	}

	// Pad content lines to match maxVisible height
	for len(contentLines) < maxVisible {
		contentLines = append(contentLines, "")
	}

	// Render scrollbar
	scrollbarHeight := maxVisible
	scrollbar := renderScrollbar(m.stationCursor, len(m.stations), scrollbarHeight)
	scrollbarLines := strings.Split(scrollbar, "\n")

	// Combine content and scrollbar side by side
	var b strings.Builder
	for i := 0; i < len(contentLines); i++ {
		line := contentLines[i]
		// Pad line to contentWidth to ensure scrollbar aligns
		lineWidth := lipgloss.Width(line)
		if lineWidth < contentWidth {
			line += strings.Repeat(" ", contentWidth-lineWidth)
		}
		b.WriteString(line)

		if i < len(scrollbarLines) {
			b.WriteString(" ") // Space before scrollbar
			b.WriteString(scrollbarLines[i])
		}

		if i < len(contentLines)-1 {
			b.WriteString("\n")
		}
	}

	return title + "\n" + b.String()
}

// renderRightPanel renders the right panel:
//
//	top row:    departures (left) | destinations (right)
//	bottom row: journey (left) | map (right)  — only when journey is open
func (m Model) renderRightPanel(width, height int) string {
	// Split top row between departures and destinations
	destWidth := width * 28 / 100
	if destWidth < 14 {
		destWidth = 14
	}
	depWidth := width - destWidth - 1 // -1 for vertical separator
	if depWidth < 20 {
		depWidth = 20
	}

	if m.showJourney && m.journey != nil {
		// Top: departures | destinations, bottom: journey | map
		topHeight := height * 45 / 100
		bottomHeight := height - topHeight - 1 // -1 for separator
		if topHeight < 4 {
			topHeight = 4
		}
		if bottomHeight < 4 {
			bottomHeight = 4
		}

		// Top row
		depView := m.renderDepartureList(depWidth, topHeight)
		destView := m.renderDestinationPanel(destWidth, topHeight)
		vSepTop := styleMuted.Render(strings.Repeat("│\n", topHeight-1) + "│")
		depBox := lipgloss.NewStyle().Width(depWidth).Height(topHeight).Render(depView)
		destBox := lipgloss.NewStyle().Width(destWidth).Height(topHeight).Render(destView)
		topRow := lipgloss.JoinHorizontal(lipgloss.Top, depBox, vSepTop, destBox)

		separator := styleMuted.Render(strings.Repeat("─", width))

		// Bottom row: journey | map
		journeyWidth := width * 55 / 100
		mapWidth := width - journeyWidth - 1
		if journeyWidth < 20 {
			journeyWidth = 20
		}
		if mapWidth < 10 {
			mapWidth = 10
		}

		legendHeight := 1
		contentHeight := bottomHeight - legendHeight
		if contentHeight < 3 {
			contentHeight = 3
		}

		journeyView := m.renderJourneyDetail(journeyWidth, contentHeight)
		currentIdx := output.FindCurrentStopIndex(m.journey.Stops, time.Now())
		boardStationIdx := findBoardStationIdx(m.journey.Stops, m.selectedStation)
		mapView := renderRouteMap(m.journey.Stops, currentIdx, m.journeyScroll, boardStationIdx, mapWidth, contentHeight)

		journeyBox := lipgloss.NewStyle().Width(journeyWidth).Height(contentHeight).Render(journeyView)
		mapBox := lipgloss.NewStyle().Width(mapWidth).Height(contentHeight).Render(mapView)
		vSep := styleMuted.Render(strings.Repeat("│\n", contentHeight-1) + "│")
		bottomRow := lipgloss.JoinHorizontal(lipgloss.Top, journeyBox, vSep, mapBox)

		legend := renderJourneyLegend(width)

		return topRow + "\n" + separator + "\n" + bottomRow + "\n" + legend
	}

	// No journey: departures | destinations side by side
	depView := m.renderDepartureList(depWidth, height)
	destView := m.renderDestinationPanel(destWidth, height)
	vSep := styleMuted.Render(strings.Repeat("│\n", height-1) + "│")
	depBox := lipgloss.NewStyle().Width(depWidth).Height(height).Render(depView)
	destBox := lipgloss.NewStyle().Width(destWidth).Height(height).Render(destView)
	return lipgloss.JoinHorizontal(lipgloss.Top, depBox, vSep, destBox)
}

// renderDepartureList renders the departure table.
func (m Model) renderDepartureList(width, height int) string {
	title := "DEPARTURES"
	if m.boardMode == boardArrival {
		title = "ARRIVALS"
	}
	if m.selectedStation != nil {
		title += " for " + truncate(m.selectedStation.Name, width-20)
	}
	if m.focus == focusDepartures {
		title = "▶ " + title // Add indicator when focused
	}
	// Show filter status in title when some destinations are inactive
	if len(m.destinationFilters) > 0 {
		active := 0
		for _, f := range m.destinationFilters {
			if f {
				active++
			}
		}
		if active < len(m.destinationFilters) {
			title += fmt.Sprintf(" (%d/%d dest)", active, len(m.destinationFilters))
		}
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

	deps := m.filteredDepartures()
	if len(deps) == 0 {
		return titleStr + "\n" + styleMuted.Render(" No departures found")
	}

	// Reserve space for scrollbar
	contentWidth := width - 2

	maxVisible := height - 2
	if maxVisible < 1 {
		maxVisible = 1
	}
	start, end := visibleRange(m.departureCursor, len(deps), maxVisible)

	// Build content lines
	var contentLines []string
	for i := start; i < end; i++ {
		dep := deps[i]
		line := renderDepartureLine(dep, contentWidth, i == m.departureCursor && m.focus == focusDepartures)
		contentLines = append(contentLines, line)
	}

	// Pad content lines to match maxVisible height
	for len(contentLines) < maxVisible {
		contentLines = append(contentLines, "")
	}

	// Render scrollbar
	scrollbar := renderScrollbar(m.departureCursor, len(deps), maxVisible)
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
		hints = "Enter:search  Tab:next  Shift+Tab:back  Esc:clear  Ctrl+C:quit"
	case focusFilters:
		hints = "h/l:move  Space:toggle  a:all  Tab:next  Shift+Tab:back  Esc:search  q:quit"
	case focusBoard:
		hints = "h/l:move  Space:select  Tab:next  Shift+Tab:back  Esc:search  q:quit"
	case focusAutoRefresh:
		hints = "Space:toggle  Tab:next  Shift+Tab:back  Esc:search  q:quit"
	case focusStations:
		hints = "j/k:nav  PgUp/PgDn:page  Home/End:jump  Enter:select  Tab/Shift+Tab:nav  /:search  q:quit"
	case focusDepartures:
		hints = "j/k:nav  PgUp/PgDn:page  Home/End:jump  Enter:journey  Tab/Shift+Tab:nav  Esc:back  q:quit"
	case focusDestinations:
		hints = "j/k:nav  Space:toggle  a:all  Tab:next  Shift+Tab:back  Esc:search  q:quit"
	case focusJourney:
		hints = "j/k:scroll  PgUp/PgDn:page  Home/End:jump  Tab/Shift+Tab:nav  Esc:back  q:quit"
	}

	// Add scroll position indicator
	var indicator string
	switch m.focus {
	case focusStations:
		indicator = scrollIndicator(m.stationCursor, len(m.stations))
	case focusDepartures:
		indicator = scrollIndicator(m.departureCursor, len(m.filteredDepartures()))
	case focusDestinations:
		indicator = scrollIndicator(m.destinationCursor, len(m.destinationList))
	case focusJourney:
		if m.journey != nil {
			indicator = scrollIndicator(m.journeyScroll, len(m.journey.Stops))
		}
	}

	statusText := " " + hints
	if indicator != "" {
		statusText += "  │  " + indicator
	}

	return styleStatusBar.Width(m.width).Render(statusText)
}

// scrollIndicator returns a position indicator string (e.g., "5/20").
func scrollIndicator(cursor, total int) string {
	if total == 0 {
		return ""
	}
	return fmt.Sprintf("%d/%d", cursor+1, total) // 1-indexed for user display
}

// renderJourneyLegend renders a one-line colour legend for the journey stop list.
func renderJourneyLegend(width int) string {
	redSquare := styleCurrentStop.Render(" ")
	greenSquare := styleBoardStation.Render(" ")
	legend := " " + redSquare + " Current Station   " + greenSquare + " Journey Station"
	return styleMuted.Width(width).Render(legend)
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

// renderScrollbar renders a vertical scrollbar for a scrollable list.
// cursor: current position (0-indexed)
// total: total number of items
// height: height of the scrollbar in lines
// Returns a string with newlines for each line of the scrollbar.
func renderScrollbar(cursor, total, height int) string {
	if total == 0 || height <= 0 {
		// Empty scrollbar
		var b strings.Builder
		for i := 0; i < height; i++ {
			b.WriteString(" ")
			if i < height-1 {
				b.WriteString("\n")
			}
		}
		return b.String()
	}

	// If all items fit, show full thumb
	if total <= height {
		var b strings.Builder
		for i := 0; i < height; i++ {
			b.WriteString(styleMuted.Render("█"))
			if i < height-1 {
				b.WriteString("\n")
			}
		}
		return b.String()
	}

	// Calculate thumb size (proportional to visible/total ratio)
	visibleRatio := float64(height) / float64(total)
	thumbSize := int(float64(height) * visibleRatio)
	if thumbSize < 1 {
		thumbSize = 1
	}

	// Calculate thumb position
	scrollRatio := float64(cursor) / float64(total-1)
	thumbStart := int(scrollRatio * float64(height-thumbSize))
	if thumbStart < 0 {
		thumbStart = 0
	}
	if thumbStart+thumbSize > height {
		thumbStart = height - thumbSize
	}

	// Build scrollbar
	var b strings.Builder
	for i := 0; i < height; i++ {
		if i >= thumbStart && i < thumbStart+thumbSize {
			b.WriteString(styleSelected.Render("█")) // Thumb
		} else {
			b.WriteString(styleMuted.Render("│")) // Track
		}
		if i < height-1 {
			b.WriteString("\n")
		}
	}

	return b.String()
}
