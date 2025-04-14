// filename: pkg/core/errors.go
package core

import "errors"

// --- Core Validation Errors ---
var (
	ErrValidationRequiredArgNil = errors.New("required argument is nil")
	ErrValidationTypeMismatch   = errors.New("argument type mismatch")
	ErrValidationArgCount       = errors.New("incorrect argument count")
)

// --- Core Tool Execution Errors ---
var (
	ErrListIndexOutOfBounds     = errors.New("list index out of bounds")
	ErrListCannotSortMixedTypes = errors.New("cannot sort list with mixed or non-sortable types")
	ErrListInvalidIndexType     = errors.New("list index must be an integer")
	ErrListInvalidAccessorType  = errors.New("invalid accessor type for collection")
	ErrMapKeyNotFound           = errors.New("key not found in map")
	ErrCannotAccessType         = errors.New("cannot perform element access on type")
	ErrCollectionIsNil          = errors.New("collection evaluated to nil")
	ErrAccessorIsNil            = errors.New("accessor evaluated to nil")
	// File System Tool Errors
	ErrPathViolation   = errors.New("path resolves outside allowed directory")
	ErrCannotCreateDir = errors.New("cannot create directory")
	ErrCannotDelete    = errors.New("cannot delete file or directory")
	// Go Tooling Errors
	ErrGoParseFailed  = errors.New("failed to parse Go source")
	ErrGoModifyFailed = errors.New("failed to modify Go AST")
	ErrGoFormatFailed = errors.New("failed to format Go AST") // *** ADDED ***
	// Add other general tool errors as needed
	ErrInternalTool = errors.New("internal tool error")
)

// --- Core Interpreter Errors ---
// (Potentially add interpreter-specific errors here later if needed)
