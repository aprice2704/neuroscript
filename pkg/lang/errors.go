// filename: pkg/lang/errors.go
// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Added the ErrMaxIterationsExceeded sentinel error for loop resource limits.
// nlines: 230
// risk_rating: LOW

package lang

import (
	"errors"
	"fmt"
	"strings"
)

// --- ErrorCode Type ---
type ErrorCode int

// --- RuntimeError ---
type RuntimeError struct {
	Code     ErrorCode
	Message  string
	Wrapped  error
	Position *Position
}

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

func NewRuntimeError(code ErrorCode, message string, wrapped error) *RuntimeError {
	return &RuntimeError{Code: code, Message: message, Wrapped: wrapped, Position: nil}
}

func (e *RuntimeError) WithPosition(pos *Position) *RuntimeError {
	if e != nil {
		e.Position = pos
	}
	return e
}

func WrapErrorWithPosition(err error, pos *Position, contextMsg string) error {
	if err == nil {
		return nil
	}
	var re *RuntimeError
	if errors.As(err, &re) {
		if re.Position == nil && pos != nil {
			re.Position = pos
		}
		if contextMsg != "" && !strings.HasPrefix(re.Message, contextMsg) {
			re.Message = fmt.Sprintf("%s: %s", contextMsg, re.Message)
		}
		return re
	}
	fullMessage := contextMsg
	if err.Error() != "" {
		fullMessage = fmt.Sprintf("%s: %s", contextMsg, err.Error())
	}
	return NewRuntimeError(ErrorCodeEvaluation, fullMessage, err).WithPosition(pos)
}

// --- Basic Runtime Error Codes ---
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
	ErrorCodeKeyNotFound         ErrorCode = 9
	ErrorCodeSecurity            ErrorCode = 10
	ErrorCodeReadOnly            ErrorCode = 11
	ErrorCodeReturnViolation     ErrorCode = 12
	ErrorCodeClearViolation      ErrorCode = 13
	ErrorCodeDivisionByZero      ErrorCode = 14
	ErrorCodeSyntax              ErrorCode = 15
	ErrorCodeLLMError            ErrorCode = 16
	ErrorCodeEvaluation          ErrorCode = 17
	ErrorCodeConfiguration       ErrorCode = 18
	ErrorCodePreconditionFailed  ErrorCode = 19
	ErrorCodeRateLimited         ErrorCode = 20
	ErrorCodeToolExecutionFailed ErrorCode = 21

	// Filesystem Specific Error Codes (start from 22)
	ErrorCodeFileNotFound     ErrorCode = 22
	ErrorCodePathTypeMismatch ErrorCode = 23
	ErrorCodePathExists       ErrorCode = 24
	ErrorCodePermissionDenied ErrorCode = 25
	ErrorCodeIOFailed         ErrorCode = 26

	// Tree Specific Error Codes (start from 27)
	ErrorCodeTreeConstraintViolation ErrorCode = 27
	ErrorCodeNodeWrongType           ErrorCode = 28
	ErrorCodeAttributeNotFound       ErrorCode = 29
	ErrorCodeUnknownKeyword          ErrorCode = 30
	ErrorCodeTypeAssertionFailed     ErrorCode = 31
	ErrorCodeExecutionFailed         ErrorCode = 32
	ErrorCodeTreeIntegrity           ErrorCode = 33
	ErrorCodePathViolation           ErrorCode = 34
	ErrorCodeNotImplemented          ErrorCode = 35
	ErrorCodeCountMismatch           ErrorCode = 36

	// Resource Error Codes (start from 37)
	ErrorCodeResourceExhaustion   ErrorCode = 37
	ErrorCodeNestingDepthExceeded ErrorCode = 38
	ErrorCodeControlFlow          ErrorCode = 39

	// --- SECURITY codes (99 900-99 999).  Stable for signing / IR play-books. ----
	securityBase ErrorCode = 99900
)

const (
	_                             ErrorCode = securityBase + iota // 99900 reserved (placeholder)
	ErrorCodeAttackPossible                                       // 99901
	ErrorCodeAttackProbable                                       // 99902
	ErrorCodeAttackCertain                                        // 99903
	ErrorCodeSubsystemCompromised                                 // 99904
	ErrorCodeSubsystemQuarantined                                 // 99905
	ErrorCodeEscapePossible                                       // 99906
	ErrorCodeEscapeProbable                                       // 99907
	ErrorCodeEscapeCertain                                        // 99908
)

///////////////////////////////////////////////////////////////////////////////
//  Sentinel errors – usable with errors.Is / errors.As
///////////////////////////////////////////////////////////////////////////////

var (

	// Security
	ErrAttackPossible       = errors.New("ATTACK possible")
	ErrAttackProbable       = errors.New("ATTACK probable")
	ErrAttackCertain        = errors.New("ATTACK underway")
	ErrSubsystemCompromised = errors.New("COMPROMISED subsystem – discard communications")
	ErrSubsystemQuarantined = errors.New("QUARANTINED subsystem – do not interact")
	ErrEscapePossible       = errors.New("ESCAPE possible")
	ErrEscapeProbable       = errors.New("ESCAPE probable")
	ErrEscapeCertain        = errors.New("ESCAPE underway")

	// Tools
	ErrorCodeToolSpecific ErrorCode = 1000
)

// --- Basic Runtime Sentinel Errors ---
var (
	ErrConfiguration = errors.New("invalid configuration")
)

// --- Core Validation Errors ---
var (
	ErrValidationRequiredArgNil     = errors.New("required argument is nil")
	ErrValidationRequiredArgMissing = errors.New("required argument is missing")
	ErrValidationTypeMismatch       = errors.New("argument type mismatch")
	ErrValidationArgCount           = errors.New("incorrect argument count")
	ErrValidationArgValue           = errors.New("invalid argument value")
	ErrMissingArgument              = errors.New("required argument missing")
	ErrInvalidArgument              = errors.New("invalid argument")
	ErrInvalidInput                 = errors.New("invalid input")
	ErrNullByteInArgument           = errors.New("argument contains null byte")
	ErrIncorrectArgCount            = errors.New("incorrect function argument count")
	ErrValidationRequired           = errors.New("validation error: missing required argument")
)

// --- Core Security & Resource Errors ---
var (
	ErrToolDenied            = errors.New("tool explicitly denied")
	ErrToolNotAllowed        = errors.New("tool not in allowlist")
	ErrToolBlocked           = errors.New("tool blocked by security policy")
	ErrSecurityViolation     = errors.New("security violation")
	ErrPathViolation         = errors.New("path resolves outside allowed directory")
	ErrInternalSecurity      = errors.New("internal security error")
	ErrInvalidPath           = errors.New("invalid path")
	ErrResourceExhaustion    = errors.New("resource exhaustion limit reached")
	ErrNestingDepthExceeded  = errors.New("maximum nesting depth exceeded")
	ErrMaxIterationsExceeded = errors.New("maximum loop iterations exceeded") // FIX: Added this line
)

// --- Core Handle Errors ---
var (
	ErrHandleInvalid   = errors.New("handle is invalid or refers to invalid data")
	ErrHandleNotFound  = errors.New("handle not found in cache")
	ErrHandleWrongType = errors.New("handle has wrong type")
)

// --- Core Tool Execution Errors (including Filesystem and Tree sentinels) ---
var (
	ErrInternalTool        = errors.New("internal tool error")
	ErrInvalidToolGroup    = errors.New("tool group is invalid")
	ErrInvalidToolName     = errors.New("tool name is invalid")
	ErrNotFound            = errors.New("item not found")
	ErrFailedPrecondition  = errors.New("operation failed due to a precondition not being met")
	ErrToolExecutionFailed = errors.New("tool execution failed")

	ErrFileNotFound      = errors.New("file not found")
	ErrPathNotFile       = errors.New("path is not a file")
	ErrPathNotDirectory  = errors.New("path is not a directory")
	ErrPathExists        = errors.New("path already exists")
	ErrPermissionDenied  = errors.New("permission denied")
	ErrIOFailed          = errors.New("i/o operation failed")
	ErrCannotCreateDir   = errors.New("cannot create directory")
	ErrCannotDelete      = errors.New("cannot delete file or directory")
	ErrSkippedBinaryFile = errors.New("skipped potentially binary file")

	ErrTreeConstraintViolation = errors.New("tree constraint violation")
	ErrNodeWrongType           = errors.New("incorrect node type for operation")
	ErrAttributeNotFound       = errors.New("attribute not found on node")
	ErrTreeJSONUnmarshal       = errors.New("failed to unmarshal JSON input")
	ErrTreeJSONMarshal         = errors.New("failed to marshal tree structure to JSON")
	ErrTreeInvalidQuery        = errors.New("invalid query map structure or values")
	ErrCannotSetValueOnType    = errors.New("cannot set Value on node types object or array")
	ErrTreeNodeNotObject       = errors.New("expected tree node to be an object type")
	ErrNodeIDExists            = errors.New("node ID already exists in tree")
	ErrCannotRemoveRoot        = errors.New("cannot remove the root node")
	ErrTreeIntegrity           = errors.New("tree integrity violation")

	ErrListIndexOutOfBounds     = errors.New("list index out of bounds")
	ErrListCannotSortMixedTypes = errors.New("cannot sort list with mixed or non-sortable types")
	ErrListInvalidIndexType     = errors.New("list index must be an integer")
	ErrListInvalidAccessorType  = errors.New("invalid accessor type for collection")
	ErrMapKeyNotFound           = errors.New("key not found in map")
	ErrCannotAccessType         = errors.New("cannot perform element access on type")
	ErrCollectionIsNil          = errors.New("collection evaluated to nil")
	ErrAccessorIsNil            = errors.New("accessor evaluated to nil")

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
)

// --- Core Interpreter Errors ---
var (
	ErrProcedureNotFound    = errors.New("procedure not found")
	ErrArgumentMismatch     = errors.New("argument mismatch")
	ErrReturnMismatch       = errors.New("procedure return count mismatch")
	ErrProcedureExists      = errors.New("procedure already defined")
	ErrMaxCallDepthExceeded = errors.New("maximum call depth exceeded")
	ErrFailStatement        = errors.New("execution halted by FAIL statement")
	ErrInternal             = errors.New("internal interpreter error")
	ErrReadOnlyViolation    = errors.New("attempt to modify read-only variable")
	ErrUnsupportedSyntax    = errors.New("unsupported syntax")
	ErrClearViolation       = errors.New("clear_error used outside on_error block")
	ErrReturnViolation      = errors.New("'return' statement is not permitted inside an on_error block")
	ErrLLMError             = errors.New("LLM interaction failed")
	ErrLLMNotConfigured     = errors.New("LLM client not configured in interpreter")
	ErrDivisionByZero       = errors.New("division by zero")
	ErrMustConditionFailed  = errors.New("must condition evaluated to false")
	ErrAssignCountMismatch  = errors.New("assignment count mismatch")
	ErrMultiAssignNonList   = errors.New("multiple assignment requires a list value on the right-hand side")

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

///////////////////////////////////////////////////////////////////////////////
//  Remediation catalogue
///////////////////////////////////////////////////////////////////////////////

// Severity indicates how loud ops should shout.
type Severity uint8

const (
	SevInfo Severity = iota
	SevWarn
	SevError
	SevCritical
)

// ErrorDetails enriches an ErrorCode with extra metadata.
type ErrorDetails struct {
	Code        ErrorCode
	Summary     string
	Remediation string
	Severity    Severity
}

// errorCatalogue holds the built-ins; tools can register more at runtime.
var errorCatalogue = map[ErrorCode]ErrorDetails{
	//  -- runtime samples
	ErrorCodeProcNotFound: {
		Code:        ErrorCodeProcNotFound,
		Summary:     "procedure not found",
		Remediation: "Check declaration spelling and package imports.",
		Severity:    SevError,
	},

	//  -- security
	ErrorCodeAttackPossible: {
		Code:        ErrorCodeAttackPossible,
		Summary:     "attack possible",
		Remediation: "Increase logging; monitor for escalation.",
		Severity:    SevWarn,
	},
	ErrorCodeAttackProbable: {
		Code:        ErrorCodeAttackProbable,
		Summary:     "attack probable",
		Remediation: "Activate heightened alert; restrict non-essential accounts.",
		Severity:    SevError,
	},
	ErrorCodeAttackCertain: {
		Code:        ErrorCodeAttackCertain,
		Summary:     "attack underway",
		Remediation: "Isolate affected subsystems; initiate incident-response plan.",
		Severity:    SevCritical,
	},
	ErrorCodeSubsystemCompromised: {
		Code:        ErrorCodeSubsystemCompromised,
		Summary:     "subsystem compromised",
		Remediation: "Quarantine subsystem; revoke credentials immediately.",
		Severity:    SevCritical,
	},
	ErrorCodeSubsystemQuarantined: {
		Code:        ErrorCodeSubsystemQuarantined,
		Summary:     "subsystem quarantined",
		Remediation: "Do not reintegrate until full forensic audit passes.",
		Severity:    SevError,
	},
	ErrorCodeEscapePossible: {
		Code:        ErrorCodeEscapePossible,
		Summary:     "escape possible",
		Remediation: "Verify container/VM boundaries; check audit logs.",
		Severity:    SevWarn,
	},
	ErrorCodeEscapeProbable: {
		Code:        ErrorCodeEscapeProbable,
		Summary:     "escape probable",
		Remediation: "Suspend workload; investigate hypervisor integrity.",
		Severity:    SevError,
	},
	ErrorCodeEscapeCertain: {
		Code:        ErrorCodeEscapeCertain,
		Summary:     "escape certain",
		Remediation: "Treat host as fully compromised; execute IR checklist.",
		Severity:    SevCritical,
	},
}

// RegisterErrorDetails lets tool packages add their own codes.
func RegisterErrorDetails(d ErrorDetails) {
	errorCatalogue[d.Code] = d
}

// LookupErrorDetails returns metadata for a code (bool=false if unknown).
func LookupErrorDetails(code ErrorCode) (ErrorDetails, bool) {
	d, ok := errorCatalogue[code]
	return d, ok
}

// FormatWithRemediation prettifies a RuntimeError for humans/dev-ops.
func FormatWithRemediation(err error) string {
	var r *RuntimeError
	if !errors.As(err, &r) {
		return err.Error()
	}
	sb := &strings.Builder{}
	sb.WriteString(r.Error())

	if d, ok := LookupErrorDetails(r.Code); ok && d.Remediation != "" {
		sb.WriteString("\n↳ Remediation: ")
		sb.WriteString(d.Remediation)
	}
	if r.Position != nil {
		sb.WriteString(fmt.Sprintf("\n↳ at %s", r.Position))
	}
	return sb.String()
}

///////////////////////////////////////////////////////////////////////////////
//  Position (stub – real one lives in parser package)
///////////////////////////////////////////////////////////////////////////////

type Pos struct {
	Filename string
	Line     int
	Column   int
}

func (p *Pos) String() string {
	if p == nil {
		return "<unknown>"
	}
	return fmt.Sprintf("%s:%d:%d", p.Filename, p.Line, p.Column)
}
