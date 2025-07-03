// filename: pkg/interpreter/testing_bits.go
package interpreter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

// RunSteps is an exported wrapper for the unexported executeSteps method, allowing it to be called by external test packages.
func (i *Interpreter) RunSteps(steps []ast.Step) (lang.Value, bool, bool, error) {
	return i.executeSteps(steps, false, nil)
}

// GetLastResult is an exported wrapper that allows external tests to retrieve the unexported lastCallResult field.
func (i *Interpreter) GetLastResult() lang.Value {
	return i.lastCallResult
}

// DebugDumpVariables is a testing helper to print the current state of variables
// in an interpreter instance. It's kept within the interpreter package to avoid
// import cycles with the testutil package.
func DebugDumpVariables(i *Interpreter, t *testing.T) {
	t.Helper()
	var sb strings.Builder
	sb.WriteString("\n--- Variable Dump ---\n")
	vars, err := i.GetAllVariables()
	if err != nil {
		sb.WriteString(fmt.Sprintf("Error getting variables: %v\n", err))
		t.Log(sb.String())
		return
	}

	if len(vars) == 0 {
		sb.WriteString("No variables set.\n")
	} else {
		for key, val := range vars {
			sb.WriteString(fmt.Sprintf("%-20s (%T):\t%#v\n", key, val, val))
		}
	}
	sb.WriteString("---------------------\n")
	t.Log(sb.String())
}
