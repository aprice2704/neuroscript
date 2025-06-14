// NeuroScript Version: 0.3.1
// File version: 0.2.2
// Purpose: Added ErrTreeIntegrity for clearer error reporting on structural issues, as suggested.
// nlines: 211
// risk_rating: LOW
// filename: pkg/core/errors.go
package core

import (
	"errors"
	"fmt"
	"strings"
)

// --- ErrorCode Type ---
type ErrorCode int

// --- RuntimeError ---
// MODIFIED: Added Position field
type RuntimeError struct {
	Code     ErrorCode
	Message  string
	Wrapped  error
	Position *Position // ADDED: To store position information
}

// MODIFIED: Updated Error() method to include position
func (e *RuntimeError) Error() string {
	msg := fmt.Sprintf("NeuroScript Error %d: %s", e.Code, e.Message)
	if e.Position != nil {
		msg = fmt.Sprintf("%s at %s", msg, e.Position.String())
	}
	if e.Wrapped != nil {
		msg = fmt.Sprintf("%s (wrapped: %v)", msg, e.Wrapped)
	}
	return msg
}
func (e *RuntimeError) Unwrap() error { return e.Wrapped }

// MODIFIED: Initialize Position to nil
func NewRuntimeError(code ErrorCode, message string, wrapped error) *RuntimeError {
	return &RuntimeError{Code: code, Message: message, Wrapped: wrapped, Position: nil}
}

// ADDED: WithPosition method for RuntimeError
func (e *RuntimeError) WithPosition(pos *Position) *RuntimeError {
	if e != nil {
		e.Position = pos
	}
	return e
}

// ADDED: wrapErrorWithPosition helper function
func WrapErrorWithPosition(err error, pos *Position, contextMsg string) error {
	if err == nil {
		return nil
	}
	var re *RuntimeError
	if errors.As(err, &re) {
		if re.Position == nil && pos != nil { // Add position if not already set
			re.Position = pos
		}
		// Prepend context message if it's not already part of the error.
		if contextMsg != "" && !strings.HasPrefix(re.Message, contextMsg) { // Avoid double-prefixing
			re.Message = fmt.Sprintf("%s: %s", contextMsg, re.Message)
		}
		return re
	}
	// If it's not already a RuntimeError, create a new one.
	fullMessage := contextMsg
	if err.Error() != "" { // Avoid "context: " if original error is empty
		fullMessage = fmt.Sprintf("%s: %s", contextMsg, err.Error())
	}
	// Use ErrorCodeEvaluation as a generic code for wrapped errors if not specified.
	return NewRuntimeError(ErrorCodeEvaluation, fullMessage, err).WithPosition(pos)
}

// --- Basic Runtime Error Codes ---
// These are general categories for runtime errors.
// Corresponding sentinel errors (ErrXyz) should be defined below.
const (
	ErrorCodeGeneric             ErrorCode = 0
	ErrorCodeFailStatement       ErrorCode = 1
	ErrorCodeProcNotFound        ErrorCode = 2
	ErrorCodeToolNotFound        ErrorCode = 3
	ErrorCodeArgMismatch         ErrorCode = 4
	ErrorCodeMustFailed          ErrorCode = 5
	ErrorCodeInternal            ErrorCode = 6
	ErrorCodeType                ErrorCode = 7
	ErrorCodeBounds              ErrorCode = 8
	ErrorCodeKeyNotFound         ErrorCode = 9 // General key not found (e.g. map key)
	ErrorCodeSecurity            ErrorCode = 10
	ErrorCodeReadOnly            ErrorCode = 11
	ErrorCodeReturnViolation     ErrorCode = 12
	ErrorCodeClearViolation      ErrorCode = 13
	ErrorCodeDivisionByZero      ErrorCode = 14
	ErrorCodeSyntax              ErrorCode = 15 // Includes parsing errors like JSON, NeuroScript syntax
	ErrorCodeLLMError            ErrorCode = 16
	ErrorCodeEvaluation          ErrorCode = 17
	ErrorCodeConfiguration       ErrorCode = 18
	ErrorCodePreconditionFailed  ErrorCode = 19
	ErrorCodeRateLimited         ErrorCode = 20
	ErrorCodeToolExecutionFailed ErrorCode = 21 // Error code for general tool execution failures
	ErrorCodeNotImplemented      ErrorCode = 30
	ErrorCodeTimeout             ErrorCode = 31

	// Filesystem Specific Error Codes (start from 22)
	ErrorCodeFileNotFound     ErrorCode = 22
	ErrorCodePathTypeMismatch ErrorCode = 23
	ErrorCodePathExists       ErrorCode = 24
	ErrorCodePermissionDenied ErrorCode = 25
	ErrorCodeIOFailed         ErrorCode = 26

	// Tree Specific Error Codes (start from 27)
	ErrorCodeTreeConstraintViolation ErrorCode = 27 // e.g., cannot set value on object, cannot remove root, ID exists
	ErrorCodeNodeWrongType           ErrorCode = 28 // e.g., expected object, got value
	ErrorCodeAttributeNotFound       ErrorCode = 29 // For metadata access
	ErrorCodeUnknownKeyword          ErrorCode = 30
	ErrorCodeTypeAssertionFailed     ErrorCode = 31
	ErrorCodeExecutionFailed         ErrorCode = 32
	ErrorCodeTreeIntegrity           ErrorCode = 33 // ADDED: e.g., child ID exists but node is missing from map
	ErrorCodePathViolation           ErrorCode = 34
	ErrorCodeFeatureNotImplemented   ErrorCode = 35

	ErrorCodeToolSpecific ErrorCode = 1000 // Base for tool-specific error codes (non-FS/Tree or highly unique cases)
)

// --- Basic Runtime Sentinel Errors ---
var (
	ErrConfiguration = errors.New("invalid configuration") // For ErrorCodeConfiguration
	// Define other basic sentinels here if needed, e.g.:
	// ErrType = errors.New("type error") // For ErrorCodeType if a general one is useful
)

// --- Core Validation Errors ---
var (
	ErrValidationRequiredArgNil     = errors.New("required argument is nil")
	ErrValidationRequiredArgMissing = errors.New("required argument is missing")
	ErrValidationTypeMismatch       = errors.New("argument type mismatch")
	ErrValidationArgCount           = errors.New("incorrect argument count")
	ErrValidationArgValue           = errors.New("invalid argument value")
	ErrMissingArgument              = errors.New("required argument missing") // Consider consolidating with ErrValidationRequiredArgMissing
	ErrInvalidArgument              = errors.New("invalid argument")
	ErrInvalidInput                 = errors.New("invalid input")
	ErrNullByteInArgument           = errors.New("argument contains null byte")
	ErrIncorrectArgCount            = errors.New("incorrect function argument count")           // Consider consolidating with ErrValidationArgCount
	ErrValidationRequired           = errors.New("validation error: missing required argument") // Consider consolidating with ErrValidationRequiredArgMissing
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
	ErrHandleNotFound  = errors.New("handle not found in cache")
	ErrHandleWrongType = errors.New("handle has wrong type")
)

// --- Core Tool Execution Errors (including Filesystem and Tree sentinels) ---
var (
	// General Tool Errors
	ErrInternalTool       = errors.New("internal tool error")
	ErrNotFound           = errors.New("item not found") // Generic not found by a tool
	ErrFailedPrecondition = errors.New("operation failed due to a precondition not being met")
	// ErrRateLimited is defined above
	// ErrNotImplemented is defined above
	ErrToolExecutionFailed = errors.New("tool execution failed") // Sentinel for ErrorCodeToolExecutionFailed

	// Filesystem Errors
	ErrFileNotFound      = errors.New("file not found")                  // For ErrorCodeFileNotFound
	ErrPathNotFile       = errors.New("path is not a file")              // For ErrorCodePathTypeMismatch
	ErrPathNotDirectory  = errors.New("path is not a directory")         // For ErrorCodePathTypeMismatch
	ErrPathExists        = errors.New("path already exists")             // For ErrorCodePathExists
	ErrPermissionDenied  = errors.New("permission denied")               // For ErrorCodePermissionDenied
	ErrIOFailed          = errors.New("i/o operation failed")            // For ErrorCodeIOFailed
	ErrCannotCreateDir   = errors.New("cannot create directory")         // Use with ErrorCodePathExists or ErrorCodeIOFailed
	ErrCannotDelete      = errors.New("cannot delete file or directory") // Use with ErrorCodePreconditionFailed (dir not empty) or ErrorCodeIOFailed
	ErrSkippedBinaryFile = errors.New("skipped potentially binary file") // Specific FS case

	// Tree Errors
	ErrTreeConstraintViolation = errors.New("tree constraint violation")                      // For ErrorCodeTreeConstraintViolation
	ErrNodeWrongType           = errors.New("incorrect node type for operation")              // For ErrorCodeNodeWrongType
	ErrAttributeNotFound       = errors.New("attribute not found on node")                    // For ErrorCodeAttributeNotFound
	ErrTreeJSONUnmarshal       = errors.New("failed to unmarshal JSON input")                 // Use with ErrorCodeSyntax
	ErrTreeJSONMarshal         = errors.New("failed to marshal tree structure to JSON")       // Use with ErrorCodeInternal
	ErrTreeInvalidQuery        = errors.New("invalid query map structure or values")          // Use with ErrorCodeArgMismatch
	ErrCannotSetValueOnType    = errors.New("cannot set Value on node types object or array") // Use with ErrorCodeTreeConstraintViolation
	ErrTreeNodeNotObject       = errors.New("expected tree node to be an object type")        // Use with ErrorCodeNodeWrongType
	ErrNodeIDExists            = errors.New("node ID already exists in tree")                 // Use with ErrorCodeTreeConstraintViolation
	ErrCannotRemoveRoot        = errors.New("cannot remove the root node")                    // Use with ErrorCodeTreeConstraintViolation
	ErrTreeIntegrity           = errors.New("tree integrity violation")                       // ADDED: For ErrorCodeTreeIntegrity

	// List/Map/Collection Errors
	ErrListIndexOutOfBounds     = errors.New("list index out of bounds") // Use with ErrorCodeBounds
	ErrListCannotSortMixedTypes = errors.New("cannot sort list with mixed or non-sortable types")
	ErrListInvalidIndexType     = errors.New("list index must be an integer")
	ErrListInvalidAccessorType  = errors.New("invalid accessor type for collection")
	ErrMapKeyNotFound           = errors.New("key not found in map") // Use with ErrorCodeKeyNotFound
	ErrCannotAccessType         = errors.New("cannot perform element access on type")
	ErrCollectionIsNil          = errors.New("collection evaluated to nil")
	ErrAccessorIsNil            = errors.New("accessor evaluated to nil")

	// Go Tool specific errors (from goast, gosemantic etc.)
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

	// Other Tool Errors
	ErrInvalidHashAlgorithm      = errors.New("invalid or unsupported hash algorithm")
	ErrCacheObjectNotFound       = errors.New("object not found in cache")
	ErrCacheObjectWrongType      = errors.New("cached object has wrong type")
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
	// ErrTypeAssertionFailed is defined above
)

// --- Core Interpreter Errors ---
var (
	ErrProcedureNotFound    = errors.New("procedure not found") // For ErrorCodeProcNotFound
	ErrArgumentMismatch     = errors.New("argument mismatch")   // For ErrorCodeArgMismatch
	ErrReturnMismatch       = errors.New("procedure return count mismatch")
	ErrProcedureExists      = errors.New("procedure already defined")
	ErrMaxCallDepthExceeded = errors.New("maximum call depth exceeded")
	// ErrUnknownKeyword is defined above
	// ErrUnhandledException is defined above
	ErrFailStatement     = errors.New("execution halted by FAIL statement")                           // For ErrorCodeFailStatement
	ErrInternal          = errors.New("internal interpreter error")                                   // For ErrorCodeInternal
	ErrReadOnlyViolation = errors.New("attempt to modify read-only variable")                         // For ErrorCodeReadOnly
	ErrUnsupportedSyntax = errors.New("unsupported syntax")                                           // For ErrorCodeSyntax
	ErrClearViolation    = errors.New("clear_error used outside on_error block")                      // For ErrorCodeClearViolation
	ErrReturnViolation   = errors.New("'return' statement is not permitted inside an on_error block") // For ErrorCodeReturnViolation
	// ErrToolNotFound is defined above
	ErrLLMError            = errors.New("LLM interaction failed") // For ErrorCodeLLMError
	ErrLLMNotConfigured    = errors.New("LLM client not configured in interpreter")
	ErrDivisionByZero      = errors.New("division by zero")                  // For ErrorCodeDivisionByZero
	ErrMustConditionFailed = errors.New("must condition evaluated to false") // For ErrorCodeMustFailed

	// AI WM Errors
	ErrAuthDetailsMissing    = errors.New("authentication details are missing")
	ErrAPIKeyNotFound        = errors.New("API key not found though configuration implies one should exist")
	ErrFeatureNotImplemented = errors.New("feature not implemented")

	ErrRateLimited         = errors.New("operation failed due to rate limiting")
	ErrToolNotFound        = errors.New("tool or tool function not found")
	ErrUnknownKeyword      = errors.New("unknown keyword encountered")
	ErrTypeAssertionFailed = errors.New("type assertion failed")
	ErrNotImplemented      = errors.New("feature or tool not implemented")
)

// --- Control Flow Sentinel Errors ---
var (
	ErrBreak    = errors.New("internal: break signal")
	ErrContinue = errors.New("internal: continue signal")
)
