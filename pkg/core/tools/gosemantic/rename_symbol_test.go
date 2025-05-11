// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Added registration of extended toolsets.
// Test file for GoRenameSymbol tool.
// filename: pkg/core/tools/gosemantic/rename_symbol_test.go

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
	"github.com/aprice2704/neuroscript/pkg/toolsets" // Required for extended toolset registration
)

// --- Fixtures ---
const renameSymbolFixturePkgAContent = `package pkga // L1

import "fmt" // L3

const GlobalConst = 123 // L5 C7
var GlobalVar = "hello" // L6 C5

type MyStruct struct { // L8 C6
	FieldA int      // L9 C2
	fieldB string   // unexported L10 C2
}

func (s *MyStruct) PointerMethod(val string) { // L13 C20
	fmt.Println("Pointer receiver method:", s.FieldA, val) // L14 C46 (s.FieldA)
}

func (s MyStruct) ValueMethod() string { // L17 C19
	return s.fieldB // L18 C12
}

type MyInterface interface { // L21 C6
	DoSomething() error
}

func TopLevelFunc(a int, b string) (string, error) { // L25 C6
	gs := MyStruct{FieldA: a, fieldB: b} // L26 C7 (MyStruct), L26 C18 (FieldA), L26 C29 (fieldB)
	gs.PointerMethod("from func")         // L27 C6 (PointerMethod)
	_ = gs.ValueMethod()                  // L28 C8 (ValueMethod)
	var localVar = "test"
	fmt.Println(localVar)
	return "ok", nil
}

func anotherFunc() { // unexported // L34 C6
	fmt.Println(GlobalVar) // L35 C14 (GlobalVar)
}
`

const renameSymbolFixtureMainContent = `package main // L1

import ( // L3
	"fmt"
	// Use the expected module path after adding go.mod
	thepkga "mytestmodule/pkga" // L6 C2
	//"os" // Commented out to avoid unused import warning in test log
)

func main() { // L10
	fmt.Println(thepkga.GlobalConst) // L11 C15 (GlobalConst)
	s := thepkga.MyStruct{FieldA: 1} // L12 C15 (MyStruct), L12 C24 (FieldA)
	s.PointerMethod("value")         // L13 C4 (PointerMethod)
	fmt.Println(s)
	res, _ := thepkga.TopLevelFunc(thepkga.GlobalConst, thepkga.GlobalVar) // L15 C20 (TopLevelFunc), L15 C46 (GlobalConst), L15 C60 (GlobalVar)
	fmt.Println(res)
}
`

// --- Helper Function ---

// sortAndFilterRenamePatches sorts a slice of patch maps based on path then original_text.
// It returns a NEW slice containing filtered maps (only path, original_text, new_text).
// It also performs basic validation on offsets.
func sortAndFilterRenamePatches(results []interface{}, t *testing.T) ([]map[string]interface{}, error) {
	t.Helper()
	filtered := make([]map[string]interface{}, 0, len(results))
	for i, item := range results {
		originalMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("item %d not map: %T", i, item)
		}
		offsetStart, okS := originalMap["offset_start"].(int64)
		offsetEnd, okE := originalMap["offset_end"].(int64)
		if !okS || !okE {
			return nil, fmt.Errorf("patch %d bad offsets: %#v", i, originalMap)
		}
		if offsetStart < 0 || offsetEnd <= offsetStart {
			return nil, fmt.Errorf("patch %d invalid offsets: %d, %d", i, offsetStart, offsetEnd)
		}
		path, okP := originalMap["path"].(string)
		originalText, okO := originalMap["original_text"].(string)
		newText, okN := originalMap["new_text"].(string)
		if !okP || !okO || !okN {
			return nil, fmt.Errorf("patch %d missing keys: %#v", i, originalMap)
		}
		filteredMap := map[string]interface{}{"path": path, "original_text": originalText, "new_text": newText}
		filtered = append(filtered, filteredMap)
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		mapI, mapJ := filtered[i], filtered[j]
		pathI, pathJ := mapI["path"].(string), mapJ["path"].(string)
		if pathI != pathJ {
			return pathI < pathJ
		}
		origI, origJ := mapI["original_text"].(string), mapJ["original_text"].(string)
		return origI < origJ
	})
	return filtered, nil
}

// --- Test Cases ---
func TestGoRenameSymbol(t *testing.T) {
	// --- Test Setup ---
	logger, _ := adapters.NewSimpleSlogAdapter(os.Stderr, logging.LogLevelDebug)
	logger.Debug("Test logger initialized")
	llmClient := adapters.NewNoOpLLMClient()
	sandboxDir := t.TempDir()

	interpreter, err := core.NewInterpreter(logger, llmClient, sandboxDir, nil, nil)
	if err != nil {
		t.Fatalf("Failed create interpreter: %v", err)
	}
	// core.RegisterCoreTools is called within NewInterpreter

	// Register extended toolsets, including GoSemantic which provides Go.IndexCode
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
	if err := os.WriteFile(filepath.Join(pkgAPath, "pkga.go"), []byte(renameSymbolFixturePkgAContent), 0644); err != nil {
		t.Fatalf("Failed write pkga.go: %v", err)
	}
	if err := os.WriteFile(filepath.Join(sandboxDir, "main.go"), []byte(renameSymbolFixtureMainContent), 0644); err != nil {
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
	indexResult, indexErr := indexTool.Func(interpreter, []interface{}{"."}) // Index current directory
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
		name        string
		query       string
		newName     string
		wantPatches []map[string]interface{} // Expected patches (path, original_text, new_text only)
		wantErr     error
	}{
		{
			name: "Rename Global Constant", query: "package:mytestmodule/pkga; const:GlobalConst", newName: "GlobalConstant",
			wantPatches: []map[string]interface{}{
				{"path": "pkga/pkga.go", "original_text": "GlobalConst", "new_text": "GlobalConstant"},
				{"path": "main.go", "original_text": "GlobalConst", "new_text": "GlobalConstant"},
				{"path": "main.go", "original_text": "GlobalConst", "new_text": "GlobalConstant"},
			},
		},
		{
			name: "Rename Global Variable", query: "package:mytestmodule/pkga; var:GlobalVar", newName: "GlobalVariable",
			wantPatches: []map[string]interface{}{
				{"path": "pkga/pkga.go", "original_text": "GlobalVar", "new_text": "GlobalVariable"},
				{"path": "pkga/pkga.go", "original_text": "GlobalVar", "new_text": "GlobalVariable"},
				{"path": "main.go", "original_text": "GlobalVar", "new_text": "GlobalVariable"},
			},
		},
		{
			name: "Rename Type", query: "package:mytestmodule/pkga; type:MyStruct", newName: "MyStructure",
			wantPatches: []map[string]interface{}{
				{"path": "pkga/pkga.go", "original_text": "MyStruct", "new_text": "MyStructure"}, {"path": "pkga/pkga.go", "original_text": "MyStruct", "new_text": "MyStructure"},
				{"path": "pkga/pkga.go", "original_text": "MyStruct", "new_text": "MyStructure"}, {"path": "pkga/pkga.go", "original_text": "MyStruct", "new_text": "MyStructure"},
				{"path": "main.go", "original_text": "MyStruct", "new_text": "MyStructure"},
			},
		},
		{
			name: "Rename Function", query: "package:mytestmodule/pkga; function:TopLevelFunc", newName: "TopLevelFunction",
			wantPatches: []map[string]interface{}{
				{"path": "pkga/pkga.go", "original_text": "TopLevelFunc", "new_text": "TopLevelFunction"},
				{"path": "main.go", "original_text": "TopLevelFunc", "new_text": "TopLevelFunction"},
			},
		},
		{
			name: "Rename Method", query: "package:mytestmodule/pkga; type:MyStruct; method:PointerMethod", newName: "PointerReceiverMethod",
			wantPatches: []map[string]interface{}{
				{"path": "pkga/pkga.go", "original_text": "PointerMethod", "new_text": "PointerReceiverMethod"},
				{"path": "pkga/pkga.go", "original_text": "PointerMethod", "new_text": "PointerReceiverMethod"},
				{"path": "main.go", "original_text": "PointerMethod", "new_text": "PointerReceiverMethod"},
			},
		},
		{
			name: "Rename Field", query: "package:mytestmodule/pkga; type:MyStruct; field:FieldA", newName: "FieldAlpha",
			wantPatches: []map[string]interface{}{
				{"path": "pkga/pkga.go", "original_text": "FieldA", "new_text": "FieldAlpha"}, {"path": "pkga/pkga.go", "original_text": "FieldA", "new_text": "FieldAlpha"},
				{"path": "pkga/pkga.go", "original_text": "FieldA", "new_text": "FieldAlpha"}, {"path": "main.go", "original_text": "FieldA", "new_text": "FieldAlpha"},
			},
		},
		{
			name: "Rename Unexported Field", query: "package:mytestmodule/pkga; type:MyStruct; field:fieldB", newName: "fieldBeta",
			wantPatches: []map[string]interface{}{
				{"path": "pkga/pkga.go", "original_text": "fieldB", "new_text": "fieldBeta"},
				{"path": "pkga/pkga.go", "original_text": "fieldB", "new_text": "fieldBeta"},
				{"path": "pkga/pkga.go", "original_text": "fieldB", "new_text": "fieldBeta"},
			},
		},
		{
			name: "Rename Unexported Function", query: "package:mytestmodule/pkga; function:anotherFunc", newName: "anotherFunction",
			wantPatches: []map[string]interface{}{{"path": "pkga/pkga.go", "original_text": "anotherFunc", "new_text": "anotherFunction"}},
		},
		{name: "Rename Symbol Not Found", query: "package:mytestmodule/pkga; function:NoSuchFunc", newName: "NewFuncName", wantPatches: []map[string]interface{}{}},
		{name: "Rename Package Not Found", query: "package:nonexistent/pkg; function:SomeFunc", newName: "NewFuncName", wantPatches: []map[string]interface{}{}},
		{name: "Rename Same Name", query: "package:mytestmodule/pkga; const:GlobalConst", newName: "GlobalConst", wantPatches: []map[string]interface{}{}}, // Should result in no patches
		{name: "Rename Invalid New Name", query: "package:mytestmodule/pkga; const:GlobalConst", newName: "Invalid-Name", wantErr: core.ErrInvalidArgument},
		{name: "Rename to Keyword", query: "package:mytestmodule/pkga; var:GlobalVar", newName: "type", wantErr: core.ErrInvalidArgument},
		{name: "Rename Builtin Type (Expect Empty)", query: "package:builtin; type:string", newName: "MyString", wantPatches: []map[string]interface{}{}}, // Should not find/rename builtins
	}

	// --- Run Tests ---
	renameTool, foundRename := interpreter.ToolRegistry().GetTool("Go.RenameSymbol")
	if !foundRename {
		t.Fatalf("Tool Go.RenameSymbol not found in registry")
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() // Disabling parallel for rename tests as they might interact with index or common state if not careful

			result, runErr := renameTool.Func(interpreter, []interface{}{indexHandle, tc.query, tc.newName})

			// --- Error Checking ---
			if tc.wantErr != nil {
				if runErr == nil {
					t.Errorf("Expected error wrapping %q, but got nil", tc.wantErr)
				} else {
					isCorrectError := errors.Is(runErr, tc.wantErr)
					// Allow direct match for ErrInvalidArgument as it might not always be wrapped by RuntimeError in the same way
					if !isCorrectError && tc.wantErr == core.ErrInvalidArgument {
						isCorrectError = errors.Is(runErr, core.ErrInvalidArgument)
					}
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
				t.Fatalf("Did not expect error for query %q, newName %q, but got: %v", tc.query, tc.newName, runErr)
			}

			// --- Result Comparison ---
			actualResultsRaw, ok := result.([]interface{})
			if !ok {
				if result == nil && (tc.wantPatches == nil || len(tc.wantPatches) == 0) {
					actualResultsRaw = []interface{}{} // Treat nil result as empty if expecting empty
				} else {
					t.Fatalf("Expected result type []interface{}, got %T: %v", result, result)
				}
			}
			actualResultsFiltered, filterErr := sortAndFilterRenamePatches(actualResultsRaw, t)
			if filterErr != nil {
				t.Fatalf("Error filtering/sorting actual results for query %q: %v\nActual Raw Results: %#v", tc.query, filterErr, actualResultsRaw)
			}

			var expectedResultsSorted []map[string]interface{}
			if tc.wantPatches != nil {
				wantResultInterfaces := make([]interface{}, len(tc.wantPatches))
				for i, v := range tc.wantPatches {
					// Add dummy offsets as they are not compared but needed by sortAndFilterRenamePatches
					vCopy := make(map[string]interface{})
					for kOriginal, vOriginal := range v {
						vCopy[kOriginal] = vOriginal
					}
					vCopy["offset_start"] = int64(0) // Dummy value
					vCopy["offset_end"] = int64(1)   // Dummy value
					wantResultInterfaces[i] = vCopy
				}
				var sortErr error
				expectedResultsSorted, sortErr = sortAndFilterRenamePatches(wantResultInterfaces, t)
				if sortErr != nil {
					t.Fatalf("Internal Test Error: Error sorting expected results for query %q: %v", tc.query, sortErr)
				}
			} else {
				expectedResultsSorted = []map[string]interface{}{}
			}

			if !reflect.DeepEqual(actualResultsFiltered, expectedResultsSorted) {
				if len(actualResultsFiltered) == 0 && len(expectedResultsSorted) == 0 {
					// This is fine.
				} else {
					t.Errorf("Result mismatch for query %q -> %q (ignoring offsets):\n Expected (sorted): %#v\n Got (sorted/filtered): %#v", tc.query, tc.newName, expectedResultsSorted, actualResultsFiltered)
					t.Logf("Original Got (unsorted, with offsets): %#v", actualResultsRaw)
				}
			} else if len(expectedResultsSorted) > 0 {
				t.Logf("Successfully matched %d patch operations for query %q", len(expectedResultsSorted), tc.query)
			}
		})
	}
}
