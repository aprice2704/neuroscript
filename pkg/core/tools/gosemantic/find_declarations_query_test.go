// NeuroScript Version: 0.3.1
// File version: 0.0.12 // Updated NewInterpreter call signature.
// Test file for GoGetDeclarationOfSymbol tool.
// filename: pkg/core/tools/gosemantic/find_declarations_query_test.go

package gosemantic

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- Test Fixture Source Code ---
// Using standard Go raw string literals for fixtures
const fixturePkgAContent = `package pkga

import "fmt"

const GlobalConst = 123
var GlobalVar = "hello"

type MyStruct struct {
	FieldA int
	fieldB string // unexported
}

func (s *MyStruct) PointerMethod(val string) {
	fmt.Println("Pointer receiver method:", s.FieldA, val)
}

func (s MyStruct) ValueMethod() string {
	return s.fieldB
}

type MyInterface interface {
	DoSomething() error
}

func TopLevelFunc(a int, b string) (string, error) {
	gs := MyStruct{FieldA: a, fieldB: b}
	gs.PointerMethod("from func")
	_ = gs.ValueMethod()
	var localVar = "test"
	fmt.Println(localVar)
	return "ok", nil
}

func anotherFunc() { // unexported
	fmt.Println(GlobalVar)
}
`

const fixtureMainContent = `package main

import (
	"fmt"
	// Use the expected module path after adding go.mod
	thepkga "mytestmodule/pkga"
	//"os" // Commented out to avoid unused import warning in test log
)

func main() {
	fmt.Println(thepkga.GlobalConst)
	s := thepkga.MyStruct{FieldA: 1}
	s.PointerMethod("value")
	fmt.Println(s)
	res, _ := thepkga.TopLevelFunc(thepkga.GlobalConst, thepkga.GlobalVar)
	fmt.Println(res)
}
`

// --- Test Cases ---
func TestGoGetDeclarationOfSymbol(t *testing.T) {
	// --- Test Setup ---
	logger, _ := adapters.NewSimpleSlogAdapter(os.Stderr, logging.LogLevelDebug)
	logger.Debug("Test logger initialized")

	llmClient := adapters.NewNoOpLLMClient()
	sandboxDir := t.TempDir()

	// *** CORRECTED NewInterpreter call with 5 arguments ***
	interpreter, err := core.NewInterpreter(logger, llmClient, sandboxDir, nil, nil) // Pass nil for initialVars and libPaths
	if err != nil {
		t.Fatalf("Failed to create core.Interpreter: %v", err)
	}
	// Note: core.RegisterCoreTools is called within NewInterpreter constructor now.
	// If gosemantic tools need registration, it must happen separately.
	// Assuming for now they use init() or a specific RegisterGosemanticTools func exists.

	err = interpreter.SetSandboxDir(sandboxDir) // Ensure sandbox is set
	if err != nil {
		t.Fatalf("Failed to set sandbox dir: %v", err)
	}

	// --- Create Fixture Files ---
	pkgAPath := filepath.Join(sandboxDir, "pkga")
	if err := os.MkdirAll(pkgAPath, 0755); err != nil {
		t.Fatalf("Failed to create fixture dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgAPath, "pkga.go"), []byte(fixturePkgAContent), 0644); err != nil {
		t.Fatalf("Failed to write fixture pkga.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sandboxDir, "main.go"), []byte(fixtureMainContent), 0644); err != nil {
		t.Fatalf("Failed to write fixture main.go: %v", err)
	}

	goModContent := []byte("module mytestmodule\n\ngo 1.21\n")
	if err := os.WriteFile(filepath.Join(sandboxDir, "go.mod"), goModContent, 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}
	logger.Info("Created go.mod in sandbox", "path", filepath.Join(sandboxDir, "go.mod"))

	// Run GoIndexCode via registry
	indexTool, found := interpreter.ToolRegistry().GetTool("GoIndexCode")
	if !found {
		t.Fatalf("Tool GoIndexCode not found")
	}
	indexResult, indexErr := indexTool.Func(interpreter, []interface{}{"."})
	if indexErr != nil {
		handleCheck, _ := indexResult.(string)
		if handleCheck == "" {
			t.Fatalf("GoIndexCode failed to produce a handle: %v", indexErr)
		} else {
			t.Logf("GoIndexCode reported an error, but returned a handle. Proceeding cautiously: %v", indexErr)
		}
	}
	indexHandle, ok := indexResult.(string)
	if !ok || indexHandle == "" {
		t.Fatalf("GoIndexCode did not return a valid handle string, got %T: %v", indexResult, indexResult)
	}
	t.Logf("Got Semantic Index Handle: %s", indexHandle)

	// Verify index content (optional debug)
	indexValue, getHandleErr := interpreter.GetHandleValue(indexHandle, semanticIndexTypeTag)
	if getHandleErr != nil {
		t.Fatalf("Failed to retrieve index from handle %s: %v", indexHandle, getHandleErr)
	}
	semanticIndex, ok := indexValue.(*SemanticIndex)
	if !ok {
		t.Fatalf("Handle %s did not contain *SemanticIndex", indexHandle)
	}
	t.Logf("--- Packages Found in Index ---")
	foundCorrectPackage := false
	expectedPkgPath := "mytestmodule/pkga"
	if semanticIndex.Packages != nil {
		for _, pkgInfo := range semanticIndex.Packages {
			if pkgInfo == nil {
				continue
			}
			t.Logf("  PkgPath: %q, ID: %q, Name: %q", pkgInfo.PkgPath, pkgInfo.ID, pkgInfo.Name)
			if strings.Contains(pkgInfo.PkgPath, expectedPkgPath) || strings.Contains(pkgInfo.ID, expectedPkgPath) {
				foundCorrectPackage = true
			}
		}
	} else {
		t.Logf("  No packages found in index!")
	}
	t.Logf("------------------------------")
	if !foundCorrectPackage {
		t.Fatalf("Setup Error: Expected package containing %q was not found in the created index.", expectedPkgPath)
	}

	// --- Define Test Cases ---
	testCases := []struct {
		name        string
		query       string
		wantErr     error
		wantResult  map[string]interface{}
		skipCompare bool
	}{
		{name: "Find Top Level Function", query: "package:mytestmodule/pkga; function:TopLevelFunc", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "TopLevelFunc", "kind": "function"}},
		{name: "Find Type Struct", query: "package:mytestmodule/pkga; type:MyStruct", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "MyStruct", "kind": "type"}},
		{name: "Find Type Interface", query: "package:mytestmodule/pkga; interface:MyInterface", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "MyInterface", "kind": "type"}},
		{name: "Find Pointer Receiver Method", query: "package:mytestmodule/pkga; type:MyStruct; method:PointerMethod", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "PointerMethod", "kind": "method"}},
		{name: "Find Pointer Receiver Method with Receiver Constraint", query: "package:mytestmodule/pkga; type:MyStruct; method:PointerMethod; receiver:*MyStruct", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "PointerMethod", "kind": "method"}},
		{name: "Method with Mismatched Receiver Constraint", query: "package:mytestmodule/pkga; type:MyStruct; method:PointerMethod; receiver:MyStruct", wantResult: nil},
		{name: "Find Value Receiver Method", query: "package:mytestmodule/pkga; type:MyStruct; method:ValueMethod", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "ValueMethod", "kind": "method"}},
		{name: "Find Value Receiver Method with Receiver Constraint", query: "package:mytestmodule/pkga; type:MyStruct; method:ValueMethod; receiver:MyStruct", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "ValueMethod", "kind": "method"}},
		{name: "Find Struct Field via 'field' alias", query: "package:mytestmodule/pkga; type:MyStruct; field:FieldA", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "FieldA", "kind": "field"}},
		{name: "Find Struct Field via 'var'", query: "package:mytestmodule/pkga; type:MyStruct; var:FieldA", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "FieldA", "kind": "field"}},
		{name: "Find Global Variable", query: "package:mytestmodule/pkga; var:GlobalVar", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "GlobalVar", "kind": "variable"}},
		{name: "Find Global Constant", query: "package:mytestmodule/pkga; const:GlobalConst", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "GlobalConst", "kind": "constant"}},
		{name: "Find Unexported Func", query: "package:mytestmodule/pkga; function:anotherFunc", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "anotherFunc", "kind": "function"}},
		{name: "Symbol Not Found - NonExistentFunc", query: "package:mytestmodule/pkga; function:NonExistentFunc", wantResult: nil},
		{name: "Symbol Not Found - Wrong Kind (var as func)", query: "package:mytestmodule/pkga; function:GlobalVar", wantResult: nil},
		{name: "Method Not Found On Type", query: "package:mytestmodule/pkga; type:MyStruct; method:DoesNotExist", wantResult: nil},
		{name: "Field Not Found On Type", query: "package:mytestmodule/pkga; type:MyStruct; field:DoesNotExist", wantResult: nil},
		{name: "Package Not Found In Index", query: "package:nonexistent/pkg; function:SomeFunc", wantResult: nil},
		{name: "Invalid Query - Missing Package", query: "type:MyStruct; function:TopLevelFunc", wantErr: ErrInvalidQueryFormat, skipCompare: true},
		{name: "Invalid Query - Multiple Symbol Keys (func and type)", query: "package:mytestmodule/pkga; function:TopLevelFunc; type:MyStruct", wantErr: ErrInvalidQueryFormat, skipCompare: true},
		{name: "Invalid Query - Malformed Pair", query: "package:mytestmodule/pkga; functionTopLevelFunc", wantErr: ErrInvalidQueryFormat, skipCompare: true},
		{name: "Invalid Query - Unknown Key", query: "package:mytestmodule/pkga; function:TopLevelFunc; badkey:abc", wantErr: ErrInvalidQueryFormat, skipCompare: true},
		{name: "Invalid Query - Method without Type/Interface", query: "package:mytestmodule/pkga; method:PointerMethod", wantErr: ErrInvalidQueryFormat, skipCompare: true},
	}

	// --- Run Tests ---
	declTool, foundDecl := interpreter.ToolRegistry().GetTool("GoGetDeclarationOfSymbol")
	if !foundDecl {
		t.Fatalf("Tool GoGetDeclarationOfSymbol not found in registry")
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			result, runErr := declTool.Func(interpreter, []interface{}{indexHandle, tc.query})

			// --- Error Checking ---
			if tc.wantErr != nil {
				if runErr == nil {
					t.Errorf("Expected error wrapping %q, but got nil", tc.wantErr)
				} else {
					isCorrectError := errors.Is(runErr, tc.wantErr)
					// Check for specific case where ErrInvalidQueryFormat might be wrapped by ErrInvalidArgument
					if !isCorrectError && errors.Is(tc.wantErr, ErrInvalidQueryFormat) {
						var rtErr *core.RuntimeError
						if errors.As(runErr, &rtErr) && errors.Is(rtErr.Wrapped, core.ErrInvalidArgument) && strings.Contains(rtErr.Message, ErrInvalidQueryFormat.Error()) {
							isCorrectError = true
						}
					}
					if !isCorrectError {
						t.Errorf("Expected error wrapping %q (or ErrInvalidArgument), but got %q (%v)", tc.wantErr, runErr, runErr)
					}
				}
				if result != nil {
					t.Errorf("Expected nil result on error, but got: %v", result)
				}
				return
			}
			if runErr != nil {
				t.Fatalf("Did not expect error for query %q, but got: %v", tc.query, runErr)
			}

			// --- Result Comparison ---
			if !tc.skipCompare {
				var resultMapForCompare map[string]interface{}
				var originalResultMap map[string]interface{}
				if result != nil {
					tempMap, ok := result.(map[string]interface{})
					if ok {
						originalResultMap = tempMap
						resultMapForCompare = make(map[string]interface{})
						for k, v := range tempMap {
							if k != "line" && k != "column" {
								resultMapForCompare[k] = v
							}
						}
					} else {
						t.Fatalf("Expected result map[string]interface{} or nil, got %T: %v", result, result)
					}
				}
				if !reflect.DeepEqual(resultMapForCompare, tc.wantResult) {
					t.Errorf("Result mismatch for query %q (ignoring line/column):\n Expected: %#v\n Got (filtered): %#v", tc.query, tc.wantResult, resultMapForCompare)
					if originalResultMap != nil {
						t.Logf("Original Got (with line/column): %#v", originalResultMap)
					} else {
						t.Logf("Original Got was: %T %v", result, result)
					}
				}
			}
		})
	}
}
