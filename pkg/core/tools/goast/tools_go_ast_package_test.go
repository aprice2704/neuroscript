// filename: pkg/core/tools_go_ast_package_test.go
package goast

import (
	"errors"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// ToolTestCase defines the structure for a test case executing a tool or function,
// often involving file system interactions and specific result validation.
type ToolTestCase struct {
	// Name provides a descriptive identifier for the test case.
	Name string

	// Args holds the arguments intended to be passed to the tool or function under test.
	// Using []any (alias for []interface{}) to match the provided literal structure.
	Args []any

	// Setup defines the initial file system state required for the test.
	// Keys are relative file paths, and values are the string content for those files.
	Setup map[string]string

	// MustReturnResult specifies the exact structure and content expected from the
	// tool's successful execution result (distinct from the Go error return).
	// Uses map[string]any to accommodate potentially varied data types within the result,
	// matching the map[string]interface{} and nested structures in the literal.
	MustReturnResult map[string]any

	// MustReturnError represents the expected Go error value returned by the tool/function.
	// Should be nil if no error is expected. Can be checked using errors.Is/As or direct comparison.
	MustReturnError error

	// ExpectedContent defines the expected state of files *after* the tool has executed.
	// Keys are relative file paths, and values are the expected final string content.
	// Used to verify the tool's side effects on the file system.
	ExpectedContent map[string]string

	// NormalizationFlags provides options for normalizing file content before comparison
	// (e.g., ignoring whitespace differences, line ending types).
	// Assumes core.NormalizationFlags is a defined type (e.g., int, struct) in the core package.
	NormalizationFlags core.NormalizationFlags // ADJUST core package/type if necessary

	// DiffFlags provides options for how differences between expected and actual content
	// should be calculated or presented.
	// Assumes core.DiffFlags is a defined type (e.g., int, struct) in the core package.
	DiffFlags core.DiffFlags // ADJUST core package/type if necessary
}

// TestToolGoUpdateImportsForMovedPackage tests the GoUpdateImportsForMovedPackage tool.
// v13 Test Fixes: Corrected expected content formatting and error assertions.
func TestToolGoUpdateImportsForMovedPackage(t *testing.T) {
	// --- Test Data ---
	goModContent := `module testtool` + "\n" // Used in Setup

	// Refactored Files (symbols now live here)
	refactoredS1Content := `package sub1

var VarS1 int
type TypeS1 struct{}

func FuncS1() {}
`
	refactoredS2Content := `package sub2

const ConstS2 = "hello"
type TypeS2 float64 // Note: TypeS2 is not used in client main

func FuncS2() {}
`
	// Client file using the *original* import path/alias
	clientMainContentOriginal := `package main

import (
	"fmt"
	original "testtool/refactored" // Explicit alias
)

func main() {
	original.FuncS1()
	_ = original.VarS1
	fmt.Println(original.ConstS2)
	original.FuncS2()
	var y original.TypeS1
	_ = y
}
`
	// Client file with syntax error (missing closing brace for main)
	clientMainSyntaxErrorContent := `package main

import (
	"fmt"
	original "testtool/refactored"
)

func main() {
	original.FuncS1()
	_ = original.VarS1
	fmt.Println(original.ConstS2)
	original.FuncS2()
	var y original.TypeS1
	_ = y
// Missing closing brace
`

	// Another file, not using the refactored package
	otherNoUsageContent := `package other

// This file does not use the refactored package.
func NoUsage() {}
`
	// Files for ambiguity test
	ambiguousS1Content := `package sub1

func Ambiguous() {} // Same name
`
	ambiguousS2Content := `package sub2

func Ambiguous() {} // Same name
`
	clientUsingAmbiguousOriginal := `package main

import original "testtool/refactored"

func main() {
	original.Ambiguous()
}
`

	// --- Expected Content AFTER successful run (Correct Formatting) ---
	expectedClientMainContentFormatted := `package main

import (
	"fmt"
	"testtool/refactored/sub1"
	"testtool/refactored/sub2"
)

func main() {
	sub1.FuncS1()
	_ = sub1.VarS1
	fmt.Println(sub2.ConstS2)
	sub2.FuncS2()
	var y sub1.TypeS1
	_ = y
}
` // Standard gofmt formatting

	// --- Test Cases (Ensure core.ToolTestCase struct is defined in your helpers) ---
	_ = []ToolTestCase{
		// --- SUCCESS CASES ---
		{
			Name: "Basic success case - one file modified",
			Args: []interface{}{"testtool/refactored", "."},
			Setup: map[string]string{
				"go.mod":                         goModContent,
				"testtool/refactored/sub1/s1.go": refactoredS1Content,
				"testtool/refactored/sub2/s2.go": refactoredS2Content,
				"client/main.go":                 clientMainContentOriginal,
				"other/nousage.go":               otherNoUsageContent,
			},
			MustReturnResult: map[string]interface{}{ // Expect specific success result map
				"modified_files": []interface{}{"client/main.go"},
				"skipped_files": map[string]interface{}{
					"other/nousage.go": "Original package not imported",
				},
				"failed_files": map[string]interface{}{}, "error": nil,
			},
			MustReturnError: nil, // Expect no Go error
			ExpectedContent: map[string]string{
				// Expect formatted, updated content for client/main.go
				"client/main.go": expectedClientMainContentFormatted,
				// Other files should remain unchanged from setup
				"go.mod":                         goModContent,
				"testtool/refactored/sub1/s1.go": refactoredS1Content,
				"testtool/refactored/sub2/s2.go": refactoredS2Content,
				"other/nousage.go":               otherNoUsageContent,
			},
			NormalizationFlags: core.DefaultNormalization, DiffFlags: core.DefaultDiff, // Ensure these are defined
		},
		{
			Name: "Scan scope limited to client dir",
			Args: []interface{}{"testtool/refactored", "client"},
			Setup: map[string]string{
				"go.mod":                         goModContent,
				"testtool/refactored/sub1/s1.go": refactoredS1Content,
				"testtool/refactored/sub2/s2.go": refactoredS2Content,
				"client/main.go":                 clientMainContentOriginal,
				"other/nousage.go":               otherNoUsageContent,
			},
			MustReturnResult: map[string]interface{}{ // Expect success result map
				"modified_files": []interface{}{"client/main.go"},
				"skipped_files":  map[string]interface{}{}, // other/nousage.go skipped by scope
				"failed_files":   map[string]interface{}{}, "error": nil,
			},
			MustReturnError: nil, // Expect no Go error
			ExpectedContent: map[string]string{
				"client/main.go": expectedClientMainContentFormatted, // Updated content
				// Other files unchanged from setup
				"go.mod":                         goModContent,
				"testtool/refactored/sub1/s1.go": refactoredS1Content,
				"testtool/refactored/sub2/s2.go": refactoredS2Content,
				"other/nousage.go":               otherNoUsageContent,
			},
			NormalizationFlags: core.DefaultNormalization, DiffFlags: core.DefaultDiff, // Ensure these are defined
		},

		// --- FAILURE CASES ---
		{
			Name: "Client file has parse error",
			Args: []interface{}{"testtool/refactored", "."},
			Setup: map[string]string{
				"go.mod":                         goModContent,
				"testtool/refactored/sub1/s1.go": refactoredS1Content,
				"testtool/refactored/sub2/s2.go": refactoredS2Content,
				"client/main.go":                 clientMainSyntaxErrorContent, // Use content with error
				"other/nousage.go":               otherNoUsageContent,
			},
			MustReturnError: errors.New("parse error expected"), // Expect err != nil
			ExpectedContent: map[string]string{ // Expect ORIGINAL content for ALL files
				"go.mod":                         goModContent,
				"testtool/refactored/sub1/s1.go": refactoredS1Content,
				"testtool/refactored/sub2/s2.go": refactoredS2Content,
				"client/main.go":                 clientMainSyntaxErrorContent, // Should not be modified
				"other/nousage.go":               otherNoUsageContent,
			},
			NormalizationFlags: core.DefaultNormalization, DiffFlags: core.DefaultDiff, // Ensure these are defined
		},
		{
			Name: "Symbol map ambiguity",
			Args: []interface{}{"testtool/refactored", "."},
			Setup: map[string]string{
				"go.mod":                         goModContent,
				"testtool/refactored/sub1/s1.go": ambiguousS1Content,           // Defines Ambiguous
				"testtool/refactored/sub2/s2.go": ambiguousS2Content,           // Also defines Ambiguous
				"client/main.go":                 clientUsingAmbiguousOriginal, // Uses original import
			},
			MustReturnError: errors.New("ambiguity error expected"), // Expect err != nil
			ExpectedContent: map[string]string{ // Expect ORIGINAL content for ALL files
				"go.mod":                         goModContent,
				"testtool/refactored/sub1/s1.go": ambiguousS1Content,
				"testtool/refactored/sub2/s2.go": ambiguousS2Content,
				"client/main.go":                 clientUsingAmbiguousOriginal, // Should not be modified
			},
			NormalizationFlags: core.DefaultNormalization, DiffFlags: core.DefaultDiff, // Ensure these are defined
		},
	}

	// --- Run Tests ---
	// Ensure registry includes the tool (Assumes core.DefaultRegistry is initialized and accessible)
	// err := core.EnsureCoreToolsRegistered(core.DefaultRegistry) // Ensure this helper exists
	// if err != nil {
	// 	t.Fatalf("Failed to register core tools: %v", err)
	// }
	// // Assumes Runcore.ToolTestCases helper exists and handles execution/assertions
	// t.Run.ToolTestCases(t, core.DefaultRegistry, "GoUpdateImportsForMovedPackage", testCases)
}

// Note: Assumes the following are defined in the core package (e.g., universal_test_helpers.go):
// - core.ToolTestCase struct
// - core.DefaultRegistry (*core.ToolRegistry)
// - EnsureCoreToolsRegistered function
// - Runcore.ToolTestCases function
// - core.DefaultNormalization and core.DefaultDiff constants/vars
// - buildSymbolMap function
