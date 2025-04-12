// filename: pkg/core/tools_fs_read.go
package core

import (
	"fmt"
	"os"
	// path/filepath needed only by SecureFilePath, which is in helpers
)

// toolReadFile reads the content of a specified file.
// Assumes path validation/sandboxing is handled by the SecurityLayer before this is called.
// *** MODIFIED: Propagate error from SecureFilePath ***
func toolReadFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// Validation guarantees args[0] is a string
	filePath := args[0].(string)

	// Although security layer validates against sandbox, we resolve here vs CWD
	// for the os.ReadFile call. The security layer prevents reading outside sandbox.
	cwd, errWd := os.Getwd()
	if errWd != nil {
		return nil, fmt.Errorf("TOOL ReadFile failed get CWD: %w", errWd) // Internal error
	}

	// Use SecureFilePath to get the absolute path, primarily for OS compatibility.
	// The security check against sandbox root happened *before* this tool was called.
	absPath, secErr := SecureFilePath(filePath, cwd)
	if secErr != nil {
		// *** Propagate the path violation error ***
		errMsg := fmt.Sprintf("ReadFile path error for '%s': %s", filePath, secErr.Error()) // Log unwrapped
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ReadFile] %s", errMsg)
		}
		// Return error message string for script, but the error itself for Go context
		return errMsg, secErr
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL ReadFile] Reading validated path: %s (Original Relative: %s)", absPath, filePath)
	}

	contentBytes, readErr := os.ReadFile(absPath)
	if readErr != nil {
		errMsg := ""
		if os.IsNotExist(readErr) {
			errMsg = fmt.Sprintf("ReadFile failed: File not found at path '%s'", filePath)
		} else {
			errMsg = fmt.Sprintf("ReadFile failed for '%s': %s", filePath, readErr.Error())
		}
		if interpreter.logger != nil {
			interpreter.logger.Printf("[TOOL ReadFile] %s", errMsg)
		}
		// Return error message string for script, and wrapped internal error for Go
		return errMsg, fmt.Errorf("%w: reading file '%s': %w", ErrInternalTool, filePath, readErr)
	}

	if interpreter.logger != nil {
		interpreter.logger.Printf("[TOOL ReadFile] Read %d bytes successfully from %s", len(contentBytes), filePath)
	}
	// Return file content as string
	return string(contentBytes), nil
}
