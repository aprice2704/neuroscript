// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Ignore line/column in comparisons due to fragility.
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
	"github.com/aprice2704/neuroscript/pkg/logging"
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
			return nil, fmt.Errorf("item at index %d is not map[string]interface{}: %T", i, item)
		}
		filteredMap := make(map[string]interface{})
		for k, v := range originalMap {
			if k == "path" || k == "name" || k == "kind" {
				filteredMap[k] = v
			}
		}
		// Ensure essential keys are present after filtering
		if _, ok := filteredMap["path"]; !ok {
			return nil, fmt.Errorf("filtered map at index %d missing 'path' key. Original: %#v", i, originalMap)
		}
		if _, ok := filteredMap["name"]; !ok {
			return nil, fmt.Errorf("filtered map at index %d missing 'name' key. Original: %#v", i, originalMap)
		}
		// Kind might be optional or less critical depending on exact tool needs, but let's keep it for now.
		// if _, ok := filteredMap["kind"]; !ok {
		// 	return nil, fmt.Errorf("filtered map at index %d missing 'kind' key. Original: %#v", i, originalMap)
		// }
		filtered = append(filtered, filteredMap)
	}

	sort.SliceStable(filtered, func(i, j int) bool {
		mapI := filtered[i]
		mapJ := filtered[j]

		// Assume keys exist after filtering logic above
		pathI := mapI["path"].(string)
		pathJ := mapJ["path"].(string)
		if pathI != pathJ {
			return pathI < pathJ
		}

		// Sort by name secondarily for deterministic order
		nameI := mapI["name"].(string)
		nameJ := mapJ["name"].(string)
		return nameI < nameJ

		// Sorting by kind might also be useful if name/path are identical, but less likely needed.
	})
	return filtered, nil
}

// --- Test Cases ---
func TestGoFindUsages(t *testing.T) {
	// --- Test Setup ---
	logger, _ := adapters.NewSimpleSlogAdapter(os.Stderr, logging.LogLevelDebug)
	logger.Debug("Test logger initialized")
	llmClient := adapters.NewNoOpLLMClient()
	sandboxDir := t.TempDir()
	interpreter, err := core.NewInterpreter(logger, llmClient, sandboxDir, nil)
	if err != nil {
		t.Fatalf("Failed create interpreter: %v", err)
	}
	err = core.RegisterCoreTools(interpreter.ToolRegistry())
	if err != nil {
		t.Fatalf("Failed register core tools: %v", err)
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
	logger.Info("Created go.mod in sandbox", "path", filepath.Join(sandboxDir, "go.mod"))
	indexResult, indexErr := toolGoIndexCode(interpreter, []interface{}{"."})
	if indexErr != nil {
		handleCheck, _ := indexResult.(string)
		if handleCheck == "" {
			t.Fatalf("GoIndexCode failed: %v", indexErr)
		} else {
			t.Logf("GoIndexCode warning: %v", indexErr)
		}
	}
	indexHandle, ok := indexResult.(string)
	if !ok || indexHandle == "" {
		t.Fatalf("GoIndexCode invalid handle: %T %v", indexResult, indexResult)
	}
	t.Logf("Got Semantic Index Handle: %s", indexHandle)

	// --- Define Test Cases ---
	// NOTE: wantResult maps now only contain path, name, kind (line/column omitted)
	testCases := []struct {
		name       string
		query      string
		wantResult []map[string]interface{} // Slice of expected usage maps (path, name, kind only)
		wantErr    error
	}{
		{
			name:  "Find Usages of Global Constant",
			query: "package:mytestmodule/pkga; const:GlobalConst",
			wantResult: []map[string]interface{}{
				{"path": "main.go", "name": "GlobalConst", "kind": "constant"},
				{"path": "main.go", "name": "GlobalConst", "kind": "constant"}, // Duplicates are ok if they represent distinct usages
			},
		},
		{
			name:  "Find Usages of Global Variable",
			query: "package:mytestmodule/pkga; var:GlobalVar",
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "GlobalVar", "kind": "variable"},
				{"path": "main.go", "name": "GlobalVar", "kind": "variable"},
			},
		},
		{
			name:  "Find Usages of Type",
			query: "package:mytestmodule/pkga; type:MyStruct",
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "MyStruct", "kind": "type"}, // Pointer receiver
				{"path": "pkga/pkga.go", "name": "MyStruct", "kind": "type"}, // Value receiver
				{"path": "pkga/pkga.go", "name": "MyStruct", "kind": "type"}, // Struct literal
				{"path": "main.go", "name": "MyStruct", "kind": "type"},      // Struct literal
			},
		},
		{
			name:  "Find Usages of Function",
			query: "package:mytestmodule/pkga; function:TopLevelFunc",
			wantResult: []map[string]interface{}{
				{"path": "main.go", "name": "TopLevelFunc", "kind": "function"},
			},
		},
		{
			name:  "Find Usages of Method",
			query: "package:mytestmodule/pkga; type:MyStruct; method:PointerMethod",
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "PointerMethod", "kind": "method"},
				{"path": "main.go", "name": "PointerMethod", "kind": "method"},
			},
		},
		{
			name:  "Find Usages of Field",
			query: "package:mytestmodule/pkga; type:MyStruct; field:FieldA",
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "FieldA", "kind": "field"}, // Usage in method
				{"path": "pkga/pkga.go", "name": "FieldA", "kind": "field"}, // Usage in struct literal
				{"path": "main.go", "name": "FieldA", "kind": "field"},      // Usage in struct literal
			},
		},
		{
			name:       "Find Usages of Unexported Function",
			query:      "package:mytestmodule/pkga; function:anotherFunc",
			wantResult: []map[string]interface{}{}, // Expect empty list
		},
		{
			name:  "Find Usages of Unexported Field",
			query: "package:mytestmodule/pkga; type:MyStruct; var:fieldB",
			wantResult: []map[string]interface{}{
				{"path": "pkga/pkga.go", "name": "fieldB", "kind": "field"},
				{"path": "pkga/pkga.go", "name": "fieldB", "kind": "field"},
			},
		},
		{
			name:       "Target Symbol Not Found via Query",
			query:      "package:mytestmodule/pkga; function:ThisDoesNotExist",
			wantResult: []map[string]interface{}{}, // Expect empty list
		},
		{
			name:       "Target Package Not Found via Query",
			query:      "package:nonexistent/pkg; function:SomeFunc",
			wantResult: []map[string]interface{}{}, // Expect empty list
		},
		{
			name:    "Invalid Query - Bad Key",
			query:   "package:mytestmodule/pkga; badkey:abc",
			wantErr: ErrInvalidQueryFormat,
		},
		{
			name:    "Invalid Query - Missing Package",
			query:   "function:TopLevelFunc",
			wantErr: ErrInvalidQueryFormat,
		},
	}

	// --- Run Tests ---
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel() // Mark tests as parallelizable

			result, runErr := toolGoFindUsages(interpreter, []interface{}{indexHandle, tc.query})

			// --- Error Checking ---
			if tc.wantErr != nil {
				if runErr == nil {
					t.Errorf("Expected error wrapping %q, but got nil", tc.wantErr)
				} else {
					isCorrectError := errors.Is(runErr, tc.wantErr) || (errors.Is(runErr, core.ErrInvalidArgument) && strings.Contains(runErr.Error(), tc.wantErr.Error()))
					if !isCorrectError {
						t.Errorf("Expected error wrapping %q (or ErrInvalidArgument), but got %q (%v)", tc.wantErr, runErr, runErr)
					}
				}
				if result != nil {
					t.Errorf("Expected nil result on error, but got: %v", result)
				}
				return
			}

			// If no error was expected, fail if one occurred
			if runErr != nil {
				t.Fatalf("Did not expect error for query %q, but got: %v", tc.query, runErr)
			}

			// --- Result Comparison ---
			actualResultsRaw, ok := result.([]interface{})
			if !ok {
				t.Fatalf("Expected result type []interface{}, but got %T: %v", result, result)
			}

			// Filter and sort actual results (ignore line/column)
			actualResultsFiltered, filterErr := sortResultsFiltered(actualResultsRaw)
			if filterErr != nil {
				t.Fatalf("Error filtering/sorting actual results for query %q: %v", tc.query, filterErr)
			}

			// Sort expected results (which already lack line/column)
			// Need a temporary slice of interface{} to use sortResultsFiltered for expected results
			wantResultInterfaces := make([]interface{}, len(tc.wantResult))
			for i, v := range tc.wantResult {
				wantResultInterfaces[i] = v
			}
			expectedResultsSorted, filterErr := sortResultsFiltered(wantResultInterfaces) // Sort expected results using the same helper
			if filterErr != nil {
				t.Fatalf("Error sorting expected results for query %q: %v", tc.query, filterErr)
			}

			// Use reflect.DeepEqual for comparison on filtered, sorted slices
			if !reflect.DeepEqual(actualResultsFiltered, expectedResultsSorted) {
				t.Errorf("Result mismatch for query %q (ignoring line/column):\n Expected (sorted): %#v\n Got (sorted/filtered): %#v", tc.query, expectedResultsSorted, actualResultsFiltered)
				// Log original raw result for debugging context
				t.Logf("Original Got (unsorted, with line/column): %#v", actualResultsRaw)
			}
		})
	}
}
