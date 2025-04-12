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
	ErrListIndexOutOfBounds     = errors.New("list index out of bounds") // Used in list access node eval, not directly by tools yet
	ErrListCannotSortMixedTypes = errors.New("cannot sort list with mixed or non-sortable types")
	ErrListInvalidIndexType     = errors.New("list index must be an integer")         // Used in list access node eval
	ErrListInvalidAccessorType  = errors.New("invalid accessor type for collection")  // Used in list access node eval
	ErrMapKeyNotFound           = errors.New("key not found in map")                  // Used in map access node eval
	ErrCannotAccessType         = errors.New("cannot perform element access on type") // Used in element access node eval
	ErrCollectionIsNil          = errors.New("collection evaluated to nil")           // Used in element access node eval
	ErrAccessorIsNil            = errors.New("accessor evaluated to nil")             // Used in element access node eval
	// File System Tool Errors
	ErrPathViolation = errors.New("path resolves outside allowed directory") // ADDED
	// Add other general tool errors as needed
	ErrInternalTool = errors.New("internal tool error")
)

// --- Core Interpreter Errors ---
// (Potentially add interpreter-specific errors here later if needed)
