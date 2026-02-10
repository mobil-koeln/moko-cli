package models

import (
	"encoding/json"
	"testing"
)

func TestLocationResponse_ToLocation(t *testing.T) {
	tests := []struct {
		name     string
		response LocationResponse
		wantName string
		wantEVA  int64
		wantType string
	}{
		{
			name: "station with EVA",
			response: LocationResponse{
				ExtID: "8000105",
				ID:    "A=1@O=Frankfurt(Main)Hbf@X=8663003@Y=50107145@U=80@L=8000105@",
				Name:  "Frankfurt(Main)Hbf",
				Type:  "ST",
				Lat:   50107145,
				Lon:   8663003,
			},
			wantName: "Frankfurt(Main)Hbf",
			wantEVA:  8000105,
			wantType: "ST",
		},
		{
			name: "POI without EVA",
			response: LocationResponse{
				ExtID: "",
				ID:    "A=4@O=Some POI@X=8660000@Y=50100000@",
				Name:  "Some POI",
				Type:  "POI",
			},
			wantName: "Some POI",
			wantEVA:  0,
			wantType: "POI",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loc := tt.response.ToLocation()

			if loc.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", loc.Name, tt.wantName)
			}
			if loc.EVA != tt.wantEVA {
				t.Errorf("EVA = %d, want %d", loc.EVA, tt.wantEVA)
			}
			if loc.Type != tt.wantType {
				t.Errorf("Type = %q, want %q", loc.Type, tt.wantType)
			}
		})
	}
}

func TestLocationResponse_JSON(t *testing.T) {
	jsonData := `[
		{
			"extId": "8000105",
			"id": "A=1@O=Frankfurt(Main)Hbf@X=8663003@Y=50107145@U=80@L=8000105@",
			"name": "Frankfurt(Main)Hbf",
			"type": "ST",
			"lat": 50107145,
			"lon": 8663003,
			"products": ["ICE", "EC_IC", "REGIONAL", "SBAHN"]
		},
		{
			"extId": "8000261",
			"id": "A=1@O=München Hbf@X=11558744@Y=48140364@U=80@L=8000261@",
			"name": "München Hbf",
			"type": "ST",
			"lat": 48140364,
			"lon": 11558744,
			"products": ["ICE", "EC_IC", "REGIONAL", "SBAHN", "UBAHN"]
		}
	]`

	var resp []LocationResponse
	if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}

	if len(resp) != 2 {
		t.Fatalf("Expected 2 locations, got %d", len(resp))
	}

	// Test first location
	loc1 := resp[0].ToLocation()
	if loc1.Name != "Frankfurt(Main)Hbf" {
		t.Errorf("First location name = %q, want %q", loc1.Name, "Frankfurt(Main)Hbf")
	}
	if loc1.EVA != 8000105 {
		t.Errorf("First location EVA = %d, want %d", loc1.EVA, 8000105)
	}

	// Test second location
	loc2 := resp[1].ToLocation()
	if loc2.Name != "München Hbf" {
		t.Errorf("Second location name = %q, want %q", loc2.Name, "München Hbf")
	}
}

func TestLocation_Coordinates(t *testing.T) {
	loc := &Location{
		Lat: 50.107145,
		Lon: 8.663003,
	}

	// Simple sanity check for coordinate values
	if loc.Lat < 47 || loc.Lat > 55 {
		t.Errorf("Latitude %f out of expected German range", loc.Lat)
	}
	if loc.Lon < 5 || loc.Lon > 15 {
		t.Errorf("Longitude %f out of expected German range", loc.Lon)
	}
}

func TestLocationResponse_CoordinateConversion(t *testing.T) {
	// When coordinates are provided directly in JSON, they're already in degrees
	response := LocationResponse{
		Lat: 50.107145,
		Lon: 8.663003,
	}

	loc := response.ToLocation()

	expectedLat := 50.107145
	expectedLon := 8.663003

	if abs(loc.Lat-expectedLat) > 0.001 {
		t.Errorf("Latitude = %f, want %f", loc.Lat, expectedLat)
	}
	if abs(loc.Lon-expectedLon) > 0.001 {
		t.Errorf("Longitude = %f, want %f", loc.Lon, expectedLon)
	}
}

func TestLocationResponse_CoordinatesFromID(t *testing.T) {
	// When coordinates are in ID string, they're in micro-degrees
	response := LocationResponse{
		ID:   "A=1@O=Test Station@X=8663003@Y=50107145@U=80@L=8000105@",
		Name: "Test Station",
		Lat:  0, // Not provided directly
		Lon:  0,
	}

	loc := response.ToLocation()

	expectedLat := 50.107145
	expectedLon := 8.663003

	if abs(loc.Lat-expectedLat) > 0.001 {
		t.Errorf("Latitude = %f, want %f", loc.Lat, expectedLat)
	}
	if abs(loc.Lon-expectedLon) > 0.001 {
		t.Errorf("Longitude = %f, want %f", loc.Lon, expectedLon)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
