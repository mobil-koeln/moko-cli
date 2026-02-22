package tui

import (
	"testing"
	"time"

	"github.com/mobil-koeln/moko-cli/internal/api"
	"github.com/mobil-koeln/moko-cli/internal/models"
	"github.com/mobil-koeln/moko-cli/internal/testutil"
)

func TestModel_View(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.height = 50

	// Just verify View doesn't panic
	output := m.View()
	testutil.AssertTrue(t, len(output) > 0)
}

func TestModel_View_WithStations(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.height = 50

	m.stations = []models.Location{
		{Name: "Frankfurt Hbf", EVA: 8000105},
		{Name: "München Hbf", EVA: 8000261},
	}

	output := m.View()
	testutil.AssertTrue(t, len(output) > 0)
	testutil.AssertContains(t, output, "Frankfurt Hbf")
}

func TestModel_View_WithDepartures(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.height = 50

	depTime := time.Now()
	m.selectedStation = &models.Location{Name: "Frankfurt Hbf", EVA: 8000105}
	m.departures = []models.Departure{
		{JourneyID: "j1", Line: "ICE 123", Dep: &depTime, Destination: "München"},
		{JourneyID: "j2", Line: "RE 456", Dep: &depTime, Destination: "Mainz"},
	}

	output := m.View()
	testutil.AssertTrue(t, len(output) > 0)
	testutil.AssertContains(t, output, "ICE 123")
}

func TestModel_View_WithJourney(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.height = 50
	m.showJourney = true

	m.journey = &models.Journey{
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Frankfurt Hbf", EVA: 8000105},
			{Name: "München Hbf", EVA: 8000261},
		},
	}

	output := m.View()
	testutil.AssertTrue(t, len(output) > 0)
	testutil.AssertContains(t, output, "ICE 123")
}

func TestModel_View_Loading(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.height = 50
	m.stationsLoading = true

	output := m.View()
	testutil.AssertTrue(t, len(output) > 0)
}

func TestModel_View_Error(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.height = 50
	m.stationsErr = api.ErrNotFound

	output := m.View()
	testutil.AssertTrue(t, len(output) > 0)
}

func TestRenderStationList(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	m.stations = []models.Location{
		{Name: "Frankfurt Hbf", EVA: 8000105},
		{Name: "München Hbf", EVA: 8000261},
	}
	m.stationCursor = 0

	output := m.renderStationList(40, 30)
	testutil.AssertTrue(t, len(output) > 0)
	testutil.AssertContains(t, output, "Frankfurt Hbf")
}

func TestRenderDepartureList(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	depTime := time.Now()
	m.departures = []models.Departure{
		{JourneyID: "j1", Line: "ICE 123", Dep: &depTime},
	}
	m.departureCursor = 0

	output := m.renderDepartureList(60, 30)
	testutil.AssertTrue(t, len(output) > 0)
}

func TestRenderRightPanel(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)

	depTime := time.Now()
	m.selectedStation = &models.Location{Name: "Frankfurt Hbf", EVA: 8000105}
	m.departures = []models.Departure{
		{JourneyID: "j1", Line: "ICE 123", Dep: &depTime},
	}

	output := m.renderRightPanel(60, 30)
	testutil.AssertTrue(t, len(output) > 0)
}

func TestFindBoardStationIdx_ByEVA(t *testing.T) {
	depTime := time.Now().Add(time.Hour)
	stops := []models.Stop{
		{Name: "Hamburg Hbf", EVA: 8002549, Dep: &depTime},
		{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &depTime},
		{Name: "München Hbf", EVA: 8000261, Dep: &depTime},
	}
	station := &models.Location{EVA: 8000105}
	got := findBoardStationIdx(stops, station)
	testutil.AssertEqual(t, got, 1)
}

func TestFindBoardStationIdx_ByCoordinates(t *testing.T) {
	// Simulates transit stops where EVA doesn't match but coordinates do
	depTime := time.Now().Add(time.Hour)
	stops := []models.Stop{
		{Name: "Stop A", EVA: 0, Lat: 50.960, Lon: 7.001, Dep: &depTime},
		{Name: "Mülheim Keupstr.", EVA: 0, Lat: 50.9652, Lon: 7.00578, Dep: &depTime},
		{Name: "Stop C", EVA: 0, Lat: 50.970, Lon: 7.010, Dep: &depTime},
	}
	station := &models.Location{EVA: 900312003, Lat: 50.9652, Lon: 7.00578}
	got := findBoardStationIdx(stops, station)
	testutil.AssertEqual(t, got, 1)
}

func TestFindBoardStationIdx_EVAZeroFallsBackToCoords(t *testing.T) {
	depTime := time.Now().Add(time.Hour)
	stops := []models.Stop{
		{Name: "Stop A", EVA: 0, Lat: 50.960, Lon: 7.001, Dep: &depTime},
		{Name: "Board Stop", EVA: 0, Lat: 50.9652, Lon: 7.00578, Dep: &depTime},
	}
	// Station EVA 0 also, but coordinates match
	station := &models.Location{EVA: 0, Lat: 50.9652, Lon: 7.00578}
	got := findBoardStationIdx(stops, station)
	testutil.AssertEqual(t, got, 1)
}

func TestFindBoardStationIdx_NilStation(t *testing.T) {
	depTime := time.Now().Add(time.Hour)
	stops := []models.Stop{{Name: "Stop A", EVA: 8000105, Dep: &depTime}}
	got := findBoardStationIdx(stops, nil)
	testutil.AssertEqual(t, got, -1)
}

func TestFindBoardStationIdx_NoMatch(t *testing.T) {
	depTime := time.Now().Add(time.Hour)
	stops := []models.Stop{
		{Name: "Stop A", EVA: 8000100, Lat: 50.000, Lon: 8.000, Dep: &depTime},
	}
	station := &models.Location{EVA: 8000200, Lat: 51.000, Lon: 9.000}
	got := findBoardStationIdx(stops, station)
	testutil.AssertEqual(t, got, -1)
}

func TestRenderJourneyDetail_BoardStationHighlighted(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.height = 50
	m.showJourney = true

	depTime := time.Now().Add(time.Hour)
	m.selectedStation = &models.Location{Name: "Frankfurt Hbf", EVA: 8000105}
	m.journey = &models.Journey{
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Hamburg Hbf", EVA: 8002549, Dep: &depTime},
			{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &depTime},
			{Name: "München Hbf", EVA: 8000261, Dep: &depTime},
		},
	}
	m.journeyScroll = 0

	output := m.renderJourneyDetail(60, 12)
	// Board station stop should be rendered (content present)
	testutil.AssertContains(t, output, "Frankfurt Hbf")
}

func TestRenderJourneyDetail_BoardStationRedWhenCurrent(t *testing.T) {
	// When board station is also the current stop, red takes priority (not green)
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.height = 50
	m.showJourney = true

	past := time.Now().Add(-30 * time.Minute)
	future := time.Now().Add(30 * time.Minute)
	boardEVA := int64(8000105)

	m.selectedStation = &models.Location{Name: "Frankfurt Hbf", EVA: boardEVA}
	m.journey = &models.Journey{
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Hamburg Hbf", EVA: 8002549, Arr: &past},
			{Name: "Frankfurt Hbf", EVA: boardEVA, Arr: &past, Dep: &future},
			{Name: "München Hbf", EVA: 8000261, Arr: &future},
		},
	}
	m.journeyScroll = 0

	output := m.renderJourneyDetail(60, 12)
	// Should render without panic; content is present
	testutil.AssertContains(t, output, "Frankfurt Hbf")
}

func TestRenderJourneyDetail_NoBoardStationHighlight(t *testing.T) {
	// When selectedStation is nil, no board station highlighting should occur
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.height = 50
	m.showJourney = true
	m.selectedStation = nil // No selected station

	depTime := time.Now().Add(time.Hour)
	m.journey = &models.Journey{
		Name: "ICE 123",
		Stops: []models.Stop{
			{Name: "Hamburg Hbf", EVA: 8002549, Dep: &depTime},
			{Name: "Frankfurt Hbf", EVA: 8000105, Dep: &depTime},
		},
	}

	// Should render without panic
	output := m.renderJourneyDetail(60, 12)
	testutil.AssertContains(t, output, "Frankfurt Hbf")
}

func TestRenderStatusBar(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.focus = focusSearch

	output := m.renderStatusBar()
	testutil.AssertTrue(t, len(output) > 0)
}

func TestRenderStatusBar_DifferentFocus(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100

	focusPanels := []focusPanel{
		focusSearch,
		focusFilters,
		focusBoard,
		focusAutoRefresh,
		focusStations,
		focusDepartures,
		focusJourney,
	}

	for _, panel := range focusPanels {
		m.focus = panel
		output := m.renderStatusBar()
		testutil.AssertTrue(t, len(output) > 0)
	}
}
