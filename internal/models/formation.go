package models

import (
	"fmt"
	"sort"
)

// Formation represents a train carriage formation/composition
type Formation struct {
	Platform     string     `json:"platform"`
	Direction    int        `json:"direction"` // 0 or 100
	TrainType    string     `json:"trainType"`
	Sectors      []Sector   `json:"sectors"`
	Carriages    []Carriage `json:"carriages"`
	Groups       []Group    `json:"groups"`
	Destinations []string   `json:"destinations,omitempty"`
	TrainNumbers []string   `json:"trainNumbers,omitempty"`
}

// Sector represents a platform sector/zone
type Sector struct {
	Name          string  `json:"name"`
	StartPercent  float64 `json:"startPercent"`
	EndPercent    float64 `json:"endPercent"`
	LengthPercent float64 `json:"lengthPercent"`
	StartMeters   float64 `json:"startMeters"`
	EndMeters     float64 `json:"endMeters"`
	LengthMeters  float64 `json:"lengthMeters"`
}

// Carriage represents a single wagon/car in the formation
type Carriage struct {
	Number        string  `json:"number"`
	Model         string  `json:"model,omitempty"`
	Type          string  `json:"type"`
	UicID         string  `json:"uicId,omitempty"`
	Section       string  `json:"section,omitempty"`
	ClassType     int     `json:"classType"` // 0=unknown, 1=first, 2=second, 12=mixed
	StartPercent  float64 `json:"startPercent"`
	EndPercent    float64 `json:"endPercent"`
	LengthPercent float64 `json:"lengthPercent"`
	StartMeters   float64 `json:"startMeters"`
	EndMeters     float64 `json:"endMeters"`
	LengthMeters  float64 `json:"lengthMeters"`

	// Status flags
	IsClosed     bool `json:"isClosed"`
	IsLocomotive bool `json:"isLocomotive"`
	IsPowercar   bool `json:"isPowercar"`
	IsDosto      bool `json:"isDosto"` // Double-decker

	// Amenities
	HasFirstClass      bool `json:"hasFirstClass"`
	HasSecondClass     bool `json:"hasSecondClass"`
	HasBistro          bool `json:"hasBistro"`
	HasAC              bool `json:"hasAc"`
	HasWheelchairSpace bool `json:"hasWheelchairSpace"`
	HasFamilyZone      bool `json:"hasFamilyZone"`
	HasQuietZone       bool `json:"hasQuietZone"`
	HasBahnComfort     bool `json:"hasBahnComfort"`
}

// Group represents a group of carriages (often a train unit)
type Group struct {
	Name         string     `json:"name"`
	Designation  string     `json:"designation,omitempty"` // ICE name like "Gießen"
	TrainType    string     `json:"trainType"`
	TrainNo      string     `json:"trainNo"`
	Destination  string     `json:"destination"`
	Description  string     `json:"description,omitempty"`
	Model        string     `json:"model,omitempty"`
	Series       string     `json:"series,omitempty"`
	Sectors      []string   `json:"sectors,omitempty"`
	Carriages    []Carriage `json:"carriages"`
	StartPercent float64    `json:"startPercent"`
	EndPercent   float64    `json:"endPercent"`
}

// FormationResponse represents the raw API response for formation
type FormationResponse struct {
	DeparturePlatform         string `json:"departurePlatform"`
	DeparturePlatformSchedule string `json:"departurePlatformSchedule"`
	Platform                  struct {
		Start   float64 `json:"start"`
		End     float64 `json:"end"`
		Sectors []struct {
			Name         string  `json:"name"`
			Start        float64 `json:"start"`
			End          float64 `json:"end"`
			CubePosition float64 `json:"cubePosition"`
		} `json:"sectors"`
	} `json:"platform"`
	Groups []struct {
		Name      string `json:"name"`
		Transport struct {
			Category    string      `json:"category"`
			Number      interface{} `json:"number"` // Can be string or int
			Destination struct {
				Name string `json:"name"`
			} `json:"destination"`
		} `json:"transport"`
		Vehicles []struct {
			WagonIdentificationNumber interface{} `json:"wagonIdentificationNumber"` // Can be string or int
			VehicleID                 string      `json:"vehicleID"`
			Status                    string      `json:"status"`
			Type                      struct {
				ConstructionType string `json:"constructionType"`
				Category         string `json:"category"`
				HasFirstClass    bool   `json:"hasFirstClass"`
				HasEconomyClass  bool   `json:"hasEconomyClass"`
			} `json:"type"`
			PlatformPosition struct {
				Start  float64 `json:"start"`
				End    float64 `json:"end"`
				Sector string  `json:"sector"`
			} `json:"platformPosition"`
			Amenities []struct {
				Type string `json:"type"`
			} `json:"amenities"`
		} `json:"vehicles"`
	} `json:"groups"`
}

// ToFormation converts the raw response to a Formation
func (r *FormationResponse) ToFormation(trainType string) *Formation {
	platformLength := r.Platform.End - r.Platform.Start
	if platformLength == 0 {
		platformLength = 1 // Avoid division by zero
	}

	f := &Formation{
		Platform:  r.DeparturePlatform,
		TrainType: trainType,
	}

	// Parse sectors
	for _, s := range r.Platform.Sectors {
		sector := Sector{
			Name:         s.Name,
			StartMeters:  s.Start,
			EndMeters:    s.End,
			LengthMeters: s.End - s.Start,
			StartPercent: (s.Start - r.Platform.Start) * 100 / platformLength,
			EndPercent:   (s.End - r.Platform.Start) * 100 / platformLength,
		}
		sector.LengthPercent = sector.EndPercent - sector.StartPercent
		f.Sectors = append(f.Sectors, sector)
	}

	// Parse groups and carriages
	destSet := make(map[string]bool)
	trainNoSet := make(map[string]bool)

	for _, g := range r.Groups {
		// Handle number which can be string or int
		trainNo := ""
		switch v := g.Transport.Number.(type) {
		case string:
			trainNo = v
		case float64:
			trainNo = fmt.Sprintf("%.0f", v)
		case int:
			trainNo = fmt.Sprintf("%d", v)
		}

		group := Group{
			Name:        g.Name,
			TrainType:   g.Transport.Category,
			TrainNo:     trainNo,
			Destination: g.Transport.Destination.Name,
		}

		destSet[group.Destination] = true
		trainNoSet[group.TrainNo] = true

		// Check for ICE designation
		group.Designation = getICEDesignation(g.Name)

		sectorSet := make(map[string]bool)

		for _, v := range g.Vehicles {
			carriage := parseCarriage(v, r.Platform.Start, platformLength)
			group.Carriages = append(group.Carriages, carriage)
			f.Carriages = append(f.Carriages, carriage)

			if carriage.Section != "" {
				sectorSet[carriage.Section] = true
			}
		}

		// Sort carriages by position
		sort.Slice(group.Carriages, func(i, j int) bool {
			return group.Carriages[i].StartPercent < group.Carriages[j].StartPercent
		})

		// Set group position
		if len(group.Carriages) > 0 {
			group.StartPercent = group.Carriages[0].StartPercent
			group.EndPercent = group.Carriages[len(group.Carriages)-1].EndPercent
		}

		// Convert sector set to slice
		for s := range sectorSet {
			group.Sectors = append(group.Sectors, s)
		}
		sort.Strings(group.Sectors)

		// Parse model/description
		group.parseDescription()

		f.Groups = append(f.Groups, group)
	}

	// Sort groups by position
	sort.Slice(f.Groups, func(i, j int) bool {
		return f.Groups[i].StartPercent < f.Groups[j].StartPercent
	})

	// Sort all carriages by position
	sort.Slice(f.Carriages, func(i, j int) bool {
		return f.Carriages[i].StartPercent < f.Carriages[j].StartPercent
	})

	// Determine direction
	if len(f.Carriages) > 1 {
		if f.Carriages[0].StartPercent > f.Carriages[len(f.Carriages)-1].StartPercent {
			f.Direction = 100
		}
	}

	// Collect destinations and train numbers
	for d := range destSet {
		f.Destinations = append(f.Destinations, d)
	}
	for t := range trainNoSet {
		f.TrainNumbers = append(f.TrainNumbers, t)
	}

	return f
}

func parseCarriage(v struct {
	WagonIdentificationNumber interface{} `json:"wagonIdentificationNumber"`
	VehicleID                 string      `json:"vehicleID"`
	Status                    string      `json:"status"`
	Type                      struct {
		ConstructionType string `json:"constructionType"`
		Category         string `json:"category"`
		HasFirstClass    bool   `json:"hasFirstClass"`
		HasEconomyClass  bool   `json:"hasEconomyClass"`
	} `json:"type"`
	PlatformPosition struct {
		Start  float64 `json:"start"`
		End    float64 `json:"end"`
		Sector string  `json:"sector"`
	} `json:"platformPosition"`
	Amenities []struct {
		Type string `json:"type"`
	} `json:"amenities"`
}, platformStart, platformLength float64) Carriage {
	// Convert wagon number to string
	wagonNumber := ""
	switch n := v.WagonIdentificationNumber.(type) {
	case string:
		wagonNumber = n
	case float64:
		wagonNumber = fmt.Sprintf("%.0f", n)
	case int:
		wagonNumber = fmt.Sprintf("%d", n)
	}

	c := Carriage{
		Number:         wagonNumber,
		UicID:          v.VehicleID,
		Type:           v.Type.ConstructionType,
		Section:        v.PlatformPosition.Sector,
		HasFirstClass:  v.Type.HasFirstClass,
		HasSecondClass: v.Type.HasEconomyClass,
		StartMeters:    v.PlatformPosition.Start,
		EndMeters:      v.PlatformPosition.End,
		IsClosed:       v.Status == "CLOSED",
	}

	// Calculate percentage positions
	if platformLength > 0 {
		c.StartPercent = (v.PlatformPosition.Start - platformStart) * 100 / platformLength
		c.EndPercent = (v.PlatformPosition.End - platformStart) * 100 / platformLength
		c.LengthPercent = c.EndPercent - c.StartPercent
	}

	c.LengthMeters = v.PlatformPosition.End - v.PlatformPosition.Start

	// Parse model from UIC ID (format: extract middle 3 digits)
	if len(v.VehicleID) >= 12 {
		c.Model = v.VehicleID[5:8]
	}

	// Determine class type from construction type
	if len(c.Type) > 0 {
		switch c.Type[0] {
		case 'D':
			c.IsDosto = true
		}
		if containsAny(c.Type, "AB") {
			c.ClassType = 12
		} else if containsAny(c.Type, "A") {
			c.ClassType = 1
		} else if containsAny(c.Type, "B", "WR") {
			c.ClassType = 2
		}
	}

	// Check category
	switch v.Type.Category {
	case "LOCOMOTIVE":
		c.IsLocomotive = true
	case "POWERCAR":
		c.IsPowercar = true
	}
	if containsStr(v.Type.Category, "DININGCAR") {
		c.HasBistro = true
	}

	// Parse amenities
	for _, a := range v.Amenities {
		switch a.Type {
		case "AIR_CONDITION":
			c.HasAC = true
		case "WHEELCHAIR_SPACE":
			c.HasWheelchairSpace = true
		case "ZONE_FAMILY":
			c.HasFamilyZone = true
		case "ZONE_QUIET":
			c.HasQuietZone = true
		case "SEATS_BAHN_COMFORT":
			c.HasBahnComfort = true
		}
	}

	return c
}

func (g *Group) parseDescription() {
	// Simplified model detection based on train numbers
	// A full implementation would analyze UIC IDs like the Perl version
	if len(g.Carriages) == 0 {
		return
	}

	// Check for common models based on model codes
	modelCounts := make(map[string]int)
	for _, c := range g.Carriages {
		if c.Model != "" {
			modelCounts[c.Model]++
		}
	}

	// Find most common model
	var maxModel string
	var maxCount int
	for m, count := range modelCounts {
		if count > maxCount {
			maxModel = m
			maxCount = count
		}
	}

	// Map model codes to descriptions
	modelMap := map[string][2]string{
		"401": {"ICE 1", "BR 401"},
		"402": {"ICE 2", "BR 402"},
		"403": {"ICE 3", "BR 403"},
		"406": {"ICE 3", "BR 406"},
		"407": {"ICE 3 Velaro", "BR 407"},
		"408": {"ICE 3neo", "BR 408"},
		"411": {"ICE T", "BR 411"},
		"412": {"ICE 4", "BR 412"},
		"415": {"ICE T", "BR 415"},
	}

	if desc, ok := modelMap[maxModel]; ok {
		g.Model = desc[0]
		g.Series = desc[1]
		g.Description = desc[0]
		if desc[0] != desc[1] {
			g.Description += " (" + desc[1] + ")"
		}
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		for _, c := range sub {
			for _, sc := range s {
				if c == sc {
					return true
				}
			}
		}
	}
	return false
}

func containsStr(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > len(substr) && (s[:len(substr)] == substr || containsStr(s[1:], substr))))
}

// ICE train designations
var iceDesignations = map[string]string{
	"101": "Gießen", "102": "Jever", "103": "Neu-Isenburg", "104": "Fulda",
	"105": "Offenbach am Main", "110": "Gelsenkirchen", "111": "Nürnberg",
	"301": "Freiburg im Breisgau", "302": "Hansestadt Lübeck", "303": "Dortmund",
	"304": "München", "305": "Baden-Baden", "1101": "Neustadt an der Weinstraße",
	"1104": "Erfurt", "1105": "Dresden", "1108": "Berlin", "1112": "Hamburg",
	"1126": "Leipzig", "1129": "Kiel", "1182": "Mainz",
}

func getICEDesignation(name string) string {
	// Extract number from ICE name like "ICE 301"
	if len(name) < 4 {
		return ""
	}
	// Try to find a number in the name
	var numStr string
	for _, c := range name {
		if c >= '0' && c <= '9' {
			numStr += string(c)
		}
	}
	if numStr == "" {
		return ""
	}
	return iceDesignations[numStr]
}
