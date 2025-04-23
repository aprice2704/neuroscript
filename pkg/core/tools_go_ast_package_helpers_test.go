// filename: pkg/core/tools_go_ast_package_helpers_test.go
package core

import (
	"bytes"
	// "errors" // No longer needed here? Keep if other tests need it.
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"

	// "os" // No longer needed here
	// "path/filepath" // No longer needed here
	"strings"
	"testing"
	// No external assertion libs needed, as per AI_README.md
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

// TestApplyAstImportChanges (remains the same as before)
func TestApplyAstImportChanges(t *testing.T) {
	testCases := []struct {
		name         string
		inputCode    string
		oldPath      string
		newImports   map[string]string // path -> alias (alias unused for now)
		expectedCode string
		expectError  bool
	}{
		{
			name: "Remove existing import, add one new",
			inputCode: `package main

import "fmt"
import "old/path"
import "log"

func main() {
	fmt.Println("Hello")
	log.Println("World")
}
`,
			oldPath:    "old/path",
			newImports: map[string]string{"new/path/one": ""},
			expectedCode: `package main

import (
	"fmt"
	"log"
	"new/path/one"
)

func main() {
	fmt.Println("Hello")
	log.Println("World")
}
`,
			expectError: false,
		},
		{
			name: "Remove existing import, add multiple new",
			inputCode: `package main

import (
	"fmt"
	"old/path"
	"log"
)

func main() {}
`,
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

func main() {}
`,
			expectError: false,
		},
		{
			name: "Remove import that does not exist, add one",
			inputCode: `package main

import "fmt"

func main() {}
`,
			oldPath:    "non/existent/path",
			newImports: map[string]string{"new/path/one": ""},
			expectedCode: `package main

import (
	"fmt"
	"new/path/one"
)

func main() {}
`,
			expectError: false,
		},
		{
			name: "Add import that already exists (idempotency)",
			inputCode: `package main

import (
	"fmt"
	"new/path/one"
)

func main() {}
`,
			oldPath:    "old/path",
			newImports: map[string]string{"new/path/one": ""},
			expectedCode: `package main

import (
	"fmt"
	"new/path/one"
)

func main() {}
`,
			expectError: false,
		},
		{
			name: "Add multiple imports, one already exists",
			inputCode: `package main

import (
	"fmt"
	"new/path/one"
)

func main() {}
`,
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

func main() {}
`,
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

func main() {}
`,
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
			astFile, err := parser.ParseFile(fset, "input.go", tc.inputCode, parser.ParseComments)
			if err != nil {
				t.Fatalf("Failed to parse input code: %v", err)
			}
			err = applyAstImportChanges(fset, astFile, tc.oldPath, tc.newImports)
			if tc.expectError {
				if err == nil {
					t.Errorf("Expected an error, but got nil")
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect an error, but got: %v", err)
				}
				formattedCode, formatErr := formatNode(fset, astFile)
				if formatErr != nil {
					t.Fatalf("Failed to format resulting AST: %v", formatErr)
				}
				expected := strings.TrimSpace(tc.expectedCode)
				actual := strings.TrimSpace(formattedCode)
				if actual != expected {
					t.Errorf("Formatted code does not match expected.\nExpected:\n---\n%s\n---\nGot:\n---\n%s\n---", expected, actual)
				}
			}
		})
	}
}

// --- REMOVED TestDetermineRefactoredDir and setupTestModule ---
