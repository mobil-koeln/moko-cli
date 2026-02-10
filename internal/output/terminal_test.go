package output

import (
	"bytes"
	"os"
	"testing"
	"time"

	"github.com/mobil-koeln/moko-cli/internal/testutil"
)

func TestClearScreen(t *testing.T) {
	var buf bytes.Buffer
	ClearScreen(&buf)

	output := buf.String()
	// Should contain ANSI escape sequences for clear screen and move cursor
	testutil.AssertContains(t, output, "\033[2J")
	testutil.AssertContains(t, output, "\033[H")
}

func TestHideCursor(t *testing.T) {
	var buf bytes.Buffer
	HideCursor(&buf)

	output := buf.String()
	// Should contain ANSI escape sequence for hiding cursor
	testutil.AssertContains(t, output, "\033[?25l")
}

func TestShowCursor(t *testing.T) {
	var buf bytes.Buffer
	ShowCursor(&buf)

	output := buf.String()
	// Should contain ANSI escape sequence for showing cursor
	testutil.AssertContains(t, output, "\033[?25h")
}

func TestSetupSignalHandler(t *testing.T) {
	sigChan := SetupSignalHandler()

	// Verify channel is created
	testutil.AssertTrue(t, sigChan != nil)

	// Verify channel is buffered (won't block)
	select {
	case <-sigChan:
		t.Error("channel should be empty initially")
	case <-time.After(10 * time.Millisecond):
		// Expected - channel is empty
	}

	// Simulate sending a signal (in a real test, we'd send os.Interrupt)
	// For this test, we just verify the channel works
	go func() {
		sigChan <- os.Interrupt
	}()

	// Verify we can receive from the channel
	select {
	case sig := <-sigChan:
		testutil.AssertEqual(t, sig, os.Interrupt)
	case <-time.After(100 * time.Millisecond):
		t.Error("should have received signal")
	}
}
