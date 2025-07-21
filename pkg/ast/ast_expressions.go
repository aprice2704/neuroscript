// filename: pkg/ast/ast_expressions.go
// NeuroScript Version: 0.5.2
// File version: 3
// Purpose: Removed redundant Pos fields and GetPos methods from all expression nodes to unify position handling via BaseNode.
// nlines: 150+
// risk_rating: MEDIUM

package ast

import (
	"fmt"
	"strconv"
	"strings"
)

// CallTarget represents the target of a function or tool call.
type CallTarget struct {
	BaseNode
	IsTool bool   // True if it's a tool call (e.g., tool.FS.Read)
	Name   string // Fully qualified name (e.g., MyProcedure, FS.Read)
}

func (ct *CallTarget) String() string {
	if ct.IsTool {
		return "tool." + ct.Name
	}
	return ct.Name
}

// CallableExprNode represents a function or tool call expression.
type CallableExprNode struct {
	BaseNode
	Target    CallTarget
	Arguments []Expression
}

func (n *CallableExprNode) expressionNode() {}
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
	BaseNode
	Name string
}

func (n *VariableNode) expressionNode() {}
func (n *VariableNode) String() string  { return n.Name }

// PlaceholderNode represents a placeholder like {{variable}} or {{LAST}}.
type PlaceholderNode struct {
	BaseNode
	Name string // "LAST" or variable name
}

func (n *PlaceholderNode) expressionNode() {}
func (n *PlaceholderNode) String() string  { return fmt.Sprintf("{{%s}}", n.Name) }

// LastNode represents the 'last' keyword.
type LastNode struct {
	BaseNode
}

func (n *LastNode) expressionNode() {}
func (n *LastNode) String() string  { return "last" }

// EvalNode represents an eval(expression) call.
type EvalNode struct {
	BaseNode
	Argument Expression
}

func (n *EvalNode) expressionNode() {}
func (n *EvalNode) String() string  { return fmt.Sprintf("eval(%s)", n.Argument.String()) }

// StringLiteralNode represents a string literal.
type StringLiteralNode struct {
	BaseNode
	Value string
	IsRaw bool // True if triple-backtick string ```...```
}

func (n *StringLiteralNode) expressionNode() {}
func (n *StringLiteralNode) String() string {
	if n.IsRaw {
		return "```" + n.Value + "```"
	}
	return strconv.Quote(n.Value)
}

// NumberLiteralNode represents a number literal (integer or float).
type NumberLiteralNode struct {
	BaseNode
	Value interface{} // Stores int64 or float64
}

func (n *NumberLiteralNode) expressionNode() {}
func (n *NumberLiteralNode) String() string  { return fmt.Sprintf("%v", n.Value) }

// BooleanLiteralNode represents a boolean literal (true or false).
type BooleanLiteralNode struct {
	BaseNode
	Value bool
}

func (n *BooleanLiteralNode) expressionNode() {}
func (n *BooleanLiteralNode) String() string  { return strconv.FormatBool(n.Value) }

// ListLiteralNode represents a list literal (e.g., [1, "two", true]).
type ListLiteralNode struct {
	BaseNode
	Elements []Expression
}

func (n *ListLiteralNode) expressionNode() {}
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
	BaseNode
	Key   *StringLiteralNode // Keys are always string literals
	Value Expression
}

func (n *MapEntryNode) String() string {
	if n.Key != nil && n.Value != nil {
		return fmt.Sprintf("%s: %s", n.Key.String(), n.Value.String())
	}
	return "<invalid_map_entry>"
}

// MapLiteralNode represents a map literal (e.g., {"key1": value1, "key2": value2}).
type MapLiteralNode struct {
	BaseNode
	Entries []*MapEntryNode
}

func (n *MapLiteralNode) expressionNode() {}
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
	BaseNode
	Collection Expression // The variable or expression yielding the collection
	Accessor   Expression // The index or key expression
}

func (n *ElementAccessNode) expressionNode() {}
func (n *ElementAccessNode) String() string {
	return fmt.Sprintf("%s[%s]", n.Collection.String(), n.Accessor.String())
}

// UnaryOpNode represents a unary operation (e.g., -value, not flag).
type UnaryOpNode struct {
	BaseNode
	Operator string
	Operand  Expression
}

func (n *UnaryOpNode) expressionNode() {}
func (n *UnaryOpNode) String() string {
	return fmt.Sprintf("%s%s", n.Operator, n.Operand.String())
}

// BinaryOpNode represents a binary operation (e.g., left + right).
type BinaryOpNode struct {
	BaseNode
	Left     Expression
	Operator string
	Right    Expression
}

func (n *BinaryOpNode) expressionNode() {}
func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Operator, n.Right.String())
}

// TypeOfNode represents a typeof(expression) call.
type TypeOfNode struct {
	BaseNode
	Argument Expression
}

func (n *TypeOfNode) expressionNode() {}
func (n *TypeOfNode) String() string  { return fmt.Sprintf("typeof(%s)", n.Argument.String()) }

// NilLiteralNode represents the 'nil' literal.
type NilLiteralNode struct {
	BaseNode
}

func (n *NilLiteralNode) expressionNode() {}
func (n *NilLiteralNode) String() string  { return "nil" }
