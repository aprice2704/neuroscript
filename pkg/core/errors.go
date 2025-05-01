// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-01 12:49:26 PDT
// filename: pkg/core/errors.go
package core

import (
	"errors"
	"fmt"
)

// --- ErrorCode Type ---
type ErrorCode int // Define a specific type for error codes

// --- RuntimeError ---
// RuntimeError represents a structured error originating from NeuroScript execution.
type RuntimeError struct {
	Code    ErrorCode // Use specific type
	Message string    // Human-readable error message
	Wrapped error     // Optional: The original underlying Go error, if any
	// Pos     *Position // Optional: Add position directly here if desired later
}

// Error implements the standard Go error interface.
func (e *RuntimeError) Error() string {
	// Optionally include Position info if added to struct:
	// posStr := ""
	// if e.Pos != nil {
	//  posStr = fmt.Sprintf(" at %s", e.Pos.String())
	// }
	if e.Wrapped != nil {
		// Consider including wrapped error details if helpful for debugging Go side
		// return fmt.Sprintf("NeuroScript Error %d%s: %s (wrapped: %v)", e.Code, posStr, e.Message, e.Wrapped)
		return fmt.Sprintf("NeuroScript Error %d: %s", e.Code, e.Message) // Simpler message for script
	}
	return fmt.Sprintf("NeuroScript Error %d: %s", e.Code, e.Message)
}

// Unwrap provides compatibility for errors.Is and errors.As.
func (e *RuntimeError) Unwrap() error {
	return e.Wrapped
}

// Helper to create a new RuntimeError
// NOTE: Does NOT accept Position directly. Position should be added to Message or struct later if needed.
func NewRuntimeError(code ErrorCode, message string, wrapped error) *RuntimeError {
	return &RuntimeError{Code: code, Message: message, Wrapped: wrapped}
}

// --- Basic Runtime Error Codes (Expand as needed) ---
// *** UPDATED: Added ErrorCodeEvaluation, ErrorCodeConfiguration ***
const (
	ErrorCodeGeneric         ErrorCode = 0    // Default or unknown script error
	ErrorCodeFailStatement   ErrorCode = 1    // Error explicitly raised by 'fail'
	ErrorCodeProcNotFound    ErrorCode = 2    // Procedure call target not found
	ErrorCodeToolNotFound    ErrorCode = 3    // Tool or tool function not found
	ErrorCodeArgMismatch     ErrorCode = 4    // Incorrect number/type of arguments
	ErrorCodeMustFailed      ErrorCode = 5    // 'must' or 'mustbe' condition failed
	ErrorCodeInternal        ErrorCode = 6    // Internal interpreter error (e.g., bad AST, nil pointer)
	ErrorCodeType            ErrorCode = 7    // Type error during operation (e.g., adding string to int)
	ErrorCodeBounds          ErrorCode = 8    // Index out of bounds
	ErrorCodeKeyNotFound     ErrorCode = 9    // Map key not found
	ErrorCodeSecurity        ErrorCode = 10   // Security policy violation
	ErrorCodeReadOnly        ErrorCode = 11   // Attempt to modify read-only variable (err_code/err_msg)
	ErrorCodeReturnViolation ErrorCode = 12   // 'return' used inside 'on_error'
	ErrorCodeClearViolation  ErrorCode = 13   // 'clear_error' used outside 'on_error'
	ErrorCodeDivisionByZero  ErrorCode = 14   // Division by zero
	ErrorCodeSyntax          ErrorCode = 15   // Syntax error during parsing or interpretation
	ErrorCodeLLMError        ErrorCode = 16   // Error during LLM API call or processing
	ErrorCodeEvaluation      ErrorCode = 17   // Error during expression evaluation (ADDED)
	ErrorCodeConfiguration   ErrorCode = 18   // Error related to interpreter/tool configuration (ADDED)
	ErrorCodeToolSpecific    ErrorCode = 1000 // Base for tool-specific errors
	// Add more codes as needed
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
	ErrAmbiguousSymbol               = errors.New("ambiguous exported symbol")
	// Cache Errors
	ErrCacheObjectNotFound  = errors.New("object not found in cache")
	ErrCacheObjectWrongType = errors.New("cached object has wrong type")
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
	ErrTypeAssertionFailed       = errors.New("type assertion failed") // <<< ADDED DEFINITION HERE

	// Verification Errors
	ErrMustConditionFailed = errors.New("must condition evaluated to false")
	// General Tool Error
	ErrInternalTool      = errors.New("internal tool error")
	ErrSkippedBinaryFile = errors.New("skipped potentially binary file")
)

// --- Core Interpreter Errors ---
// *** UPDATED: Added ErrLLMNotConfigured ***
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
	ErrLLMError             = errors.New("LLM interaction failed")
	ErrLLMNotConfigured     = errors.New("LLM client not configured in interpreter") // ADDED
)

// --- ADDED: Control Flow Sentinel Errors ---
var (
	// These are used internally to signal control flow, not typically user-facing errors.
	ErrBreak    = errors.New("internal: break signal")
	ErrContinue = errors.New("internal: continue signal")
	// TODO: Consider if ErrReturn should also be a simple sentinel error here,
	// or if it needs to carry values (currently handled via panic/recover in interpreter?).
	// For now, defining ErrBreak and ErrContinue is sufficient for this task.
)
