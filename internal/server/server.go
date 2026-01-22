package server

import (
	"github.com/mark3labs/mcp-go/server"

	"github.com/kratos/mcp-notes/internal/tools"
	"github.com/kratos/mcp-notes/internal/vault"
)

// NewServer creates a new MCP server configured with all note tools.
// It initializes the server with the "notes" identifier and registers
// all tools provided by the tools package.
//
// The vault parameter provides access to the notes storage backend.
func NewServer(v vault.Vault) *server.MCPServer {
	// Create MCP server with name "notes" and version "1.0.0"
	srv := server.NewMCPServer("notes", "1.0.0")

	// Create handlers with vault dependency
	handlers := tools.NewHandlers(v)

	// Register all tools with the server
	handlers.RegisterTools(srv)

	return srv
}
