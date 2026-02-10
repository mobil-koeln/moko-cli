package output

import (
	"os"

	"github.com/fatih/color"
	"github.com/mattn/go-isatty"
)

// ColorMode represents the color output mode
type ColorMode int

const (
	// ColorAuto enables colors if output is a TTY
	ColorAuto ColorMode = iota
	// ColorAlways forces colors on
	ColorAlways
	// ColorNever disables colors
	ColorNever
)

// Colors holds the color functions for different output types
type Colors struct {
	Time      func(format string, a ...interface{}) string
	Delay     func(format string, a ...interface{}) string
	DelayHigh func(format string, a ...interface{}) string
	OnTime    func(format string, a ...interface{}) string
	Line      func(format string, a ...interface{}) string
	Platform  func(format string, a ...interface{}) string
	Dest      func(format string, a ...interface{}) string
	Canceled  func(format string, a ...interface{}) string
	Via       func(format string, a ...interface{}) string
	Header    func(format string, a ...interface{}) string
	Muted     func(format string, a ...interface{}) string
}

// NewColors creates a new Colors instance based on the color mode
func NewColors(mode ColorMode) *Colors {
	// Determine if we should use colors
	useColors := false
	switch mode {
	case ColorAlways:
		useColors = true
		color.NoColor = false // Force colors on
	case ColorNever:
		useColors = false
	case ColorAuto:
		useColors = isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
	}

	if !useColors {
		// Return no-op color functions
		noColor := func(format string, a ...interface{}) string {
			if len(a) == 0 {
				return format
			}
			return color.New().Sprintf(format, a...)
		}
		return &Colors{
			Time:      noColor,
			Delay:     noColor,
			DelayHigh: noColor,
			OnTime:    noColor,
			Line:      noColor,
			Platform:  noColor,
			Dest:      noColor,
			Canceled:  noColor,
			Via:       noColor,
			Header:    noColor,
			Muted:     noColor,
		}
	}

	// Create colored functions
	return &Colors{
		Time:      color.New(color.FgWhite, color.Bold).SprintfFunc(),
		Delay:     color.New(color.FgYellow).SprintfFunc(),
		DelayHigh: color.New(color.FgRed, color.Bold).SprintfFunc(),
		OnTime:    color.New(color.FgGreen).SprintfFunc(),
		Line:      color.New(color.FgCyan, color.Bold).SprintfFunc(),
		Platform:  color.New(color.FgMagenta).SprintfFunc(),
		Dest:      color.New(color.FgWhite).SprintfFunc(),
		Canceled:  color.New(color.FgRed, color.Bold).SprintfFunc(),
		Via:       color.New(color.FgHiBlack).SprintfFunc(),
		Header:    color.New(color.FgWhite, color.Bold).SprintfFunc(),
		Muted:     color.New(color.FgHiBlack).SprintfFunc(),
	}
}

// FormatDelay formats a delay value with appropriate color (fixed 4-char width)
func (c *Colors) FormatDelay(delay int) string {
	if delay == 0 {
		return "    " // 4 spaces for alignment
	}
	if delay > 0 {
		if delay >= 10 {
			return c.DelayHigh("%+4d", delay)
		}
		return c.Delay("%+4d", delay)
	}
	return c.OnTime("%4d", delay)
}

// ParseColorMode parses a color mode string
func ParseColorMode(s string) ColorMode {
	switch s {
	case "always":
		return ColorAlways
	case "never":
		return ColorNever
	default:
		return ColorAuto
	}
}
