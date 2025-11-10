// filename: pkg/parser/ast_nodekind_test.go
// NeuroScript Version: 0.8.0
// File version: 12
// Purpose: STRENGTHENED assertions to catch both KindUnknown(0) and invalid out-of-range Kinds by using types.KindMarker.
// nlines: 115
// risk_rating: MEDIUM

package parser

import (
	"io/ioutil"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
	"github.com/aprice2704/neuroscript/pkg/types"
)

func TestAllNodesHaveValidKind(t *testing.T) {
	testCases := []struct {
		name     string
		filePath string
	}{
		{"Comprehensive Syntax", filepath.Join("testdata", "valid_comprehensive_syntax.ns.txt")},
		{"Comprehensive Grammar", filepath.Join("..", "antlr", "comprehensive_grammar.ns")},
		{"Command Block", filepath.Join("..", "antlr", "command_block.ns")},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			scriptBytes, err := ioutil.ReadFile(tc.filePath)
			if err != nil {
				t.Fatalf("Failed to read test file '%s': %v", tc.filePath, err)
			}
			script := string(scriptBytes)

			prog := testParseAndBuild(t, script)
			if prog == nil {
				t.Fatal("Parsing and building resulted in a nil program.")
			}

			// visited map prevents infinite loops on cyclic pointers, though our AST should be a DAG.
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

				// If it's not a struct, we can't inspect its fields.
				if v.Kind() != reflect.Struct {
					return
				}

				// Check if the struct itself is (or can be addressed as) an ast.Node.
				if v.CanAddr() {
					if node, ok := v.Addr().Interface().(ast.Node); ok {
						// This is our primary check.
						baseNodeField := v.FieldByName("BaseNode")
						if baseNodeField.IsValid() {
							if bn, ok := baseNodeField.Addr().Interface().(*ast.BaseNode); ok {
								t.Logf("Visiting Node: Type=%-25T, Kind=%-20s(%d), Pos=%s", node, bn.NodeKind, bn.NodeKind, bn.StartPos)

								// --- STRENGTHENED ASSERTIONS ---
								// Use the sentinels from pkg/types/kind.go
								if bn.NodeKind <= types.KindUnknown {
									t.Errorf("Found AST node with uninitialized or Unknown NodeKind (<= 0).\n- Node Type: %T\n- Kind: %d\n- Start Pos: %s", node, bn.NodeKind, bn.StartPos)
								}
								if bn.NodeKind >= types.KindMarker {
									t.Errorf("Found AST node with invalid out-of-range NodeKind. This is the Kind(57) bug!\n- Node Type: %T\n- Invalid Kind: %d (>= KindMarker %d)\n- Start Pos: %s", node, bn.NodeKind, types.KindMarker, bn.StartPos)
								}
								// --- END ASSERTIONS ---
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
							walk(iter.Value()) // We only care about values, not keys.
						}
					}
				}
			}

			// Start the walk from the root of the program.
			walk(reflect.ValueOf(prog))
		})
	}
}
