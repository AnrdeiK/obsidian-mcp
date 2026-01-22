package vault

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewVault(t *testing.T) {
	tests := []struct {
		name      string
		basePath  string
		wantError bool
	}{
		{
			name:      "valid directory",
			basePath:  t.TempDir(),
			wantError: false,
		},
		{
			name:      "nonexistent directory",
			basePath:  "/nonexistent/path",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := NewVault(tt.basePath)
			if tt.wantError {
				if err == nil {
					t.Error("Expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if v == nil {
					t.Error("Expected vault, got nil")
				}
			}
		})
	}
}

func TestNewVaultWithFile(t *testing.T) {
	tmpDir := t.TempDir()
	tmpFile := filepath.Join(tmpDir, "file.txt")
	if err := os.WriteFile(tmpFile, []byte("content"), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	_, err := NewVault(tmpFile)
	if err == nil {
		t.Error("Expected error when creating vault with file path")
	}
}

func TestValidatePath(t *testing.T) {
	tmpDir := t.TempDir()
	v, err := NewVault(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	vaultImpl := v.(*vault)

	tests := []struct {
		name      string
		path      string
		wantError error
	}{
		{
			name:      "valid path",
			path:      "notes/test.md",
			wantError: nil,
		},
		{
			name:      "path traversal with ..",
			path:      "../../../etc/passwd.md",
			wantError: ErrPathTraversal,
		},
		{
			name:      "path traversal hidden",
			path:      "notes/../../etc/passwd.md",
			wantError: ErrPathTraversal,
		},
		{
			name:      "non-markdown file",
			path:      "notes/test.txt",
			wantError: ErrNotMarkdown,
		},
		{
			name:      "empty path",
			path:      "",
			wantError: ErrInvalidPath,
		},
		{
			name:      "markdown file without extension",
			path:      "notes/test",
			wantError: ErrNotMarkdown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := vaultImpl.validatePath(tt.path)
			if tt.wantError != nil {
				if !errors.Is(err, tt.wantError) {
					t.Errorf("Expected error %v, got %v", tt.wantError, err)
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
			}
		})
	}
}

func setupTestVault(t *testing.T) (Vault, string) {
	tmpDir := t.TempDir()

	// Create test notes
	notes := map[string]string{
		"note1.md":             "This is note 1 with #tag1 and #tag2",
		"note2.md":             "This is note 2 with #tag2 and #tag3",
		"subdir/note3.md":      "This is note 3 in subdir with #tag1",
		"subdir/deep/note4.md": "This is note 4 deep down with #tag4",
		"other/note5.md":       "This is note 5 in other with #other",
		"readme.txt":           "This should be ignored",
		"subdir/.hidden.md":    "Hidden file with #hidden",
	}

	for path, content := range notes {
		fullPath := filepath.Join(tmpDir, path)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directory: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to create test file %s: %v", path, err)
		}
	}

	v, err := NewVault(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	return v, tmpDir
}

func TestList(t *testing.T) {
	v, _ := setupTestVault(t)
	ctx := context.Background()

	t.Run("list all non-recursive", func(t *testing.T) {
		notes, err := v.List(ctx, "", false)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		// Should only include note1.md and note2.md (root level only)
		if len(notes) != 2 {
			t.Errorf("Expected 2 notes, got %d", len(notes))
		}
	})

	t.Run("list all recursive", func(t *testing.T) {
		notes, err := v.List(ctx, "", true)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		// Should include all .md files (6 total)
		if len(notes) != 6 {
			t.Errorf("Expected 6 notes, got %d", len(notes))
		}
	})

	t.Run("list subdir non-recursive", func(t *testing.T) {
		notes, err := v.List(ctx, "subdir", false)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		// Should only include subdir/note3.md
		if len(notes) != 2 {
			t.Errorf("Expected 2 notes, got %d", len(notes))
		}
	})

	t.Run("list subdir recursive", func(t *testing.T) {
		notes, err := v.List(ctx, "subdir", true)
		if err != nil {
			t.Fatalf("List() error = %v", err)
		}

		// Should include subdir/note3.md, subdir/deep/note4.md, and .hidden.md
		if len(notes) != 3 {
			t.Errorf("Expected 3 notes, got %d", len(notes))
		}
	})

	t.Run("list with path traversal", func(t *testing.T) {
		_, err := v.List(ctx, "../../../etc", false)
		if !errors.Is(err, ErrPathTraversal) {
			t.Errorf("Expected ErrPathTraversal, got %v", err)
		}
	})
}

func TestSearch(t *testing.T) {
	v, _ := setupTestVault(t)
	ctx := context.Background()

	t.Run("search by content", func(t *testing.T) {
		notes, err := v.Search(ctx, "note 1", "", nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		if len(notes) != 1 {
			t.Errorf("Expected 1 note, got %d", len(notes))
		}

		if len(notes) > 0 && !strings.Contains(notes[0].Path, "note1.md") {
			t.Errorf("Expected note1.md, got %s", notes[0].Path)
		}
	})

	t.Run("search by tag", func(t *testing.T) {
		notes, err := v.Search(ctx, "", "", []string{"tag1"})
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		// Should find note1.md and subdir/note3.md
		if len(notes) != 2 {
			t.Errorf("Expected 2 notes with tag1, got %d", len(notes))
		}
	})

	t.Run("search by multiple tags", func(t *testing.T) {
		notes, err := v.Search(ctx, "", "", []string{"tag2", "tag3"})
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		// Should find notes with either tag2 OR tag3
		if len(notes) < 2 {
			t.Errorf("Expected at least 2 notes, got %d", len(notes))
		}
	})

	t.Run("search by content and tag", func(t *testing.T) {
		notes, err := v.Search(ctx, "subdir", "", []string{"tag1"})
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		// Should find only subdir/note3.md
		if len(notes) != 1 {
			t.Errorf("Expected 1 note, got %d", len(notes))
		}
	})

	t.Run("search with invalid regex", func(t *testing.T) {
		_, err := v.Search(ctx, "[invalid(", "", nil)
		if err == nil {
			t.Error("Expected error for invalid regex")
		}
	})

	t.Run("search in subpath", func(t *testing.T) {
		notes, err := v.Search(ctx, "", "subdir", nil)
		if err != nil {
			t.Fatalf("Search() error = %v", err)
		}

		// Should only search in subdir
		for _, note := range notes {
			if !strings.HasPrefix(note.Path, "subdir") {
				t.Errorf("Note %s is not in subdir", note.Path)
			}
		}
	})
}

func TestRead(t *testing.T) {
	v, tmpDir := setupTestVault(t)
	ctx := context.Background()

	t.Run("read existing note", func(t *testing.T) {
		content, err := v.Read(ctx, "note1.md")
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}

		if !strings.Contains(content, "note 1") {
			t.Errorf("Unexpected content: %s", content)
		}
	})

	t.Run("read from cache", func(t *testing.T) {
		// First read
		_, err := v.Read(ctx, "note1.md")
		if err != nil {
			t.Fatalf("First Read() error = %v", err)
		}

		// Second read (should hit cache)
		content, err := v.Read(ctx, "note1.md")
		if err != nil {
			t.Fatalf("Second Read() error = %v", err)
		}

		if !strings.Contains(content, "note 1") {
			t.Errorf("Unexpected cached content: %s", content)
		}
	})

	t.Run("read nonexistent note", func(t *testing.T) {
		_, err := v.Read(ctx, "nonexistent.md")
		if !errors.Is(err, ErrNoteNotFound) {
			t.Errorf("Expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("read with path traversal", func(t *testing.T) {
		_, err := v.Read(ctx, "../../../etc/passwd.md")
		if !errors.Is(err, ErrPathTraversal) {
			t.Errorf("Expected ErrPathTraversal, got %v", err)
		}
	})

	t.Run("read non-markdown file", func(t *testing.T) {
		_, err := v.Read(ctx, "readme.txt")
		if !errors.Is(err, ErrNotMarkdown) {
			t.Errorf("Expected ErrNotMarkdown, got %v", err)
		}
	})

	t.Run("read note in subdirectory", func(t *testing.T) {
		content, err := v.Read(ctx, "subdir/note3.md")
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}

		if !strings.Contains(content, "note 3") {
			t.Errorf("Unexpected content: %s", content)
		}
	})

	// Test cache invalidation on external modification
	t.Run("cache invalidation on file modification", func(t *testing.T) {
		// First read
		originalContent, err := v.Read(ctx, "note1.md")
		if err != nil {
			t.Fatalf("First Read() error = %v", err)
		}

		// Wait to ensure mtime changes (some filesystems have 1-second resolution)
		time.Sleep(1100 * time.Millisecond)

		// Modify file externally
		newContent := "Modified content with #newtag"
		fullPath := filepath.Join(tmpDir, "note1.md")
		if err := os.WriteFile(fullPath, []byte(newContent), 0644); err != nil {
			t.Fatalf("Failed to modify file: %v", err)
		}

		// Read again - should get new content
		updatedContent, err := v.Read(ctx, "note1.md")
		if err != nil {
			t.Fatalf("Second Read() error = %v", err)
		}

		if updatedContent == originalContent {
			t.Error("Expected cache to be invalidated after file modification")
		}

		if updatedContent != newContent {
			t.Errorf("Expected new content, got %s", updatedContent)
		}
	})
}

func TestCreate(t *testing.T) {
	tmpDir := t.TempDir()
	v, err := NewVault(tmpDir)
	if err != nil {
		t.Fatalf("Failed to create vault: %v", err)
	}

	ctx := context.Background()

	t.Run("create new note", func(t *testing.T) {
		content := "New note with #tag"
		err := v.Create(ctx, "new.md", content)
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		// Verify file exists
		fullPath := filepath.Join(tmpDir, "new.md")
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read created file: %v", err)
		}

		if string(data) != content {
			t.Errorf("File content = %s, want %s", string(data), content)
		}
	})

	t.Run("create note in subdirectory", func(t *testing.T) {
		err := v.Create(ctx, "subdir/nested/note.md", "Content")
		if err != nil {
			t.Fatalf("Create() error = %v", err)
		}

		fullPath := filepath.Join(tmpDir, "subdir/nested/note.md")
		if _, err := os.Stat(fullPath); err != nil {
			t.Error("Expected file to exist")
		}
	})

	t.Run("create existing note", func(t *testing.T) {
		// Create first time
		err := v.Create(ctx, "existing.md", "Content")
		if err != nil {
			t.Fatalf("First Create() error = %v", err)
		}

		// Try to create again
		err = v.Create(ctx, "existing.md", "New content")
		if err == nil {
			t.Error("Expected error when creating existing note")
		}
	})

	t.Run("create with path traversal", func(t *testing.T) {
		err := v.Create(ctx, "../../../tmp/evil.md", "Content")
		if !errors.Is(err, ErrPathTraversal) {
			t.Errorf("Expected ErrPathTraversal, got %v", err)
		}
	})

	t.Run("create non-markdown file", func(t *testing.T) {
		err := v.Create(ctx, "file.txt", "Content")
		if !errors.Is(err, ErrNotMarkdown) {
			t.Errorf("Expected ErrNotMarkdown, got %v", err)
		}
	})
}

func TestUpdate(t *testing.T) {
	v, tmpDir := setupTestVault(t)
	ctx := context.Background()

	t.Run("update existing note", func(t *testing.T) {
		newContent := "Updated content with #newtag"
		err := v.Update(ctx, "note1.md", newContent)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		// Verify update
		content, err := v.Read(ctx, "note1.md")
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}

		if content != newContent {
			t.Errorf("Content = %s, want %s", content, newContent)
		}

		// Verify file on disk
		fullPath := filepath.Join(tmpDir, "note1.md")
		data, err := os.ReadFile(fullPath)
		if err != nil {
			t.Fatalf("Failed to read file: %v", err)
		}

		if string(data) != newContent {
			t.Errorf("File content = %s, want %s", string(data), newContent)
		}
	})

	t.Run("update nonexistent note", func(t *testing.T) {
		err := v.Update(ctx, "nonexistent.md", "Content")
		if !errors.Is(err, ErrNoteNotFound) {
			t.Errorf("Expected ErrNoteNotFound, got %v", err)
		}
	})

	t.Run("update with path traversal", func(t *testing.T) {
		err := v.Update(ctx, "../../../etc/passwd.md", "Content")
		if !errors.Is(err, ErrPathTraversal) {
			t.Errorf("Expected ErrPathTraversal, got %v", err)
		}
	})

	t.Run("update non-markdown file", func(t *testing.T) {
		err := v.Update(ctx, "readme.txt", "Content")
		if !errors.Is(err, ErrNotMarkdown) {
			t.Errorf("Expected ErrNotMarkdown, got %v", err)
		}
	})

	t.Run("update note in subdirectory", func(t *testing.T) {
		newContent := "Updated subdir note"
		err := v.Update(ctx, "subdir/note3.md", newContent)
		if err != nil {
			t.Fatalf("Update() error = %v", err)
		}

		content, err := v.Read(ctx, "subdir/note3.md")
		if err != nil {
			t.Fatalf("Read() error = %v", err)
		}

		if content != newContent {
			t.Errorf("Content = %s, want %s", content, newContent)
		}
	})
}

func TestContextCancellation(t *testing.T) {
	v, _ := setupTestVault(t)

	t.Run("list with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel() // Cancel immediately

		_, err := v.List(ctx, "", true)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})

	t.Run("search with cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		_, err := v.Search(ctx, "query", "", nil)
		if !errors.Is(err, context.Canceled) {
			t.Errorf("Expected context.Canceled, got %v", err)
		}
	})
}

func TestNoteInfoTags(t *testing.T) {
	v, _ := setupTestVault(t)
	ctx := context.Background()

	notes, err := v.List(ctx, "", true)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	// Verify tags are extracted
	foundTaggedNote := false
	for _, note := range notes {
		if strings.Contains(note.Path, "note1.md") {
			foundTaggedNote = true
			if len(note.Tags) == 0 {
				t.Error("Expected note1.md to have tags")
			}

			// Check if tag1 is present
			hasTag1 := false
			for _, tag := range note.Tags {
				if tag == "tag1" {
					hasTag1 = true
					break
				}
			}

			if !hasTag1 {
				t.Errorf("Expected tag1 in note1.md tags: %v", note.Tags)
			}
		}
	}

	if !foundTaggedNote {
		t.Error("Did not find note1.md in list")
	}
}
