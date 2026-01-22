package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// ReadNoteTool returns the ServerTool for reading a note's content.
func (h *Handlers) ReadNoteTool() server.ServerTool {
	tool := mcp.NewTool(
		"read_note",
		mcp.WithDescription("Read the full content of a note by its path."),
		mcp.WithString(
			"path",
			mcp.Description("Path to the note file (relative to vault root, must end with .md)."),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	return server.ServerTool{
		Tool:    tool,
		Handler: h.handleReadNote,
	}
}

// handleReadNote implements the read_note tool handler.
func (h *Handlers) handleReadNote(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	path, err := request.RequireString("path")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Missing required parameter 'path': %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Call vault
	content, err := h.vault.Read(ctx, path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: formatVaultError(err, "reading", path),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: content,
			},
		},
		IsError: false,
	}, nil
}
