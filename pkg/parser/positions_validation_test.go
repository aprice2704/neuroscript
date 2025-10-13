// NeuroScript Version: 0.7.2
// File version: 1
// Purpose: Adds a comprehensive test to validate that all nodes in the AST have non-nil start and end positions.
// filename: pkg/parser/ast_builder_positions_validation_test.go
// nlines: 80
// risk_rating: MEDIUM

package parser

import (
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

// TestAllNodesHaveValidPositions walks the entire AST of a comprehensive
// script and asserts that every node has a non-nil StartPos and StopPos.
// This is critical for downstream tools like LSPs and debuggers.
func TestAllNodesHaveValidPositions(t *testing.T) {
	// A script that includes a variety of node types.
	script := `
:: key: value
func main(needs p1) means
	set x = p1 + 1
	if x > 10
		return "big"
	else
		for each item in [1, 2]
			emit item
		endfor
	endif
	return "small"
endfunc
`
	prog := testParseAndBuild(t, script)
	if prog == nil {
		t.Fatal("Parsing and building resulted in a nil program.")
	}

	// visited map prevents infinite loops on cyclic pointers.
	visited := make(map[uintptr]bool)
	var walk func(v reflect.Value)

	walk = func(v reflect.Value) {
		// Dereference pointers and interfaces to get to the concrete value.
		if v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
			if v.IsNil() {
				return
			}
			// Check for cycles.
			if v.Kind() == reflect.Ptr {
				ptr := v.Pointer()
				if visited[ptr] {
					return
				}
				visited[ptr] = true
			}
			v = v.Elem()
		}

		if v.Kind() != reflect.Struct {
			return
		}

		// Check if the struct is an ast.Node and validate its positions.
		if v.CanAddr() {
			if node, ok := v.Addr().Interface().(ast.Node); ok {
				baseNodeField := v.FieldByName("BaseNode")
				if baseNodeField.IsValid() {
					if bn, ok := baseNodeField.Addr().Interface().(*ast.BaseNode); ok {
						if bn.NodeKind != types.KindUnknown { // Skip nodes that aren't fully formed
							if bn.StartPos == nil {
								t.Errorf("Found AST node with nil StartPos.\n- Node Type: %T", node)
							}
							if bn.StopPos == nil {
								t.Errorf("Found AST node with nil StopPos.\n- Node Type: %T", node)
							}
						}
					}
				}
			}
		}

		// Recursively walk through the fields of the struct.
		for i := 0; i < v.NumField(); i++ {
			fieldVal := v.Field(i)
			switch fieldVal.Kind() {
			case reflect.Ptr, reflect.Interface:
				if !fieldVal.IsNil() {
					walk(fieldVal)
				}
			case reflect.Struct:
				walk(fieldVal)
			case reflect.Slice:
				for j := 0; j < fieldVal.Len(); j++ {
					walk(fieldVal.Index(j))
				}
			case reflect.Map:
				iter := fieldVal.MapRange()
				for iter.Next() {
					walk(iter.Value())
				}
			}
		}
	}

	// Start the walk from the root of the program.
	walk(reflect.ValueOf(prog))
}
