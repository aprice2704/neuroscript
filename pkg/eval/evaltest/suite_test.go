// NeuroScript Version: 0.8.0
// File version: 2
// Purpose: Corrects type assertion for calling mock procedures in the suite's self-test.
// filename: pkg/eval/evaltest/suite_test.go
// nlines: 95
// risk_rating: LOW

package evaltest

import (
	"fmt"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// mockConformingRuntime is a minimal, "golden path" implementation of the
// eval.Runtime interface, designed specifically to pass the conformance suite.
type mockConformingRuntime struct {
	vars  map[string]lang.Value
	tools map[types.FullName]eval.ToolSpec
	procs map[string]lang.Callable
}

func newMockConformingRuntime() *mockConformingRuntime {
	return &mockConformingRuntime{
		vars:  make(map[string]lang.Value),
		tools: make(map[types.FullName]eval.ToolSpec),
		procs: make(map[string]lang.Callable),
	}
}

// Implement the eval.Runtime interface
func (m *mockConformingRuntime) GetVariable(name string) (lang.Value, bool) {
	v, ok := m.vars[name]
	return v, ok
}

func (m *mockConformingRuntime) GetToolSpec(toolName types.FullName) (eval.ToolSpec, bool) {
	s, ok := m.tools[toolName]
	return s, ok
}

func (m *mockConformingRuntime) ExecuteTool(toolName types.FullName, args map[string]lang.Value) (lang.Value, error) {
	switch toolName {
	case "tool.test.add":
		a, _ := lang.ToFloat64(args["a"])
		b, _ := lang.ToFloat64(args["b"])
		return lang.NumberValue{Value: a + b}, nil
	case "tool.test.get_raw_map":
		// The runtime's job is to wrap.
		return lang.MapValue{Value: map[string]lang.Value{"raw": lang.BoolValue{Value: true}}}, nil
	case "tool.test.get_panic":
		// The runtime's job is to recover and return an error.
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, "panic recovered", nil)
	default:
		return nil, lang.ErrToolNotFound
	}
}

func (m *mockConformingRuntime) RunProcedure(procName string, args ...lang.Value) (lang.Value, error) {
	proc, ok := m.procs[procName]
	if !ok {
		return nil, lang.ErrProcedureNotFound
	}

	// FIX: Type assert to the concrete mock type to access Arity() and Call()
	mockProc, ok := proc.(*MockHostProc)
	if !ok {
		// This would be a failure in the test setup itself.
		return nil, lang.NewRuntimeError(lang.ErrorCodeInternal, fmt.Sprintf("internal test error: expected *MockHostProc, got %T", proc), nil)
	}

	if mockProc.Arity() != len(args) {
		return nil, lang.ErrArgumentMismatch
	}
	// This is a simplified call for the mock
	return mockProc.Call(m, args)
}

// TestConformanceSuite runs the exported conformance suite against our own
// minimal, correct "golden path" implementation. If this test passes,
// the suite itself is considered valid and usable by implementors.
func TestConformanceSuite(t *testing.T) {
	// 1. Define the factory function as required by the suite
	factory := func(t *testing.T) eval.Runtime {
		rt := newMockConformingRuntime()

		// 2. Pre-load the runtime with the exact state the suite expects
		rt.vars["foo"] = lang.StringValue{Value: "bar"}
		rt.vars["my_map"] = lang.MapValue{Value: map[string]lang.Value{"a": lang.NumberValue{1}}}

		rt.tools["tool.test.add"] = (&MockHostTool{}).GetSpec()
		rt.tools["tool.test.get_raw_map"] = (&MockHostToolRawMap{}).GetSpec()
		rt.tools["tool.test.get_panic"] = (&MockHostToolPanic{}).GetSpec()

		rt.procs["my_proc"] = &MockHostProc{}

		return rt
	}

	// 3. Run the conformance suite. This test will only pass if the
	// mockConformingRuntime correctly implements the runtime contract
	// and the test suite itself is logically sound.
	RunConformanceTests(t, factory)
}
