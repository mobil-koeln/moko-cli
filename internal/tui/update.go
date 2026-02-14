package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mobil-koeln/moko-cli/internal/output"
)

// Update handles all messages and key events.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case searchResultMsg:
		return m.handleSearchResult(msg)

	case departuresResultMsg:
		return m.handleDeparturesResult(msg)

	case journeyResultMsg:
		return m.handleJourneyResult(msg)

	case autoRefreshTickMsg:
		return m.handleAutoRefreshTick()

	case countdownTickMsg:
		return m.handleCountdownTick()

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Pass remaining messages to textinput when focused
	if m.focus == focusSearch {
		var cmd tea.Cmd
		m.searchInput, cmd = m.searchInput.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleSearchResult(msg searchResultMsg) (tea.Model, tea.Cmd) {
	// Ignore stale results
	if msg.seq != m.searchSeq {
		return m, nil
	}
	m.stationsLoading = false
	m.stationsErr = msg.err
	if msg.err != nil {
		return m, nil
	}

	m.stations = msg.locations
	m.stationCursor = 0

	// Auto-select first station and fetch departures
	if len(m.stations) > 0 {
		m.focus = focusStations
		m.searchInput.Blur()
		station := m.stations[0]
		m.selectedStation = &station
		m.departuresLoading = true
		m.departuresErr = nil
		m.departures = nil
		m.departureCursor = 0
		m.showJourney = false
		return m, fetchBoard(m.client, station, m.selectedModes(), m.boardMode)
	}

	return m, nil
}

func (m Model) handleDeparturesResult(msg departuresResultMsg) (tea.Model, tea.Cmd) {
	// Ignore if station changed
	if m.selectedStation == nil || msg.stationEVA != m.selectedStation.EVA {
		return m, nil
	}
	m.departuresLoading = false
	m.departuresErr = msg.err
	if msg.err == nil {
		hadData := len(m.departures) > 0
		m.departures = msg.departures
		if hadData && m.selectedJourneyID != "" {
			// Re-locate the selected journey in the refreshed list
			found := false
			for i, dep := range m.departures {
				if dep.JourneyID == m.selectedJourneyID {
					m.departureCursor = i
					found = true
					break
				}
			}
			if !found {
				// Journey left the board — close the journey view
				m.showJourney = false
				m.journey = nil
				m.selectedJourneyID = ""
			}
		} else if !hadData {
			m.departureCursor = 0
		}
		// Clamp cursor if list shrank
		if m.departureCursor >= len(m.departures) && len(m.departures) > 0 {
			m.departureCursor = len(m.departures) - 1
		}
		m.lastUpdate = time.Now()
	}
	return m, nil
}

func (m Model) handleJourneyResult(msg journeyResultMsg) (tea.Model, tea.Cmd) {
	m.journeyLoading = false
	m.journeyErr = msg.err
	if msg.err == nil {
		wasShowing := m.showJourney && m.journey != nil
		m.journey = msg.journey
		m.showJourney = true

		// ALWAYS clamp scroll position after setting journey (strengthened reactive clamping)
		if m.journey != nil && len(m.journey.Stops) > 0 {
			if m.journeyScroll >= len(m.journey.Stops) {
				m.journeyScroll = len(m.journey.Stops) - 1
			}
			if m.journeyScroll < 0 {
				m.journeyScroll = 0
			}
		}

		if wasShowing && m.journeyManualScroll {
			// User manually scrolled — position already clamped above
		} else {
			// New journey or no manual scroll — auto-scroll to current station.
			m.journeyManualScroll = false
			currentIdx := output.FindCurrentStopIndex(m.journey.Stops, time.Now())
			if currentIdx >= 0 {
				m.journeyScroll = currentIdx
			} else {
				m.journeyScroll = 0
			}
		}
	}
	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Global keys
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	}

	switch m.focus {
	case focusSearch:
		return m.handleSearchKeys(msg)
	case focusFilters:
		return m.handleFilterKeys(msg)
	case focusBoard:
		return m.handleBoardKeys(msg)
	case focusAutoRefresh:
		return m.handleAutoRefreshKeys(msg)
	case focusStations:
		return m.handleStationKeys(msg)
	case focusDepartures:
		return m.handleDepartureKeys(msg)
	case focusJourney:
		return m.handleJourneyKeys(msg)
	}

	return m, nil
}

func (m Model) handleSearchKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		query := strings.TrimSpace(m.searchInput.Value())
		if query == "" {
			return m, nil
		}
		m.searchSeq++
		m.stationsLoading = true
		m.stationsErr = nil
		return m, searchStations(m.client, query, m.searchSeq)

	case "esc":
		m.searchInput.SetValue("")
		return m, nil

	case "tab":
		m.focus = focusFilters
		m.searchInput.Blur()
		return m, nil

	case "shift+tab":
		// Navigate backward to last available panel
		if m.showJourney {
			m.focus = focusJourney
		} else if len(m.departures) > 0 {
			m.focus = focusDepartures
		} else if len(m.stations) > 0 {
			m.focus = focusStations
		} else {
			m.focus = focusAutoRefresh
		}
		m.searchInput.Blur()
		return m, nil
	}

	// Forward to textinput
	var cmd tea.Cmd
	m.searchInput, cmd = m.searchInput.Update(msg)
	return m, cmd
}

func (m Model) handleStationKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Defensive clamp at start of handler to prevent out-of-bounds scroll
	if len(m.stations) > 0 {
		if m.stationCursor < 0 {
			m.stationCursor = 0
		}
		if m.stationCursor >= len(m.stations) {
			m.stationCursor = len(m.stations) - 1
		}
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "tab":
		if len(m.departures) > 0 {
			m.focus = focusDepartures
			return m, nil
		}
		m.focus = focusSearch
		m.searchInput.Focus()
		return m, nil

	case "shift+tab":
		m.focus = focusAutoRefresh
		return m, nil

	case "esc", "/":
		m.focus = focusSearch
		m.searchInput.Focus()
		return m, nil

	case "j", "down":
		if m.stationCursor < len(m.stations)-1 {
			m.stationCursor++
		}
		return m, nil

	case "k", "up":
		if m.stationCursor > 0 {
			m.stationCursor--
		}
		return m, nil

	case "pgdown":
		if len(m.stations) > 0 {
			// Calculate page size based on viewport height
			pageSize := m.height - 10 // conservative estimate: minus header, filter bar, status
			if pageSize < 1 {
				pageSize = 10 // fallback
			}
			m.stationCursor += pageSize
			if m.stationCursor >= len(m.stations) {
				m.stationCursor = len(m.stations) - 1
			}
		}
		return m, nil

	case "pgup":
		if len(m.stations) > 0 {
			pageSize := m.height - 10
			if pageSize < 1 {
				pageSize = 10
			}
			m.stationCursor -= pageSize
			if m.stationCursor < 0 {
				m.stationCursor = 0
			}
		}
		return m, nil

	case "home":
		m.stationCursor = 0
		return m, nil

	case "end":
		if len(m.stations) > 0 {
			m.stationCursor = len(m.stations) - 1
		}
		return m, nil

	case "enter":
		if len(m.stations) > 0 {
			station := m.stations[m.stationCursor]
			m.selectedStation = &station
			m.departuresLoading = true
			m.departuresErr = nil
			m.departures = nil
			m.departureCursor = 0
			m.showJourney = false
			m.journey = nil
			return m, fetchBoard(m.client, station, m.selectedModes(), m.boardMode)
		}
	}

	return m, nil
}

func (m Model) handleDepartureKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Defensive clamp at start of handler to prevent out-of-bounds scroll
	if len(m.departures) > 0 {
		if m.departureCursor < 0 {
			m.departureCursor = 0
		}
		if m.departureCursor >= len(m.departures) {
			m.departureCursor = len(m.departures) - 1
		}
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "tab":
		if m.showJourney {
			m.focus = focusJourney
		} else {
			m.focus = focusSearch
			m.searchInput.Focus()
		}
		return m, nil

	case "shift+tab":
		m.focus = focusStations
		return m, nil

	case "esc":
		if m.showJourney {
			m.showJourney = false
			m.journey = nil
			m.selectedJourneyID = ""
			return m, nil
		}
		m.focus = focusStations
		return m, nil

	case "/":
		m.focus = focusSearch
		m.searchInput.Focus()
		return m, nil

	case "j", "down":
		if m.departureCursor < len(m.departures)-1 {
			m.departureCursor++
		}
		return m, nil

	case "k", "up":
		if m.departureCursor > 0 {
			m.departureCursor--
		}
		return m, nil

	case "pgdown":
		if len(m.departures) > 0 {
			// Calculate page size based on viewport height
			pageSize := m.height - 10 // conservative estimate
			if pageSize < 1 {
				pageSize = 10
			}
			m.departureCursor += pageSize
			if m.departureCursor >= len(m.departures) {
				m.departureCursor = len(m.departures) - 1
			}
		}
		return m, nil

	case "pgup":
		if len(m.departures) > 0 {
			pageSize := m.height - 10
			if pageSize < 1 {
				pageSize = 10
			}
			m.departureCursor -= pageSize
			if m.departureCursor < 0 {
				m.departureCursor = 0
			}
		}
		return m, nil

	case "home":
		m.departureCursor = 0
		return m, nil

	case "end":
		if len(m.departures) > 0 {
			m.departureCursor = len(m.departures) - 1
		}
		return m, nil

	case "enter":
		if len(m.departures) > 0 {
			dep := m.departures[m.departureCursor]
			if dep.JourneyID != "" {
				m.selectedJourneyID = dep.JourneyID
				m.journeyLoading = true
				m.journeyErr = nil
				m.journey = nil
				return m, fetchJourney(m.client, dep.JourneyID)
			}
		}
	}

	return m, nil
}

func (m Model) handleAutoRefreshKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case " ", "enter":
		m.autoRefresh = !m.autoRefresh
		if m.autoRefresh {
			// Do immediate update when enabling auto-refresh
			var cmds []tea.Cmd
			cmds = append(cmds, autoRefreshTick(), countdownTick())

			// Immediately refresh board if a station is selected
			if m.selectedStation != nil {
				cmds = append(cmds, fetchBoard(m.client, *m.selectedStation, m.selectedModes(), m.boardMode))
			}

			// Immediately refresh journey if one is displayed
			if m.showJourney && m.selectedJourneyID != "" {
				cmds = append(cmds, fetchJourney(m.client, m.selectedJourneyID))
			}

			return m, tea.Batch(cmds...)
		}
		return m, nil

	case "tab":
		if len(m.stations) > 0 {
			m.focus = focusStations
			return m, nil
		}
		m.focus = focusSearch
		m.searchInput.Focus()
		return m, nil

	case "shift+tab":
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

func (m Model) handleAutoRefreshTick() (tea.Model, tea.Cmd) {
	if !m.autoRefresh {
		return m, nil
	}

	var cmds []tea.Cmd

	// Schedule next tick
	cmds = append(cmds, autoRefreshTick())

	// Silently refresh board — keep existing data visible until new data arrives
	if m.selectedStation != nil {
		cmds = append(cmds, fetchBoard(m.client, *m.selectedStation, m.selectedModes(), m.boardMode))
	}

	// Silently refresh journey if one is displayed
	if m.showJourney && m.selectedJourneyID != "" {
		cmds = append(cmds, fetchJourney(m.client, m.selectedJourneyID))
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleCountdownTick() (tea.Model, tea.Cmd) {
	if !m.autoRefresh {
		return m, nil
	}
	// Schedule next countdown tick
	return m, countdownTick()
}

func (m Model) handleJourneyKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Defensive clamp at start of handler to prevent out-of-bounds scroll
	if m.journey != nil && len(m.journey.Stops) > 0 {
		if m.journeyScroll < 0 {
			m.journeyScroll = 0
		}
		if m.journeyScroll >= len(m.journey.Stops) {
			m.journeyScroll = len(m.journey.Stops) - 1
		}
	}

	switch msg.String() {
	case "q":
		return m, tea.Quit

	case "tab", "/":
		m.focus = focusSearch
		m.searchInput.Focus()
		return m, nil

	case "shift+tab":
		m.focus = focusDepartures
		return m, nil

	case "esc":
		m.focus = focusDepartures
		return m, nil

	case "j", "down":
		if m.journey != nil && m.journeyScroll < len(m.journey.Stops)-1 {
			m.journeyScroll++
			m.journeyManualScroll = true
		}
		return m, nil

	case "k", "up":
		if m.journeyScroll > 0 {
			m.journeyScroll--
			m.journeyManualScroll = true
		}
		return m, nil

	case "pgdown":
		if m.journey != nil && len(m.journey.Stops) > 0 {
			// Journey stops take ~3 lines each, so page size is smaller
			viewportHeight := m.height - 10
			pageSize := viewportHeight / 3
			if pageSize < 1 {
				pageSize = 5 // fallback
			}
			m.journeyScroll += pageSize
			if m.journeyScroll >= len(m.journey.Stops) {
				m.journeyScroll = len(m.journey.Stops) - 1
			}
			m.journeyManualScroll = true
		}
		return m, nil

	case "pgup":
		if m.journey != nil && len(m.journey.Stops) > 0 {
			viewportHeight := m.height - 10
			pageSize := viewportHeight / 3
			if pageSize < 1 {
				pageSize = 5
			}
			m.journeyScroll -= pageSize
			if m.journeyScroll < 0 {
				m.journeyScroll = 0
			}
			m.journeyManualScroll = true
		}
		return m, nil

	case "home":
		m.journeyScroll = 0
		m.journeyManualScroll = true
		return m, nil

	case "end":
		if m.journey != nil && len(m.journey.Stops) > 0 {
			m.journeyScroll = len(m.journey.Stops) - 1
			m.journeyManualScroll = true
		}
		return m, nil
	}

	return m, nil
}
