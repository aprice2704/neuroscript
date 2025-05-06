// NeuroScript Version: 0.3.1
// File version: 0.0.2
// Correct expected column for MyStruct usage, expect empty slice for not found cases.
// filename: pkg/core/tools_go_find_usages_test.go

package core

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// Re-use the same test code as find declarations for finding usages
// Line numbers (L<n>) are added as comments for easy reference in test cases.
/*
const testGoCodeForFindDecl = `package main // L1

import "fmt" // L3

var GlobalVar = "initial" // L5

type MyStruct struct { // L7
	Field int // L8
} // L9

func main() { // L11
	message := Greet("World") // L12 (message def C2, Greet usage C13)
	fmt.Println(message)      // L13 (message usage C14, fmt usage C2, Println C6)
	GlobalVar = "changed"   // L14 (GlobalVar usage C2)

	instance := &MyStruct{Field: 10} // L16 (instance def C2, MyStruct usage C15, Field usage C24 ) Corrected C15
	instance.Method()                // L17 (instance usage C2, Method usage C11)
	fmt.Println(instance.Field)      // L18 (instance usage C14, Field usage C23)
} // L19

// Greet generates a greeting. // L21
func Greet(name string) string { // L22 (Greet def C6, name def C12)
	return "Hello, " + name // L23 (name usage C21)
} // L24

// Method is a method on MyStruct // L26
func (m *MyStruct) Method() { // L27 (m def C7, MyStruct usage C10, Method def C20) Corrected C10
	m.Field++ // L28 (m usage C2, Field usage C4)
} // L29
`

const testGoMod = `module findusagestest

go 1.21
`
*/
// NOTE: Using the constants from tools_go_find_declarations_test.go directly
// to avoid duplication. Ensure that file exists and is part of the test build.

// setupFindUsagesTest creates a temporary directory, writes test Go code and go.mod,
// initializes an interpreter, creates a semantic index, and returns the interpreter and handle.
func setupFindUsagesTest(t *testing.T) (*Interpreter, string) {
	t.Helper()
	// Use the exact same setup code structure as TestToolGoFindDeclarations
	tempDir := t.TempDir()
	mainGoPath := filepath.Join(tempDir, "main.go")
	goModPath := filepath.Join(tempDir, "go.mod")
	// t.Logf("Test sandbox created: %s", tempDir) // Reduce log noise

	if err := os.WriteFile(mainGoPath, []byte(testGoCodeForFindDecl), 0644); err != nil {
		t.Fatalf("Setup: Failed to write test main.go: %v", err)
	}
	// t.Logf("Test file created: %s", mainGoPath)
	if err := os.WriteFile(goModPath, []byte(testGoMod), 0644); err != nil {
		t.Fatalf("Setup: Failed to write test go.mod: %v", err)
	}
	// t.Logf("Test go.mod created: %s", goModPath)

	interpreter, _ := NewDefaultTestInterpreter(t) // Ignore initial sandbox dir from helper
	interpreter.SetSandboxDir(tempDir)
	if interpreter.SandboxDir() != tempDir {
		t.Fatalf("Setup: Interpreter sandbox directory mismatch. Expected: %s, Got: %s", tempDir, interpreter.SandboxDir())
	}
	// t.Logf("Interpreter sandbox explicitly set to: %s", tempDir)

	// Create Semantic Index
	indexArgs := []interface{}{"."}
	handleValue, indexErr := toolGoIndexCode(interpreter, indexArgs)

	var handle string
	if handleValue != nil {
		var ok bool
		handle, ok = handleValue.(string)
		if !ok {
			t.Fatalf("Setup: toolGoIndexCode returned a non-string handle (%T)", handleValue)
		}
	}
	if indexErr != nil {
		// Check if it's just package load errors vs fatal index error
		indexVal, getErr := interpreter.GetHandleValue(handle, "semantic_index")
		semanticIndex, isSemanticIndex := indexVal.(*SemanticIndex)
		if getErr != nil || !isSemanticIndex || len(semanticIndex.LoadErrs) == 0 {
			t.Fatalf("Setup: toolGoIndexCode failed unexpectedly: %v (get handle err: %v, index valid: %t)", indexErr, getErr, isSemanticIndex)
		} else {
			// t.Logf("Setup Note: toolGoIndexCode reported package load errors: %v", semanticIndex.LoadErrs)
		}
	}
	if handle == "" {
		t.Fatalf("Setup: toolGoIndexCode returned nil or invalid handle")
	}
	// t.Logf("Index created with handle: %s", handle)
	return interpreter, handle
}

// sortResults sorts the slice of usage maps for consistent comparison.
func sortResults(results []map[string]interface{}) {
	sort.SliceStable(results, func(i, j int) bool {
		resI, resJ := results[i], results[j]
		pathI, _ := resI["path"].(string)
		pathJ, _ := resJ["path"].(string)
		if pathI != pathJ {
			return pathI < pathJ
		}
		lineI, _ := resI["line"].(int64)
		lineJ, _ := resJ["line"].(int64)
		if lineI != lineJ {
			return lineI < lineJ
		}
		colI, _ := resI["column"].(int64)
		colJ, _ := resJ["column"].(int64)
		return colI < colJ
	})
}

// compareUsageResults checks if two slices of usage results are deeply equal after sorting.
func compareUsageResults(t *testing.T, expected, actual interface{}) bool {
	t.Helper()

	// --- Updated Comparison Logic ---
	var expectedSlice []map[string]interface{}
	var actualSlice []map[string]interface{}
	var ok bool

	// Handle expected nil or empty slice
	if expected == nil {
		expectedSlice = []map[string]interface{}{} // Treat expected nil as empty slice for comparison
	} else {
		expectedSlice, ok = expected.([]map[string]interface{})
		if !ok {
			t.Errorf("Type mismatch: Expected value is not []map[string]interface{} or nil, got %T", expected)
			return false
		}
		if expectedSlice == nil { // Handle case where expected is explicitly `[]map[string]interface{}(nil)`
			expectedSlice = []map[string]interface{}{}
		}
	}

	// Handle actual nil or empty slice
	if actual == nil {
		actualSlice = []map[string]interface{}{} // Treat actual nil as empty slice for comparison
	} else {
		actualSlice, ok = actual.([]map[string]interface{})
		if !ok {
			t.Errorf("Type mismatch: Actual value is not []map[string]interface{} or nil, got %T", actual)
			return false
		}
		if actualSlice == nil { // Handle case where actual is explicitly `[]map[string]interface{}(nil)`
			actualSlice = []map[string]interface{}{}
		}
	}
	// --- End Updated Comparison Logic ---

	// Sort both slices before comparison
	sortResults(expectedSlice)
	sortResults(actualSlice)

	if !reflect.DeepEqual(expectedSlice, actualSlice) {
		t.Errorf("Result mismatch (after sorting and normalizing nil/empty):\nExpected: %#v\nActual:   %#v", expectedSlice, actualSlice)
		return false
	}
	return true
}

// TestToolGoFindUsages tests the toolGoFindUsages function.
func TestToolGoFindUsages(t *testing.T) {
	interpreter, handle := setupFindUsagesTest(t) // Setup shared index

	testCases := []struct {
		name        string
		path        string      // Relative path
		line        int64       // Line of the identifier to find usages for
		column      int64       // Column of the identifier
		expected    interface{} // Use interface{} to allow expecting nil or empty slice
		expectedErr error
	}{
		// --- Happy Path Cases ---
		{
			name:   "Find usages of GlobalVar (from def)",
			path:   "main.go",
			line:   5, // Definition: var GlobalVar = "initial"
			column: 5, // 'G'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(14), "column": int64(2), "name": "GlobalVar"}, // Usage: GlobalVar = "changed"
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of GlobalVar (from usage)",
			path:   "main.go",
			line:   14, // Usage: GlobalVar = "changed"
			column: 2,  // 'G'
			expected: []map[string]interface{}{
				// Expect the same usage list, even starting from a usage site
				{"path": "main.go", "line": int64(14), "column": int64(2), "name": "GlobalVar"},
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of MyStruct (from def)",
			path:   "main.go",
			line:   7, // Definition: type MyStruct struct {
			column: 6, // 'M'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(16), "column": int64(15), "name": "MyStruct"}, // Usage: instance := &MyStruct{...}
				{"path": "main.go", "line": int64(27), "column": int64(10), "name": "MyStruct"}, // Usage: func (m *MyStruct) Method() <-- CORRECTED COLUMN
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of MyStruct (from usage in receiver)",
			path:   "main.go",
			line:   27, // Usage: func (m *MyStruct) Method()
			column: 10, // 'M' <-- CORRECTED COLUMN
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(16), "column": int64(15), "name": "MyStruct"},
				{"path": "main.go", "line": int64(27), "column": int64(10), "name": "MyStruct"}, // <-- CORRECTED COLUMN
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of Field (from def)",
			path:   "main.go",
			line:   8, // Definition: Field int
			column: 2, // 'F'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(16), "column": int64(24), "name": "Field"}, // Usage: MyStruct{Field: 10}
				{"path": "main.go", "line": int64(18), "column": int64(23), "name": "Field"}, // Usage: instance.Field
				{"path": "main.go", "line": int64(28), "column": int64(4), "name": "Field"},  // Usage: m.Field++
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of Field (from usage m.Field++)",
			path:   "main.go",
			line:   28, // Usage: m.Field++
			column: 4,  // 'F'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(16), "column": int64(24), "name": "Field"},
				{"path": "main.go", "line": int64(18), "column": int64(23), "name": "Field"},
				{"path": "main.go", "line": int64(28), "column": int64(4), "name": "Field"},
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of Greet (from def)",
			path:   "main.go",
			line:   22, // Definition: func Greet(...)
			column: 6,  // 'G'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(12), "column": int64(13), "name": "Greet"}, // Usage: message := Greet("World")
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of message (from def)",
			path:   "main.go",
			line:   12, // Definition: message := ...
			column: 2,  // 'm'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(13), "column": int64(14), "name": "message"}, // Usage: fmt.Println(message)
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of instance (from def)",
			path:   "main.go",
			line:   16, // Definition: instance := ...
			column: 2,  // 'i'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(17), "column": int64(2), "name": "instance"},  // Usage: instance.Method()
				{"path": "main.go", "line": int64(18), "column": int64(14), "name": "instance"}, // Usage: fmt.Println(instance.Field)
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of name param (from def)",
			path:   "main.go",
			line:   22, // Definition: func Greet(name string)
			column: 12, // 'n'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(23), "column": int64(21), "name": "name"}, // Usage: return "Hello, " + name
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of m receiver (from def)",
			path:   "main.go",
			line:   27, // Definition: func (m *MyStruct)
			column: 7,  // 'm'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(28), "column": int64(2), "name": "m"}, // Usage: m.Field++
			},
			expectedErr: nil,
		},
		{
			name:   "Find usages of Method (from def)",
			path:   "main.go",
			line:   27, // Definition: Method()
			column: 20, // 'M'
			expected: []map[string]interface{}{
				{"path": "main.go", "line": int64(17), "column": int64(11), "name": "Method"}, // Usage: instance.Method()
			},
			expectedErr: nil,
		},
		// --- Not Found / Error Cases ---
		{
			name:        "Target identifier not found (in comment)",
			path:        "main.go",
			line:        21,                         // // Greet generates a greeting. // L21
			column:      4,                          // 'e' in Greet
			expected:    []map[string]interface{}{}, // Expect empty slice for not found <-- CORRECTED EXPECTATION
			expectedErr: nil,
		},
		{
			name:        "Target is package name (fmt)",
			path:        "main.go",
			line:        13,                         // fmt.Println(message)
			column:      2,                          // 'f' in fmt
			expected:    []map[string]interface{}{}, // Expect empty slice for package names
			expectedErr: nil,
		},
		{
			name:        "Invalid position (out of bounds line)",
			path:        "main.go",
			line:        100,
			column:      1,
			expected:    []map[string]interface{}{}, // Expect empty slice for invalid pos <-- CORRECTED EXPECTATION
			expectedErr: nil,
		},
		{
			name:        "Invalid handle",
			path:        "main.go",
			line:        5, // GlobalVar def
			column:      5,
			expected:    nil, // Expect nil interface{} for error cases still
			expectedErr: ErrHandleWrongType,
		},
		{
			name:        "Invalid argument type (line)",
			path:        "main.go",
			line:        0, // Invalid line number, will fail validation
			column:      5,
			expected:    nil,
			expectedErr: ErrInvalidArgument,
		},
	}

	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			currentHandle := handle
			inputArgs := []interface{}{currentHandle, tc.path, tc.line, tc.column}

			// Special setup for invalid handle test
			if tc.name == "Invalid handle" {
				badHandle, _ := interpreter.RegisterHandle("not an index", "string")
				inputArgs[0] = badHandle
			}

			resultValue, err := toolGoFindUsages(interpreter, inputArgs)

			// Assert Error
			if tc.expectedErr != nil {
				if err == nil {
					t.Errorf("FAIL: Expected error satisfying %v, but got nil", tc.expectedErr)
				} else if !errors.Is(err, tc.expectedErr) {
					t.Errorf("FAIL: Expected error satisfying %v, but got %v (%T)", tc.expectedErr, err, err)
				}
				// Don't check result value if an error was expected
				return
			}
			// No error expected
			if err != nil {
				t.Fatalf("FAIL: Expected success, but got error: %v", err)
			}

			// Assert Result Value (using custom comparison)
			if !compareUsageResults(t, tc.expected, resultValue) {
				// compareUsageResults already logged the error details
			}
		})
	}
}
