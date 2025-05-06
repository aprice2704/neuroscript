// NeuroScript Version: 0.3.1
// File version: 0.0.11 // Correctly ignore only line/column in comparison. Use raw string fixtures.
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

	interpreter, err := core.NewInterpreter(logger, llmClient, sandboxDir, nil)
	if err != nil {
		t.Fatalf("Failed to create core.Interpreter: %v", err)
	}
	err = core.RegisterCoreTools(interpreter.ToolRegistry())
	if err != nil {
		t.Fatalf("Failed to register core tools: %v", err)
	}
	err = interpreter.SetSandboxDir(sandboxDir)
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

	// Run GoIndexCode
	indexResult, indexErr := toolGoIndexCode(interpreter, []interface{}{"."})
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

	// Log the actual loaded package paths/IDs for debugging verification
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
			if pkgInfo.PkgPath == expectedPkgPath || pkgInfo.ID == expectedPkgPath || (pkgInfo.Types != nil && pkgInfo.Types.Path() == expectedPkgPath) {
				foundCorrectPackage = true
			}
		}
	} else {
		t.Logf("  No packages found in index!")
	}
	t.Logf("------------------------------")

	if !foundCorrectPackage {
		t.Fatalf("Setup Error: Expected package %q was not found in the created index. Check GoIndexCode logs and fixture setup.", expectedPkgPath)
	}

	// --- Define Test Cases ---
	// wantResult maps contain expected path, name, kind (line/column are removed before comparison)
	testCases := []struct {
		name        string
		query       string
		wantErr     error
		wantResult  map[string]interface{} // nil means expect nil, non-nil compared after removing line/col
		skipCompare bool
	}{
		{name: "Find Top Level Function", query: "package:mytestmodule/pkga; function:TopLevelFunc", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "TopLevelFunc", "kind": "function"}},
		{name: "Find Type Struct", query: "package:mytestmodule/pkga; type:MyStruct", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "MyStruct", "kind": "type"}},
		{name: "Find Type Interface", query: "package:mytestmodule/pkga; interface:MyInterface", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "MyInterface", "kind": "type"}},
		{name: "Find Pointer Receiver Method", query: "package:mytestmodule/pkga; type:MyStruct; method:PointerMethod", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "PointerMethod", "kind": "method"}},

		// --- Receiver Constraint Tests (Keep expected values as-is to highlight the remaining logic bug) ---
		{name: "Find Pointer Receiver Method with Receiver Constraint", query: "package:mytestmodule/pkga; type:MyStruct; method:PointerMethod; receiver:*MyStruct", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "PointerMethod", "kind": "method"}}, // Expected non-nil, but might fail due to logic bug
		{name: "Method with Mismatched Receiver Constraint", query: "package:mytestmodule/pkga; type:MyStruct; method:PointerMethod; receiver:MyStruct", wantResult: nil},                                                                                                   // Expected nil, but might fail due to logic bug
		// --- End Receiver Constraint Tests ---

		{name: "Find Value Receiver Method", query: "package:mytestmodule/pkga; type:MyStruct; method:ValueMethod", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "ValueMethod", "kind": "method"}},
		{name: "Find Value Receiver Method with Receiver Constraint", query: "package:mytestmodule/pkga; type:MyStruct; method:ValueMethod; receiver:MyStruct", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "ValueMethod", "kind": "method"}},
		{name: "Find Struct Field via 'field' alias", query: "package:mytestmodule/pkga; type:MyStruct; field:FieldA", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "FieldA", "kind": "field"}},
		{name: "Find Struct Field via 'var'", query: "package:mytestmodule/pkga; type:MyStruct; var:FieldA", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "FieldA", "kind": "field"}},
		{name: "Find Global Variable", query: "package:mytestmodule/pkga; var:GlobalVar", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "GlobalVar", "kind": "variable"}},
		{name: "Find Global Constant", query: "package:mytestmodule/pkga; const:GlobalConst", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "GlobalConst", "kind": "constant"}},
		{name: "Find Unexported Func", query: "package:mytestmodule/pkga; function:anotherFunc", wantResult: map[string]interface{}{"path": "pkga/pkga.go", "name": "anotherFunc", "kind": "function"}},

		// --- Negative Test Cases (wantResult = nil will be compared) ---
		{name: "Symbol Not Found - NonExistentFunc", query: "package:mytestmodule/pkga; function:NonExistentFunc", wantResult: nil},
		{name: "Symbol Not Found - Wrong Kind (var as func)", query: "package:mytestmodule/pkga; function:GlobalVar", wantResult: nil},
		{name: "Method Not Found On Type", query: "package:mytestmodule/pkga; type:MyStruct; method:DoesNotExist", wantResult: nil},
		{name: "Field Not Found On Type", query: "package:mytestmodule/pkga; type:MyStruct; field:DoesNotExist", wantResult: nil},
		{name: "Package Not Found In Index", query: "package:nonexistent/pkg; function:SomeFunc", wantResult: nil},

		// --- Invalid Query Format Tests (wantErr != nil) ---
		{name: "Invalid Query - Missing Package", query: "type:MyStruct; function:TopLevelFunc", wantErr: ErrInvalidQueryFormat, skipCompare: true},
		{name: "Invalid Query - Multiple Symbol Keys (func and type)", query: "package:mytestmodule/pkga; function:TopLevelFunc; type:MyStruct", wantErr: ErrInvalidQueryFormat, skipCompare: true},
		{name: "Invalid Query - Malformed Pair", query: "package:mytestmodule/pkga; functionTopLevelFunc", wantErr: ErrInvalidQueryFormat, skipCompare: true},
		{name: "Invalid Query - Unknown Key", query: "package:mytestmodule/pkga; function:TopLevelFunc; badkey:abc", wantErr: ErrInvalidQueryFormat, skipCompare: true},
		{name: "Invalid Query - Method without Type/Interface", query: "package:mytestmodule/pkga; method:PointerMethod", wantErr: ErrInvalidQueryFormat, skipCompare: true},
	}

	// --- Run Tests ---
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Mark tests as parallelizable

			result, runErr := toolGoGetDeclarationOfSymbol(interpreter, []interface{}{indexHandle, tc.query})

			// --- Error Checking ---
			if tc.wantErr != nil {
				if runErr == nil {
					t.Errorf("Expected error wrapping %q, but got nil", tc.wantErr)
				} else {
					isCorrectError := errors.Is(runErr, tc.wantErr)
					if !isCorrectError && errors.Is(tc.wantErr, ErrInvalidQueryFormat) {
						isCorrectError = errors.Is(runErr, core.ErrInvalidArgument) && strings.Contains(runErr.Error(), ErrInvalidQueryFormat.Error())
					}
					if !isCorrectError {
						t.Errorf("Expected error wrapping %q (or ErrInvalidArgument wrapping it), but got %q (%v)", tc.wantErr, runErr, runErr)
					}
				}
				if result != nil {
					t.Errorf("Expected nil result on error, but got: %v", result)
				}
				return // End test case for expected error
			}

			// If no error was expected, fail if one occurred
			if runErr != nil {
				t.Fatalf("Did not expect error for query %q, but got: %v", tc.query, runErr)
			}

			// --- Result Comparison (Ignoring line/column for non-nil results) ---
			if !tc.skipCompare { // skipCompare is true only for error tests handled above
				var resultMapForCompare map[string]interface{} // Map used for comparison (line/col removed)
				var originalResultMap map[string]interface{}   // Keep original for logging if needed

				// Prepare resultMapForCompare based on the actual result
				if result != nil {
					tempMap, ok := result.(map[string]interface{})
					if ok {
						originalResultMap = tempMap // Store the original
						resultMapForCompare = make(map[string]interface{})
						// Copy keys except line and column
						for k, v := range tempMap {
							if k != "line" && k != "column" {
								resultMapForCompare[k] = v
							}
						}
					} else {
						// If result is not nil but not a map, it's an unexpected type
						t.Fatalf("Expected result type map[string]interface{} or nil, but got %T: %v", result, result)
					}
					// If result was nil, resultMapForCompare remains nil, which is correct
				}

				// Compare the modified actual result (resultMapForCompare) with the expected result (tc.wantResult)
				if !reflect.DeepEqual(resultMapForCompare, tc.wantResult) {
					t.Errorf("Result mismatch for query %q (ignoring line/column):\n Expected: %#v\n Got (filtered): %#v", tc.query, tc.wantResult, resultMapForCompare)
					// Log the original unfiltered map from the tool for context
					if originalResultMap != nil {
						t.Logf("Original Got (with line/column): %#v", originalResultMap)
					} else {
						t.Logf("Original Got was: %T %v", result, result) // Log if it wasn't even a map
					}
				}
			}
		})
	}
}
