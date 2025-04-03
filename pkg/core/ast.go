// pkg/core/ast.go
package core

// --- Expression Node Types ---

// Represents a variable reference (e.g., my_var)
type VariableNode struct {
	Name string
}

// Represents a placeholder (e.g., {{my_placeholder}})
type PlaceholderNode struct {
	Name string // The name inside the braces
}

// Represents the special __last_call_result keyword
type LastCallResultNode struct{}

// Represents a string literal (stores unquoted value)
type StringLiteralNode struct {
	Value string
}

// Represents a number literal (int64 or float64)
type NumberLiteralNode struct {
	Value interface{} // Holds int64 or float64
}

// Represents a boolean literal (true or false)
type BooleanLiteralNode struct {
	Value bool
}

// Represents a list literal (e.g., [1, "a", {{v}}])
type ListLiteralNode struct {
	Elements []interface{} // Slice of expression nodes
}

// Represents a single key-value pair within a map literal
type MapEntryNode struct {
	Key   StringLiteralNode // Key is always a string literal node
	Value interface{}       // An expression node for the value
}

// Represents a map literal (e.g., {"key": val, "lit": "abc"})
type MapLiteralNode struct {
	Entries []MapEntryNode
}

// Represents a string concatenation using '+'
type ConcatenationNode struct {
	Operands []interface{} // Slice of expression nodes
}

// Represents a comparison operation
type ComparisonNode struct {
	Left     interface{} // Left-hand side expression node
	Operator string      // Operator symbol (e.g., "==", "!=", ">")
	Right    interface{} // Right-hand side expression node
}

// Represents Element Access e.g., my_list[index_expr] or my_map[key_expr]
type ElementAccessNode struct {
	Collection interface{} // Expression node evaluating to the list/map (e.g., VariableNode, ListLiteralNode)
	Accessor   interface{} // Expression node evaluating to the index/key
}

// --- Existing Structures ---

// Docstring sections remain the same
type Docstring struct {
	Purpose    string
	InputLines []string
	Inputs     map[string]string
	Output     string
	Algorithm  string
	Caveats    string
	Examples   string
}

// Procedure definition remains the same
type Procedure struct {
	Name      string
	Params    []string
	Docstring Docstring
	Steps     []Step
}

// Step struct updated for ELSE block
type Step struct {
	Type      string        // "SET", "CALL", "IF", "WHILE", "FOR", "RETURN", "EMIT" etc.
	Target    string        // Variable name (SET, FOR loop var), Procedure/LLM/Tool name (CALL)
	Value     interface{}   // Body []Step (IF/WHILE/FOR THEN-block), Expression Node (SET, RETURN, EMIT)
	ElseValue interface{}   // *** ADDED: Body []Step for ELSE block ***
	Args      []interface{} // Slice of Expression Nodes (CALL arguments)
	Cond      interface{}   // Expression Node OR ComparisonNode (IF, WHILE), Expression Node (FOR collection)
}

// newStep creates a new Step instance (updated for ElseValue)
func newStep(typ string, target string, condNode interface{}, valueNode interface{}, elseValueNode interface{}, argNodes []interface{}) Step {
	return Step{Type: typ, Target: target, Cond: condNode, Value: valueNode, ElseValue: elseValueNode, Args: argNodes}
}
