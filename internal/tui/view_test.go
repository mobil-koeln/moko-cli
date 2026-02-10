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
