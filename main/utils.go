package main

import (
	"fmt"
	"os"
	"path/filepath"
	"math/rand"
	"strings"
	"time"
	"log"
)

func isValidPath(path string) bool {
	// Check if path is absolute
	if !filepath.IsAbs(path) {
		return false
	}

	// Check if path contains .. or .
	if strings.Contains(path, "..") || path == "." {
		return false
	}

	// Check if path is root
	if path == "/" || path == "\\" {
		return false
	}

	// Try to stat the directory
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Verify it's a directory
	return info.IsDir()
}

func generateUniqueFilename(destPath string, prompt string) (string, error) {
	log.Printf("Generating unique filename for path: %s", destPath)

	// Split path into directory and filename
	destDir := filepath.Dir(destPath)
	origFilename := filepath.Base(destPath)

	// If no extension provided or wrong extension, append .webp
	if !strings.HasSuffix(strings.ToLower(origFilename), ".webp") {
		origFilename = strings.TrimSuffix(origFilename, filepath.Ext(origFilename)) + ".webp"
	}

	// Validate destination directory
	if !isValidPath(destDir) {
		defaultPath := os.Getenv("DEFAULT_DOWNLOAD_PATH")
		if defaultPath == "" {
			// If env var not set, use user's Downloads directory
			homeDir, err := os.UserHomeDir()
			if err != nil {
				return "", fmt.Errorf("failed to get user home directory: %v", err)
			}
			defaultPath = filepath.Join(homeDir, "Downloads")
		}
		
		log.Printf("Invalid destination directory '%s', using default: %s", destDir, defaultPath)
		destDir = defaultPath
		
		// Ensure default download directory exists
		if err := os.MkdirAll(destDir, 0755); err != nil {
			return "", fmt.Errorf("failed to create default download directory: %v", err)
		}
	}

	// First try with original filename
	finalPath := filepath.Join(destDir, origFilename)
	if _, err := os.Stat(finalPath); os.IsNotExist(err) {
		return finalPath, nil
	}

	// If file exists, try with random numbers
	ext := filepath.Ext(origFilename)
	nameWithoutExt := strings.TrimSuffix(origFilename, ext)
	
	// Initialize random number generator with current time
	rand.Seed(time.Now().UnixNano())
	
	// Try up to 100 times
	for i := 0; i < 100; i++ {
		// Generate a random number between 1000 and 9999
		randomNum := rand.Intn(9000) + 1000
		newFilename := fmt.Sprintf("%s-%d%s", nameWithoutExt, randomNum, ext)
		finalPath = filepath.Join(destDir, newFilename)
		
		if _, err := os.Stat(finalPath); os.IsNotExist(err) {
			return finalPath, nil
		}
	}

	return "", fmt.Errorf("could not generate unique filename after 100 attempts")
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
