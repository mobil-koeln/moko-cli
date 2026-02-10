package tui

import (
	"context"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mobil-koeln/moko-cli/internal/api"
	"github.com/mobil-koeln/moko-cli/internal/models"
)

const (
	apiTimeout          = 5 * time.Second
	autoRefreshInterval = 30 * time.Second
)

// autoRefreshTick returns a tea.Cmd that sends a tick after the refresh interval.
func autoRefreshTick() tea.Cmd {
	return tea.Tick(autoRefreshInterval, func(t time.Time) tea.Msg {
		return autoRefreshTickMsg(t)
	})
}

// countdownTick returns a tea.Cmd that sends a tick every second for countdown display.
func countdownTick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return countdownTickMsg(t)
	})
}

// searchStations returns a tea.Cmd that searches for stations.
func searchStations(client *api.Client, query string, seq int) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		locations, err := client.SearchLocations(ctx, query)
		return searchResultMsg{
			seq:       seq,
			locations: locations,
			err:       err,
		}
	}
}

// fetchBoard returns a tea.Cmd that fetches departures or arrivals for a station.
func fetchBoard(client *api.Client, station models.Location, modes []string, mode boardMode) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		req := api.StationBoardRequest{
			EVA:            station.EVA,
			StationID:      station.ID,
			NumVias:        5,
			ModesOfTransit: modes,
		}
		var departures []models.Departure
		var err error
		if mode == boardArrival {
			departures, err = client.GetArrivals(ctx, req)
		} else {
			departures, err = client.GetDepartures(ctx, req)
		}
		return departuresResultMsg{
			stationEVA: station.EVA,
			departures: departures,
			err:        err,
		}
	}
}

// fetchJourney returns a tea.Cmd that fetches journey details.
func fetchJourney(client *api.Client, journeyID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), apiTimeout)
		defer cancel()

		journey, err := client.GetJourney(ctx, journeyID, false)
		return journeyResultMsg{
			journeyID: journeyID,
			journey:   journey,
			err:       err,
		}
	}
}
