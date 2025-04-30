// filename: pkg/core/tools_file_helpers.go
package core

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// readFileContent reads the entire content of a file specified by path.
// It respects the interpreter's sandbox.
func readFileContent(interp *Interpreter, path string) (string, error) {
	// Use FileAPI for sandboxed access
	absPath, err := interp.FileAPI().ResolvePath(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path '%s': %w", path, err)
	}

	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to read file '%s': %w", absPath, err)
	}
	return string(contentBytes), nil
}

// writeFileContent writes content to a file specified by path.
// It respects the interpreter's sandbox.
func writeFileContent(interp *Interpreter, path string, content string) error {
	// Use FileAPI for sandboxed access
	absPath, err := interp.FileAPI().ResolvePath(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path '%s': %w", path, err)
	}

	// Ensure the directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0750); err != nil { // Use appropriate permissions
		return fmt.Errorf("failed to create directory '%s': %w", dir, err)
	}

	err = os.WriteFile(absPath, []byte(content), 0640) // Use appropriate permissions
	if err != nil {
		return fmt.Errorf("failed to write file '%s': %w", absPath, err)
	}
	return nil
}

// calculateFileHash calculates the SHA256 hash of a file.
// It respects the interpreter's sandbox.
func calculateFileHash(interp *Interpreter, path string) (string, error) {
	absPath, err := interp.FileAPI().ResolvePath(path)
	if err != nil {
		return "", fmt.Errorf("failed to resolve path '%s': %w", path, err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return "", fmt.Errorf("failed to open file '%s' for hashing: %w", absPath, err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file content for '%s': %w", absPath, err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// embedFileContent generates embeddings for the file content.
// It respects the interpreter's sandbox and uses the configured LLM client.
func embedFileContent(ctx context.Context, interp *Interpreter, path string) ([]float32, error) {
	if interp.llmClient == nil {
		return nil, fmt.Errorf("LLM client not configured in interpreter, cannot generate embeddings")
	}

	content, err := readFileContent(interp, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s' for embedding: %w", path, err)
	}

	// Use the Embed method directly from the LLMClient interface
	// REMOVED: Access to interpreter.llmClient.Client
	embeddings, err := interp.llmClient.Embed(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embeddings for file '%s': %w", path, err)
	}

	return embeddings, nil
}

// findFiles walks the directory tree starting from startPath within the sandbox
// and returns a list of absolute file paths matching the criteria.
// TODO: Add matching criteria (e.g., glob patterns, filters).
func findFiles(interp *Interpreter, startPath string) ([]string, error) {
	absStartPath, err := interp.FileAPI().ResolvePath(startPath)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve start path '%s': %w", startPath, err)
	}

	var files []string
	err = filepath.Walk(absStartPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Log the error but continue walking if possible
			interp.Logger().Warn("Error accessing path during walk", "path", path, "error", err)
			// Decide whether to skip subtree (e.g., permission denied)
			// return filepath.SkipDir // Example
			return nil // Continue walking other parts
		}

		// Ensure we stay within the sandbox (redundant if ResolvePath is robust, but good defense)
		if !strings.HasPrefix(path, interp.FileAPI().sandboxRoot) {
			interp.Logger().Error("Path escaped sandbox during walk", "path", path, "sandbox", interp.FileAPI().sandboxRoot)
			return fmt.Errorf("security violation: path '%s' escaped sandbox '%s'", path, interp.FileAPI().sandboxRoot)
		}

		if !info.IsDir() {
			// TODO: Add filtering logic here based on patterns/criteria
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		// This error is likely the security violation from above or a critical walk error
		return nil, fmt.Errorf("error walking directory '%s': %w", absStartPath, err)
	}

	return files, nil
}

// scanFileLines reads a file line by line and applies a callback function.
// Stops if the callback returns false. Respects the sandbox.
func scanFileLines(interp *Interpreter, path string, callback func(line string) bool) error {
	absPath, err := interp.FileAPI().ResolvePath(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path '%s': %w", path, err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		return fmt.Errorf("failed to open file '%s' for scanning: %w", absPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if !callback(scanner.Text()) {
			break // Callback requested stop
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file '%s': %w", absPath, err)
	}
	return nil
}
