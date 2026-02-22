package models

import (
	"testing"
)

func TestParseIntFromString_PlainNumber(t *testing.T) {
	if got := parseIntFromString("8000105"); got != 8000105 {
		t.Errorf("got %d, want 8000105", got)
	}
}

func TestParseIntFromString_TransitEVA(t *testing.T) {
	if got := parseIntFromString("900312003"); got != 900312003 {
		t.Errorf("got %d, want 900312003", got)
	}
}

func TestParseIntFromString_HafasID(t *testing.T) {
	// Hafas format: plain number not at the start, EVA is in @L=<eva>@
	hafas := "A=1@O=Köln, Mülheim Keupstr.@X=7005780@Y=50965200@U=80@L=900312003@B=1@"
	if got := parseIntFromString(hafas); got != 900312003 {
		t.Errorf("got %d, want 900312003", got)
	}
}

func TestParseIntFromString_Empty(t *testing.T) {
	if got := parseIntFromString(""); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestParseIntFromString_NoEVA(t *testing.T) {
	if got := parseIntFromString("A=1@O=SomeName@X=123@Y=456@"); got != 0 {
		t.Errorf("got %d, want 0", got)
	}
}

func TestToJourney_EVAFromHafasExtID(t *testing.T) {
	resp := &JourneyResponse{
		ZugName: "Bus 150",
		Halte: []struct {
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
		}{
			{
				Name:      "Mülheim Keupstr., Köln",
				EVANumber: 0, // not set by API for transit stops
				ExtID:     "A=1@O=Mülheim Keupstr.@X=7005780@Y=50965200@U=80@L=900312003@B=1@",
			},
		},
	}

	journey := resp.ToJourney("test-id", nil)
	if len(journey.Stops) != 1 {
		t.Fatalf("expected 1 stop, got %d", len(journey.Stops))
	}
	if journey.Stops[0].EVA != 900312003 {
		t.Errorf("EVA: got %d, want 900312003", journey.Stops[0].EVA)
	}
}

func TestToJourney_EVAFromEVANumber(t *testing.T) {
	resp := &JourneyResponse{
		ZugName: "ICE 123",
		Halte: []struct {
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
		}{
			{
				Name:      "Frankfurt Hbf",
				EVANumber: 8000105, // set directly
				ExtID:     "8000105",
			},
		},
	}

	journey := resp.ToJourney("test-id", nil)
	if journey.Stops[0].EVA != 8000105 {
		t.Errorf("EVA: got %d, want 8000105", journey.Stops[0].EVA)
	}
}
