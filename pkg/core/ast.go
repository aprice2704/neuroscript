package core

// Docstring sections (PURPOSE, INPUTS, etc.)
type Docstring struct {
	Purpose   string
	Inputs    map[string]string // "argName": "description"
	Output    string
	Algorithm string
	Caveats   string
	Examples  string
}

// Procedure definition
type Procedure struct {
	Name      string
	Params    []string // Input parameter names
	Docstring Docstring
	Steps     []Step
}

// Individual step (SET, CALL, etc.)
type Step struct {
	Type   string      // "SET", "CALL", "IF", "RETURN"
	Target string      // Variable or procedure name
	Value  interface{} // For SET: value to assign
	Args   []string    // For CALL: arguments
	Cond   string      // For IF: condition
}
