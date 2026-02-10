package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDepartureResponse_ToDeparture(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	tests := []struct {
		name      string
		response  DepartureResponse
		wantLine  string
		wantDest  string
		wantDelay int
	}{
		{
			name: "basic departure",
			response: DepartureResponse{
				JourneyID: "test-journey-1",
				Terminus:  "München Hbf",
				Gleis:     "5",
				Zeit:      "2025-01-15T10:00:00",
				EZZeit:    "2025-01-15T10:05:00",
				Verkehrmittel: struct {
					KurzText   string `json:"kurzText"`
					MittelText string `json:"mittelText"`
					LangText   string `json:"langText"`
					Name       string `json:"name"`
				}{
					KurzText:   "ICE",
					MittelText: "ICE 123",
					LangText:   "ICE 123",
					Name:       "ICE 123",
				},
			},
			wantLine:  "ICE 123",
			wantDest:  "München Hbf",
			wantDelay: 5,
		},
		{
			name: "on time departure",
			response: DepartureResponse{
				JourneyID: "test-journey-2",
				Terminus:  "Berlin Hbf",
				Gleis:     "10",
				Zeit:      "2025-01-15T14:30:00",
				EZZeit:    "2025-01-15T14:30:00",
				Verkehrmittel: struct {
					KurzText   string `json:"kurzText"`
					MittelText string `json:"mittelText"`
					LangText   string `json:"langText"`
					Name       string `json:"name"`
				}{
					KurzText:   "RE",
					MittelText: "RE 50",
					LangText:   "RE 50",
					Name:       "RE 50",
				},
			},
			wantLine:  "RE 50",
			wantDest:  "Berlin Hbf",
			wantDelay: 0,
		},
		{
			name: "cancelled departure",
			response: DepartureResponse{
				JourneyID: "test-journey-3",
				Terminus:  "Hamburg Hbf",
				Zeit:      "2025-01-15T16:00:00",
				Meldungen: []struct {
					Type string `json:"type"`
					Text string `json:"text"`
				}{
					{Type: "HALT_AUSFALL", Text: "Zug fällt aus"},
				},
				Verkehrmittel: struct {
					KurzText   string `json:"kurzText"`
					MittelText string `json:"mittelText"`
					LangText   string `json:"langText"`
					Name       string `json:"name"`
				}{
					KurzText:   "ICE",
					MittelText: "ICE 500",
					LangText:   "ICE 500",
					Name:       "ICE 500",
				},
			},
			wantLine: "ICE 500",
			wantDest: "Hamburg Hbf",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := tt.response.ToDeparture(loc)

			if dep.Line != tt.wantLine {
				t.Errorf("Line = %q, want %q", dep.Line, tt.wantLine)
			}
			if dep.Destination != tt.wantDest {
				t.Errorf("Destination = %q, want %q", dep.Destination, tt.wantDest)
			}
			if dep.Delay != tt.wantDelay {
				t.Errorf("Delay = %d, want %d", dep.Delay, tt.wantDelay)
			}
		})
	}
}

func TestDeparture_EffectivePlatform(t *testing.T) {
	tests := []struct {
		name       string
		platform   string
		rtPlatform string
		want       string
	}{
		{
			name:       "scheduled platform only",
			platform:   "5",
			rtPlatform: "",
			want:       "5",
		},
		{
			name:       "realtime platform override",
			platform:   "5",
			rtPlatform: "7",
			want:       "7",
		},
		{
			name:       "no platform",
			platform:   "",
			rtPlatform: "",
			want:       "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dep := &Departure{
				Platform:   tt.platform,
				RTPlatform: tt.rtPlatform,
			}
			if got := dep.EffectivePlatform(); got != tt.want {
				t.Errorf("EffectivePlatform() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestDeparturesResponse_JSON(t *testing.T) {
	jsonData := `{
		"entries": [
			{
				"journeyId": "2|#VN#1#ST#...",
				"terminus": "München Hbf",
				"gleis": "5",
				"zeit": "2025-01-15T10:00:00",
				"ezZeit": "2025-01-15T10:05:00",
				"ueber": ["Mannheim Hbf", "Stuttgart Hbf"],
				"verkehrmittel": {
					"kurzText": "ICE",
					"mittelText": "ICE 123",
					"langText": "ICE 123",
					"name": "ICE 123"
				}
			}
		]
	}`

	var resp DeparturesResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(resp.Entries) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(resp.Entries))
	}

	entry := resp.Entries[0]
	if entry.Terminus != "München Hbf" {
		t.Errorf("Terminus = %q, want %q", entry.Terminus, "München Hbf")
	}
	if len(entry.Ueber) != 2 {
		t.Errorf("Ueber length = %d, want 2", len(entry.Ueber))
	}
}

func TestParseTime(t *testing.T) {
	loc, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		t.Fatalf("Failed to load timezone: %v", err)
	}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{
			name:    "valid time",
			input:   "2025-01-15T10:00:00",
			wantErr: false,
		},
		{
			name:    "time with Z suffix",
			input:   "2025-01-15T10:00:00Z",
			wantErr: false,
		},
		{
			name:    "time with timezone offset",
			input:   "2025-01-15T10:00:00+01:00",
			wantErr: false,
		},
		{
			name:    "invalid format",
			input:   "not-a-time",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := parseTime(tt.input, loc)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseTime() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
