package tools

import (
	"context"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// CreateNoteTool returns the ServerTool for creating a new note.
func (h *Handlers) CreateNoteTool() server.ServerTool {
	tool := mcp.NewTool(
		"create_note",
		mcp.WithDescription("Create a new note with the given content. Parent directories are created automatically if needed."),
		mcp.WithString(
			"path",
			mcp.Description("Path for the new note (relative to vault root, must end with .md)."),
			mcp.Required(),
		),
		mcp.WithString(
			"content",
			mcp.Description("Content of the note in markdown format."),
			mcp.Required(),
		),
		mcp.WithReadOnlyHintAnnotation(false),
		mcp.WithDestructiveHintAnnotation(false),
	)

	return server.ServerTool{
		Tool:    tool,
		Handler: h.handleCreateNote,
	}
}

// handleCreateNote implements the create_note tool handler.
func (h *Handlers) handleCreateNote(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
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
	err = h.vault.Create(ctx, path, content)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: formatVaultError(err, "creating", path),
				},
			},
			IsError: true,
		}, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			mcp.TextContent{
				Type: "text",
				Text: fmt.Sprintf("Successfully created note: %s", path),
			},
		},
		IsError: false,
	}, nil
}
