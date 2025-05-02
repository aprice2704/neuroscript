// filename: pkg/core/tools_go_ast.go
// UPDATED: Register GoGetNodeInfo
package goast

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/core"
)

// --- Helper Struct ---
// --- (CachedAst struct remains unchanged) ---
type CachedAst struct {
	File *ast.File
	Fset *token.FileSet
}

const golangASTTypeTag = "GolangAST"

// --- GoParseFile Tool ---
// --- (toolGoParseFile remains unchanged) ---
func toolGoParseFile(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	var content string
	var filePath string
	var pathArgProvided, contentArgProvided bool
	var pathHasValue, contentHasValue bool
	pathArg := args[0]
	contentArg := args[1]
	if pathArg != nil {
		pathArgProvided = true
		pathStr, ok := pathArg.(string)
		if ok && pathStr != "" {
			filePath = pathStr
			pathHasValue = true
		}
	}
	if contentArg != nil {
		contentArgProvided = true
		contentStr, ok := contentArg.(string)
		if ok && contentStr != "" {
			if !pathHasValue {
				content = contentStr
			}
			contentHasValue = true
		}
	}
	if pathHasValue && contentHasValue {
		return "GoParseFile requires exactly one of 'path' or 'content' argument, both provided.", nil
	}
	if !pathHasValue && !contentHasValue {
		if pathArgProvided && !pathHasValue && !contentArgProvided {
			return fmt.Sprintf("GoParseFile: 'path' argument was provided but empty, and 'content' was not provided."), nil
		}
		if contentArgProvided && !contentHasValue && !pathArgProvided {
			return fmt.Sprintf("GoParseFile: 'content' argument was provided but empty, and 'path' was not provided."), nil
		}
		if pathArgProvided && contentArgProvided && !pathHasValue && !contentHasValue {
			return fmt.Sprintf("GoParseFile: Both 'path' and 'content' arguments were provided but empty."), nil
		}
		return "GoParseFile requires 'path' (string) or 'content' (string) argument.", nil
	}
	var sourceName string
	if pathHasValue {
		sourceName = filePath
		sandboxRoot := interpreter.SandboxDir()
		if sandboxRoot == "" {
			sandboxRoot = "."
		}
		absPath, secErr := core.SecureFilePath(filePath, sandboxRoot)
		if secErr != nil {
			return fmt.Sprintf("GoParseFile path error for '%s': %s", filePath, secErr.Error()), secErr
		}
		contentBytes, readErr := os.ReadFile(absPath)
		if readErr != nil {
			return fmt.Sprintf("GoParseFile failed to read file '%s': %s", filePath, readErr.Error()), fmt.Errorf("%w: reading file '%s': %w", core.ErrInternalTool, filePath, readErr)
		}
		content = string(contentBytes)
	} else if contentHasValue {
		sourceName = "<content string>"
	} else {
		return nil, fmt.Errorf("GoParseFile: Internal logic error determining input source: %w", core.ErrInternalTool)
	}
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, sourceName, content, parser.ParseComments)
	if err != nil {
		return fmt.Sprintf("GoParseFile failed: %s", err.Error()), fmt.Errorf("%w: %w", core.ErrGoParseFailed, err)
	}
	cachedData := CachedAst{File: astFile, Fset: fset}
	handleID, err := interpreter.RegisterHandle(cachedData, golangASTTypeTag)
	if err != nil {
		interpreter.Logger().Error("Failed to register AST handle for '%s': %v", sourceName, err)
		return nil, fmt.Errorf("failed to register AST handle: %w", err)
	}
	interpreter.Logger().Info("Tool: GoParseFile] Successfully parsed '%s'. Stored AST+FileSet with handle ID: %s", sourceName, handleID)
	return handleID, nil
}

// --- GoFormatASTNode Tool ---
// --- (toolGoFormatAST remains unchanged) ---
func toolGoFormatAST(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	interpreter.Logger().Info("Tool: GoFormatAST ENTRY] Received args: %v", args)
	if len(args) != 1 {
		return nil, fmt.Errorf("GoFormatASTNode: Requires exactly 1 argument: handleID (string): %w", core.ErrValidationArgCount)
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("GoFormatASTNode: Expected string handle ID, got %T: %w", args[0], core.ErrValidationTypeMismatch)
	}
	if handleID == "" {
		return nil, fmt.Errorf("GoFormatASTNode: Handle ID cannot be empty: %w", core.ErrValidationRequiredArgNil)
	}
	interpreter.Logger().Info("Tool: GoFormatASTNode] Validated handle ID: %s", handleID)
	obj, err := interpreter.GetHandleValue(handleID, golangASTTypeTag)
	if err != nil {
		return nil, fmt.Errorf("GoFormatASTNode: %w", errors.Join(core.ErrGoFormatFailed, err))
	}
	cachedAst, ok := obj.(CachedAst)
	if !ok || cachedAst.File == nil || cachedAst.Fset == nil {
		errInternal := fmt.Errorf("internal error - retrieved object for handle '%s' is invalid (%T)", handleID, obj)
		return nil, fmt.Errorf("%w: %w", core.ErrGoFormatFailed, errInternal)
	}
	interpreter.Logger().Info("Tool: GoFormatASTNode] Successfully retrieved CachedAst for handle '%s'.", handleID)
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	err = cfg.Fprint(&buf, cachedAst.Fset, cachedAst.File)
	if err != nil {
		errInternal := fmt.Errorf("failed to format AST for handle '%s': %w", handleID, err)
		return nil, fmt.Errorf("%w: %w", core.ErrGoFormatFailed, errInternal)
	}
	formattedCode := buf.String()
	interpreter.Logger().Info("Tool: GoFormatASTNode] Successfully formatted AST. Returning code string (len: %d).", len(formattedCode))
	return formattedCode, nil
}

// --- Registration ---
// Registers Go AST tools (Parse, Modify, Format, Find, Analyze).
func registerGoAstTools(registry *core.ToolRegistry) error {
	var registrationErrors []error
	collectRegErr := func(toolName string, err error) {
		if err != nil {
			fmt.Printf("!!! CRITICAL: Failed to register Go AST tool %s: %v\n", toolName, err)
			registrationErrors = append(registrationErrors, fmt.Errorf("failed to register %s: %w", toolName, err))
		}
	}

	// GoParseFile
	collectRegErr("GoParseFile", registry.RegisterTool(core.ToolImplementation{Spec: core.ToolSpec{Name: "GoParseFile", Description: "Parses Go source code from path or content string. Returns AST handle.", Args: []core.ArgSpec{{Name: "path", Type: core.ArgTypeString, Required: false}, {Name: "content", Type: core.ArgTypeString, Required: false}}, ReturnType: core.ArgTypeString}, Func: toolGoParseFile}))
	// GoModifyAST
	collectRegErr("GoModifyAST", registry.RegisterTool(core.ToolImplementation{Spec: core.ToolSpec{Name: "GoModifyAST", Description: "Modifies Go AST (handle) using directives. Returns NEW handle on success.", Args: []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true}, {Name: "modifications", Type: core.ArgTypeMap, Required: true}}, ReturnType: core.ArgTypeString}, Func: toolGoModifyAST})) // Assumes toolGoModifyAST exists
	// GoFormatASTNode
	collectRegErr("GoFormatASTNode", registry.RegisterTool(core.ToolImplementation{Spec: core.ToolSpec{Name: "GoFormatASTNode", Description: "Formats Go AST (handle). Returns formatted code string.", Args: []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true}}, ReturnType: core.ArgTypeString}, Func: toolGoFormatAST}))
	// GoFindIdentifiers
	collectRegErr("GoFindIdentifiers", registry.RegisterTool(core.ToolImplementation{Spec: core.ToolSpec{Name: "GoFindIdentifiers", Description: "Finds occurrences of qualified identifiers (pkg.Symbol) in a Go AST (handle). Returns list of positions.", Args: []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle for the AST."}, {Name: "pkg_name", Type: core.ArgTypeString, Required: true, Description: "Package name part (e.g., 'fmt')."}, {Name: "identifier", Type: core.ArgTypeString, Required: true, Description: "Identifier name part (e.g., 'Println')."}}, ReturnType: core.ArgTypeSliceMap}, Func: toolGoFindIdentifiers})) // Assumes toolGoFindIdentifiers exists

	// +++ Add GoGetNodeInfo registration +++
	collectRegErr("GoGetNodeInfo", registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name:        "GoGetNodeInfo",
			Description: "Finds the AST node at a specific position (offset or line/column) within an AST handle and returns information about it.",
			Args: []core.ArgSpec{
				{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle for the AST (from GoParseFile)."},
				{Name: "position", Type: core.ArgTypeMap, Required: true, Description: "Map specifying position, e.g. {\"offset\": 123} or {\"line\": 10, \"column\": 5} (1-based)."},
			},
			ReturnType: core.ArgTypeMap, // Returns map describing the node, or nil if not found/error
		},
		Func: toolGoGetNodeInfo, // Assumes toolGoGetNodeInfo exists (in tools_go_ast_analyze.go)
	}))
	// +++ End GoGetNodeInfo registration +++

	if len(registrationErrors) > 0 {
		return fmt.Errorf("errors registering basic Go AST tools: %s", errors.Join(registrationErrors...))
	} // Use errors.Join
	return nil // Success
}

// --- (NewDefaultTestInterpreter remains unchanged) ---
func NewDefaultTestInterpreter(t *testing.T) (*core.Interpreter, string) {
	t.Helper()
	return core.NewTestInterpreter(t, nil, nil)
}
