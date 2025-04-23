// filename: pkg/core/tools_go_ast_package_test.go
package core

import (
	"io" // For io.Discard
	"log"
	"os"
	"path/filepath"
	"reflect"
	"sort" // For sorting slices before DeepEqual
	"strings"
	"testing"
	// Removed time import as NewRateLimiter is removed
)

// --- Test Setup Helpers ---

// setupToolTestFixture creates the core structure for tool tests.
// rootDir/
//
//	go.mod (module testtool)
//	refactored/
//	  sub1/
//	    s1.go (package sub1; func FuncS1(){}, var VarS1)
//	  sub2/
//	    s2.go (package sub2; type TypeS2 struct{}, const ConstS2)
//	client/
//	  main.go (package main; import original "testtool/refactored"; uses original.FuncS1, original.VarS1, original.TypeS2, original.ConstS2)
//	other/
//	  nousage.go (package other; import "fmt")
func setupToolTestFixture(t *testing.T, rootDir string) {
	t.Helper()
	// Module file
	writeFileHelper(t, filepath.Join(rootDir, "go.mod"), "module testtool\n\ngo 1.21\n")

	// Refactored package structure (where symbols now live)
	sub1Dir := filepath.Join(rootDir, "refactored", "sub1")
	sub2Dir := filepath.Join(rootDir, "refactored", "sub2")
	if err := os.MkdirAll(sub1Dir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.MkdirAll(sub2Dir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	writeFileHelper(t, filepath.Join(sub1Dir, "s1.go"), `package sub1

// FuncS1 docs
func FuncS1() {}
var VarS1 int // Add a var
`)
	writeFileHelper(t, filepath.Join(sub2Dir, "s2.go"), `package sub2

// TypeS2 docs
type TypeS2 struct{}
const ConstS2 = "hello" // Add a const
`)

	// Client package using the *original* import path
	clientDir := filepath.Join(rootDir, "client")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	clientContent := `package main

import (
	"fmt"
	// Pretend original path was "testtool/refactored"
	original "testtool/refactored"
)

func main() {
	original.FuncS1() // Use func from sub1
	_ = original.VarS1 // Use var from sub1
	var x original.TypeS2 // Use type from sub2
	fmt.Println(x)
	fmt.Println(original.ConstS2) // Use const from sub2
}
`
	writeFileHelper(t, filepath.Join(clientDir, "main.go"), clientContent)

	// Another package not using the refactored import
	otherDir := filepath.Join(rootDir, "other")
	if err := os.MkdirAll(otherDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	writeFileHelper(t, filepath.Join(otherDir, "nousage.go"), `package other

import "fmt"

func Run() { fmt.Println("Other package") }
`)
}

// setupToolTestFixtureInvalidClient creates a fixture with invalid Go code in the client.
func setupToolTestFixtureInvalidClient(t *testing.T, rootDir string) {
	t.Helper()
	// Setup base structure same as setupToolTestFixture
	setupToolTestFixture(t, rootDir)

	// Overwrite client/main.go with invalid syntax
	clientDir := filepath.Join(rootDir, "client")
	invalidClientContent := `package main

import "testtool/refactored"

func main() {
	original.FuncS1()
	invalid syntax here // <<<< INVALID
}
`
	// Use the local writeFileHelper
	writeFileHelper(t, filepath.Join(clientDir, "main.go"), invalidClientContent)
}

// setupToolTestFixtureAmbiguous creates a fixture where buildSymbolMap should fail.
func setupToolTestFixtureAmbiguous(t *testing.T, rootDir string) {
	t.Helper()
	writeFileHelper(t, filepath.Join(rootDir, "go.mod"), "module testtool\n\ngo 1.21\n")
	sub1Dir := filepath.Join(rootDir, "refactored", "sub1") // Changed suba to sub1
	sub2Dir := filepath.Join(rootDir, "refactored", "sub2") // Changed subb to sub2
	if err := os.MkdirAll(sub1Dir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	if err := os.MkdirAll(sub2Dir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	// Use the local writeFileHelper
	writeFileHelper(t, filepath.Join(sub1Dir, "s1.go"), "package sub1\nfunc Ambiguous() {}")
	writeFileHelper(t, filepath.Join(sub2Dir, "s2.go"), "package sub2\nfunc Ambiguous() {}") // Same symbol name

	// Add a client file (doesn't matter if it uses the symbol, buildSymbolMap fails first)
	clientDir := filepath.Join(rootDir, "client")
	if err := os.MkdirAll(clientDir, 0755); err != nil {
		t.Fatalf("MkdirAll failed: %v", err)
	}
	writeFileHelper(t, filepath.Join(clientDir, "main.go"), "package main\n\nimport \"testtool/refactored\"\n\nfunc main() {}")
}

// --- Test Function ---

func TestToolGoUpdateImportsForMovedPackage(t *testing.T) {
	// Use NewInterpreter constructor and remove invalid fields
	dummyLogger := log.New(io.Discard, "TestToolGoUpdateImports: ", 0) // Discard logs or use os.Stderr
	dummyInterpreter := NewInterpreter(dummyLogger, nil)               // Provide nil for LLMClient

	testCases := []struct {
		name                 string
		fixtureSetupFunc     func(t *testing.T, rootDir string)
		refactoredPkgPath    string                 // Import path of the *original* package
		scanScope            string                 // Usually "." relative to rootDir
		expectedResult       map[string]interface{} // Compare modified, skipped, failed, error
		expectedFileContents map[string]string      // map[relPath]expectedContent
	}{
		{
			name:              "Basic success case - one file modified",
			fixtureSetupFunc:  setupToolTestFixture,
			refactoredPkgPath: "testtool/refactored",
			scanScope:         ".",
			expectedResult: map[string]interface{}{
				"modified_files": []interface{}{"client/main.go"}, // Paths expected relative to rootDir
				"skipped_files":  map[string]interface{}{"other/nousage.go": "Original package not imported"},
				"failed_files":   map[string]interface{}{},
				"error":          nil,
			},
			expectedFileContents: map[string]string{
				"client/main.go": `package main

import (
	"fmt"
	"testtool/refactored/sub1" // Added
	"testtool/refactored/sub2" // Added
)

func main() {
	sub1.FuncS1() // Qualifier needs manual update! Tool only changes imports.
	_ = sub1.VarS1
	var x sub2.TypeS2 // Qualifier needs manual update!
	fmt.Println(x)
	fmt.Println(sub2.ConstS2)
}
`, // Note: Original import line with alias is removed correctly by astutil
			},
		},
		{
			name:              "Scan scope limited to client dir",
			fixtureSetupFunc:  setupToolTestFixture,
			refactoredPkgPath: "testtool/refactored",
			scanScope:         "client", // Scan only within client directory
			expectedResult: map[string]interface{}{
				"modified_files": []interface{}{"client/main.go"},
				"skipped_files":  map[string]interface{}{}, // other/nousage.go is outside scope
				"failed_files":   map[string]interface{}{},
				"error":          nil,
			},
			expectedFileContents: map[string]string{
				"client/main.go": `package main

import (
	"fmt"
	"testtool/refactored/sub1"
	"testtool/refactored/sub2"
)

func main() {
	sub1.FuncS1()
	_ = sub1.VarS1
	var x sub2.TypeS2
	fmt.Println(x)
	fmt.Println(sub2.ConstS2)
}
`,
			},
		},
		{
			name:              "Client file has parse error",
			fixtureSetupFunc:  setupToolTestFixtureInvalidClient,
			refactoredPkgPath: "testtool/refactored",
			scanScope:         ".",
			expectedResult: map[string]interface{}{
				"modified_files": []interface{}{},
				"skipped_files":  map[string]interface{}{"other/nousage.go": "Original package not imported"},
				"failed_files":   map[string]interface{}{}, // packages.Load error is top-level
				// Error message might vary slightly depending on go version/env
				"error": "errors encountered loading packages: file=client/main.go: client/main.go:7:2: expected operand, found 'invalid'",
			},
			expectedFileContents: nil,
		},
		{
			name:              "Symbol map ambiguity",
			fixtureSetupFunc:  setupToolTestFixtureAmbiguous,
			refactoredPkgPath: "testtool/refactored",
			scanScope:         ".",
			expectedResult: map[string]interface{}{
				"modified_files": nil, // buildSymbolMap fails
				"skipped_files":  nil,
				"failed_files":   nil,
				// Adjusted ambiguous path based on corrected fixture helper
				"error": "Failed to build symbol map: ambiguous exported symbols found: symbol 'Ambiguous' (found in testtool/refactored/sub1 and testtool/refactored/sub2)",
			},
			expectedFileContents: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootDir := t.TempDir()
			t.Logf("Test rootDir: %s", rootDir)
			tc.fixtureSetupFunc(t, rootDir)
			dummyInterpreter.sandboxDir = rootDir // Set sandbox for the test case

			// --- Execute Tool ---
			args := []interface{}{tc.refactoredPkgPath, tc.scanScope}
			resultIntf, execErr := toolGoUpdateImportsForMovedPackage(dummyInterpreter, args)

			if execErr != nil {
				t.Fatalf("Tool execution failed unexpectedly: %v", execErr)
			}
			if resultIntf == nil {
				t.Fatalf("Tool returned nil result unexpectedly")
			}
			resultMap, ok := resultIntf.(map[string]interface{})
			if !ok {
				t.Fatalf("Tool result was not a map[string]interface{}, but %T", resultIntf)
			}

			// --- Compare Result Map ---
			// Normalize paths function (can be refined)
			normalizePaths := func(data interface{}, basePath string) interface{} {
				switch v := data.(type) {
				case []interface{}: // For modified_files list
					normList := make([]interface{}, len(v))
					strList := make([]string, 0, len(v))
					for _, item := range v {
						if strItem, ok := item.(string); ok {
							absPath := filepath.Join(basePath, strItem)
							relPath, err := filepath.Rel(basePath, absPath)
							if err != nil {
								t.Logf("Warning: could not make path relative: %s", strItem)
								relPath = strItem
							}
							strList = append(strList, filepath.ToSlash(relPath))
						} else {
							// Handle non-string items if necessary, or default/error
							normList = append(normList, item) // Keep non-strings for now
							return normList                   // Return mixed list if non-string found? Or filter?
						}
					}
					sort.Strings(strList) // Sort the normalized string paths
					for i, s := range strList {
						normList[i] = s
					} // Put sorted strings back
					return normList
				case map[string]interface{}: // For skipped_files, failed_files maps
					normMap := make(map[string]interface{})
					for key, val := range v {
						absKey := filepath.Join(basePath, key)
						relKey, err := filepath.Rel(basePath, absKey)
						if err != nil {
							t.Logf("Warning: could not make key path relative: %s", key)
							relKey = key
						}
						normMap[filepath.ToSlash(relKey)] = val
					}
					return normMap
				case string:
					return v // Keep error string as is
				case nil:
					return nil
				default:
					return v // Return other types unmodified
				}
			}

			expectedNormalized := make(map[string]interface{})
			actualNormalized := make(map[string]interface{})

			// Helper to safely get map value or default
			_ = func(m map[string]interface{}, key string, defaultVal interface{}) interface{} {
				if val, ok := m[key]; ok && val != nil {
					return val
				}
				return defaultVal
			}

			// Normalize expected and actual results before comparison
			for k, v := range tc.expectedResult {
				if k == "modified_files" || k == "skipped_files" || k == "failed_files" {
					expectedNormalized[k] = normalizePaths(v, rootDir)
				} else {
					expectedNormalized[k] = v
				}
			}
			for k, v := range resultMap {
				if k == "modified_files" || k == "skipped_files" || k == "failed_files" {
					actualNormalized[k] = normalizePaths(v, rootDir)
				} else {
					actualNormalized[k] = v
				}
			}

			// Ensure all keys exist in both maps for comparison, defaulting appropriately
			keys := []string{"modified_files", "skipped_files", "failed_files", "error"}
			for _, key := range keys {
				var defaultVal interface{}
				switch key {
				case "modified_files":
					defaultVal = []interface{}{}
				case "skipped_files", "failed_files":
					defaultVal = map[string]interface{}{}
				case "error":
					defaultVal = nil
				}
				if _, ok := expectedNormalized[key]; !ok {
					expectedNormalized[key] = defaultVal
				}
				if _, ok := actualNormalized[key]; !ok {
					actualNormalized[key] = defaultVal
				}
			}

			// Compare error string separately (allowing actual to contain expected)
			expectedErrStr, _ := expectedNormalized["error"].(string)
			actualErrStr, _ := actualNormalized["error"].(string)
			errorMatch := false
			if expectedErrStr == "" && actualErrStr == "" {
				errorMatch = true
			} else if expectedErrStr != "" && actualErrStr != "" && strings.Contains(actualErrStr, expectedErrStr) {
				errorMatch = true
			} else if expectedErrStr == "" && actualErrStr == "" { // Explicit nil check
				errorMatch = true
			}

			if !errorMatch {
				t.Errorf("Error message mismatch.\nExpected (or contains): %q\nGot: %q", expectedErrStr, actualErrStr)
			}
			// Remove errors before DeepEqual comparison
			delete(expectedNormalized, "error")
			delete(actualNormalized, "error")

			// Compare the rest of the maps
			if !reflect.DeepEqual(actualNormalized, expectedNormalized) {
				t.Errorf("Result map (excluding error string) does not match expected.\nExpected: %#v\nGot:      %#v", expectedNormalized, actualNormalized)
			}

			// --- Compare File Contents ---
			if tc.expectedFileContents != nil {
				for relPath, expectedContent := range tc.expectedFileContents {
					absPath := filepath.Join(rootDir, relPath)
					actualBytes, err := os.ReadFile(absPath)
					if err != nil {
						t.Errorf("Failed to read expected modified file '%s': %v", relPath, err)
						continue
					}
					actualContent := string(actualBytes)
					expected := strings.TrimSpace(expectedContent)
					actual := strings.TrimSpace(actualContent)
					if actual != expected {
						t.Errorf("Content mismatch for file '%s'.\nExpected:\n---\n%s\n---\nGot:\n---\n%s\n---", relPath, expected, actual)
					}
				}
			}
		})
	}
}
