package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// UpdateNoteTool returns the ServerTool for updating an existing note.
func (h *Handlers) UpdateNoteTool() server.ServerTool {
	tool := mcp.NewTool(
		"update_note",
		mcp.WithDescription("Update an existing note with new content. The note must already exist."),
		mcp.WithString(
			"path",
			mcp.Description("Path to the note to update (relative to vault root, must end with .md)."),
			mcp.Required(),
		),
		mcp.WithString(
			"content",
			mcp.Description("New content for the note in markdown format."),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(true),
	)

	return server.ServerTool{
		Tool:    tool,
		Handler: h.handleUpdateNote,
	}
}

// handleUpdateNote implements the update_note tool handler.
func (h *Handlers) handleUpdateNote(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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

	content, err := request.RequireString("content")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Missing required parameter 'content': %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	// Call vault
	err = h.vault.Update(ctx, path, content)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: formatVaultError(err, "updating", path),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully updated note: %s", path),
			},
		},
		IsError: false,
	}, nil
}
