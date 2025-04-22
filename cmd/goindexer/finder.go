package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// findRepoPaths tries to find the repo root (containing go.mod) and the module path.
// It searches upwards from the provided startDir.
func findRepoPaths(startDir string) (rootPath string, modulePath string, err error) {
	dir, err := filepath.Abs(startDir)
	if err != nil {
		return "", "", fmt.Errorf("failed to get absolute path for starting directory %q: %w", startDir, err)
	}
	dir = filepath.Clean(dir)

	info, err := os.Stat(dir)
	if err == nil && !info.IsDir() {
		dir = filepath.Dir(dir)
	} else if err != nil && !os.IsNotExist(err) {
		return "", "", fmt.Errorf("failed to stat starting path %q: %w", startDir, err)
	}

	log.Printf("Starting go.mod search upwards from: %s", dir)
	checkedPaths := []string{} // Keep track to avoid infinite loops on weird filesystems

	for {
		// Prevent potential infinite loops
		for _, checked := range checkedPaths {
			if checked == dir {
				return "", "", fmt.Errorf("filesystem loop detected or traversed too high during go.mod search near %s", dir)
			}
		}
		checkedPaths = append(checkedPaths, dir)

		goModPath := filepath.Join(dir, "go.mod")
		log.Printf("  Checking: %s", goModPath)
		_, statErr := os.Stat(goModPath)

		if statErr == nil {
			// Found go.mod
			rootPath = dir
			modPath, parseErr := parseModulePath(goModPath)
			if parseErr != nil {
				// Return the found rootPath, but a specific error about parsing THAT file
				return rootPath, "", fmt.Errorf("found '%s' but failed to parse module path: %w", goModPath, parseErr)
			}
			modulePath = modPath
			log.Printf("  Found go.mod at: %s (Module: %s)", rootPath, modulePath)
			return rootPath, modulePath, nil // Success
		}

		if !errors.Is(statErr, os.ErrNotExist) {
			// Error other than file not existing encountered during search
			return "", "", fmt.Errorf("error checking for go.mod at %s: %w", goModPath, statErr)
		}

		// Check if we've reached the root
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			// Reached filesystem root without finding go.mod
			return "", "", fmt.Errorf("go.mod not found in %s or any parent directory", startDir)
		}
		// Go up one level
		dir = parentDir
	}
}

// parseModulePath reads a go.mod file and extracts the module path.
// More robust parsing attempt.
func parseModulePath(goModPath string) (string, error) {
	file, err := os.Open(goModPath)
	if err != nil {
		return "", fmt.Errorf("failed to open %s: %w", goModPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := scanner.Text()
		// Remove comments first
		if idx := strings.Index(line, "//"); idx != -1 {
			line = line[:idx]
		}
		// Trim leading/trailing whitespace AFTER removing comments
		line = strings.TrimSpace(line)

		// Check for module directive
		if strings.HasPrefix(line, "module") {
			fields := strings.Fields(line)
			if len(fields) >= 2 {
				// Module path is the second field, potentially handling paths with spaces if quoted (rare)
				modulePath := fields[1]
				// Basic trim of surrounding quotes if present (unlikely for module)
				modulePath = strings.Trim(modulePath, `"`)
				if modulePath == "" { // Handle case like "module \n // comment"
					continue
				}
				log.Printf("    Parsed module path '%s' from line %d of %s", modulePath, lineNumber, goModPath)
				return modulePath, nil // Found module path
			} else {
				// Found 'module' keyword but line format is wrong
				return "", fmt.Errorf("malformed module line (line %d) in %s: %q", lineNumber, goModPath, scanner.Text())
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading %s: %w", goModPath, err)
	}

	// Scanned whole file, no module line found
	return "", fmt.Errorf("module directive not found in %s", goModPath)
}
