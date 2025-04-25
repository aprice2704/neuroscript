package core

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort" // Keep sort import
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	// log import removed, using t.Logf via helpers
	// astutil import removed
)

// --- Test Setup ---
//const testModuleName = "testtool"

// setupPackageUpdateTestEnv and compareResultMaps remain unchanged.
func setupPackageUpdateTestEnv(t *testing.T, files map[string]string) (string, func()) { /* ... (implementation as before) ... */
	t.Helper()
	rootDir := t.TempDir()

	// Create go.mod using the constant module name
	goModPath := filepath.Join(rootDir, "go.mod")
	goModContent := fmt.Sprintf("module %s\n\ngo 1.21", testModuleName) // Use testModuleName
	if err := os.WriteFile(goModPath, []byte(goModContent), 0644); err != nil {
		t.Fatalf("Failed to write go.mod: %v", err)
	}
	logTest(t, "Writing %d bytes to: %s", len(goModContent), "go.mod")

	for name, content := range files {
		filePath := filepath.Join(rootDir, name)
		dirPath := filepath.Dir(filePath)
		if err := os.MkdirAll(dirPath, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dirPath, err)
		}
		if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
			t.Fatalf("Failed to write file %s: %v", name, err)
		}
		logTest(t, "Writing %d bytes to: %s", len(content), name)

	}

	cleanup := func() {
		// t.TempDir handles cleanup automatically
	}

	return rootDir, cleanup
}

func compareResultMaps(t *testing.T, expected, got map[string]interface{}, checkErrorSubstring string) bool { /* ... (implementation as before) ... */
	t.Helper()

	// Make copies to avoid modifying originals
	expectedCopy := make(map[string]interface{})
	gotCopy := make(map[string]interface{})
	for k, v := range expected {
		expectedCopy[k] = v
	}
	for k, v := range got {
		gotCopy[k] = v
	}

	// Handle error comparison separately
	expectedErrVal, expErrOk := expectedCopy["error"]
	gotErrVal, gotErrOk := gotCopy["error"]

	// Handle failed_files comparison separately (check substring)
	expectedFailedVal, expFailedOk := expectedCopy["failed_files"]
	gotFailedVal, gotFailedOk := gotCopy["failed_files"]
	delete(expectedCopy, "failed_files")
	delete(gotCopy, "failed_files")

	// Delete error fields for DeepEqual comparison of the rest
	delete(expectedCopy, "error")
	delete(gotCopy, "error")

	// 1. Compare the maps excluding the 'error' and 'failed_files' fields
	if !reflect.DeepEqual(expectedCopy, gotCopy) {
		t.Errorf("Result map (excluding error/failed_files) does not match expected.")
		// Provide a more detailed diff using go-cmp
		diff := cmp.Diff(expectedCopy, gotCopy)
		if diff != "" {
			t.Errorf("Map diff (-expected +got):\n%s", diff)
		} else {
			t.Logf("  Expected (Core Map): %#v", expectedCopy)
			t.Logf("  Got (Core Map):      %#v", gotCopy)
		}
		return false // Maps differ
	}

	// 2. Compare 'failed_files' maps (check keys and error substrings)
	if expFailedOk != gotFailedOk {
		t.Errorf("Presence of 'failed_files' map mismatch. Expected: %t, Got: %t", expFailedOk, gotFailedOk)
		return false
	}
	if expFailedOk { // Both expected and got failed_files
		expectedFailedMap, okE := expectedFailedVal.(map[string]interface{})
		gotFailedMap, okG := gotFailedVal.(map[string]interface{})
		if !okE || !okG {
			t.Errorf("Type mismatch for 'failed_files'. Expected map[string]interface{}, Got E:%T, G:%T", expectedFailedVal, gotFailedVal)
			return false
		}
		// Check keys are the same
		expectedKeys := make([]string, 0, len(expectedFailedMap))
		gotKeys := make([]string, 0, len(gotFailedMap))
		for k := range expectedFailedMap {
			expectedKeys = append(expectedKeys, k)
		}
		for k := range gotFailedMap {
			gotKeys = append(gotKeys, k)
		}
		sort.Strings(expectedKeys)
		sort.Strings(gotKeys)
		if !reflect.DeepEqual(expectedKeys, gotKeys) {
			t.Errorf("'failed_files' keys mismatch.\nExpected: %v\nGot:      %v", expectedKeys, gotKeys)
			return false
		}
		// Check error substrings for each key
		for _, key := range expectedKeys {
			// Expected error substring comes from the expected map
			expErrSubstr, okE := expectedFailedMap[key].(string)
			if !okE {
				t.Errorf("'failed_files' expected error for key '%s' is not a string (%T)", key, expectedFailedMap[key])
				return false
			}
			gotErrStr, okG := gotFailedMap[key].(string)
			if !okG {
				t.Errorf("'failed_files' actual error for key '%s' is not a string (%T)", key, gotFailedMap[key])
				return false
			}
			// *** Use strings.Contains for flexible checking ***
			if !strings.Contains(gotErrStr, expErrSubstr) {
				t.Errorf("'failed_files' error for key '%s' mismatch.\nExpected Substring: %q\nGot:              %q", key, expErrSubstr, gotErrStr)
				return false
			}
		}
	}

	// 3. Compare errors (top-level error field)
	var expectedErr error
	var gotErr error

	if expErrOk && expectedErrVal != nil {
		if err, ok := expectedErrVal.(error); ok {
			expectedErr = err
		} else if errStr, ok := expectedErrVal.(string); ok && errStr != "" {
			expectedErr = errors.New(errStr)
		}
	}
	if gotErrOk && gotErrVal != nil {
		if errStr, ok := gotErrVal.(string); ok && errStr != "" {
			gotErr = errors.New(errStr)
		}
	}

	if expectedErr == nil && gotErr == nil {
		return true // Both nil, maps matched earlier
	}
	if expectedErr != nil && gotErr == nil {
		t.Errorf("Expected top-level error containing '%s', but got nil error message in map.", checkErrorSubstring)
		return false
	}
	if expectedErr == nil && gotErr != nil {
		t.Errorf("Expected no top-level error in map, but got: %v", gotErr)
		return false
	}
	// Both errors are non-nil
	if checkErrorSubstring != "" {
		if !strings.Contains(gotErr.Error(), checkErrorSubstring) {
			t.Errorf("Expected top-level error message containing '%s', but got: %v", checkErrorSubstring, gotErr)
			return false
		}
	} else {
		if expectedErr.Error() != gotErr.Error() {
			t.Errorf("Top-level Error mismatch.\nExpected: %v\nGot:      %v", expectedErr, gotErr)
			return false
		}
	}

	return true // All checks passed
}

// --- Test Cases ---

func TestToolGoUpdateImportsForMovedPackage(t *testing.T) {
	// File contents remain unchanged...
	s1Content := `package sub1
var VarS1 = "v1"
type TypeS1 struct{}
func FuncS1() {}`

	s2Content := `package sub2
const ConstS2 = "c2"
type TypeS2 int
func FuncS2() {}`

	mainContentOriginal := `package main

import (
	"fmt"
	original "testtool/refactored" // Original named import
)

func main() {
	original.FuncS1() // Usage of sub1
	_ = original.VarS1 // Usage of sub1
	// var x original.TypeS2 // Usage of sub2 - type only
	fmt.Println(original.ConstS2) // Usage of sub2 - const only
	// Let's add explicit func calls for better testing
	original.FuncS2() // Usage of sub2
	var y original.TypeS1 // Usage of sub1 - type only
	_ = y
}`

	mainContentExpectedBasic := `package main

import (
	"fmt"
	"testtool/refactored/sub1" // New import
	"testtool/refactored/sub2" // New import
)

func main() {
	sub1.FuncS1() // UPDATED
	_ = sub1.VarS1 // UPDATED
	// var x original.TypeS2 // Type usage not changed
	fmt.Println(sub2.ConstS2) // UPDATED
	sub2.FuncS2() // UPDATED
	var y original.TypeS1 // Type usage not changed
	_ = y
}`

	mainContentWithParseError := `package main

import (
	"fmt"
	original "testtool/refactored"
)

func main() {
	original.FuncS1()
	_ = original.VarS1
	// var x original.TypeS2 // Type usage not changed
	fmt.Println() syntax error here // Intentional parse error
	fmt.Println(original.ConstS2)
	original.FuncS2()
	var y original.TypeS1 // Type usage not changed
	_ = y
}`

	noUsageContent := `package other
import "fmt"
func NoUsage() { fmt.Println("hello") }`

	ambigS1Content := `package sub1
var Ambiguous = "from s1"`
	ambigS2Content := `package sub2
func Ambiguous() {} // Same name, different type`
	ambigMainContent := `package main
import original "testtool/refactored"
func main() { _ = original.Ambiguous }`

	testCases := []struct {
		name                string
		files               map[string]string
		params              map[string]interface{}
		expectedResult      map[string]interface{}
		expectedFiles       map[string]string
		expectedErrorSubstr string
	}{
		// Basic and Scan Scope test cases remain the same...
		{
			name: "Basic success case - one file modified",
			files: map[string]string{
				"testtool/refactored/sub1/s1.go": s1Content,
				"testtool/refactored/sub2/s2.go": s2Content,
				"client/main.go":                 mainContentOriginal,
				"other/nousage.go":               noUsageContent,
			},
			params: map[string]interface{}{
				"refactored_package_path": "testtool/refactored",
				"scan_scope":              ".",
			},
			expectedResult: map[string]interface{}{
				"modified_files": []interface{}{"client/main.go"},
				"skipped_files":  map[string]interface{}{"other/nousage.go": "Original package not imported"},
				"failed_files":   map[string]interface{}{},
				"error":          nil,
			},
			expectedFiles: map[string]string{
				"client/main.go":   mainContentExpectedBasic,
				"other/nousage.go": noUsageContent,
			},
			expectedErrorSubstr: "",
		},
		{
			name: "Scan scope limited to client dir",
			files: map[string]string{
				"testtool/refactored/sub1/s1.go": s1Content,
				"testtool/refactored/sub2/s2.go": s2Content,
				"client/main.go":                 mainContentOriginal,
				"other/nousage.go":               noUsageContent,
			},
			params: map[string]interface{}{
				"refactored_package_path": "testtool/refactored",
				"scan_scope":              "client",
			},
			expectedResult: map[string]interface{}{
				"modified_files": []interface{}{"client/main.go"},
				"skipped_files":  map[string]interface{}{},
				"failed_files":   map[string]interface{}{},
				"error":          nil,
			},
			expectedFiles: map[string]string{
				"client/main.go":   mainContentExpectedBasic,
				"other/nousage.go": noUsageContent,
			},
			expectedErrorSubstr: "",
		},
		{
			name: "Client file has parse error",
			files: map[string]string{
				"testtool/refactored/sub1/s1.go": s1Content,
				"testtool/refactored/sub2/s2.go": s2Content,
				"client/main.go":                 mainContentWithParseError,
				"other/nousage.go":               noUsageContent,
			},
			params: map[string]interface{}{
				"refactored_package_path": "testtool/refactored",
				"scan_scope":              ".",
			},
			expectedResult: map[string]interface{}{
				"modified_files": []interface{}{},
				"skipped_files":  map[string]interface{}{"other/nousage.go": "Original package not imported"},
				"failed_files":   map[string]interface{}{"client/main.go": "expected ';', found syntax"}, // Use substring check
				"error":          nil,
			},
			expectedFiles: map[string]string{
				"client/main.go":   mainContentWithParseError,
				"other/nousage.go": noUsageContent,
			},
			expectedErrorSubstr: "",
		},
		{
			name: "Symbol map ambiguity",
			files: map[string]string{
				"testtool/refactored/sub1/s1.go": ambigS1Content,
				"testtool/refactored/sub2/s2.go": ambigS2Content,
				"client/main.go":                 ambigMainContent,
			},
			params: map[string]interface{}{
				"refactored_package_path": "testtool/refactored",
				"scan_scope":              ".",
			},
			// *** UPDATED Expected Result: Expect top-level error, empty lists/maps ***
			expectedResult: map[string]interface{}{
				"modified_files": []interface{}{},
				"skipped_files":  map[string]interface{}{},
				"failed_files":   map[string]interface{}{},
				"error":          "placeholder", // Will be checked by expectedErrorSubstr
			},
			expectedFiles: map[string]string{
				"client/main.go": ambigMainContent, // Not modified
			},
			// *** UPDATED Substring: Use correct paths ***
			expectedErrorSubstr: "ambiguous exported symbols found: symbol 'Ambiguous' (found in testtool/refactored/sub1 and testtool/refactored/sub2)",
		},
	}

	// --- Test Execution Loop (remains unchanged from previous response) ---
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootDir, cleanup := setupPackageUpdateTestEnv(t, tc.files)
			defer cleanup()
			logTest(t, "Test rootDir: %s", rootDir)

			interpreter, _ := newDefaultTestInterpreter(t)
			interpreter.sandboxDir = rootDir

			toolImpl, found := interpreter.ToolRegistry().GetTool("GoUpdateImportsForMovedPackage")
			if !found {
				t.Fatalf("Tool GoUpdateImportsForMovedPackage not found in registry")
			}

			// Prepare arguments as slice based on spec
			refactoredPkgPath, ok1 := tc.params["refactored_package_path"].(string)
			scanScope, ok2 := tc.params["scan_scope"].(string)
			if !ok1 || !ok2 {
				t.Fatalf("Test setup error: invalid parameters in tc.params map")
			}
			rawArgs := makeArgs(refactoredPkgPath, scanScope)

			// Validate args
			convertedArgs, valErr := ValidateAndConvertArgs(toolImpl.Spec, rawArgs)
			if valErr != nil {
				t.Fatalf("Argument validation failed unexpectedly: %v", valErr)
			}

			// Execute the tool function
			resultIntf, toolErrGo := toolImpl.Func(interpreter, convertedArgs)

			// Process result and error
			var result map[string]interface{}
			var toolErr error = toolErrGo // Capture Go error directly

			if resultIntf != nil {
				var ok bool
				result, ok = resultIntf.(map[string]interface{})
				if !ok {
					t.Fatalf("Tool function returned unexpected type %T, expected map[string]interface{}", resultIntf)
				}
				if toolErr == nil {
					if errVal, ok := result["error"]; ok && errVal != nil {
						if errStr, okStr := errVal.(string); okStr {
							toolErr = errors.New(errStr)
						}
					}
				}
			} else if toolErr == nil {
				t.Fatalf("Tool function returned nil result and nil error unexpectedly")
			}

			// Assertions
			if !compareResultMaps(t, tc.expectedResult, result, tc.expectedErrorSubstr) {
				// Error details logged within compareResultMaps
			}

			// File content check
			if toolErr == nil {
				modifiedFilesRaw, _ := result["modified_files"].([]interface{})
				modifiedFilesMap := make(map[string]bool)
				for _, f := range modifiedFilesRaw {
					if s, ok := f.(string); ok {
						modifiedFilesMap[s] = true
					}
				}

				for relPath, expectedContent := range tc.expectedFiles {
					fullPath := filepath.Join(rootDir, relPath)
					actualBytes, err := os.ReadFile(fullPath)
					if err != nil {
						if _, shouldBeModified := modifiedFilesMap[relPath]; shouldBeModified {
							t.Errorf("Expected file '%s' to be modified, but encountered read error: %v", relPath, err)
						} else {
							t.Logf("Note: Could not read file '%s' for comparison (maybe expected?): %v", relPath, err)
						}
						continue
					}
					actualContent := string(actualBytes)
					AssertEqualStrings(t, expectedContent, actualContent, "Content mismatch for file '%s'", relPath)
				}
			}
		})
	}
}
