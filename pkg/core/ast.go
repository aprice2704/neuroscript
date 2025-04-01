// pkg/core/ast.go
package core

// --- Expression Node Types ---
// We use interface{} for now to hold various node types.
// A more rigorous approach might use a dedicated ExpressionNode interface.

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
	// Elements will hold the evaluated AST nodes for each element expression
	Elements []interface{} // Slice of expression nodes (VariableNode, LiteralNode, etc.)
}

// Represents a single key-value pair within a map literal
type MapEntryNode struct {
	// Key is always a string literal node after parsing
	Key StringLiteralNode
	// Value holds the evaluated AST node for the value expression
	Value interface{} // An expression node
}

// Represents a map literal (e.g., {"key": val, "lit": "abc"})
type MapLiteralNode struct {
	Entries []MapEntryNode
}

// Represents a string concatenation using '+'
type ConcatenationNode struct {
	// Operands holds the evaluated AST nodes for the parts being concatenated
	Operands []interface{} // Slice of expression nodes
}

// --- Existing Structures (Modified) ---

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

// Step struct updated to hold specific node types for expressions/args/conditions
type Step struct {
	Type   string        // "SET", "CALL", "IF", "WHILE", "FOR", "RETURN", "EMIT" etc.
	Target string        // Variable name (SET, FOR loop var), Procedure/LLM/Tool name (CALL)
	Value  interface{}   // Body []Step (IF/WHILE/FOR), Expression Node (SET, RETURN, EMIT)
	Args   []interface{} // Slice of Expression Nodes (CALL arguments)
	Cond   interface{}   // Expression Node (IF, WHILE condition), Expression Node (FOR collection)
}

// newStep creates a new Step instance - signature updated slightly for clarity/consistency
// Note: We might adjust how steps are created in the listener based on node types
func newStep(typ string, target string, condNode interface{}, valueNode interface{}, argNodes []interface{}) Step {
	return Step{Type: typ, Target: target, Cond: condNode, Value: valueNode, Args: argNodes}
}
