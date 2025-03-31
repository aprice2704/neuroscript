// pkg/core/ast.go
package core

// Docstring sections (PURPOSE, INPUTS, etc.)
type Docstring struct {
	Purpose    string
	InputLines []string          // NEW: Store raw input lines to preserve order
	Inputs     map[string]string // Keep map for potential future parsing (currently unused by listener)
	Output     string
	Algorithm  string
	Caveats    string
	Examples   string
}

// Procedure definition
type Procedure struct {
	Name      string
	Params    []string // Input parameter names
	Docstring Docstring
	Steps     []Step
}

// Individual step (SET, CALL, etc.)
// For block statements (IF, WHILE, FOR), Value will hold []Step
type Step struct {
	Type   string      // "SET", "CALL", "IF", "WHILE", "FOR", "RETURN", etc.
	Target string      // Variable name (SET, FOR) or Procedure/LLM/Tool name (CALL)
	Value  interface{} // Value (SET, RETURN), Arguments (CALL), Block Body ([]Step for IF/WHILE/FOR)
	Args   []string    // Arguments (CALL) - Potentially redundant if moved to Value? Keep for now.
	Cond   string      // Condition (IF, WHILE), Collection (FOR)
}

// newStep creates a new Step instance. Moved from neuroscript.y.go.
func newStep(typ string, target string, cond string, value interface{}, args []string) Step {
	return Step{Type: typ, Target: target, Cond: cond, Value: value, Args: args}
}
