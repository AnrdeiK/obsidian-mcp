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

// CacheInterface defines the contract for note caching
// Implementations must be thread-safe
type CacheInterface interface {
	// Get retrieves a cache entry if it exists and is valid
	Get(path string) (CacheEntry, bool)
	// Set stores a cache entry with the given metadata
	Set(path string, content string, tags []string, mtime time.Time)
	// Delete removes a cache entry
	Delete(path string)
}

// Cache provides thread-safe caching of note content and metadata
// Cache entries are validated against file modification time
type Cache struct {
	mu      sync.RWMutex
	entries map[string]CacheEntry
}

// Ensure Cache implements CacheInterface
var _ CacheInterface = (*Cache)(nil)

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
		// Re-acquire lock to verify entry hasn't changed, then delete
		c.mu.Lock()
		if current, stillExists := c.entries[path]; stillExists && current.Mtime.Equal(entryMtime) {
			delete(c.entries, path)
		}
		c.mu.Unlock()
		return CacheEntry{}, false
	}

	fileMtime := stat.ModTime()

	// Re-acquire lock to compare and ensure entry hasn't been modified by another goroutine
	c.mu.RLock()
	current, stillExists := c.entries[path]
	c.mu.RUnlock()

	// If entry was modified/deleted while we were checking stat, return cache miss
	if !stillExists || !current.Mtime.Equal(entryMtime) {
		return CacheEntry{}, false
	}

	// Now check if file has been modified on disk
	if !fileMtime.Equal(entryMtime) {
		// Delete stale entry
		c.mu.Lock()
		if current, stillExists := c.entries[path]; stillExists && current.Mtime.Equal(entryMtime) {
			delete(c.entries, path)
		}
		c.mu.Unlock()
		return CacheEntry{}, false
	}

	// Create defensive copy of tags slice to prevent external modification
	tagsCopy := make([]string, len(entry.Tags))
	copy(tagsCopy, entry.Tags)

	return CacheEntry{
		Content: entry.Content,
		Tags:    tagsCopy,
		Mtime:   entry.Mtime,
	}, true
}

// Set stores a cache entry with the given metadata
func (c *Cache) Set(path string, content string, tags []string, mtime time.Time) {
	// Create defensive copy of tags to prevent external modification
	tagsCopy := make([]string, len(tags))
	copy(tagsCopy, tags)

	c.mu.Lock()
	c.entries[path] = CacheEntry{
		Content: content,
		Tags:    tagsCopy,
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
