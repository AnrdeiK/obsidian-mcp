// Package main provides the entry point for the MCP notes server.
// It initializes the vault and starts the MCP server with stdio transport.
package main

import (
	"fmt"
	"log"
	"os"

	"github.com/mark3labs/mcp-go/server"

	internalserver "github.com/kratos/mcp-notes/internal/server"
	"github.com/kratos/mcp-notes/internal/vault"
)

func main() {
	// Parse command-line arguments
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s <vault-path>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  %s /path/to/obsidian/vault\n", os.Args[0])
		os.Exit(1)
	}

	vaultPath := os.Args[1]

	// Create vault instance
	// NewVault validates that the path exists and is accessible
	v, err := vault.NewVault(vaultPath)
	if err != nil {
		log.Fatalf("Failed to create vault: %v", err)
	}

	// Create MCP server with registered tools
	srv := internalserver.NewServer(v)

	// Serve via stdio transport
	// This blocks until the server is shut down or an error occurs
	if err := server.ServeStdio(srv); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
