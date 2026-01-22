package vault

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// NoteInfo represents metadata about a note
type NoteInfo struct {
	Path string   `json:"path"` // Relative path from vault root
	Tags []string `json:"tags"` // Extracted tags from content
}

// Vault provides operations for managing a collection of markdown notes
type Vault interface {
	// List returns all notes in the given subpath
	// If recursive is true, includes notes from subdirectories
	List(ctx context.Context, subpath string, recursive bool) ([]NoteInfo, error)

	// Search finds notes matching the query string and optional tag filters
	// Query is matched against note content using regex
	Search(ctx context.Context, query, subpath string, tags []string) ([]NoteInfo, error)

	// Read returns the content of a note
	Read(ctx context.Context, path string) (string, error)

	// Create creates a new note with the given content
	// Creates parent directories if they don't exist
	Create(ctx context.Context, path, content string) error

	// Update modifies an existing note
	Update(ctx context.Context, path, content string) error
}

// vault implements the Vault interface
var _ Vault = (*vault)(nil)

type vault struct {
	basePath string
	cache    *Cache
}

// NewVault creates a new vault instance
// basePath must exist and be a valid directory
func NewVault(basePath string) (Vault, error) {
	// Validate base path exists
	stat, err := os.Stat(basePath)
	if err != nil {
		return nil, fmt.Errorf("base path validation failed: %w", err)
	}

	if !stat.IsDir() {
		return nil, fmt.Errorf("base path is not a directory: %s", basePath)
	}

	// Clean and absolutize base path
	absPath, err := filepath.Abs(basePath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve absolute path: %w", err)
	}

	return &vault{
		basePath: absPath,
		cache:    NewCache(),
	}, nil
}

// validatePath ensures the path is safe and returns the full filesystem path
func (v *vault) validatePath(path string) (string, error) {
	if path == "" {
		return "", ErrInvalidPath
	}

	// Clean the path to resolve . and ..
	cleaned := filepath.Clean(path)

	// Check for path traversal attempts
	if strings.Contains(cleaned, "..") {
		return "", ErrPathTraversal
	}

	// Build full path
	fullPath := filepath.Join(v.basePath, cleaned)

	// Ensure the resolved path is still within basePath
	if !strings.HasPrefix(fullPath, v.basePath) {
		return "", ErrPathTraversal
	}

	// Ensure it's a markdown file
	if !strings.HasSuffix(fullPath, ".md") {
		return "", ErrNotMarkdown
	}

	return fullPath, nil
}

// List returns all notes in the given subpath
func (v *vault) List(ctx context.Context, subpath string, recursive bool) ([]NoteInfo, error) {
	// Build search directory
	searchPath := v.basePath
	if subpath != "" {
		cleaned := filepath.Clean(subpath)
		if strings.Contains(cleaned, "..") {
			return nil, ErrPathTraversal
		}
		searchPath = filepath.Join(v.basePath, cleaned)
	}

	// Ensure search path is within vault
	if !strings.HasPrefix(searchPath, v.basePath) {
		return nil, ErrPathTraversal
	}

	var notes []NoteInfo

	walkFn := func(path string, info os.FileInfo, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			// Skip inaccessible files/directories
			return nil
		}

		// Skip directories
		if info.IsDir() {
			if !recursive && path != searchPath {
				return filepath.SkipDir
			}
			return nil
		}

		// Only include .md files
		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Get relative path from vault root
		relPath, err := filepath.Rel(v.basePath, path)
		if err != nil {
			return nil // Skip files we can't get relative path for
		}

		// Try to get tags from cache
		var tags []string
		if entry, ok := v.cache.Get(path); ok {
			tags = entry.Tags
		} else {
			// Read file to extract tags
			content, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip unreadable files
			}

			tags = ExtractTags(string(content))
			v.cache.Set(path, string(content), tags, info.ModTime())
		}

		notes = append(notes, NoteInfo{
			Path: relPath,
			Tags: tags,
		})

		return nil
	}

	if err := filepath.Walk(searchPath, walkFn); err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return notes, nil
}

// Search finds notes matching the query and optional tag filters
func (v *vault) Search(ctx context.Context, query, subpath string, tags []string) ([]NoteInfo, error) {
	// Build search directory
	searchPath := v.basePath
	if subpath != "" {
		cleaned := filepath.Clean(subpath)
		if strings.Contains(cleaned, "..") {
			return nil, ErrPathTraversal
		}
		searchPath = filepath.Join(v.basePath, cleaned)
	}

	// Ensure search path is within vault
	if !strings.HasPrefix(searchPath, v.basePath) {
		return nil, ErrPathTraversal
	}

	// Compile query regex if provided
	var queryRegex *regexp.Regexp
	if query != "" {
		var err error
		queryRegex, err = regexp.Compile("(?i)" + query) // Case-insensitive
		if err != nil {
			return nil, fmt.Errorf("invalid query regex: %w", err)
		}
	}

	// Normalize tag filter to lowercase for comparison
	tagFilter := make(map[string]struct{})
	for _, tag := range tags {
		tagFilter[strings.ToLower(tag)] = struct{}{}
	}

	var results []NoteInfo

	walkFn := func(path string, info os.FileInfo, err error) error {
		// Check context cancellation
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if err != nil {
			return nil // Skip inaccessible files
		}

		if info.IsDir() {
			return nil
		}

		if !strings.HasSuffix(path, ".md") {
			return nil
		}

		// Read file content
		var content string
		var noteTags []string

		if entry, ok := v.cache.Get(path); ok {
			content = entry.Content
			noteTags = entry.Tags
		} else {
			data, err := os.ReadFile(path)
			if err != nil {
				return nil // Skip unreadable files
			}
			content = string(data)
			noteTags = ExtractTags(content)
			v.cache.Set(path, content, noteTags, info.ModTime())
		}

		// Apply query filter
		if queryRegex != nil && !queryRegex.MatchString(content) {
			return nil
		}

		// Apply tag filter
		if len(tagFilter) > 0 {
			hasMatchingTag := false
			for _, tag := range noteTags {
				if _, exists := tagFilter[tag]; exists {
					hasMatchingTag = true
					break
				}
			}
			if !hasMatchingTag {
				return nil
			}
		}

		// Get relative path
		relPath, err := filepath.Rel(v.basePath, path)
		if err != nil {
			return nil
		}

		results = append(results, NoteInfo{
			Path: relPath,
			Tags: noteTags,
		})

		return nil
	}

	if err := filepath.Walk(searchPath, walkFn); err != nil {
		return nil, fmt.Errorf("failed to walk directory: %w", err)
	}

	return results, nil
}

// Read returns the content of a note
func (v *vault) Read(ctx context.Context, path string) (string, error) {
	fullPath, err := v.validatePath(path)
	if err != nil {
		return "", err
	}

	// Check if file exists
	stat, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", ErrNoteNotFound
		}
		return "", fmt.Errorf("failed to stat file: %w", err)
	}

	// Check cache first
	if entry, ok := v.cache.Get(fullPath); ok {
		return entry.Content, nil
	}

	// Read from filesystem
	data, err := os.ReadFile(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %w", err)
	}

	content := string(data)
	tags := ExtractTags(content)
	v.cache.Set(fullPath, content, tags, stat.ModTime())

	return content, nil
}

// Create creates a new note with the given content
func (v *vault) Create(ctx context.Context, path, content string) error {
	fullPath, err := v.validatePath(path)
	if err != nil {
		return err
	}

	// Check if file already exists
	if _, err := os.Stat(fullPath); err == nil {
		return fmt.Errorf("note already exists: %s", path)
	}

	// Create parent directories
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update cache
	stat, err := os.Stat(fullPath)
	if err == nil {
		tags := ExtractTags(content)
		v.cache.Set(fullPath, content, tags, stat.ModTime())
	}

	return nil
}

// Update modifies an existing note
func (v *vault) Update(ctx context.Context, path, content string) error {
	fullPath, err := v.validatePath(path)
	if err != nil {
		return err
	}

	// Check if file exists
	if _, err := os.Stat(fullPath); err != nil {
		if os.IsNotExist(err) {
			return ErrNoteNotFound
		}
		return fmt.Errorf("failed to stat file: %w", err)
	}

	// Write file
	if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	// Update cache
	stat, err := os.Stat(fullPath)
	if err == nil {
		tags := ExtractTags(content)
		v.cache.Set(fullPath, content, tags, stat.ModTime())
	}

	return nil
}
