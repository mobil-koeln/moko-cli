package tui

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mobil-koeln/moko-cli/internal/api"
	"github.com/mobil-koeln/moko-cli/internal/models"
	"github.com/mobil-koeln/moko-cli/internal/testutil"
)

// --- visibleRange tests ---

func TestVisibleRange_AllFit(t *testing.T) {
	// When total items fit in viewport, show all
	start, end := visibleRange(0, 5, 10)
	testutil.AssertEqual(t, start, 0)
	testutil.AssertEqual(t, end, 5)
}

func TestVisibleRange_EmptyList(t *testing.T) {
	start, end := visibleRange(0, 0, 10)
	testutil.AssertEqual(t, start, 0)
	testutil.AssertEqual(t, end, 0)
}

func TestVisibleRange_SingleItem(t *testing.T) {
	start, end := visibleRange(0, 1, 10)
	testutil.AssertEqual(t, start, 0)
	testutil.AssertEqual(t, end, 1)
}

func TestVisibleRange_CursorAtStart(t *testing.T) {
	// cursor=0, 20 items, viewport=10 → should show 0-10
	start, end := visibleRange(0, 20, 10)
	testutil.AssertEqual(t, start, 0)
	testutil.AssertEqual(t, end, 10)
}

func TestVisibleRange_CursorInMiddle(t *testing.T) {
	// cursor=10, 20 items, viewport=10 → centered: start=5, end=15
	start, end := visibleRange(10, 20, 10)
	testutil.AssertEqual(t, start, 5)
	testutil.AssertEqual(t, end, 15)
}

func TestVisibleRange_CursorAtEnd(t *testing.T) {
	// cursor=19, 20 items, viewport=10 → end-clamped: start=10, end=20
	start, end := visibleRange(19, 20, 10)
	testutil.AssertEqual(t, start, 10)
	testutil.AssertEqual(t, end, 20)
}

func TestVisibleRange_CursorNearStart(t *testing.T) {
	// cursor=2, 20 items, viewport=10 → start clamped to 0
	start, end := visibleRange(2, 20, 10)
	testutil.AssertEqual(t, start, 0)
	testutil.AssertEqual(t, end, 10)
}

func TestVisibleRange_CursorNearEnd(t *testing.T) {
	// cursor=18, 20 items, viewport=10 → end clamped
	start, end := visibleRange(18, 20, 10)
	testutil.AssertEqual(t, start, 10)
	testutil.AssertEqual(t, end, 20)
}

func TestVisibleRange_ViewportOfOne(t *testing.T) {
	start, end := visibleRange(5, 20, 1)
	testutil.AssertEqual(t, start, 5)
	testutil.AssertEqual(t, end, 6)
}

func TestVisibleRange_ExactFit(t *testing.T) {
	// total == maxVisible
	start, end := visibleRange(5, 10, 10)
	testutil.AssertEqual(t, start, 0)
	testutil.AssertEqual(t, end, 10)
}

func TestVisibleRange_CursorAlwaysVisible(t *testing.T) {
	// Property test: cursor should always be within [start, end)
	for total := 1; total <= 30; total++ {
		for maxVis := 1; maxVis <= 15; maxVis++ {
			for cursor := 0; cursor < total; cursor++ {
				start, end := visibleRange(cursor, total, maxVis)
				if cursor < start || cursor >= end {
					t.Errorf("cursor %d not in [%d, %d) for total=%d, maxVisible=%d",
						cursor, start, end, total, maxVis)
				}
				if end-start > maxVis {
					t.Errorf("visible range %d exceeds maxVisible %d", end-start, maxVis)
				}
			}
		}
	}
}

// --- Station list scrolling tests ---

func makeStations(n int) []models.Location {
	stations := make([]models.Location, n)
	for i := 0; i < n; i++ {
		stations[i] = models.Location{
			Name: fmt.Sprintf("Station %d", i),
			EVA:  int64(8000100 + i),
			ID:   fmt.Sprintf("id-%d", i),
		}
	}
	return stations
}

func makeDepartures(n int) []models.Departure {
	depTime := time.Now()
	deps := make([]models.Departure, n)
	for i := 0; i < n; i++ {
		deps[i] = models.Departure{
			JourneyID:   fmt.Sprintf("journey-%d", i),
			Line:        fmt.Sprintf("ICE %d", i),
			Dep:         &depTime,
			Destination: fmt.Sprintf("Destination %d", i),
		}
	}
	return deps
}

func makeStops(n int) []models.Stop {
	stops := make([]models.Stop, n)
	now := time.Now()
	for i := 0; i < n; i++ {
		depTime := now.Add(time.Duration(i*10) * time.Minute)
		stops[i] = models.Stop{
			Name: fmt.Sprintf("Stop %d", i),
			EVA:  int64(8000100 + i),
			Dep:  &depTime,
		}
	}
	return stops
}

func newTestModel() Model {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 120
	m.height = 40
	return m
}

func TestStationScrolling_NavigateDown(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(25)
	m.focus = focusStations
	m.stationCursor = 0

	// Navigate down through stations
	for i := 0; i < 24; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = newModel.(Model)
		testutil.AssertEqual(t, m.stationCursor, i+1)
	}

	// Should stop at last item
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 24)
}

func TestStationScrolling_NavigateUp(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(25)
	m.focus = focusStations
	m.stationCursor = 24

	// Navigate up
	for i := 24; i > 0; i-- {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		m = newModel.(Model)
		testutil.AssertEqual(t, m.stationCursor, i-1)
	}

	// Should stop at first item
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)
}

func TestStationScrolling_RenderWithScroll(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(25)
	m.stationCursor = 20

	output := m.renderStationList(40, 12)
	// Should contain the cursor station
	testutil.AssertContains(t, output, "Station 20")
	// Should NOT contain stations far from cursor (use exact names to avoid substring matches)
	testutil.AssertNotContains(t, output, "Station 0\n")
	testutil.AssertNotContains(t, output, " Station 5\n")
}

func TestStationScrolling_CursorAtTopRender(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(25)
	m.stationCursor = 0

	output := m.renderStationList(40, 12)
	testutil.AssertContains(t, output, "Station 0")
	testutil.AssertContains(t, output, ">")
}

func TestStationScrolling_ResetOnNewSearch(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(25)
	m.stationCursor = 20
	m.searchSeq = 1

	msg := searchResultMsg{
		seq:       1,
		locations: makeStations(10),
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	testutil.AssertEqual(t, m.stationCursor, 0)
	testutil.AssertLen(t, m.stations, 10)
}

// --- Departure list scrolling tests ---

func TestDepartureScrolling_NavigateDown(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(20)
	m.focus = focusDepartures
	m.departureCursor = 0

	for i := 0; i < 19; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = newModel.(Model)
		testutil.AssertEqual(t, m.departureCursor, i+1)
	}

	// Stop at end
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.departureCursor, 19)
}

func TestDepartureScrolling_NavigateUp(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(20)
	m.focus = focusDepartures
	m.departureCursor = 19

	for i := 19; i > 0; i-- {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		m = newModel.(Model)
		testutil.AssertEqual(t, m.departureCursor, i-1)
	}

	// Stop at start
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.departureCursor, 0)
}

func TestDepartureScrolling_RenderWithScroll(t *testing.T) {
	m := newTestModel()
	m.selectedStation = &models.Location{Name: "Test Station", EVA: 8000100}
	m.departures = makeDepartures(20)
	m.departureCursor = 15
	m.focus = focusDepartures

	output := m.renderDepartureList(80, 12)
	// Should contain cursor departure
	testutil.AssertContains(t, output, "ICE 15")
}

func TestDepartureScrolling_ResetOnStationChange(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(20)
	m.departureCursor = 15
	m.stations = makeStations(5)
	m.focus = focusStations
	m.stationCursor = 2

	// Press enter to select a new station
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.departureCursor, 0)
	testutil.AssertTrue(t, m.departuresLoading)
}

// --- Journey scrolling tests ---

func TestJourneyScrolling_NavigateDown(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{
		Name:  "ICE 123",
		Stops: makeStops(20),
	}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 0

	for i := 0; i < 19; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = newModel.(Model)
		testutil.AssertEqual(t, m.journeyScroll, i+1)
	}

	// Stop at end
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.journeyScroll, 19)
}

func TestJourneyScrolling_NavigateUp(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{
		Name:  "ICE 123",
		Stops: makeStops(20),
	}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 19

	for i := 19; i > 0; i-- {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		m = newModel.(Model)
		testutil.AssertEqual(t, m.journeyScroll, i-1)
	}

	// Stop at start
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.journeyScroll, 0)
}

func TestJourneyScrolling_ResetOnNewJourney(t *testing.T) {
	m := newTestModel()
	m.showJourney = false
	m.journey = nil
	m.journeyScroll = 10

	msg := journeyResultMsg{
		journeyID: "j-new",
		journey: &models.Journey{
			Name:  "RE 456",
			Stops: makeStops(15),
		},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	testutil.AssertEqual(t, m.journeyScroll, 0)
	testutil.AssertTrue(t, m.showJourney)
}

func TestJourneyScrolling_PreservedOnRefreshWithManualScroll(t *testing.T) {
	m := newTestModel()
	m.showJourney = true
	m.journey = &models.Journey{
		Name:  "ICE 123",
		Stops: makeStops(20),
	}
	m.journeyScroll = 12
	m.journeyManualScroll = true // User manually scrolled

	// Simulate auto-refresh returning updated journey
	msg := journeyResultMsg{
		journeyID: "j-123",
		journey: &models.Journey{
			Name:  "ICE 123",
			Stops: makeStops(20),
		},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Scroll position should be preserved because user manually scrolled
	testutil.AssertEqual(t, m.journeyScroll, 12)
	testutil.AssertTrue(t, m.showJourney)
	testutil.AssertTrue(t, m.journeyManualScroll)
}

func TestJourneyScrolling_AutoScrollsOnRefreshWithoutManualScroll(t *testing.T) {
	m := newTestModel()
	m.showJourney = true
	m.journeyManualScroll = false

	now := time.Now()
	past := now.Add(-30 * time.Minute)
	future := now.Add(30 * time.Minute)

	// FindCurrentStopIndex uses Arr times, so we must set those
	m.journey = &models.Journey{
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Stop A", EVA: 1, Arr: &past, Dep: &past},
			{Name: "Stop B", EVA: 2, Arr: &now, Dep: &now},
			{Name: "Stop C", EVA: 3, Arr: &future, Dep: &future},
		},
	}
	m.journeyScroll = 0

	// Refresh should auto-scroll to current station (not preserve position)
	msg := journeyResultMsg{
		journeyID: "j-123",
		journey: &models.Journey{
			Name: "ICE 123",
			Stops: []models.Stop{
				{Name: "Stop A", EVA: 1, Arr: &past, Dep: &past},
				{Name: "Stop B", EVA: 2, Arr: &now, Dep: &now},
				{Name: "Stop C", EVA: 3, Arr: &future, Dep: &future},
			},
		},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Should auto-scroll to current stop (index 1, since Stop B's Arr <= now)
	testutil.AssertTrue(t, m.journeyScroll > 0)
	testutil.AssertFalse(t, m.journeyManualScroll)
}

func TestJourneyScrolling_RenderWithScroll(t *testing.T) {
	m := newTestModel()
	m.showJourney = true
	m.journey = &models.Journey{
		Name:  "ICE 123",
		Stops: makeStops(25),
	}
	m.journeyScroll = 20

	output := m.renderJourneyDetail(60, 12)
	testutil.AssertContains(t, output, "Stop 20")
}

// --- Focus preservation tests ---

func TestFocusSwitch_PreservesStationCursor(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(25)
	m.departures = makeDepartures(20)
	m.focus = focusStations
	m.stationCursor = 15

	// Tab to departures
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.focus, focusDepartures)
	testutil.AssertEqual(t, m.stationCursor, 15) // Preserved

	// Esc back to stations
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.focus, focusStations)
	testutil.AssertEqual(t, m.stationCursor, 15) // Still preserved
}

func TestFocusSwitch_PreservesDepartureCursor(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(20)
	m.focus = focusDepartures
	m.departureCursor = 12
	m.showJourney = true
	m.journey = &models.Journey{Name: "Test", Stops: makeStops(5)}

	// Tab to journey
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.focus, focusJourney)
	testutil.AssertEqual(t, m.departureCursor, 12) // Preserved

	// Esc back to departures
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.focus, focusDepartures)
	testutil.AssertEqual(t, m.departureCursor, 12) // Still preserved
}

func TestFocusSwitch_PreservesJourneyScroll(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "Test", Stops: makeStops(20)}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 8

	// Tab to search
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.focus, focusSearch)
	testutil.AssertEqual(t, m.journeyScroll, 8) // Preserved
}

// --- Auto-refresh preservation tests ---

func TestAutoRefresh_PreservesDepartureCursorByJourneyID(t *testing.T) {
	m := newTestModel()
	station := &models.Location{Name: "Frankfurt Hbf", EVA: 8000105}
	m.selectedStation = station
	m.departures = makeDepartures(20)
	m.departureCursor = 10
	m.selectedJourneyID = "journey-10"

	// Simulate refresh with reordered list: journey-10 is now at index 5
	refreshed := makeDepartures(20)
	// Swap so that journey-10 is at a different position
	refreshed[5], refreshed[10] = refreshed[10], refreshed[5]

	msg := departuresResultMsg{
		stationEVA: 8000105,
		departures: refreshed,
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Cursor should follow journey-10 to its new position
	testutil.AssertEqual(t, m.departureCursor, 5)
}

func TestAutoRefresh_JourneyLeftBoard(t *testing.T) {
	m := newTestModel()
	station := &models.Location{Name: "Frankfurt Hbf", EVA: 8000105}
	m.selectedStation = station
	m.departures = makeDepartures(20)
	m.departureCursor = 10
	m.selectedJourneyID = "journey-10"
	m.showJourney = true
	m.journey = &models.Journey{Name: "ICE 10", Stops: makeStops(5)}

	// Refresh without journey-10 in the list
	refreshed := makeDepartures(15)
	// None have journeyID "journey-10" since range is 0-14
	// Actually they do: journey-10 is at index 10
	// Create departures without the target journey
	for i := range refreshed {
		refreshed[i].JourneyID = fmt.Sprintf("other-journey-%d", i)
	}

	msg := departuresResultMsg{
		stationEVA: 8000105,
		departures: refreshed,
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Journey should be closed
	testutil.AssertFalse(t, m.showJourney)
	testutil.AssertTrue(t, m.journey == nil)
	testutil.AssertEqual(t, m.selectedJourneyID, "")
}

func TestAutoRefresh_ClampsOverflowingCursor(t *testing.T) {
	m := newTestModel()
	station := &models.Location{Name: "Frankfurt Hbf", EVA: 8000105}
	m.selectedStation = station
	m.departures = makeDepartures(20)
	m.departureCursor = 18

	// Refresh with fewer items
	msg := departuresResultMsg{
		stationEVA: 8000105,
		departures: makeDepartures(10),
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Cursor should be clamped to last item
	testutil.AssertEqual(t, m.departureCursor, 9)
}

// --- Edge cases ---

func TestScrolling_EmptyStationList(t *testing.T) {
	m := newTestModel()
	m.stations = nil
	m.focus = focusStations

	// Navigation on empty list should not panic
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)
}

func TestScrolling_EmptyDepartureList(t *testing.T) {
	m := newTestModel()
	m.departures = nil
	m.focus = focusDepartures

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.departureCursor, 0)
}

func TestScrolling_NilJourney(t *testing.T) {
	m := newTestModel()
	m.journey = nil
	m.focus = focusJourney

	// Should not panic
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.journeyScroll, 0)
}

func TestScrolling_SingleStation(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(1)
	m.focus = focusStations
	m.stationCursor = 0

	// Can't go down
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)

	// Can't go up
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)
}

func TestScrolling_SingleDeparture(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(1)
	m.focus = focusDepartures
	m.departureCursor = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.departureCursor, 0)
}

func TestScrolling_SingleJourneyStop(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "Test", Stops: makeStops(1)}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.journeyScroll, 0)
}

func TestRenderStationList_SmallViewport(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(25)
	m.stationCursor = 10

	// Very small viewport - should not panic
	output := m.renderStationList(30, 5)
	testutil.AssertTrue(t, len(output) > 0)
	testutil.AssertContains(t, output, "Station 10")
}

func TestRenderDepartureList_SmallViewport(t *testing.T) {
	m := newTestModel()
	m.selectedStation = &models.Location{Name: "Test", EVA: 1}
	m.departures = makeDepartures(20)
	m.departureCursor = 15

	output := m.renderDepartureList(60, 5)
	testutil.AssertTrue(t, len(output) > 0)
}

func TestRenderJourneyDetail_SmallViewport(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "ICE 1", Stops: makeStops(25)}
	m.showJourney = true
	m.journeyScroll = 20

	output := m.renderJourneyDetail(50, 5)
	testutil.AssertTrue(t, len(output) > 0)
}

// --- Arrow key tests (down/up in addition to j/k) ---

func TestScrolling_ArrowKeys(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(10)
	m.focus = focusStations
	m.stationCursor = 0

	// Down arrow
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 1)

	// Up arrow
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)
}

// --- Journey scroll not clamped on refresh with fewer stops (potential issue) ---

func TestJourneyScroll_ClampedOnRefreshWithFewerStops(t *testing.T) {
	m := newTestModel()
	m.showJourney = true
	m.journey = &models.Journey{
		Name:  "ICE 123",
		Stops: makeStops(20),
	}
	m.journeyScroll = 18
	m.journeyManualScroll = true // User manually scrolled

	// Refresh with fewer stops
	msg := journeyResultMsg{
		journeyID: "j-123",
		journey: &models.Journey{
			Name:  "ICE 123",
			Stops: makeStops(10),
		},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// journeyScroll should be clamped to last valid index (10-1=9)
	testutil.AssertEqual(t, m.journeyScroll, 9)
	testutil.AssertTrue(t, m.journeyManualScroll)

	// Rendering should not panic
	output := m.renderJourneyDetail(60, 12)
	testutil.AssertTrue(t, len(output) > 0)

	// Pressing 'k' should work to scroll back
	m.focus = focusJourney
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.journeyScroll, 8)

	// 'j' at last stop should not go further
	m.journeyScroll = 9
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.journeyScroll, 9)
}

// --- Defensive Clamping Tests (Invisible Scroll Bug Fix) ---

func TestDefensiveClamping_StationRapidKeyPresses(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(10)
	m.focus = focusStations
	m.stationCursor = 9 // At last item

	// Rapidly press 'j' 20 times beyond bounds
	for i := 0; i < 20; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = newModel.(Model)
	}

	// Cursor should be clamped at last valid index, not overflowed
	testutil.AssertEqual(t, m.stationCursor, 9)

	// Should be able to navigate back immediately (no invisible scroll)
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 8)
}

func TestDefensiveClamping_DepartureRapidKeyPresses(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(15)
	m.focus = focusDepartures
	m.departureCursor = 14 // At last item

	// Rapidly press 'j' beyond bounds
	for i := 0; i < 30; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = newModel.(Model)
	}

	// Cursor should stay at last valid index
	testutil.AssertEqual(t, m.departureCursor, 14)

	// Should be able to navigate back
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.departureCursor, 13)
}

func TestDefensiveClamping_JourneyRapidKeyPresses(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "ICE 1", Stops: makeStops(12)}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 11 // At last stop

	// Rapidly press 'j' beyond bounds
	for i := 0; i < 25; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
		m = newModel.(Model)
	}

	// Scroll should stay at last valid index
	testutil.AssertEqual(t, m.journeyScroll, 11)

	// Should be able to navigate back
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.journeyScroll, 10)
}

func TestDefensiveClamping_StationNegativeBound(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(10)
	m.focus = focusStations
	m.stationCursor = 0 // At first item

	// Rapidly press 'k' to try going negative
	for i := 0; i < 15; i++ {
		newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
		m = newModel.(Model)
	}

	// Cursor should stay at 0, not go negative
	testutil.AssertEqual(t, m.stationCursor, 0)

	// Should be able to navigate forward
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 1)
}

func TestDefensiveClamping_JourneyListShrink(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "ICE 1", Stops: makeStops(20)}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 18
	m.journeyManualScroll = true

	// Simulate refresh with fewer stops (journey list shrinks)
	msg := journeyResultMsg{
		journeyID: "j-1",
		journey:   &models.Journey{Name: "ICE 1", Stops: makeStops(10)},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Scroll should be clamped to new max (9)
	testutil.AssertEqual(t, m.journeyScroll, 9)

	// Next key press should work correctly (no invisible scroll)
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.journeyScroll, 9) // Still at max

	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.journeyScroll, 8) // Can navigate back
}

// --- Page Scroll Tests (PgUp/PgDn) ---

func TestPageScroll_StationPageDown(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(50)
	m.focus = focusStations
	m.stationCursor = 0

	// Press Page Down
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = newModel.(Model)

	// Should jump by ~page size (height - 10 = 30)
	testutil.AssertTrue(t, m.stationCursor > 10) // At least jumped significantly
	testutil.AssertTrue(t, m.stationCursor < 50) // But not beyond bounds
}

func TestPageScroll_StationPageUp(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(50)
	m.focus = focusStations
	m.stationCursor = 40

	// Press Page Up
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = newModel.(Model)

	// Should jump back by ~page size
	testutil.AssertTrue(t, m.stationCursor < 30) // Moved back significantly
	testutil.AssertTrue(t, m.stationCursor >= 0) // But not negative
}

func TestPageScroll_DeparturePageDown(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(40)
	m.focus = focusDepartures
	m.departureCursor = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = newModel.(Model)

	testutil.AssertTrue(t, m.departureCursor > 10)
	testutil.AssertTrue(t, m.departureCursor < 40)
}

func TestPageScroll_DeparturePageUp(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(40)
	m.focus = focusDepartures
	m.departureCursor = 35

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = newModel.(Model)

	testutil.AssertTrue(t, m.departureCursor < 25)
	testutil.AssertTrue(t, m.departureCursor >= 0)
}

func TestPageScroll_JourneyPageDown(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "ICE 1", Stops: makeStops(30)}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 0

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = newModel.(Model)

	// Journey uses smaller page size (height/3)
	testutil.AssertTrue(t, m.journeyScroll > 3)
	testutil.AssertTrue(t, m.journeyScroll < 30)
	testutil.AssertTrue(t, m.journeyManualScroll) // Should set manual scroll flag
}

func TestPageScroll_JourneyPageUp(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "ICE 1", Stops: makeStops(30)}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 25

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = newModel.(Model)

	testutil.AssertTrue(t, m.journeyScroll < 20)
	testutil.AssertTrue(t, m.journeyScroll >= 0)
	testutil.AssertTrue(t, m.journeyManualScroll)
}

func TestPageScroll_PageDownAtEnd(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(20)
	m.focus = focusStations
	m.stationCursor = 18

	// Page down when near end should clamp to last item
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.stationCursor, 19) // Clamped to last
}

func TestPageScroll_PageUpAtStart(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(20)
	m.focus = focusStations
	m.stationCursor = 2

	// Page up when near start should clamp to first item
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.stationCursor, 0) // Clamped to first
}

func TestPageScroll_EmptyList(t *testing.T) {
	m := newTestModel()
	m.stations = nil
	m.focus = focusStations

	// Page down on empty list should not panic
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)

	// Page up on empty list should not panic
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)
}

// --- Home/End Key Tests ---

func TestHomeEnd_StationHome(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(30)
	m.focus = focusStations
	m.stationCursor = 20

	// Press Home
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.stationCursor, 0)
}

func TestHomeEnd_StationEnd(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(30)
	m.focus = focusStations
	m.stationCursor = 5

	// Press End
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.stationCursor, 29)
}

func TestHomeEnd_DepartureHome(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(25)
	m.focus = focusDepartures
	m.departureCursor = 18

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.departureCursor, 0)
}

func TestHomeEnd_DepartureEnd(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(25)
	m.focus = focusDepartures
	m.departureCursor = 3

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.departureCursor, 24)
}

func TestHomeEnd_JourneyHome(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "ICE 1", Stops: makeStops(20)}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 15

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.journeyScroll, 0)
	testutil.AssertTrue(t, m.journeyManualScroll)
}

func TestHomeEnd_JourneyEnd(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "ICE 1", Stops: makeStops(20)}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 2

	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.journeyScroll, 19)
	testutil.AssertTrue(t, m.journeyManualScroll)
}

func TestHomeEnd_EmptyList(t *testing.T) {
	m := newTestModel()
	m.stations = nil
	m.focus = focusStations

	// Home on empty list
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)

	// End on empty list
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)
}

func TestHomeEnd_SingleItem(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(1)
	m.focus = focusStations
	m.stationCursor = 0

	// Home on single item
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)

	// End on single item
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = newModel.(Model)
	testutil.AssertEqual(t, m.stationCursor, 0)
}

// --- Scroll Indicator Tests ---

func TestScrollIndicator_EmptyList(t *testing.T) {
	indicator := scrollIndicator(0, 0)
	testutil.AssertEqual(t, indicator, "")
}

func TestScrollIndicator_SingleItem(t *testing.T) {
	indicator := scrollIndicator(0, 1)
	testutil.AssertEqual(t, indicator, "1/1")
}

func TestScrollIndicator_FirstOfMany(t *testing.T) {
	indicator := scrollIndicator(0, 20)
	testutil.AssertEqual(t, indicator, "1/20")
}

func TestScrollIndicator_MiddleItem(t *testing.T) {
	indicator := scrollIndicator(9, 20)
	testutil.AssertEqual(t, indicator, "10/20")
}

func TestScrollIndicator_LastItem(t *testing.T) {
	indicator := scrollIndicator(19, 20)
	testutil.AssertEqual(t, indicator, "20/20")
}

func TestRenderStatusBar_WithStationIndicator(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(25)
	m.focus = focusStations
	m.stationCursor = 14

	statusBar := m.renderStatusBar()

	// Should contain the scroll indicator
	testutil.AssertContains(t, statusBar, "15/25")
}

func TestRenderStatusBar_WithDepartureIndicator(t *testing.T) {
	m := newTestModel()
	m.departures = makeDepartures(18)
	m.focus = focusDepartures
	m.departureCursor = 7

	statusBar := m.renderStatusBar()

	testutil.AssertContains(t, statusBar, "8/18")
}

func TestRenderStatusBar_WithJourneyIndicator(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "ICE 1", Stops: makeStops(12)}
	m.showJourney = true
	m.focus = focusJourney
	m.journeyScroll = 5

	statusBar := m.renderStatusBar()

	testutil.AssertContains(t, statusBar, "6/12")
}

func TestRenderStatusBar_NoIndicatorWhenNotFocused(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(25)
	m.stationCursor = 10
	m.focus = focusSearch // Not focused on stations

	statusBar := m.renderStatusBar()

	// Should NOT contain station indicator
	testutil.AssertNotContains(t, statusBar, "11/25")
}

func TestRenderStatusBar_IndicatorWithEmptyList(t *testing.T) {
	m := newTestModel()
	m.stations = nil
	m.focus = focusStations

	statusBar := m.renderStatusBar()

	// Should not show numeric indicator (e.g., "1/10") for empty list
	// Note: "/" appears in keyboard hints like "PgUp/PgDn" and "/:search"
	// So we check that there's no "│  X/Y" pattern (indicator separator + numbers)
	testutil.AssertNotContains(t, statusBar, "│  1/")
	testutil.AssertNotContains(t, statusBar, "│  0/")
}

func TestRenderStatusBar_UpdatedHintsForNewKeys(t *testing.T) {
	m := newTestModel()
	m.focus = focusStations

	statusBar := m.renderStatusBar()

	// Should contain new navigation hints
	testutil.AssertContains(t, statusBar, "PgUp/PgDn:page")
	testutil.AssertContains(t, statusBar, "Home/End:jump")
}

// --- Scrollbar Rendering Tests ---

func TestRenderScrollbar_EmptyList(t *testing.T) {
	scrollbar := renderScrollbar(0, 0, 10)
	lines := strings.Split(scrollbar, "\n")
	testutil.AssertEqual(t, len(lines), 10)
	// All should be empty spaces
	for _, line := range lines {
		testutil.AssertTrue(t, line == " " || line == "")
	}
}

func TestRenderScrollbar_AllItemsFit(t *testing.T) {
	// 10 items in 20 lines - all fit, so full thumb
	scrollbar := renderScrollbar(5, 10, 20)
	lines := strings.Split(scrollbar, "\n")
	testutil.AssertEqual(t, len(lines), 20)
	// All lines should have thumb (█)
	for _, line := range lines {
		testutil.AssertContains(t, line, "█")
	}
}

func TestRenderScrollbar_ScrollAtTop(t *testing.T) {
	// Cursor at 0, 50 items, 10 lines
	scrollbar := renderScrollbar(0, 50, 10)
	lines := strings.Split(scrollbar, "\n")
	testutil.AssertEqual(t, len(lines), 10)
	// First line should have thumb
	testutil.AssertContains(t, lines[0], "█")
}

func TestRenderScrollbar_ScrollAtBottom(t *testing.T) {
	// Cursor at last item (49), 50 items, 10 lines
	scrollbar := renderScrollbar(49, 50, 10)
	lines := strings.Split(scrollbar, "\n")
	testutil.AssertEqual(t, len(lines), 10)
	// Last line should have thumb
	testutil.AssertContains(t, lines[9], "█")
}

func TestRenderScrollbar_ScrollInMiddle(t *testing.T) {
	// Cursor in middle (25), 50 items, 10 lines
	scrollbar := renderScrollbar(25, 50, 10)
	lines := strings.Split(scrollbar, "\n")
	testutil.AssertEqual(t, len(lines), 10)
	// Thumb should be somewhere in the middle
	hasThumb := false
	for i := 3; i < 7; i++ {
		if strings.Contains(lines[i], "█") {
			hasThumb = true
			break
		}
	}
	testutil.AssertTrue(t, hasThumb)
}

func TestRenderScrollbar_SingleLineHeight(t *testing.T) {
	scrollbar := renderScrollbar(10, 50, 1)
	lines := strings.Split(scrollbar, "\n")
	testutil.AssertEqual(t, len(lines), 1)
	// Should still render something
	testutil.AssertTrue(t, len(lines[0]) > 0)
}

func TestRenderScrollbar_LargeList(t *testing.T) {
	// 100 items, 10 lines, cursor at 50
	scrollbar := renderScrollbar(50, 100, 10)
	lines := strings.Split(scrollbar, "\n")
	testutil.AssertEqual(t, len(lines), 10)
	// Should have a small thumb (less than half the height)
	thumbCount := 0
	for _, line := range lines {
		if strings.Contains(line, "█") {
			thumbCount++
		}
	}
	testutil.AssertTrue(t, thumbCount > 0 && thumbCount < 5)
}

func TestRenderScrollbar_ZeroHeight(t *testing.T) {
	scrollbar := renderScrollbar(5, 20, 0)
	// Should return empty string or handle gracefully
	testutil.AssertTrue(t, len(scrollbar) == 0 || scrollbar == "")
}

func TestRenderScrollbar_IntegrationWithStationList(t *testing.T) {
	m := newTestModel()
	m.stations = makeStations(50)
	m.stationCursor = 25

	output := m.renderStationList(40, 15)
	// Should contain scrollbar symbols
	testutil.AssertTrue(t, strings.Contains(output, "█") || strings.Contains(output, "│"))
}

func TestRenderScrollbar_IntegrationWithDepartureList(t *testing.T) {
	m := newTestModel()
	m.selectedStation = &models.Location{Name: "Test Station", EVA: 8000100}
	m.departures = makeDepartures(30)
	m.departureCursor = 15

	output := m.renderDepartureList(80, 15)
	// Should contain scrollbar symbols
	testutil.AssertTrue(t, strings.Contains(output, "█") || strings.Contains(output, "│"))
}

func TestRenderScrollbar_IntegrationWithJourneyDetail(t *testing.T) {
	m := newTestModel()
	m.journey = &models.Journey{Name: "ICE 1", Stops: makeStops(40)}
	m.showJourney = true
	m.journeyScroll = 20

	output := m.renderJourneyDetail(60, 15)
	// Should contain scrollbar symbols
	testutil.AssertTrue(t, strings.Contains(output, "█") || strings.Contains(output, "│"))
}

// --- Truncate helper tests ---

func TestTruncate(t *testing.T) {
	tests := []struct {
		name  string
		input string
		width int
		want  string
	}{
		{"no truncation", "Hello", 10, "Hello"},
		{"exact length", "Hello", 5, "Hello"},
		{"truncated with tilde", "Hello World", 8, "Hello W~"},
		{"very short width", "Hello", 2, "He"},
		{"zero width", "Hello", 0, ""},
		{"negative width", "Hello", -1, ""},
		{"empty string", "", 10, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncate(tt.input, tt.width)
			testutil.AssertEqual(t, got, tt.want)
		})
	}
}
