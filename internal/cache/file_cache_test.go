package cache

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewFileCache(t *testing.T) {
	dir := t.TempDir()
	ttl := 60 * time.Second

	cache, err := NewFileCache(dir, ttl)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}
	if cache == nil {
		t.Fatal("NewFileCache() returned nil")
	}
}

func TestFileCache_SetAndGet(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, 60*time.Second)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	key := "https://example.com/api/test"
	value := []byte(`{"test": "data"}`)

	// Set the value
	if err := cache.Set(key, value); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Get the value
	got, ok := cache.Get(key)
	if !ok {
		t.Fatal("Get() returned false, want true")
	}
	if string(got) != string(value) {
		t.Errorf("Get() = %q, want %q", got, value)
	}
}

func TestFileCache_GetMissing(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, 60*time.Second)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	_, ok := cache.Get("non-existent-key")
	if ok {
		t.Error("Get() returned true for non-existent key")
	}
}

func TestFileCache_Expiration(t *testing.T) {
	dir := t.TempDir()
	// Very short TTL for testing
	cache, err := NewFileCache(dir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	key := "https://example.com/api/expire-test"
	value := []byte(`{"test": "expiration"}`)

	// Set the value
	if err := cache.Set(key, value); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Value should be available immediately
	if _, ok := cache.Get(key); !ok {
		t.Error("Get() returned false immediately after Set()")
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Value should be expired now
	if _, ok := cache.Get(key); ok {
		t.Error("Get() returned true for expired key")
	}
}

func TestFileCache_HashKey(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, 60*time.Second)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	// Test that different keys produce different cache files
	key1 := "https://example.com/api/1"
	key2 := "https://example.com/api/2"

	if err := cache.Set(key1, []byte("data1")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}
	if err := cache.Set(key2, []byte("data2")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Both should be retrievable
	data1, ok1 := cache.Get(key1)
	data2, ok2 := cache.Get(key2)

	if !ok1 || !ok2 {
		t.Error("Failed to retrieve one or both keys")
	}
	if string(data1) != "data1" || string(data2) != "data2" {
		t.Error("Data mismatch")
	}
}

func TestFileCache_CreateDirectory(t *testing.T) {
	// Use a nested directory that doesn't exist
	baseDir := t.TempDir()
	nestedDir := filepath.Join(baseDir, "nested", "cache", "dir")

	cache, err := NewFileCache(nestedDir, 60*time.Second)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	// Verify directory was created
	if _, err := os.Stat(nestedDir); os.IsNotExist(err) {
		t.Error("Cache directory was not created")
	}

	// Verify cache works
	if err := cache.Set("test", []byte("data")); err != nil {
		t.Errorf("Set() error = %v", err)
	}
}

func TestDefaultCacheDir(t *testing.T) {
	dir := DefaultCacheDir()
	if dir == "" {
		t.Error("DefaultCacheDir() returned empty string")
	}
}

func TestFileCache_Clear(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, 60*time.Second)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	// Add multiple cache entries
	keys := []string{
		"https://example.com/api/1",
		"https://example.com/api/2",
		"https://example.com/api/3",
	}
	for i, key := range keys {
		value := []byte(`{"data": "` + string(rune('0'+i)) + `"}`)
		if err := cache.Set(key, value); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// Verify all entries exist
	for _, key := range keys {
		if _, ok := cache.Get(key); !ok {
			t.Errorf("Get(%q) returned false before Clear()", key)
		}
	}

	// Clear the cache
	if err := cache.Clear(); err != nil {
		t.Fatalf("Clear() error = %v", err)
	}

	// Verify all entries are gone
	for _, key := range keys {
		if _, ok := cache.Get(key); ok {
			t.Errorf("Get(%q) returned true after Clear()", key)
		}
	}

	// Verify cache directory still exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		t.Error("Cache directory was deleted by Clear()")
	}
}

func TestFileCache_ClearEmptyCache(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, 60*time.Second)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	// Clear empty cache should not error
	if err := cache.Clear(); err != nil {
		t.Errorf("Clear() on empty cache error = %v", err)
	}
}

func TestFileCache_Cleanup(t *testing.T) {
	dir := t.TempDir()
	// Short TTL for testing
	cache, err := NewFileCache(dir, 100*time.Millisecond)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	// Add some entries
	oldKeys := []string{
		"https://example.com/api/old1",
		"https://example.com/api/old2",
	}
	for _, key := range oldKeys {
		if err := cache.Set(key, []byte("old data")); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// Wait for entries to expire
	time.Sleep(150 * time.Millisecond)

	// Add fresh entry
	freshKey := "https://example.com/api/fresh"
	if err := cache.Set(freshKey, []byte("fresh data")); err != nil {
		t.Fatalf("Set() error = %v", err)
	}

	// Run cleanup
	if err := cache.Cleanup(); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	// Old entries should be gone
	for _, key := range oldKeys {
		if _, ok := cache.Get(key); ok {
			t.Errorf("Get(%q) returned true after Cleanup(), expired entry not removed", key)
		}
	}

	// Fresh entry should still exist
	if _, ok := cache.Get(freshKey); !ok {
		t.Error("Get(freshKey) returned false after Cleanup(), fresh entry was removed")
	}
}

func TestFileCache_CleanupEmptyCache(t *testing.T) {
	dir := t.TempDir()
	cache, err := NewFileCache(dir, 60*time.Second)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	// Cleanup empty cache should not error
	if err := cache.Cleanup(); err != nil {
		t.Errorf("Cleanup() on empty cache error = %v", err)
	}
}

func TestFileCache_CleanupAllExpired(t *testing.T) {
	dir := t.TempDir()
	// Very short TTL
	cache, err := NewFileCache(dir, 50*time.Millisecond)
	if err != nil {
		t.Fatalf("NewFileCache() error = %v", err)
	}

	// Add entries
	keys := []string{
		"https://example.com/api/1",
		"https://example.com/api/2",
		"https://example.com/api/3",
	}
	for _, key := range keys {
		if err := cache.Set(key, []byte("data")); err != nil {
			t.Fatalf("Set() error = %v", err)
		}
	}

	// Wait for all to expire
	time.Sleep(100 * time.Millisecond)

	// Cleanup should remove all
	if err := cache.Cleanup(); err != nil {
		t.Fatalf("Cleanup() error = %v", err)
	}

	// All should be gone
	for _, key := range keys {
		if _, ok := cache.Get(key); ok {
			t.Errorf("Get(%q) returned true after Cleanup(), all entries should be expired", key)
		}
	}
}
