// filename: pkg/core/tools_fs_hash.go
package core

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// toolFileHash calculates the SHA256 hash of a specified file within the sandbox.
// Returns the hex-encoded hash string on success, or an empty string and error on failure.
func toolFileHash(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Validation ---
	if len(args) != 1 {
		return "", fmt.Errorf("%w: expected 1 argument (filepath), got %d", ErrValidationArgCount, len(args))
	}
	filePathRel, ok := args[0].(string)
	if !ok {
		return "", fmt.Errorf("%w: expected argument 1 (filepath) to be a string, got %T", ErrValidationTypeMismatch, args[0])
	}
	if filePathRel == "" {
		return "", fmt.Errorf("%w: filepath cannot be empty", ErrValidationArgValue)
	}

	// --- Path Validation ---
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		if interpreter.logger != nil {
			interpreter.logger.Warn("TOOL FileHash] Interpreter sandboxDir is empty, using default relative path validation.")
		}
		sandboxRoot = "." // Default to current directory if sandbox is not set
	}

	absPath, secErr := SecureFilePath(filePathRel, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("FileHash path error for '%s': %s", filePathRel, secErr.Error())
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: FileHash] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		}
		// Return empty string for script, but propagate original error for Go context
		return "", secErr
	}

	// --- File Hashing ---
	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: FileHash] Attempting to hash validated path: %s (Original Relative: %s, Sandbox: %s)", absPath, filePathRel, sandboxRoot)
	}

	file, err := os.Open(absPath)
	if err != nil {
		errMsg := ""
		if os.IsNotExist(err) {
			errMsg = fmt.Sprintf("FileHash failed: File not found at path '%s'", filePathRel)
		} else {
			errMsg = fmt.Sprintf("FileHash failed to open '%s': %s", filePathRel, err.Error())
		}
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: FileHash] %s", errMsg)
		}
		// Return empty string for script, wrap error for Go context
		return "", fmt.Errorf("%w: opening file '%s': %w", ErrInternalTool, filePathRel, err)
	}
	defer file.Close() // Ensure file is closed

	// Check if it's a directory (we shouldn't hash directories)
	stat, err := file.Stat()
	if err != nil {
		errMsg := fmt.Sprintf("FileHash failed to stat file '%s': %s", filePathRel, err.Error())
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: FileHash] %s", errMsg)
		}
		return "", fmt.Errorf("%w: stating file '%s': %w", ErrInternalTool, filePathRel, err)
	}
	if stat.IsDir() {
		errMsg := fmt.Sprintf("FileHash failed: path '%s' is a directory, not a file", filePathRel)
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: FileHash] %s", errMsg)
		}
		return "", fmt.Errorf("%w: %s", ErrValidationArgValue, errMsg)
	}

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		errMsg := fmt.Sprintf("FileHash failed to read file '%s' for hashing: %s", filePathRel, err.Error())
		if interpreter.logger != nil {
			interpreter.logger.Info("Tool: FileHash] %s", errMsg)
		}
		// Return empty string for script, wrap error for Go context
		return "", fmt.Errorf("%w: reading file '%s' for hashing: %w", ErrInternalTool, filePathRel, err)
	}

	hashBytes := hasher.Sum(nil)
	hashString := fmt.Sprintf("%x", hashBytes)

	if interpreter.logger != nil {
		interpreter.logger.Info("Tool: FileHash] Successfully calculated SHA256 hash for %s: %s", filePathRel, hashString)
	}

	return hashString, nil
}

// registerFsHashTool registers the FileHash tool.
func registerFsHashTools(registry *ToolRegistry) error {
	return registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "FileHash",
			Description: "Calculates the SHA256 hash of a specified file within the sandbox. Returns the hex-encoded hash string.",
			Args: []ArgSpec{
				{Name: "filepath", Type: ArgTypeString, Required: true, Description: "The relative path (within the sandbox) of the file to hash."},
			},
			ReturnType: ArgTypeString, // Returns hash string or empty string on error
		},
		Func: toolFileHash,
	})
}
