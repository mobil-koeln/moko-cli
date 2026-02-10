package tui

import (
	"time"

	"github.com/mobil-koeln/moko-cli/internal/models"
)

// autoRefreshTickMsg is sent every 30 seconds when auto-refresh is enabled.
type autoRefreshTickMsg time.Time

// countdownTickMsg is sent every second when auto-refresh is enabled to update countdown display.
type countdownTickMsg time.Time

// searchResultMsg carries station search results back to the model.
// seq is used for stale-result detection.
type searchResultMsg struct {
	seq       int
	locations []models.Location
	err       error
}

// departuresResultMsg carries departure results for a specific station.
type departuresResultMsg struct {
	stationEVA int64
	departures []models.Departure
	err        error
}

// journeyResultMsg carries journey detail results.
type journeyResultMsg struct {
	journeyID string
	journey   *models.Journey
	err       error
}
