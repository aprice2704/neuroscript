// NeuroScript Version: 0.3.1
// File version: 0.0.6
// Fix toolset registration call arguments and remove placeholder vars.
// filename: pkg/core/tools/goast/tools_go_ast.go

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

	// "testing" // testing import removed

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/toolsets" // Import toolsets
)

// --- Helper Struct ---
type CachedAst struct {
	File *ast.File
	Fset *token.FileSet
}

const golangASTTypeTag = "GolangAST"

// --- GoParseFile Tool ---
// (Implementation unchanged)
func toolGoParseFile(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	var content string
	var filePath string
	var pathHasValue, contentHasValue bool
	if len(args) < 2 {
		return nil, fmt.Errorf("%w: GoParseFile expects path and content args", core.ErrInvalidArgument)
	}
	pathArg := args[0]
	contentArg := args[1]
	if pathArg != nil {
		pathStr, ok := pathArg.(string)
		if ok && pathStr != "" {
			filePath = pathStr
			pathHasValue = true
		}
	}
	if contentArg != nil {
		contentStr, ok := contentArg.(string)
		if ok && contentStr != "" {
			if !pathHasValue {
				content = contentStr
			}
			contentHasValue = true
		}
	}
	if pathHasValue && contentHasValue {
		contentHasValue = false
	} // Prefer path
	if !pathHasValue && !contentHasValue {
		return nil, fmt.Errorf("%w: GoParseFile requires non-empty 'path' or 'content' argument", core.ErrInvalidArgument)
	}
	var sourceName string
	if pathHasValue {
		sourceName = filePath
		sandboxRoot := interpreter.SandboxDir()
		absPath, secErr := core.SecureFilePath(filePath, sandboxRoot)
		if secErr != nil {
			return nil, fmt.Errorf("GoParseFile path error for '%s': %w", filePath, secErr)
		}
		contentBytes, readErr := os.ReadFile(absPath)
		if readErr != nil {
			return nil, fmt.Errorf("GoParseFile failed to read file '%s': %w", filePath, readErr)
		}
		content = string(contentBytes)
	} else {
		sourceName = "<content string>"
	}
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, sourceName, content, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("GoParseFile failed: %w", err)
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
// (Implementation unchanged)
func toolGoFormatAST(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("%w: GoFormatASTNode requires exactly 1 argument: handleID (string)", core.ErrInvalidArgument)
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("%w: GoFormatASTNode expected string handle ID, got %T", core.ErrInvalidArgument, args[0])
	}
	if handleID == "" {
		return nil, fmt.Errorf("%w: GoFormatASTNode handle ID cannot be empty", core.ErrInvalidArgument)
	}
	obj, err := interpreter.GetHandleValue(handleID, golangASTTypeTag)
	if err != nil {
		return nil, fmt.Errorf("GoFormatASTNode: %w", err)
	}
	cachedAst, ok := obj.(CachedAst)
	if !ok || cachedAst.File == nil || cachedAst.Fset == nil {
		return nil, fmt.Errorf("%w: GoFormatASTNode retrieved invalid object for handle '%s' (%T)", core.ErrInternalTool, handleID, obj)
	}
	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	err = cfg.Fprint(&buf, cachedAst.Fset, cachedAst.File)
	if err != nil {
		return nil, fmt.Errorf("GoFormatASTNode failed to format AST for handle '%s': %w", handleID, err)
	}
	return buf.String(), nil
}

// --- init function for toolset registration ---
func init() {
	// Provide a name for the toolset ("goast") and the registration function
	toolsets.AddToolsetRegistration("goast", RegisterGoAstTools) // +++ Added toolset name +++
}

// --- Registration Function ---
// RegisterGoAstTools registers Go AST tools. Called via the toolsets mechanism.
func RegisterGoAstTools(registry core.ToolRegistrar) error {
	var registrationErrors []error
	collectRegErr := func(toolName string, err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "!!! CRITICAL: Failed to register Go AST tool %s: %v\n", toolName, err)
			registrationErrors = append(registrationErrors, fmt.Errorf("failed to register %s: %w", toolName, err))
		}
	}

	// Register tools using ToolImplementation variables defined in other files within this package
	// (toolGoParseFile and toolGoFormatAST defined inline for now as their impl is in this file)

	collectRegErr("GoParseFile", registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{Name: "GoParseFile" /* ... */}, Func: toolGoParseFile,
	}))
	collectRegErr("GoModifyAST", registry.RegisterTool(toolGoModifyASTImpl))
	collectRegErr("GoFormatASTNode", registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{Name: "GoFormatASTNode" /* ... */}, Func: toolGoFormatAST,
	}))
	collectRegErr("GoGetNodeInfo", registry.RegisterTool(toolGoGetNodeInfoImpl))
	collectRegErr("GoFindIdentifiers", registry.RegisterTool(toolGoFindIdentifiersImpl))

	if len(registrationErrors) > 0 {
		return fmt.Errorf("errors registering Go AST tools: %w", errors.Join(registrationErrors...))
	}
	fmt.Println("--- Go AST Tools Registered ---")
	return nil
}

// --- Removed Placeholders ---
// Removed placeholder vars like:
// var toolGoModifyASTImpl core.ToolImplementation
// var toolGoGetNodeInfoImpl core.ToolImplementation
// var toolGoFindIdentifiersImpl core.ToolImplementation
// as their definitions are in the respective implementation files.
