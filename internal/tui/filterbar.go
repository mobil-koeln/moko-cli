package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// renderFilterBar renders two independent bordered boxes side by side:
// transport mode chips on the left, departure/arrival on the right.
func (m Model) renderFilterBar() string {
	// --- Transport modes box ---
	var modes strings.Builder
	for i, ml := range modeLabels {
		focused := m.focus == focusFilters && m.filterCursor == i
		modes.WriteString(m.renderChip(ml.label, m.modeFilters[i], focused))
		if i < len(modeLabels)-1 {
			modes.WriteString(" ")
		}
	}

	modesBorder := stylePanelNormal
	if m.focus == focusFilters {
		modesBorder = stylePanelFocused
	}
	modesBox := modesBorder.Render(modes.String())

	// --- Departure/Arrival box ---
	var board strings.Builder

	depActive := m.boardMode == boardDeparture
	depFocused := m.focus == focusBoard && m.boardCursor == 0
	board.WriteString(m.renderChip("Departure", depActive, depFocused))
	board.WriteString(" ")

	arrActive := m.boardMode == boardArrival
	arrFocused := m.focus == focusBoard && m.boardCursor == 1
	board.WriteString(m.renderChip("Arrival", arrActive, arrFocused))

	boardBorder := stylePanelNormal
	if m.focus == focusBoard {
		boardBorder = stylePanelFocused
	}
	boardBox := boardBorder.Render(board.String())

	// --- Auto-refresh box ---
	refreshFocused := m.focus == focusAutoRefresh
	refreshChip := m.renderChip("Auto-refresh 30s", m.autoRefresh, refreshFocused)

	refreshBorder := stylePanelNormal
	if refreshFocused {
		refreshBorder = stylePanelFocused
	}
	refreshBox := refreshBorder.Render(refreshChip)

	boxes := lipgloss.JoinHorizontal(lipgloss.Top, modesBox, boardBox, refreshBox)

	// Last update line above the boxes
	if !m.lastUpdate.IsZero() {
		updateText := "  Last update:\t" + m.lastUpdate.Format("15:04:05")

		// Add countdown if auto-refresh is enabled
		if m.autoRefresh {
			elapsed := time.Since(m.lastUpdate)
			remaining := autoRefreshInterval - elapsed
			if remaining < 0 {
				remaining = 0
			}
			seconds := int(remaining.Seconds())
			updateText += fmt.Sprintf("\t(refresh in %ds)", seconds)
		}

		updateLine := styleMuted.Render(updateText)
		return updateLine + "\n" + boxes
	}

	return boxes
}

// renderChip renders a single chip with cursor highlighting.
func (m Model) renderChip(label string, active bool, focused bool) string {
	if focused {
		if active {
			return styleChipCursor.Render("[" + label + "]")
		}
		return styleChipCursor.Render(" " + label + " ")
	}
	if active {
		return styleLine.Render("[" + label + "]")
	}
	return styleMuted.Render(" " + label + " ")
}

// handleFilterKeys handles key events when the transport modes box is focused.
func (m Model) handleFilterKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left":
		if m.filterCursor > 0 {
			m.filterCursor--
		}
		return m, nil

	case "l", "right":
		if m.filterCursor < len(modeLabels)-1 {
			m.filterCursor++
		}
		return m, nil

	case " ", "enter":
		m.modeFilters[m.filterCursor] = !m.modeFilters[m.filterCursor]
		return m.refetchBoard()

	case "a":
		return m.toggleAllModes()

	case "tab":
		m.focus = focusBoard
		return m, nil

	case "esc", "/":
		m.focus = focusSearch
		m.searchInput.Focus()
		return m, nil

	case "q":
		return m, tea.Quit
	}

	return m, nil
}

// handleBoardKeys handles key events when the departure/arrival box is focused.
func (m Model) handleBoardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "h", "left":
		if m.boardCursor > 0 {
			m.boardCursor--
		}
		return m, nil

	case "l", "right":
		if m.boardCursor < 1 {
			m.boardCursor++
		}
		return m, nil

	case " ", "enter":
		if m.boardCursor == 0 {
			m.boardMode = boardDeparture
		} else {
			m.boardMode = boardArrival
		}
		return m.refetchBoard()

	case "tab":
		m.focus = focusAutoRefresh
		return m, nil

	case "esc", "/":
		m.focus = focusSearch
		m.searchInput.Focus()
		return m, nil

	case "q":
		return m, tea.Quit
	}

	return m, nil
}

// toggleAllModes toggles all transport modes on or off.
func (m Model) toggleAllModes() (tea.Model, tea.Cmd) {
	anyOff := false
	for _, active := range m.modeFilters {
		if !active {
			anyOff = true
			break
		}
	}
	for i := range m.modeFilters {
		m.modeFilters[i] = anyOff
	}

	return m.refetchBoard()
}

// refetchBoard re-fetches departures/arrivals if a station is selected.
func (m Model) refetchBoard() (tea.Model, tea.Cmd) {
	if m.selectedStation != nil {
		m.departuresLoading = true
		m.departuresErr = nil
		m.departures = nil
		m.departureCursor = 0
		m.showJourney = false
		m.journey = nil
		return m, fetchBoard(m.client, *m.selectedStation, m.selectedModes(), m.boardMode)
	}
	return m, nil
}
