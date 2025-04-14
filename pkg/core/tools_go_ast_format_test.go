// filename: pkg/core/tools_go_ast_format_test.go
package core

import (
	"bytes"
	"errors"
	"fmt"    // Import fmt if used in stubs/logging
	"go/ast" // Import ast
	"go/format"
	"go/parser"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath" // Import strconv if used in stubs
	"strings"       // Import sync if used in stubs
	"testing"
)

// --- START: Local Helpers and Minimal Stubs ---
// Includes necessary functions and types locally to ensure compilation.

type cachedObject struct {
	obj     interface{}
	typeTag string
}
type ToolRegistryAPI interface {
	GetTool(name string) (*ToolDefinition, bool)
	RegisterTool(name string, tool *ToolDefinition)
}
type ToolDefinition struct {
	Spec *ToolParameterSpec
	Func ToolFunc
}
type ToolParameterSpec struct { /* Minimal stub - define fields if ValidateAndConvertArgs needs them */
}

// Registry Stubs
type mockToolRegistry struct{ tools map[string]*ToolDefinition }

func (m *mockToolRegistry) GetTool(name string) (*ToolDefinition, bool) {
	if m.tools == nil {
		return nil, false
	}
	tool, found := m.tools[name]
	return tool, found
}
func (m *mockToolRegistry) RegisterTool(name string, tool *ToolDefinition) {
	if m.tools == nil {
		m.tools = make(map[string]*ToolDefinition)
	}
	m.tools[name] = tool
}
func NewMockToolRegistry() ToolRegistryAPI {
	return &mockToolRegistry{tools: make(map[string]*ToolDefinition)}
}

// Tool Registration Stubs (Minimal mocks of tool behavior)
func RegisterGoFormatTool(registry ToolRegistryAPI) {
	formatFunc := func(interp *Interpreter, args []interface{}) (interface{}, error) {
		// Validation is assumed to be done by ValidateAndConvertArgs stub
		handle := args[0].(string) // Assume validation passed

		nodeIntf, typeTag, found := interp.getCachedObjectAndType(handle)
		// Wrap actual errors under the expected ErrGoFormatFailed
		if !found {
			return nil, fmt.Errorf("%w: %w: handle '%s'", ErrGoFormatFailed, ErrCacheObjectNotFound, handle)
		}
		if typeTag != golangASTTypeTag {
			return nil, fmt.Errorf("%w: %w: handle '%s' has type '%s', expected '%s'", ErrGoFormatFailed, ErrCacheObjectWrongType, handle, typeTag, golangASTTypeTag)
		}

		node, ok := nodeIntf.(ast.Node)
		if !ok {
			return nil, fmt.Errorf("%w: internal error: cached object is not ast.Node (%T)", ErrGoFormatFailed, nodeIntf)
		}

		var buf bytes.Buffer
		fset := token.NewFileSet()
		err := format.Node(&buf, fset, node)
		if err != nil {
			return nil, fmt.Errorf("%w: format.Node error: %v", ErrGoFormatFailed, err)
		}
		return buf.String(), nil
	}
	registry.RegisterTool("GoFormatASTNode", &ToolDefinition{Spec: &ToolParameterSpec{}, Func: formatFunc})
}
func RegisterGoParseTool(registry ToolRegistryAPI) {
	parseFunc := func(interp *Interpreter, args []interface{}) (interface{}, error) {
		// Basic validation within stub
		if len(args) != 2 {
			return nil, errors.New("GoParseFile stub expects 2 args (nil, content)")
		}
		if args[1] == nil {
			return nil, errors.New("GoParseFile stub content arg is nil")
		}
		content, ok := args[1].(string)
		if !ok {
			return nil, errors.New("GoParseFile stub content arg not string")
		}

		fset := token.NewFileSet()
		node, err := parser.ParseFile(fset, "<content string>", content, parser.ParseComments)
		if err != nil {
			return nil, fmt.Errorf("go parse failed in stub: %w", err)
		} // Return underlying parse error
		handle := interp.storeObjectInCache(node, golangASTTypeTag)
		return handle, nil
	}
	registry.RegisterTool("GoParseFile", &ToolDefinition{Spec: &ToolParameterSpec{}, Func: parseFunc})
}

// Fixture Loading Helper (Corrected version from previous step)
var findFixtureDirForFormatTest = filepath.Join("test_fixtures", "find_fixtures")
var fixtureDirForFormatTest = "test_fixtures"

func loadGoFixture(t *testing.T, baseFilename string) string {
	t.Helper()
	pathsToCheck := []string{
		filepath.Join(fixtureDirForFormatTest, baseFilename),
		filepath.Join(findFixtureDirForFormatTest, baseFilename),
	}
	if !strings.HasSuffix(baseFilename, ".txt") {
		pathsToCheck = append(pathsToCheck, filepath.Join(fixtureDirForFormatTest, baseFilename+".txt"))
		pathsToCheck = append(pathsToCheck, filepath.Join(findFixtureDirForFormatTest, baseFilename+".txt"))
	} else {
		baseWithoutTxt := strings.TrimSuffix(baseFilename, ".txt")
		pathsToCheck = append(pathsToCheck, filepath.Join(fixtureDirForFormatTest, baseWithoutTxt))
		pathsToCheck = append(pathsToCheck, filepath.Join(findFixtureDirForFormatTest, baseWithoutTxt))
	}
	var triedPaths []string
	uniquePaths := make(map[string]bool)
	for _, p := range pathsToCheck {
		cleanPath := filepath.Clean(p)
		if _, exists := uniquePaths[cleanPath]; !exists {
			triedPaths = append(triedPaths, cleanPath)
			uniquePaths[cleanPath] = true
		} else {
			continue
		}
		content, err := os.ReadFile(cleanPath)
		if err == nil {
			t.Logf("Loaded fixture: %s", cleanPath)
			return string(content)
		}
		if !os.IsNotExist(err) {
			t.Fatalf("Error reading potential fixture file %s: %v", cleanPath, err)
		}
	}
	t.Fatalf("Failed to find fixture file %s (checked variants: %v)", baseFilename, triedPaths)
	return ""
}

// Setup helper using the local parse stub
func setupParseGoTest(t *testing.T, interp *Interpreter, content string) string {
	t.Helper()
	handleID, err := toolGoParseFile(interp, makeArgs(nil, content)) // Uses local stub
	if err != nil {
		t.Logf("Content that failed parsing in setupParseGoTest:\n%s", content)
		t.Fatalf("setupParseGoTest: toolGoParseFile failed: %v", err)
	}
	handleStr, ok := handleID.(string)
	if !ok || handleStr == "" {
		t.Fatalf("setupParseGoTest: toolGoParseFile did not return a valid handle string, got %T: %v", handleID, handleID)
	}
	_, _, found := interp.getCachedObjectAndType(handleStr) // Use local stub method
	if !found {
		t.Fatalf("setupParseGoTest: Handle '%s' not found in cache immediately after creation!", handleStr)
	}
	return handleStr
}

// --- END: Local Helpers and Minimal Stubs ---

// --- START: Helpers Specific to this Test ---

// countComments counts the number of comments in Go source code using the scanner.
func countComments(t *testing.T, content string) int {
	t.Helper()
	fset := token.NewFileSet()
	var s scanner.Scanner
	s.Init(fset.AddFile("", fset.Base(), len(content)), []byte(content), func(pos token.Position, msg string) { /* ignore errors */ }, scanner.ScanComments) // Requires "go/scanner", "go/token"
	count := 0
	for {
		_, tok, _ := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.COMMENT {
			count++
		}
	}
	return count
}

// checkGoSyntax checks if the given Go source content is syntactically valid.
func checkGoSyntax(t *testing.T, content string) error {
	t.Helper()
	fset := token.NewFileSet()
	_, err := parser.ParseFile(fset, "<formatted output>", content, parser.AllErrors|parser.ParseComments) // Requires "go/parser"
	return err
}

// --- END: Specific Helpers ---

// --- Test Function ---
func TestToolGoFormatASTNode(t *testing.T) {
	// Load fixture content once using the local helper
	findBasicContent := loadGoFixture(t, "find_basic.go.txt")
	findMultiplePkgsContent := loadGoFixture(t, "find_multiple_pkgs.go.txt")
	findAliasedContent := loadGoFixture(t, "find_aliased.go.txt")
	largeComplexContent := loadGoFixture(t, "large_complex_source.go.txt")
	formatUnformattedContent := loadGoFixture(t, "format_unformatted.txt")
	formatFormattedContent := loadGoFixture(t, "format_formatted.txt")

	tests := []struct {
		name                         string
		sourceContent                string
		wantErrIs                    error
		valWantErrIs                 error
		checkCommentCount            bool
		checkSyntax                  bool
		checkAgainstFormattedFixture bool
	}{
		// Test cases (same as before)
		{
			name:              "Format basic file",
			sourceContent:     findBasicContent,
			checkCommentCount: true,
			checkSyntax:       true,
		},
		{
			name:              "Format multi-package file",
			sourceContent:     findMultiplePkgsContent,
			checkCommentCount: true,
			checkSyntax:       true,
		},
		{
			name:              "Format aliased file",
			sourceContent:     findAliasedContent,
			checkCommentCount: true,
			checkSyntax:       true,
		},
		{
			name:              "Format large complex file",
			sourceContent:     largeComplexContent,
			checkCommentCount: true,
			checkSyntax:       true,
		},
		{
			name:                         "Format unformatted file and compare to formatted",
			sourceContent:                formatUnformattedContent,
			checkCommentCount:            true,
			checkSyntax:                  true,
			checkAgainstFormattedFixture: true,
		},
		{
			name:          "Error: Invalid Handle",
			sourceContent: findBasicContent,
			wantErrIs:     ErrGoFormatFailed,
		},
		{
			name:          "Error: Handle Wrong Type",
			sourceContent: findBasicContent,
			wantErrIs:     ErrGoFormatFailed,
		},
		{
			name:          "Validation: Wrong Arg Count (no handle)",
			sourceContent: "",
			valWantErrIs:  ErrValidationArgCount,
		},
		{
			name:          "Validation: Nil Handle",
			sourceContent: "",
			valWantErrIs:  ErrValidationRequiredArgNil,
		},
	}

	for _, tt := range tests {
		tc := tt // Capture range variable
		t.Run(tc.name, func(t *testing.T) {
			// Setup interpreter using local helper/stub
			interp, _ := newDefaultTestInterpreter(t) // Gets interp with mock tools registered

			var handleID string
			var initialCommentCount int

			// Parse only if the test expects valid execution
			if tc.wantErrIs == nil && tc.valWantErrIs == nil {
				if tc.sourceContent == "" {
					t.Fatalf("Test setup error: sourceContent is empty for test '%s'", tc.name)
				}
				if tc.checkCommentCount {
					initialCommentCount = countComments(t, tc.sourceContent) // Use local helper
					t.Logf("Initial comment count for %s: %d", tc.name, initialCommentCount)
				}
				handleID = setupParseGoTest(t, interp, tc.sourceContent) // Use local helper
			}

			var rawArgs []interface{}
			// Construct args (same logic as before)
			switch tc.name {
			case "Error: Invalid Handle":
				if handleID == "" {
					handleID = setupParseGoTest(t, interp, tc.sourceContent)
				}
				rawArgs = makeArgs("non-existent-handle")
			case "Error: Handle Wrong Type":
				if handleID == "" {
					handleID = setupParseGoTest(t, interp, tc.sourceContent)
				}
				wrongTypeHandle := interp.storeObjectInCache("just a string", "WrongType") // Use local stub method
				rawArgs = makeArgs(wrongTypeHandle)
			case "Validation: Wrong Arg Count (no handle)":
				rawArgs = makeArgs()
			case "Validation: Nil Handle":
				rawArgs = makeArgs(nil)
			default: // Happy paths
				rawArgs = makeArgs(handleID)
			}

			// --- Tool Lookup & Validation ---
			toolImpl, found := interp.ToolRegistry().GetTool("GoFormatASTNode") // Use local stub method
			if !found {
				t.Fatalf("Tool GoFormatASTNode not found in registry")
			}

			spec := toolImpl.Spec
			convertedArgs, valErr := ValidateAndConvertArgs(spec, rawArgs) // Use local stub

			// Check Validation Error Expectation (same logic as before)
			if tc.valWantErrIs != nil {
				if valErr == nil {
					t.Errorf("ValidateAndConvertArgs() expected error [%v], but got nil", tc.valWantErrIs)
				} else if !errors.Is(valErr, tc.valWantErrIs) {
					t.Errorf("ValidateAndConvertArgs() expected error type [%T], got [%T]: %v", tc.valWantErrIs, valErr, valErr)
				}
				return
			}
			if valErr != nil && tc.valWantErrIs == nil {
				t.Fatalf("ValidateAndConvertArgs() unexpected validation error: %v", valErr)
			}

			// --- Execution ---
			gotResultIntf, toolErr := toolImpl.Func(interp, convertedArgs)

			// Check Tool Execution Error Expectation (same logic as before)
			if tc.wantErrIs != nil {
				if toolErr == nil {
					t.Errorf("Execute: expected Go error type [%T], but got nil error. Result: %v", tc.wantErrIs, gotResultIntf)
				} else if !errors.Is(toolErr, tc.wantErrIs) {
					isCacheErrorCase := strings.Contains(tc.name, "Invalid Handle") || strings.Contains(tc.name, "Handle Wrong Type")
					expectedUnderlying := ErrCacheObjectNotFound
					if strings.Contains(tc.name, "Handle Wrong Type") {
						expectedUnderlying = ErrCacheObjectWrongType
					}
					// Check if the specific format error occurred and if it wraps the expected cache error
					if errors.Is(tc.wantErrIs, ErrGoFormatFailed) && isCacheErrorCase && errors.Is(toolErr, expectedUnderlying) {
						t.Logf("Execute: Got expected wrapped error type [%T] wrapping [%T]", tc.wantErrIs, expectedUnderlying)
					} else {
						t.Errorf("Execute: wrong Go error type. \n got error: %v\nwant error type [%T]", toolErr, tc.wantErrIs)
					}
				}
				if gotResultIntf != nil {
					t.Errorf("Execute: expected nil result when error [%v] is returned, but got: %v (%T)", toolErr, gotResultIntf, gotResultIntf)
				}
				return
			}
			if toolErr != nil && tc.wantErrIs == nil {
				t.Fatalf("Execute: unexpected Go error: %v. Result: %v (%T)", toolErr, gotResultIntf, gotResultIntf)
			}

			// --- Success Result Verification ---
			gotFormatted, ok := gotResultIntf.(string)
			if !ok {
				t.Fatalf("Execute Success: Expected result type string, got %T", gotResultIntf)
			}

			// Verification logic (same as before)
			syntaxCheckPassed := true
			if tc.checkSyntax {
				syntaxErr := checkGoSyntax(t, gotFormatted) // Use local helper
				if syntaxErr != nil {
					t.Errorf("Formatted output failed syntax check: %v", syntaxErr)
					t.Logf("Formatted output that failed syntax check:\n%s", gotFormatted)
					syntaxCheckPassed = false
				}
			}

			commentCheckPassed := true
			if tc.checkCommentCount {
				finalCommentCount := countComments(t, gotFormatted) // Use local helper
				t.Logf("Final comment count for %s: %d", tc.name, finalCommentCount)
				if initialCommentCount != finalCommentCount {
					t.Errorf("Comment count mismatch: initial=%d, final=%d", initialCommentCount, finalCommentCount)
					commentCheckPassed = false
				}
			}

			formatFixtureCheckPassed := true
			if tc.checkAgainstFormattedFixture {
				wantFormattedNormalized := strings.ReplaceAll(formatFormattedContent, "\r\n", "\n")
				gotFormattedNormalized := strings.ReplaceAll(gotFormatted, "\r\n", "\n")
				if gotFormattedNormalized != wantFormattedNormalized {
					t.Errorf("Formatted output does not match format_formatted.txt content.")
					formatFixtureCheckPassed = false
				}
			}

			if syntaxCheckPassed && commentCheckPassed && formatFixtureCheckPassed {
				t.Logf("Checks passed for %s.", tc.name)
			}
		})
	}
}
