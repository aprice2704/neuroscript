// filename: pkg/core/tools_go_ast.go
// UPDATED: Register new tool GoFindIdentifiers
// UPDATED: Use RegisterHandle and GetHandleValue
// UPDATED: REMOVED registration for GoUpdateImportsForMovedPackage (moved to tools_go_ast_package.go)
package goast

import (
	"bytes"
	"errors" // Added for Join
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"strings" // Needed for registration error joining fallback

	"github.com/aprice2704/neuroscript/pkg/core"
	// "golang.org/x/tools/go/ast/astutil" // Not needed here anymore? Or maybe by find/modify? Keep for now.
)

// --- Helper Struct ---
type CachedAst struct {
	File *ast.File
	Fset *token.FileSet
}

const golangASTTypeTag = "GolangAST" // Use this as the type prefix for handles

// --- GoParseFile Tool ---
// (Implementation remains the same as before)
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
	} // Return error msg string
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
	} // Return error msg string
	var sourceName string
	if pathHasValue {
		sourceName = filePath
		sandboxRoot := interpreter.SandboxDir() // TODO: Get sandbox from AgentContext if applicable?
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
	} // Return Go error for internal issues
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, sourceName, content, parser.ParseComments)
	if err != nil {
		return fmt.Sprintf("GoParseFile failed: %s", err.Error()), fmt.Errorf("%w: %w", core.ErrGoParseFailed, err)
	} // Return wrapped Go error
	cachedData := CachedAst{File: astFile, Fset: fset}
	handleID, err := interpreter.RegisterHandle(cachedData, golangASTTypeTag)
	if err != nil {
		interpreter.Logger().Error("Failed to register AST handle for '%s': %v", sourceName, err)
		return nil, fmt.Errorf("failed to register AST handle: %w", err) // Return internal error
	}
	interpreter.Logger().Info("Tool: GoParseFile] Successfully parsed '%s'. Stored AST+FileSet with handle ID: %s", sourceName, handleID)
	return handleID, nil
}

// --- GoFormatASTNode Tool ---
// (Implementation remains the same as before)
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
// Registers BASIC Go AST tools (Parse, Modify, Format, Find).
// Package-level refactoring tools are registered separately.
func registerGoAstTools(registry *core.ToolRegistry) error {
	var registrationErrors []error

	// Helper to collect registration errors
	collectRegErr := func(toolName string, err error) {
		if err != nil {
			fmt.Printf("!!! CRITICAL: Failed to register Go AST tool %s: %v\n", toolName, err)
			registrationErrors = append(registrationErrors, fmt.Errorf("failed to register %s: %w", toolName, err))
		}
	}

	// GoParseFile registration
	collectRegErr("GoParseFile", registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name: "GoParseFile", Description: "Parses Go source code from path or content string. Returns AST handle.",
			Args: []core.ArgSpec{{Name: "path", Type: core.ArgTypeString, Required: false}, {Name: "content", Type: core.ArgTypeString, Required: false}}, ReturnType: core.ArgTypeString,
		}, Func: toolGoParseFile,
	}))

	// GoModifyAST registration (Func defined in tools_go_ast_modify.go)
	collectRegErr("GoModifyAST", registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name: "GoModifyAST", Description: "Modifies Go AST (handle) using directives (change_package, add/remove/replace_import, replace_id). Returns NEW handle on success.",
			Args: []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true}, {Name: "modifications", Type: core.ArgTypeAny, Required: true}}, ReturnType: core.ArgTypeString,
		}, Func: toolGoModifyAST, // Assumes toolGoModifyAST is accessible (defined in another file in the same package)
	}))

	// GoFormatASTNode registration
	collectRegErr("GoFormatASTNode", registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name:        "GoFormatASTNode", // Changed from "GoFormatAST"
			Description: "Formats Go AST (handle). Returns formatted code string.",
			Args:        []core.ArgSpec{{Name: "handle", Type: core.ArgTypeString, Required: true}}, ReturnType: core.ArgTypeString,
		}, Func: toolGoFormatAST, // Keep the function name toolGoFormatAST
	}))

	// GoFindIdentifiers registration
	collectRegErr("GoFindIdentifiers", registry.RegisterTool(core.ToolImplementation{
		Spec: core.ToolSpec{
			Name:        "GoFindIdentifiers",
			Description: "Finds occurrences of qualified identifiers (pkg.Symbol) in a Go AST (handle). Returns list of positions.",
			Args: []core.ArgSpec{
				{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle for the AST."},
				{Name: "pkg_name", Type: core.ArgTypeString, Required: true, Description: "Package name part (e.g., 'fmt')."},
				{Name: "identifier", Type: core.ArgTypeString, Required: true, Description: "Identifier name part (e.g., 'Println')."},
			},
			ReturnType: core.ArgTypeSliceAny, // Returns a list of maps [{filename, line, column}, ...]
		},
		Func: toolGoFindIdentifiers, // Assumes toolGoFindIdentifiers is accessible (defined in another file in the same package)
	}))

	// --- REMOVED: GoUpdateImportsForMovedPackage registration (moved to tools_go_ast_package.go) ---

	// Combine collected errors if any
	if len(registrationErrors) > 0 {
		errorMessages := make([]string, len(registrationErrors))
		for i, e := range registrationErrors {
			errorMessages[i] = e.Error()
		}
		// Use errors.Join if Go 1.20+, otherwise fallback
		// return errors.Join(registrationErrors...) // Preferred
		return fmt.Errorf("errors registering basic Go AST tools: %s", strings.Join(errorMessages, "; ")) // Fallback
	}

	return nil // Success
}
