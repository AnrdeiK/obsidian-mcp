package vault

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewCache(t *testing.T) {
	cache := NewCache()
	if cache == nil {
		t.Fatal("NewCache() returned nil")
	}
	if cache.entries == nil {
		t.Error("Cache entries map is nil")
	}
}

func TestCacheSetAndGet(t *testing.T) {
	cache := NewCache()

	// Create a temporary file for testing
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	content := "test content"
	tags := []string{"tag1", "tag2"}

	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	stat, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	// Set cache entry
	cache.Set(tmpFile, content, tags, stat.ModTime())

	// Get cache entry
	entry, ok := cache.Get(tmpFile)
	if !ok {
		t.Error("Expected to get cache entry")
	}

	if entry.Content != content {
		t.Errorf("Content = %v, want %v", entry.Content, content)
	}

	if len(entry.Tags) != len(tags) {
		t.Errorf("Tags length = %d, want %d", len(entry.Tags), len(tags))
	}
}

func TestCacheGetNonExistent(t *testing.T) {
	cache := NewCache()

	_, ok := cache.Get("nonexistent.md")
	if ok {
		t.Error("Expected false for nonexistent cache entry")
	}
}

func TestCacheInvalidation(t *testing.T) {
	cache := NewCache()

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")
	content := "initial content"

	if err := os.WriteFile(tmpFile, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	stat, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	// Set cache entry
	cache.Set(tmpFile, content, []string{"tag"}, stat.ModTime())

	// Verify cache hit
	_, ok := cache.Get(tmpFile)
	if !ok {
		t.Error("Expected cache hit")
	}

	// Wait a bit and modify the file
	time.Sleep(10 * time.Millisecond)
	newContent := "updated content"
	if err := os.WriteFile(tmpFile, []byte(newContent), 0644); err != nil {
		t.Fatalf("Failed to update test file: %v", err)
	}

	// Cache should be invalidated
	_, ok = cache.Get(tmpFile)
	if ok {
		t.Error("Expected cache miss after file modification")
	}
}

func TestCacheInvalidationFileDeleted(t *testing.T) {
	cache := NewCache()

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")

	if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	stat, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	// Set cache entry
	cache.Set(tmpFile, "content", []string{}, stat.ModTime())

	// Delete the file
	if err := os.Remove(tmpFile); err != nil {
		t.Fatalf("Failed to delete test file: %v", err)
	}

	// Cache should be invalidated
	_, ok := cache.Get(tmpFile)
	if ok {
		t.Error("Expected cache miss after file deletion")
	}
}

func TestCacheDelete(t *testing.T) {
	cache := NewCache()

	// Create a temporary file
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "test.md")

	if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	stat, err := os.Stat(tmpFile)
	if err != nil {
		t.Fatalf("Failed to stat test file: %v", err)
	}

	// Set and verify cache entry exists
	cache.Set(tmpFile, "content", []string{}, stat.ModTime())
	_, ok := cache.Get(tmpFile)
	if !ok {
		t.Error("Expected cache hit")
	}

	// Delete cache entry
	cache.Delete(tmpFile)

	// Verify deletion
	_, ok = cache.Get(tmpFile)
	if ok {
		t.Error("Expected cache miss after deletion")
	}
}

func TestCacheConcurrency(t *testing.T) {
	cache := NewCache()
	tmpDir := t.TempDir()

	// Create multiple test files
	files := make([]string, 10)
	for i := 0; i < 10; i++ {
		tmpFile := filepath.Join(tmpDir, "test_"+fmt.Sprintf("%d", i)+".md")
		if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}
		files[i] = tmpFile
	}

	// Concurrent writes and reads
	done := make(chan bool)

	// Writers
	for i := 0; i < 5; i++ {
		go func(idx int) {
			for j := 0; j < 100; j++ {
				file := files[idx%len(files)]
				stat, _ := os.Stat(file)
				cache.Set(file, "content", []string{"tag"}, stat.ModTime())
			}
			done <- true
		}(i)
	}

	// Readers
	for i := 0; i < 5; i++ {
		go func(idx int) {
			for j := 0; j < 100; j++ {
				file := files[idx%len(files)]
				cache.Get(file)
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}
