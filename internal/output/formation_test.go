package output

import (
	"bytes"
	"testing"

	"github.com/mobil-koeln/moko-cli/internal/models"
	"github.com/mobil-koeln/moko-cli/internal/testutil"
)

func TestRenderFormation_Nil(t *testing.T) {
	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderFormation(&buf, nil, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "No formation data available")
}

func TestRenderFormation_Basic(t *testing.T) {
	formation := &models.Formation{
		Platform:  "7",
		Direction: 100, // Forward
		TrainType: "ICE",
		Sectors: []models.Sector{
			{
				Name:          "A",
				StartPercent:  10,
				EndPercent:    40,
				LengthPercent: 30,
			},
			{
				Name:          "B",
				StartPercent:  40,
				EndPercent:    70,
				LengthPercent: 30,
			},
		},
		Carriages: []models.Carriage{
			{
				Number:        "1",
				Type:          "FirstClass",
				ClassType:     1,
				StartPercent:  15,
				EndPercent:    25,
				LengthPercent: 10,
				IsLocomotive:  false,
			},
			{
				Number:        "2",
				Type:          "SecondClass",
				ClassType:     2,
				StartPercent:  25,
				EndPercent:    35,
				LengthPercent: 10,
				IsLocomotive:  false,
			},
		},
		Groups: []models.Group{
			{
				Name:        "ICE 123",
				TrainType:   "ICE",
				TrainNo:     "123",
				Destination: "München Hbf",
				Description: "Train",
				Sectors:     []string{"A", "B"},
				Carriages: []models.Carriage{
					{
						Number:    "1",
						Model:     "411",
						Type:      "Apekzf",
						ClassType: 1,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderFormation(&buf, formation, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "Platform:")
	testutil.AssertContains(t, output, "7")
	testutil.AssertContains(t, output, "A")
	testutil.AssertContains(t, output, "B")
	testutil.AssertContains(t, output, "ICE")
	testutil.AssertContains(t, output, "München Hbf")
}

func TestRenderFormation_WithLocomotive(t *testing.T) {
	formation := &models.Formation{
		Platform:  "7",
		Direction: 0, // Backward
		Carriages: []models.Carriage{
			{
				Number:        "[]",
				Type:          "Locomotive",
				IsLocomotive:  true,
				StartPercent:  10,
				LengthPercent: 5,
			},
			{
				Number:        "1",
				Type:          "FirstClass",
				ClassType:     1,
				StartPercent:  15,
				LengthPercent: 10,
			},
		},
		Groups: []models.Group{
			{
				Name:        "ICE 123",
				TrainType:   "ICE",
				TrainNo:     "123",
				Destination: "München Hbf",
				Carriages: []models.Carriage{
					{
						Number:       "Lok",
						IsLocomotive: true,
					},
					{
						Number:    "1",
						ClassType: 1,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderFormation(&buf, formation, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "Lok")
	testutil.AssertContains(t, output, "<") // Backward direction
}

func TestRenderFormation_WithClosedCarriage(t *testing.T) {
	formation := &models.Formation{
		Platform:  "7",
		Direction: 100,
		Carriages: []models.Carriage{
			{
				Number:        "X",
				Type:          "Closed",
				IsClosed:      true,
				StartPercent:  10,
				LengthPercent: 10,
			},
		},
		Groups: []models.Group{
			{
				Name:        "ICE 123",
				TrainType:   "ICE",
				TrainNo:     "123",
				Destination: "München Hbf",
				Carriages: []models.Carriage{
					{
						Number:   "X",
						IsClosed: true,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderFormation(&buf, formation, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "X")
}

func TestRenderFormation_WithAmenities(t *testing.T) {
	formation := &models.Formation{
		Platform:  "7",
		Direction: 100,
		Groups: []models.Group{
			{
				Name:        "ICE 123",
				TrainType:   "ICE",
				TrainNo:     "123",
				Destination: "München Hbf",
				Carriages: []models.Carriage{
					{
						Number:        "1",
						Model:         "411",
						Type:          "Apekzf",
						ClassType:     1,
						IsDosto:       true,
						HasBistro:     true,
						HasQuietZone:  true,
						HasFamilyZone: true,
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderFormation(&buf, formation, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "Doppelstock")
	testutil.AssertContains(t, output, "Bistro")
	testutil.AssertContains(t, output, "Ruhebereich")
	testutil.AssertContains(t, output, "Familienbereich")
}

func TestRenderFormation_WithDesignation(t *testing.T) {
	formation := &models.Formation{
		Platform:  "7",
		Direction: 100,
		Groups: []models.Group{
			{
				Name:        "ICE 123",
				Designation: "Gießen", // ICE train name
				TrainType:   "ICE",
				TrainNo:     "123",
				Destination: "München Hbf",
				Description: "Train",
				Sectors:     []string{"A", "B"},
				Carriages:   []models.Carriage{},
			},
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderFormation(&buf, formation, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "Gießen")
	testutil.AssertContains(t, output, "(AB)")
}

func TestRenderFormation_MixedClass(t *testing.T) {
	formation := &models.Formation{
		Platform:  "7",
		Direction: 100,
		Groups: []models.Group{
			{
				Name:        "ICE 123",
				TrainType:   "ICE",
				TrainNo:     "123",
				Destination: "München Hbf",
				Carriages: []models.Carriage{
					{
						Number:    "1",
						Model:     "411",
						Type:      "Mixed",
						ClassType: 12, // Mixed class
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderFormation(&buf, formation, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "1./2.")
}

func TestRenderFormation_EmptyCarriages(t *testing.T) {
	formation := &models.Formation{
		Platform:  "7",
		Direction: 100,
		Sectors: []models.Sector{
			{
				Name:          "A",
				LengthPercent: 50,
			},
		},
		Groups: []models.Group{
			{
				Name:        "ICE 123",
				TrainType:   "ICE",
				TrainNo:     "123",
				Destination: "München Hbf",
				Carriages:   []models.Carriage{},
			},
		},
	}

	var buf bytes.Buffer
	opts := TableOptions{Colors: NewColors(ColorNever)}

	RenderFormation(&buf, formation, opts)

	output := buf.String()
	testutil.AssertContains(t, output, "Platform:")
	testutil.AssertContains(t, output, "A")
}

func TestRenderFormation_Direction(t *testing.T) {
	tests := []struct {
		name      string
		direction int
		wantChar  string
	}{
		{"forward", 100, ">"},
		{"backward", 0, "<"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			formation := &models.Formation{
				Platform:  "7",
				Direction: tt.direction,
				Carriages: []models.Carriage{
					{
						Number:        "1",
						StartPercent:  10,
						LengthPercent: 10,
					},
				},
			}

			var buf bytes.Buffer
			opts := TableOptions{Colors: NewColors(ColorNever)}

			RenderFormation(&buf, formation, opts)

			output := buf.String()
			testutil.AssertContains(t, output, tt.wantChar)
		})
	}
}
