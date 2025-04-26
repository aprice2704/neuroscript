// filename: pkg/core/tools_go_ast_symbol_map_test.go
package goast

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	// log import removed
)

// --- Test Setup ---

// const testModuleName = "testtool" // <<< REMOVED duplicate declaration <<<

// Helper function to create a temporary test environment for symbol map tests
func setupSymbolMapTestEnv(t *testing.T, files map[string]string) (string, func()) {
	t.Helper()
	rootDir := t.TempDir()
	logTest(t, "Test rootDir: %s", rootDir) // Use logTest from universal_test_helpers

	// Create go.mod using the constant defined in tools_go_ast_package_test.go
	goModPath := filepath.Join(rootDir, "go.mod")
	// Assuming testModuleName is accessible (defined in tools_go_ast_package_test.go)
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21\n", testModuleName)
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}
	logTest(t, "Writing %d bytes to: %s", len(goModContent), "go.mod") // Use logTest

	for name, content := range files {
		filePath := filepath.Join(rootDir, name)
		dirPath := filepath.Dir(filePath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", name, err)
		}
		logTest(t, "Writing %d bytes to: %s", len(content), name) // Use logTest

	}

	cleanup := func() {
		// t.TempDir handles cleanup automatically
	}

	return rootDir, cleanup
}

// --- Test Cases ---

func TestBuildSymbolMapLogic(t *testing.T) {
	// Define struct locally within the test function
	type testCase struct {
		name              string
		files             map[string]string
		packagePath       string // Path relative to module root where refactored pkgs live
		expectedMap       map[string]string
		expectedErrorType error // Use sentinel error type
	}

	testCases := []testCase{
		{
			name: "Basic case with multiple types",
			files: map[string]string{
				"testtool/refactored/sub1/file1.go": `package sub1
import "fmt"
var ExportedVarOne = "hello"
type ExportedTypeOne struct{}
func ExportedFuncOne() { fmt.Println(ExportedVarOne) }
const internalConst = 1
`,
				"testtool/refactored/sub2/file2.go": `package sub2
const ExportedConstTwo = 123
type ExportedTypeTwo int
var internalVar = "world"
func internalFunc() {}
`,
				"testtool/refactored/sub2/empty.go":      "package sub2",
				"testtool/refactored/sub2/file2_test.go": `package sub2; import "testing"; func TestDummy(t *testing.T) {}`,
			},
			packagePath: "testtool/refactored", // This is the input path relative to root
			expectedMap: map[string]string{
				// Expected paths should be the correct canonical import paths
				"ExportedVarOne":   "testtool/refactored/sub1",
				"ExportedTypeOne":  "testtool/refactored/sub1",
				"ExportedFuncOne":  "testtool/refactored/sub1",
				"ExportedConstTwo": "testtool/refactored/sub2",
				"ExportedTypeTwo":  "testtool/refactored/sub2",
			},
			expectedErrorType: nil,
		},
		{
			name: "Ambiguous symbols",
			files: map[string]string{
				"testtool/refactored/sub1/ambig1.go": `package sub1
var AmbiguousVar = "from sub1"
func AmbiguousFunc() {}
`,
				"testtool/refactored/sub2/ambig2.go": `package sub2
var AnotherVar = 1
func AmbiguousFunc() {}
type AmbiguousType struct{}
`,
				"testtool/refactored/sub3/ambig3.go": `package sub3
const AmbiguousConst = 1
type AmbiguousType struct{}
`,
			},
			packagePath:       "testtool/refactored",
			expectedMap:       nil,
			expectedErrorType: ErrAmbiguousSymbol,
		},
		{
			name: "No Go files",
			files: map[string]string{
				"testtool/refactored/sub1/README.md": "Just a readme",
			},
			packagePath:       "testtool/refactored",
			expectedMap:       map[string]string{},
			expectedErrorType: nil,
		},
		{
			name: "Directory not found",
			files: map[string]string{
				"otherdir/file.go": "package otherdir",
			},
			packagePath:       "nonexistent/path",
			expectedMap:       nil,
			expectedErrorType: ErrRefactoredPathNotFound,
		},
		{
			name: "Nested packages",
			files: map[string]string{
				"nested/pkg/sub/file1.go": `package sub
var ExportedVar = "v"`,
				"nested/pkg/file2.go": `package pkg
const ExportedConst = 1`,
			},
			packagePath: "nested/pkg", // Path relative to module root
			expectedMap: map[string]string{
				// Expect paths relative to module root
				"ExportedVar":   "nested/pkg/sub",
				"ExportedConst": "nested/pkg",
			},
			expectedErrorType: nil,
		},
		{
			name: "Package path outside root",
			files: map[string]string{
				"actual/pkg/file.go": "package pkg",
			},
			packagePath:       "../outside",
			expectedMap:       nil,
			expectedErrorType: ErrPathViolation, // Expect sentinel error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootDir, cleanup := setupSymbolMapTestEnv(t, tc.files)
			defer cleanup()

			interpreter, _ := newDefaultTestInterpreter(t)
			interpreter.sandboxDir = rootDir

			// --- Execute ---
			symbolMap, err := buildSymbolMap(tc.packagePath, interpreter)

			// --- Assert Error using errors.Is ---
			if tc.expectedErrorType != nil {
				if err == nil {
					t.Fatalf("Expected error wrapping [%v], but got nil error", tc.expectedErrorType)
				}
				if !errors.Is(err, tc.expectedErrorType) {
					t.Errorf("Expected error to wrap '%v', but errors.Is is false. Got error: %v (Type: %T)", tc.expectedErrorType, err, err)
				} else {
					t.Logf("Got expected error type: %v", err)
				}
			} else {
				if err != nil {
					t.Fatalf("Expected no error, but got: %v", err)
				}
			}

			// --- Assert Map Content (only if no error expected) ---
			if tc.expectedErrorType == nil {
				expectedIsEmpty := len(tc.expectedMap) == 0
				actualIsEmpty := len(symbolMap) == 0

				if expectedIsEmpty && actualIsEmpty {
					t.Logf("Both expected and actual maps are empty/nil.")
				} else if !reflect.DeepEqual(tc.expectedMap, symbolMap) {
					t.Errorf("Returned map does not match expected.")
					diff := cmp.Diff(tc.expectedMap, symbolMap)
					if diff != "" {
						t.Errorf("Map diff (-expected +got):\n%s", diff)
					} else {
						t.Logf("  Expected: %#v", tc.expectedMap)
						t.Logf("  Got:      %#v", symbolMap)
					}
				}
			} else {
				if len(symbolMap) > 0 {
					t.Errorf("Expected nil or empty map when error [%v] occurred, but got: %#v", tc.expectedErrorType, symbolMap)
				}
			}
		})
	}
}
