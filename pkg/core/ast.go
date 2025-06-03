// NeuroScript Version: 0.5.2
// File version: 0.1.0
// Purpose: Modify AST nodes to support complex lvalues for indexed/field assignments.
// filename: pkg/core/ast.go
// nlines: 270
// risk_rating: HIGH

package core

import (
	"fmt"
	"strconv"
	"strings"
)

// Position represents a location in the source code.
type Position struct {
	Line   int
	Column int
	File   string // Optional: filename or source identifier
}

func (p *Position) GetPos() *Position { return p }
func (p *Position) String() string {
	if p == nil {
		return "(unknown pos)"
	}
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("line %d, col %d", p.Line, p.Column)
}

// Expression is an interface for all expression AST nodes.
type Expression interface {
	GetPos() *Position
	expressionNode() // Marker method
	String() string  // For debugging and string representation
}

// Program represents the entire parsed NeuroScript program.
type Program struct {
	Pos        *Position
	Metadata   map[string]string
	Procedures map[string]*Procedure
}

func (p *Program) GetPos() *Position { return p.Pos }

// ErrorNode represents a parsing or semantic error encountered during AST construction.
type ErrorNode struct {
	Pos     *Position
	Message string
}

func (n *ErrorNode) GetPos() *Position { return n.Pos }
func (n *ErrorNode) expressionNode()   {}
func (n *ErrorNode) String() string    { return fmt.Sprintf("Error(%s): %s", n.Pos, n.Message) }

// CallTarget represents the target of a function or tool call.
type CallTarget struct {
	Pos    *Position
	IsTool bool   // True if it's a tool call (e.g., tool.FS.Read)
	Name   string // Fully qualified name (e.g., MyProcedure, FS.Read)
}

func (ct *CallTarget) GetPos() *Position { return ct.Pos }
func (ct *CallTarget) String() string {
	if ct.IsTool {
		return "tool." + ct.Name
	}
	return ct.Name
}

// CallableExprNode represents a function or tool call expression.
type CallableExprNode struct {
	Pos       *Position
	Target    CallTarget
	Arguments []Expression
}

func (n *CallableExprNode) GetPos() *Position { return n.Pos }
func (n *CallableExprNode) expressionNode()   {}
func (n *CallableExprNode) String() string {
	args := make([]string, len(n.Arguments))
	for i, arg := range n.Arguments {
		// Ensure arg is not nil to prevent panic with %v
		if arg != nil {
			args[i] = arg.String()
		} else {
			args[i] = "<nil_expr>"
		}
	}
	return fmt.Sprintf("%s(%s)", n.Target.String(), strings.Join(args, ", "))
}

// VariableNode represents a variable reference.
type VariableNode struct {
	Pos  *Position
	Name string
}

func (n *VariableNode) GetPos() *Position { return n.Pos }
func (n *VariableNode) expressionNode()   {}
func (n *VariableNode) String() string    { return n.Name }

// PlaceholderNode represents a placeholder like {{variable}} or {{LAST}}.
type PlaceholderNode struct {
	Pos  *Position
	Name string // "LAST" or variable name
}

func (n *PlaceholderNode) GetPos() *Position { return n.Pos }
func (n *PlaceholderNode) expressionNode()   {}
func (n *PlaceholderNode) String() string    { return fmt.Sprintf("{{%s}}", n.Name) }

// LastNode represents the 'last' keyword.
type LastNode struct {
	Pos *Position
}

func (n *LastNode) GetPos() *Position { return n.Pos }
func (n *LastNode) expressionNode()   {}
func (n *LastNode) String() string    { return "last" }

// EvalNode represents an eval(expression) call.
type EvalNode struct {
	Pos      *Position
	Argument Expression
}

func (n *EvalNode) GetPos() *Position { return n.Pos }
func (n *EvalNode) expressionNode()   {}
func (n *EvalNode) String() string    { return fmt.Sprintf("eval(%s)", n.Argument.String()) }

// StringLiteralNode represents a string literal.
type StringLiteralNode struct {
	Pos   *Position
	Value string
	IsRaw bool // True if triple-backtick string ```...```
}

func (n *StringLiteralNode) GetPos() *Position { return n.Pos }
func (n *StringLiteralNode) expressionNode()   {}
func (n StringLiteralNode) String() string {
	if n.IsRaw {
		return "```" + n.Value + "```"
	}
	return strconv.Quote(n.Value)
}

// NumberLiteralNode represents a number literal (integer or float).
type NumberLiteralNode struct {
	Pos   *Position
	Value interface{} // Stores int64 or float64
}

func (n *NumberLiteralNode) GetPos() *Position { return n.Pos }
func (n *NumberLiteralNode) expressionNode()   {}
func (n *NumberLiteralNode) String() string    { return fmt.Sprintf("%v", n.Value) }

// BooleanLiteralNode represents a boolean literal (true or false).
type BooleanLiteralNode struct {
	Pos   *Position
	Value bool
}

func (n *BooleanLiteralNode) GetPos() *Position { return n.Pos }
func (n *BooleanLiteralNode) expressionNode()   {}
func (n *BooleanLiteralNode) String() string    { return strconv.FormatBool(n.Value) }

// ListLiteralNode represents a list literal (e.g., [1, "two", true]).
type ListLiteralNode struct {
	Pos      *Position
	Elements []Expression
}

func (n *ListLiteralNode) GetPos() *Position { return n.Pos }
func (n *ListLiteralNode) expressionNode()   {}
func (n *ListLiteralNode) String() string {
	elems := make([]string, len(n.Elements))
	for i, el := range n.Elements {
		if el != nil {
			elems[i] = el.String()
		} else {
			elems[i] = "<nil_expr>"
		}
	}
	return "[" + strings.Join(elems, ", ") + "]"
}

// MapEntryNode represents a single key-value pair in a map literal.
type MapEntryNode struct {
	Pos   *Position
	Key   *StringLiteralNode // Keys are always string literals
	Value Expression
}

func (n *MapEntryNode) GetPos() *Position { return n.Pos }
func (n *MapEntryNode) String() string {
	if n.Key != nil && n.Value != nil {
		return fmt.Sprintf("%s: %s", n.Key.String(), n.Value.String())
	}
	return "<invalid_map_entry>"
}

// MapLiteralNode represents a map literal (e.g., {"key1": value1, "key2": value2}).
type MapLiteralNode struct {
	Pos     *Position
	Entries []*MapEntryNode
}

func (n *MapLiteralNode) GetPos() *Position { return n.Pos }
func (n *MapLiteralNode) expressionNode()   {}
func (n *MapLiteralNode) String() string {
	entries := make([]string, len(n.Entries))
	for i, entry := range n.Entries {
		if entry != nil {
			entries[i] = entry.String()
		} else {
			entries[i] = "<nil_entry>"
		}
	}
	return "{" + strings.Join(entries, ", ") + "}"
}

// ElementAccessNode represents accessing an element of a list or map (e.g., myList[0], myMap["key"]).
type ElementAccessNode struct {
	Pos        *Position
	Collection Expression // The variable or expression yielding the collection
	Accessor   Expression // The index or key expression
}

func (n *ElementAccessNode) GetPos() *Position { return n.Pos }
func (n *ElementAccessNode) expressionNode()   {}
func (n *ElementAccessNode) String() string {
	return fmt.Sprintf("%s[%s]", n.Collection.String(), n.Accessor.String())
}

// UnaryOpNode represents a unary operation (e.g., -value, not flag).
type UnaryOpNode struct {
	Pos      *Position
	Operator string
	Operand  Expression
}

func (n *UnaryOpNode) GetPos() *Position { return n.Pos }
func (n *UnaryOpNode) expressionNode()   {}
func (n *UnaryOpNode) String() string {
	return fmt.Sprintf("%s%s", n.Operator, n.Operand.String())
}

// BinaryOpNode represents a binary operation (e.g., left + right).
type BinaryOpNode struct {
	Pos      *Position
	Left     Expression
	Operator string
	Right    Expression
}

func (n *BinaryOpNode) GetPos() *Position { return n.Pos }
func (n *BinaryOpNode) expressionNode()   {}
func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Operator, n.Right.String())
}

// TypeOfNode represents a typeof(expression) call.
type TypeOfNode struct {
	Pos      *Position
	Argument Expression
}

func (n *TypeOfNode) GetPos() *Position { return n.Pos }
func (n *TypeOfNode) expressionNode()   {}
func (n *TypeOfNode) String() string    { return fmt.Sprintf("typeof(%s)", n.Argument.String()) }

// NilLiteralNode represents the 'nil' literal.
type NilLiteralNode struct {
	Pos *Position
}

func (n *NilLiteralNode) GetPos() *Position { return n.Pos }
func (n *NilLiteralNode) expressionNode()   {}
func (n *NilLiteralNode) String() string    { return "nil" }

// AccessorType distinguishes between bracket and dot access in an LValue
type AccessorType int

const (
	BracketAccess AccessorType = iota // e.g., a[expression]
	DotAccess                         // e.g., a.field
)

// AccessorNode represents one part of an lvalue's accessor chain (e.g., "[index]" or ".field")
type AccessorNode struct {
	Pos        *Position
	Type       AccessorType
	IndexOrKey Expression // For BracketAccess (LBRACK expression RBRACK)
	FieldName  string     // For DotAccess (DOT IDENTIFIER)
}

func (an *AccessorNode) String() string {
	if an.Type == BracketAccess {
		return fmt.Sprintf("[%s]", an.IndexOrKey.String())
	}
	return fmt.Sprintf(".%s", an.FieldName)
}

// LValueNode represents the left-hand side of an assignment that can be complex
type LValueNode struct {
	Pos        *Position
	Identifier string         // The base variable name (e.g., 'a' in a["key"])
	Accessors  []AccessorNode // Sequence of bracket or dot accessors
}

func (n *LValueNode) GetPos() *Position { return n.Pos }
func (n *LValueNode) String() string {
	var sb strings.Builder
	sb.WriteString(n.Identifier)
	for _, acc := range n.Accessors {
		sb.WriteString(acc.String())
	}
	return sb.String()
}

// ParamSpec defines a parameter in a procedure signature.
type ParamSpec struct {
	Name         string
	DefaultValue interface{} // For optional parameters
}

// Procedure represents a user-defined function.
type Procedure struct {
	Pos               *Position
	Name              string
	RequiredParams    []string
	OptionalParams    []ParamSpec // Name and default value
	Variadic          bool        // If the last param is variadic
	VariadicParamName string
	ReturnVarNames    []string // Names of variables to return
	Steps             []Step
	OriginalSignature string            // For debugging/LSP
	Metadata          map[string]string // Procedure-level metadata
}

func (p *Procedure) GetPos() *Position { return p.Pos }

// Step represents a single statement or control flow structure in a procedure.
// Note: The structure of Step for "set" statements now uses LValue.
// Other parts of the codebase (AST builder, interpreter) will need to be updated to use step.LValue.
type Step struct {
	Pos  *Position
	Type string // e.g., "set", "call", "if", "return", "emit", "must", "fail", "clear_error", "ask", "on_error"

	// For "set":
	LValue *LValueNode // Target of the assignment, can be complex (e.g., var[index].field)
	Value  Expression  // RHS expression for set

	// For "call":
	Call *CallableExprNode // Details of the function/tool call

	// For "if", "while", "must" (conditional variant):
	Cond Expression // Condition expression

	// For "if", "while", "for", "on_error":
	Body []Step // Block of steps

	// For "if":
	Else []Step // Else block for if statements

	// For "for each":
	LoopVarName string     // Variable for each item in the loop (e.g., 'item' in 'for each item in myList')
	Collection  Expression // The collection expression to iterate over

	// For "ask":
	PromptExpr Expression // The prompt expression for 'ask'
	AskIntoVar string     // Optional variable to store the result of 'ask'

	// For "return", "emit", "fail", "must" (unconditional variant):
	// Value Expression (already defined above) is used for these.
	// For "return" specifically, if multiple values are returned, this might be a ListLiteralNode or similar.
	// The grammar supports `return_statement: KW_RETURN expression_list?`,
	// so `Value` might hold a single Expression or an ExpressionList equivalent (e.g. ListLiteralNode).
	// For simplicity, we'll assume Value holds the primary expression, and multiple returns are packed by the builder if needed.
	// Let's keep Values []Expression if it was there before for multiple return values.
	Values []Expression // For return statements that might return multiple values (if grammar supports it explicitly)

}

func (s *Step) GetPos() *Position { return s.Pos }

// --- Helper for getting position from any node that implements Expression or Step ---
func getExpressionPosition(val interface{}) *Position {
	if expr, ok := val.(Expression); ok {
		return expr.GetPos()
	}
	if step, ok := val.(Step); ok {
		return step.GetPos()
	}
	return nil
}
