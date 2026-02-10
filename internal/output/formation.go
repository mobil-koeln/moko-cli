package output

import (
	"fmt"
	"io"
	"strings"

	"github.com/mobil-koeln/moko-cli/internal/models"
)

// RenderFormation renders a train formation as ASCII art
func RenderFormation(w io.Writer, formation *models.Formation, opts TableOptions) {
	if formation == nil {
		_, _ = fmt.Fprintln(w, "No formation data available.")
		return
	}

	c := opts.Colors
	if c == nil {
		c = NewColors(ColorNever)
	}

	// Platform header
	_, _ = fmt.Fprintf(w, "%s %s\n\n", c.Header("Platform:"), c.Platform(formation.Platform))

	// Render sectors
	if len(formation.Sectors) > 0 {
		renderSectors(w, formation.Sectors, c)
	}

	// Render carriages
	if len(formation.Carriages) > 0 {
		renderCarriages(w, formation, c)
	}

	_, _ = fmt.Fprintln(w)

	// Render groups with details
	for _, group := range formation.Groups {
		renderGroup(w, &group, c)
	}
}

func renderSectors(w io.Writer, sectors []models.Sector, c *Colors) {
	var sb strings.Builder

	for _, sector := range sectors {
		sectorLength := int(sector.LengthPercent)
		if sectorLength < 3 {
			sectorLength = 3
		}

		spacingLeft := (sectorLength - 2) / 2
		spacingRight := (sectorLength - 2) - spacingLeft

		if spacingLeft < 0 {
			spacingLeft = 0
		}
		if spacingRight < 0 {
			spacingRight = 0
		}

		sb.WriteString("▏")
		sb.WriteString(strings.Repeat(" ", spacingLeft))
		sb.WriteString(c.Header(sector.Name))
		sb.WriteString(strings.Repeat(" ", spacingRight))
		sb.WriteString("▕")
	}

	_, _ = fmt.Fprintln(w, sb.String())
}

func renderCarriages(w io.Writer, formation *models.Formation, c *Colors) {
	var sb strings.Builder

	// Find minimum start position for padding
	minStart := 100.0
	for _, carriage := range formation.Carriages {
		if carriage.StartPercent < minStart {
			minStart = carriage.StartPercent
		}
	}

	// Add initial padding
	if minStart > 1 {
		sb.WriteString(strings.Repeat(" ", int(minStart-1)))
	}

	// Direction indicator
	if formation.Direction == 100 {
		sb.WriteString(c.Muted(">"))
	} else {
		sb.WriteString(c.Muted("<"))
	}

	// Render each carriage
	for _, carriage := range formation.Carriages {
		wagonLength := int(carriage.LengthPercent)
		if wagonLength < 5 {
			wagonLength = 5
		}

		spacingLeft := wagonLength/2 - 2
		spacingRight := wagonLength/2 - 1

		if wagonLength%2 == 1 {
			spacingLeft++
		}

		if spacingLeft < 0 {
			spacingLeft = 0
		}
		if spacingRight < 0 {
			spacingRight = 0
		}

		// Determine wagon description
		wagonDesc := carriage.Number
		if wagonDesc == "" {
			wagonDesc = "?"
		}

		if carriage.IsClosed {
			wagonDesc = "X"
		}

		if carriage.IsLocomotive || carriage.IsPowercar {
			wagonDesc = "[]"
		}

		// Truncate to 3 characters
		if len(wagonDesc) > 3 {
			wagonDesc = wagonDesc[:3]
		}

		// Apply class color
		var coloredDesc string
		switch carriage.ClassType {
		case 1:
			coloredDesc = c.DelayHigh("%3s", wagonDesc) // First class in red/bold
		case 2:
			coloredDesc = c.Line("%3s", wagonDesc) // Second class in cyan
		case 12:
			coloredDesc = c.Delay("%3s", wagonDesc) // Mixed class in yellow
		default:
			coloredDesc = fmt.Sprintf("%3s", wagonDesc)
		}

		sb.WriteString(strings.Repeat(" ", spacingLeft))
		sb.WriteString(coloredDesc)
		sb.WriteString(strings.Repeat(" ", spacingRight))
	}

	// Closing direction indicator
	if formation.Direction == 100 {
		sb.WriteString(c.Muted(">"))
	} else {
		sb.WriteString(c.Muted("<"))
	}

	_, _ = fmt.Fprintln(w, sb.String())
}

func renderGroup(w io.Writer, group *models.Group, c *Colors) {
	// Group header
	desc := group.Description
	if desc == "" {
		desc = "Train"
	}

	designation := ""
	if group.Designation != "" {
		designation = fmt.Sprintf(" \"%s\"", group.Designation)
	}

	sectors := ""
	if len(group.Sectors) > 0 {
		sectors = " (" + strings.Join(group.Sectors, "") + ")"
	}

	_, _ = fmt.Fprintf(w, "%s%s%s\n", c.Header(desc), c.Muted(designation), c.Muted(sectors))
	_, _ = fmt.Fprintf(w, "%s %s  %s %s\n\n",
		c.Line(group.TrainType),
		c.Line(group.TrainNo),
		c.Muted("→"),
		group.Destination,
	)

	// Carriage details
	for _, carriage := range group.Carriages {
		number := carriage.Number
		if number == "" {
			number = "?"
		}
		if carriage.IsClosed {
			number = "X"
		}
		if carriage.IsLocomotive {
			number = "Lok"
		}

		model := carriage.Model
		if model == "" {
			model = "???"
		}

		carriageType := carriage.Type
		if carriageType == "" {
			carriageType = "?"
		}

		// Build amenities string
		var amenities []string
		if carriage.IsDosto {
			amenities = append(amenities, "Doppelstock")
		}
		if carriage.HasBistro {
			amenities = append(amenities, "Bistro")
		}
		if carriage.HasQuietZone {
			amenities = append(amenities, "Ruhebereich")
		}
		if carriage.HasFamilyZone {
			amenities = append(amenities, "Familienbereich")
		}

		amenityStr := ""
		if len(amenities) > 0 {
			amenityStr = "  " + strings.Join(amenities, "  ")
		}

		// Class indicator
		classStr := ""
		switch carriage.ClassType {
		case 1:
			classStr = c.DelayHigh("1.")
		case 2:
			classStr = c.Line("2.")
		case 12:
			classStr = c.Delay("1./2.")
		}

		_, _ = fmt.Fprintf(w, "%3s: %3s %10s  %s%s\n",
			number,
			model,
			carriageType,
			classStr,
			c.Muted(amenityStr),
		)
	}

	_, _ = fmt.Fprintln(w)
}
