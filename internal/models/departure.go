package models

import (
	"strings"
	"time"
)

// Departure represents a single departure/arrival at a station
type Departure struct {
	JourneyID   string     `json:"journeyId"`
	Type        string     `json:"type"`
	Line        string     `json:"line"`
	Train       string     `json:"train"`
	TrainShort  string     `json:"trainShort"`
	TrainMid    string     `json:"trainMid"`
	TrainLong   string     `json:"trainLong"`
	StopEVA     string     `json:"stopEva"`
	Destination string     `json:"destination"`
	Platform    string     `json:"platform"`
	RTPlatform  string     `json:"rtPlatform"`
	Via         []string   `json:"via,omitempty"`
	ViaLast     string     `json:"viaLast,omitempty"`
	SchedDep    *time.Time `json:"schedDep,omitempty"`
	RTDep       *time.Time `json:"rtDep,omitempty"`
	Dep         *time.Time `json:"dep,omitempty"`
	Delay       int        `json:"delay"`
	IsCancelled bool       `json:"isCancelled"`
	Messages    []Message  `json:"messages,omitempty"`
}

// Message represents an alert/notification for a departure
type Message struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// DepartureResponse represents the raw JSON for a single departure entry
type DepartureResponse struct {
	JourneyID     string   `json:"journeyId"`
	BahnhofsID    string   `json:"bahnhofsId"`
	Terminus      string   `json:"terminus"`
	Gleis         string   `json:"gleis"`
	EZGleis       string   `json:"ezGleis"`
	Zeit          string   `json:"zeit"`
	EZZeit        string   `json:"ezZeit"`
	Ueber         []string `json:"ueber"`
	Verkehrmittel struct {
		KurzText   string `json:"kurzText"`
		MittelText string `json:"mittelText"`
		LangText   string `json:"langText"`
		Name       string `json:"name"`
	} `json:"verkehrmittel"`
	Meldungen []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"meldungen"`
}

// DeparturesResponse represents the full API response for departures
type DeparturesResponse struct {
	Entries []DepartureResponse `json:"entries"`
}

// ToDeparture converts the raw response to a Departure
func (r *DepartureResponse) ToDeparture(loc *time.Location) *Departure {
	dep := &Departure{
		JourneyID:   r.JourneyID,
		Type:        r.Verkehrmittel.KurzText,
		Line:        r.Verkehrmittel.MittelText,
		Train:       r.Verkehrmittel.Name,
		TrainShort:  r.Verkehrmittel.KurzText,
		TrainMid:    r.Verkehrmittel.MittelText,
		TrainLong:   r.Verkehrmittel.LangText,
		StopEVA:     r.BahnhofsID,
		Destination: r.Terminus,
		Platform:    r.Gleis,
		RTPlatform:  r.EZGleis,
	}

	// Process via stations (skip first entry as in Perl version)
	if len(r.Ueber) > 1 {
		dep.Via = r.Ueber[1:]
		dep.ViaLast = r.Ueber[len(r.Ueber)-1]
	} else if len(r.Ueber) == 1 {
		dep.ViaLast = r.Ueber[0]
	}

	// Parse times
	if r.Zeit != "" {
		if t, err := parseTime(r.Zeit, loc); err == nil {
			dep.SchedDep = &t
		}
	}
	if r.EZZeit != "" {
		if t, err := parseTime(r.EZZeit, loc); err == nil {
			dep.RTDep = &t
		}
	}

	// Set effective departure time
	if dep.RTDep != nil {
		dep.Dep = dep.RTDep
	} else {
		dep.Dep = dep.SchedDep
	}

	// Calculate delay
	if dep.SchedDep != nil && dep.RTDep != nil {
		dep.Delay = int(dep.RTDep.Sub(*dep.SchedDep).Minutes())
	}

	// Process messages
	for _, msg := range r.Meldungen {
		dep.Messages = append(dep.Messages, Message{
			Type: msg.Type,
			Text: msg.Text,
		})
		if msg.Type == "HALT_AUSFALL" {
			dep.IsCancelled = true
		}
	}

	return dep
}

// parseTime parses a time string in format "2006-01-02T15:04:05"
func parseTime(s string, loc *time.Location) (time.Time, error) {
	// Handle timezone suffix if present
	s = strings.TrimSuffix(s, "Z")
	if idx := strings.Index(s, "+"); idx > 0 {
		s = s[:idx]
	}
	return time.ParseInLocation("2006-01-02T15:04:05", s, loc)
}

// EffectivePlatform returns the real-time platform if available, otherwise scheduled
func (d *Departure) EffectivePlatform() string {
	if d.RTPlatform != "" {
		return d.RTPlatform
	}
	return d.Platform
}
