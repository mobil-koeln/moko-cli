package output

import (
	"strings"
	"testing"

	"github.com/fatih/color"
	"github.com/mobil-koeln/moko-cli/internal/testutil"
)

func TestParseColorMode(t *testing.T) {
	tests := []struct {
		input string
		want  ColorMode
	}{
		{"always", ColorAlways},
		{"never", ColorNever},
		{"auto", ColorAuto},
		{"", ColorAuto},        // default
		{"invalid", ColorAuto}, // default
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseColorMode(tt.input)
			testutil.AssertEqual(t, got, tt.want)
		})
	}
}

func TestNewColors_NeverMode(t *testing.T) {
	// Save and restore color state
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = true

	c := NewColors(ColorNever)

	// Test that all color functions return uncolored strings
	testutil.AssertEqual(t, c.Time("15:04"), "15:04")
	testutil.AssertEqual(t, c.Delay("+2"), "+2")
	testutil.AssertEqual(t, c.DelayHigh("+12"), "+12")
	testutil.AssertEqual(t, c.OnTime("-1"), "-1")
	testutil.AssertEqual(t, c.Line("ICE 123"), "ICE 123")
	testutil.AssertEqual(t, c.Platform("Pl.7"), "Pl.7")
	testutil.AssertEqual(t, c.Dest("München"), "München")
	testutil.AssertEqual(t, c.Canceled("CANCELED"), "CANCELED")
	testutil.AssertEqual(t, c.Via("via Frankfurt"), "via Frankfurt")
	testutil.AssertEqual(t, c.Header("Departures"), "Departures")
	testutil.AssertEqual(t, c.Muted("details"), "details")
}

func TestNewColors_AlwaysMode(t *testing.T) {
	c := NewColors(ColorAlways)

	// Test that color functions return ANSI-escaped strings
	// We check for ANSI escape sequences (starting with \033[)
	result := c.Time("15:04")
	testutil.AssertContains(t, result, "\033[")
	testutil.AssertContains(t, result, "15:04")

	result = c.DelayHigh("+12")
	testutil.AssertContains(t, result, "\033[")
	testutil.AssertContains(t, result, "+12")

	result = c.Line("ICE 123")
	testutil.AssertContains(t, result, "\033[")
	testutil.AssertContains(t, result, "ICE 123")
}

func TestFormatDelay_NoColor(t *testing.T) {
	// Save and restore color state
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = true

	c := NewColors(ColorNever)

	tests := []struct {
		name  string
		delay int
		want  string
	}{
		{"zero delay", 0, "    "},      // 4 spaces
		{"minor delay", 5, "  +5"},     // right-aligned +5
		{"major delay", 12, " +12"},    // right-aligned +12
		{"negative delay", -3, "  -3"}, // early arrival
		{"single digit", 1, "  +1"},    // single digit delay
		{"large delay", 123, "+123"},   // large delay
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.FormatDelay(tt.delay)
			testutil.AssertEqual(t, got, tt.want)
		})
	}
}

func TestFormatDelay_WithColor(t *testing.T) {
	c := NewColors(ColorAlways)

	tests := []struct {
		name          string
		delay         int
		shouldHave    string
		shouldNotHave string
	}{
		{"zero delay", 0, "    ", "\033["}, // no color codes
		{"minor delay", 5, "\033[", ""},    // yellow color
		{"major delay", 12, "\033[", ""},   // red color
		{"negative", -3, "\033[", ""},      // green color (on-time)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := c.FormatDelay(tt.delay)

			if tt.shouldHave != "" {
				testutil.AssertContains(t, got, tt.shouldHave)
			}

			if tt.shouldNotHave != "" {
				testutil.AssertNotContains(t, got, tt.shouldNotHave)
			}

			// Verify delay value is present (strip ANSI codes for comparison)
			stripped := stripANSI(got)
			if tt.delay == 0 {
				testutil.AssertEqual(t, stripped, "    ")
			} else {
				testutil.AssertContains(t, stripped, formatDelayValue(tt.delay))
			}
		})
	}
}

func TestFormatDelay_Width(t *testing.T) {
	// Save and restore color state
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = true

	c := NewColors(ColorNever)

	// All formatted delays should be exactly 4 characters wide (without ANSI codes)
	tests := []int{0, 1, 5, 9, 10, 15, 99, 100, 999}

	for _, delay := range tests {
		t.Run(formatDelayValue(delay), func(t *testing.T) {
			got := c.FormatDelay(delay)
			testutil.AssertEqual(t, len(got), 4)
		})
	}
}

func TestColors_Sprintf(t *testing.T) {
	// Save and restore color state
	oldNoColor := color.NoColor
	defer func() { color.NoColor = oldNoColor }()
	color.NoColor = true

	c := NewColors(ColorNever)

	// Test sprintf formatting
	testutil.AssertEqual(t, c.Time("%02d:%02d", 14, 30), "14:30")
	testutil.AssertEqual(t, c.Line("ICE %d", 123), "ICE 123")
	testutil.AssertEqual(t, c.Platform("Pl.%s", "7"), "Pl.7")
}

// Helper functions

func stripANSI(s string) string {
	// Simple ANSI stripper for testing
	var result strings.Builder
	inEscape := false

	for _, r := range s {
		if r == '\033' {
			inEscape = true
			continue
		}
		if inEscape {
			if r == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteRune(r)
	}

	return result.String()
}

func formatDelayValue(delay int) string {
	if delay > 0 {
		return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(
			strings.TrimSpace(formatInt(delay, true)), " ", ""), "\t", ""))
	}
	return strings.TrimSpace(formatInt(delay, false))
}

func formatInt(val int, positive bool) string {
	if positive && val > 0 {
		return "+" + strings.TrimSpace(formatInt(val, false))
	}
	if val == 0 {
		return ""
	}
	s := ""
	if val < 0 {
		s = "-"
		val = -val
	}
	// Simple int to string
	digits := []rune{}
	for val > 0 {
		digits = append([]rune{rune('0' + val%10)}, digits...)
		val /= 10
	}
	return s + string(digits)
}
