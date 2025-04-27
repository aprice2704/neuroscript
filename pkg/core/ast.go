// pkg/core/ast.go
package core

// --- Expression Node Types (Unchanged from previous version) ---
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
	Key   StringLiteralNode
	Value interface{}
}
type MapLiteralNode struct{ Entries []MapEntryNode }
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

// --- Procedure & Step Structures (Revised for v0.2.0) ---

type Procedure struct {
	Name              string
	RequiredParams    []string // From 'needs' clause
	OptionalParams    []string // From 'optional' clause
	ReturnVarNames    []string // From 'returns' clause
	Steps             []Step
	OriginalSignature string
}

type Step struct {
	Type           string        // "set", "call", "return", "emit", "if", "while", "for", "must", "mustbe", "try", "fail"
	Target         string        // Variable name (SET, FOR), Call target (CALL, mustbe), Catch var (CATCH)
	Cond           interface{}   // Condition expr (IF, WHILE), Collection expr (FOR)
	Value          interface{}   // RHS (SET), Return val (RETURN), Emit val (EMIT), Must expr/call (MUST/MUSTBE), Body steps []Step (IF-THEN, WHILE, FOR, TRY)
	ElseValue      interface{}   // Else body steps []Step (IF)
	Args           []interface{} // Arguments (CALL)
	CatchVar       string        // Variable name for caught error in CATCH block
	CatchSteps     []Step        // Steps for the CATCH block
	FinallySteps   []Step        // Steps for the FINALLY block
	SourceLineInfo string        // Optional: Store original line number/content for debugging
}

// newStep helper function - ensures correct assignment for Args
func newStep(typ, target string, cond, value, elseValue interface{}, args []interface{}) Step {
	// Signature correctly takes args as []interface{}
	s := Step{
		Type:         typ,
		Target:       target,
		Cond:         cond,
		Value:        value,
		ElseValue:    elseValue,
		Args:         args, // Direct assignment is correct here
		CatchVar:     "",   // Initialize explicitly
		CatchSteps:   nil,  // Initialize explicitly
		FinallySteps: nil,  // Initialize explicitly
	}
	// Specific setting for CatchVar, CatchSteps, FinallySteps done in ExitTry_statement listener method.
	return s
}
