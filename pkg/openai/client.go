package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

// ImageRequest represents the request body for image generation
type ImageRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	N      int    `json:"n"`
	Size   string `json:"size"`
}

// ImageResponse represents the API response structure
type ImageResponse struct {
	Data []struct {
		URL string `json:"url"`
	} `json:"data"`
}

// GenerateImage calls OpenAI API to generate an image
func GenerateImage(prompt string, width, height int) (string, error) {
	log.Printf("GenerateImage called with prompt: %s, dimensions: %dx%d", prompt, width, height)

	// Determine the closest supported DALL-E size
	size := "1024x1024" // Default
	if width >= 1920 || height >= 1080 {
		size = "1792x1024" // HD aspect ratio
	}
	log.Printf("Selected DALL-E size: %s", size)

	reqBody := ImageRequest{
		Model:  "dall-e-3",
		Prompt: prompt,
		N:      1,
		Size:   size,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("error marshaling request: %v", err)
	}

	// Add timeout to the request
	client := &http.Client{
		Timeout: 60 * time.Second, // Set timeout to 60 seconds
	}

	req, err := http.NewRequest("POST", "https://api.openai.com/v1/images/generations",
		bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("error creating request: %v", err)
	}

	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)

	log.Printf("Sending request to OpenAI API...")
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("OpenAI API error response: %s", string(bodyBytes))
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result ImageResponse
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		log.Printf("Error decoding response body: %s", string(bodyBytes))
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if len(result.Data) == 0 || result.Data[0].URL == "" {
		return "", fmt.Errorf("no image URL in response")
	}

	log.Printf("Successfully received image URL from OpenAI")
	return result.Data[0].URL, nil
}

func DownloadImage(url, destPath string) error {
	log.Printf("Starting download of image from: %s", url)

	client := &http.Client{
		Timeout: 30 * time.Second, // Set timeout for download
	}

	resp, err := client.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading image: status code %d", resp.StatusCode)
	}

	log.Printf("Image download successful, creating file: %s", destPath)
	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer out.Close()

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error saving image: %v", err)
	}
	log.Printf("Successfully wrote %d bytes to file", written)

	return nil
}
