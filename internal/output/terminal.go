package output

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
)

// ClearScreen clears the terminal screen and moves cursor to top-left
func ClearScreen(w io.Writer) {
	_, _ = fmt.Fprint(w, "\033[2J\033[H")
}

// HideCursor hides the terminal cursor
func HideCursor(w io.Writer) {
	_, _ = fmt.Fprint(w, "\033[?25l")
}

// ShowCursor shows the terminal cursor
func ShowCursor(w io.Writer) {
	_, _ = fmt.Fprint(w, "\033[?25h")
}

// SetupSignalHandler returns a channel that receives interrupt signals
func SetupSignalHandler() chan os.Signal {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	return sigChan
}
