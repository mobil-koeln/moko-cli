package tui

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mobil-koeln/moko-cli/internal/api"
	"github.com/mobil-koeln/moko-cli/internal/models"
	"github.com/mobil-koeln/moko-cli/internal/testutil"
)

func TestNew(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Check initial state
	testutil.AssertTrue(t, m.client != nil)
	testutil.AssertEqual(t, m.focus, focusSearch)
	testutil.AssertLen(t, m.modeFilters, len(modeLabels))

	// All filters should be enabled by default
	for i, filter := range m.modeFilters {
		if !filter {
			t.Errorf("mode filter %d should be enabled by default", i)
		}
	}
}

func TestModel_Init(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	cmd := m.Init()
	testutil.AssertTrue(t, cmd != nil)
}

func TestModel_SelectedModes(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// All modes selected by default
	modes := m.selectedModes()
	testutil.AssertEqual(t, len(modes), len(modeLabels))

	// Test with some modes disabled
	m.modeFilters[0] = false // ICE
	m.modeFilters[1] = false // EC_IC

	modes = m.selectedModes()
	testutil.AssertEqual(t, len(modes), len(modeLabels)-2)

	// Verify ICE is not in the list
	for _, mode := range modes {
		if mode == "ICE" {
			t.Error("ICE should not be in selected modes")
		}
	}
}

func TestModel_SelectedModes_AllDisabled(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Disable all modes
	for i := range m.modeFilters {
		m.modeFilters[i] = false
	}

	modes := m.selectedModes()
	testutil.AssertLen(t, modes, 0)
}

func TestModel_SelectedModes_SingleMode(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Disable all except ICE
	for i := range m.modeFilters {
		m.modeFilters[i] = false
	}
	m.modeFilters[0] = true // ICE

	modes := m.selectedModes()
	testutil.AssertLen(t, modes, 1)
	testutil.AssertEqual(t, modes[0], "ICE")
}

func TestFocusPanel_Constants(t *testing.T) {
	// Verify focus panel constants exist and are distinct
	panels := []focusPanel{
		focusSearch,
		focusFilters,
		focusBoard,
		focusAutoRefresh,
		focusStations,
		focusDepartures,
		focusJourney,
	}

	// Check that all panels have different values
	seen := make(map[focusPanel]bool)
	for _, panel := range panels {
		if seen[panel] {
			t.Errorf("duplicate focus panel value: %d", panel)
		}
		seen[panel] = true
	}
}

func TestBoardMode_Constants(t *testing.T) {
	testutil.AssertTrue(t, boardDeparture != boardArrival)
	testutil.AssertEqual(t, int(boardDeparture), 0)
	testutil.AssertEqual(t, int(boardArrival), 1)
}

func TestModeLabels_Count(t *testing.T) {
	// Should have 10 transport modes
	testutil.AssertEqual(t, len(modeLabels), 10)

	// Verify some expected modes exist
	found := make(map[string]bool)
	for _, label := range modeLabels {
		found[label.apiName] = true
	}

	expectedModes := []string{"ICE", "EC_IC", "SBAHN", "BUS"}
	for _, mode := range expectedModes {
		if !found[mode] {
			t.Errorf("expected mode %s not found in modeLabels", mode)
		}
	}
}

func TestModel_InitialState(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Check initial values
	testutil.AssertEqual(t, m.stationCursor, 0)
	testutil.AssertEqual(t, m.departureCursor, 0)
	testutil.AssertEqual(t, m.filterCursor, 0)
	testutil.AssertEqual(t, m.boardCursor, 0)
	testutil.AssertFalse(t, m.stationsLoading)
	testutil.AssertFalse(t, m.departuresLoading)
	testutil.AssertFalse(t, m.journeyLoading)
	testutil.AssertFalse(t, m.showJourney)
	testutil.AssertFalse(t, m.autoRefresh)
	testutil.AssertEqual(t, m.searchSeq, 0)
	testutil.AssertEqual(t, m.boardMode, boardDeparture)
}

func TestModel_WindowSize(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Initial size should be 0
	testutil.AssertEqual(t, m.width, 0)
	testutil.AssertEqual(t, m.height, 0)

	// Update with WindowSizeMsg
	newModel, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.width, 100)
	testutil.AssertEqual(t, m.height, 50)
}

func TestSearchResultMsg_Success(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.searchSeq = 1
	m.stationsLoading = true

	locations := []models.Location{
		{Name: "Frankfurt Hbf", EVA: 8000105, ID: "test-id-1"},
		{Name: "München Hbf", EVA: 8000261, ID: "test-id-2"},
	}

	msg := searchResultMsg{
		seq:       1,
		locations: locations,
		err:       nil,
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	testutil.AssertLen(t, m.stations, 2)
	testutil.AssertFalse(t, m.stationsLoading)
	testutil.AssertNil(t, m.stationsErr)
	testutil.AssertEqual(t, m.stationCursor, 0)
}

func TestSearchResultMsg_StaleResult(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.searchSeq = 2 // Current sequence is 2
	m.stationsLoading = true

	locations := []models.Location{
		{Name: "Frankfurt Hbf", EVA: 8000105},
	}

	// Old result with seq=1
	msg := searchResultMsg{
		seq:       1,
		locations: locations,
		err:       nil,
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Should ignore stale result
	testutil.AssertLen(t, m.stations, 0)
	testutil.AssertTrue(t, m.stationsLoading) // Still loading
}

func TestSearchResultMsg_Error(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.searchSeq = 1
	m.stationsLoading = true

	msg := searchResultMsg{
		seq:       1,
		locations: nil,
		err:       api.ErrNotFound,
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	testutil.AssertLen(t, m.stations, 0)
	testutil.AssertFalse(t, m.stationsLoading)
	testutil.AssertError(t, m.stationsErr)
}

func TestDeparturesResultMsg_Success(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Set up selected station
	station := &models.Location{Name: "Frankfurt Hbf", EVA: 8000105}
	m.selectedStation = station
	m.departuresLoading = true

	depTime := time.Now()
	departures := []models.Departure{
		{JourneyID: "journey-1", Line: "ICE 123", Dep: &depTime},
		{JourneyID: "journey-2", Line: "RE 456", Dep: &depTime},
	}

	msg := departuresResultMsg{
		stationEVA: 8000105,
		departures: departures,
		err:        nil,
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	testutil.AssertLen(t, m.departures, 2)
	testutil.AssertFalse(t, m.departuresLoading)
	testutil.AssertNil(t, m.departuresErr)
}

func TestDeparturesResultMsg_WrongStation(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Set up selected station
	station := &models.Location{Name: "Frankfurt Hbf", EVA: 8000105}
	m.selectedStation = station
	m.departuresLoading = true

	// Result for different station
	msg := departuresResultMsg{
		stationEVA: 8000261, // München, not Frankfurt
		departures: []models.Departure{},
		err:        nil,
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Should ignore result for wrong station
	testutil.AssertLen(t, m.departures, 0)
	testutil.AssertTrue(t, m.departuresLoading) // Still loading
}

func TestJourneyResultMsg_Success(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.selectedJourneyID = "journey-123"
	m.journeyLoading = true

	journey := &models.Journey{
		ID:   "journey-123",
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Frankfurt Hbf", EVA: 8000105},
			{Name: "München Hbf", EVA: 8000261},
		},
	}

	msg := journeyResultMsg{
		journeyID: "journey-123",
		journey:   journey,
		err:       nil,
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	testutil.AssertTrue(t, m.journey != nil)
	testutil.AssertFalse(t, m.journeyLoading)
	testutil.AssertNil(t, m.journeyErr)
	testutil.AssertEqual(t, m.journey.Name, "ICE 123")
}

func TestJourneyResult_RefreshPreservesManualScroll(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Set up an already-displayed journey
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)
	m.showJourney = true
	m.journey = &models.Journey{
		ID:   "journey-123",
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &past},
			{Name: "Mannheim Hbf", EVA: 8000244, Arr: &now, Dep: &now},
			{Name: "Stuttgart Hbf", EVA: 8000096, Arr: &future},
			{Name: "München Hbf", EVA: 8000261, Arr: &future},
		},
	}

	// User manually scrolled to stop index 3
	m.journeyScroll = 3
	m.journeyManualScroll = true

	// Simulate a refresh with the same journey data
	msg := journeyResultMsg{
		journeyID: "journey-123",
		journey: &models.Journey{
			ID:   "journey-123",
			Name: "ICE 123",
			Stops: []models.Stop{
				{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &past},
				{Name: "Mannheim Hbf", EVA: 8000244, Arr: &now, Dep: &now},
				{Name: "Stuttgart Hbf", EVA: 8000096, Arr: &future},
				{Name: "München Hbf", EVA: 8000261, Arr: &future},
			},
		},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Manual scroll position should be preserved
	testutil.AssertEqual(t, m.journeyScroll, 3)
	testutil.AssertTrue(t, m.journeyManualScroll)
}

func TestJourneyResult_RefreshAutoScrollsWhenNotManual(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Set up an already-displayed journey with NO manual scroll
	now := time.Now()
	past := now.Add(-1 * time.Hour)
	future := now.Add(1 * time.Hour)
	m.showJourney = true
	m.journeyManualScroll = false
	m.journeyScroll = 0
	m.journey = &models.Journey{
		ID:   "journey-123",
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &past},
			{Name: "Mannheim Hbf", EVA: 8000244, Arr: &past, Dep: &past},
			{Name: "Stuttgart Hbf", EVA: 8000096, Arr: &future},
			{Name: "München Hbf", EVA: 8000261, Arr: &future},
		},
	}

	// Refresh with same data — should auto-scroll to current station
	msg := journeyResultMsg{
		journeyID: "journey-123",
		journey: &models.Journey{
			ID:   "journey-123",
			Name: "ICE 123",
			Stops: []models.Stop{
				{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &past},
				{Name: "Mannheim Hbf", EVA: 8000244, Arr: &past, Dep: &past},
				{Name: "Stuttgart Hbf", EVA: 8000096, Arr: &future},
				{Name: "München Hbf", EVA: 8000261, Arr: &future},
			},
		},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Should NOT stay at 0 — should auto-scroll to the current stop
	// The current stop should be around index 1 or 2 (between past and future stops)
	testutil.AssertTrue(t, m.journeyScroll > 0)
	testutil.AssertFalse(t, m.journeyManualScroll)
}

func TestJourneyResult_NewJourneyResetsManualScroll(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// Journey is NOT currently shown — this is a new journey open
	m.showJourney = false
	m.journey = nil
	m.journeyManualScroll = true // leftover from previous journey
	m.journeyScroll = 5          // leftover scroll position

	now := time.Now()
	past := now.Add(-1 * time.Hour)

	msg := journeyResultMsg{
		journeyID: "journey-456",
		journey: &models.Journey{
			ID:   "journey-456",
			Name: "RE 456",
			Stops: []models.Stop{
				{Name: "Köln Hbf", EVA: 8000207, Dep: &past},
				{Name: "Bonn Hbf", EVA: 8000044, Arr: &now},
			},
		},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Manual scroll flag should be reset for new journey
	testutil.AssertFalse(t, m.journeyManualScroll)
	testutil.AssertTrue(t, m.showJourney)
}

func TestJourneyResult_RefreshClampsScrollWhenStopsShrank(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	now := time.Now()
	past := now.Add(-1 * time.Hour)
	m.showJourney = true
	m.journey = &models.Journey{
		ID:   "journey-123",
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &past},
			{Name: "Mannheim Hbf", EVA: 8000244, Arr: &now},
			{Name: "Stuttgart Hbf", EVA: 8000096, Arr: &now},
			{Name: "München Hbf", EVA: 8000261, Arr: &now},
		},
	}
	m.journeyScroll = 3 // Scrolled to last stop
	m.journeyManualScroll = true

	// Refresh with fewer stops
	msg := journeyResultMsg{
		journeyID: "journey-123",
		journey: &models.Journey{
			ID:   "journey-123",
			Name: "ICE 123",
			Stops: []models.Stop{
				{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &past},
				{Name: "Mannheim Hbf", EVA: 8000244, Arr: &now},
			},
		},
	}

	newModel, _ := m.Update(msg)
	m = newModel.(Model)

	// Scroll should be clamped to last valid index
	testutil.AssertEqual(t, m.journeyScroll, 1)
	testutil.AssertTrue(t, m.journeyManualScroll)
}

func TestJourneyKeys_ManualScrollSetsFlag(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.focus = focusJourney
	m.showJourney = true
	m.journeyManualScroll = false

	now := time.Now()
	m.journey = &models.Journey{
		ID:   "journey-123",
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &now},
			{Name: "Mannheim Hbf", EVA: 8000244, Arr: &now},
			{Name: "Stuttgart Hbf", EVA: 8000096, Arr: &now},
		},
	}
	m.journeyScroll = 0

	// Press "j" to scroll down
	newModel, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.journeyScroll, 1)
	testutil.AssertTrue(t, m.journeyManualScroll)

	// Press "k" to scroll back up
	newModel, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	m = newModel.(Model)

	testutil.AssertEqual(t, m.journeyScroll, 0)
	testutil.AssertTrue(t, m.journeyManualScroll) // Still true
}

func TestAutoRefreshTickMsg(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.autoRefresh = true

	station := &models.Location{Name: "Frankfurt Hbf", EVA: 8000105, ID: "test-id"}
	m.selectedStation = station

	msg := autoRefreshTickMsg(time.Now())

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	// Should return a batch command for refresh
	testutil.AssertTrue(t, cmd != nil)
	testutil.AssertTrue(t, m.autoRefresh) // Still enabled
}

func TestCountdownTickMsg(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.autoRefresh = true

	msg := countdownTickMsg(time.Now())

	newModel, cmd := m.Update(msg)
	m = newModel.(Model)

	// Should schedule next tick
	testutil.AssertTrue(t, cmd != nil)
}

func TestModel_QuitMsg(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	// QuitMsg is handled by bubbletea itself
	// Just verify Update doesn't panic
	newModel, _ := m.Update(tea.QuitMsg{})
	testutil.AssertTrue(t, newModel != nil)
}
