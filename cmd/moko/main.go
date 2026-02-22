package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/mobil-koeln/moko-cli/internal/api"
	"github.com/mobil-koeln/moko-cli/internal/models"
	"github.com/mobil-koeln/moko-cli/internal/output"
	"github.com/mobil-koeln/moko-cli/internal/tui"
	"github.com/spf13/cobra"
)

var version = "0.4.0"

func main() {
	if err := rootCmd.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

var rootCmd = &cobra.Command{
	Use:   "moko",
	Short: "CLI for querying Deutsche Bahn real-time transit information",
	Long: `moko is a command-line interface for querying Deutsche Bahn (DB)
real-time transit information from the bahn.de API.

Features:
  - Departure and arrival boards at any station
  - Journey/trip details with all stops
  - Station search by name or geographic coordinates
  - Train carriage formation (Wagenreihung)
  - Filter by transport modes (ICE, EC/IC, Regional, S-Bahn, etc.)
  - JSON output for scripting
  - Response caching for faster repeated queries

Quick Start:
  1. Launch TUI:               moko (or moko tui)
  2. Search for a station:     moko search "Frankfurt Hbf"
  3. Show departures:          moko departures <eva>:<station_id>
  4. Show arrivals:            moko arrivals <eva>:<station_id>
  5. Find nearby stations:     moko nearby 50.107:8.663
  6. Get journey details:      moko journey <journey_id>
  7. Show train formation:     moko formation <eva> ICE 623`,
	Version: version,
	RunE: func(cmd *cobra.Command, args []string) error {
		// If no subcommand is provided, launch TUI
		if len(args) == 0 {
			return runTUI(cmd, args)
		}
		return cmd.Help()
	},
}

// Global flags
var (
	flagDate    string
	flagTime    string
	flagJSON    bool
	flagRawJSON bool
	flagColor   string
	flagNoCache bool
	flagShowVia bool
)

// Departures/Arrivals flags
var (
	flagNumVias   int
	flagModes     []string
	flagLine      string
	flagDirection string
	flagWatch     bool
	flagJourney   bool
)

func init() {
	// Add subcommands
	rootCmd.AddCommand(departuresCmd)
	rootCmd.AddCommand(arrivalsCmd)
	rootCmd.AddCommand(searchCmd)
	rootCmd.AddCommand(nearbyCmd)
	rootCmd.AddCommand(journeyCmd)
	rootCmd.AddCommand(formationCmd)
	rootCmd.AddCommand(tuiCmd)

	// Global flags
	rootCmd.PersistentFlags().StringVarP(&flagDate, "date", "d", "", "Date (DD.MM.YYYY or YYYY-MM-DD)")
	rootCmd.PersistentFlags().StringVarP(&flagTime, "time", "t", "", "Time (HH:MM)")
	rootCmd.PersistentFlags().BoolVar(&flagJSON, "json", false, "Output as JSON")
	rootCmd.PersistentFlags().BoolVar(&flagRawJSON, "raw-json", false, "Output raw API response")
	rootCmd.PersistentFlags().StringVar(&flagColor, "color", "auto", "Color output: auto, always, never")
	rootCmd.PersistentFlags().BoolVar(&flagNoCache, "no-cache", false, "Disable response caching")

	// Departures-specific flags
	departuresCmd.Flags().IntVar(&flagNumVias, "vias", 5, "Number of intermediate stops to show")
	departuresCmd.Flags().StringSliceVarP(&flagModes, "modes", "m", nil, "Filter by transport modes (ICE,EC_IC,REGIONAL,SBAHN,BUS,UBAHN,TRAM)")
	departuresCmd.Flags().BoolVarP(&flagShowVia, "via", "v", false, "Show intermediate stops")
	departuresCmd.Flags().StringVarP(&flagLine, "line", "l", "", "Filter by line number (exact match)")
	departuresCmd.Flags().StringVar(&flagDirection, "direction", "", "Filter by destination (substring match)")
	departuresCmd.Flags().BoolVarP(&flagWatch, "watch", "w", false, "Watch mode: refresh every 30 seconds")
	departuresCmd.Flags().BoolVarP(&flagJourney, "journey", "j", false, "Show journey ID for each departure")

	// Arrivals-specific flags (same as departures)
	arrivalsCmd.Flags().IntVar(&flagNumVias, "vias", 5, "Number of intermediate stops to show")
	arrivalsCmd.Flags().StringSliceVarP(&flagModes, "modes", "m", nil, "Filter by transport modes (ICE,EC_IC,REGIONAL,SBAHN,BUS,UBAHN,TRAM)")
	arrivalsCmd.Flags().BoolVarP(&flagShowVia, "via", "v", false, "Show intermediate stops")
	arrivalsCmd.Flags().StringVarP(&flagLine, "line", "l", "", "Filter by line number (exact match)")
	arrivalsCmd.Flags().StringVar(&flagDirection, "direction", "", "Filter by destination (substring match)")
	arrivalsCmd.Flags().BoolVarP(&flagWatch, "watch", "w", false, "Watch mode: refresh every 30 seconds")
	arrivalsCmd.Flags().BoolVarP(&flagJourney, "journey", "j", false, "Show journey ID for each arrival")

	// Journey-specific flags
	journeyCmd.Flags().BoolVarP(&flagWatch, "watch", "w", false, "Watch mode: refresh every 30 seconds")
}

// createClient creates an API client with common options
func createClient() (*api.Client, error) {
	opts := []api.ClientOption{}

	// Enable caching unless disabled
	if !flagNoCache {
		opts = append(opts, api.WithDefaultCache())
	}

	return api.NewClient(opts...)
}

// getColorMode returns the color mode based on flag
func getColorMode() output.ColorMode {
	return output.ParseColorMode(flagColor)
}

var departuresCmd = &cobra.Command{
	Use:   "departures <eva>:<station_id>",
	Short: "Show departures at a station",
	Long: `Show upcoming departures at a station.

The station must be specified as EVA:ID format, e.g.:
  moko departures 8000105:A=1@O=Frankfurt(Main)Hbf@X=8663003@Y=50107145@U=80@L=8000105@B=1@p=1234567890

Use 'moko search <name>' to find station IDs.

Available transport modes for --modes flag:
  ICE          - ICE trains
  EC_IC        - EuroCity/InterCity trains
  IR           - InterRegio trains
  REGIONAL     - Regional trains (RE, RB)
  SBAHN        - S-Bahn trains
  BUS          - Buses
  UBAHN        - U-Bahn (subway)
  TRAM         - Trams/Streetcars
  SCHIFF       - Ferries/Ships
  ANRUFPFLICHTIG - On-demand services

Filtering:
  --line, -l <line>      Filter by line number (exact match, e.g., S1, 623)
  --direction <dest>     Filter by destination (substring match)

Additional Output:
  --journey, -j          Show journey ID (use with 'moko journey <id>')
  --watch, -w            Refresh every 30 seconds (full-screen mode)

Examples:
  moko departures 8000105:...                    # All departures
  moko departures 8000105:... --modes ICE,EC_IC  # Only long-distance trains
  moko departures 8000105:... --modes SBAHN      # Only S-Bahn
  moko departures 8000105:... --via              # Show intermediate stops
  moko departures 8000105:... --line S1          # Only S1 line
  moko departures 8000105:... --direction Frankfurt  # Going to Frankfurt
  moko departures 8000105:... -l ICE --direction München
  moko departures 8000105:... --journey          # Show journey IDs
  moko departures 8000105:... --watch            # Watch mode with 30s refresh
  moko departures 8000105:... --line S1 --watch  # Watch only S1 line`,
	Args: cobra.ExactArgs(1),
	RunE: runDepartures,
}

var arrivalsCmd = &cobra.Command{
	Use:   "arrivals <eva>:<station_id>",
	Short: "Show arrivals at a station",
	Long: `Show upcoming arrivals at a station.

The station must be specified as EVA:ID format, e.g.:
  moko arrivals 8000105:A=1@O=Frankfurt(Main)Hbf@X=8663003@Y=50107145@U=80@L=8000105@B=1@p=1234567890

Use 'moko search <name>' to find station IDs.

Filtering:
  --line, -l <line>      Filter by line number (exact match, e.g., S1, 623)
  --direction <dest>     Filter by origin (substring match)

Additional Output:
  --journey, -j          Show journey ID (use with 'moko journey <id>')
  --watch, -w            Refresh every 30 seconds (full-screen mode)

Examples:
  moko arrivals 8000105:...                    # All arrivals
  moko arrivals 8000105:... --line S1          # Only S1 line
  moko arrivals 8000105:... --direction Berlin # Coming from Berlin
  moko arrivals 8000105:... --journey          # Show journey IDs
  moko arrivals 8000105:... --watch            # Watch mode with 30s refresh`,
	Args: cobra.ExactArgs(1),
	RunE: runArrivals,
}

var searchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for stations by name",
	Long: `Search for stations by name.

Example:
  moko search "Frankfurt Hbf"
  moko search München`,
	Args: cobra.ExactArgs(1),
	RunE: runSearch,
}

var nearbyCmd = &cobra.Command{
	Use:   "nearby <lat>:<lon>",
	Short: "Search for stations near a location",
	Long: `Search for stations near a geographic location.

The location must be specified as latitude:longitude in decimal degrees.

Example:
  moko nearby 50.107:8.663
  moko nearby 52.520:13.405`,
	Args: cobra.ExactArgs(1),
	RunE: runNearby,
}

var journeyCmd = &cobra.Command{
	Use:   "journey <journey_id>",
	Short: "Show journey details",
	Long: `Show detailed information about a journey/trip.

The journey ID can be obtained from the departures output using --journey or --json.

Watch Mode:
  --watch, -w            Refresh every 30 seconds (full-screen mode)

Examples:
  moko journey "2|#VN#1#ST#..."
  moko journey "2|#VN#1#ST#..." --watch    # Track journey in real-time`,
	Args: cobra.ExactArgs(1),
	RunE: runJourney,
}

var formationCmd = &cobra.Command{
	Use:   "formation <eva> <train_type> <train_number>",
	Short: "Show train carriage formation",
	Long: `Show the carriage formation/composition of a train.

Example:
  moko formation 8000105 ICE 623
  moko formation 8000105 ICE 623 -d 28.12.2025 -t 12:00

The train must depart from the station within the next ~60 minutes,
or use -d and -t to specify the departure time.`,
	Args: cobra.ExactArgs(3),
	RunE: runFormation,
}

var tuiCmd = &cobra.Command{
	Use:   "tui",
	Short: "Launch interactive full-screen TUI",
	Long: `Launch an interactive full-screen terminal UI for browsing
stations, departures, and journey details.

Keyboard:
  Tab          Cycle focus between panels
  j/k or arrows  Navigate lists
  Enter        Select / confirm
  Esc          Go back
  /            Jump to search
  q            Quit`,
	RunE: runTUI,
}

func runTUI(cmd *cobra.Command, args []string) error {
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	model := tui.New(client)
	p := tea.NewProgram(model, tea.WithAltScreen())
	_, err = p.Run()
	return err
}

// filterDepartures filters departures by line and/or direction
func filterDepartures(deps []models.Departure, line, direction string) []models.Departure {
	if line == "" && direction == "" {
		return deps
	}

	filtered := make([]models.Departure, 0, len(deps))
	for _, d := range deps {
		// Line filter: exact match (case-insensitive)
		if line != "" && !strings.EqualFold(d.Line, line) {
			continue
		}
		// Direction filter: substring match (case-insensitive)
		if direction != "" && !strings.Contains(strings.ToLower(d.Destination), strings.ToLower(direction)) {
			continue
		}
		filtered = append(filtered, d)
	}
	return filtered
}

// runWatch runs a continuous refresh loop for watch mode
func runWatch(fetchAndRender func() error) error {
	const refreshInterval = 30 * time.Second

	sigChan := output.SetupSignalHandler()
	ticker := time.NewTicker(refreshInterval)
	defer ticker.Stop()

	// Hide cursor during watch mode
	output.HideCursor(os.Stdout)
	defer output.ShowCursor(os.Stdout)

	// Initial render
	for {
		output.ClearScreen(os.Stdout)

		// Show header with timestamp
		now := time.Now()
		fmt.Printf("Last update: %s | Next refresh in 30s | Press Ctrl+C to exit\n\n",
			now.Format("15:04:05"))

		// Fetch and render data
		if err := fetchAndRender(); err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}

		// Wait for next tick or interrupt
		select {
		case <-ticker.C:
			continue
		case <-sigChan:
			output.ClearScreen(os.Stdout)
			fmt.Println("Watch mode ended.")
			return nil
		}
	}
}

func runDepartures(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse station argument (format: eva:id)
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("station must be in format EVA:ID (e.g., 8000105:A=1@O=...)\nUse 'moko search <name>' to find station IDs")
	}

	eva, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid EVA number: %w", err)
	}
	stationID := parts[1]

	// Create API client
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	req := api.DepartureRequest{
		EVA:            eva,
		StationID:      stationID,
		NumVias:        flagNumVias,
		ModesOfTransit: flagModes,
	}

	// Parse date/time if provided
	if flagDate != "" || flagTime != "" {
		req.DateTime = parseDateTime(flagDate, flagTime, client.Timezone())
	}

	// Watch mode
	if flagWatch {
		return runWatch(func() error {
			colors := output.NewColors(getColorMode())
			deps, err := client.GetDepartures(ctx, req)
			if err != nil {
				return err
			}
			deps = filterDepartures(deps, flagLine, flagDirection)
			output.RenderDepartures(os.Stdout, deps, output.TableOptions{
				Colors:    colors,
				ShowVia:   flagShowVia,
				ShowRoute: flagJourney,
			})
			return nil
		})
	}

	// Raw JSON output
	if flagRawJSON {
		raw, err := client.GetDeparturesRaw(ctx, req)
		if err != nil {
			return err
		}
		return printPrettyJSON(raw)
	}

	// Get departures
	departures, err := client.GetDepartures(ctx, req)
	if err != nil {
		return err
	}

	// Apply line/direction filters
	departures = filterDepartures(departures, flagLine, flagDirection)

	// JSON output
	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(departures)
	}

	// Text output with colors
	colors := output.NewColors(getColorMode())
	output.RenderDepartures(os.Stdout, departures, output.TableOptions{
		Colors:    colors,
		ShowVia:   flagShowVia,
		ShowRoute: flagJourney,
	})

	return nil
}

func runArrivals(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse station argument (format: eva:id)
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("station must be in format EVA:ID (e.g., 8000105:A=1@O=...)\nUse 'moko search <name>' to find station IDs")
	}

	eva, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid EVA number: %w", err)
	}
	stationID := parts[1]

	// Create API client
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	req := api.StationBoardRequest{
		EVA:            eva,
		StationID:      stationID,
		NumVias:        flagNumVias,
		ModesOfTransit: flagModes,
	}

	// Parse date/time if provided
	if flagDate != "" || flagTime != "" {
		req.DateTime = parseDateTime(flagDate, flagTime, client.Timezone())
	}

	// Watch mode
	if flagWatch {
		return runWatch(func() error {
			colors := output.NewColors(getColorMode())
			arrs, err := client.GetArrivals(ctx, req)
			if err != nil {
				return err
			}
			arrs = filterDepartures(arrs, flagLine, flagDirection)
			output.RenderDepartures(os.Stdout, arrs, output.TableOptions{
				Colors:    colors,
				ShowVia:   flagShowVia,
				ShowRoute: flagJourney,
			})
			return nil
		})
	}

	// Raw JSON output
	if flagRawJSON {
		raw, err := client.GetArrivalsRaw(ctx, req)
		if err != nil {
			return err
		}
		return printPrettyJSON(raw)
	}

	// Get arrivals
	arrivals, err := client.GetArrivals(ctx, req)
	if err != nil {
		return err
	}

	// Apply line/direction filters
	arrivals = filterDepartures(arrivals, flagLine, flagDirection)

	// JSON output
	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(arrivals)
	}

	// Text output with colors
	colors := output.NewColors(getColorMode())
	output.RenderDepartures(os.Stdout, arrivals, output.TableOptions{
		Colors:    colors,
		ShowVia:   flagShowVia,
		ShowRoute: flagJourney,
	})

	return nil
}

func runSearch(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	query := args[0]

	// Create API client
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Raw JSON output
	if flagRawJSON {
		raw, err := client.SearchLocationsRaw(ctx, query)
		if err != nil {
			return err
		}
		return printPrettyJSON(raw)
	}

	// Get locations
	locations, err := client.SearchLocations(ctx, query)
	if err != nil {
		return err
	}

	// JSON output
	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(locations)
	}

	// Text output with colors
	colors := output.NewColors(getColorMode())
	output.RenderLocations(os.Stdout, locations, output.TableOptions{
		Colors: colors,
	})

	return nil
}

func runNearby(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse coordinates (format: lat:lon)
	parts := strings.SplitN(args[0], ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("coordinates must be in format LAT:LON (e.g., 50.107:8.663)")
	}

	lat, err := strconv.ParseFloat(parts[0], 64)
	if err != nil {
		return fmt.Errorf("invalid latitude: %w", err)
	}
	lon, err := strconv.ParseFloat(parts[1], 64)
	if err != nil {
		return fmt.Errorf("invalid longitude: %w", err)
	}

	// Create API client
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	req := api.NearbyRequest{
		Latitude:  lat,
		Longitude: lon,
	}

	// Raw JSON output
	if flagRawJSON {
		raw, err := client.SearchNearbyRaw(ctx, req)
		if err != nil {
			return err
		}
		return printPrettyJSON(raw)
	}

	// Get nearby stations
	locations, err := client.SearchNearby(ctx, req)
	if err != nil {
		return err
	}

	// JSON output
	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(locations)
	}

	// Text output with colors
	colors := output.NewColors(getColorMode())
	output.RenderLocations(os.Stdout, locations, output.TableOptions{
		Colors: colors,
	})

	return nil
}

func runJourney(cmd *cobra.Command, args []string) error {
	ctx := context.Background()
	journeyID := args[0]

	// Create API client
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	// Watch mode
	if flagWatch {
		return runWatch(func() error {
			colors := output.NewColors(getColorMode())
			j, err := client.GetJourney(ctx, journeyID, false)
			if err != nil {
				return err
			}
			output.RenderJourney(os.Stdout, j, output.TableOptions{
				Colors: colors,
			})
			return nil
		})
	}

	// Raw JSON output
	if flagRawJSON {
		raw, err := client.GetJourneyRaw(ctx, journeyID, false)
		if err != nil {
			return err
		}
		return printPrettyJSON(raw)
	}

	// Get journey
	journey, err := client.GetJourney(ctx, journeyID, false)
	if err != nil {
		return err
	}

	// JSON output
	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(journey)
	}

	// Text output with colors
	colors := output.NewColors(getColorMode())
	output.RenderJourney(os.Stdout, journey, output.TableOptions{
		Colors: colors,
	})

	return nil
}

func runFormation(cmd *cobra.Command, args []string) error {
	ctx := context.Background()

	// Parse arguments
	eva, err := strconv.ParseInt(args[0], 10, 64)
	if err != nil {
		return fmt.Errorf("invalid EVA number: %w", err)
	}
	trainType := args[1]
	trainNumber := args[2]

	// Create API client
	client, err := createClient()
	if err != nil {
		return fmt.Errorf("failed to create API client: %w", err)
	}

	req := api.FormationRequest{
		EVA:         eva,
		TrainType:   trainType,
		TrainNumber: trainNumber,
	}

	// Parse date/time if provided
	if flagDate != "" || flagTime != "" {
		req.Departure = parseDateTime(flagDate, flagTime, client.Timezone())
	}

	// Raw JSON output
	if flagRawJSON {
		raw, err := client.GetFormationRaw(ctx, req)
		if err != nil {
			return err
		}
		return printPrettyJSON(raw)
	}

	// Get formation
	formation, err := client.GetFormation(ctx, req)
	if err != nil {
		return fmt.Errorf("formation not available: %w", err)
	}

	// JSON output
	if flagJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(formation)
	}

	// Text output with colors
	colors := output.NewColors(getColorMode())
	output.RenderFormation(os.Stdout, formation, output.TableOptions{
		Colors: colors,
	})

	return nil
}

func parseDateTime(dateStr, timeStr string, loc *time.Location) time.Time {
	now := time.Now().In(loc)

	year := now.Year()
	month := now.Month()
	day := now.Day()
	hour := now.Hour()
	minute := now.Minute()

	// Parse date
	if dateStr != "" {
		// Try DD.MM.YYYY format
		if strings.Contains(dateStr, ".") {
			parts := strings.Split(dateStr, ".")
			if len(parts) >= 2 {
				if d, err := strconv.Atoi(parts[0]); err == nil {
					day = d
				}
				if m, err := strconv.Atoi(parts[1]); err == nil {
					month = time.Month(m)
				}
				if len(parts) == 3 {
					if y, err := strconv.Atoi(parts[2]); err == nil {
						if y < 100 {
							y += 2000
						}
						year = y
					}
				}
			}
		} else if strings.Contains(dateStr, "-") {
			// Try YYYY-MM-DD format
			t, err := time.ParseInLocation("2006-01-02", dateStr, loc)
			if err == nil {
				year = t.Year()
				month = t.Month()
				day = t.Day()
			}
		}
	}

	// Parse time
	if timeStr != "" {
		parts := strings.Split(timeStr, ":")
		if len(parts) >= 2 {
			if h, err := strconv.Atoi(parts[0]); err == nil {
				hour = h
			}
			if m, err := strconv.Atoi(parts[1]); err == nil {
				minute = m
			}
		}
	}

	return time.Date(year, month, day, hour, minute, 0, 0, loc)
}

func printPrettyJSON(data []byte) error {
	var prettyJSON interface{}
	if err := json.Unmarshal(data, &prettyJSON); err != nil {
		// If we can't parse it, just print raw
		fmt.Println(string(data))
		return err
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(prettyJSON)
}
