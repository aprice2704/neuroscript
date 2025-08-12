// filename: pkg/parser/parser_test.go
package parser

import (
	"os"
	"path/filepath"
	"strings" // Import strings package
	"testing"

	"github.com/aprice2704/neuroscript/pkg/logging"
	// Import necessary packages
	// "github.com/antlr4-go/antlr/v4" // No longer needed directly here
	// For logger interface/struct
)

func TestNeuroScriptParser(t *testing.T) {
	// Use NoOpLogger struct literal directly
	logger := logging.NewNoOpLogger()
	parserAPI := NewParserAPI(logger)

	// Use a map for easier management if test cases grow numerous
	testCases := map[string]struct {
		expectError bool
		desc        string
	}{
		"valid_basic.ns.txt":                {false, "Basic valid syntax with metadata"},
		"valid_metadata.ns.txt":             {false, "Valid file and procedure metadata, allows blank lines between funcs"},
		"valid_control_flow.ns.txt":         {false, "Valid if/else/endif blocks (corrected)"},
		"valid_error_handler.ns.txt":        {false, "Valid on_error/endon syntax"},
		"valid_tool_call.ns.txt":            {false, "Valid tool call (using LAST) and map access syntax"},
		"invalid_keyword_case.ns.txt":       {true, "Invalid uppercase keywords"},
		"invalid_syntax_structure.ns.txt":   {true, "Invalid structure (missing means, mismatched end)"},
		"valid_metadata_format.ns.txt":      {false, "Syntactically valid metadata format (semantic check needed later)"},
		"invalid_line_continuation.ns.txt":  {true, "Invalid use of the line continuation character"},
		"invalid_map_key.ns.txt":            {true, "Using a variable as a map key, which is not allowed"},
		"valid_recursive_call.ns.txt":       {false, "A function calling itself should parse correctly"},
		"valid_complex_accessor.ns.txt":     {false, "Using a function call as a map accessor"},
		"valid_comprehensive_syntax.ns.txt": {false, "A wide range of valid syntax constructs"},
		"invalid_syntax_collection.ns.txt":  {true, "A collection of various common syntax errors"},
	}

	fixtureDir := "testdata" // Define fixture directory

	// Read all entries in the testdata directory
	fixtures, err := os.ReadDir(fixtureDir)
	if err != nil {
		t.Fatalf("Failed to read test fixture directory '%s': %v", fixtureDir, err)
	}

	// Iterate over files found in the directory
	for _, fixture := range fixtures {
		if fixture.IsDir() {
			continue // Skip directories
		}

		filename := fixture.Name()
		// Check if the found file is one of our defined test cases
		tc, ok := testCases[filename]
		if !ok {
			// Optionally, fail or warn if an untested file is found
			// t.Logf("Skipping file '%s' found in testdata directory, not in testCases map.", filename)
			continue
		}

		// Use t.Run to create subtests for each file
		t.Run(filename, func(t *testing.T) {
			t.Parallel() // Allow tests to run in parallel

			filePath := filepath.Join(fixtureDir, filename)
			contentBytes, readErr := os.ReadFile(filePath)
			if readErr != nil {
				t.Errorf("Failed to read test fixture file '%s': %v", filePath, readErr)
				return // Fail this subtest
			}
			content := string(contentBytes)

			// --- ADDED CHECK FOR INVALID COMMENTS ---
			if strings.Contains(content, "//") {
				t.Errorf("Test file '%s' contains invalid '//' comments. Use '#' or '--'.", filename)
				return // Fail this subtest
			}
			// --- END ADDED CHECK ---

			t.Logf("Parsing file: %s (ExpectError: %v)", filename, tc.expectError)

			// Parse using the API (which now returns only an error)
			_, parseErr := parserAPI.Parse(content)

			hasError := parseErr != nil

			// Check expectations
			if tc.expectError && !hasError {
				t.Errorf("Expected parsing errors for '%s', but got nil error.", filename)
			}

			if !tc.expectError && hasError {
				// Use %+v to potentially get more detailed error info if available
				t.Errorf("Expected no parsing errors for '%s', but got: %+v", filename, parseErr)
			}

			// Optional: Add basic AST validation if needed later
			// if !hasError && tree != nil {
			//     // Perform simple checks on the returned tree if necessary
			// }

			t.Logf("Finished parsing %s. Errors found: %v", filename, hasError)
		})
	}
}

// Removed unused getTestErrorListener function
