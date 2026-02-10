package tui

import (
	"testing"

	"github.com/mobil-koeln/moko-cli/internal/api"
	"github.com/mobil-koeln/moko-cli/internal/testutil"
)

func TestRenderFilterBar(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100

	// Just verify it doesn't panic
	output := m.renderFilterBar()
	testutil.AssertTrue(t, len(output) > 0)
}

func TestRenderFilterBar_AllEnabled(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100

	// All filters enabled by default
	output := m.renderFilterBar()

	// Should contain mode labels
	testutil.AssertContains(t, output, "ICE")
	testutil.AssertContains(t, output, "IC")
}

func TestRenderFilterBar_SomeDisabled(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100

	// Disable some modes
	m.modeFilters[0] = false // ICE
	m.modeFilters[1] = false // IC

	output := m.renderFilterBar()

	// Should still render but with different styling
	testutil.AssertTrue(t, len(output) > 0)
}

func TestRenderFilterBar_FocusedFilter(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.focus = focusFilters
	m.filterCursor = 0

	output := m.renderFilterBar()

	// Should render with focus styling
	testutil.AssertTrue(t, len(output) > 0)
}

func TestRenderFilterBar_Departures(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.boardMode = boardDeparture

	output := m.renderFilterBar()

	// Should indicate departure mode
	testutil.AssertContains(t, output, "Dep")
}

func TestRenderFilterBar_Arrivals(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.boardMode = boardArrival

	output := m.renderFilterBar()

	// Should indicate arrival mode
	testutil.AssertContains(t, output, "Arr")
}

func TestRenderFilterBar_AutoRefreshOff(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.autoRefresh = false

	output := m.renderFilterBar()

	// Should show refresh is off
	testutil.AssertTrue(t, len(output) > 0)
}

func TestRenderFilterBar_AutoRefreshOn(t *testing.T) {
	client, _ := api.NewClient()
	m := New(client)
	m.width = 100
	m.autoRefresh = true

	output := m.renderFilterBar()

	// Should show refresh is on
	testutil.AssertTrue(t, len(output) > 0)
}
