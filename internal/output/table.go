package output

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/mobil-koeln/moko-cli/internal/models"
)

// TableOptions configures the table output
type TableOptions struct {
	Colors    *Colors
	ShowVia   bool
	ShowRoute bool
}

// RenderDepartures renders departures as a formatted table
func RenderDepartures(w io.Writer, departures []models.Departure, opts TableOptions) {
	if len(departures) == 0 {
		_, _ = fmt.Fprintln(w, "No departures found.")
		return
	}

	c := opts.Colors
	if c == nil {
		c = NewColors(ColorNever)
	}

	for _, dep := range departures {
		// Time
		timeStr := "??:??"
		if dep.Dep != nil {
			timeStr = dep.Dep.Format("15:04")
		}

		// Delay (fixed 4-char width)
		delayStr := c.FormatDelay(dep.Delay)

		// Line/Train (truncate/pad to 10 chars)
		line := dep.Line
		if line == "" {
			line = dep.TrainShort
		}
		if len(line) > 10 {
			line = line[:10]
		}
		lineStr := fmt.Sprintf("%-10s", line)

		// Platform (fixed 7-char width: "Pl.XXX" or spaces)
		platform := dep.EffectivePlatform()
		platformStr := "       " // 7 spaces
		if platform != "" {
			if len(platform) > 3 {
				platform = platform[:3]
			}
			platformStr = fmt.Sprintf("Pl.%-3s ", platform)
		}

		// Destination
		dest := dep.Destination
		if dep.IsCancelled {
			dest = c.Canceled("%s [CANCELED]", dest)
		}

		// Format the line: TIME DELAY LINE     PLATFORM DEST
		_, _ = fmt.Fprintf(w, "%s %s  %s  %s %s\n",
			c.Time(timeStr),
			delayStr,
			c.Line(lineStr),
			c.Platform(platformStr),
			dest,
		)

		// Show via stations if requested
		if opts.ShowVia && len(dep.Via) > 0 {
			viaStr := strings.Join(dep.Via, " - ")
			_, _ = fmt.Fprintf(w, "                              %s\n", c.Via("via %s", viaStr))
		}

		// Show journey ID if requested
		if opts.ShowRoute && dep.JourneyID != "" {
			_, _ = fmt.Fprintf(w, "                              %s %s\n",
				c.Muted("Journey:"),
				c.Via(dep.JourneyID))
		}
	}
}

// RenderLocations renders locations as a formatted list
func RenderLocations(w io.Writer, locations []models.Location, opts TableOptions) {
	if len(locations) == 0 {
		_, _ = fmt.Fprintln(w, "No stations found.")
		return
	}

	c := opts.Colors
	if c == nil {
		c = NewColors(ColorNever)
	}

	_, _ = fmt.Fprintln(w, c.Header("Found stations:"))
	_, _ = fmt.Fprintln(w)

	for _, loc := range locations {
		_, _ = fmt.Fprintf(w, "  %s\n", c.Line(loc.Name))
		_, _ = fmt.Fprintf(w, "    %s %d\n", c.Muted("EVA:"), loc.EVA)
		if loc.EVA != 0 {
			_, _ = fmt.Fprintf(w, "    %s moko departures %d:%s\n",
				c.Muted("Use:"),
				loc.EVA,
				loc.ID,
			)
		}
		_, _ = fmt.Fprintln(w)
	}
}

// FindCurrentStopIndex determines which stop the journey is currently at or approaching.
// Logic:
// 1. Look at current time and find where train SHOULD be (scheduled)
// 2. Get the delay at that point
// 3. Virtual time = current time - delay (where train actually is on the schedule)
// 4. Find the station matching that virtual time
func FindCurrentStopIndex(stops []models.Stop, now time.Time) int {
	if len(stops) == 0 {
		return -1
	}

	// Step 1: Find where the train SHOULD be based on current time
	// The displayed times (Arr/Dep) ARE the scheduled times
	// Find the station whose scheduled time is closest to now
	delay := 0
	for i := len(stops) - 1; i >= 0; i-- {
		// Bounds check before accessing array
		if i < 0 || i >= len(stops) {
			continue
		}
		if stops[i].Arr != nil && !now.Before(*stops[i].Arr) {
			delay = stops[i].Delay
			break
		}
	}

	// Step 2: Calculate virtual time (where train actually is on schedule)
	// A train with +6 delay at 19:01 is actually at the 18:55 position
	virtualNow := now.Add(-time.Duration(delay) * time.Minute)

	// Step 3: Find the station at virtual time
	// Compare directly to scheduled times (Arr) - no subtraction needed
	for i := len(stops) - 1; i >= 0; i-- {
		// Bounds check before accessing array
		if i < 0 || i >= len(stops) {
			continue
		}
		if stops[i].Arr != nil && !virtualNow.Before(*stops[i].Arr) {
			return i
		}
	}

	return 0
}

// RenderJourney renders a journey with all stops
func RenderJourney(w io.Writer, journey *models.Journey, opts TableOptions) {
	if journey == nil {
		_, _ = fmt.Fprintln(w, "No journey data found.")
		return
	}

	c := opts.Colors
	if c == nil {
		c = NewColors(ColorNever)
	}

	// Header
	_, _ = fmt.Fprintf(w, "%s %s\n",
		c.Header("Journey:"),
		c.Line(journey.Name),
	)

	if journey.Operator != "" {
		_, _ = fmt.Fprintf(w, "%s %s\n", c.Muted("Operator:"), journey.Operator)
	}

	_, _ = fmt.Fprintln(w)
	_, _ = fmt.Fprintln(w, c.Header("Route:"))
	_, _ = fmt.Fprintln(w)

	// Find current position
	now := time.Now()
	currentIdx := FindCurrentStopIndex(journey.Stops, now)

	// Stops
	for i, stop := range journey.Stops {
		// Determine if this is first, last, or intermediate stop
		isFirst := i == 0
		isLast := i == len(journey.Stops)-1
		isCurrent := i == currentIdx

		// Arrival time
		arrStr := "     "
		if stop.Arr != nil && !isFirst {
			arrStr = stop.Arr.Format("15:04")
		}

		// Departure time
		depStr := "     "
		if stop.Dep != nil && !isLast {
			depStr = stop.Dep.Format("15:04")
		}

		// Delay
		delayStr := "    "
		if stop.Delay != 0 {
			delayStr = c.FormatDelay(stop.Delay)
		}

		// Platform
		platform := stop.Platform
		platformStr := "        "
		if platform != "" {
			platformStr = fmt.Sprintf("Pl.%-4s", platform)
			if !isCurrent {
				platformStr = c.Platform("%s", platformStr)
			}
		}

		// Station name
		name := stop.Name
		if stop.IsCancelled {
			name = c.Canceled("%s [CANCELED]", name)
		}

		// Connection symbol
		symbol := "├"
		if isFirst {
			symbol = "┌"
		} else if isLast {
			symbol = "└"
		}

		// Current station indicator
		indicator := " "
		if isCurrent {
			indicator = ">"
		}

		// Format output - highlight current station in red
		if isCurrent && !stop.IsCancelled {
			_, _ = fmt.Fprintf(w, "%s %s %s  %s %-4s  %-8s  %s\n",
				c.Canceled(indicator),
				c.Muted(symbol),
				c.Canceled(arrStr),
				c.Canceled(depStr),
				delayStr,
				c.Canceled(platformStr),
				c.Canceled(name),
			)
		} else {
			_, _ = fmt.Fprintf(w, "%s %s %s  %s %-4s  %-8s  %s\n",
				indicator,
				c.Muted(symbol),
				c.Time(arrStr),
				c.Time(depStr),
				delayStr,
				platformStr,
				name,
			)
		}
	}
}
