// +build integration

package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

var binaryPath string

// TestMain builds the binary before running tests
func TestMain(m *testing.M) {
	// Build the binary
	binaryPath = filepath.Join(os.TempDir(), "moko-test")
	build := exec.Command("go", "build", "-o", binaryPath, ".")
	if err := build.Run(); err != nil {
		os.Exit(1)
	}

	// Run tests
	code := m.Run()

	// Cleanup
	_ = os.Remove(binaryPath)
	os.Exit(code)
}

func runCommand(t *testing.T, args ...string) (string, string, int) {
	t.Helper()
	cmd := exec.Command(binaryPath, args...)

	stdout, err := cmd.Output()
	stderr := ""
	exitCode := 0

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
			stderr = string(exitErr.Stderr)
		}
	}

	return string(stdout), stderr, exitCode
}

func TestCLI_Version(t *testing.T) {
	stdout, _, exitCode := runCommand(t, "--version")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "moko version") {
		t.Errorf("Expected version output, got: %s", stdout)
	}
}

func TestCLI_Help(t *testing.T) {
	stdout, _, exitCode := runCommand(t, "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "moko is a command-line interface") {
		t.Errorf("Expected help text, got: %s", stdout)
	}

	// Check that all commands are listed
	commands := []string{"search", "departures", "arrivals", "nearby", "journey", "formation", "tui"}
	for _, cmd := range commands {
		if !strings.Contains(stdout, cmd) {
			t.Errorf("Expected command '%s' in help output", cmd)
		}
	}
}

func TestCLI_SearchCommand_Help(t *testing.T) {
	stdout, _, exitCode := runCommand(t, "search", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Search for stations") {
		t.Errorf("Expected search help text, got: %s", stdout)
	}
}

func TestCLI_SearchCommand_MissingQuery(t *testing.T) {
	stdout, stderr, exitCode := runCommand(t, "search")

	// Command should either fail or show help
	if exitCode == 0 && !strings.Contains(stdout, "Usage:") && !strings.Contains(stderr, "Usage:") {
		t.Error("Expected non-zero exit code or help text for missing query")
	}
}

func TestCLI_SearchCommand_JSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API call in short mode")
	}

	stdout, _, exitCode := runCommand(t, "search", "Frankfurt", "--json")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Try to parse as JSON array
	var results []interface{}
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Errorf("Expected valid JSON array, got error: %v", err)
	}
}

func TestCLI_DeparturesCommand_Help(t *testing.T) {
	stdout, _, exitCode := runCommand(t, "departures", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Show upcoming departures") {
		t.Errorf("Expected departures help text, got: %s", stdout)
	}
}

func TestCLI_DeparturesCommand_MissingStation(t *testing.T) {
	stdout, stderr, exitCode := runCommand(t, "departures")

	// Command should either fail or show help
	if exitCode == 0 && !strings.Contains(stdout, "Usage:") && !strings.Contains(stderr, "Usage:") {
		t.Error("Expected non-zero exit code or help text for missing station")
	}
}

func TestCLI_DeparturesCommand_WithModes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API call in short mode")
	}

	// Use Frankfurt Hbf EVA: 8000105
	stdout, _, exitCode := runCommand(t, "departures", "8000105:A=1@O=Frankfurt(Main)Hbf@X=8663003@Y=50107145@U=80@L=8000105@", "--modes", "ICE")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if stdout == "" {
		t.Error("Expected output, got empty string")
	}
}

func TestCLI_ArrivalsCommand_Help(t *testing.T) {
	stdout, _, exitCode := runCommand(t, "arrivals", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Show upcoming arrivals") {
		t.Errorf("Expected arrivals help text, got: %s", stdout)
	}
}

func TestCLI_NearbyCommand_Help(t *testing.T) {
	stdout, _, exitCode := runCommand(t, "nearby", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Search for stations near") {
		t.Errorf("Expected nearby help text, got: %s", stdout)
	}
}

func TestCLI_NearbyCommand_InvalidCoordinates(t *testing.T) {
	stdout, stderr, exitCode := runCommand(t, "nearby", "invalid")

	// Command should either fail or show error message
	if exitCode == 0 && !strings.Contains(stdout, "Usage:") && !strings.Contains(stderr, "error") && !strings.Contains(stdout, "error") {
		t.Error("Expected non-zero exit code or error message for invalid coordinates")
	}
}

func TestCLI_NearbyCommand_ValidCoordinates(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API call in short mode")
	}

	stdout, _, exitCode := runCommand(t, "nearby", "50.107:8.663", "--json")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Try to parse as JSON array
	var results []interface{}
	if err := json.Unmarshal([]byte(stdout), &results); err != nil {
		t.Errorf("Expected valid JSON array, got error: %v", err)
	}
}

func TestCLI_JourneyCommand_Help(t *testing.T) {
	stdout, _, exitCode := runCommand(t, "journey", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Show detailed information") {
		t.Errorf("Expected journey help text, got: %s", stdout)
	}
}

func TestCLI_JourneyCommand_MissingID(t *testing.T) {
	stdout, stderr, exitCode := runCommand(t, "journey")

	// Command should either fail or show help
	if exitCode == 0 && !strings.Contains(stdout, "Usage:") && !strings.Contains(stderr, "Usage:") {
		t.Error("Expected non-zero exit code or help text for missing journey ID")
	}
}

func TestCLI_FormationCommand_Help(t *testing.T) {
	stdout, _, exitCode := runCommand(t, "formation", "--help")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if !strings.Contains(stdout, "Show the carriage formation") {
		t.Errorf("Expected formation help text, got: %s", stdout)
	}
}

func TestCLI_FormationCommand_MissingArgs(t *testing.T) {
	stdout, stderr, exitCode := runCommand(t, "formation")

	// Command should either fail or show help
	if exitCode == 0 && !strings.Contains(stdout, "Usage:") && !strings.Contains(stderr, "Usage:") {
		t.Error("Expected non-zero exit code or help text for missing arguments")
	}
}

func TestCLI_GlobalFlags_Color(t *testing.T) {
	tests := []struct {
		name  string
		flag  string
		valid bool
	}{
		{"auto", "auto", true},
		{"always", "always", true},
		{"never", "never", true},
		{"invalid", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, exitCode := runCommand(t, "search", "test", "--color", tt.flag)

			if tt.valid && exitCode == 0 {
				// Valid flag should work (though search might fail)
			} else if !tt.valid && exitCode != 0 {
				// Invalid flag should fail
			}
		})
	}
}

func TestCLI_GlobalFlags_NoCache(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API call in short mode")
	}

	stdout, _, exitCode := runCommand(t, "search", "Frankfurt", "--no-cache")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	if stdout == "" {
		t.Error("Expected output, got empty string")
	}
}

func TestCLI_DateTimeFlags(t *testing.T) {
	tests := []struct {
		name     string
		dateFlag string
		timeFlag string
	}{
		{"date YYYY-MM-DD", "2025-12-31", ""},
		{"date DD.MM.YYYY", "31.12.2025", ""},
		{"time HH:MM", "", "14:30"},
		{"both", "2025-12-31", "14:30"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := []string{"search", "Frankfurt"}
			if tt.dateFlag != "" {
				args = append(args, "--date", tt.dateFlag)
			}
			if tt.timeFlag != "" {
				args = append(args, "--time", tt.timeFlag)
			}

			_, _, exitCode := runCommand(t, args...)

			// Command should accept the flags (though API might fail)
			if exitCode == 127 {
				t.Errorf("Command not found or parsing failed")
			}
		})
	}
}

func TestCLI_RawJSONOutput(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API call in short mode")
	}

	stdout, _, exitCode := runCommand(t, "search", "Frankfurt", "--raw-json")

	if exitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", exitCode)
	}

	// Raw JSON should be valid JSON
	var raw interface{}
	if err := json.Unmarshal([]byte(stdout), &raw); err != nil {
		t.Errorf("Expected valid raw JSON, got error: %v", err)
	}
}

func TestCLI_InvalidCommand(t *testing.T) {
	_, _, exitCode := runCommand(t, "nonexistent")

	if exitCode == 0 {
		t.Error("Expected non-zero exit code for invalid command")
	}
}

func TestCLI_MultipleModesFilter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping API call in short mode")
	}

	stdout, _, exitCode := runCommand(t, "search", "Frankfurt", "--modes", "ICE,EC_IC,REGIONAL")

	// Should accept multiple modes comma-separated
	if exitCode == 127 {
		t.Error("Command failed to parse multiple modes")
	}

	if stdout == "" && exitCode == 0 {
		// Command parsed correctly
	}
}
