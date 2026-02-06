package config

import "os"

// fileExists checks if a file exists at the given path.
func fileExists(path string) bool {
	// Get file info
	_, err := os.Stat(path)

	// If no error, file exists
	if err == nil {
		return true
	}

	// If error is "file not found", return false
	if os.IsNotExist(err) {
		return false
	}

	// For other errors (permissions, etc), assume file doesn't exist
	return false
}

// isExecutable checks if a file has executable permissions.
func isExecutable(path string) bool {
	// Get file info
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if any execute bit is set (owner, group, or other)
	// 0111 in binary checks all three execute permission bits
	mode := info.Mode()
	return mode&0111 != 0
}