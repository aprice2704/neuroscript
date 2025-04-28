// filename: pkg/core/errors.go
package core

import (
	"errors"
	"fmt"
)

// --- NEW: RuntimeError ---
// RuntimeError represents a structured error originating from NeuroScript execution.
type RuntimeError struct {
	Code    int    // Numeric code for categorization
	Message string // Human-readable error message
	Wrapped error  // Optional: The original underlying Go error, if any
}

// Error implements the standard Go error interface.
func (e *RuntimeError) Error() string {
	if e.Wrapped != nil {
		// Consider including wrapped error details if helpful for debugging Go side
		// return fmt.Sprintf("NeuroScript Error %d: %s (wrapped: %v)", e.Code, e.Message, e.Wrapped)
		return fmt.Sprintf("NeuroScript Error %d: %s", e.Code, e.Message) // Simpler message for script
	}
	return fmt.Sprintf("NeuroScript Error %d: %s", e.Code, e.Message)
}

// Unwrap provides compatibility for errors.Is and errors.As.
func (e *RuntimeError) Unwrap() error {
	return e.Wrapped
}

// Helper to create a new RuntimeError
func NewRuntimeError(code int, message string, wrapped error) *RuntimeError {
	return &RuntimeError{Code: code, Message: message, Wrapped: wrapped}
}

// --- Basic Runtime Error Codes (Expand as needed) ---
const (
	ErrorCodeGeneric         = 0  // Default or unknown script error
	ErrorCodeFailStatement   = 1  // Error explicitly raised by 'fail'
	ErrorCodeProcNotFound    = 2  // Procedure call target not found
	ErrorCodeToolNotFound    = 3  // Tool or tool function not found
	ErrorCodeArgMismatch     = 4  // Incorrect number/type of arguments
	ErrorCodeMustFailed      = 5  // 'must' or 'mustbe' condition failed
	ErrorCodeInternal        = 6  // Internal interpreter error (e.g., bad AST, nil pointer)
	ErrorCodeType            = 7  // Type error during operation (e.g., adding string to int)
	ErrorCodeBounds          = 8  // Index out of bounds
	ErrorCodeKeyNotFound     = 9  // Map key not found
	ErrorCodeSecurity        = 10 // Security policy violation
	ErrorCodeReadOnly        = 11 // Attempt to modify read-only variable (err_code/err_msg)
	ErrorCodeReturnViolation = 12 // 'return' used inside 'on_error'
	ErrorCodeClearViolation  = 13 // 'clear_error' used outside 'on_error'
	ErrorCodeDivisionByZero  = 14 // Division by zero
	ErrorCodeSyntax          = 15 // Syntax error during parsing or interpretation

	// --- NEW Codes for ask... ---
	ErrorCodeLLMError               = 16 // Error during LLM API call or processing
	ErrorCodeHumanInteractionError  = 17 // Error during askHuman interaction (future use)
	ErrorCodeComputerExecutionError = 18 // Error during askComputer specific logic (if distinct from tool errors)
	// --- End NEW Codes ---

	// ... add more specific codes ...
	ErrorCodeToolSpecific = 1000 // Base for tool-specific errors
)

// --- Core Validation Errors ---
var (
	ErrValidationRequiredArgNil = errors.New("required argument is nil")
	ErrValidationTypeMismatch   = errors.New("argument type mismatch")
	ErrValidationArgCount       = errors.New("incorrect argument count")
	ErrValidationArgValue       = errors.New("invalid argument value")
	ErrMissingArgument          = errors.New("required argument missing")
	ErrInvalidArgument          = errors.New("invalid argument")
	ErrNullByteInArgument       = errors.New("argument contains null byte")
	ErrIncorrectArgCount        = errors.New("incorrect function argument count")
)

// --- Core Security Errors ---
var (
	ErrToolDenied        = errors.New("tool explicitly denied")
	ErrToolNotAllowed    = errors.New("tool not in allowlist")
	ErrToolBlocked       = errors.New("tool blocked by security policy")
	ErrSecurityViolation = errors.New("security violation")
	ErrPathViolation     = errors.New("path resolves outside allowed directory")
	ErrInternalSecurity  = errors.New("internal security error")
	ErrInvalidPath       = errors.New("invalid path")
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
	ErrFileNotFound         = errors.New("file not found")
	// Go Tooling Errors
	ErrGoParseFailed                 = errors.New("failed to parse Go source")
	ErrGoModifyFailed                = errors.New("failed to modify Go AST")
	ErrGoFormatFailed                = errors.New("failed to format Go AST")
	ErrGoModifyInvalidDirectiveValue = errors.New("invalid value for GoModifyAST directive")
	ErrGoModifyMissingMapKey         = errors.New("missing required key in GoModifyAST directive map")
	ErrGoModifyEmptyMap              = errors.New("GoModifyAST modifications map cannot be empty")
	ErrGoModifyUnknownDirective      = errors.New("GoModifyAST modifications map contains no known directives")
	ErrGoInvalidIdentifierFormat     = errors.New("invalid identifier format (e.g., empty string)")
	ErrRefactoredPathNotFound        = errors.New("refactored package path not found for symbol mapping")
	ErrSymbolMappingFailed           = errors.New("failed to build symbol map from refactored packages")
	ErrSymbolNotFoundInMap           = errors.New("symbol used from original package not found in new location map")
	ErrAmbiguousSymbol               = errors.New("ambiguous exported symbol")
	// Cache Errors
	ErrCacheObjectNotFound  = errors.New("object not found in cache")
	ErrCacheObjectWrongType = errors.New("object found in cache has wrong type")
	// Math/Evaluation Errors
	ErrDivisionByZero            = errors.New("division by zero")
	ErrInvalidOperandType        = errors.New("invalid operand type")
	ErrInvalidOperandTypeNumeric = errors.New("requires numeric operand(s)")
	ErrInvalidOperandTypeInteger = errors.New("requires integer operand(s)")
	ErrInvalidOperandTypeString  = errors.New("requires string operand(s)")
	ErrInvalidOperandTypeBool    = errors.New("requires boolean operand(s)")
	ErrInvalidFunctionArgument   = errors.New("invalid argument for function")
	ErrVariableNotFound          = errors.New("variable not found")
	ErrUnsupportedOperator       = errors.New("unsupported operator")
	ErrNilOperand                = errors.New("operation received nil operand")
	ErrUnknownFunction           = errors.New("unknown function called")

	// Verification Errors
	ErrMustConditionFailed = errors.New("must condition evaluated to false")
	// General Tool Error
	ErrInternalTool      = errors.New("internal tool error")
	ErrSkippedBinaryFile = errors.New("skipped potentially binary file")
)

// --- Core Interpreter Errors ---
var (
	ErrProcedureNotFound    = errors.New("procedure not found")
	ErrArgumentMismatch     = errors.New("argument mismatch")
	ErrMaxCallDepthExceeded = errors.New("maximum call depth exceeded")
	ErrUnknownKeyword       = errors.New("unknown keyword")
	ErrUnhandledException   = errors.New("unhandled exception during execution")
	ErrFailStatement        = errors.New("execution halted by FAIL statement")
	ErrInternal             = errors.New("internal interpreter error")
	ErrReadOnlyViolation    = errors.New("attempt to modify read-only variable")
	ErrUnsupportedSyntax    = errors.New("unsupported syntax")
	ErrClearViolation       = errors.New("clear_error used outside on_error block")
	ErrReturnViolation      = errors.New("'return' statement is not permitted inside an on_error block")
	ErrToolNotFound         = errors.New("tool or tool function not found")
	// --- NEW Sentinels for ask... ---
	ErrLLMError = errors.New("LLM interaction failed") // Used in executeAskAI
	// --- End NEW Sentinels ---
)
