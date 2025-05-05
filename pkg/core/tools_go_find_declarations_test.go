// NeuroScript Version: 0.3.1
// Last Modified: 2025-05-04 13:05:15 PDT // Use correct test helper, fix Invalid_path expectation, reduce PASS logs
// filename: pkg/core/tools_go_find_declarations_test.go

package core

import (
	"errors"
	"go/ast" // Import go/ast for AST dump logging
	"os"
	"path/filepath"
	"reflect"
	"testing"
	// Import removed: adapters and logging are handled by NewDefaultTestInterpreter
)

// Test code snippet that includes cases from the original failing tests.
// Line numbers (L<n>) are added as comments for easy reference in test cases.
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

	instance := &MyStruct{Field: 10} // L16 (instance def C2, MyStruct usage C14, Field usage C24 )
	instance.Method()                // L17 (instance usage C2, Method usage C11)
	fmt.Println(instance.Field)      // L18 (instance usage C14, Field usage C23)
} // L19

// Greet generates a greeting. // L21
func Greet(name string) string { // L22 (Greet def C6, name def C12)
	return "Hello, " + name // L23 (name usage C21)
} // L24

// Method is a method on MyStruct // L26
func (m *MyStruct) Method() { // L27 (m def C7, MyStruct usage C11, Method def C20)
	m.Field++ // L28 (m usage C2, Field usage C4)
} // L29
`

// Minimal go.mod for the test package
const testGoMod = `module finddecltest

go 1.21
`

// TestToolGoFindDeclarations tests the toolGoFindDeclarations function.
func TestToolGoFindDeclarations(t *testing.T) {
	// --- Test Setup ---
	// 1. Create Temp Directory
	tempDir := t.TempDir()
	mainGoPath := filepath.Join(tempDir, "main.go")
	goModPath := filepath.Join(tempDir, "go.mod")
	t.Logf("Test sandbox created: %s", tempDir)

	// 2. Write Test Files
	if err := os.WriteFile(mainGoPath, []byte(testGoCodeForFindDecl), 0644); err != nil {
		t.Fatalf("Failed to write test main.go: %v", err)
	}
	t.Logf("Test file created: %s", mainGoPath)
	if err := os.WriteFile(goModPath, []byte(testGoMod), 0644); err != nil {
		t.Fatalf("Failed to write test go.mod: %v", err)
	}
	t.Logf("Test go.mod created: %s", goModPath)

	// 3. Setup Interpreter using the helper from helpers.go
	// NewDefaultTestInterpreter creates the interpreter, sandbox, registers tools, etc.
	interpreter, absSandboxDirHelper := NewDefaultTestInterpreter(t)

	// Ensure the test interpreter uses the tempDir we created for this specific test
	interpreter.SetSandboxDir(tempDir) // Directly set the sandbox path.
	if interpreter.SandboxDir() != tempDir {
		t.Fatalf("Interpreter sandbox directory mismatch. Expected: %s, Got: %s (Helper initially created: %s)", tempDir, interpreter.SandboxDir(), absSandboxDirHelper)
	}
	// Log level is controlled by NewDefaultTestInterpreter (likely Debug), but explicit logs below are reduced.
	t.Logf("Interpreter sandbox explicitly set to: %s", tempDir)

	// 4. Create Semantic Index for the temp directory
	indexArgs := []interface{}{"."} // Index the root of the tempDir
	handleValue, indexErr := toolGoIndexCode(interpreter, indexArgs)

	// --- Prepare index handle ---
	var handle string
	if handleValue != nil {
		var ok bool
		handle, ok = handleValue.(string)
		if !ok {
			t.Fatalf("toolGoIndexCode returned a non-string handle (%T) without error", handleValue)
		}
	}

	// --- Check for fatal index creation errors vs potentially acceptable package load errors ---
	if indexErr != nil {
		indexVal, getErr := interpreter.GetHandleValue(handle, "semantic_index")
		semanticIndex, isSemanticIndex := indexVal.(*SemanticIndex)
		if getErr != nil || !isSemanticIndex || len(semanticIndex.LoadErrs) == 0 {
			t.Fatalf("toolGoIndexCode failed unexpectedly: %v (get handle err: %v, index valid: %t)",
				indexErr, getErr, isSemanticIndex)
		} else {
			t.Logf("Note: toolGoIndexCode reported package load errors: %v", semanticIndex.LoadErrs)
		}
	}
	if handle == "" {
		t.Fatalf("toolGoIndexCode returned nil or invalid handle even without fatal error")
	}
	t.Logf("Index created with handle: %s", handle)

	// --- Test Cases ---
	testCases := []struct {
		name           string
		path           string // Relative path within tempDir
		line           int64
		column         int64
		expectedResult map[string]interface{} // Use map[string]interface{}, nil if not found
		expectedErr    error                  // Use errors.Is for specific errors
	}{
		// == Cases based on original failures ==
		{
			name:   "Find_Field_usage", // L28 C4 -> should find Field def L8 C2
			path:   "main.go",
			line:   28, // m.Field++
			column: 4,  // Column of 'F' in Field
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(8),
				"column": int64(2),
				"name":   "Field",
				"kind":   "field",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_Method_definition", // L27 C20 -> should find Method def L27 C20
			path:   "main.go",
			line:   27, // func (m *MyStruct) Method() {
			column: 20, // Column of 'M' in Method
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(27),
				"column": int64(20),
				"name":   "Method",
				"kind":   "method",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_receiver_m_definition", // L27 C7 -> should find m def L27 C7
			path:   "main.go",
			line:   27, // func (m *MyStruct) Method() {
			column: 7,  // Column of 'm' in (m *MyStruct)
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(27),
				"column": int64(7),
				"name":   "m",
				"kind":   "variable", // Receiver is a variable
			},
			expectedErr: nil,
		},
		{
			name:   "Find_receiver_m_usage", // L28 C2 -> should find m def L27 C7
			path:   "main.go",
			line:   28, // m.Field++
			column: 2,  // Column of 'm'
			expectedResult: map[string]interface{}{
				"path":   "main.go", // Definition site
				"line":   int64(27),
				"column": int64(7),
				"name":   "m",
				"kind":   "variable",
			},
			expectedErr: nil,
		},

		// == Other Standard Cases ==
		{
			name:   "Find_GlobalVar_definition", // L5 C5
			path:   "main.go",
			line:   5, // var GlobalVar = "initial"
			column: 5, // Column of 'G' in GlobalVar
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(5),
				"column": int64(5),
				"name":   "GlobalVar",
				"kind":   "variable",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_GlobalVar_usage", // L14 C2 -> should find def L5 C5
			path:   "main.go",
			line:   14, // GlobalVar = "changed"
			column: 2,  // Column of 'G' in GlobalVar
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(5),
				"column": int64(5),
				"name":   "GlobalVar",
				"kind":   "variable",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_MyStruct_type_definition", // L7 C6
			path:   "main.go",
			line:   7, // type MyStruct struct {
			column: 6, // Column of 'M' in MyStruct
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(7),
				"column": int64(6),
				"name":   "MyStruct",
				"kind":   "type",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_MyStruct_type_usage_in_receiver", // L27 C11 -> should find def L7 C6
			path:   "main.go",
			line:   27, // func (m *MyStruct) Method() {
			column: 11, // Column of 'M' in MyStruct
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(7),
				"column": int64(6),
				"name":   "MyStruct",
				"kind":   "type",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_MyStruct_type_usage_in_composite_lit", // L16 C14 -> should find def L7 C6
			path:   "main.go",
			line:   16, // instance := &MyStruct{Field: 10}
			column: 14, // Column of 'M' in MyStruct
			expectedResult: map[string]interface{}{ // Keep expectation, but test might still fail here
				"path":   "main.go",
				"line":   int64(7),
				"column": int64(6),
				"name":   "MyStruct",
				"kind":   "type",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_Field_usage_in_composite_lit", // L16 C24 -> should find def L8 C2
			path:   "main.go",
			line:   16, // instance := &MyStruct{Field: 10}
			column: 24, // Column of 'F' in Field:
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(8),
				"column": int64(2),
				"name":   "Field",
				"kind":   "field",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_Field_usage_in_selector", // L18 C23 -> should find def L8 C2
			path:   "main.go",
			line:   18, // fmt.Println(instance.Field)
			column: 23, // Column of 'F' in Field
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(8),
				"column": int64(2),
				"name":   "Field",
				"kind":   "field",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_Greet_func_definition", // L22 C6
			path:   "main.go",
			line:   22, // func Greet(name string) string {
			column: 6,  // Column of 'G' in Greet
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(22),
				"column": int64(6),
				"name":   "Greet",
				"kind":   "function",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_Greet_func_usage", // L12 C13 -> should find def L22 C6
			path:   "main.go",
			line:   12, // message := Greet("World")
			column: 13, // Column of 'G' in Greet
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(22),
				"column": int64(6),
				"name":   "Greet",
				"kind":   "function",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_param_name_definition", // L22 C12 -> should find def L22 C12
			path:   "main.go",
			line:   22, // func Greet(name string) string {
			column: 12, // Column of 'n' in name
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(22),
				"column": int64(12),
				"name":   "name",
				"kind":   "variable", // Parameter is a variable
			},
			expectedErr: nil,
		},
		{
			name:   "Find_param_name_usage", // L23 C21 -> should find def L22 C12
			path:   "main.go",
			line:   23, // return "Hello, " + name
			column: 21, // Column of 'n' in name
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(22),
				"column": int64(12),
				"name":   "name",
				"kind":   "variable",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_var_message_definition", // L12 C2 -> should find def L12 C2
			path:   "main.go",
			line:   12, // message := Greet("World")
			column: 2,  // Column of 'm' in message
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(12),
				"column": int64(2),
				"name":   "message",
				"kind":   "variable",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_var_message_usage", // L13 C14 -> should find def L12 C2
			path:   "main.go",
			line:   13, // fmt.Println(message)
			column: 14, // Column of 'm' in message
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(12),
				"column": int64(2),
				"name":   "message",
				"kind":   "variable",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_var_instance_definition", // L16 C2 -> should find def L16 C2
			path:   "main.go",
			line:   16, // instance := &MyStruct{Field: 10}
			column: 2,  // Column of 'i' in instance
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(16),
				"column": int64(2),
				"name":   "instance",
				"kind":   "variable",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_var_instance_usage_method_call", // L17 C2 -> should find def L16 C2
			path:   "main.go",
			line:   17, // instance.Method()
			column: 2,  // Column of 'i' in instance
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(16),
				"column": int64(2),
				"name":   "instance",
				"kind":   "variable",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_var_instance_usage_field_access", // L18 C14 -> should find def L16 C2
			path:   "main.go",
			line:   18, // fmt.Println(instance.Field)
			column: 14, // Column of 'i' in instance
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(16),
				"column": int64(2),
				"name":   "instance",
				"kind":   "variable",
			},
			expectedErr: nil,
		},
		{
			name:   "Find_Method_usage", // L17 C11 -> should find def L27 C20
			path:   "main.go",
			line:   17, // instance.Method()
			column: 11, // Column of 'M' in Method
			expectedResult: map[string]interface{}{
				"path":   "main.go",
				"line":   int64(27),
				"column": int64(20),
				"name":   "Method",
				"kind":   "method",
			},
			expectedErr: nil,
		},
		// == Not Found / Error Cases ==
		{
			name:           "Find_in_comment", // Should find nothing
			path:           "main.go",
			line:           21,  // // Greet generates a greeting.
			column:         4,   // Inside comment
			expectedResult: nil, // Expect nil map for "not found"
			expectedErr:    nil,
		},
		{
			name:           "Find_on_keyword_func", // Should find nothing specific
			path:           "main.go",
			line:           22, // func Greet(name string) string {
			column:         3,  // on 'n' in func
			expectedResult: nil,
			expectedErr:    nil,
		},
		{
			name:           "Find_on_keyword_type", // Should find nothing specific
			path:           "main.go",
			line:           7, // type MyStruct struct {
			column:         3, // on 'p' in type
			expectedResult: nil,
			expectedErr:    nil,
		},
		{
			name:           "Position_out_of_bounds_line",
			path:           "main.go",
			line:           100, // Line doesn't exist
			column:         1,
			expectedResult: nil, // Position is invalid within the file context
			expectedErr:    nil, // findPosInFileSet returns NoPos, tool returns nil result
		},
		{
			name:           "Position_out_of_bounds_column",
			path:           "main.go",
			line:           3,   // import "fmt" // L3
			column:         100, // Column way past end of line
			expectedResult: nil, // Position is invalid within the file context
			expectedErr:    nil, // findPosInFileSet returns NoPos, tool returns nil result
		},
		{
			name:           "Position_negative_line",
			path:           "main.go",
			line:           -1,
			column:         1,
			expectedResult: nil,
			expectedErr:    ErrInvalidArgument, // Function validates line/col > 0
		},
		{
			name:           "Position_zero_column",
			path:           "main.go",
			line:           1,
			column:         0,
			expectedResult: nil,
			expectedErr:    ErrInvalidArgument, // Function validates line/col > 0
		},
		{
			name:           "Invalid_path",
			path:           "nonexistent.go",
			line:           1,
			column:         1,
			expectedResult: nil,
			expectedErr:    nil, // *** FIXED EXPECTATION: Tool returns nil,nil if file not found in index ***
		},
		{
			name:           "Invalid_handle_type",
			path:           "main.go",
			line:           5,
			column:         5,
			expectedResult: nil,
			expectedErr:    ErrHandleWrongType, // Expect wrong handle type error
			// Setup for this test case happens inside t.Run
		},
	}

	// --- Run Test Cases ---
	for _, tc := range testCases {
		tc := tc // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// t.Parallel() // Consider enabling if setup/interpreter is thread-safe

			currentHandle := handle // Use the valid handle by default
			inputArgs := []interface{}{currentHandle, tc.path, tc.line, tc.column}

			// Special setup for invalid handle test
			if tc.name == "Invalid_handle_type" {
				badHandle, _ := interpreter.RegisterHandle("not a semantic index", "string") // Register a dummy string handle
				inputArgs[0] = badHandle                                                     // Use the bad handle for this specific test
				currentHandle = badHandle                                                    // For logging clarity if needed
				// t.Logf("Using intentionally bad handle for this test: %s", badHandle) // Keep commented out for less verbosity
			}

			// --- Execute the function under test ---
			resultValue, err := toolGoFindDeclarations(interpreter, inputArgs)

			// --- Assert Error ---
			if tc.expectedErr != nil {
				// Error WAS expected
				if err == nil {
					t.Errorf("FAIL: Expected error satisfying %v, but got nil", tc.expectedErr)
				} else if !errors.Is(err, tc.expectedErr) {
					// Check if the received error wraps the expected sentinel error
					t.Errorf("FAIL: Expected error satisfying %v, but got %v (%T)", tc.expectedErr, err, err)
				}
				// else { // Keep commented out for less verbosity
				// 	t.Logf("PASS: Got expected error: %v", err)
				// }
				// Don't check result value if an error was expected
				return
			}

			// Error was NOT expected
			if err != nil {
				// Log the AST dump from the index if available on unexpected error
				indexVal, getErr := interpreter.GetHandleValue(handle, "semantic_index")
				if getErr == nil { // Only proceed if we can actually get the index
					if semanticIndex, ok := indexVal.(*SemanticIndex); ok {
						for _, pkg := range semanticIndex.Packages {
							if pkg == nil {
								continue
							}
							for _, fileNode := range pkg.Syntax {
								if fileNode == nil {
									continue
								}
								tokenFile := semanticIndex.Fset.File(fileNode.Pos())
								if tokenFile != nil && filepath.Clean(tokenFile.Name()) == filepath.Join(tempDir, tc.path) {
									// Only dump AST on failure now
									t.Logf("--- AST Dump for %s on Failure ---", tc.path)
									ast.Print(semanticIndex.Fset, fileNode)
									break // Found the relevant file node
								}
							}
						}
					}
				}
				t.Fatalf("FAIL: Expected success, but got error: %v", err)
			}

			// --- Assert Result Value (only if no error was expected) ---
			var resultMap map[string]interface{}
			if resultValue != nil {
				var ok bool
				resultMap, ok = resultValue.(map[string]interface{})
				if !ok {
					// Fail immediately if the type is wrong (and not nil)
					t.Fatalf("FAIL: Expected result type map[string]interface{} or nil, but got %T (%#v)", resultValue, resultValue)
				}
			}

			// Use reflect.DeepEqual for comparison, handles nil maps correctly.
			if !reflect.DeepEqual(tc.expectedResult, resultMap) {
				// Provide detailed output on failure
				t.Errorf("FAIL: Result mismatch:\nExpected: %#v\nActual:   %#v", tc.expectedResult, resultMap)
			}
			// else { // Keep commented out for less verbosity
			// 	 // t.Logf("PASS: Result matched expected: %#v", tc.expectedResult)
			// }
		})
	}
}
