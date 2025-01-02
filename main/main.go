package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/prasanth/myservers/imagegen-go/pkg/openai"
)

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)
	// Set up logging to stderr
	log.SetOutput(os.Stderr)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	log.Printf("Starting imagegen-go MCP server ...")

	for {
		var request JSONRPCRequest
		if err := decoder.Decode(&request); err != nil {
			log.Printf("Error decoding request: %v", err)
			sendError(encoder, nil, ParseError, "Failed to parse JSON")
			break
		}

		log.Printf("Received request: %v", PrettyJSON(request))

		if request.JSONRPC != "2.0" {
			sendError(encoder, request.ID, InvalidRequest, "Only JSON-RPC 2.0 is supported")
			continue
		}

		var response interface{}

		switch request.Method {
		case "initialize":
			response = JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result: InitializeResult{
					ProtocolVersion: "2024-11-05",
					ServerInfo: ServerInfo{
						Name:    "imagegen-go",
						Version: "1.0.0",
					},
					Capabilities: Capabilities{
						Tools: map[string]interface{}{},
					},
				},
			}

		case "notifications/initialized", "initialized":
			log.Printf("Server initialized successfully")
			continue // Skip sending response for notifications

		case "tools/list":
			toolSchema := json.RawMessage(`{
                "type": "object",
                "properties": {
                    "prompt": {
                        "type": "string",
                        "description": "Description of the image to generate"
                    },
                    "width": {
                        "type": "number",
                        "description": "Width of the image in pixels",
                        "default": 512
                    },
                    "height": {
                        "type": "number",
                        "description": "Height of the image in pixels",
                        "default": 512
                    },
                    "destination": {
                        "type": "string",
                        "description": "Path where the generated image should be saved"
                    }
                },
                "required": ["prompt", "destination"]
            }`)

			response = JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result: ListToolsResult{
					Tools: []Tool{
						{
							Name:        "generate-image",
							Description: "Generate an image using Stable Diffusion based on a text prompt",
							InputSchema: toolSchema,
						},
					},
				},
			}

		case "resources/list":
			response = JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result: ListResourcesResult{
					Resources: []Resource{},
				},
			}

		case "prompts/list":
			response = JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result: ListPromptsResult{
					Prompts: []Prompt{},
				},
			}
		case "tools/call":
			params, ok := request.Params.(map[string]interface{})
			if !ok {
				sendError(encoder, request.ID, InvalidParams, "Invalid parameters")
				continue
			}

			toolName, ok := params["name"].(string)
			if !ok || toolName != "generate-image" {
				sendError(encoder, request.ID, MethodNotFound, "Unknown tool")
				continue
			}

			args, ok := params["arguments"].(map[string]interface{})
			if !ok {
				sendError(encoder, request.ID, InvalidParams, "Invalid arguments")
				continue
			}

			prompt, ok := args["prompt"].(string)
			if !ok || prompt == "" {
				sendError(encoder, request.ID, InvalidParams, "Prompt is required")
				continue
			}

			// Get destination path or use default Downloads folder
			var destPath string
			if dest, ok := args["destination"].(string); ok && dest != "" {
				destPath = dest
			} else {
				downloadsDir, err := getDefaultDownloadsPath()
				if err != nil {
					sendError(encoder, request.ID, InternalError, "Could not determine downloads directory")
					continue
				}
				destPath = downloadsDir
			}

			// Generate unique filename
			fullPath := generateUniqueFilename(destPath, prompt)

			// Get dimensions (use defaults if not provided)
			width := 1920
			height := 1080
			if w, ok := args["width"].(float64); ok {
				width = int(w)
			}
			if h, ok := args["height"].(float64); ok {
				height = int(h)
			}

			// Generate image
			imageURL, err := openai.GenerateImage(prompt, width, height)
			if err != nil {
				sendError(encoder, request.ID, InternalError, fmt.Sprintf("Error generating image: %v", err))
				continue
			}

			// Download image
			if err := openai.DownloadImage(imageURL, fullPath); err != nil {
				sendError(encoder, request.ID, InternalError, fmt.Sprintf("Error saving image: %v", err))
				continue
			}

			response = JSONRPCResponse{
				JSONRPC: "2.0",
				ID:      request.ID,
				Result: map[string]interface{}{
					"content": []map[string]interface{}{
						{
							"type": "text",
							"text": fmt.Sprintf("Image generated and saved to: %s", fullPath),
						},
					},
				},
			}

		default:
			sendError(encoder, request.ID, MethodNotFound, "Method not implemented")
			continue
		}

		if response != nil {
			log.Printf("Sending response: %v", PrettyJSON(response))
			sendResponse(encoder, response)
		}
	}

	log.Printf("imagegen-go MCP server out of loop...")
}

func sendError(encoder *json.Encoder, id interface{}, code int, message string) {
	response := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error: &JSONRPCError{
			Code:    code,
			Message: message,
		},
	}
	sendResponse(encoder, response)
}

func sendResponse(encoder *json.Encoder, response interface{}) {

	if err := encoder.Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}
