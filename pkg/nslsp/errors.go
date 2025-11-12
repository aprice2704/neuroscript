// NeuroScript Version: 0.7.0
// File version: 7
// Purpose: Defines standardized error codes for LSP diagnostics. FIX: Added DiagCodeUninitializedVar.
// filename: pkg/nslsp/errors.go
// nlines: 22
// risk_rating: LOW

package nslsp

// DiagnosticCode is a string identifier for a specific type of diagnostic error.
type DiagnosticCode string

const (
	// DiagCodeToolNotFound indicates that a tool definition could not be found.
	DiagCodeToolNotFound DiagnosticCode = "ToolNotFound"
	// DiagCodeArgCountMismatch indicates a mismatch between expected and actual arguments.
	DiagCodeArgCountMismatch DiagnosticCode = "ArgCountMismatch"
	// DiagCodeProcNotFound indicates that a procedure definition could not be found in the workspace.
	DiagCodeProcNotFound DiagnosticCode = "ProcNotFound"
	// DiagCodeOptionalArgMissing indicates that a tool is called with a valid number of arguments, but is missing one or more optional ones.
	DiagCodeOptionalArgMissing DiagnosticCode = "OptionalArgMissing"
	// DiagCodeUninitializedVar indicates that a variable is read before it has been assigned a value.
	DiagCodeUninitializedVar DiagnosticCode = "UninitializedVar"
)
