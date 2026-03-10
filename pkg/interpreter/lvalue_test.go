// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 6
// :: description: Updated multi-assignment test to check returned values rather than internal sandbox variables.
// :: latestChange: Multi-assignment test now returns x, y for assertion.
// :: filename: pkg/interpreter/lvalue_test.go
// :: serialization: go

package interpreter_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestLValueAssignments(t *testing.T) {
	t.Run("Vivification of nested maps", func(t *testing.T) {
		h := NewTestHarness(t)
		script := `
			func main() means
				set a.b.c = "deeply nested"
				return a
			endfunc
		`
		result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		// Expected structure: map[b:map[c:deeply nested]]
		aMap, ok := result.(*lang.MapValue)
		if !ok {
			t.Fatalf("Expected a map result, got %T", result)
		}
		b, _ := aMap.Value["b"].(*lang.MapValue)
		if b == nil {
			t.Fatalf("Expected key 'b' to be a map, got %T", aMap.Value["b"])
		}
		c, _ := b.Value["c"].(lang.StringValue)

		if c.Value != "deeply nested" {
			t.Errorf("Vivification failed. Expected 'deeply nested', got '%s'", c.Value)
		}
	})

	t.Run("Vivification of nested lists and maps", func(t *testing.T) {
		h := NewTestHarness(t)
		script := `
			func main() means
				set data[0].name = "Alice"
				set data[1].name = "Bob"
				return data
			endfunc
		`
		result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		resultList, ok := result.(*lang.ListValue)
		if !ok {
			t.Fatalf("Expected a list result, got %T", result)
		}
		if len(resultList.Value) != 2 {
			t.Fatalf("Expected list of length 2, got %d", len(resultList.Value))
		}

		person0, _ := resultList.Value[0].(*lang.MapValue)
		if person0 == nil {
			t.Fatalf("Expected item 0 to be a map, got %T", resultList.Value[0])
		}
		name0, _ := person0.Value["name"].(lang.StringValue)
		if name0.Value != "Alice" {
			t.Errorf("Expected name 'Alice', got '%s'", name0.Value)
		}
	})

	t.Run("Vivification with various types", func(t *testing.T) {
		h := NewTestHarness(t)
		script := `
            func main() means
                set config.port = 8080
                set config.enabled = true
                set config.features[0] = "login"
                return config
            endfunc
        `
		result, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		resultMap, ok := result.(*lang.MapValue)
		if !ok {
			t.Fatalf("Expected a map result, got %T", result)
		}

		port, _ := resultMap.Value["port"].(lang.NumberValue)
		if port.Value != 8080 {
			t.Errorf("Expected port 8080, got %v", port.Value)
		}

		enabled, _ := resultMap.Value["enabled"].(lang.BoolValue)
		if !enabled.Value {
			t.Error("Expected enabled to be true")
		}

		features, _ := resultMap.Value["features"].(*lang.ListValue)
		if features == nil {
			t.Fatalf("Expected 'features' to be a list, got %T", resultMap.Value["features"])
		}
		feature0, _ := features.Value[0].(lang.StringValue)
		if feature0.Value != "login" {
			t.Errorf("Expected feature 'login', got '%s'", feature0.Value)
		}
	})

	t.Run("Error on indexing a non-container", func(t *testing.T) {
		h := NewTestHarness(t)
		script := `
			func main() means
				set my_string = "hello"
				set my_string[0] = "H"
			endfunc
		`
		_, err := h.Interpreter.ExecuteScriptString("main", script, nil)
		if err == nil {
			t.Fatal("Expected an error when indexing a string, but got nil")
		}
		// This is an internal error because the logic assumes containers.
		if !strings.Contains(err.Error(), "traverseAndSet called on a non-container") {
			t.Errorf("Expected a container error, but got: %v", err)
		}
	})

	t.Run("Error on multi-assignment count mismatch", func(t *testing.T) {
		h := NewTestHarness(t)
		script := `
			func get_two() means
				return 1, 2
			endfunc
			func main() means
				set a, b, c = get_two()
			endfunc
		`
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		h.Interpreter.Load(&interfaces.Tree{Root: program})

		_, err := h.Interpreter.Run("main")
		if err == nil {
			t.Fatal("Expected an error for assignment count mismatch, but got nil")
		}
		if !strings.Contains(err.Error(), "LHS count 3 doesn't match RHS list length 2") {
			t.Errorf("Expected an assignment count mismatch error, but got: %v", err)
		}
	})

	t.Run("Multi-assignment with correct count", func(t *testing.T) {
		h := NewTestHarness(t)
		script := `
			func get_vals() means
				return "a", 10
			endfunc
			func main() means
				set x, y = get_vals()
				return x, y
			endfunc
		`
		tree, _ := h.Parser.Parse(script)
		program, _, _ := h.ASTBuilder.Build(tree)
		h.Interpreter.Load(&interfaces.Tree{Root: program})
		result, err := h.Interpreter.Run("main")
		if err != nil {
			t.Fatalf("Script execution failed: %v", err)
		}

		resList, ok := result.(lang.ListValue)
		if !ok {
			t.Fatalf("Expected ListValue, got %T", result)
		}
		if len(resList.Value) != 2 {
			t.Fatalf("Expected 2 elements, got %d", len(resList.Value))
		}
		x := resList.Value[0]
		y := resList.Value[1]

		expectedX := lang.StringValue{Value: "a"}
		expectedY := lang.NumberValue{Value: 10}

		if !reflect.DeepEqual(x, expectedX) {
			t.Errorf("Variable 'x' mismatch. Got: %#v, Want: %#v", x, expectedX)
		}
		if !reflect.DeepEqual(y, expectedY) {
			t.Errorf("Variable 'y' mismatch. Got: %#v, Want: %#v", y, expectedY)
		}
	})
}
