// NeuroScript Version: 0.5.2
// File version: 2
// Purpose: Consolidated file for all AST node structs and their helper types.
// filename: pkg/ast/ast_nodes.go
// nlines: 320
// risk_rating: HIGH

package ast

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// NOTE: This file exceeds the 200-line limit specified in AI_RULES.md.
// This was done to follow the explicit request to consolidate all node
// definitions into a single file during this refactoring phase.

// --- Root Node ---

// Program represents the entire parsed NeuroScript program. It is the root of the AST.
type Program struct {
	BaseNode
	Metadata    map[string]string
	Procedures  map[string]*Procedure
	Events      []*OnEventDecl
	Expressions []Expression
	Commands    []*CommandNode
}

// Kind returns the specific type for this node.
func (n *Program) Kind() Kind { return KindProgram }

// NewProgram creates and initializes a new Program node.
func NewProgram() *Program {
	return &Program{
		Metadata:    make(map[string]string),
		Procedures:  make(map[string]*Procedure),
		Events:      make([]*OnEventDecl, 0),
		Commands:    make([]*CommandNode, 0),
		Expressions: make([]Expression, 0),
	}
}

// --- Top-Level Declarations ---

// CommandNode represents a single 'command ... endcommand' block.
type CommandNode struct {
	BaseNode
	Metadata      map[string]string
	Body          []Step
	ErrorHandlers []*Step
}

// Kind returns the specific type for this node.
func (n *CommandNode) Kind() Kind { return KindCommandNode }
func (n *CommandNode) String() string {
	return "command ... endcommand"
}

// OnEventDecl represents a top-level 'on event ...' declaration.
type OnEventDecl struct {
	BaseNode
	EventNameExpr Expression
	HandlerName   string
	EventVarName  string
	Body          []Step
}

// Kind returns the specific type for this node.
func (n *OnEventDecl) Kind() Kind { return KindOnEventDecl }

// Procedure represents a declared function.
type Procedure struct {
	BaseNode
	name              string
	Metadata          map[string]string
	RequiredParams    []string
	OptionalParams    []*ParamSpec
	Variadic          bool
	VariadicParamName string
	ReturnVarNames    []string
	ErrorHandlers     []*Step
	Steps             []Step
}

// Kind returns the specific type for this node.
func (p *Procedure) Kind() Kind { return KindProcedure }

// Name returns the procedure's name.
func (p *Procedure) Name() string { return p.name }

// SetName sets the procedure's name.
func (p *Procedure) SetName(name string) { p.name = name }

// IsCallable marks this as a callable type for the interpreter.
func (p *Procedure) IsCallable() {}

// ParamSpec defines a parameter for a procedure. It is not an AST node.
type ParamSpec struct {
	Name    string
	Default lang.Value
}

// --- Statements & Steps ---

// Step represents a single statement or instruction within a procedure's body.
type Step struct {
	BaseNode
	Type           string
	LValues        []*LValueNode
	Values         []Expression
	Cond           Expression
	Body           []Step
	ElseBody       []Step
	LoopVarName    string
	IndexVarName   string
	Collection     Expression
	Call           *CallableExprNode
	OnEvent        *OnEventDecl
	AskIntoVar     string
	IsFinal        bool
	ErrorName      string
	tool           interfaces.Tool
	ExpressionStmt *ExpressionStatementNode
}

// Kind returns the specific type for this node.
func (s *Step) Kind() Kind { return KindStep }

// ExpressionStatementNode represents a statement that consists of a single expression.
type ExpressionStatementNode struct {
	BaseNode
	Expression Expression
}

// Kind returns the specific type for this node.
func (n *ExpressionStatementNode) Kind() Kind { return KindExpressionStmt }
func (n *ExpressionStatementNode) String() string {
	if n.Expression != nil {
		return n.Expression.String()
	}
	return "<nil_expr_stmt>"
}

// --- Expressions ---

// Expression is an interface for all expression AST nodes.
type Expression interface {
	Node
	expressionNode() // Marker method
	String() string
}

// LValueNode represents a "left-value" in an assignment.
type LValueNode struct {
	BaseNode
	Identifier string
	Accessors  []*AccessorNode
}

func (n *LValueNode) Kind() Kind      { return KindLValue }
func (n *LValueNode) expressionNode() {}
func (n *LValueNode) String() string {
	return n.Identifier
}

// AccessorNode represents a single part of an element access chain (e.g., `[key]` or `.field`).
type AccessorNode struct {
	BaseNode
	Type       AccessorType
	Key        Expression
	IsOptional bool
}

// AccessorType defines how an element is accessed. It is not an AST node.
type AccessorType int

const (
	BracketAccess AccessorType = iota
	DotAccess
)

// CallableExprNode represents a function or tool call expression.
type CallableExprNode struct {
	BaseNode
	Target    CallTarget
	Arguments []Expression
}

func (n *CallableExprNode) Kind() Kind      { return KindCallableExpr }
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

// CallTarget represents the target of a function or tool call. It is not an AST node.
type CallTarget struct {
	BaseNode
	IsTool bool
	Name   string
}

func (ct *CallTarget) String() string {
	if ct.IsTool {
		return "tool." + ct.Name
	}
	return ct.Name
}

// VariableNode represents a variable reference.
type VariableNode struct {
	BaseNode
	Name string
}

func (n *VariableNode) Kind() Kind      { return KindVariable }
func (n *VariableNode) expressionNode() {}
func (n *VariableNode) String() string  { return n.Name }

// StringLiteralNode represents a string literal.
type StringLiteralNode struct {
	BaseNode
	Value string
	IsRaw bool
}

func (n *StringLiteralNode) Kind() Kind      { return KindStringLiteral }
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
	Value interface{}
}

func (n *NumberLiteralNode) Kind() Kind      { return KindNumberLiteral }
func (n *NumberLiteralNode) expressionNode() {}
func (n *NumberLiteralNode) String() string  { return fmt.Sprintf("%v", n.Value) }

// BooleanLiteralNode represents a boolean literal (true or false).
type BooleanLiteralNode struct {
	BaseNode
	Value bool
}

func (n *BooleanLiteralNode) Kind() Kind      { return KindBooleanLiteral }
func (n *BooleanLiteralNode) expressionNode() {}
func (n *BooleanLiteralNode) String() string  { return strconv.FormatBool(n.Value) }

// NilLiteralNode represents the 'nil' literal.
type NilLiteralNode struct {
	BaseNode
}

func (n *NilLiteralNode) Kind() Kind      { return KindNilLiteral }
func (n *NilLiteralNode) expressionNode() {}
func (n *NilLiteralNode) String() string  { return "nil" }

// ListLiteralNode represents a list literal (e.g., [1, "two", true]).
type ListLiteralNode struct {
	BaseNode
	Elements []Expression
}

func (n *ListLiteralNode) Kind() Kind      { return KindListLiteral }
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

// MapLiteralNode represents a map literal (e.g., {"key1": value1, "key2": value2}).
type MapLiteralNode struct {
	BaseNode
	Entries []*MapEntryNode
}

func (n *MapLiteralNode) Kind() Kind      { return KindMapLiteral }
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

// MapEntryNode represents a single key-value pair in a map literal.
type MapEntryNode struct {
	BaseNode
	Key   *StringLiteralNode
	Value Expression
}

func (n *MapEntryNode) Kind() Kind { return KindMapEntry }
func (n *MapEntryNode) String() string {
	if n.Key != nil && n.Value != nil {
		return fmt.Sprintf("%s: %s", n.Key.String(), n.Value.String())
	}
	return "<invalid_map_entry>"
}

// ElementAccessNode represents accessing an element of a list or map (e.g., myList[0]).
type ElementAccessNode struct {
	BaseNode
	Collection Expression
	Accessor   Expression
}

func (n *ElementAccessNode) Kind() Kind      { return KindElementAccess }
func (n *ElementAccessNode) expressionNode() {}
func (n *ElementAccessNode) String() string {
	return fmt.Sprintf("%s[%s]", n.Collection.String(), n.Accessor.String())
}

// BinaryOpNode represents a binary operation (e.g., left + right).
type BinaryOpNode struct {
	BaseNode
	Left     Expression
	Operator string
	Right    Expression
}

func (n *BinaryOpNode) Kind() Kind      { return KindBinaryOp }
func (n *BinaryOpNode) expressionNode() {}
func (n *BinaryOpNode) String() string {
	return fmt.Sprintf("(%s %s %s)", n.Left.String(), n.Operator, n.Right.String())
}

// --- Miscellaneous Nodes ---

// ErrorNode captures a parsing or semantic error.
type ErrorNode struct {
	BaseNode
	Message string
}

func (n *ErrorNode) Kind() Kind      { return KindError }
func (n *ErrorNode) expressionNode() {}
func (n *ErrorNode) String() string {
	if n == nil {
		return "<nil error node>"
	}
	return fmt.Sprintf("Error at %s: %s", n.Pos(), n.Message)
}

// MetadataLine represents a single `:: key: value` line. It is not an AST node.
type MetadataLine struct {
	Pos   *Position
	Key   string
	Value string
}
