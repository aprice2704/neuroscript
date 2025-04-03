// pkg/core/ast.go
package core

// --- Expression Node Types ---

type VariableNode struct{ Name string }
type PlaceholderNode struct{ Name string }   // Represents {{...}} syntax - value is raw name
type LastNode struct{}                       // Represents LAST keyword
type EvalNode struct{ Argument interface{} } // NEW: Represents EVAL(expression)

type StringLiteralNode struct{ Value string } // Value is RAW string content
type NumberLiteralNode struct{ Value interface{} }
type BooleanLiteralNode struct{ Value bool }
type ListLiteralNode struct{ Elements []interface{} }
type MapEntryNode struct {
	Key   StringLiteralNode
	Value interface{}
}
type MapLiteralNode struct{ Entries []MapEntryNode }
type ConcatenationNode struct{ Operands []interface{} }
type ComparisonNode struct {
	Left     interface{}
	Operator string
	Right    interface{}
}
type ElementAccessNode struct {
	Collection interface{}
	Accessor   interface{}
}

// --- REMOVED RawTextNode and RawString ---

// --- Procedure & Step Structures (Unchanged) ---

type Docstring struct {
	Purpose, Output, Algorithm, Caveats, Examples string
	InputLines                                    []string
	Inputs                                        map[string]string
}
type Procedure struct {
	Name      string
	Params    []string
	Docstring Docstring
	Steps     []Step
}
type Step struct {
	Type, Target           string
	Value, ElseValue, Cond interface{}
	Args                   []interface{}
}

func newStep(typ, target string, cond, value, elseValue interface{}, args []interface{}) Step {
	return Step{Type: typ, Target: target, Cond: cond, Value: value, ElseValue: elseValue, Args: args}
}
