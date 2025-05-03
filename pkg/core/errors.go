// NeuroScript Version: 0.3.0
// Last Modified: 2025-05-02 17:36:24 PDT // Add ErrCannotRemoveRoot
// filename: pkg/core/errors.go
package core

import (
	"errors"
	"fmt"
)

// --- ErrorCode Type ---
type ErrorCode int

// --- RuntimeError ---
type RuntimeError struct {
	Code    ErrorCode
	Message string
	Wrapped error
}

func (e *RuntimeError) Error() string {
	if e.Wrapped != nil {
		// Including wrapped error might be too verbose for user-facing errors.
		// Consider logging it separately.
		return fmt.Sprintf("NeuroScript Error %d: %s", e.Code, e.Message)
	}
	return fmt.Sprintf("NeuroScript Error %d: %s", e.Code, e.Message)
}
func (e *RuntimeError) Unwrap() error { return e.Wrapped }
func NewRuntimeError(code ErrorCode, message string, wrapped error) *RuntimeError {
	return &RuntimeError{Code: code, Message: message, Wrapped: wrapped}
}

// --- Basic Runtime Error Codes ---
const (
	ErrorCodeGeneric         ErrorCode = 0
	ErrorCodeFailStatement   ErrorCode = 1
	ErrorCodeProcNotFound    ErrorCode = 2
	ErrorCodeToolNotFound    ErrorCode = 3
	ErrorCodeArgMismatch     ErrorCode = 4
	ErrorCodeMustFailed      ErrorCode = 5
	ErrorCodeInternal        ErrorCode = 6
	ErrorCodeType            ErrorCode = 7
	ErrorCodeBounds          ErrorCode = 8
	ErrorCodeKeyNotFound     ErrorCode = 9
	ErrorCodeSecurity        ErrorCode = 10
	ErrorCodeReadOnly        ErrorCode = 11
	ErrorCodeReturnViolation ErrorCode = 12
	ErrorCodeClearViolation  ErrorCode = 13
	ErrorCodeDivisionByZero  ErrorCode = 14
	ErrorCodeSyntax          ErrorCode = 15
	ErrorCodeLLMError        ErrorCode = 16
	ErrorCodeEvaluation      ErrorCode = 17
	ErrorCodeConfiguration   ErrorCode = 18
	ErrorCodeToolSpecific    ErrorCode = 1000
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
	ErrValidationRequired       = errors.New("validation error: missing required argument")
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

// --- Core Handle Errors ---
var (
	ErrHandleInvalid   = errors.New("handle is invalid or refers to invalid data")
	ErrHandleNotFound  = errors.New("handle not found")
	ErrHandleWrongType = errors.New("handle has wrong type")
)

// --- Core Tool Execution Errors ---
var (
	ErrInternalTool                  = errors.New("internal tool error")
	ErrNotFound                      = errors.New("item not found") // Generic not found
	ErrListIndexOutOfBounds          = errors.New("list index out of bounds")
	ErrListCannotSortMixedTypes      = errors.New("cannot sort list with mixed or non-sortable types")
	ErrListInvalidIndexType          = errors.New("list index must be an integer")
	ErrListInvalidAccessorType       = errors.New("invalid accessor type for collection")
	ErrMapKeyNotFound                = errors.New("key not found in map")
	ErrCannotAccessType              = errors.New("cannot perform element access on type")
	ErrCollectionIsNil               = errors.New("collection evaluated to nil")
	ErrAccessorIsNil                 = errors.New("accessor evaluated to nil")
	ErrCannotCreateDir               = errors.New("cannot create directory")
	ErrCannotDelete                  = errors.New("cannot delete file or directory")
	ErrInvalidHashAlgorithm          = errors.New("invalid or unsupported hash algorithm")
	ErrFileNotFound                  = errors.New("file not found") // Specific file not found
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
	ErrCacheObjectNotFound           = errors.New("object not found in cache")
	ErrCacheObjectWrongType          = errors.New("cached object has wrong type")
	ErrDivisionByZero                = errors.New("division by zero")
	ErrInvalidOperandType            = errors.New("invalid operand type")
	ErrInvalidOperandTypeNumeric     = errors.New("requires numeric operand(s)")
	ErrInvalidOperandTypeInteger     = errors.New("requires integer operand(s)")
	ErrInvalidOperandTypeString      = errors.New("requires string operand(s)")
	ErrInvalidOperandTypeBool        = errors.New("requires boolean operand(s)")
	ErrInvalidFunctionArgument       = errors.New("invalid argument for function")
	ErrVariableNotFound              = errors.New("variable not found")
	ErrUnsupportedOperator           = errors.New("unsupported operator")
	ErrNilOperand                    = errors.New("operation received nil operand")
	ErrUnknownFunction               = errors.New("unknown function called")
	ErrTypeAssertionFailed           = errors.New("type assertion failed")
	ErrMustConditionFailed           = errors.New("must condition evaluated to false")
	ErrSkippedBinaryFile             = errors.New("skipped potentially binary file")
	ErrTreeJSONUnmarshal             = errors.New("failed to unmarshal JSON input")
	ErrTreeBuildFailed               = errors.New("failed to build internal tree structure")
	ErrTreeFormatFailed              = errors.New("failed to reconstruct data for formatting")
	ErrTreeJSONMarshal               = errors.New("failed to marshal tree data to JSON")
	ErrTreeInvalidQuery              = errors.New("invalid query map structure or values")
	ErrTreeCannotSetValueOnType      = errors.New("cannot set Value on node types object or array")
	ErrTreeNodeNotObject             = errors.New("target node is not type object")
	ErrAttributeNotFound             = errors.New("attribute key not found on node")
	ErrNodeIDExists                  = errors.New("node ID already exists in tree")
	ErrCannotRemoveRoot              = errors.New("cannot remove the root node") // <<< ADDED
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
	ErrLLMError             = errors.New("LLM interaction failed")
	ErrLLMNotConfigured     = errors.New("LLM client not configured in interpreter")
)

// --- Control Flow Sentinel Errors ---
var (
	ErrBreak    = errors.New("internal: break signal")
	ErrContinue = errors.New("internal: continue signal")
)
