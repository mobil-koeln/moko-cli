# CLAUDE.md - moko Project Guide

This file helps Claude Code instances understand the moko project architecture and conventions.

## Project Overview

**moko** is a Go CLI tool for querying Deutsche Bahn real-time transit information from the bahn.de API. It provides both a command-line interface (via Cobra) and an interactive TUI (via Bubble Tea).

## Quick Commands

### Build
```bash
make build          # Standard build → ./moko
make build-linux    # Cross-compile for Linux
make build-darwin   # Cross-compile for macOS
make build-windows  # Cross-compile for Windows
```

### Test & Lint
```bash
make test           # Run all tests with: go test -v ./...
make lint           # Run go vet ./...
go test ./...       # Direct test invocation
```

### Run
```bash
./moko                                            # Interactive TUI (default)
./moko tui                                        # Interactive TUI (explicit)
./moko search "Frankfurt Hbf"                    # CLI mode
./moko departures <eva>:<station_id> --modes ICE # CLI with filters
```

## Architecture

### High-Level Structure

```
┌─────────────────────────────────────────────────┐
│  cmd/moko/main.go (Cobra CLI)                  │
│  ├─ departures, arrivals, search, nearby,       │
│  │  journey, formation commands                 │
│  └─ tui command → launches Bubble Tea app       │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│  internal/api/client.go (API Layer)             │
│  ├─ HTTP client with browser emulation          │
│  ├─ Cookie jar for session persistence          │
│  └─ Optional file-based caching                 │
└─────────────────────────────────────────────────┘
                      │
                      ▼
┌─────────────────────────────────────────────────┐
│  https://www.bahn.de/web/api                    │
│  ├─ /fahrplan/v1/location (search)              │
│  ├─ /fahrplan/v1/departureBoard (departures)    │
│  ├─ /fahrplan/v1/arrivalBoard (arrivals)        │
│  └─ /ris/v1/wagenreihung (train formation)      │
└─────────────────────────────────────────────────┘
```

### Dual Interface Pattern

**CLI Mode** (cmd/moko/main.go):
- Cobra commands for one-off queries
- Three output formats: colored text (default), --json, --raw-json
- Global flags: --date, --time, --modes, --no-cache, --watch
- Each command creates API client → makes request → formats output → exits

**TUI Mode** (internal/tui/):
- Bubble Tea interactive application
- Model-View-Update architecture
- Focus panel system for navigation
- Real-time updates with auto-refresh

### TUI Architecture (internal/tui/)

The TUI is built with [Bubble Tea](https://github.com/charmbracelet/bubbletea) and follows the Elm architecture:

**Files:**
- `model.go` - State struct, focus panels, initialization
- `update.go` - Event handling (keyboard, messages)
- `view.go` - Rendering functions
- `commands.go` - Async operations (API calls)
- `messages.go` - Custom message types
- `styles.go` - Lipgloss styling
- `filterbar.go` - Transport mode filter UI
- `journey.go` - Journey detail rendering
- `map.go` - Route map visualization

**Focus Panel System:**

The TUI uses a `focusPanel` enum to track which UI element has keyboard focus:

```go
const (
    focusSearch      // Search input field
    focusFilters     // Transport mode chips
    focusBoard       // Departure/Arrival toggle
    focusAutoRefresh // Auto-refresh toggle
    focusStations    // Left panel - station list
    focusDepartures  // Right panel - departure/arrival list
    focusJourney     // Right panel - journey detail view
)
```

Navigation uses **Tab** to cycle forward through panels and **Esc** to return to search.

**Layout:**

```
┌──────────────────────────────────────────────────────────┐
│ [ASCII Logo]  mobil.koeln                                │
├──────────────────────────────────────────────────────────┤
│ Search: [Frankfurt Hbf_]                                 │
├──────────────────────────────────────────────────────────┤
│ Filters: [ICE] [IC] [RE] [S] [Bus] │ Dep/Arr │ Refresh  │
├──────────────────┬───────────────────────────────────────┤
│ STATIONS (35%)   │ DEPARTURES (65%)                      │
│ > Frankfurt Hbf  │ 14:30  +2  ICE 123  Pl.7  München     │
│   Frankfurt Süd  │ 14:35      RE 4567  Pl.12 Mainz       │
│                  ├───────────────────────────────────────┤
│                  │ JOURNEY DETAIL (if selected)          │
│                  │ [Stop list]    │ [Route Map]          │
└──────────────────┴───────────────────────────────────────┘
│ Status bar: j/k:navigate Enter:select Tab:next /:search  │
└──────────────────────────────────────────────────────────┘
```

**Auto-Refresh Behavior:**

When enabled (30-second interval):
- Silently re-fetches board data without clearing UI
- Preserves cursor position by journey ID
- Auto-closes journey view if selected train left the board
- Updates "Last update" timestamp

### API Client (internal/api/client.go)

**Key Features:**
1. **Browser Emulation** - Rotates through Firefox/Chrome profiles to avoid bot detection
2. **Cookie Persistence** - Uses `http.CookieJar` for session cookies
3. **Timezone Handling** - All times are Europe/Berlin
4. **Caching** - Optional file-based cache (default: `~/.cache/moko/`, 60s TTL)

**Request Pattern:**
```go
// All API methods follow this pattern:
func (c *Client) GetDepartures(ctx context.Context, req StationBoardRequest) ([]models.Departure, error)
func (c *Client) GetDeparturesRawJSON(ctx context.Context, req StationBoardRequest) (json.RawMessage, error)
```

**Transport Modes:**

The API uses 10 transport mode constants (defined in `internal/api/endpoints.go`):

```go
var ModesOfTransit = []string{
    "ICE", "EC_IC", "IR", "REGIONAL", "SBAHN",
    "BUS", "UBAHN", "TRAM", "SCHIFF", "ANRUFPFLICHTIG",
}
```

These are exposed in the CLI via `--modes` flag and in the TUI via filter chips.

### Data Flow Example (TUI)

1. User types "Frankfurt" and presses Enter
2. `handleSearchKeys()` increments `searchSeq` and calls `searchStations()`
3. `searchStations()` spawns goroutine → calls `client.SearchLocations()`
4. Result arrives as `searchResultMsg` with sequence number
5. `handleSearchResult()` checks sequence (ignore stale), updates `m.stations`
6. Auto-selects first station, calls `fetchBoard()` with current filter modes
7. `fetchBoard()` calls either `GetDepartures()` or `GetArrivals()` based on `boardMode`
8. Result arrives as `departuresResultMsg` with station EVA for validation
9. `handleDeparturesResult()` updates `m.departures`, preserves cursor by journey ID
10. View re-renders with new data

## Important Patterns

### Date/Time Handling

- **All API times are Europe/Berlin** - Set in `api.Client` initialization
- **Parsing** - `parseDateTime()` in main.go accepts DD.MM.YYYY or YYYY-MM-DD
- **Display** - Uses `Format("15:04")` for HH:MM, `Format("02.01.2006")` for dates
- **Zero values** - Check `dep.Dep != nil` before formatting times

### Stale Result Prevention

The TUI uses sequence numbers to prevent race conditions:

```go
m.searchSeq++
m.stationsLoading = true
return m, searchStations(m.client, query, m.searchSeq)

// Later in handler:
if msg.seq != m.searchSeq {
    return m, nil  // Ignore stale result
}
```

Station changes use EVA comparison:
```go
if msg.stationEVA != m.selectedStation.EVA {
    return m, nil  // Ignore result for old station
}
```

### Journey Persistence Across Refreshes

```go
// When opening a journey
m.selectedJourneyID = dep.JourneyID

// When refreshing departures
for i, dep := range m.departures {
    if dep.JourneyID == m.selectedJourneyID {
        m.departureCursor = i  // Re-locate journey
        found = true
        break
    }
}
if !found {
    m.showJourney = false  // Journey left the board
    m.selectedJourneyID = ""
}
```

### Output Formatting (internal/output/)

Three modules handle terminal output:
- `colors.go` - ANSI color codes with --color flag support
- `table.go` - Tabular departure/arrival formatting
- `terminal.go` - Additional helpers

**Color scheme:**
- Green: On-time (delay <= 0)
- Yellow: Minor delay (1-9 min)
- Red: Major delay (>= 10 min) or cancelled
- Cyan: Train line numbers
- Magenta: Platform numbers

## Testing Guidelines

### Test Structure

- Unit tests: `*_test.go` alongside source files
- Integration tests: `integration_test.go` with build tag `// +build integration`
- Test utilities: `internal/testutil/` (helpers, mock server, fixtures)

### Running Tests

```bash
make test              # All tests
make test-short        # Skip integration tests
make test-coverage     # With coverage report
make test-integration  # Integration tests only
make coverage-check    # Verify 70% coverage threshold
```

### Coverage Goals

| Package | Target | Current |
|---------|--------|---------|
| Overall | 70%+ | ~19.5% |
| internal/output | 60%+ | 92.9% ✅ |
| internal/models | 80%+ | 15.7% |
| internal/api | 70%+ | 5.9% |
| internal/tui | 50%+ | 0% |

### Writing Tests

```go
import "github.com/mobil-koeln/moko-cli/internal/testutil"

func TestExample(t *testing.T) {
    // Use testutil helpers
    testutil.AssertEqual(t, got, want)
    testutil.AssertNil(t, err)
    testutil.AssertContains(t, haystack, needle)

    // Mock HTTP server
    ms := testutil.NewMockServer(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte(testutil.SampleDepartureResponse))
    })
    defer ms.Close()

    // Table-driven tests preferred
    tests := []struct {
        name string
        input int
        want string
    }{
        {"case 1", 1, "one"},
        {"case 2", 2, "two"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := convert(tt.input)
            testutil.AssertEqual(t, got, tt.want)
        })
    }
}
```

### TUI Testing

Test message handlers in isolation:
```go
func TestHandleSearchResult(t *testing.T) {
    m := NewModel()
    m.searchSeq = 1

    msg := searchResultMsg{seq: 1, results: locations}
    newModel, cmd := m.Update(msg)

    testutil.AssertLen(t, newModel.(Model).stations, 2)
}
```

### Best Practices

- Use `t.Helper()` in test utilities
- Mock API responses where possible
- TUI components should test message handlers in isolation
- Table tests preferred for multiple cases
- Test error cases, not just happy paths
- Never do synchronous API calls in tests
- Use fixtures from `testutil.fixtures.go`

See `TESTING.md` for comprehensive testing guide.

## Common Gotchas

1. **Empty EVA/StationID** - Always validate location has both fields before API calls
2. **Nil time pointers** - Check `dep.Dep != nil` and `dep.Arr != nil`
3. **Platform changes** - Use `dep.EffectivePlatform()` not `dep.Platform` directly
4. **Journey ID format** - Opaque string from API, don't parse or modify
5. **Lipgloss Width** - ANSI codes count as zero width, use `lipgloss.Width()` not `len()`
6. **TUI blocking** - Never do synchronous API calls in Update(), always use Cmd

## Dependencies

Key external packages:
- `github.com/spf13/cobra` - CLI framework
- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/lipgloss` - Terminal styling
- `github.com/charmbracelet/bubbles/textinput` - Text input widget
- `github.com/fatih/color` - ANSI colors for CLI output

## Build Configuration

- Go version: 1.25+ (toolchain 1.25.x)
- Version injection: `go build -ldflags="-X main.version=0.1.0"`
- Cross-compilation: Use `GOOS` and `GOARCH` env vars
- Compression: UPX supported for Linux/Windows (not macOS)

## Future Extensions

When adding new features, consider:
- **New transport modes** - Update `ModesOfTransit` in endpoints.go and `modeLabels` in tui/model.go
- **New API endpoints** - Add constants to endpoints.go, methods to client.go, models to internal/models/
- **New TUI panels** - Add focus panel constant, key handlers, view functions
- **New output formats** - Add flag to main.go, formatter to internal/output/

## Release Process

### Homebrew Distribution

The project is distributed via Homebrew. GoReleaser handles automated releases:

1. **Create a tag**: `git tag -a v0.1.0 -m "Release v0.1.0" && git push origin v0.1.0`
2. **GitHub Actions** automatically:
   - Runs tests
   - Builds binaries for multiple platforms
   - Creates GitHub release
   - Updates Homebrew tap at `mobil-koeln/homebrew-tap`

3. **Installation**:
   ```bash
   brew tap mobil-koeln/tap
   brew install moko
   ```

### Manual Release

```bash
export GITHUB_TOKEN="your_token"
export HOMEBREW_TAP_GITHUB_TOKEN="your_token"
goreleaser release --clean
```

See `RELEASING.md` for detailed release instructions.

## Git Workflow

This project uses conventional commits:
- `feat:` - New features
- `fix:` - Bug fixes
- `refactor:` - Code restructuring
- `docs:` - Documentation updates
- `test:` - Test additions/updates
- `chore:` - Build/release changes

Commit with co-authorship:
```bash
git commit -m "feat: add new feature

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```
