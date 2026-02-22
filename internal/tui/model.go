package tui

import (
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mobil-koeln/moko-cli/internal/api"
	"github.com/mobil-koeln/moko-cli/internal/models"
)

type focusPanel int

const (
	focusSearch focusPanel = iota
	focusFilters
	focusBoard
	focusAutoRefresh
	focusStations
	focusDepartures
	focusDestinations
	focusJourney
)

type boardMode int

const (
	boardDeparture boardMode = iota
	boardArrival
)

var modeLabels = []struct {
	apiName string
	label   string
}{
	{"ICE", "ICE"},
	{"EC_IC", "IC"},
	{"IR", "IR"},
	{"REGIONAL", "RE"},
	{"SBAHN", "S"},
	{"BUS", "Bus"},
	{"SCHIFF", "Ship"},
	{"UBAHN", "U"},
	{"TRAM", "Tram"},
	{"ANRUFPFLICHTIG", "On-call"},
}

// Model is the root Bubble Tea model for the TUI.
type Model struct {
	client *api.Client
	width  int
	height int

	searchInput textinput.Model
	focus       focusPanel

	// Filter bar - transport modes
	modeFilters  []bool
	filterCursor int

	// Board mode - departure/arrival
	boardMode   boardMode
	boardCursor int

	// Auto-refresh
	autoRefresh bool
	lastUpdate  time.Time

	// Left panel - stations
	stations        []models.Location
	stationCursor   int
	stationsLoading bool
	stationsErr     error
	searchSeq       int

	// Right panel - departures
	selectedStation   *models.Location
	departures        []models.Departure
	departureCursor   int
	departuresLoading bool
	departuresErr     error

	// Right panel - destination filter
	destinationList    []string
	destinationFilters []bool
	destinationCursor  int

	// Right panel - journey detail
	selectedJourneyID   string
	journey             *models.Journey
	journeyLoading      bool
	journeyErr          error
	showJourney         bool
	journeyScroll       int
	journeyManualScroll bool // true when user has manually scrolled in journey view
}

// New creates a new TUI model.
func New(client *api.Client) Model {
	ti := textinput.New()
	ti.Placeholder = "Search station..."
	ti.Focus()
	ti.CharLimit = 100
	ti.Width = 40

	filters := make([]bool, len(modeLabels))
	for i := range filters {
		filters[i] = true
	}

	return Model{
		client:      client,
		searchInput: ti,
		focus:       focusSearch,
		modeFilters: filters,
	}
}

// selectedModes returns the API mode names for active filters.
func (m Model) selectedModes() []string {
	var modes []string
	for i, active := range m.modeFilters {
		if active {
			modes = append(modes, modeLabels[i].apiName)
		}
	}
	return modes
}

// Init returns the initial command (textinput blink).
func (m Model) Init() tea.Cmd {
	return textinput.Blink
}
