// filename: pkg/tool/fileapi/tools_file_helpers.go
package fileapi

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// readFileContent reads the entire content of a file specified by path.
// It respects the interpreter's sandbox.
func readFileContent(interp *Interpreter, path string) (string, error) {
	// Use FileAPI getter method for sandboxed access
	absPath, err := interp.FileAPI().ResolvePath(path)	// <<< USE GETTER
	if err != nil {
		if errors.Is(err, ErrPathViolation) {
			return "", err
		}
		return "", fmt.Errorf("failed to resolve path '%s': %w", path, err)
	}

	interp.Logger().Debug("Reading file", "path", absPath)	// Use Logger() getter
	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return "", fmt.Errorf("%w: file not found at '%s'", ErrFileNotFound, path)
		}
		return "", fmt.Errorf("failed to read file '%s': %w", absPath, err)
	}
	return string(contentBytes), nil
}

// writeFileContent writes content to a file specified by path.
// It respects the interpreter's sandbox.
func writeFileContent(interp *Interpreter, path string, content string) error {
	// Use FileAPI getter method for sandboxed access
	absPath, err := interp.FileAPI().ResolvePath(path)	// <<< USE GETTER
	if err != nil {
		if errors.Is(err, ErrPathViolation) {
			return err
		}
		return fmt.Errorf("failed to resolve path '%s': %w", path, err)
	}

	interp.Logger().Debug("Writing file", "path", absPath, "content_length", len(content))	// Use Logger() getter

	// Ensure the target directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory '%s' for writing: %w", dir, err)
	}

	// Write the file - use appropriate permissions
	err = os.WriteFile(absPath, []byte(content), 0640)
	if err != nil {
		return fmt.Errorf("failed to write file '%s': %w", absPath, err)
	}
	interp.Logger().Debug("File written successfully", "path", absPath)	// Use Logger() getter
	return nil
}

// embedFileContent generates embeddings for the file content.
// It respects the interpreter's sandbox and uses the configured LLM client.
func embedFileContent(ctx context.Context, interp *Interpreter, path string) ([]float32, error) {
	interp.Logger().Debug("Requesting embedding for file", "path", path)	// Use getter

	if interp.llmClient == nil {
		return nil, ErrLLMNotConfigured
	}

	content, err := readFileContent(interp, path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s' for embedding: %w", path, err)
	}

	embeddings, err := interp.llmClient.Embed(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("%w: generating embeddings for file '%s': %w", ErrLLMError, path, err)
	}

	interp.Logger().Debug("Embedding generated successfully", "path", path, "vector_length", len(embeddings))	// Use getter
	return embeddings, nil
}

// findFiles walks the directory tree starting from startPath within the sandbox
// and returns a list of absolute file paths matching the criteria.
func findFiles(interp *Interpreter, startPath string) ([]string, error) {
	// Use FileAPI getter method
	absStartPath, err := interp.FileAPI().ResolvePath(startPath)	// <<< USE GETTER
	if err != nil {
		if errors.Is(err, ErrPathViolation) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to resolve start path '%s': %w", startPath, err)
	}

	var files []string
	err = filepath.Walk(absStartPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			interp.Logger().Warn("Error accessing path during walk", "path", path, "error", err)	// Use getter
			return nil
		}

		// Use FileAPI getter method to get sandbox root for comparison
		sandboxRoot := interp.FileAPI().sandboxRoot	// <<< USE GETTER
		if !strings.HasPrefix(path, sandboxRoot) {
			interp.Logger().Error("Path escaped sandbox during walk", "path", path, "sandbox", sandboxRoot)	// Use getter
			return fmt.Errorf("security violation: path '%s' escaped sandbox '%s'", path, sandboxRoot)
		}

		if !info.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory '%s': %w", absStartPath, err)
	}

	return files, nil
}

// scanFileLines reads a file line by line and applies a callback function.
// Stops if the callback returns false. Respects the sandbox.
func scanFileLines(interp *Interpreter, path string, callback func(line string) bool) error {
	// Use FileAPI getter method
	absPath, err := interp.FileAPI().ResolvePath(path)	// <<< USE GETTER
	if err != nil {
		if errors.Is(err, ErrPathViolation) {
			return err
		}
		return fmt.Errorf("failed to resolve path '%s': %w", path, err)
	}

	file, err := os.Open(absPath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("%w: %w", ErrFileNotFound, err)
		}
		return fmt.Errorf("failed to open file '%s' for scanning: %w", absPath, err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if !callback(scanner.Text()) {
			break
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error scanning file '%s': %w", absPath, err)
	}
	return nil
}