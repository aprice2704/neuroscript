// pkg/core/ast.go
package core

import "fmt" // Ensure fmt is imported if needed by String() methods etc.

// --- Expression Node Types (Unchanged from user-provided version) ---
// Note: These are defined as interfaces in the provided code snippet,
// assuming concrete types like StringLiteralNode etc. exist elsewhere or
// these interfaces are populated with concrete types by the expression builder.
type VariableNode struct{ Name string }
type PlaceholderNode struct{ Name string }
type LastNode struct{}
type EvalNode struct{ Argument interface{} } // For EVAL() function
type StringLiteralNode struct {
	Value string
	IsRaw bool
}
type NumberLiteralNode struct{ Value interface{} } // Holds int64 or float64
type BooleanLiteralNode struct{ Value bool }
type ListLiteralNode struct{ Elements []interface{} }
type MapEntryNode struct {
	Key   StringLiteralNode // Assuming StringLiteralNode is a concrete type
	Value interface{}
}
type MapLiteralNode struct{ Entries []MapEntryNode } // Assuming MapEntryNode is concrete
type ElementAccessNode struct {
	Collection interface{}
	Accessor   interface{}
}
type UnaryOpNode struct {
	Operator string
	Operand  interface{}
}
type BinaryOpNode struct {
	Left     interface{}
	Operator string
	Right    interface{}
}
type FunctionCallNode struct {
	FunctionName string
	Arguments    []interface{}
}

// --- Procedure & Step Structures (Revised for v0.2.0 based on on_error) ---

type Procedure struct {
	Name              string
	RequiredParams    []string // From 'needs' clause
	OptionalParams    []string // From 'optional' clause
	ReturnVarNames    []string // From 'returns' clause
	Steps             []Step
	OriginalSignature string
	// Add Metadata map if needed based on other files
	Metadata map[string]string
}

type Step struct {
	Type      string        // "set", "call", "return", "emit", "if", "while", "for", "must", "mustbe", "fail", "on_error", "clear_error" // Updated types
	Target    string        // Variable name (SET, FOR), Call target (CALL, mustbe)
	Cond      interface{}   // Condition expr (IF, WHILE), Collection expr (FOR)
	Value     interface{}   // RHS (SET), Return val(s) (RETURN), Emit val (EMIT), Must expr/call (MUST/MUSTBE), Fail expr (FAIL), Body steps []Step (IF-THEN, WHILE, FOR, ON_ERROR)
	ElseValue interface{}   // Else body steps []Step (IF)
	Args      []interface{} // Arguments (CALL)
	// CatchVar       string      // REMOVED
	// CatchSteps     []Step      // REMOVED
	// FinallySteps   []Step      // REMOVED
	SourceLineInfo string // Optional: Store original line number/content for debugging
	// Ensure Metadata field exists if Procedure has it and it's needed per-step
	Metadata map[string]string
}

// newStep helper function - revised for removed fields
func newStep(typ, target string, cond, value, elseValue interface{}, args []interface{}) Step {
	s := Step{
		Type:      typ,
		Target:    target,
		Cond:      cond,
		Value:     value,
		ElseValue: elseValue,
		Args:      args,
		// CatchVar, CatchSteps, FinallySteps removed
		Metadata: make(map[string]string), // Initialize if Metadata field exists
	}
	// Specific setting for fields like Value (for block steps) or Target (loop var)
	// happens in the respective ExitXxx listener methods.
	return s
}

// Ensure necessary concrete types like StringLiteralNode are defined if not already.
// If expression nodes are interfaces, ensure they conform to a common Node interface if needed.
// Add String() methods to nodes for debugging if helpful.
func (s StringLiteralNode) String() string {
	if s.IsRaw {
		return fmt.Sprintf("```%s```", s.Value)
	}
	return fmt.Sprintf("%q", s.Value)
}

// ... Add other String() methods for expression nodes as needed ...
