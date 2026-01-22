// Package tools provides MCP tool handlers for the notes server.
package tools

import (
	"github.com/kratos/mcp-notes/internal/vault"
	"github.com/mark3labs/mcp-go/server"
)

// Handlers aggregates all tool handlers for the MCP notes server.
// It provides a central point for registering tools with the MCP server.
type Handlers struct {
	vault vault.Vault
}

// NewHandlers creates a new Handlers instance with the given vault.
func NewHandlers(v vault.Vault) *Handlers {
	return &Handlers{
		vault: v,
	}
}

// RegisterTools registers all tool handlers with the MCP server.
// This should be called during server initialization.
func (h *Handlers) RegisterTools(srv *server.MCPServer) {
	srv.AddTools(
		h.ListNotesTool(),
		h.SearchNotesTool(),
		h.ReadNoteTool(),
		h.CreateNoteTool(),
		h.UpdateNoteTool(),
	)
}
