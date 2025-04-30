// filename: pkg/core/tools_file_api.go
package core

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aprice2704/neuroscript/pkg/logging"
	// For generating unique IDs if needed by FileAPI state
)

// FileAPI provides sandboxed file system operations for the interpreter.
type FileAPI struct {
	sandboxRoot string
	logger      logging.Logger
	// Potentially add state like open file handles if needed
	// openFiles map[string]*os.File
}

// NewFileAPI creates a new FileAPI instance.
// It requires the absolute path to the sandbox root directory and a logger.
func NewFileAPI(sandboxRoot string, logger logging.Logger) *FileAPI {
	if logger == nil {
		logger = &coreNoOpLogger{}
		// logger.Warn("FileAPI created with nil logger, using internal NoOpLogger.")
	}
	if sandboxRoot == "" {
		// This should be prevented by the Interpreter's SetSandbox logic
		logger.Error("FATAL: FileAPI created with empty sandboxRoot.")
		panic("FATAL: FileAPI requires a non-empty sandbox root directory.")
	}
	// Ensure the path is absolute and clean
	absRoot, err := filepath.Abs(sandboxRoot)
	if err != nil {
		logger.Error("FATAL: Failed to get absolute path for sandbox root", "path", sandboxRoot, "error", err)
		panic(fmt.Sprintf("FATAL: Invalid sandbox root '%s': %v", sandboxRoot, err))
	}

	logger.Info("Initializing FileAPI.", "sandboxRoot", absRoot)
	return &FileAPI{
		sandboxRoot: absRoot,
		logger:      logger,
		// openFiles:   make(map[string]*os.File),
	}
}

// ResolvePath converts a relative path provided by the script into an absolute path
// confined within the sandbox. It returns an error if the path tries to escape.
func (f *FileAPI) ResolvePath(relativePath string) (string, error) {
	if filepath.IsAbs(relativePath) {
		// Forbid absolute paths from the script for security
		return "", fmt.Errorf("absolute paths are not allowed: '%s'", relativePath)
	}

	// Clean the path to prevent tricks like '..' escaping
	cleanedPath := filepath.Clean(relativePath)

	// Check for '..' components that might lead outside the root *after* joining
	// Note: filepath.Join calls Clean internally, but explicit check adds clarity/safety.
	if strings.HasPrefix(cleanedPath, ".."+string(filepath.Separator)) || cleanedPath == ".." {
		return "", fmt.Errorf("path attempts to traverse above sandbox root: '%s'", relativePath)
	}

	// Join with the sandbox root
	absPath := filepath.Join(f.sandboxRoot, cleanedPath)

	// Final check: Ensure the resulting absolute path is still within the sandbox root
	// This protects against symlink tricks or complex '..' scenarios missed by simple checks.
	if !strings.HasPrefix(absPath, f.sandboxRoot) {
		f.logger.Error("Path resolution resulted in escaping sandbox", "relativePath", relativePath, "resolvedPath", absPath, "sandbox", f.sandboxRoot)
		return "", fmt.Errorf("resolved path '%s' is outside sandbox '%s'", absPath, f.sandboxRoot)
	}

	f.logger.Debug("Resolved path", "relative", relativePath, "absolute", absPath)
	return absPath, nil
}

// --- File Operation Methods ---

// Read reads the entire content of a file within the sandbox.
func (f *FileAPI) Read(path string) (string, error) {
	absPath, err := f.ResolvePath(path)
	if err != nil {
		return "", err // Error includes context from ResolvePath
	}

	f.logger.Debug("Reading file", "path", absPath)
	contentBytes, err := os.ReadFile(absPath)
	if err != nil {
		// Check if it's a "not found" error vs. other read error
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: file not found at '%s'", ErrFileNotFound, path) // Use specific error type
		}
		return "", fmt.Errorf("failed to read file '%s': %w", absPath, err)
	}
	return string(contentBytes), nil
}

// Write writes content to a file within the sandbox, overwriting if it exists.
// Creates directories if they don't exist.
func (f *FileAPI) Write(path string, content string) error {
	absPath, err := f.ResolvePath(path)
	if err != nil {
		return err
	}

	f.logger.Debug("Writing file", "path", absPath, "content_length", len(content))

	// Ensure the target directory exists
	dir := filepath.Dir(absPath)
	// Use MkdirAll for robustness - requires appropriate permissions
	if err := os.MkdirAll(dir, 0750); err != nil {
		return fmt.Errorf("failed to create directory '%s' for writing: %w", dir, err)
	}

	// Write the file - use appropriate permissions
	err = os.WriteFile(absPath, []byte(content), 0640)
	if err != nil {
		return fmt.Errorf("failed to write file '%s': %w", absPath, err)
	}
	f.logger.Info("File written successfully", "path", absPath)
	return nil
}

// Delete removes a file or an empty directory within the sandbox.
func (f *FileAPI) Delete(path string) error {
	absPath, err := f.ResolvePath(path)
	if err != nil {
		return err
	}

	// Prevent deleting the sandbox root itself
	if absPath == f.sandboxRoot {
		return fmt.Errorf("cannot delete the sandbox root directory")
	}

	f.logger.Debug("Deleting path", "path", absPath)
	err = os.Remove(absPath) // os.Remove handles both files and empty directories
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("%w: cannot delete, path not found at '%s'", ErrFileNotFound, path)
		}
		// Check if it's a "directory not empty" error
		// This requires checking the specific error type or string, which can be fragile.
		// Example check (may vary across OS):
		// if strings.Contains(err.Error(), "directory not empty") {
		//     return fmt.Errorf("cannot delete directory '%s': it is not empty", path)
		// }
		return fmt.Errorf("failed to delete path '%s': %w", absPath, err)
	}
	f.logger.Info("Path deleted successfully", "path", absPath)
	return nil
}

// Stat returns information about a file or directory within the sandbox.
func (f *FileAPI) Stat(path string) (os.FileInfo, error) {
	absPath, err := f.ResolvePath(path)
	if err != nil {
		return nil, err
	}

	f.logger.Debug("Stating path", "path", absPath)
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: path not found at '%s'", ErrFileNotFound, path)
		}
		return nil, fmt.Errorf("failed to stat path '%s': %w", absPath, err)
	}
	return info, nil
}

// ListDir lists the contents of a directory within the sandbox.
func (f *FileAPI) ListDir(path string) ([]string, error) {
	absPath, err := f.ResolvePath(path)
	if err != nil {
		return nil, err
	}

	f.logger.Debug("Listing directory", "path", absPath)
	info, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("%w: directory not found at '%s'", ErrFileNotFound, path)
		}
		return nil, fmt.Errorf("failed to stat directory '%s': %w", absPath, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("path is not a directory: '%s'", path)
	}

	entries, err := os.ReadDir(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory '%s': %w", absPath, err)
	}

	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	f.logger.Debug("Directory listed successfully", "path", absPath, "entry_count", len(names))
	return names, nil
}

// Embed generates embeddings for a file's content using the interpreter's LLM client.
// This method requires the Interpreter context to access the LLMClient.
func (f *FileAPI) Embed(ctx context.Context, interp *Interpreter, path string) ([]float32, error) {
	f.logger.Debug("Requesting embedding for file", "path", path)

	// CORRECTED: Replace checkGenAIClient with direct nil check
	if interp.llmClient == nil {
		return nil, fmt.Errorf("LLM client not configured in interpreter, cannot generate embeddings")
	}

	// Read file content using the FileAPI's own Read method to ensure sandboxing
	content, err := f.Read(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file '%s' for embedding: %w", path, err)
	}

	// Call the Embed method on the LLMClient interface instance
	embeddings, err := interp.llmClient.Embed(ctx, content)
	if err != nil {
		return nil, fmt.Errorf("LLM Embed call failed for file '%s': %w", path, err)
	}

	f.logger.Debug("Embedding generated successfully", "path", path, "vector_length", len(embeddings))
	return embeddings, nil
}

// --- Sync Related Methods (Placeholder - requires sync logic integration) ---

// GetFileHash calculates the SHA256 hash of a file.
func (f *FileAPI) GetFileHash(path string) (string, error) {
	absPath, err := f.ResolvePath(path)
	if err != nil {
		return "", err
	}
	file, err := os.Open(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", fmt.Errorf("%w: file not found at '%s'", ErrFileNotFound, path)
		}
		return "", fmt.Errorf("failed to open file '%s' for hashing: %w", absPath, err)
	}
	defer file.Close()
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to hash file content for '%s': %w", absPath, err)
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}

// SyncState represents the state needed for synchronization.
type SyncState struct {
	Files map[string]string // Map of relative path -> hash
	// Add other relevant state like last sync time, etc.
	LastSyncTime time.Time
}

// GetSyncState calculates the current state of the sandbox for synchronization.
func (f *FileAPI) GetSyncState() (*SyncState, error) {
	f.logger.Debug("Calculating sandbox sync state...")
	state := &SyncState{
		Files:        make(map[string]string),
		LastSyncTime: time.Now().UTC(), // Record current time
	}

	err := filepath.Walk(f.sandboxRoot, func(absPath string, info os.FileInfo, err error) error {
		if err != nil {
			f.logger.Warn("Error accessing path during sync state calculation", "path", absPath, "error", err)
			return nil // Continue walking if possible
		}
		// Skip the root directory itself
		if absPath == f.sandboxRoot {
			return nil
		}
		// Ensure we are still in sandbox (defense in depth)
		if !strings.HasPrefix(absPath, f.sandboxRoot) {
			return fmt.Errorf("security violation: path '%s' escaped sandbox '%s' during sync state calculation", absPath, f.sandboxRoot)
		}

		if !info.IsDir() {
			// Calculate relative path
			relPath, err := filepath.Rel(f.sandboxRoot, absPath)
			if err != nil {
				// Should not happen if absPath starts with sandboxRoot
				f.logger.Error("Failed to calculate relative path", "absolute", absPath, "root", f.sandboxRoot, "error", err)
				return err // Stop the walk
			}

			// Calculate hash
			hash, err := f.GetFileHash(relPath) // Use relative path for GetFileHash
			if err != nil {
				// Log error but continue if possible? Or fail sync state?
				f.logger.Error("Failed to get hash for file during sync state calculation", "path", relPath, "error", err)
				// Decide whether to skip this file or abort
				return nil // Skip this file
			}
			state.Files[relPath] = hash
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking sandbox for sync state: %w", err)
	}

	f.logger.Debug("Sandbox sync state calculated", "file_count", len(state.Files))
	return state, nil
}

// --- Assumed Definitions ---
// var ErrFileNotFound = errors.New("file not found") // Assumed defined in errors.go
