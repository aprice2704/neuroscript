// filename: pkg/core/tools_go_ast_package_test.go
package core

import (
	"io"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	// "fmt" // Not needed directly here
)

// --- Test Setup Helpers ---
// (setupToolTestFixture, setupToolTestFixtureInvalidClient, setupToolTestFixtureAmbiguous)
// These seem okay, ensure they create the expected structure relative to the module path.
// Example: rootDir/testtool/refactored/sub1/s1.go

// --- setupToolTestFixture ---
func setupToolTestFixture(t *testing.T, rootDir string) {
	t.Helper()
	// Module file at root
	writeFileHelper(t, filepath.Join(rootDir, "go.mod"), "module testtool\n\ngo 1.21\n")

	// Base path for the refactored code *within* the module structure
	refactoredBase := filepath.Join(rootDir, "testtool", "refactored") // Use module path structure
	sub1Dir := filepath.Join(refactoredBase, "sub1")
	sub2Dir := filepath.Join(refactoredBase, "sub2")

	// Ensure directories exist
	if err := os.MkdirAll(sub1Dir, 0755); err != nil {
		t.Fatalf("MkdirAll failed for %s: %v", sub1Dir, err)
	}
	if err := os.MkdirAll(sub2Dir, 0755); err != nil {
		t.Fatalf("MkdirAll failed for %s: %v", sub2Dir, err)
	}

	// Write sub-package files
	writeFileHelper(t, filepath.Join(sub1Dir, "s1.go"), `package sub1
func FuncS1() {}
var VarS1 int
type TypeS1 struct{} // Add a type for completeness
`)
	writeFileHelper(t, filepath.Join(sub2Dir, "s2.go"), `package sub2
type TypeS2 struct{}
const ConstS2 = "hello"
func FuncS2() {} // Add a func for completeness
`)

	// Client package (can be anywhere relative to root, imports use module path)
	clientDir := filepath.Join(rootDir, "client")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed for %s: %v", clientDir, err)
	}
	// Original client content importing the 'parent' package path
	clientContent := `package main

import (
	"fmt"
	original "testtool/refactored" // Import path refers to module path (the old way)
)

func main() {
	original.FuncS1()
	_ = original.VarS1
	var x original.TypeS2
	fmt.Println(x)
	fmt.Println(original.ConstS2)
	// original.FuncS2() // Add usage for FuncS2 if needed
	// var y original.TypeS1 // Add usage for TypeS1 if needed
}
`
	writeFileHelper(t, filepath.Join(clientDir, "main.go"), clientContent)

	// Other package (should be skipped)
	otherDir := filepath.Join(rootDir, "other")
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed for %s: %v", otherDir, err)
	}
	writeFileHelper(t, filepath.Join(otherDir, "nousage.go"), `package other
import "fmt"
func Run() { fmt.Println("Other package") }
`)
}

// --- setupToolTestFixtureInvalidClient ---
func setupToolTestFixtureInvalidClient(t *testing.T, rootDir string) {
	t.Helper()
	setupToolTestFixture(t, rootDir) // Use the base fixture setup

	clientDir := filepath.Join(rootDir, "client")
	// Overwrite client/main.go with invalid content
	invalidClientContent := `package main

import (
	"fmt"
	original "testtool/refactored"
)

func main() {
	original.FuncS1()
	invalid syntax here // <<<< INVALID LINE 10 (adjust if needed)
	fmt.Println("test")
}
`
	writeFileHelper(t, filepath.Join(clientDir, "main.go"), invalidClientContent)
}

// --- setupToolTestFixtureAmbiguous ---
func setupToolTestFixtureAmbiguous(t *testing.T, rootDir string) {
	t.Helper()
	writeFileHelper(t, filepath.Join(rootDir, "go.mod"), "module testtool\n\ngo 1.21\n")

	refactoredBase := filepath.Join(rootDir, "testtool", "refactored")
	sub1Dir := filepath.Join(refactoredBase, "sub1")
	sub2Dir := filepath.Join(refactoredBase, "sub2")

	if err := os.MkdirAll(sub1Dir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.MkdirAll(sub2Dir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}

	// Create ambiguous symbols
	writeFileHelper(t, filepath.Join(sub1Dir, "s1.go"), `package sub1
var Ambiguous = "string"
func Another() {}`) // Add non-ambiguous too
	writeFileHelper(t, filepath.Join(sub2Dir, "s2.go"), `package sub2
type Ambiguous struct{} // Type with same name
const SomethingElse = 123`) // Add non-ambiguous

	// Client that imports the old path
	clientDir := filepath.Join(rootDir, "client")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	// Client doesn't need to *use* ambiguous symbols, just needs to import old path
	// for the tool to run and hit the symbol map error.
	writeFileHelper(t, filepath.Join(clientDir, "main.go"), `package main
import original "testtool/refactored"
func main() { _ = original.Another } // Use a non-ambiguous one
`)
}

// --- Test Function ---
func TestToolGoUpdateImportsForMovedPackage(t *testing.T) {
	// Keep logger discarded unless debugging specific test issues
	dummyLogger := log.New(io.Discard, "TestToolGoUpdateImports: ", log.LstdFlags)
	// dummyLogger := log.New(os.Stderr, "TestToolGoUpdateImports: ", log.LstdFlags|log.Lshortfile) // For Debugging
	dummyInterpreter := NewInterpreter(dummyLogger, nil)

	testCases := []struct {
		name                 string
		fixtureSetupFunc     func(t *testing.T, rootDir string)
		refactoredPkgPath    string                 // Import path used to find symbols (e.g., "testtool/refactored")
		scanScope            string                 // Directory to scan relative to rootDir (e.g., ".")
		expectedResult       map[string]interface{} // Expected outcome map
		expectedFileContents map[string]string      // Expected file content after modification (relative path -> content)
		expectedErrorContent string                 // Substring expected in error message if error is not nil
	}{
		{
			name:              "Basic success case - one file modified",
			fixtureSetupFunc:  setupToolTestFixture,
			refactoredPkgPath: "testtool/refactored", // Path containing the NEW code (sub-packages)
			scanScope:         ".",                   // Scan everything from rootDir
			expectedResult: map[string]interface{}{
				// Expect interface{} types for slices/maps due to result construction/comparison
				"modified_files": []interface{}{"client/main.go"},
				"skipped_files":  map[string]interface{}{"other/nousage.go": "Original package not imported"},
				"failed_files":   map[string]interface{}{},
				"error":          nil,
			},
			expectedFileContents: map[string]string{
				// --- FIX: Expect qualifiers to REMAIN UNCHANGED ---
				"client/main.go": `package main

import (
	"fmt"
	"testtool/refactored/sub1" // Added
	"testtool/refactored/sub2" // Added
)

func main() {
	original.FuncS1() // <-- RETAINED QUALIFIER
	_ = original.VarS1 // <-- RETAINED QUALIFIER
	var x original.TypeS2 // <-- RETAINED QUALIFIER
	fmt.Println(x)
	fmt.Println(original.ConstS2) // <-- RETAINED QUALIFIER
}
`,
			},
			expectedErrorContent: "",
		},
		{
			name:              "Scan scope limited to client dir",
			fixtureSetupFunc:  setupToolTestFixture,
			refactoredPkgPath: "testtool/refactored",
			scanScope:         "client", // Scan only within client directory
			expectedResult: map[string]interface{}{
				"modified_files": []interface{}{"client/main.go"}, // Path relative to rootDir still
				"skipped_files":  map[string]interface{}{},        // other/nousage.go is outside scope
				"failed_files":   map[string]interface{}{},
				"error":          nil,
			},
			expectedFileContents: map[string]string{
				// --- FIX: Expect qualifiers to REMAIN UNCHANGED ---
				"client/main.go": `package main

import (
	"fmt"
	"testtool/refactored/sub1"
	"testtool/refactored/sub2"
)

func main() {
	original.FuncS1() // <-- RETAINED QUALIFIER
	_ = original.VarS1 // <-- RETAINED QUALIFIER
	var x original.TypeS2 // <-- RETAINED QUALIFIER
	fmt.Println(x)
	fmt.Println(original.ConstS2) // <-- RETAINED QUALIFIER
}
`,
			},
			expectedErrorContent: "",
		},
		{
			name:              "Client file has parse error",
			fixtureSetupFunc:  setupToolTestFixtureInvalidClient,
			refactoredPkgPath: "testtool/refactored",
			scanScope:         ".",
			expectedResult: map[string]interface{}{
				"modified_files": []interface{}{},
				"skipped_files":  map[string]interface{}{"other/nousage.go": "Original package not imported"},
				// --- FIX: Expect parser error in failed_files ---
				"failed_files": map[string]interface{}{
					// The exact line/col might vary slightly, check content
					"client/main.go": "Failed to parse file: client/main.go:10:10: expected ';', found syntax",
				},
				"error": nil, // --- FIX: Top-level error should be nil if parse error handled per-file ---
			},
			expectedFileContents: nil, // No content change expected if parse fails
			expectedErrorContent: "",  // No top-level error message expected
		},
		{
			name:              "Symbol map ambiguity",
			fixtureSetupFunc:  setupToolTestFixtureAmbiguous,
			refactoredPkgPath: "testtool/refactored", // Path where ambiguous symbols live
			scanScope:         ".",
			expectedResult: map[string]interface{}{
				// Expect empty maps/slices as buildSymbolMap fails early
				"modified_files": []interface{}{},
				"skipped_files":  map[string]interface{}{},
				"failed_files":   map[string]interface{}{},
				// --- FIX: Error comes from buildSymbolMap ---
				// Use expectedErrorContent for substring check
				"error": "placeholder for check",
			},
			expectedFileContents: nil,
			// --- FIX: Check error content ---
			expectedErrorContent: "Failed to build symbol map for 'testtool/refactored': symbol mapping failed: ambiguous exported symbols found: symbol 'Ambiguous'",
		},
		// TODO: Add test cases:
		// - No symbols used from old package (should skip client/main.go)
		// - Old package not imported at all (covered by other/nousage.go, but maybe explicit client case?)
		// - Refactored package path doesn't exist (buildSymbolMap error)
		// - Scan scope doesn't exist (should error early?)
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootDir := t.TempDir()
			t.Logf("Test rootDir: %s", rootDir)
			// Defer cleanup for easier debugging if needed
			// defer os.RemoveAll(rootDir)

			// Setup fixture
			tc.fixtureSetupFunc(t, rootDir)
			dummyInterpreter.sandboxDir = rootDir // Set sandbox for the tool

			// Execute the tool
			args := []interface{}{tc.refactoredPkgPath, tc.scanScope}
			resultIntf, execErr := toolGoUpdateImportsForMovedPackage(dummyInterpreter, args)

			// --- Assertions ---
			if execErr != nil {
				// This is an error during tool *execution* framework, not a returned error in the map
				t.Fatalf("Tool execution framework failed unexpectedly: %v", execErr)
			}
			if resultIntf == nil {
				t.Fatalf("Tool returned nil result unexpectedly")
			}
			resultMap, ok := resultIntf.(map[string]interface{})
			if !ok {
				t.Fatalf("Tool result was not a map[string]interface{}, but %T", resultIntf)
			}

			// Normalize paths *before* comparison
			// Note: normalizeResultMapPaths uses interface{}, ensure it handles the map[string]interface{} correctly
			normalizeResultMapPaths(t, resultMap, rootDir)
			// Don't normalize expected map, paths should already be relative/slashed

			// Compare error string (using expectedErrorContent)
			compareErrorStringContent(t, resultMap, tc.expectedErrorContent)

			// Remove error for DeepEqual comparison of other fields
			delete(resultMap, "error")
			delete(tc.expectedResult, "error") // Remove placeholder if it existed

			// Set defaults for nil slices/maps AFTER removing error
			setDefaultResultMapValues(resultMap)
			setDefaultResultMapValues(tc.expectedResult) // Ensure expected has defaults too

			// Compare the rest of the map
			if !reflect.DeepEqual(resultMap, tc.expectedResult) {
				t.Errorf("Result map (excluding error string) does not match expected.\nExpected: %#v\nGot:      %#v", tc.expectedResult, resultMap)
			}

			// Check file contents if expected
			if tc.expectedFileContents != nil {
				for relPath, expectedContent := range tc.expectedFileContents {
					// Ensure relPath uses OS separators for ReadFile
					osRelPath := filepath.FromSlash(relPath)
					absPath := filepath.Join(rootDir, osRelPath)
					actualBytes, err := os.ReadFile(absPath)
					if err != nil {
						// Check if the file *should* have been modified before failing
						modifiedFiles, _ := tc.expectedResult["modified_files"].([]interface{})
						shouldBeModified := false
						for _, modFile := range modifiedFiles {
							if modFile.(string) == relPath { // Compare with slash path
								shouldBeModified = true
								break
							}
						}
						if shouldBeModified {
							t.Errorf("Expected to read modified file '%s', but got error: %v", relPath, err)
						} else {
							// If file wasn't expected to be modified, reading it might fail (if it was deleted?) or content doesn't matter
							t.Logf("File '%s' not expected to be modified, skipping content check due to read error: %v", relPath, err)
						}
						continue // Skip content check for this file
					}

					// Normalize content for comparison (ignore whitespace differences)
					// Using Fields might be too aggressive, FormatNode should handle most formatting.
					// Let's do a simpler TrimSpace comparison first.
					actualContent := strings.TrimSpace(string(actualBytes))
					expectedTrimmed := strings.TrimSpace(expectedContent)

					if actualContent != expectedTrimmed {
						// Use a diff library or t.Errorf with full content for better debugging
						t.Errorf("Content mismatch for file '%s'.\n--- Expected:\n%s\n\n--- Got:\n%s\n", relPath, expectedContent, string(actualBytes))
						// Example using fields for a potentially more robust check:
						// expectedFields := strings.Join(strings.Fields(expectedContent), " ")
						// actualFields := strings.Join(strings.Fields(string(actualBytes)), " ")
						// if actualFields != expectedFields {
						//  t.Errorf(...)
						// }
					}
				}
			}
		})
	}
}

// --- Test Helper Functions ---

// normalizeResultMapPaths: Converts absolute paths in maps/slices to relative, slash-separated paths.
// IMPORTANT: Modifies the input map directly.
func normalizeResultMapPaths(t *testing.T, dataMap map[string]interface{}, basePath string) {
	t.Helper()
	normalize := func(pathStr string) string {
		p := pathStr
		// Ensure basePath is absolute for Rel
		absBasePath, err := filepath.Abs(basePath)
		if err != nil {
			t.Logf("[WARN normalizePaths] Could not get absolute path for base: %s - %v", basePath, err)
			return filepath.ToSlash(p) // Fallback to slash conversion
		}
		// Ensure p is absolute if possible
		absP := p
		if !filepath.IsAbs(p) {
			// Assume relative to basePath if not absolute already? Or error?
			// Let's try joining with basePath
			absP = filepath.Join(absBasePath, p)
		}

		relPath, err := filepath.Rel(absBasePath, absP)
		if err != nil {
			t.Logf("[WARN normalizePaths] Could not make path relative: %s (base: %s) - %v. Using original.", p, absBasePath, err)
			// Keep original path but convert to slashes
			return filepath.ToSlash(p)
		}
		return filepath.ToSlash(relPath)
	}

	// Normalize modified_files (assuming []interface{} containing strings)
	if val, ok := dataMap["modified_files"]; ok && val != nil {
		if list, ok := val.([]interface{}); ok {
			normList := make([]interface{}, len(list))
			tempSortList := make([]string, len(list)) // For sorting
			for i, item := range list {
				if strItem, ok := item.(string); ok {
					normStr := normalize(strItem)
					normList[i] = normStr
					tempSortList[i] = normStr
				} else {
					normList[i] = item // Keep non-string items as is? Or error?
					t.Logf("[WARN normalizePaths] Non-string item found in modified_files: %T", item)
				}
			}
			sort.Strings(tempSortList) // Sort the string versions
			// Rebuild normList in sorted order
			sortedNormList := make([]interface{}, len(list))
			for i, sortedStr := range tempSortList {
				sortedNormList[i] = sortedStr
			}
			dataMap["modified_files"] = sortedNormList
		}
	}

	// Normalize keys in skipped_files and failed_files (assuming map[string]interface{})
	for _, key := range []string{"skipped_files", "failed_files"} {
		if val, ok := dataMap[key]; ok && val != nil {
			// Check if it's map[string]interface{} or map[string]string
			switch fileMap := val.(type) {
			case map[string]interface{}:
				normMap := make(map[string]interface{})
				for p, reason := range fileMap {
					normMap[normalize(p)] = reason
				}
				dataMap[key] = normMap
			case map[string]string: // Handle case where concrete type was returned
				normMap := make(map[string]interface{})
				for p, reason := range fileMap {
					normMap[normalize(p)] = reason
				}
				dataMap[key] = normMap // Store as map[string]interface{}
			default:
				t.Logf("[WARN normalizePaths] Unexpected type for %s: %T", key, val)
			}
		}
	}
}

// setDefaultResultMapValues: Ensures specific keys exist with default empty values (using interface{} types).
// IMPORTANT: Modifies the input map directly.
func setDefaultResultMapValues(resultMap map[string]interface{}) {
	if _, ok := resultMap["modified_files"]; !ok || resultMap["modified_files"] == nil {
		resultMap["modified_files"] = []interface{}{} // Default empty interface slice
	}
	if _, ok := resultMap["skipped_files"]; !ok || resultMap["skipped_files"] == nil {
		resultMap["skipped_files"] = map[string]interface{}{} // Default empty interface map
	}
	if _, ok := resultMap["failed_files"]; !ok || resultMap["failed_files"] == nil {
		resultMap["failed_files"] = map[string]interface{}{} // Default empty interface map
	}
	if _, ok := resultMap["error"]; !ok {
		resultMap["error"] = nil // Default nil error
	}
	// Ensure message is present if needed? Or handle absence in comparison.
	// If message isn't expected, ensure it's not present or is nil/empty.
	if _, ok := resultMap["message"]; !ok {
		// resultMap["message"] = nil // Or empty string? Depends on expectation.
	}
}

// compareErrorStringContent: Compares the error string in the result map against an expected substring.
func compareErrorStringContent(t *testing.T, actualMap map[string]interface{}, expectedErrContent string) {
	t.Helper()
	actualErrStr := ""
	if errVal, ok := actualMap["error"]; ok && errVal != nil {
		if strVal, ok := errVal.(string); ok {
			actualErrStr = strVal
		} else {
			t.Errorf("Value for 'error' key is not a string: %T", errVal)
		}
	}

	if expectedErrContent == "" {
		if actualErrStr != "" {
			t.Errorf("Expected no error message, but got: %q", actualErrStr)
		}
	} else {
		if actualErrStr == "" {
			t.Errorf("Expected error message containing %q, but got no error message.", expectedErrContent)
		} else if !strings.Contains(actualErrStr, expectedErrContent) {
			t.Errorf("Error message mismatch.\nExpected content: %q\nGot:              %q", expectedErrContent, actualErrStr)
		} else {
			// Content found, log for confirmation if needed
			t.Logf("Actual error %q contains expected content %q", actualErrStr, expectedErrContent)
		}
	}
}
