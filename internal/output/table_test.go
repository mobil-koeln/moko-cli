package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/mobil-koeln/moko-cli/internal/models"
	"github.com/mobil-koeln/moko-cli/internal/testutil"
)

func TestRenderDepartures_Empty(t *testing.T) {
	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderDepartures(&buf, []models.Departure{}, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "No departures found")
}

func TestRenderDepartures_SingleDeparture(t *testing.T) {
	// Create test departure
	depTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	dep := models.Departure{
		JourneyID:   "1|123456|0|80|1012024",
		Dep:         &depTime,
		Delay:       0,
		Line:        "ICE 123",
		Platform:    "7",
		Destination: "München Hbf",
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderDepartures(&buf, []models.Departure{dep}, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "14:30")
	testutil.AssertContains(t, output, "ICE 123")
	testutil.AssertContains(t, output, "Pl.7")
	testutil.AssertContains(t, output, "München Hbf")
}

func TestRenderDepartures_WithDelay(t *testing.T) {
	depTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	dep := models.Departure{
		JourneyID:   "1|123456|0|80|1012024",
		Dep:         &depTime,
		Delay:       5,
		Line:        "ICE 123",
		Platform:    "7",
		Destination: "München Hbf",
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderDepartures(&buf, []models.Departure{dep}, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "14:30")
	testutil.AssertContains(t, output, "+5")
}

func TestRenderDepartures_Canceled(t *testing.T) {
	depTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	dep := models.Departure{
		JourneyID:   "1|123456|0|80|1012024",
		Dep:         &depTime,
		Line:        "ICE 123",
		Platform:    "7",
		Destination: "München Hbf",
		IsCancelled: true,
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderDepartures(&buf, []models.Departure{dep}, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "CANCELED")
	testutil.AssertContains(t, output, "München Hbf")
}

func TestRenderDepartures_WithVia(t *testing.T) {
	depTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	dep := models.Departure{
		JourneyID:   "1|123456|0|80|1012024",
		Dep:         &depTime,
		Line:        "ICE 123",
		Platform:    "7",
		Destination: "München Hbf",
		Via:         []string{"Mannheim", "Stuttgart"},
	}

	var buf bytes.Buffer
	opts := TableOptions{
		Colors:  NewColors(ColorNever),
		ShowVia: true,
	}

	RenderDepartures(&buf, []models.Departure{dep}, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "via Mannheim - Stuttgart")
}

func TestRenderDepartures_WithRoute(t *testing.T) {
	depTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	dep := models.Departure{
		JourneyID:   "1|123456|0|80|1012024",
		Dep:         &depTime,
		Line:        "ICE 123",
		Platform:    "7",
		Destination: "München Hbf",
	}

	var buf bytes.Buffer
	opts := TableOptions{
		Colors:    NewColors(ColorNever),
		ShowRoute: true,
	}

	RenderDepartures(&buf, []models.Departure{dep}, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "Journey:")
	testutil.AssertContains(t, output, "1|123456|0|80|1012024")
}

func TestRenderDepartures_NilDeparture(t *testing.T) {
	dep := models.Departure{
		JourneyID:   "1|123456|0|80|1012024",
		Dep:         nil, // No departure time
		Line:        "ICE 123",
		Platform:    "7",
		Destination: "München Hbf",
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderDepartures(&buf, []models.Departure{dep}, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "??:??")
}

func TestRenderDepartures_PlatformChange(t *testing.T) {
	depTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	dep := models.Departure{
		JourneyID:   "1|123456|0|80|1012024",
		Dep:         &depTime,
		Line:        "ICE 123",
		Platform:    "7",
		RTPlatform:  "8", // Platform changed
		Destination: "München Hbf",
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderDepartures(&buf, []models.Departure{dep}, opts)

	output := buf.String()
	// Should show effective platform (RTPlatform)
	testutil.AssertContains(t, output, "Pl.8")
}

func TestRenderDepartures_LongLineName(t *testing.T) {
	depTime := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	dep := models.Departure{
		JourneyID:   "1|123456|0|80|1012024",
		Dep:         &depTime,
		Line:        "VeryLongTrainNameThatExceeds", // > 10 chars
		Platform:    "7",
		Destination: "München Hbf",
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderDepartures(&buf, []models.Departure{dep}, opts)

	output := buf.String()
	// Should be truncated to 10 chars
	testutil.AssertContains(t, output, "VeryLongTr")
	testutil.AssertNotContains(t, output, "VeryLongTrainNameThatExceeds")
}

func TestRenderDepartures_MultipleDepartures(t *testing.T) {
	depTime1 := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	depTime2 := time.Date(2024, 1, 1, 14, 45, 0, 0, time.UTC)

	departures := []models.Departure{
		{
			JourneyID:   "1|123456|0|80|1012024",
			Dep:         &depTime1,
			Line:        "ICE 123",
			Platform:    "7",
			Destination: "München Hbf",
		},
		{
			JourneyID:   "1|654321|0|80|1012024",
			Dep:         &depTime2,
			Line:        "RE 4567",
			Platform:    "12",
			Destination: "Mainz Hbf",
			Delay:       2,
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderDepartures(&buf, departures, opts)

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	// Should have at least 2 departure lines
	testutil.AssertTrue(t, len(lines) >= 2)
}

func TestRenderLocations_Empty(t *testing.T) {
	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderLocations(&buf, []models.Location{}, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "No stations found")
}

func TestRenderLocations_Single(t *testing.T) {
	locations := []models.Location{
		{
			Name: "Frankfurt(Main)Hbf",
			EVA:  8000105,
			ID:   "A=1@O=Frankfurt(Main)Hbf@X=8663785@Y=50107145@U=80@L=8000105@",
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderLocations(&buf, locations, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "Found stations:")
	testutil.AssertContains(t, output, "Frankfurt(Main)Hbf")
	testutil.AssertContains(t, output, "EVA:")
	testutil.AssertContains(t, output, "8000105")
	testutil.AssertContains(t, output, "moko departures")
}

func TestRenderLocations_Multiple(t *testing.T) {
	locations := []models.Location{
		{
			Name: "Frankfurt(Main)Hbf",
			EVA:  8000105,
			ID:   "A=1@O=Frankfurt(Main)Hbf@",
		},
		{
			Name: "Frankfurt(Main) Süd",
			EVA:  8002041,
			ID:   "A=1@O=Frankfurt(Main) Süd@",
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderLocations(&buf, locations, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "Frankfurt(Main)Hbf")
	testutil.AssertContains(t, output, "Frankfurt(Main) Süd")
	testutil.AssertContains(t, output, "8000105")
	testutil.AssertContains(t, output, "8002041")
}

func TestFindCurrentStopIndex_EmptyStops(t *testing.T) {
	now := time.Now()
	idx := FindCurrentStopIndex([]models.Stop{}, now)
	testutil.AssertEqual(t, idx, -1)
}

func TestFindCurrentStopIndex_BeforeFirstStop(t *testing.T) {
	now := time.Date(2024, 1, 1, 14, 0, 0, 0, time.UTC)
	arr1 := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)

	stops := []models.Stop{
		{Name: "Station 1", Arr: &arr1, Delay: 0},
	}

	idx := FindCurrentStopIndex(stops, now)
	testutil.AssertEqual(t, idx, 0)
}

func TestFindCurrentStopIndex_AtFirstStop(t *testing.T) {
	now := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	arr1 := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)

	stops := []models.Stop{
		{Name: "Station 1", Arr: &arr1, Delay: 0},
	}

	idx := FindCurrentStopIndex(stops, now)
	testutil.AssertEqual(t, idx, 0)
}

func TestFindCurrentStopIndex_BetweenStops(t *testing.T) {
	now := time.Date(2024, 1, 1, 14, 45, 0, 0, time.UTC)
	arr1 := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	arr2 := time.Date(2024, 1, 1, 15, 0, 0, 0, time.UTC)

	stops := []models.Stop{
		{Name: "Station 1", Arr: &arr1, Delay: 0},
		{Name: "Station 2", Arr: &arr2, Delay: 0},
	}

	idx := FindCurrentStopIndex(stops, now)
	testutil.AssertEqual(t, idx, 0) // Still at first stop
}

func TestFindCurrentStopIndex_WithDelay(t *testing.T) {
	// Train is delayed by 6 minutes
	// Current time is 19:01, but train is at the 18:55 position on schedule
	now := time.Date(2024, 1, 1, 19, 1, 0, 0, time.UTC)
	arr1 := time.Date(2024, 1, 1, 18, 55, 0, 0, time.UTC)
	arr2 := time.Date(2024, 1, 1, 19, 10, 0, 0, time.UTC)

	stops := []models.Stop{
		{Name: "Station 1", Arr: &arr1, Delay: 6},
		{Name: "Station 2", Arr: &arr2, Delay: 6},
	}

	idx := FindCurrentStopIndex(stops, now)
	testutil.AssertEqual(t, idx, 0) // At first stop due to delay
}

func TestRenderJourney_Nil(t *testing.T) {
	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderJourney(&buf, nil, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "No journey data found")
}

func TestRenderJourney_Basic(t *testing.T) {
	arr1 := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)
	dep1 := time.Date(2024, 1, 1, 14, 32, 0, 0, time.UTC)
	arr2 := time.Date(2024, 1, 1, 15, 15, 0, 0, time.UTC)

	journey := &models.Journey{
		Name:     "ICE 123",
		Operator: "DB Fernverkehr AG",
		Stops: []models.Stop{
			{
				Name:     "Frankfurt Hbf",
				Platform: "7",
				Arr:      &arr1,
				Dep:      &dep1,
				Delay:    2,
			},
			{
				Name:     "München Hbf",
				Platform: "18",
				Arr:      &arr2,
				Delay:    5,
			},
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderJourney(&buf, journey, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "Journey:")
	testutil.AssertContains(t, output, "ICE 123")
	testutil.AssertContains(t, output, "Operator:")
	testutil.AssertContains(t, output, "DB Fernverkehr AG")
	testutil.AssertContains(t, output, "Route:")
	testutil.AssertContains(t, output, "Frankfurt Hbf")
	testutil.AssertContains(t, output, "München Hbf")
	testutil.AssertContains(t, output, "Pl.7")
	testutil.AssertContains(t, output, "Pl.18")
}

func TestRenderJourney_CanceledStop(t *testing.T) {
	arr1 := time.Date(2024, 1, 1, 14, 30, 0, 0, time.UTC)

	journey := &models.Journey{
		Name: "ICE 123",
		Stops: []models.Stop{
			{
				Name:        "Frankfurt Hbf",
				Platform:    "7",
				Arr:         &arr1,
				IsCancelled: true,
			},
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderJourney(&buf, journey, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "CANCELED")
	testutil.AssertContains(t, output, "Frankfurt Hbf")
}
