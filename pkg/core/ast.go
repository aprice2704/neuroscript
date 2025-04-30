// filename: pkg/core/ast.go
package core

import "fmt" // Ensure fmt is imported if needed by String() methods etc.

// --- Program Root Node ---

// Program represents the root of the NeuroScript Abstract Syntax Tree (AST).
type Program struct {
	Metadata   map[string]string // Stores file-level metadata (:: key: value)
	Procedures []Procedure       // List of procedures defined in the script
}

// --- Expression Node Types ---
// Using concrete types for expression nodes

// Represents the target of a function or tool call
type CallTarget struct {
	IsTool bool   // True if the target starts with 'tool.'
	Name   string // The name of the procedure or tool (e.g., "myFunc", "ReadFile")
}

// Node representing a function/tool call within an expression
type CallableExprNode struct {
	Target    CallTarget    // The function or tool being called
	Arguments []interface{} // Evaluated arguments (expression nodes)
}

// --- REMOVED old FunctionCallNode (replaced by CallableExprNode) ---
// type FunctionCallNode struct {
// 	FunctionName string
// 	Arguments    []interface{}
// }

type VariableNode struct{ Name string }
type PlaceholderNode struct{ Name string }
type LastNode struct{}
type EvalNode struct{ Argument interface{} } // For EVAL() function
type StringLiteralNode struct {
	Value string
	IsRaw bool // True for ```...``` strings
}
type NumberLiteralNode struct{ Value interface{} } // Holds int64 or float64
type BooleanLiteralNode struct{ Value bool }
type ListLiteralNode struct{ Elements []interface{} } // Elements are expression nodes
type MapEntryNode struct {
	Key   StringLiteralNode // Keys must be string literals currently
	Value interface{}       // Value is an expression node
}
type MapLiteralNode struct{ Entries []MapEntryNode }
type ElementAccessNode struct {
	Collection interface{} // Expression node yielding collection
	Accessor   interface{} // Expression node yielding index/key
}
type UnaryOpNode struct {
	Operator string      // e.g., "-", "not", "no", "some"
	Operand  interface{} // Expression node
}
type BinaryOpNode struct {
	Left     interface{} // Expression node
	Operator string      // e.g., "+", "==", "and"
	Right    interface{} // Expression node
}

// --- Procedure & Step Structures ---

type Procedure struct {
	Name              string
	RequiredParams    []string          // From 'needs' clause
	OptionalParams    []string          // From 'optional' clause
	ReturnVarNames    []string          // From 'returns' clause
	Steps             []Step            // The sequence of operations
	OriginalSignature string            // Original text for reference/debugging
	Metadata          map[string]string // Metadata defined within the procedure (:: key: value)
}

// Step represents a single operation within a procedure.
// Fields are used based on the 'Type'.
type Step struct {
	Type string // "set", "return", "emit", "if", "while", "for", "must", "mustbe", "fail", "on_error", "clear_error"
	// --- Field Usage by Type ---
	Target    string        // Variable name (set, for), Check function (mustbe)
	Cond      interface{}   // Condition expression node (if, while), Collection expression node (for)
	Value     interface{}   // RHS expression node (set), Return value(s) []expression node (return), Emit value expression node (emit), Must condition expression node (must), Fail expression node (fail), Body steps []Step (if-then, while, for, on_error)
	ElseValue interface{}   // Else body steps []Step (if)
	Args      []interface{} // REMOVED: Arguments were part of CallStep, now calls are expressions. Keep for compatibility? Let's remove.
	// OriginalSignature string        // Optional: Store original line number/content? Maybe add later.
	Metadata map[string]string // Metadata associated with this step (future use)
}

// newStep helper function - REMOVED Args parameter
func newStep(typ, target string, cond, value, elseValue interface{}) Step {
	return Step{
		Type:      typ,
		Target:    target,
		Cond:      cond,
		Value:     value,
		ElseValue: elseValue,
		Metadata:  make(map[string]string), // Initialize metadata map
	}
}

// String() methods for debugging (Example)
func (s StringLiteralNode) String() string {
	if s.IsRaw {
		return fmt.Sprintf("```%s```", s.Value)
	}
	return fmt.Sprintf("%q", s.Value)
}

func (c CallableExprNode) String() string {
	targetStr := c.Target.Name
	if c.Target.IsTool {
		targetStr = "tool." + targetStr
	}
	return fmt.Sprintf("%s(... %d args ...)", targetStr, len(c.Arguments))
}

// ... Add other String() methods for expression nodes as needed ...
