package main

import (
	"encoding/json"
	"log"
	"os"
)

func main() {
	decoder := json.NewDecoder(os.Stdin)
	encoder := json.NewEncoder(os.Stdout)

	var req map[string]interface{}
	if err := decoder.Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		return
	}

	log.Printf("Received request: %v", PrettyJSON(req))

	method, hasMethod := req["method"].(string)
	id, hasId := req["id"]
	var response map[string]interface{}

	if hasMethod && hasId {
		switch method {
		case "initialize":
			response = map[string]interface{}{
				"jsonrpc": "2.0",
				"id":      id,
				"result": map[string]interface{}{
					"protocolVersion": "2024-11-05",
					"serverInfo": map[string]interface{}{
						"name":    "imagegen-go",
						"version": "1.0.0",
					},
					"capabilities": map[string]interface{}{
						"tools": map[string]interface{}{},
					},
				},
			}
		}

	}

	if response != nil {
		log.Printf("Sending response: %v", PrettyJSON(response))

		err := encoder.Encode(response)

		if err != nil {
			log.Printf("Error encoding response: %v", err)
		}
	}

}
