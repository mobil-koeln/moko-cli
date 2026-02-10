package models

import (
	"regexp"
	"strconv"
)

// Location represents a station/stop from search results or route entries
type Location struct {
	EVA      int64    `json:"eva"`
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Lat      float64  `json:"lat"`
	Lon      float64  `json:"lon"`
	Type     string   `json:"type"`
	Products []string `json:"products,omitempty"`
}

// LocationResponse represents the raw JSON response for location search
type LocationResponse struct {
	ExtID     string   `json:"extId"`     // API returns as string
	EVANumber int64    `json:"evaNumber"` // Used in some responses
	ID        string   `json:"id"`
	Name      string   `json:"name"`
	Lat       float64  `json:"lat"`
	Lon       float64  `json:"lon"`
	Type      string   `json:"type"`
	Products  []string `json:"products"`
}

// ToLocation converts the raw response to a Location
func (r *LocationResponse) ToLocation() *Location {
	// Parse EVA from string
	var eva int64
	if r.ExtID != "" {
		if parsed, err := strconv.ParseInt(r.ExtID, 10, 64); err == nil {
			eva = parsed
		}
	}
	if eva == 0 {
		eva = r.EVANumber
	}

	loc := &Location{
		EVA:      eva,
		ID:       r.ID,
		Name:     r.Name,
		Lat:      r.Lat,
		Lon:      r.Lon,
		Type:     r.Type,
		Products: r.Products,
	}

	// Parse lat/lon from ID if not provided directly
	// Format: ...@X=<lon>@Y=<lat>...
	if loc.Lat == 0 && loc.Lon == 0 && loc.ID != "" {
		loc.parseCoordinatesFromID()
	}

	return loc
}

var coordRegex = regexp.MustCompile(`@X=(-?\d+)@Y=(-?\d+)`)

func (l *Location) parseCoordinatesFromID() {
	matches := coordRegex.FindStringSubmatch(l.ID)
	if len(matches) == 3 {
		if lon, err := strconv.ParseFloat(matches[1], 64); err == nil {
			l.Lon = lon / 1e6
		}
		if lat, err := strconv.ParseFloat(matches[2], 64); err == nil {
			l.Lat = lat / 1e6
		}
	}
}
