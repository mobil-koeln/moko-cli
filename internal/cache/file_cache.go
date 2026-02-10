package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// FileCache implements a file-based cache with TTL
type FileCache struct {
	dir string
	ttl time.Duration
}

// cacheEntry represents a cached item with expiration
type cacheEntry struct {
	Data      []byte    `json:"data"`
	ExpiresAt time.Time `json:"expires_at"`
}

// NewFileCache creates a new file cache
func NewFileCache(dir string, ttl time.Duration) (*FileCache, error) {
	// Create cache directory if it doesn't exist (0750 for security)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return nil, err
	}

	return &FileCache{
		dir: dir,
		ttl: ttl,
	}, nil
}

// DefaultCacheDir returns the default cache directory
func DefaultCacheDir() string {
	// Check XDG_CACHE_HOME first
	if xdgCache := os.Getenv("XDG_CACHE_HOME"); xdgCache != "" {
		return filepath.Join(xdgCache, "moko")
	}

	// Fall back to ~/.cache/moko
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(os.TempDir(), "moko-cache")
	}

	return filepath.Join(home, ".cache", "moko")
}

// keyToFilename converts a cache key (URL) to a filename
func (c *FileCache) keyToFilename(key string) string {
	hash := sha256.Sum256([]byte(key))
	return filepath.Join(c.dir, hex.EncodeToString(hash[:])+".json")
}

// Get retrieves a value from the cache
func (c *FileCache) Get(key string) ([]byte, bool) {
	filename := c.keyToFilename(key)

	// #nosec G304 -- filename is derived from hash of cache key, not user input
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, false
	}

	var entry cacheEntry
	if err := json.Unmarshal(data, &entry); err != nil {
		// Invalid cache entry, remove it
		_ = os.Remove(filename)
		return nil, false
	}

	// Check if expired
	if time.Now().After(entry.ExpiresAt) {
		_ = os.Remove(filename)
		return nil, false
	}

	return entry.Data, true
}

// Set stores a value in the cache
func (c *FileCache) Set(key string, value []byte) error {
	entry := cacheEntry{
		Data:      value,
		ExpiresAt: time.Now().Add(c.ttl),
	}

	data, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	filename := c.keyToFilename(key)
	// Use 0600 for cache files to restrict access to owner only
	return os.WriteFile(filename, data, 0600)
}

// Clear removes all cache entries
func (c *FileCache) Clear() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ".json" {
			_ = os.Remove(filepath.Join(c.dir, entry.Name()))
		}
	}

	return nil
}

// Cleanup removes expired entries
func (c *FileCache) Cleanup() error {
	entries, err := os.ReadDir(c.dir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		filename := filepath.Join(c.dir, entry.Name())
		// #nosec G304 -- filename is from ReadDir within cache directory
		data, err := os.ReadFile(filename)
		if err != nil {
			continue
		}

		var ce cacheEntry
		if err := json.Unmarshal(data, &ce); err != nil {
			_ = os.Remove(filename)
			continue
		}

		if time.Now().After(ce.ExpiresAt) {
			_ = os.Remove(filename)
		}
	}

	return nil
}
