package testutil

import (
	"testing"
	"time"
)

// AssertEqual checks if two values are equal
func AssertEqual[T comparable](t *testing.T, got, want T) {
	t.Helper()
	if got != want {
		t.Errorf("got %v, want %v", got, want)
	}
}

// AssertNil checks if error is nil
func AssertNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// AssertError checks if error is not nil
func AssertError(t *testing.T, err error) {
	t.Helper()
	if err == nil {
		t.Fatal("expected error but got nil")
	}
}

// AssertContains checks if string contains substring
func AssertContains(t *testing.T, got, want string) {
	t.Helper()
	if !contains(got, want) {
		t.Errorf("got %q, want it to contain %q", got, want)
	}
}

// AssertNotContains checks if string does not contain substring
func AssertNotContains(t *testing.T, got, notWant string) {
	t.Helper()
	if contains(got, notWant) {
		t.Errorf("got %q, want it to not contain %q", got, notWant)
	}
}

// AssertTimeEqual checks if two times are equal within tolerance
func AssertTimeEqual(t *testing.T, got, want time.Time, tolerance time.Duration) {
	t.Helper()
	diff := got.Sub(want)
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("time difference %v exceeds tolerance %v", diff, tolerance)
	}
}

// AssertFloatEqual checks if two floats are equal within tolerance
func AssertFloatEqual(t *testing.T, got, want, tolerance float64) {
	t.Helper()
	diff := got - want
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Errorf("got %v, want %v (tolerance: %v)", got, want, tolerance)
	}
}

// AssertTrue checks if condition is true
func AssertTrue(t *testing.T, condition bool) {
	t.Helper()
	if !condition {
		t.Error("expected true but got false")
	}
}

// AssertFalse checks if condition is false
func AssertFalse(t *testing.T, condition bool) {
	t.Helper()
	if condition {
		t.Error("expected false but got true")
	}
}

// AssertLen checks if slice/map has expected length
func AssertLen[T any](t *testing.T, items []T, want int) {
	t.Helper()
	got := len(items)
	if got != want {
		t.Errorf("got length %d, want %d", got, want)
	}
}

// contains checks if s contains substr
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || indexInString(s, substr) >= 0)
}

// indexInString finds substr in s
func indexInString(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
