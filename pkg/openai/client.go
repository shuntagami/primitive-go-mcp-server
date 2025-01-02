package openai

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
	// Determine the closest supported DALL-E size
	size := "1024x1024" // Default
	if width >= 1920 || height >= 1080 {
		size = "1792x1024" // HD aspect ratio
	}

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

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(bodyBytes))
	}

	var result ImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("error decoding response: %v", err)
	}

	if len(result.Data) == 0 || result.Data[0].URL == "" {
		return "", fmt.Errorf("no image URL in response")
	}

	return result.Data[0].URL, nil
}

// DownloadImage downloads the image from URL and saves it to the specified path
func DownloadImage(url, destPath string) error {
	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("error downloading image: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error downloading image: status code %d", resp.StatusCode)
	}

	out, err := os.Create(destPath)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("error saving image: %v", err)
	}

	return nil
}
