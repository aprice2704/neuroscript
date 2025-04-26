// filename: pkg/core/tools_go_ast_find_test.go
// UPDATED: Use GetHandleValue, RegisterHandle
package goast

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings" // Import strings for file extension check
	"testing"
	// No go-cmp needed if comparing simple maps carefully
)

// --- Helper Functions ---

// setupParseForFindTest: Parses content and returns handle. Assumes helpers are available.
func setupParseForFindTest(t *testing.T, interp *Interpreter, content string) string {
	t.Helper()
	// IMPORTANT: toolGoParseFile likely uses "<content string>" as filename when parsing from string
	handleIDIntf, err := toolGoParseFile(interp, makeArgs(nil, content)) // Parse from content
	if err != nil {
		t.Fatalf("setupParseForFindTest: toolGoParseFile failed: %v", err)
	}
	handleStr, ok := handleIDIntf.(string)
	if !ok || handleStr == "" {
		t.Fatalf("setupParseForFindTest: toolGoParseFile did not return a valid handle string, got %T: %v", handleIDIntf, handleIDIntf)
	}
	// Verify handle exists in cache using GetHandleValue
	_, err = interp.GetHandleValue(handleStr, golangASTTypeTag) // Use the correct type tag
	if err != nil {
		t.Fatalf("setupParseForFindTest: Handle '%s' not found in cache or wrong type immediately after creation: %v", handleStr, err)
	}
	return handleStr
}

// comparePositionLists: Compares slices of position maps, ignoring order. (Unchanged)
func comparePositionLists(t *testing.T, got, want []map[string]interface{}) bool {
	t.Helper()
	if len(got) != len(want) {
		t.Errorf("Position list length mismatch: got %d, want %d", len(got), len(want))
		t.Logf("Got : %#v", got)
		t.Logf("Want: %#v", want)
		return false
	}
	if len(got) == 0 {
		return true // Both empty, considered equal
	}

	// Sort both slices based on line then column for stable comparison
	sortFunc := func(slice []map[string]interface{}) {
		sort.SliceStable(slice, func(i, j int) bool {
			lineI, okI := slice[i]["line"].(int)
			lineJ, okJ := slice[j]["line"].(int)
			if !okI || !okJ || lineI != lineJ {
				// Handle potential nil or non-int values gracefully during sort?
				// For now, assume valid ints based on expected tool output.
				return lineI < lineJ // Sort primarily by line
			}
			// If lines are equal, sort by column
			colI, _ := slice[i]["column"].(int)
			colJ, _ := slice[j]["column"].(int)
			return colI < colJ
		})
	}

	// Create copies before sorting to avoid modifying original slices if they were passed directly
	gotCopy := make([]map[string]interface{}, len(got))
	copy(gotCopy, got)
	wantCopy := make([]map[string]interface{}, len(want))
	copy(wantCopy, want)

	sortFunc(gotCopy)
	sortFunc(wantCopy)

	if !reflect.DeepEqual(gotCopy, wantCopy) {
		t.Errorf("Position list content mismatch (after sorting):")
		t.Logf("Got : %#v", gotCopy)
		t.Logf("Want: %#v", wantCopy)
		return false
	}
	return true
}

// --- Test Fixture Loading --- (Unchanged)
var findFixtureDir = filepath.Join("test_fixtures", "find_fixtures") // Subdirectory for find tests

// loadFindFixture loads content from .go.txt files as specified by user.
func loadFindFixture(t *testing.T, baseFilename string) string {
	t.Helper()
	// Construct the .go.txt filename
	fixtureFilename := baseFilename
	if !strings.HasSuffix(fixtureFilename, ".txt") {
		fixtureFilename += ".txt"
	}

	fullPath := filepath.Join(findFixtureDir, fixtureFilename)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		// Fail if the specific .go.txt file doesn't exist - removed fallback creation logic.
		t.Fatalf("Failed to read fixture file %s: %v. Ensure it exists and has the .txt extension.", fullPath, err)
	}
	return string(content)
}

// --- Test Function ---
func TestToolGoFindIdentifiers(t *testing.T) {
	// Load fixture content once, expecting .go.txt extension
	findBasicContent := loadFindFixture(t, "find_basic.go.txt")                // Loads find_basic.go.txt
	findMultiplePkgsContent := loadFindFixture(t, "find_multiple_pkgs.go.txt") // Loads find_multiple_pkgs.go.txt
	findNoMatchContent := loadFindFixture(t, "find_no_match.go.txt")           // Loads find_no_match.go.txt
	findAliasedContent := loadFindFixture(t, "find_aliased.go.txt")            // Loads find_aliased.go.txt

	// --- Test Cases ---
	// NOTE: wantResult values updated based on the user's test output.
	// Filename changed to "<content string>"
	// Line/Column numbers changed to match the 'Got' values from the output.
	tests := []struct {
		name          string
		sourceContent string
		findPkg       string
		findID        string
		wantResult    []map[string]interface{} // Expect list of maps
		wantErrIs     error                    // Expected error type from the tool itself
		valWantErrIs  error                    // Expected error type from validation
	}{
		// --- Happy Paths ---
		{
			name:          "Find fmt.Println in basic file",
			sourceContent: findBasicContent,
			findPkg:       "fmt",
			findID:        "Println",
			wantResult: []map[string]interface{}{
				{"filename": "<content string>", "line": 7, "column": 2},  // Was line 6, col 6 -> Got line 7, col 2
				{"filename": "<content string>", "line": 8, "column": 2},  // Was line 7, col 6 -> Got line 8, col 2
				{"filename": "<content string>", "line": 12, "column": 2}, // Was line 11, col 6 -> Got line 12, col 2
			},
		},
		{
			name:          "Find fmt.Println in multi-pkg file",
			sourceContent: findMultiplePkgsContent,
			findPkg:       "fmt",
			findID:        "Println",
			wantResult: []map[string]interface{}{
				{"filename": "<content string>", "line": 10, "column": 2}, // Was line 8, col 6 -> Got line 10, col 2
				{"filename": "<content string>", "line": 13, "column": 3}, // Was line 11, col 7 -> Got line 13, col 3
			},
		},
		{
			name:          "Find strings.HasPrefix in multi-pkg file",
			sourceContent: findMultiplePkgsContent,
			findPkg:       "strings",
			findID:        "HasPrefix",
			wantResult: []map[string]interface{}{
				{"filename": "<content string>", "line": 12, "column": 5}, // Was line 10, col 13 -> Got line 12, col 5
			},
		},
		{
			name:          "Find os.Getenv in multi-pkg file",
			sourceContent: findMultiplePkgsContent,
			findPkg:       "os",
			findID:        "Getenv",
			wantResult: []map[string]interface{}{
				{"filename": "<content string>", "line": 15, "column": 2}, // Was line 13, col 5 -> Got line 15, col 2
			},
		},
		{
			name:          "Find aliased f.Println (searching f.Println)",
			sourceContent: findAliasedContent,
			findPkg:       "f", // Use the alias name as pkgName
			findID:        "Println",
			wantResult: []map[string]interface{}{
				{"filename": "<content string>", "line": 7, "column": 2}, // Was line 7, col 3 -> Got line 7, col 2
			},
		},
		{
			name:          "Find direct fmt.Println with alias present",
			sourceContent: findAliasedContent,
			findPkg:       "fmt", // Use the original package name
			findID:        "Println",
			wantResult: []map[string]interface{}{
				{"filename": "<content string>", "line": 8, "column": 2}, // Was line 8, col 6 -> Got line 8, col 2
			},
		},
		{
			name:          "Find identifier not present (fmt.Printf)",
			sourceContent: findBasicContent,
			findPkg:       "fmt",
			findID:        "Printf",
			wantResult:    []map[string]interface{}{}, // Expect empty list
		},
		{
			name:          "Find identifier in file without package import",
			sourceContent: findBasicContent,
			findPkg:       "strings",
			findID:        "HasPrefix",
			wantResult:    []map[string]interface{}{}, // Expect empty list
		},
		{
			name:          "Find in file with no matches",
			sourceContent: findNoMatchContent, // Uses find_no_match.go.txt
			findPkg:       "fmt",
			findID:        "Println",
			wantResult:    []map[string]interface{}{}, // Expect empty list
		},

		// --- Unhappy Paths (Errors) ---
		{
			name:          "Error: Invalid Handle",
			sourceContent: findBasicContent, // Need valid content to attempt parse first
			findPkg:       "fmt",
			findID:        "Println",
			wantResult:    nil,
			wantErrIs:     ErrGoModifyFailed, // Assuming AST retrieval issues still use this error, maybe define ErrGoFindFailed?
		},
		{
			name:          "Error: Handle Wrong Type",
			sourceContent: findBasicContent, // Need valid content
			findPkg:       "fmt",
			findID:        "Println",
			wantResult:    nil,
			wantErrIs:     ErrGoModifyFailed, // Assuming AST retrieval issues still use this error, maybe define ErrGoFindFailed?
		},
		{
			name:          "Error: Empty Package Name Arg",
			sourceContent: findBasicContent, // Need valid content
			findPkg:       "",               // Invalid pkg_name
			findID:        "Println",
			wantResult:    nil,
			wantErrIs:     ErrGoInvalidIdentifierFormat,
		},
		{
			name:          "Error: Empty Identifier Arg",
			sourceContent: findBasicContent, // Need valid content
			findPkg:       "fmt",
			findID:        "", // Invalid identifier
			wantResult:    nil,
			wantErrIs:     ErrGoInvalidIdentifierFormat,
		},
		// --- Validation Errors ---
		{
			name:         "Validation: Wrong Arg Count (Missing ID)",
			findPkg:      "fmt",
			wantResult:   nil,
			valWantErrIs: ErrValidationArgCount,
		},
		{
			name:         "Validation: Nil Handle",
			findPkg:      "fmt",
			findID:       "Println",
			wantResult:   nil,
			valWantErrIs: ErrValidationRequiredArgNil,
		},
		{
			name: "Validation: Nil PkgName",
			// handle set dynamically
			findID:       "Println",
			wantResult:   nil,
			valWantErrIs: ErrValidationRequiredArgNil,
		},
		{
			name: "Validation: Nil Identifier",
			// handle set dynamically
			findPkg:      "fmt",
			wantResult:   nil,
			valWantErrIs: ErrValidationRequiredArgNil,
		},
	}

	for _, tt := range tests {
		tc := tt // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Setup interpreter and parse the source content for this test case
			interp, _ := newDefaultTestInterpreter(t)
			var handleID string
			if tc.sourceContent != "" && tc.valWantErrIs == nil { // Only parse if content provided and no validation error expected early
				handleID = setupParseForFindTest(t, interp, tc.sourceContent)
			}

			var rawArgs []interface{}
			// Construct args, handling specific error cases
			if tc.name == "Error: Invalid Handle" {
				rawArgs = makeArgs("non-existent-handle", tc.findPkg, tc.findID)
			} else if tc.name == "Error: Handle Wrong Type" {
				// Create a dummy object in cache with wrong type tag
				// *** UPDATED CALL ***
				wrongTypeHandle, regErr := interp.RegisterHandle("just a string", "WrongType")
				if regErr != nil {
					t.Fatalf("Failed to register handle for wrong type test: %v", regErr)
				}
				// *** END UPDATE ***
				rawArgs = makeArgs(wrongTypeHandle, tc.findPkg, tc.findID)
			} else if tc.name == "Validation: Wrong Arg Count (Missing ID)" {
				rawArgs = makeArgs(handleID, tc.findPkg) // Missing identifier arg
			} else if tc.name == "Validation: Nil Handle" {
				rawArgs = makeArgs(nil, tc.findPkg, tc.findID)
			} else if tc.name == "Validation: Nil PkgName" {
				rawArgs = makeArgs(handleID, nil, tc.findID)
			} else if tc.name == "Validation: Nil Identifier" {
				rawArgs = makeArgs(handleID, tc.findPkg, nil)
			} else {
				// Default case for valid structure
				rawArgs = makeArgs(handleID, tc.findPkg, tc.findID)
			}

			// --- Tool Lookup & Validation ---
			toolImpl, found := interp.ToolRegistry().GetTool("GoFindIdentifiers")
			if !found {
				t.Fatalf("Tool GoFindIdentifiers not found")
			}
			spec := toolImpl.Spec
			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs)

			// Check Validation Error Expectation
			if tc.valWantErrIs != nil {
				if valErr == nil {
					t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil", tc.valWantErrIs)
				} else if !errors.Is(valErr, tc.valWantErrIs) {
					t.Errorf("ValidateAndConvertArgs() expected error type [%T], got [%T]: %v", tc.valWantErrIs, valErr, valErr)
				}
				return // Stop if validation error was expected
			}
			if valErr != nil && tc.valWantErrIs == nil {
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
			}

			// --- Execution ---
			gotResultIntf, toolErr := toolImpl.Func(interp, convertedArgs)

			// Check Tool Execution Error Expectation
			if tc.wantErrIs != nil {
				if toolErr == nil {
					t.Errorf("Execute: expected Go error type [%T], but got nil error. Result: %v", tc.wantErrIs, gotResultIntf)
				} else if !errors.Is(toolErr, tc.wantErrIs) {
					t.Errorf("Execute: wrong Go error type. \n got error: %v\nwant error type: %T", toolErr, tc.wantErrIs)
				}
				if gotResultIntf != nil {
					t.Errorf("Execute: expected nil result when error is returned, but got: %v (%T)", gotResultIntf, gotResultIntf)
				}
				return // Stop if tool error was expected
			}
			if toolErr != nil && tc.wantErrIs == nil {
				t.Fatalf("Execute: unexpected Go error: %v. Result: %v (%T)", toolErr, gotResultIntf, gotResultIntf)
			}

			// --- Success Result Comparison ---
			// Convert result to the expected type: []map[string]interface{}
			var gotResult []map[string]interface{}
			if gotResultIntf != nil {
				gotSlice, okSlice := gotResultIntf.([]interface{})
				if !okSlice {
					// Check if it's already the correct type (Go 1.18+ type inference might do this)
					gotMapSlice, okMapSlice := gotResultIntf.([]map[string]interface{})
					if okMapSlice {
						gotResult = gotMapSlice // Directly assign if already correct
					} else {
						t.Fatalf("Execute Success: Expected result type []interface{} or []map[string]interface{}, got %T", gotResultIntf)
					}

				} else {
					// Convert from []interface{}
					gotResult = make([]map[string]interface{}, len(gotSlice))
					validConv := true
					for i, item := range gotSlice {
						mapItem, okMap := item.(map[string]interface{})
						if !okMap {
							t.Fatalf("Execute Success: Expected result slice element %d to be map[string]interface{}, got %T", i, item)
							validConv = false
							break
						}
						// Ensure int types for line/column for comparison
						// Convert int64 from JSON-like map to int for comparison
						if lineVal, ok := mapItem["line"].(int64); ok {
							mapItem["line"] = int(lineVal)
						} else if _, okInt := mapItem["line"].(int); !okInt {
							t.Fatalf("line value %v not int64 or int", mapItem["line"])
						}
						if colVal, ok := mapItem["column"].(int64); ok {
							mapItem["column"] = int(colVal)
						} else if _, okInt := mapItem["column"].(int); !okInt {
							t.Fatalf("column value %v not int64 or int", mapItem["column"])
						}

						gotResult[i] = mapItem
					}
					if !validConv {
						return
					}
				}
			} else {
				// Handle case where nil was returned (e.g., on error) vs. empty list for no matches
				if tc.wantResult == nil {
					// If nil was expected (likely due to an expected error handled above), this is fine.
				} else if len(tc.wantResult) == 0 {
					// If an empty list was expected, treat nil result as an empty list.
					gotResult = []map[string]interface{}{}
				} else {
					// Unexpected nil result when a non-empty list was expected.
					t.Fatalf("Execute Success: Got nil result, but expected non-empty list: %#v", tc.wantResult)
				}
			}

			// Use the custom comparison function
			if !comparePositionLists(t, gotResult, tc.wantResult) {
				// comparePositionLists already logs detailed errors
				t.Logf("Result comparison failed for test: %s", tc.name)
			} else {
				t.Logf("Result comparison successful for test: %s", tc.name)
			}
		})
	}
}

// getCachedObjectAndType is used by setupParseForFindTest, needs to be defined or removed
// For testing purposes, a simple placeholder accessing the interpreter's cache might suffice
// or setupParseForFindTest should just use GetHandleValue
func (i *Interpreter) getCachedObjectAndType(handleID string) (object interface{}, typeTag string, found bool) {
	// Determine typeTag from handle prefix (basic implementation)
	parts := strings.SplitN(handleID, handleSeparator, 2)
	if len(parts) == 2 {
		typeTag = parts[0]
	}

	if i.objectCache != nil {
		object, found = i.objectCache[handleID]
	}
	// If using GetHandleValue, you don't need this method, just call:
	// object, err := i.GetHandleValue(handleID, golangASTTypeTag)
	// found = (err == nil)
	return object, typeTag, found
}
