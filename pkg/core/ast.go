// pkg/core/ast.go
package core

// --- Expression Node Types (Revised for Arithmetic/Boolean/Functions) ---

// Basic Value Nodes (Unchanged)
type VariableNode struct{ Name string }
type PlaceholderNode struct{ Name string }   // Represents {{...}} syntax - value is raw name
type LastNode struct{}                       // Represents LAST keyword
type EvalNode struct{ Argument interface{} } // Represents EVAL(expression)

// Literal Nodes (Unchanged)
type StringLiteralNode struct{ Value string }
type NumberLiteralNode struct{ Value interface{} } // Holds int64 or float64
type BooleanLiteralNode struct{ Value bool }
type ListLiteralNode struct{ Elements []interface{} }
type MapEntryNode struct {
	Key   StringLiteralNode // Key must be string literal
	Value interface{}
}
type MapLiteralNode struct{ Entries []MapEntryNode }

// Access Node (Unchanged)
type ElementAccessNode struct {
	Collection interface{} // List or Map node/variable
	Accessor   interface{} // Index (int) or Key (string) expression node
}

// --- NEW/REVISED Operation Nodes ---

// UnaryOpNode: Represents prefix operations like negation (-) or NOT.
type UnaryOpNode struct {
	Operator string // e.g., "-", "NOT"
	Operand  interface{}
}

// BinaryOpNode: Represents all infix binary operations (+, -, *, /, %, **, ==, !=, <, >, <=, >=, AND, OR, &, |, ^).
// Replaces previous ConcatenationNode and ComparisonNode.
type BinaryOpNode struct {
	Left     interface{}
	Operator string // e.g., "+", "*", "==", "AND", "&"
	Right    interface{}
}

// FunctionCallNode: Represents built-in function calls like SIN(x), LN(x).
type FunctionCallNode struct {
	FunctionName string        // e.g., "SIN", "LN"
	Arguments    []interface{} // List of argument expression nodes
}

// --- REMOVED ConcatenationNode and ComparisonNode ---

// --- Procedure & Step Structures ---

// Docstring: Includes LangVersion (Unchanged from previous update)
type Docstring struct {
	Purpose, Output, Algorithm, Caveats, Examples string
	LangVersion                                   string
	InputLines                                    []string
	Inputs                                        map[string]string
}
type Procedure struct {
	Name      string
	Params    []string
	Docstring Docstring
	Steps     []Step
}

// Step: Value, Cond, ElseValue, Args can now hold new node types like BinaryOpNode etc.
type Step struct {
	Type, Target           string
	Value, ElseValue, Cond interface{}   // e.g., Cond might be a BinaryOpNode (a == b) or UnaryOpNode (NOT flag)
	Args                   []interface{} // Arguments for CALL (remain expression nodes)
}

// newStep helper (Unchanged)
func newStep(typ, target string, cond, value, elseValue interface{}, args []interface{}) Step {
	return Step{Type: typ, Target: target, Cond: cond, Value: value, ElseValue: elseValue, Args: args}
}
