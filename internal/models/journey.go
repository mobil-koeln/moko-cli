package models

import (
	"regexp"
	"strings"
	"time"

	"github.com/mobil-koeln/moko-cli/internal/operators"
)

// hafasEVARegex extracts the EVA number from a Hafas ID string (e.g. "@L=900312003@")
var hafasEVARegex = regexp.MustCompile(`@L=(\d+)@`)

// Journey represents a complete trip/journey with all stops
type Journey struct {
	ID          string     `json:"id"`
	Name        string     `json:"name"`
	Type        string     `json:"type"`
	TripNo      string     `json:"tripNo,omitempty"`
	LineNo      string     `json:"lineNo,omitempty"`
	Operator    string     `json:"operator,omitempty"`
	Day         *time.Time `json:"day,omitempty"`
	IsCancelled bool       `json:"isCancelled"`
	Stops       []Stop     `json:"stops"`
	Messages    []Message  `json:"messages,omitempty"`
}

// Stop represents a single stop along a journey route
type Stop struct {
	EVA          int64      `json:"eva"`
	Name         string     `json:"name"`
	Lat          float64    `json:"lat,omitempty"`
	Lon          float64    `json:"lon,omitempty"`
	Platform     string     `json:"platform,omitempty"`
	RTPlatform   string     `json:"rtPlatform,omitempty"`
	SchedArr     *time.Time `json:"schedArr,omitempty"`
	RTArr        *time.Time `json:"rtArr,omitempty"`
	Arr          *time.Time `json:"arr,omitempty"`
	SchedDep     *time.Time `json:"schedDep,omitempty"`
	RTDep        *time.Time `json:"rtDep,omitempty"`
	Dep          *time.Time `json:"dep,omitempty"`
	ArrDelay     int        `json:"arrDelay,omitempty"`
	DepDelay     int        `json:"depDelay,omitempty"`
	Delay        int        `json:"delay,omitempty"`
	IsCancelled  bool       `json:"isCancelled"`
	IsAdditional bool       `json:"isAdditional"`
}

// JourneyResponse represents the raw API response for a journey
type JourneyResponse struct {
	Reisetag  string `json:"reisetag"`
	ZugName   string `json:"zugName"`
	Cancelled bool   `json:"cancelled"`
	Halte     []struct {
		Name                  string `json:"name"`
		ExtID                 string `json:"extId"`
		EVANumber             int64  `json:"evaNumber"`
		ID                    string `json:"id"`
		Gleis                 string `json:"gleis"`
		EZGleis               string `json:"ezGleis"`
		AbfahrtsZeitpunkt     string `json:"abfahrtsZeitpunkt"`
		EZAbfahrtsZeitpunkt   string `json:"ezAbfahrtsZeitpunkt"`
		AnkunftsZeitpunkt     string `json:"ankunftsZeitpunkt"`
		EZAnkunftsZeitpunkt   string `json:"ezAnkunftsZeitpunkt"`
		AdminID               string `json:"adminID"`
		Nummer                string `json:"nummer"`
		Kategorie             string `json:"kategorie"`
		Canceled              bool   `json:"canceled"`
		Additional            bool   `json:"additional"`
		PriorisierteMeldungen []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"priorisierteMeldungen"`
		RisMeldungen []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"risMeldungen"`
	} `json:"halte"`
	HimMeldungen []struct {
		Prioritaet   string `json:"prioritaet"`
		Ueberschrift string `json:"ueberschrift"`
		Text         string `json:"text"`
	} `json:"himMeldungen"`
	PriorisierteMeldungen []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"priorisierteMeldungen"`
}

// ToJourney converts the raw response to a Journey
func (r *JourneyResponse) ToJourney(id string, loc *time.Location) *Journey {
	j := &Journey{
		ID:          id,
		Name:        r.ZugName,
		IsCancelled: r.Cancelled,
		Stops:       make([]Stop, 0, len(r.Halte)),
	}

	// Parse day
	if r.Reisetag != "" {
		if t, err := time.ParseInLocation("2006-01-02", r.Reisetag, loc); err == nil {
			j.Day = &t
		}
	}

	// Extract type and number from name
	if r.ZugName != "" {
		parts := strings.Fields(r.ZugName)
		if len(parts) >= 1 {
			j.Type = parts[0]
		}
		if len(parts) >= 2 {
			j.TripNo = parts[len(parts)-1]
		}
	}

	// Track admin IDs and types for most common
	adminIDCount := make(map[string]int)
	typeCount := make(map[string]int)
	tripNoCount := make(map[string]int)

	// Process stops
	for _, h := range r.Halte {
		stop := Stop{
			Name:         h.Name,
			EVA:          h.EVANumber,
			Platform:     h.Gleis,
			RTPlatform:   h.EZGleis,
			IsCancelled:  h.Canceled,
			IsAdditional: h.Additional,
		}

		// Parse EVA from extId if needed
		if stop.EVA == 0 && h.ExtID != "" {
			// Try to parse as int
			stop.EVA = parseIntFromString(h.ExtID)
		}

		// Parse coordinates from ID
		if h.ID != "" {
			matches := coordRegex.FindStringSubmatch(h.ID)
			if len(matches) == 3 {
				stop.Lon = parseFloat(matches[1]) / 1e6
				stop.Lat = parseFloat(matches[2]) / 1e6
			}
		}

		// Parse times
		if h.AbfahrtsZeitpunkt != "" {
			if t, err := parseTime(h.AbfahrtsZeitpunkt, loc); err == nil {
				stop.SchedDep = &t
			}
		}
		if h.EZAbfahrtsZeitpunkt != "" {
			if t, err := parseTime(h.EZAbfahrtsZeitpunkt, loc); err == nil {
				stop.RTDep = &t
			}
		}
		if h.AnkunftsZeitpunkt != "" {
			if t, err := parseTime(h.AnkunftsZeitpunkt, loc); err == nil {
				stop.SchedArr = &t
			}
		}
		if h.EZAnkunftsZeitpunkt != "" {
			if t, err := parseTime(h.EZAnkunftsZeitpunkt, loc); err == nil {
				stop.RTArr = &t
			}
		}

		// Set effective times
		if stop.RTDep != nil {
			stop.Dep = stop.RTDep
		} else {
			stop.Dep = stop.SchedDep
		}
		if stop.RTArr != nil {
			stop.Arr = stop.RTArr
		} else {
			stop.Arr = stop.SchedArr
		}

		// Calculate delays
		if stop.SchedDep != nil && stop.RTDep != nil {
			stop.DepDelay = int(stop.RTDep.Sub(*stop.SchedDep).Minutes())
		}
		if stop.SchedArr != nil && stop.RTArr != nil {
			stop.ArrDelay = int(stop.RTArr.Sub(*stop.SchedArr).Minutes())
		}
		stop.Delay = stop.ArrDelay
		if stop.Delay == 0 {
			stop.Delay = stop.DepDelay
		}

		// Check for cancellation in messages
		for _, msg := range h.PriorisierteMeldungen {
			if msg.Type == "HALT_AUSFALL" {
				stop.IsCancelled = true
			}
		}
		for _, msg := range h.RisMeldungen {
			if msg.Key == "text.realtime.stop.cancelled" {
				stop.IsCancelled = true
			}
		}

		// Use effective platform
		if stop.RTPlatform == "" {
			stop.RTPlatform = stop.Platform
		}

		j.Stops = append(j.Stops, stop)

		// Count for most common values
		if h.AdminID != "" {
			adminIDCount[h.AdminID]++
		}
		if h.Kategorie != "" {
			typeCount[h.Kategorie]++
		}
		if h.Nummer != "" {
			tripNoCount[h.Nummer]++
		}
	}

	// Set type from most common if not already set
	if j.Type == "" {
		j.Type = mostCommon(typeCount)
	}

	// Set trip number from most common
	if j.TripNo == "" {
		j.TripNo = mostCommon(tripNoCount)
	}

	// Set operator from most common admin ID
	if adminID := mostCommon(adminIDCount); adminID != "" {
		j.Operator = operators.GetOperatorName(adminID)
	}

	// Process messages
	for _, msg := range r.HimMeldungen {
		j.Messages = append(j.Messages, Message{
			Type: msg.Prioritaet,
			Text: msg.Ueberschrift + ": " + msg.Text,
		})
	}
	for _, msg := range r.PriorisierteMeldungen {
		j.Messages = append(j.Messages, Message{
			Type: msg.Type,
			Text: msg.Text,
		})
	}

	return j
}

// Helper to get the platform (effective)
func (s *Stop) EffectivePlatform() string {
	if s.RTPlatform != "" {
		return s.RTPlatform
	}
	return s.Platform
}

// Helper to find most common value
func mostCommon(m map[string]int) string {
	var maxKey string
	var maxCount int
	for k, v := range m {
		if v > maxCount {
			maxKey = k
			maxCount = v
		}
	}
	return maxKey
}

// Helper to parse int from string.
// Handles both plain numbers ("900312003") and Hafas ID strings ("A=1@O=...@L=900312003@").
func parseIntFromString(s string) int64 {
	// Fast path: plain decimal number
	var result int64
	allDigits := true
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + int64(c-'0')
		} else {
			allDigits = false
			break
		}
	}
	if allDigits && result != 0 {
		return result
	}

	// Slow path: Hafas ID format "@L=<eva>@"
	if matches := hafasEVARegex.FindStringSubmatch(s); len(matches) == 2 {
		var n int64
		for _, c := range matches[1] {
			if c >= '0' && c <= '9' {
				n = n*10 + int64(c-'0')
			}
		}
		return n
	}

	return 0
}

// Helper to parse float from string
func parseFloat(s string) float64 {
	var result float64
	negative := false
	if len(s) > 0 && s[0] == '-' {
		negative = true
		s = s[1:]
	}
	for _, c := range s {
		if c >= '0' && c <= '9' {
			result = result*10 + float64(c-'0')
		}
	}
	if negative {
		result = -result
	}
	return result
}
