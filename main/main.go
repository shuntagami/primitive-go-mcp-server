package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	// Read the request
	var request JSONRPCRequest
	if err := decoder.Decode(&request); err != nil {
		log.Printf("Error decoding request: %v", err)
		sendError(encoder, nil, ParseError, "Failed to parse JSON")
		return
	}

	log.Printf("Received request: %v", PrettyJSON(request))

	// Basic validation
	if request.JSONRPC != "2.0" {
		sendError(encoder, request.ID, InvalidRequest, "Only JSON-RPC 2.0 is supported")
		return
	}

	// Handle methods
	switch request.Method {
	case "initialize":
		response := JSONRPCResponse{
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

		log.Printf("Sending response: %v", PrettyJSON(response))
		if err := encoder.Encode(response); err != nil {
			log.Printf("Error encoding response: %v", err)
		}

	default:
		sendError(encoder, request.ID, MethodNotFound, "Method not implemented")
	}
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

	log.Printf("Sending error: %v", PrettyJSON(response))
	if err := encoder.Encode(response); err != nil {
		log.Printf("Error encoding error response: %v", err)
	}
}
