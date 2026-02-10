package operators

import "testing"

func TestGetOperator(t *testing.T) {
	tests := []struct {
		name     string
		adminID  string
		wantAbbr string
		wantName string
		wantNil  bool
	}{
		{
			name:     "DB Fernverkehr",
			adminID:  "80",
			wantAbbr: "DB",
			wantName: "DB Fernverkehr AG",
		},
		{
			name:     "SBB",
			adminID:  "85",
			wantAbbr: "SBB",
			wantName: "SBB",
		},
		{
			name:     "ÖBB",
			adminID:  "81",
			wantAbbr: "ÖBB",
			wantName: "Österreichische Bundesbahnen",
		},
		{
			name:     "FlixTrain",
			adminID:  "FLX10",
			wantAbbr: "FLX",
			wantName: "FlixTrain",
		},
		{
			name:    "unknown operator",
			adminID: "UNKNOWN_ID",
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			op := GetOperator(tt.adminID)

			if tt.wantNil {
				if op != nil {
					t.Errorf("GetOperator() = %v, want nil", op)
				}
				return
			}

			if op == nil {
				t.Fatalf("GetOperator() returned nil, want operator")
			}
			if op.Abbr != tt.wantAbbr {
				t.Errorf("Abbr = %q, want %q", op.Abbr, tt.wantAbbr)
			}
			if op.Name != tt.wantName {
				t.Errorf("Name = %q, want %q", op.Name, tt.wantName)
			}
		})
	}
}

func TestGetOperatorAbbr(t *testing.T) {
	tests := []struct {
		adminID string
		want    string
	}{
		{"80", "DB"},
		{"85", "SBB"},
		{"81", "ÖBB"},
		{"UNKNOWN", ""},
	}

	for _, tt := range tests {
		t.Run(tt.adminID, func(t *testing.T) {
			if got := GetOperatorAbbr(tt.adminID); got != tt.want {
				t.Errorf("GetOperatorAbbr(%q) = %q, want %q", tt.adminID, got, tt.want)
			}
		})
	}
}

func TestGetOperatorName(t *testing.T) {
	tests := []struct {
		adminID string
		want    string
	}{
		{"80", "DB Fernverkehr AG"},
		{"85", "SBB"},
		{"81", "Österreichische Bundesbahnen"},
		{"UNKNOWN", ""},
	}

	for _, tt := range tests {
		t.Run(tt.adminID, func(t *testing.T) {
			if got := GetOperatorName(tt.adminID); got != tt.want {
				t.Errorf("GetOperatorName(%q) = %q, want %q", tt.adminID, got, tt.want)
			}
		})
	}
}

func TestOperatorMapContainsExpectedEntries(t *testing.T) {
	// Verify that common operators are present
	expectedOperators := []string{
		"80", // DB
		"81", // ÖBB
		"85", // SBB
		"87", // SNCF
		"84", // NS
	}

	for _, id := range expectedOperators {
		if GetOperator(id) == nil {
			t.Errorf("Expected operator with ID %q not found", id)
		}
	}
}
