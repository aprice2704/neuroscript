// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Corrected lang.NewRuntimeError calls with standard ErrorCodes/Sentinels.
// nlines: 77
// risk_rating: LOW
// filename: pkg/tool/fs/tools_fs_utils.go
package fs

import (
	"errors" // Required for errors.Is
	"fmt"
	"os"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- Tool Implementations (Functions only) ---

// toolLineCountFile counts lines in a specified file.
func toolLineCountFile(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return int64(-1), lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "LineCountFile: expected 1 argument (filepath)", lang.ErrArgumentMismatch)
	}
	filePath, ok := args[0].(string)
	if !ok {
		// Using ErrorCodeType for wrong type, but wrapping ErrInvalidArgument as the specific type mismatch is an invalid argument for this tool.
		return int64(-1), lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("LineCountFile: filepath argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}
	if filePath == "" {
		// Empty path is treated as an invalid argument value.
		return int64(-1), lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "LineCountFile: filepath cannot be empty", lang.ErrInvalidArgument)
	}

	sandboxRoot := interpreter.SandboxDir()
	absPath, secErr := security.SecureFilePath(filePath, sandboxRoot)
	if secErr != nil {
		interpreter.GetLogger().Warn("TOOL LineCountFile] Path validation failed", "path", filePath, "error", secErr, "sandbox_root", sandboxRoot)
		// SecureFilePath returns a RuntimeError already, directly return it.
		// Ensure SecureFilePath wraps appropriate sentinels like ErrPathViolation.
		return int64(-1), secErr
	}

	// interpreter.GetLogger().Debug("Tool: LineCountFile] Attempting to read validated path", "absolute_path", absPath, "original_path", filePath, "sandbox", sandboxRoot)
	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		interpreter.GetLogger().Warn("TOOL LineCountFile] Read error", "path", filePath, "error", readErr)
		if errors.Is(readErr, os.ErrNotExist) {
			// Use the specific ErrorCodeFileNotFound and ErrFileNotFound sentinel
			return int64(-1), lang.NewRuntimeError(lang.ErrorCodeFileNotFound, fmt.Sprintf("LineCountFile: file not found '%s'", filePath), lang.ErrFileNotFound)
		}
		if errors.Is(readErr, os.ErrPermission) {
			// Use the specific ErrorCodePermissionDenied and ErrPermissionDenied sentinel
			return int64(-1), lang.NewRuntimeError(lang.ErrorCodePermissionDenied, fmt.Sprintf("LineCountFile: permission denied for '%s'", filePath), lang.ErrPermissionDenied)
		}
		// For other I/O errors, use ErrorCodeIOFailed and wrap the specific error
		return int64(-1), lang.NewRuntimeError(lang.ErrorCodeIOFailed, fmt.Sprintf("LineCountFile: error reading file '%s'", filePath), errors.Join(lang.ErrIOFailed, readErr))
	}

	content := string(contentBytes)
	if len(content) == 0 {
		// interpreter.GetLogger().Debug("Tool: LineCountFile] Counted 0 lines (empty file)", "file_path", filePath)
		return int64(0), nil
	}

	lineCount := int64(strings.Count(content, "\n"))
	if !strings.HasSuffix(content, "\n") {
		lineCount++
	}

	// interpreter.GetLogger().Debug("Tool: LineCountFile] Counted lines", "count", lineCount, "file_path", filePath)
	return lineCount, nil
}

// toolSanitizeFilename calls the exported helper function SanitizeFilename (from security.go).
func toolSanitizeFilename(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return "", lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "SanitizeFilename: expected 1 argument (name)", lang.ErrArgumentMismatch)
	}
	name, ok := args[0].(string)
	if !ok {
		return "", lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("SanitizeFilename: name argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}

	// SanitizeFilename itself doesn't currently return an error. If it did, we'd handle it here.
	sanitized := security.SanitizeFilename(name)
	// interpreter.GetLogger().Debug("Tool: SanitizeFilename", "input", name, "output", sanitized)
	return sanitized, nil
}
