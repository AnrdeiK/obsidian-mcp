# MCP Notes Server

MCP server for working with markdown notes through Claude Code.

## Features

- Full-text search with regex support
- Tag-based search and filtering (#tag)
- In-memory cache with mtime-based invalidation
- Path traversal protection
- Markdown files only (.md)

## Installation

### 1. Build the binary

```bash
cd ~/Devs/mcp-notes
go install .
```

Binary will be at `~/go/bin/mcp-notes`.

### 2. Create Claude Code plugin

Create the plugin directory structure:

```bash
mkdir -p ~/.claude/plugins/marketplaces/claude-plugins-official/external_plugins/notes/.claude-plugin
```

Create `.claude-plugin/plugin.json`:

```json
{
  "name": "notes",
  "description": "Obsidian-style markdown notes MCP server",
  "author": {
    "name": "your-name"
  }
}
```

Create `.mcp.json` (in the notes folder, not in .claude-plugin):

```json
{
  "notes": {
    "command": "/home/you/go/bin/mcp-notes",
    "args": ["/path/to/your/notes"]
  }
}
```

### 3. Restart Claude Code

```bash
claude --resume
```

The `mcp__notes__*` tools will now be available to Claude and all agents.

## Tools

| Tool | Description | Parameters |
|------|-------------|------------|
| `list_notes` | List .md files | `path?`, `recursive?` |
| `search_notes` | Search by content and tags | `query`, `path?`, `tags?` |
| `read_note` | Read note content | `path` |
| `create_note` | Create a new note | `path`, `content` |
| `update_note` | Update existing note | `path`, `content` |

## Usage Examples

```
# List all notes
mcp__notes__list_notes

# Recursive list in a folder
mcp__notes__list_notes path="projects" recursive=true

# Search by text
mcp__notes__search_notes query="TODO"

# Search by tags
mcp__notes__search_notes tags=["work", "important"]

# Read a note
mcp__notes__read_note path="projects/ideas.md"

# Create a note
mcp__notes__create_note path="inbox/new-idea.md" content="# New Idea\n\nContent here"

# Update a note
mcp__notes__update_note path="inbox/new-idea.md" content="# Updated\n\nNew content"
```

## Project Structure

```
mcp-notes/
├── main.go                 # Entry point
├── internal/
│   ├── server/             # MCP server setup
│   ├── tools/              # Tool handlers
│   └── vault/              # Storage + cache
├── go.mod
└── go.sum
```

## Testing

```bash
go test ./... -v
go test ./... -cover
```

## Security

- Vault path is passed as a command-line argument
- Path traversal is forbidden (`..` in paths)
- Operations restricted to the specified vault directory
- Only .md files are accessible
- No authentication needed — stdio transport, local subprocess

## License

MIT
