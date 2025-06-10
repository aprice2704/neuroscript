// NeuroScript Version: 0.3.1
// File version: 0.0.6 // Call Go.FindUsages with 2 args (handle, query); update ToolSpec logging.
// Test file for GoFindUsages tool.
// filename: pkg/core/tools/gosemantic/find_usages_test.go

package gosemantic

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
	"github.com/aprice2704/neuroscript/pkg/toolsets"
)

// --- Fixtures ---
const findUsagesFixturePkgAContent = `package pkga

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

const findUsagesFixtureMainContent = `package main

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

// --- Helper Function ---

// sortResultsFiltered sorts a slice of usage maps based on path then name.
// It returns a NEW slice containing filtered maps (only path, name, kind).
func sortResultsFiltered(results []interface{}) ([]map[string]interface{}, error) {
	filtered := make([]map[string]interface{}, 0, len(results))
	for i, item := range results {
		originalMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("item %d not map: %T", i, item)
		}
		filteredMap := make(map[string]interface{})
		for k, v := range originalMap {
			if k == "path" || k == "name" || k == "kind" { // Only include these for stable comparison
				filteredMap[k] = v
			}
		}
		if _, okPath := filteredMap["path"]; !okPath {
			return nil, fmt.Errorf("filtered map %d missing 'path'. Original: %#v", i, originalMap)
		}
		if _, okName := filteredMap["name"]; !okName {
			return nil, fmt.Errorf("filtered map %d missing 'name'. Original: %#v", i, originalMap)
		}
		filtered = append(filtered, filteredMap)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		mapI, mapJ := filtered[i], filtered[j]
		pathI, pathJ := mapI["path"].(string), mapJ["path"].(string)
		if pathI != pathJ {
			return pathI < pathJ
		}
		nameI, nameJ := mapI["name"].(string), mapJ["name"].(string)
		return nameI < nameJ
	})
	return filtered, nil
}

// --- Test Cases ---
func TestGoFindUsages(t *testing.T) {
	// --- Test Setup ---
	logger, _ := adapters.NewSimpleSlogAdapter(os.Stderr, interfaces.LogLevelDebug)
	logger.Debug("Test logger initialized")
	llmClient := adapters.NewNoOpLLMClient()
	sandboxDir := t.TempDir()

	interpreter, err := core.NewInterpreter(logger, llmClient, sandboxDir, nil, nil)
	if err != nil {
		t.Fatalf("Failed create interpreter: %v", err)
	}

	err = toolsets.RegisterExtendedTools(interpreter)
	if err != nil {
		t.Fatalf("Failed to register extended tools: %v", err)
	}

	err = interpreter.SetSandboxDir(sandboxDir)
	if err != nil {
		t.Fatalf("Failed set sandbox dir: %v", err)
	}
	pkgAPath := filepath.Join(sandboxDir, "pkga")
	if err := os.MkdirAll(pkgAPath, 0755); err != nil {
		t.Fatalf("Failed create fixture dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(pkgAPath, "pkga.go"), []byte(findUsagesFixturePkgAContent), 0644); err != nil {
		t.Fatalf("Failed write pkga.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sandboxDir, "main.go"), []byte(findUsagesFixtureMainContent), 0644); err != nil {
		t.Fatalf("Failed write main.go: %v", err)
	}
	goModContent := []byte("module mytestmodule\n\ngo 1.21\n")
	if err := os.WriteFile(filepath.Join(sandboxDir, "go.mod"), goModContent, 0644); err != nil {
		t.Fatalf("Failed write go.mod: %v", err)
	}
	logger.Debug("Created go.mod in sandbox", "path", filepath.Join(sandboxDir, "go.mod"))

	indexTool, found := interpreter.ToolRegistry().GetTool("Go.IndexCode")
	if !found {
		t.Fatalf("Tool Go.IndexCode not found")
	}
	indexResult, indexErr := indexTool.Func(interpreter, []interface{}{"."})
	if indexErr != nil {
		handleCheck, _ := indexResult.(string)
		if handleCheck == "" {
			t.Fatalf("Go.IndexCode failed: %v", indexErr)
		} else {
			t.Logf("Go.IndexCode warning: %v", indexErr)
		}
	}
	indexHandle, ok := indexResult.(string)
	if !ok || indexHandle == "" {
		t.Fatalf("Go.IndexCode invalid handle: %T %v", indexResult, indexResult)
	}
	t.Logf("Got Semantic Index Handle: %s", indexHandle)

	// --- Define Test Cases ---
	testCases := []struct {
		name       string
		query      string // Semantic query string for the symbol
		wantResult []map[string]interface{}
		wantErr    error
	}{
		{
			name: "Find Usages of Global Constant", query: "package:mytestmodule/pkga; const:GlobalConst",
			wantResult: []map[string]interface{}{
				// Note: Usages typically don't include the declaration itself unless specified by tool.
				// These are expected usages in main.go
				{"path": "main.go", "name": "GlobalConst", "kind": "constant"},
				{"path": "main.go", "name": "GlobalConst", "kind": "constant"},
			},
		},
		{
			name: "Find Usages of Global Variable", query: "package:mytestmodule/pkga; var:GlobalVar",
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "GlobalVar", "kind": "variable"}, // Usage in anotherFunc
				{"path": "main.go", "name": "GlobalVar", "kind": "variable"},      // Usage in main.go
			},
		},
		{
			name: "Find Usages of Type", query: "package:mytestmodule/pkga; type:MyStruct",
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "MyStruct", "kind": "type"}, // PointerMethod receiver type
				{"path": "pkga/pkga.go", "name": "MyStruct", "kind": "type"}, // ValueMethod receiver type
				{"path": "pkga/pkga.go", "name": "MyStruct", "kind": "type"}, // Usage in TopLevelFunc (gs := MyStruct...)
				{"path": "main.go", "name": "MyStruct", "kind": "type"},      // Usage in main (s := thepkga.MyStruct...)
			},
		},
		{
			name: "Find Usages of Function", query: "package:mytestmodule/pkga; function:TopLevelFunc",
			wantResult: []map[string]interface{}{{"path": "main.go", "name": "TopLevelFunc", "kind": "function"}},
		},
		{
			name: "Find Usages of Method", query: "package:mytestmodule/pkga; type:MyStruct; method:PointerMethod",
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "PointerMethod", "kind": "method"}, // Usage in TopLevelFunc (gs.PointerMethod)
				{"path": "main.go", "name": "PointerMethod", "kind": "method"},      // Usage in main (s.PointerMethod)
			},
		},
		{
			name: "Find Usages of Field", query: "package:mytestmodule/pkga; type:MyStruct; field:FieldA",
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "FieldA", "kind": "field"}, // Usage in PointerMethod (s.FieldA)
				{"path": "pkga/pkga.go", "name": "FieldA", "kind": "field"}, // Usage in TopLevelFunc (gs := MyStruct{FieldA: ...})
				{"path": "main.go", "name": "FieldA", "kind": "field"},      // Usage in main (s := thepkga.MyStruct{FieldA: ...})
			},
		},
		{name: "Find Usages of Unexported Function", query: "package:mytestmodule/pkga; function:anotherFunc", wantResult: []map[string]interface{}{}}, // No external usages counted by this tool usually
		{
			name: "Find Usages of Unexported Field", query: "package:mytestmodule/pkga; type:MyStruct; field:fieldB", // Using 'field' for consistency
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "fieldB", "kind": "field"}, // Usage in ValueMethod (return s.fieldB)
				{"path": "pkga/pkga.go", "name": "fieldB", "kind": "field"}, // Usage in TopLevelFunc (gs := MyStruct{..., fieldB: ...})
			},
		},
		{name: "Target Symbol Not Found via Query", query: "package:mytestmodule/pkga; function:ThisDoesNotExist", wantResult: []map[string]interface{}{}},
		{name: "Target Package Not Found via Query", query: "package:nonexistent/pkg; function:SomeFunc", wantResult: []map[string]interface{}{}},
		{name: "Invalid Query - Bad Key", query: "package:mytestmodule/pkga; badkey:abc", wantErr: ErrInvalidQueryFormat}, // This error comes from parseSemanticQuery
		{name: "Invalid Query - Missing Package", query: "function:TopLevelFunc", wantErr: ErrInvalidQueryFormat},         // This error comes from parseSemanticQuery
	}

	// --- Run Tests ---
	findTool, foundFind := interpreter.ToolRegistry().GetTool("Go.FindUsages")
	if !foundFind {
		t.Fatalf("Tool Go.FindUsages not found in registry")
	}

	// Diagnostic logging for the retrieved ToolSpec
	t.Logf("DEBUG: Retrieved ToolSpec for Go.FindUsages. Name: %s, Description: %s, NumArgsExpected: %d",
		findTool.Spec.Name, findTool.Spec.Description, len(findTool.Spec.Args))
	for i, argSpec := range findTool.Spec.Args {
		t.Logf("DEBUG: Arg %d: Name: %s, Type: %s, Required: %t", i, argSpec.Name, argSpec.Type, argSpec.Required)
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			// Call Go.FindUsages with index_handle and the semantic query string
			callArgs := []interface{}{indexHandle, tc.query}
			t.Logf("DEBUG: Calling Go.FindUsages for query '%s' with args: handle=%s, query=%s", tc.query, indexHandle, tc.query)
			result, runErr := findTool.Func(interpreter, callArgs)

			// --- Error Checking ---
			if tc.wantErr != nil {
				if runErr == nil {
					t.Errorf("Expected error wrapping %q, but got nil. Call args: %#v", tc.wantErr, callArgs)
				} else {
					isCorrectError := errors.Is(runErr, tc.wantErr)
					// If wantErr is ErrInvalidQueryFormat, it might be wrapped by ErrInvalidArgument by the tool
					if !isCorrectError && errors.Is(tc.wantErr, ErrInvalidQueryFormat) {
						var rtErr *core.RuntimeError
						if errors.As(runErr, &rtErr) && errors.Is(rtErr.Wrapped, core.ErrInvalidArgument) {
							// Check if the message of the runtime error contains the specific format error
							if underlyingFormatErr := errors.Unwrap(rtErr.Wrapped); underlyingFormatErr != nil && errors.Is(underlyingFormatErr, ErrInvalidQueryFormat) {
								isCorrectError = true
							} else if strings.Contains(rtErr.Message, ErrInvalidQueryFormat.Error()) { // Fallback to string check if not directly wrapped
								isCorrectError = true
							}
						}
					}
					if !isCorrectError {
						t.Errorf("Expected error wrapping %q, but got %q (%T: %v). Call args: %#v", tc.wantErr, runErr, runErr, runErr, callArgs)
					}
				}
				// For error cases, the exact nature of 'result' might vary (nil or empty slice).
				// The primary check is for the correct error.
				return
			}
			if runErr != nil {
				t.Fatalf("Did not expect error for query %q, but got: %v. Call args: %#v", tc.query, runErr, callArgs)
			}

			// --- Result Comparison ---
			actualResultsRaw, ok := result.([]interface{})
			if !ok {
				if result == nil && (tc.wantResult == nil || len(tc.wantResult) == 0) {
					actualResultsRaw = []interface{}{} // Treat nil result as empty if expecting empty
				} else {
					t.Fatalf("Expected result type []interface{}, got %T: %v", result, result)
				}
			}
			actualResultsFiltered, filterErr := sortResultsFiltered(actualResultsRaw)
			if filterErr != nil {
				t.Fatalf("Error filtering/sorting actual results for query %q: %v", tc.query, filterErr)
			}

			var expectedResultsSorted []map[string]interface{}
			if tc.wantResult != nil {
				wantResultInterfaces := make([]interface{}, len(tc.wantResult))
				for i, v := range tc.wantResult {
					wantResultInterfaces[i] = v
				}
				var sortErr error
				expectedResultsSorted, sortErr = sortResultsFiltered(wantResultInterfaces)
				if sortErr != nil {
					t.Fatalf("Internal Test Error: Error sorting expected results for query %q: %v", tc.query, sortErr)
				}
			} else {
				expectedResultsSorted = []map[string]interface{}{} // Handles case where tc.wantResult is nil or empty
			}

			if !reflect.DeepEqual(actualResultsFiltered, expectedResultsSorted) {
				t.Errorf("Result mismatch for query %q (ignoring line/column):\n Expected (sorted): %#v\n Got (sorted/filtered): %#v", tc.query, expectedResultsSorted, actualResultsFiltered)
				t.Logf("Original Got (unsorted, with line/column): %#v", actualResultsRaw)
			}
		})
	}
}
