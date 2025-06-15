// NeuroScript Version: 0.4.0
// File version: 6
// Purpose: Added toolAppendFile and a shared writeFileHelper to implement FS.Append functionality.
// nlines: 105
// risk_rating: MEDIUM
// filename: pkg/core/tools_fs_write.go
package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// writeFileHelper is a private helper that handles the common logic for both writing and appending files.
func writeFileHelper(interpreter *Interpreter, args []interface{}, append bool) (interface{}, error) {
	if len(args) != 2 {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, fmt.Sprintf("expected 2 arguments (filepath, content), got %d", len(args)), ErrArgumentMismatch)
	}

	relPath, ok := args[0].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("filepath argument must be a string, got %T", args[0]), ErrInvalidArgument)
	}
	content, ok := args[1].(string)
	if !ok {
		return nil, NewRuntimeError(ErrorCodeType, fmt.Sprintf("content argument must be a string, got %T", args[1]), ErrInvalidArgument)
	}

	if relPath == "" {
		return nil, NewRuntimeError(ErrorCodeArgMismatch, "filepath argument cannot be empty", ErrInvalidArgument)
	}

	absPath, secErr := ResolveAndSecurePath(relPath, interpreter.SandboxDir())
	if secErr != nil {
		return nil, secErr
	}

	parentDir := filepath.Dir(absPath)
	if err := os.MkdirAll(parentDir, 0755); err != nil {
		return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("failed to create parent directory for '%s'", relPath), errors.Join(ErrCannotCreateDir, err))
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
		return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("failed to open file '%s'", relPath), errors.Join(ErrIOFailed, err))
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if err != nil {
		return nil, NewRuntimeError(ErrorCodeIOFailed, fmt.Sprintf("failed to write to file '%s'", relPath), errors.Join(ErrIOFailed, err))
	}

	return "OK", nil
}

// toolWriteFile implements FS.Write.
// It creates parent directories if they don't exist and overwrites existing files.
func toolWriteFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return writeFileHelper(interpreter, args, false)
}

// toolAppendFile implements FS.Append.
// It creates parent directories and the file if they don't exist, and appends to existing files.
func toolAppendFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	return writeFileHelper(interpreter, args, true)
}
