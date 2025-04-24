// filename: core/tools_go_ast_symbol_map_test.go
package core

import (
	"errors"
	"log" // Needed for mock interpreter logger
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	// NOTE: No longer need go/ast, go/parser, go/token for manual parsing
)

// --- Test Setup Helpers ---

// writeFileHelper writes content to a file, failing the test on error.
func writeFileHelper(t *testing.T, path string, content string) {
	t.Helper()
	// Use MkdirAll to ensure parent directories exist
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		t.Fatalf("Failed to create directory %s: %v", dir, err)
	}

	t.Logf("Writing %d bytes to: %s", len(content), path)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
	// Optional: Read back check can be added here if needed for debugging
}

// setupBasicFixture creates a simple package structure with unique exported symbols.
func setupBasicFixture(t *testing.T, rootDir string) {
	t.Helper()
	// Note: refactoredPkgPath is like "testbuildmap/original"
	// We need to create files within rootDir/<refactoredPkgPath>/sub1 etc.
	// buildSymbolMap expects the structure inside rootDir to match the refactoredPkgPath
	baseDir := filepath.Join(rootDir, "testbuildmap", "original") // Directory structure inside rootDir
	sub1Dir := filepath.Join(baseDir, "sub1")
	sub2Dir := filepath.Join(baseDir, "sub2")

	// Content for sub1
	sub1Content := `package sub1

import "fmt"

var ExportedVarOne = "hello"

func ExportedFuncOne() {
	fmt.Println(ExportedVarOne)
}

type internalType struct{} // Unexported

func internalFunc() {} // Unexported
`
	writeFileHelper(t, filepath.Join(sub1Dir, "file1.go"), sub1Content)

	// Content for sub2
	sub2Content := `package sub2

const ExportedConstTwo = 123

type ExportedTypeTwo struct {
	Field int
}

// Method, should be ignored by symbol map
func (e ExportedTypeTwo) Method() {}

var internalVar = 456 // Unexported
`
	writeFileHelper(t, filepath.Join(sub2Dir, "file2.go"), sub2Content)
	// Add an empty file to ensure it's handled correctly
	writeFileHelper(t, filepath.Join(sub2Dir, "empty.go"), "package sub2")
	// Add a _test.go file to ensure it's ignored
	writeFileHelper(t, filepath.Join(sub2Dir, "file2_test.go"), "package sub2\n\nimport \"testing\"\n\nfunc TestDummy(t *testing.T) {}")
}

// setupAmbiguousFixture creates a package structure where the same symbol is exported from multiple subpackages.
func setupAmbiguousFixture(t *testing.T, rootDir string) {
	t.Helper()
	baseDir := filepath.Join(rootDir, "testbuildmap", "original")
	sub1Dir := filepath.Join(baseDir, "sub1")
	sub2Dir := filepath.Join(baseDir, "sub2")

	// Content for sub1
	sub1Content := `package sub1

var AmbiguousVar = "from sub1"

func AmbiguousFunc() {}
`
	writeFileHelper(t, filepath.Join(sub1Dir, "ambig1.go"), sub1Content)

	// Content for sub2
	sub2Content := `package sub2

type AmbiguousVar struct{} // Same name as var in sub1, different type

const AmbiguousFunc = 1 // Same name as func in sub1, different type
`
	writeFileHelper(t, filepath.Join(sub2Dir, "ambig2.go"), sub2Content)
}

// setupNoExportedFixture creates a package structure with only unexported symbols.
func setupNoExportedFixture(t *testing.T, rootDir string) {
	t.Helper()
	baseDir := filepath.Join(rootDir, "testbuildmap", "original")
	sub1Dir := filepath.Join(baseDir, "sub1")
	sub2Dir := filepath.Join(baseDir, "sub2")

	// Content for sub1
	sub1Content := `package sub1

var localVarOne = "hello"

func localFuncOne() {}
`
	writeFileHelper(t, filepath.Join(sub1Dir, "local1.go"), sub1Content)

	// Content for sub2
	sub2Content := `package sub2

const localConstTwo = 123

type localTypeTwo struct{}
`
	writeFileHelper(t, filepath.Join(sub2Dir, "local2.go"), sub2Content)
}

// setupNoSubPackagesFixture creates a package directory but no subdirectories containing Go code.
func setupNoSubPackagesFixture(t *testing.T, rootDir string) {
	t.Helper()
	baseDir := filepath.Join(rootDir, "testbuildmap", "original")
	// Create the base directory, but no subdirectories with Go files
	err := os.MkdirAll(baseDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir %s: %v", baseDir, err)
	}
	// Optionally add an empty directory or a directory with non-Go files
	emptySubDir := filepath.Join(baseDir, "empty")
	err = os.Mkdir(emptySubDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create dir %s: %v", emptySubDir, err)
	}
	writeFileHelper(t, filepath.Join(emptySubDir, "readme.txt"), "This is not a go file")
}

// --- Test Function ---

func TestBuildSymbolMapLogic(t *testing.T) {
	testCases := []struct {
		name              string
		fixtureSetupFunc  func(t *testing.T, rootDir string)
		refactoredPkgPath string // The simulated original package path
		expectedMap       map[string]string
		expectedErrorIs   error // Use errors.Is for comparison
	}{
		{
			name:              "Basic case with multiple types",
			fixtureSetupFunc:  setupBasicFixture,
			refactoredPkgPath: "testbuildmap/original",
			expectedMap: map[string]string{
				"ExportedFuncOne":  "testbuildmap/original/sub1",
				"ExportedVarOne":   "testbuildmap/original/sub1",
				"ExportedTypeTwo":  "testbuildmap/original/sub2",
				"ExportedConstTwo": "testbuildmap/original/sub2",
			},
			expectedErrorIs: nil,
		},
		{
			name:              "Ambiguous symbols",
			fixtureSetupFunc:  setupAmbiguousFixture,
			refactoredPkgPath: "testbuildmap/original",
			expectedMap:       nil,
			expectedErrorIs:   ErrSymbolMappingFailed, // Expecting the wrapped error
		},
		{
			name:              "No exported symbols in subpackages",
			fixtureSetupFunc:  setupNoExportedFixture,
			refactoredPkgPath: "testbuildmap/original",
			expectedMap:       map[string]string{}, // Expect an empty map
			expectedErrorIs:   nil,
		},
		{
			name:              "No subpackages with go code",
			fixtureSetupFunc:  setupNoSubPackagesFixture,
			refactoredPkgPath: "testbuildmap/original",
			expectedMap:       map[string]string{}, // Expect an empty map
			expectedErrorIs:   nil,
		},
		{
			name:              "Original package path does not exist",
			fixtureSetupFunc:  setupBasicFixture, // Fixture doesn't matter, path check is first
			refactoredPkgPath: "testbuildmap/nonexistent",
			expectedMap:       nil,
			expectedErrorIs:   ErrRefactoredPathNotFound, // Expecting the specific error
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootDir := t.TempDir()
			t.Logf("Test rootDir: %s", rootDir)

			// Create the necessary file structure within the temp dir
			tc.fixtureSetupFunc(t, rootDir)

			// Create a minimal interpreter for the test
			// Use t.Logf for logging within the function being tested
			mockLogger := log.New(testWriter{t}, "", log.LstdFlags)
			interp := &Interpreter{
				logger:     mockLogger,
				sandboxDir: rootDir, // buildSymbolMap expects files relative to sandboxDir
				// Other fields can be zero/nil for this test
			}

			// Call the ACTUAL function being tested
			actualMap, err := buildSymbolMap(tc.refactoredPkgPath, interp)

			// --- Assertions ---
			if tc.expectedErrorIs != nil {
				if err == nil {
					t.Errorf("Expected error wrapping '%v', but got nil error", tc.expectedErrorIs)
				} else if !errors.Is(err, tc.expectedErrorIs) {
					// Check if the error message contains the expected error string for more flexibility
					// This helps if the error is wrapped multiple times or formatted differently.
					expectedErrStr := tc.expectedErrorIs.Error()
					if !strings.Contains(err.Error(), expectedErrStr) {
						t.Errorf("Expected error wrapping '%v' or containing its text, but got error: %v (type %T)", tc.expectedErrorIs, err, err)
					} else {
						t.Logf("Got expected error type/text: %v", err) // Log success for clarity
					}
				} else {
					t.Logf("Got expected error type via errors.Is: %v", err) // Log success for clarity
				}
				// If error was expected, map content doesn't matter as much, but check it's nil/empty
				if actualMap != nil && len(actualMap) > 0 {
					t.Errorf("Expected nil or empty map when error occurred, but got: %v", actualMap)
				}
			} else { // No error expected
				if err != nil {
					t.Errorf("Did not expect error, but got: %v", err)
				}
				// Check map equality only if no error was expected and none occurred
				if !reflect.DeepEqual(actualMap, tc.expectedMap) {
					t.Errorf("Returned map does not match expected.\nExpected: %v\nGot:      %v", tc.expectedMap, actualMap)
				}
			}
		})
	}
}

// testWriter is a helper to redirect log output to t.Logf
type testWriter struct {
	t *testing.T
}

// func (tw testWriter) Write(p []byte) (n int, err error) {
// 	tw.t.Logf("%s", p)
// 	return len(p), nil
// }

// Ensure buildSymbolMap and required error variables are defined in core package
var (
	_ = buildSymbolMap // Reference to ensure it exists
	_ = ErrSymbolMappingFailed
	_ = ErrRefactoredPathNotFound
)
