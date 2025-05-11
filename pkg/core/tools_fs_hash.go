// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Corrected NewRuntimeError calls with standard ErrorCodes/Sentinels.
// nlines: 89
// risk_rating: LOW
// filename: pkg/core/tools_fs_hash.go
package core

import (
	"crypto/sha256"
	"errors" // Required for errors.Is, errors.Join
	"fmt"
	"io"
	"os"
)

// toolFileHash calculates the SHA256 hash of a specified file within the sandbox.
// Returns the hex-encoded hash string on success, or an empty string and error on failure.
// Implements the FileHash tool.
func toolFileHash(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// --- Argument Validation ---
	if len(args) != 1 {
		return "", NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("FileHash: expected 1 argument (filepath), got %d", len(args)), ErrArgumentMismatch)
	}
	filePathRel, ok := args[0].(string)
	if !ok {
		return "", NewRuntimeError(ErrorCodeType, fmt.Sprintf("FileHash: filepath argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	if filePathRel == "" {
		return "", NewRuntimeError(ErrorCodeArgMismatch, "FileHash: filepath cannot be empty", ErrInvalidArgument)
	}

	// --- Sandbox Check ---
	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		interpreter.Logger().Error("Tool: FileHash] Interpreter sandboxDir is empty, cannot proceed.")
		return "", NewRuntimeError(ErrorCodeConfiguration, "FileHash: interpreter sandbox directory is not set", ErrConfiguration)
	}

	// --- Path Validation ---
	absPath, secErr := SecureFilePath(filePathRel, sandboxRoot)
	if secErr != nil {
		errMsg := fmt.Sprintf("FileHash: path security error for '%s': %v", filePathRel, secErr)
		interpreter.Logger().Debug("Tool: FileHash] %s (Sandbox Root: %s)", errMsg, sandboxRoot)
		// Return the RuntimeError from SecureFilePath directly
		return "", secErr
	}

	// --- File Hashing ---
	interpreter.Logger().Debug("Tool: FileHash attempting to hash validated path", "validated_path", absPath, "original_relative_path", filePathRel, "sandbox_root", sandboxRoot)

	file, openErr := os.Open(absPath)
	if openErr != nil {
		errMsg := ""
		if errors.Is(openErr, os.ErrNotExist) {
			errMsg = fmt.Sprintf("FileHash: file not found at path '%s'", filePathRel)
			interpreter.Logger().Debug("Tool: FileHash] %s", errMsg)
			return "", NewRuntimeError(ErrorCodeFileNotFound, errMsg, ErrFileNotFound)
		}
		if errors.Is(openErr, os.ErrPermission) {
			errMsg = fmt.Sprintf("FileHash: permission denied opening file '%s'", filePathRel)
			interpreter.Logger().Warn("Tool: FileHash] %s", errMsg)
			return "", NewRuntimeError(ErrorCodePermissionDenied, errMsg, ErrPermissionDenied)
		}
		// Other open errors
		errMsg = fmt.Sprintf("FileHash: failed to open file '%s'", filePathRel)
		interpreter.Logger().Error("Tool: FileHash] %s: %v", errMsg, openErr)
		return "", NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, openErr))
	}
	defer file.Close() // Ensure file is closed

	// Check if it's a directory
	stat, statErr := file.Stat()
	if statErr != nil {
		errMsg := fmt.Sprintf("FileHash: failed to stat opened file '%s'", filePathRel)
		interpreter.Logger().Error("Tool: FileHash] %s: %v", errMsg, statErr)
		// Use ErrorCodeIOFailed as stat after successful open should ideally not fail often without I/O issues
		return "", NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, statErr))
	}
	if stat.IsDir() {
		errMsg := fmt.Sprintf("FileHash: path '%s' is a directory, not a file", filePathRel)
		interpreter.Logger().Debug("Tool: FileHash] %s", errMsg)
		// Use ErrorCodePathTypeMismatch and ErrPathNotFile sentinel
		return "", NewRuntimeError(ErrorCodePathTypeMismatch, errMsg, ErrPathNotFile)
	}

	// Hash the file content
	hasher := sha256.New()
	_, copyErr := io.Copy(hasher, file)
	if copyErr != nil {
		errMsg := fmt.Sprintf("FileHash: failed to read file '%s' for hashing", filePathRel)
		interpreter.Logger().Error("Tool: FileHash] %s: %v", errMsg, copyErr)
		// Use ErrorCodeIOFailed for copy errors
		return "", NewRuntimeError(ErrorCodeIOFailed, errMsg, errors.Join(ErrIOFailed, copyErr))
	}

	hashBytes := hasher.Sum(nil)
	hashString := fmt.Sprintf("%x", hashBytes)

	interpreter.Logger().Debug("Tool: FileHash] Successfully calculated SHA256 hash", "file_path", filePathRel, "hash", hashString)
	return hashString, nil
}
