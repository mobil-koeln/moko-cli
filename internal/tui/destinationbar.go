package tui

import (
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mobil-koeln/moko-cli/internal/models"
)

// renderDestinationPanel renders the vertical destination filter panel.
func (m Model) renderDestinationPanel(width, height int) string {
	title := "DESTINATIONS"
	if m.focus == focusDestinations {
		title = "▶ " + title
	}
	titleStr := styleHeader.Render(title)

	if len(m.destinationList) == 0 {
		empty := styleMuted.Render(" No data")
		return titleStr + "\n" + empty
	}

	// Reserve space for scrollbar
	contentWidth := width - 2

	maxVisible := height - 2
	if maxVisible < 1 {
		maxVisible = 1
	}
	start, end := visibleRange(m.destinationCursor, len(m.destinationList), maxVisible)

	var contentLines []string
	for i := start; i < end; i++ {
		dest := m.destinationList[i]
		focused := m.focus == focusDestinations && m.destinationCursor == i
		active := i < len(m.destinationFilters) && m.destinationFilters[i]
		chip := m.renderChip(truncate(dest, contentWidth-3), active, focused)
		contentLines = append(contentLines, chip)
	}

	for len(contentLines) < maxVisible {
		contentLines = append(contentLines, "")
	}

	scrollbarHeight := maxVisible
	scrollbar := renderScrollbar(m.destinationCursor, len(m.destinationList), scrollbarHeight)
	scrollbarLines := strings.Split(scrollbar, "\n")

	var b strings.Builder
	for i := 0; i < len(contentLines); i++ {
		line := contentLines[i]
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

// handleDestinationKeys handles key events when the destination panel is focused.
func (m Model) handleDestinationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "j", "down":
		if m.destinationCursor < len(m.destinationList)-1 {
			m.destinationCursor++
		}
		return m, nil

	case "k", "up":
		if m.destinationCursor > 0 {
			m.destinationCursor--
		}
		return m, nil

	case "pgdown":
		if len(m.destinationList) > 0 {
			pageSize := m.height - 10
			if pageSize < 1 {
				pageSize = 10
			}
			m.destinationCursor += pageSize
			if m.destinationCursor >= len(m.destinationList) {
				m.destinationCursor = len(m.destinationList) - 1
			}
		}
		return m, nil

	case "pgup":
		if len(m.destinationList) > 0 {
			pageSize := m.height - 10
			if pageSize < 1 {
				pageSize = 10
			}
			m.destinationCursor -= pageSize
			if m.destinationCursor < 0 {
				m.destinationCursor = 0
			}
		}
		return m, nil

	case "home":
		m.destinationCursor = 0
		return m, nil

	case "end":
		if len(m.destinationList) > 0 {
			m.destinationCursor = len(m.destinationList) - 1
		}
		return m, nil

	case " ", "enter":
		if m.destinationCursor < len(m.destinationFilters) {
			m.destinationFilters[m.destinationCursor] = !m.destinationFilters[m.destinationCursor]
		}
		return m, nil

	case "a":
		anyOff := false
		for _, f := range m.destinationFilters {
			if !f {
				anyOff = true
				break
			}
		}
		for i := range m.destinationFilters {
			m.destinationFilters[i] = anyOff
		}
		return m, nil

	case "tab":
		if m.showJourney {
			m.focus = focusJourney
		} else {
			m.focus = focusSearch
			m.searchInput.Focus()
		}
		return m, nil

	case "shift+tab":
		m.focus = focusDepartures
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

// rebuildDestinationList extracts unique destinations from departures, sorts them,
// and preserves existing toggle states.
func (m Model) rebuildDestinationList() Model {
	seen := make(map[string]bool)
	var newList []string
	for _, dep := range m.departures {
		if dep.Destination != "" && !seen[dep.Destination] {
			seen[dep.Destination] = true
			newList = append(newList, dep.Destination)
		}
	}
	sort.Strings(newList)

	prevStates := make(map[string]bool)
	for i, dest := range m.destinationList {
		if i < len(m.destinationFilters) {
			prevStates[dest] = m.destinationFilters[i]
		}
	}
	newFilters := make([]bool, len(newList))
	for i, dest := range newList {
		if prev, exists := prevStates[dest]; exists {
			newFilters[i] = prev
		} else {
			newFilters[i] = true
		}
	}

	m.destinationList = newList
	m.destinationFilters = newFilters
	if len(newList) == 0 {
		m.destinationCursor = 0
	} else if m.destinationCursor >= len(newList) {
		m.destinationCursor = len(newList) - 1
	}
	return m
}

// filteredDepartures returns departures filtered by active destinations.
// If all destinations are active (or list is empty), returns all departures.
func (m Model) filteredDepartures() []models.Departure {
	if len(m.destinationList) == 0 {
		return m.departures
	}
	allActive := true
	for _, f := range m.destinationFilters {
		if !f {
			allActive = false
			break
		}
	}
	if allActive {
		return m.departures
	}
	active := make(map[string]bool, len(m.destinationList))
	for i, dest := range m.destinationList {
		if i < len(m.destinationFilters) && m.destinationFilters[i] {
			active[dest] = true
		}
	}
	var result []models.Departure
	for _, dep := range m.departures {
		if active[dep.Destination] {
			result = append(result, dep)
		}
	}
	return result
}
