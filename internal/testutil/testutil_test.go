package testutil

import (
	"errors"
	"testing"
	"time"
)

func TestAssertEqual(t *testing.T) {
	// Should pass
	AssertEqual(t, 42, 42)
	AssertEqual(t, "hello", "hello")
	AssertEqual(t, true, true)
}

func TestAssertNil(t *testing.T) {
	// Should pass
	AssertNil(t, nil)
}

func TestAssertError(t *testing.T) {
	// Should pass
	AssertError(t, errors.New("test error"))
}

func TestAssertContains(t *testing.T) {
	// Should pass
	AssertContains(t, "hello world", "world")
	AssertContains(t, "Frankfurt Hbf", "Frankfurt")
}

func TestAssertNotContains(t *testing.T) {
	// Should pass
	AssertNotContains(t, "hello world", "foo")
}

func TestAssertTimeEqual(t *testing.T) {
	now := time.Now()
	later := now.Add(100 * time.Millisecond)

	// Should pass with tolerance
	AssertTimeEqual(t, now, later, 200*time.Millisecond)
}

func TestAssertFloatEqual(t *testing.T) {
	// Should pass
	AssertFloatEqual(t, 3.14159, 3.14, 0.01)
}

func TestAssertTrue(t *testing.T) {
	// Should pass
	AssertTrue(t, true)
	AssertTrue(t, 2 > 1)
}

func TestAssertFalse(t *testing.T) {
	// Should pass
	AssertFalse(t, false)
	AssertFalse(t, 1 == 2)
}

func TestAssertLen(t *testing.T) {
	// Should pass
	AssertLen(t, []int{1, 2, 3}, 3)
	AssertLen(t, []string{"a", "b"}, 2)
	AssertLen(t, []int{}, 0)
}
