// filename: pkg/core/tools_go_ast_package_helpers_test.go
package goast

import (
	"bytes"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	"strings"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// formatNode helper (remains the same)
func formatNode(fset *token.FileSet, node ast.Node) (string, error) {
	var buf bytes.Buffer
	err := format.Node(&buf, fset, node)
	if err != nil {
		return "", err
	}
	return buf.String(), nil
}

// TestApplyAstImportChanges tests the import manipulation logic.
func TestApplyAstImportChanges(t *testing.T) {
	testInterpreter, interpErr := core.NewDefaultTestInterpreter(t)
	if len(interpErr) > 0 {
		t.Fatalf("Failed to create default test interpreter: %v", interpErr)
	}

	testCases := []struct {
		name         string
		inputCode    string
		oldPath      string
		newImports   map[string]string // path -> "" (value unused)
		expectedCode string
		expectError  bool
	}{
		{
			name: "Remove existing import, add one new",
			inputCode: `package main
import "fmt"
import "old/path"
import "log"
func main() { fmt.Println("Hello"); log.Println("World") }`,
			oldPath:    "old/path",
			newImports: map[string]string{"new/path/one": ""},
			expectedCode: `package main
import (
	"fmt"
	"log"
	"new/path/one"
)
func main() { fmt.Println("Hello"); log.Println("World") }`,
			expectError: false,
		},
		{
			name: "Remove existing import, add multiple new",
			inputCode: `package main
import ( "fmt"; "old/path"; "log" )
func main() {}`,
			oldPath: "old/path",
			newImports: map[string]string{
				"new/path/one": "",
				"new/path/two": "",
			},
			expectedCode: `package main
import (
	"fmt"
	"log"
	"new/path/one"
	"new/path/two"
)
func main() {}`,
			expectError: false,
		},
		{
			name: "Remove import that does not exist, add one",
			inputCode: `package main
import "fmt"
func main() {}`,
			oldPath:    "non/existent/path",
			newImports: map[string]string{"new/path/one": ""},
			expectedCode: `package main
import (
	"fmt"
	"new/path/one"
)
func main() {}`,
			expectError: false, // Should not error if old path not found
		},
		{
			name: "Add import that already exists (idempotency)",
			inputCode: `package main
import ( "fmt"; "new/path/one" )
func main() {}`,
			oldPath:    "old/path",                            // Doesn't exist, will be ignored
			newImports: map[string]string{"new/path/one": ""}, // Already exists
			expectedCode: `package main
import (
	"fmt"
	"new/path/one"
)
func main() {}`, // Code should remain unchanged after formatting
			expectError: false,
		},
		{
			name: "Add multiple imports, one already exists",
			inputCode: `package main
import ( "fmt"; "new/path/one" )
func main() {}`,
			oldPath: "old/path",
			newImports: map[string]string{
				"new/path/one": "", // Already exists
				"new/path/two": "", // New
			},
			expectedCode: `package main
import (
	"fmt"
	"new/path/one"
	"new/path/two"
)
func main() {}`,
			expectError: false,
		},
		{
			name:      "File with no imports, add one",
			inputCode: `package main`,
			oldPath:   "old/path",
			newImports: map[string]string{
				"new/path/one": "",
			},
			expectedCode: `package main

import "new/path/one"
`,
			expectError: false,
		},
		{
			name: "Remove only import, add none",
			inputCode: `package main
import "old/path"
func main() {}`,
			oldPath:    "old/path",
			newImports: map[string]string{},
			expectedCode: `package main

func main() {}
`,
			expectError: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fset := token.NewFileSet()
			astFile, parseErr := parser.ParseFile(fset, "input.go", tc.inputCode, parser.ParseComments) // Renamed err variable
			if parseErr != nil {
				t.Fatalf("Failed to parse input code: %v", parseErr)
			}

			// *** Call applyAstImportChanges with the interpreter from your existing helper ***
			callErr := applyAstImportChanges(fset, astFile, tc.oldPath, tc.newImports, testInterpreter) // Renamed err variable

			if tc.expectError {
				if callErr == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if callErr != nil {
					t.Errorf("Did not expect an error, but got: %v", callErr)
				}
				formattedCode, formatErr := formatNode(fset, astFile)
				if formatErr != nil {
					t.Fatalf("Failed to format resulting AST: %v", formatErr)
				}

				// Normalize whitespace for comparison
				normalize := func(s string) string { return strings.Join(strings.Fields(s), " ") }
				expectedNorm := normalize(tc.expectedCode)
				actualNorm := normalize(formattedCode)

				if actualNorm != expectedNorm {
					t.Errorf("Normalized code mismatch.\nExpected (raw):\n---\n%s\n---\nGot (raw):\n---\n%s\n---", tc.expectedCode, formattedCode)
				}
			}
		})
	}
}
