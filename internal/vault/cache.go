package vault

import (
	"os"
	"sync"
	"time"
)

// CacheEntry represents a cached note with its metadata
type CacheEntry struct {
	Content string    // File content
	Tags    []string  // Extracted tags
	Mtime   time.Time // File modification time
}

// Cache provides thread-safe caching of note content and metadata
// Cache entries are validated against file modification time
type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
}

// NewCache creates a new cache instance
func NewCache() *Cache {
	return &Cache{
		entries: make(map[string]CacheEntry),
	}
}

// Get retrieves a cache entry if it exists and is valid
// Returns the entry and true if found and valid, otherwise empty entry and false
// Validates cache freshness by comparing modification times
func (c *Cache) Get(path string) (CacheEntry, bool) {
	c.mu.RLock()
	entry, exists := c.entries[path]
	entryMtime := entry.Mtime
	c.mu.RUnlock()

	if !exists {
		return CacheEntry{}, false
	}

	// Check mtime outside lock
	stat, err := os.Stat(path)
	if err != nil {
		// Double-check before deletion
		c.mu.Lock()
		if current, stillExists := c.entries[path]; stillExists && current.Mtime.Equal(entryMtime) {
			delete(c.entries, path)
		}
		c.mu.Unlock()
		return CacheEntry{}, false
	}

	if !stat.ModTime().Equal(entryMtime) {
		// Double-check before deletion
		c.mu.Lock()
		if current, stillExists := c.entries[path]; stillExists && current.Mtime.Equal(entryMtime) {
			delete(c.entries, path)
		}
		c.mu.Unlock()
		return CacheEntry{}, false
	}

	return entry, true
}

// Set stores a cache entry with the given metadata
func (c *Cache) Set(path string, content string, tags []string, mtime time.Time) {
	c.mu.Lock()
	c.entries[path] = CacheEntry{
		Content: content,
		Tags:    tags,
		Mtime:   mtime,
	}
	c.mu.Unlock()
}

// Delete removes a cache entry
func (c *Cache) Delete(path string) {
	c.mu.Lock()
	delete(c.entries, path)
	c.mu.Unlock()
}
