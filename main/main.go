package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	var request JSONRPCRequest
	if err := decoder.Decode(&request); err != nil {
		log.Printf("Error decoding request: %v", err)
		sendError(encoder, nil, ParseError, "Failed to parse JSON")
		return
	}

	log.Printf("Received request: %v", PrettyJSON(request))

	if request.JSONRPC != "2.0" {
		sendError(encoder, request.ID, InvalidRequest, "Only JSON-RPC 2.0 is supported")
		return
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
		response = JSONRPCResponse{
			JSONRPC: "2.0",
			ID:      request.ID,
		}
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

	default:
		sendError(encoder, request.ID, MethodNotFound, "Method not implemented")
		return
	}

	log.Printf("Sending response: %v", PrettyJSON(response))
	sendResponse(encoder, response)
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
