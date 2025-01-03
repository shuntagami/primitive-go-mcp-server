# MCP Image Generation Server

A Go implementation of an MCP (Model Context Protocol) server that generates images using OpenAI's DALL-E API. This server demonstrates how to build MCP tools that can be used by Large Language Models like Claude.

## Features

- Generate images from text descriptions
- Automatic handling of save locations
- Configurable image dimensions
- Proper error handling and logging
- Works with Claude Desktop and other MCP clients

## Prerequisites

- Go 1.19 or higher
- OpenAI API key
- Claude Desktop (for testing)

## Build command
```
go build -o ./bin/imagegen-go ./main
```

## Configuration

Add this server to your Claude Desktop configuration at `~/Library/Application Support/Claude/claude_desktop_config.json`:

```json
{
  "mcpServers": {
   "imagegen-go": {
      "command": "/path/to/imagegen-go/bin/imagegen-go",
      "env": {
        "OPENAI_API_KEY": "your-api-key",
        "DEFAULT_DOWNLOAD_PATH":"/path/to/downloads"
      }
    }
  }
}
```

## Usage

1. Build the server using the command above
2. Configure Claude Desktop with your server path and API key
3. Restart Claude Desktop
4. Ask Claude to generate images!

Example prompt:
"Can you generate an image of a riverside home in cinematic style?"

## Implementation Details

This server implements the MCP tools capability and provides a single tool:
- `generate-image`: Generates an image from a text prompt using OpenAI's DALL-E


## License

MIT License
