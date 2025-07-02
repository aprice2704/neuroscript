// NeuroScript Version: 0.3.1
// File version: 0.0.1
// filename: pkg/tool/fileapi/tools_file_api.go
// nlines: 106
// risk_rating: MEDIUM
package fileapi

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// FileAPI handles sandboxed file system access for the interpreter.
// It ensures that all file operations requested by the script remain
// within a designated root directory (sandbox).
type FileAPI struct {
	sandboxRoot string // The absolute, cleaned path to the sandbox directory.
	logger      interfaces.Logger
}

// NewFileAPI creates a new FileAPI instance.
// It resolves the provided sandboxDir to an absolute path and ensures it exists.
// If sandboxDir is empty or ".", it defaults to the current working directory.
func NewFileAPI(sandboxDir string, logger interfaces.Logger) *FileAPI {
	if logger == nil {
		// Fallback to a basic logger if none provided, although Interpreter should always provide one.
		logger = &logging.NewNoOpLogger{}
		logger.Warn("NewFileAPI created with nil logger, using internal NoOpLogger.")
	}

	effectiveSandboxDir := sandboxDir
	if effectiveSandboxDir == "" || effectiveSandboxDir == "." {
		cwd, err := os.Getwd()
		if err != nil {
			// This is a fatal error during setup.
			panic(fmt.Sprintf("Failed to get current working directory for default sandbox: %v", err))
		}
		effectiveSandboxDir = cwd
		logger.Debug("Using current working directory as sandbox root.", "path", effectiveSandboxDir) // Changed from Info
	}

	// Clean and ensure the sandbox path is absolute.
	absSandbox, err := filepath.Abs(effectiveSandboxDir)
	if err != nil {
		// This is a fatal error during setup.
		panic(fmt.Sprintf("Failed to get absolute path for sandbox directory '%s': %v", effectiveSandboxDir, err))
	}
	absSandbox = filepath.Clean(absSandbox)

	// Check if the sandbox directory exists and is a directory
	info, err := os.Stat(absSandbox)
	if err != nil {
		if os.IsNotExist(err) {
			// Attempt to create the sandbox directory
			logger.Debug("Sandbox directory does not exist, attempting to create.", "path", absSandbox) // Changed from Info
			if mkErr := os.MkdirAll(absSandbox, 0750); mkErr != nil {
				panic(fmt.Sprintf("Failed to create sandbox directory '%s': %v", absSandbox, mkErr))
			}
			logger.Debug("Sandbox directory created successfully.", "path", absSandbox) // Changed from Info
		} else {
			// Other error (e.g., permission denied)
			panic(fmt.Sprintf("Failed to stat sandbox directory '%s': %v", absSandbox, err))
		}
	} else if !info.IsDir() {
		panic(fmt.Sprintf("Sandbox path '%s' exists but is not a directory", absSandbox))
	}

	logger.Debug("FileAPI initialized.", "sandbox_root", absSandbox) // Changed from Info

	return &FileAPI{
		sandboxRoot: absSandbox,
		logger:      logger,
	}
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
		return "", errors.New("internal error: FileAPI sandbox root not initialized")
	}

	// Clean the input path first to handle redundant separators, ".", etc.
	cleanedRelPath := filepath.Clean(relPath)

	// If the cleaned path is absolute, check if it's within the sandbox.
	// Otherwise, join it with the sandbox root.
	var absPath string
	if filepath.IsAbs(cleanedRelPath) {
		// If the user provided an absolute path, it *must* already be inside the sandbox.
		// We still clean it to ensure consistent formatting.
		absPath = cleanedRelPath
		f.logger.Debug("ResolvePath: Received absolute path.", "input", relPath, "cleaned_abs", absPath)
	} else {
		// Join the relative path with the sandbox root.
		// filepath.Join automatically cleans the result.
		absPath = filepath.Join(f.sandboxRoot, cleanedRelPath)
		f.logger.Debug("ResolvePath: Received relative path.", "input", relPath, "joined_abs", absPath)
	}

	// SECURITY CHECK: Ensure the final absolute path is still prefixed
	// by the sandbox root directory. Add a path separator to the root
	// to prevent partial matches (e.g., /sandbox/../sandbox-evil).
	// Note: filepath.Clean removes trailing separators, so we add it back for the check.
	prefix := f.sandboxRoot
	if !strings.HasSuffix(prefix, string(os.PathSeparator)) {
		prefix += string(os.PathSeparator)
	}
	// Also check if the resolved path is *exactly* the sandbox root
	isRoot := absPath == f.sandboxRoot

	if !strings.HasPrefix(absPath, prefix) && !isRoot {
		f.logger.Warn("Path traversal attempt detected!", "requested_path", relPath, "resolved_path", absPath, "sandbox_root", f.sandboxRoot)
		return "", fmt.Errorf("%w: path '%s' resolves outside sandbox '%s'", lang.ErrPathViolation, relPath, f.sandboxRoot)
	}

	f.logger.Debug("Path resolved successfully within sandbox.", "input", relPath, "output", absPath)
	return absPath, nil
}
