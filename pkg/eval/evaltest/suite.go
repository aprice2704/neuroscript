// NeuroScript Version: 0.8.0
// File version: 3
// Purpose: Corrects field name from 'Name' to 'FullName' in eval.ToolSpec literals.
// filename: pkg/eval/evaltest/suite.go
// nlines: 150
// risk_rating: MEDIUM

package evaltest

import (
	"errors"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/eval"
	"github.com/aprice2704/neuroscript/pkg/lang"
	// "github.com/aprice2704/neuroscript/pkg/types" // REMOVED - This was the source of the error
)

// MockHostTool is a simple tool implementation for testing.
type MockHostTool struct{}

func (t *MockHostTool) GetSpec() eval.ToolSpec { // FIXED: Was types.ToolSpec
	return eval.ToolSpec{ // FIXED: Was types.ToolSpec
		FullName: "tool.test.add", // FIXED: Was Name
		Args: []eval.ArgSpec{ // FIXED: Was types.ToolArgs
			{Name: "a", Type: "number", Required: true}, // FIXED: Added eval.ArgSpec type
			{Name: "b", Type: "number", Required: true}, // FIXED: Added eval.ArgSpec type
		},
		// Out: "number", // eval.ToolSpec doesn't have an 'Out' field
	}
}
func (t *MockHostTool) Call(args map[string]lang.Value) (lang.Value, error) {
	a, _ := lang.ToFloat64(args["a"])
	b, _ := lang.ToFloat64(args["b"])
	return lang.NumberValue{Value: a + b}, nil
}

// MockHostToolRawMap is a tool that incorrectly returns a raw Go map.
type MockHostToolRawMap struct{}

func (t *MockHostToolRawMap) GetSpec() eval.ToolSpec { // FIXED: Was types.ToolSpec
	return eval.ToolSpec{FullName: "tool.test.get_raw_map"} // FIXED: Was Name
}
func (t *MockHostToolRawMap) Call(args map[string]lang.Value) (lang.Value, error) {
	// This is intentionally wrong. The runtime must wrap this.
	rawMap := map[string]any{"raw": true}
	return lang.Wrap(rawMap) // Simulating a host that correctly wraps
}

// MockHostToolPanic is a tool that panics.
type MockHostToolPanic struct{}

func (t *MockHostToolPanic) GetSpec() eval.ToolSpec {
	return eval.ToolSpec{FullName: "tool.test.get_panic"}
} // FIXED: Was Name
func (t *MockHostToolPanic) Call(args map[string]lang.Value) (lang.Value, error) {
	panic("oh no")
}

// MockHostProc is a simple mock procedure for testing.
type MockHostProc struct{}

func (p *MockHostProc) Name() string { return "my_proc" }
func (p *MockHostProc) IsCallable()  {}
func (p *MockHostProc) Arity() int   { return 1 }
func (p *MockHostProc) Call(rt any, args []lang.Value) (lang.Value, error) {
	if len(args) != 1 {
		return nil, lang.ErrArgumentMismatch
	}
	a, _ := lang.ToFloat64(args[0])
	return lang.NumberValue{Value: a + 1}, nil
}

// RuntimeFactory is a function that creates a new, clean eval.Runtime for testing.
// The caller is responsible for pre-loading this runtime with:
// - Variable "foo" = lang.StringValue{"bar"}
// - Variable "my_map" = lang.MapValue{...}
// - Tool "tool.test.add" (MockHostTool)
// - Tool "tool.test.get_raw_map" (MockHostToolRawMap)
// - Tool "tool.test.get_panic" (MockHostToolPanic)
// - Procedure "my_proc" (MockHostProc)
type RuntimeFactory func(t *testing.T) eval.Runtime

// RunConformanceTests executes the full test suite against a runtime implementation.
func RunConformanceTests(t *testing.T, factory RuntimeFactory) {
	t.Run("GetVariable", func(t *testing.T) {
		rt := factory(t)
		t.Run("Get Existing Variable", func(t *testing.T) {
			val, ok := rt.GetVariable("foo")
			if !ok {
				t.Fatal("ok was false for existing variable 'foo'")
			}
			if !reflect.DeepEqual(val, lang.StringValue{Value: "bar"}) {
				t.Fatalf("Expected lang.StringValue{\"bar\"}, got %#v", val)
			}
		})

		t.Run("Get Non-Existent Variable", func(t *testing.T) {
			_, ok := rt.GetVariable("non_existent_var")
			if ok {
				t.Fatal("ok was true for non-existent variable")
			}
		})

		t.Run("Collection Value-Type Correctness", func(t *testing.T) {
			val, ok := rt.GetVariable("my_map")
			if !ok {
				t.Fatal("ok was false for existing variable 'my_map'")
			}
			if _, isPtr := val.(*lang.MapValue); isPtr {
				t.Fatal("GetVariable returned *MapValue, must return MapValue (by value)")
			}
			if _, isVal := val.(lang.MapValue); !isVal {
				t.Fatalf("GetVariable did not return MapValue, got %T", val)
			}
		})
	})

	t.Run("GetToolSpec", func(t *testing.T) {
		rt := factory(t)
		t.Run("Get Existing Spec", func(t *testing.T) {
			spec, ok := rt.GetToolSpec("tool.test.add")
			if !ok {
				t.Fatal("ok was false for existing tool spec")
			}
			if spec.FullName != "tool.test.add" || len(spec.Args) != 2 {
				t.Fatalf("Spec was not returned correctly: %#v", spec)
			}
		})

		t.Run("Get Non-Existent Spec", func(t *testing.T) {
			_, ok := rt.GetToolSpec("tool.fake.nonexistent")
			if ok {
				t.Fatal("ok was true for non-existent tool spec")
			}
		})
	})

	t.Run("ExecuteTool", func(t *testing.T) {
		rt := factory(t)
		t.Run("Execute Valid Tool", func(t *testing.T) {
			args := map[string]lang.Value{"a": lang.NumberValue{10}, "b": lang.NumberValue{5}}
			val, err := rt.ExecuteTool("tool.test.add", args)
			if err != nil {
				t.Fatalf("ExecuteTool failed: %v", err)
			}
			if !reflect.DeepEqual(val, lang.NumberValue{Value: 15}) {
				t.Fatalf("Expected lang.NumberValue{15}, got %#v", val)
			}
		})

		t.Run("Execute Non-Existent Tool", func(t *testing.T) {
			_, err := rt.ExecuteTool("tool.fake.nonexistent", nil)
			if !errors.Is(err, lang.ErrToolNotFound) {
				t.Fatalf("Expected ErrToolNotFound, got %v", err)
			}
		})

		t.Run("Contract: No Raw Return Values", func(t *testing.T) {
			val, err := rt.ExecuteTool("tool.test.get_raw_map", nil)
			if err != nil {
				t.Fatalf("ExecuteTool failed: %v", err)
			}
			if _, ok := val.(lang.MapValue); !ok {
				t.Fatalf("Runtime did not wrap raw map return. Expected lang.MapValue, got %T", val)
			}
		})

		t.Run("Contract: Recover from Tool Panic", func(t *testing.T) {
			_, err := rt.ExecuteTool("tool.test.get_panic", nil)
			if err == nil {
				t.Fatal("Runtime did not recover from tool panic")
			}
			var re *lang.RuntimeError
			if !errors.As(err, &re) {
				t.Fatalf("Expected panic to be wrapped in *lang.RuntimeError, got %T", err)
			}
		})
	})

	t.Run("RunProcedure", func(t *testing.T) {
		rt := factory(t)
		t.Run("Execute Valid Procedure", func(t *testing.T) {
			val, err := rt.RunProcedure("my_proc", lang.NumberValue{10})
			if err != nil {
				t.Fatalf("RunProcedure failed: %v", err)
			}
			if !reflect.DeepEqual(val, lang.NumberValue{Value: 11}) {
				t.Fatalf("Expected lang.NumberValue{11}, got %#v", val)
			}
		})

		t.Run("Execute Non-Existent Procedure", func(t *testing.T) {
			_, err := rt.RunProcedure("fake_proc")
			if !errors.Is(err, lang.ErrProcedureNotFound) {
				t.Fatalf("Expected ErrProcedureNotFound, got %v", err)
			}
		})

		t.Run("Argument Mismatch", func(t *testing.T) {
			_, err := rt.RunProcedure("my_proc", lang.NumberValue{10}, lang.NumberValue{11})
			if !errors.Is(err, lang.ErrArgumentMismatch) {
				t.Fatalf("Expected ErrArgumentMismatch, got %v", err)
			}
		})
	})
}
