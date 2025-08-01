// NeuroScript Version: 0.4.0
// File version: 7
// Purpose: Added toolAppendFile and a shared writeFileHelper to implement FS.Append functionality. Corrected error handling for writing to a directory.
// nlines: 105
// risk_rating: MEDIUM
// filename: pkg/tool/fs/tools_fs_write.go
package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/security"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// writeFileHelper is a private helper that handles the common logic for both writing and appending files.
func writeFileHelper(interpreter tool.Runtime, args []interface{}, append bool) (interface{}, error) {
	if len(args) != 2 {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, fmt.Sprintf("expected 2 arguments (filepath, content), got %d", len(args)), lang.ErrArgumentMismatch)
	}

	relPath, ok := args[0].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("filepath argument must be a string, got %T", args[0]), lang.ErrInvalidArgument)
	}
	content, ok := args[1].(string)
	if !ok {
		return nil, lang.NewRuntimeError(lang.ErrorCodeType, fmt.Sprintf("content argument must be a string, got %T", args[1]), lang.ErrInvalidArgument)
	}

	if relPath == "" {
		return nil, lang.NewRuntimeError(lang.ErrorCodeArgMismatch, "filepath argument cannot be empty", lang.ErrInvalidArgument)
	}

	absPath, secErr := security.ResolveAndSecurePath(relPath, interpreter.SandboxDir())
	if secErr != nil {
		return nil, secErr
	}

	// Check if the path is a directory before trying to open it for writing
	info, statErr := os.Stat(absPath)
	if statErr == nil && info.IsDir() {
		return nil, lang.NewRuntimeError(lang.ErrorCodePathTypeMismatch, fmt.Sprintf("path '%s' is a directory, not a file", relPath), lang.ErrPathNotFile)
	}

	parentDir := filepath.Dir(absPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, fmt.Sprintf("failed to create parent directory for '%s'", relPath), errors.Join(lang.ErrCannotCreateDir, err))
	}

	var file *os.File
	var err error

	openFlags := os.O_WRONLY | os.O_CREATE
	if append {
		openFlags |= os.O_APPEND
	} else {
		openFlags |= os.O_TRUNC // Truncate the file if we are overwriting
	}

	file, err = os.OpenFile(absPath, openFlags, 0644)
	if err != nil {
		// This will now correctly handle the "is a directory" error from the OS
		if errors.Is(err, os.ErrExist) || strings.Contains(err.Error(), "is a directory") {
			return nil, lang.NewRuntimeError(lang.ErrorCodePathTypeMismatch, fmt.Sprintf("path '%s' is a directory, not a file", relPath), lang.ErrPathNotFile)
		}
		return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, fmt.Sprintf("failed to open file '%s'", relPath), errors.Join(lang.ErrIOFailed, err))
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return nil, lang.NewRuntimeError(lang.ErrorCodeIOFailed, fmt.Sprintf("failed to write to file '%s'", relPath), errors.Join(lang.ErrIOFailed, err))
	}

	return "OK", nil
}

// toolWriteFile implements FS.Write.
// It creates parent directories if they don't exist and overwrites existing files.
var toolWriteFile tool.ToolFunc = func(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	return writeFileHelper(interpreter, args, false)
}

// toolAppendFile implements FS.Append.
// It creates parent directories and the file if they don't exist, and appends to existing files.
func toolAppendFile(interpreter tool.Runtime, args []interface{}) (interface{}, error) {
	return writeFileHelper(interpreter, args, true)
}
