// NeuroScript Version: 0.5.2
// File version: 1
// Purpose: Contains all AST nodes that implement the Expression interface.
// filename: pkg/ast/ast_expressions.go
// nlines: 195
// risk_rating: MEDIUM

package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

// CallTarget represents the target of a function or tool call.
type CallTarget struct {
	Pos	*lang.Position
	IsTool	bool	// True if it's a tool call (e.g., tool.FS.Read)
	Name	string	// Fully qualified name (e.g., MyProcedure, FS.Read)
}

func (ct *CallTarget) GetPos() *lang.Position	{ return ct.Pos }
func (ct *CallTarget) String() string {
	if ct.IsTool {
		return "tool." + ct.Name
	}
	return ct.Name
}

// CallableExprNode represents a function or tool call expression.
type CallableExprNode struct {
	Pos		*lang.Position
	Target		CallTarget
	Arguments	[]Expression
}

func (n *CallableExprNode) GetPos() *lang.Position	{ return n.Pos }
func (n *CallableExprNode) expressionNode()		{}
func (n *CallableExprNode) String() string {
	args := make([]string, len(n.Arguments))
	for i, arg := range n.Arguments {
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
	Pos	*lang.Position
	Name	string
}

func (n *VariableNode) GetPos() *lang.Position	{ return n.Pos }
func (n *VariableNode) expressionNode()		{}
func (n *VariableNode) String() string		{ return n.Name }

// PlaceholderNode represents a placeholder like {{variable}} or {{LAST}}.
type PlaceholderNode struct {
	Pos	*lang.Position
	Name	string	// "LAST" or variable name
}

func (n *PlaceholderNode) GetPos() *lang.Position	{ return n.Pos }
func (n *PlaceholderNode) expressionNode()		{}
func (n *PlaceholderNode) String() string		{ return fmt.Sprintf("{{%s}}", n.Name) }

// LastNode represents the 'last' keyword.
type LastNode struct {
	Pos *lang.Position
}

func (n *LastNode) GetPos() *lang.Position	{ return n.Pos }
func (n *LastNode) expressionNode()		{}
func (n *LastNode) String() string		{ return "last" }

// EvalNode represents an eval(expression) call.
type EvalNode struct {
	Pos		*lang.Position
	Argument	Expression
}

func (n *EvalNode) GetPos() *lang.Position	{ return n.Pos }
func (n *EvalNode) expressionNode()		{}
func (n *EvalNode) String() string		{ return fmt.Sprintf("eval(%s)", n.Argument.String()) }

// StringLiteralNode represents a string literal.
type StringLiteralNode struct {
	Pos	*lang.Position
	Value	string
	IsRaw	bool	// True if triple-backtick string ```...```
}

func (n *StringLiteralNode) GetPos() *lang.Position	{ return n.Pos }
func (n *StringLiteralNode) expressionNode()		{}
func (n StringLiteralNode) String() string {
	if n.IsRaw {
		return "```" + n.Value + "```"
	}
	return strconv.Quote(n.Value)
}

// NumberLiteralNode represents a number literal (integer or float).
type NumberLiteralNode struct {
	Pos	*lang.Position
	Value	interface{}	// Stores int64 or float64
}

func (n *NumberLiteralNode) GetPos() *lang.Position	{ return n.Pos }
func (n *NumberLiteralNode) expressionNode()		{}
func (n *NumberLiteralNode) String() string		{ return fmt.Sprintf("%v", n.Value) }

// BooleanLiteralNode represents a boolean literal (true or false).
type BooleanLiteralNode struct {
	Pos	*lang.Position
	Value	bool
}

func (n *BooleanLiteralNode) GetPos() *lang.Position	{ return n.Pos }
func (n *BooleanLiteralNode) expressionNode()		{}
func (n *BooleanLiteralNode) String() string		{ return strconv.FormatBool(n.Value) }

// ListLiteralNode represents a list literal (e.g., [1, "two", true]).
type ListLiteralNode struct {
	Pos		*lang.Position
	Elements	[]Expression
}

func (n *ListLiteralNode) GetPos() *lang.Position	{ return n.Pos }
func (n *ListLiteralNode) expressionNode()		{}
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
	Pos	*lang.Position
	Key	*StringLiteralNode	// Keys are always string literals
	Value	Expression
}

func (n *MapEntryNode) GetPos() *lang.Position	{ return n.Pos }
func (n *MapEntryNode) String() string {
	if n.Key != nil && n.Value != nil {
		return fmt.Sprintf("%s: %s", n.Key.String(), n.Value.String())
	}
	return "<invalid_map_entry>"
}

// MapLiteralNode represents a map literal (e.g., {"key1": value1, "key2": value2}).
type MapLiteralNode struct {
	Pos	*lang.Position
	Entries	[]*MapEntryNode
}

func (n *MapLiteralNode) GetPos() *lang.Position	{ return n.Pos }
func (n *MapLiteralNode) expressionNode()		{}
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
	Pos		*lang.Position
	Collection	Expression	// The variable or expression yielding the collection
	Accessor	Expression	// The index or key expression
}

func (n *ElementAccessNode) GetPos() *lang.Position	{ return n.Pos }
func (n *ElementAccessNode) expressionNode()		{}
func (n *ElementAccessNode) String() string {
	return fmt.Sprintf("%s[%s]", n.Collection.String(), n.Accessor.String())
}

// UnaryOpNode represents a unary operation (e.g., -value, not flag).
type UnaryOpNode struct {
	Pos		*lang.Position
	Operator	string
	Operand		Expression
}

func (n *UnaryOpNode) GetPos() *lang.Position	{ return n.Pos }
func (n *UnaryOpNode) expressionNode()		{}
func (n *UnaryOpNode) String() string {
	return fmt.Sprintf("%s%s", n.Operator, n.Operand.String())
}

// BinaryOpNode represents a binary operation (e.g., left + right).
type BinaryOpNode struct {
	Pos		*lang.Position
	Left		Expression
	Operator	string
	Right		Expression
}

func (n *BinaryOpNode) GetPos() *lang.Position	{ return n.Pos }
func (n *BinaryOpNode) expressionNode()		{}
func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Operator, n.Right.String())
}

// TypeOfNode represents a typeof(expression) call.
type TypeOfNode struct {
	Pos		*lang.Position
	Argument	Expression
}

func (n *TypeOfNode) GetPos() *lang.Position	{ return n.Pos }
func (n *TypeOfNode) expressionNode()		{}
func (n *TypeOfNode) String() string		{ return fmt.Sprintf("typeof(%s)", n.Argument.String()) }

// NilLiteralNode represents the 'nil' literal.
type NilLiteralNode struct {
	Pos *lang.Position
}

func (n *NilLiteralNode) GetPos() *lang.Position	{ return n.Pos }
func (n *NilLiteralNode) expressionNode()		{}
func (n *NilLiteralNode) String() string		{ return "nil" }