package tools

import (
	"errors"
	"fmt"

	"github.com/kratos/mcp-notes/internal/vault"
)

// Error message constants
const (
	errMsgPathTraversal = "Invalid path: path traversal not allowed"
	errMsgInvalidPath   = "Invalid path format"
	errMsgNotMarkdown   = "Only .md files are allowed"
)

// formatVaultError converts vault errors to user-friendly messages
func formatVaultError(err error, operation, path string) string {
	switch {
	case errors.Is(err, vault.ErrNoteNotFound):
		return fmt.Sprintf("Note not found: %s", path)
	case errors.Is(err, vault.ErrPathTraversal):
		return errMsgPathTraversal
	case errors.Is(err, vault.ErrInvalidPath):
		return errMsgInvalidPath
	case errors.Is(err, vault.ErrNotMarkdown):
		return errMsgNotMarkdown
	default:
		return fmt.Sprintf("Error %s note: %v", operation, err)
	}
}
