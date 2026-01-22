package tools

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

// SearchNotesTool returns the ServerTool for searching notes in the vault.
func (h *Handlers) SearchNotesTool() server.ServerTool {
	tool := mcp.NewTool(
		"search_notes",
		mcp.WithDescription("Search for notes matching a query and/or tag filters. Query uses regex pattern matching (case-insensitive)."),
		mcp.WithString(
			"query",
			mcp.Description("Regex pattern to search for in note content. Case-insensitive."),
			mcp.Required(),
		),
		mcp.WithString(
			"path",
			mcp.Description("Optional subdirectory path to search within. If empty, searches entire vault."),
		),
		mcp.WithArray(
			"tags",
			mcp.Description("Optional list of tags to filter by. Notes must have at least one of these tags."),
			mcp.WithStringItems(),
		),
		mcp.WithReadOnlyHintAnnotation(true),
		mcp.WithDestructiveHintAnnotation(false),
	)

	return server.ServerTool{
		Tool:    tool,
		Handler: h.handleSearchNotes,
	}
}

// handleSearchNotes implements the search_notes tool handler.
func (h *Handlers) handleSearchNotes(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	// Extract parameters
	query, err := request.RequireString("query")
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Missing required parameter 'query': %v", err),
				},
			},
			IsError: true,
		}, nil
	}

	path := request.GetString("path", "")
	tags := request.GetStringSlice("tags", nil)

	// Call vault
	notes, err := h.vault.Search(ctx, query, path, tags)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				mcp.TextContent{
					Type: "text",
					Text: fmt.Sprintf("Error searching notes: %v", err),
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
					Text: fmt.Sprintf("Error marshaling search results: %v", err),
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
