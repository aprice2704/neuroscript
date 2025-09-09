// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Defines sentinel errors for the json_lite package.
// filename: pkg/json-lite/errors.go
// nlines: 32
// risk_rating: LOW

package json_lite

import "errors"

var (
	// --- Path Errors ---

	// ErrInvalidPath indicates a path string has an invalid format, such as duplicate
	// dots, unterminated brackets, or disallowed characters.
	ErrInvalidPath = errors.New("invalid path format")
	// ErrNestingDepthExceeded is returned when a path or shape definition exceeds the
	// maximum allowed nesting depth.
	ErrNestingDepthExceeded = errors.New("nesting depth exceeded")
	// ErrMapKeyNotFound indicates a key specified in a path does not exist in the map.
	ErrMapKeyNotFound = errors.New("map key not found")
	// ErrListIndexOutOfBounds indicates an index is outside the valid range of a list.
	ErrListIndexOutOfBounds = errors.New("list index out of bounds")
	// ErrListInvalidIndexType indicates a list index in a path string is not a valid integer.
	ErrListInvalidIndexType = errors.New("invalid list index type")
	// ErrCannotAccessType is returned when trying to perform an operation on an incorrect
	// data type, like accessing a map key on a list.
	ErrCannotAccessType = errors.New("cannot access this type with the given operation")
	// ErrCollectionIsNil is returned when a path attempts to traverse through a nil map or slice.
	ErrCollectionIsNil = errors.New("collection is nil")

	// --- Shape/Validation Errors ---

	// ErrInvalidArgument is returned for invalid arguments, such as empty keys in shapes,
	// overlong path segments, or unexpected keys when allowExtra is false.
	ErrInvalidArgument = errors.New("invalid argument")
	// ErrValidationTypeMismatch is returned when a value's type does not match the
	// type specified in the shape definition.
	ErrValidationTypeMismatch = errors.New("validation type mismatch")
	// ErrValidationRequiredArgMissing is returned when a required key is missing from the data.
	ErrValidationRequiredArgMissing = errors.New("required argument missing")
	// ErrValidationFailed is returned for special string types (e.g., email, url)
	// when the value does not conform to the required format.
	ErrValidationFailed = errors.New("validation failed")
)
