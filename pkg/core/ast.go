// filename: pkg/core/ast.go
package core

import (
	"fmt"
	"log" // Import log for the helper function warning
)

// --- Position Information ---

// Position represents a location in the source file.
type Position struct {
	Line   int
	Column int
	File   string // Optional: File name if handling multiple files
}

// GetPos returns the position pointer. Useful for satisfying interfaces.
func (p *Position) GetPos() *Position {
	return p
}

// String returns a string representation like "Line 10, Col 5".
func (p *Position) String() string {
	if p == nil {
		return "<nil position>"
	}
	// TODO: Add File if available
	return fmt.Sprintf("Line %d, Col %d", p.Line, p.Column)
}

// --- Program Root Node ---

// Program represents the root of the NeuroScript Abstract Syntax Tree (AST).
type Program struct {
	Metadata   map[string]string // Stores file-level metadata (:: key: value)
	Procedures []Procedure       // List of procedures defined in the script
	Pos        *Position         // Position of the start of the program (e.g., first token)
}

// GetPos returns the program's starting position.
func (p *Program) GetPos() *Position {
	return p.Pos
}

// --- Expression Interface ---

// Expression is the interface implemented by all AST nodes that can be evaluated
// to produce a value.
type Expression interface {
	GetPos() *Position // All expression nodes must report their position
	// String() string    // Optional: For debugging via String() method
}

// --- Expression Node Types ---
// Using concrete types for expression nodes
// *** Includes Pos *Position and GetPos() method on all expression nodes ***

// Represents the target of a function or tool call
type CallTarget struct {
	Pos    *Position // Position of the target identifier
	IsTool bool      // True if the target starts with 'tool.'
	Name   string    // The name of the procedure or tool (e.g., "myFunc", "ReadFile")
}

// GetPos returns the target's position.
func (ct *CallTarget) GetPos() *Position {
	return ct.Pos
}

// Node representing a function/tool call within an expression
type CallableExprNode struct {
	Pos       *Position    // Position of the start of the call (e.g., target name)
	Target    CallTarget   // The function or tool being called
	Arguments []Expression // Evaluated arguments (now require Expression interface)
}

// GetPos returns the call's starting position.
func (n *CallableExprNode) GetPos() *Position {
	return n.Pos
}

type VariableNode struct {
	Pos  *Position // Position of the variable name
	Name string
}

// GetPos returns the variable's position.
func (n *VariableNode) GetPos() *Position {
	return n.Pos
}

type PlaceholderNode struct {
	Pos  *Position // Position of the placeholder
	Name string
}

// GetPos returns the placeholder's position.
func (n *PlaceholderNode) GetPos() *Position {
	return n.Pos
}

type LastNode struct {
	Pos *Position // Position of the 'last' keyword
}

// GetPos returns the 'last' keyword's position.
func (n *LastNode) GetPos() *Position {
	return n.Pos
}

type EvalNode struct {
	Pos      *Position  // Position of the 'eval' keyword
	Argument Expression // Argument must also be an Expression
}

// GetPos returns the 'eval' keyword's position.
func (n *EvalNode) GetPos() *Position {
	return n.Pos
}

type StringLiteralNode struct {
	Pos   *Position // Position of the start of the literal
	Value string
	IsRaw bool // True for ```...``` strings
}

// GetPos returns the string literal's position.
func (n *StringLiteralNode) GetPos() *Position {
	return n.Pos
}

type NumberLiteralNode struct {
	Pos   *Position   // Position of the start of the literal
	Value interface{} // Holds int64 or float64
}

// GetPos returns the number literal's position.
func (n *NumberLiteralNode) GetPos() *Position {
	return n.Pos
}

type BooleanLiteralNode struct {
	Pos   *Position // Position of the literal ('true' or 'false')
	Value bool
}

// GetPos returns the boolean literal's position.
func (n *BooleanLiteralNode) GetPos() *Position {
	return n.Pos
}

type ListLiteralNode struct {
	Pos      *Position    // Position of the opening bracket '['
	Elements []Expression // Elements must be Expressions
}

// GetPos returns the list literal's starting position.
func (n *ListLiteralNode) GetPos() *Position {
	return n.Pos
}

type MapEntryNode struct {
	Pos   *Position         // Position of the key string literal
	Key   StringLiteralNode // Keys must be string literals currently
	Value Expression        // Value must be an Expression
}

// GetPos returns the map entry key's position.
func (n *MapEntryNode) GetPos() *Position {
	// Return position of the Key, as it's the start of the entry
	return n.Key.GetPos()
}

type MapLiteralNode struct {
	Pos     *Position      // Position of the opening brace '{'
	Entries []MapEntryNode // Entries hold key/value pairs
}

// GetPos returns the map literal's starting position.
func (n *MapLiteralNode) GetPos() *Position {
	return n.Pos
}

type ElementAccessNode struct {
	Pos        *Position  // Position of the opening bracket '[' used for access
	Collection Expression // Expression node yielding collection
	Accessor   Expression // Expression node yielding index/key
}

// GetPos returns the element access's starting position.
func (n *ElementAccessNode) GetPos() *Position {
	return n.Pos
}

type UnaryOpNode struct {
	Pos      *Position  // Position of the operator token
	Operator string     // e.g., "-", "not", "no", "some"
	Operand  Expression // Operand must be an Expression
}

// GetPos returns the unary operator's position.
func (n *UnaryOpNode) GetPos() *Position {
	return n.Pos
}

type BinaryOpNode struct {
	Pos      *Position  // Position of the operator token
	Left     Expression // Left operand must be an Expression
	Operator string     // e.g., "+", "==", "and"
	Right    Expression // Right operand must be an Expression
}

// GetPos returns the binary operator's position.
func (n *BinaryOpNode) GetPos() *Position {
	return n.Pos
}

// --- Procedure & Step Structures ---

type Procedure struct {
	Pos               *Position         // Position of the 'func' or procedure keyword
	Name              string            // Position is part of the identifier token
	RequiredParams    []string          // Names only for now, positions could be added later if needed
	OptionalParams    []string          // Names only
	ReturnVarNames    []string          // Names only
	Steps             []Step            // The sequence of operations
	OriginalSignature string            // Original text for reference/debugging
	Metadata          map[string]string // Metadata defined within the procedure (:: key: value)
}

// GetPos returns the procedure's starting position.
func (p *Procedure) GetPos() *Position {
	return p.Pos
}

// Step represents a single operation within a procedure.
// Fields are used based on the 'Type'.
type Step struct {
	Pos       *Position         // Position of the start of the step (e.g., 'set', 'if', 'return' keyword)
	Type      string            // "set", "return", "emit", "if", "while", "for", "must", "mustbe", "fail", "on_error", "clear_error", "ask"
	Target    string            // Variable name (set, for), Check function (mustbe)
	Cond      Expression        // Condition expression (if, while), Collection expression (for) - Requires Expression interface
	Value     interface{}       // RHS expression (set->Expression), Return value(s) ([]Expression) (return), Emit value (Expression) (emit), Must condition (Expression) (must), Fail expression (Expression) (fail), Body steps ([]Step) (if-then, while, for, on_error), Prompt (Expression) (ask) - Type varies, use Expression where applicable
	ElseValue interface{}       // Else body steps ([]Step) (if)
	Metadata  map[string]string // Metadata associated with this step (future use)
}

// GetPos returns the step's starting position.
func (s *Step) GetPos() *Position {
	return s.Pos
}

// newStep helper function - Position should be set by the caller (AST Builder).
// *** MODIFIED: Added type assertion for 'cond' parameter. ***
func newStep(typ, target string, cond, value, elseValue interface{}) Step {
	// Type assert 'cond' if it's provided and expected to be an Expression
	// (e.g., for 'if', 'while', 'must' steps).
	var condExpr Expression
	if cond != nil {
		var ok bool
		condExpr, ok = cond.(Expression)
		if !ok {
			// Log a warning if the assertion fails, as this helper is likely for tests.
			// The real builder should add a proper error with position info.
			log.Printf("Warning: newStep helper received 'cond' of type %T, expected Expression", cond)
			condExpr = nil // Ensure Cond is nil if assertion fails
		}
	}

	// NOTE: Similar assertions might be needed for 'value' depending on 'typ',
	// but we are only fixing the reported error for 'Cond' for now.
	// The primary AST construction logic in the builder needs careful handling.

	return Step{
		// Pos: Set by caller (AST Builder) using tokenToPosition(ctx.GetStart())
		Type:      typ,
		Target:    target,
		Cond:      condExpr,                // Assign the result of the type assertion
		Value:     value,                   // Keep as interface{} for now
		ElseValue: elseValue,               // Keep as interface{} for now
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
		targetStr = "tool." + targetStr // Adjust prefix if needed
	}
	// Avoid printing complex arguments in default String()
	argSummary := fmt.Sprintf("%d args", len(c.Arguments))
	return fmt.Sprintf("%s(%s)", targetStr, argSummary)
}

// ... Add other String() methods for expression nodes as needed ...
