// NeuroScript Version: 0.3.1
// File version: 0.1.0 // Add ReadFile method.
// nlines: 120 // Approximate
// risk_rating: MEDIUM
// filename: pkg/core/file_api.go
package core

import (
	"errors"
	"fmt"
	"io/fs" // Import fs for file modes
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/logging"
)

// FileAPI handles sandboxed file system access for the interpreter.
// It ensures that all file operations requested by the script remain
// within a designated root directory (sandbox).
type FileAPI struct {
	sandboxRoot string // The absolute, cleaned path to the sandbox directory.
	logger      logging.Logger
}

// NewFileAPI creates a new FileAPI instance.
// It resolves the provided sandboxDir to an absolute path and ensures it exists.
// If sandboxDir is empty or ".", it defaults to the current working directory.
func NewFileAPI(sandboxDir string, logger logging.Logger) (*FileAPI, error) {
	if logger == nil {
		// Cannot proceed without a logger, return error or panic? Returning error is safer.
		// return nil, errors.New("cannot create FileAPI with nil logger")
		// For now, use NoOpLogger as fallback but log loudly
		logger = &coreNoOpLogger{}
		logger.Error("NewFileAPI created with nil logger, using internal NoOpLogger. This is not recommended.")
	}

	effectiveSandboxDir := sandboxDir
	if effectiveSandboxDir == "" || effectiveSandboxDir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			logger.Error("Failed to get current working directory for default sandbox.", "error", err)
			return nil, fmt.Errorf("failed to get current working directory for default sandbox: %w", err)
		}
		effectiveSandboxDir = cwd
		logger.Info("Using current working directory as sandbox root.", "path", effectiveSandboxDir)
	}

	// Clean and ensure the sandbox path is absolute.
	absSandbox, err := filepath.Abs(effectiveSandboxDir)
	if err != nil {
		logger.Error("Failed to get absolute path for sandbox directory.", "path", effectiveSandboxDir, "error", err)
		return nil, fmt.Errorf("failed to get absolute path for sandbox directory '%s': %w", effectiveSandboxDir, err)
	}
	absSandbox = filepath.Clean(absSandbox)

	// Check if the sandbox directory exists and is a directory
	info, err := os.Stat(absSandbox)
	if err != nil {
		if os.IsNotExist(err) {
			logger.Info("Sandbox directory does not exist, attempting to create.", "path", absSandbox)
			if mkErr := os.MkdirAll(absSandbox, 0750); mkErr != nil { // Use appropriate permissions
				logger.Error("Failed to create sandbox directory.", "path", absSandbox, "error", mkErr)
				return nil, fmt.Errorf("failed to create sandbox directory '%s': %w", absSandbox, mkErr)
			}
			logger.Info("Sandbox directory created successfully.", "path", absSandbox)
		} else {
			logger.Error("Failed to stat sandbox directory.", "path", absSandbox, "error", err)
			return nil, fmt.Errorf("failed to stat sandbox directory '%s': %w", absSandbox, err)
		}
	} else if !info.IsDir() {
		logger.Error("Sandbox path exists but is not a directory.", "path", absSandbox)
		return nil, fmt.Errorf("sandbox path '%s' exists but is not a directory", absSandbox)
	}

	logger.Info("FileAPI initialized.", "sandbox_root", absSandbox)

	return &FileAPI{
		sandboxRoot: absSandbox,
		logger:      logger,
	}, nil // Return nil error on success
}

// ResolvePath takes a relative or absolute path provided by the user/script
// and resolves it to an absolute path strictly within the sandbox root.
// It performs security checks to prevent directory traversal attacks.
// Returns the absolute, cleaned path within the sandbox, or an error
// (specifically ErrPathViolation if the path tries to escape the sandbox).
func (f *FileAPI) ResolvePath(relPath string) (string, error) {
	if f == nil {
		return "", errors.New("FileAPI receiver is nil") // Should not happen
	}
	if f.sandboxRoot == "" {
		f.logger.Error("ResolvePath called but sandboxRoot is empty!")
		return "", NewRuntimeError(ErrorCodeInternal, "internal error: FileAPI sandbox root not initialized", ErrInternal)
	}
	if relPath == "" {
		// Treat empty path as request for sandbox root itself? Or error?
		// Error seems safer to avoid ambiguity.
		return "", NewRuntimeError(ErrorCodeArgMismatch, "path cannot be empty", ErrInvalidArgument)
	}

	cleanedRelPath := filepath.Clean(relPath)

	var absPath string
	if filepath.IsAbs(cleanedRelPath) {
		absPath = cleanedRelPath
		f.logger.Debug("ResolvePath: Received absolute path.", "input", relPath, "cleaned_abs", absPath)
	} else {
		absPath = filepath.Join(f.sandboxRoot, cleanedRelPath)
		f.logger.Debug("ResolvePath: Received relative path.", "input", relPath, "joined_abs", absPath)
	}

	prefix := f.sandboxRoot
	if !strings.HasSuffix(prefix, string(os.PathSeparator)) {
		prefix += string(os.PathSeparator)
	}
	isRootOrExact := absPath == f.sandboxRoot

	// Security Check: Path must be sandbox root itself or within the sandbox root prefix.
	if !strings.HasPrefix(absPath, prefix) && !isRootOrExact {
		f.logger.Warn("Path traversal attempt detected!", "requested_path", relPath, "resolved_path", absPath, "sandbox_root", f.sandboxRoot)
		// Return error using standard NeuroScript error structure
		return "", NewRuntimeError(ErrorCodeSecurity,
			fmt.Sprintf("path '%s' resolves outside sandbox '%s'", relPath, f.sandboxRoot),
			ErrPathViolation, // Use the specific sentinel error
		)
	}

	f.logger.Debug("Path resolved successfully within sandbox.", "input", relPath, "output", absPath)
	return absPath, nil
}

// ReadFile reads the content of a file within the sandbox.
// Takes a relative path, resolves it securely, and reads the file.
func (f *FileAPI) ReadFile(relPath string) ([]byte, error) {
	if f == nil {
		return nil, errors.New("FileAPI receiver is nil")
	}
	f.logger.Debug("ReadFile requested.", "relative_path", relPath)

	absPath, err := f.ResolvePath(relPath)
	if err != nil {
		// ResolvePath already returns a RuntimeError with appropriate sentinel
		f.logger.Error("ReadFile failed during path resolution.", "relative_path", relPath, "error", err)
		return nil, err
	}

	// Check if it's actually a file
	info, statErr := os.Stat(absPath)
	if statErr != nil {
		sentinel := ErrIOFailed
		ec := ErrorCodeIOFailed
		if os.IsNotExist(statErr) {
			sentinel = ErrNotFound
			ec = ErrorCodeFileNotFound
		} else if os.IsPermission(statErr) {
			sentinel = ErrPermissionDenied
			ec = ErrorCodePermissionDenied
		}
		errMsg := fmt.Sprintf("cannot stat path '%s' before reading: %v", relPath, statErr)
		f.logger.Error(errMsg, "absolute_path", absPath)
		return nil, NewRuntimeError(ec, errMsg, errors.Join(sentinel, statErr))
	}
	if info.IsDir() {
		errMsg := fmt.Sprintf("cannot read: path '%s' is a directory", relPath)
		f.logger.Error(errMsg, "absolute_path", absPath)
		return nil, NewRuntimeError(ErrorCodePathTypeMismatch, errMsg, ErrPathNotFile) // Specific error for type mismatch
	}

	// Read the file content
	content, readErr := os.ReadFile(absPath)
	if readErr != nil {
		sentinel := ErrIOFailed
		ec := ErrorCodeIOFailed
		if os.IsPermission(readErr) { // Check specific errors if needed
			sentinel = ErrPermissionDenied
			ec = ErrorCodePermissionDenied
		}
		errMsg := fmt.Sprintf("failed to read file '%s': %v", relPath, readErr)
		f.logger.Error(errMsg, "absolute_path", absPath)
		return nil, NewRuntimeError(ec, errMsg, errors.Join(sentinel, readErr))
	}

	f.logger.Debug("ReadFile successful.", "relative_path", relPath, "bytes_read", len(content))
	return content, nil
}

// Add other FileAPI methods here as needed (WriteFile, Stat, ListDir, Mkdir, Delete, Move, WalkDir...)
// Example placeholder for Stat:
func (f *FileAPI) Stat(relPath string) (fs.FileInfo, error) {
	if f == nil {
		return nil, errors.New("FileAPI receiver is nil")
	}
	absPath, err := f.ResolvePath(relPath)
	if err != nil {
		return nil, err
	}

	info, statErr := os.Stat(absPath)
	if statErr != nil {
		sentinel := ErrIOFailed
		ec := ErrorCodeIOFailed
		if os.IsNotExist(statErr) {
			sentinel = ErrNotFound
			ec = ErrorCodeFileNotFound
		} else if os.IsPermission(statErr) {
			sentinel = ErrPermissionDenied
			ec = ErrorCodePermissionDenied
		}
		errMsg := fmt.Sprintf("cannot stat path '%s': %v", relPath, statErr)
		f.logger.Error(errMsg, "absolute_path", absPath)
		return nil, NewRuntimeError(ec, errMsg, errors.Join(sentinel, statErr))
	}
	return info, nil
}
