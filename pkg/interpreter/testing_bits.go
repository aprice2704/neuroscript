// NeuroScript Version: 0.7.0
// File version: 6
// Purpose: Added the 'model:read' capability to the privileged test policy to allow tools like 'agentmodel.get' to run in tests.
// filename: pkg/interpreter/testing_bits.go
// nlines: 95
// risk_rating: LOW
package interpreter

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/logging"
	"github.com/aprice2704/neuroscript/pkg/policy"
	"github.com/aprice2704/neuroscript/pkg/policy/capability"
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

// NewTestInterpreter is an exported test helper for creating a pre-configured
// interpreter instance, accessible from other packages.
func NewTestInterpreter(t *testing.T, initialVars map[string]lang.Value, lastResult lang.Value, privileged bool) (*Interpreter, error) {
	t.Helper()
	testLogger := logging.NewTestLogger(t)
	testLogger.SetLevel(interfaces.LogLevelInfo)
	sandboxDir := t.TempDir()

	opts := []InterpreterOption{
		WithLogger(testLogger),
		WithSandboxDir(sandboxDir),
	}

	if privileged {
		policy := &policy.ExecPolicy{
			Context: policy.ContextConfig, // Allows trusted tools
			Allow:   []string{"*"},
			Grants: capability.NewGrantSet(
				[]capability.Capability{
					{Resource: "model", Verbs: []string{"admin", "use", "read"}, Scopes: []string{"*"}}, // FIX: Added 'read' verb
					{Resource: "account", Verbs: []string{"admin"}, Scopes: []string{"*"}},
					{Resource: "env", Verbs: []string{"read"}, Scopes: []string{"*"}},
					{Resource: "net", Verbs: []string{"read"}, Scopes: []string{"*"}},
				},
				capability.Limits{},
			),
		}
		opts = append(opts, WithExecPolicy(policy))
	}

	interp := NewInterpreter(opts...)

	for k, v := range initialVars {
		if err := interp.SetInitialVariable(k, v); err != nil {
			return nil, fmt.Errorf("failed to set initial variable %q: %w", k, err)
		}
	}
	if lastResult != nil {
		interp.lastCallResult = lastResult
	}
	return interp, nil
}
