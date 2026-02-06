package config

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
)

// Config holds all service configuration
type Config struct {
	ChromiumPath string
	ServerPort   string
	MaxBrowsers  int
}

// Function to load the configuration
func Load() (*Config, error) {
	// Find Chromium binary
	chromiumPath, err := findChromium()
	if err != nil {
		return nil, err
	}

	// Read SERVER_PORT from environment variable, default to "8080"
	serverPort := os.Getenv("SERVER_PORT")
	if serverPort == "" {
		serverPort = "8080"
	}

	// Read MAX_BROWSERS from environment variable, default to 5
	maxBrowsers := os.Getenv("MAX_BROWSERS")
	if maxBrowsers == "" {
		maxBrowsers = "5"
	}

	maxBrowsersInt, err := strconv.Atoi(maxBrowsers)
	if err != nil {
		return nil, err
	}

	// Return the config
	return &Config{
		ChromiumPath: chromiumPath,
		ServerPort:   serverPort,
		MaxBrowsers:  maxBrowsersInt,
	}, nil
}

// Function to find the Chromium binary path
func findChromium() (string, error) {
	
	// Check if CHROMIUM_PATH environment variable is set
	customPath := os.Getenv("CHROMIUM_PATH")
	if customPath != "" {
		
		// Validate the custom path exists
		if !fileExists(customPath) {
			return "", fmt.Errorf("chromium binary not found at path: %s", customPath)
		}

		// Validate the custom path is executable
		if !isExecutable(customPath) {
			return "", fmt.Errorf("chromium binary found but not executable: %s", customPath)
		}
		return customPath, nil
	}

	// Get current operating system
	currentOS := runtime.GOOS

	// Get common paths for this OS
	paths := getChromiumPaths(currentOS)

	// Search through common paths
	for _, path := range paths {
		if fileExists(path) && isExecutable(path) {
			return path, nil
		}
	}

	// If we get here, chromium wasn't found anywhere
	return "", fmt.Errorf("chromium not found in common paths for %s, set CHROMIUM_PATH environment variable", currentOS)
}

// getChromiumPaths returns common Chromium installation paths based on OS.
func getChromiumPaths(operatingSystem string) []string {
	// macOS paths
	if operatingSystem == "darwin" {
		return []string{
			"/Applications/Chromium.app/Contents/MacOS/Chromium",
			"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
		}
	}

	// Linux paths
	if operatingSystem == "linux" {
		return []string{
			"/usr/bin/chromium-browser",
			"/usr/bin/chromium",
			"/snap/bin/chromium",
		}
	}

	// TODO: Add Windows paths later

	// Unsupported OS
	return []string{}
}