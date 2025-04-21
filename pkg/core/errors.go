// filename: pkg/core/errors.go
package core

import "errors"

// --- Core Validation Errors (for ValidateAndConvertArgs & Security) ---
var (
	ErrValidationRequiredArgNil = errors.New("required argument is nil")
	ErrValidationTypeMismatch   = errors.New("argument type mismatch")
	ErrValidationArgCount       = errors.New("incorrect argument count")
	ErrValidationArgValue       = errors.New("invalid argument value")
	ErrMissingArgument          = errors.New("required argument missing")
	ErrInvalidArgument          = errors.New("invalid argument")
	ErrNullByteInArgument       = errors.New("argument contains null byte")
)

// --- Core Security Errors ---
var (
	ErrToolDenied        = errors.New("tool explicitly denied")
	ErrToolNotAllowed    = errors.New("tool not in allowlist")
	ErrToolBlocked       = errors.New("tool blocked by security policy")
	ErrSecurityViolation = errors.New("security violation")
	ErrPathViolation     = errors.New("path resolves outside allowed directory")
	ErrInternalSecurity  = errors.New("internal security error")
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
	ErrCannotCreateDir      = errors.New("cannot create directory")
	ErrCannotDelete         = errors.New("cannot delete file or directory")
	ErrInvalidHashAlgorithm = errors.New("invalid or unsupported hash algorithm")
	ErrFileNotFound         = errors.New("file not found") // <<< ADDED HERE <<<
	// Go Tooling Errors
	ErrGoParseFailed                 = errors.New("failed to parse Go source")
	ErrGoModifyFailed                = errors.New("failed to modify Go AST")
	ErrGoFormatFailed                = errors.New("failed to format Go AST")
	ErrGoModifyInvalidDirectiveValue = errors.New("invalid value for GoModifyAST directive")
	ErrGoModifyMissingMapKey         = errors.New("missing required key in GoModifyAST directive map")
	ErrGoModifyEmptyMap              = errors.New("GoModifyAST modifications map cannot be empty")
	ErrGoModifyUnknownDirective      = errors.New("GoModifyAST modifications map contains no known directives")
	ErrGoInvalidIdentifierFormat     = errors.New("invalid identifier format (e.g., empty string)")
	// Cache Errors
	ErrCacheObjectNotFound  = errors.New("object not found in cache")
	ErrCacheObjectWrongType = errors.New("object found in cache has wrong type")
	// General Tool Error
	ErrInternalTool      = errors.New("internal tool error")
	ErrSkippedBinaryFile = errors.New("skipped potentially binary file")
)

// --- Core Interpreter Errors ---
var (
	ErrProcedureNotFound    = errors.New("procedure not found")         // Moved here from general
	ErrArgumentMismatch     = errors.New("argument mismatch")           // Moved here
	ErrMaxCallDepthExceeded = errors.New("maximum call depth exceeded") // Moved here
	ErrUnknownKeyword       = errors.New("unknown keyword")             // Moved here
	// Add others like ErrVariableNotFound if needed
)
