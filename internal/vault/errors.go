package vault

import "errors"

// Sentinel errors for vault operations
var (
	// ErrNoteNotFound indicates the requested note does not exist
	ErrNoteNotFound = errors.New("note not found")

	// ErrPathTraversal indicates an attempt to access a path outside the vault
	ErrPathTraversal = errors.New("path traversal not allowed")

	// ErrInvalidPath indicates the path format is invalid
	ErrInvalidPath = errors.New("invalid path")

	// ErrNotMarkdown indicates the file is not a markdown file
	ErrNotMarkdown = errors.New("only .md files allowed")
)
