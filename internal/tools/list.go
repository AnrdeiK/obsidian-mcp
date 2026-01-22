package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ListNotesTool returns the ServerTool for listing notes in the vault.
func (h *Handlers) ListNotesTool() server.ServerTool {
	tool := mcp.NewTool(
		"list_notes",
		mcp.WithDescription("List all notes in the vault or a specific subdirectory. Returns note paths with their tags."),
		mcp.WithString(
			"path",
			mcp.Description("Optional subdirectory path to list notes from. If empty, lists from vault root."),
		),
		mcp.WithBoolean(
			"recursive",
			mcp.Description("Whether to recursively list notes in subdirectories."),
			mcp.DefaultBool(true),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	return server.ServerTool{
		Tool:    tool,
		Handler: h.handleListNotes,
	}
}

// handleListNotes implements the list_notes tool handler.
func (h *Handlers) handleListNotes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	path := request.GetString("path", "")
	recursive := request.GetBool("recursive", true)

	// Call vault
	notes, err := h.vault.List(ctx, path, recursive)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error listing notes: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Marshal notes to JSON
	notesJSON, err := json.MarshalIndent(notes, "", "  ")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error marshaling notes: %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: string(notesJSON),
			},
		},
		IsError: false,
	}, nil
}
