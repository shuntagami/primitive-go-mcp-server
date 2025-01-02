package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// File handling functions
func generateUniqueFilename(basePath string, prompt string) string {
	timestamp := time.Now().Format("20060102-150405")
	baseFilename := fmt.Sprintf("%s-%s.png", sanitizeFilename(prompt), timestamp)

	finalPath := filepath.Join(basePath, baseFilename)
	counter := 1
	for {
		if _, err := os.Stat(finalPath); os.IsNotExist(err) {
			break
		}
		baseFilename = fmt.Sprintf("%s-%s-%d.png", sanitizeFilename(prompt), timestamp, counter)
		finalPath = filepath.Join(basePath, baseFilename)
		counter++
	}

	return finalPath
}

func sanitizeFilename(filename string) string {
	words := strings.Fields(filename)
	if len(words) > 4 {
		words = words[:4]
	}
	sanitized := strings.Join(words, "-")

	sanitized = strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '-'
	}, sanitized)

	if len(sanitized) > 50 {
		sanitized = sanitized[:50]
	}

	return strings.Trim(sanitized, "-")
}

func getDefaultDownloadsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(homeDir, "Downloads"), nil
}
