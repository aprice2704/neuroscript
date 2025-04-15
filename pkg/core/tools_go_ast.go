// filename: pkg/core/tools_go_ast.go
// UPDATED: Register new tool GoFindIdentifiers
package core

import (
	"bytes"
	"errors" // Added for Join
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	// "golang.org/x/tools/go/ast/astutil" // Not needed here
)

// --- Helper Struct ---
type CachedAst struct {
	File *ast.File
	Fset *token.FileSet
}

const golangASTTypeTag = "GolangAST"

// --- GoParseFile Tool ---
// (implementation unchanged)
func toolGoParseFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
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
		sandboxRoot := interpreter.sandboxDir
		if sandboxRoot == "" {
			sandboxRoot = "."
		}
		absPath, secErr := SecureFilePath(filePath, sandboxRoot)
		if secErr != nil {
			return fmt.Sprintf("GoParseFile path error for '%s': %s", filePath, secErr.Error()), secErr
		}
		contentBytes, readErr := os.ReadFile(absPath)
		if readErr != nil {
			return fmt.Sprintf("GoParseFile failed to read file '%s': %s", filePath, readErr.Error()), fmt.Errorf("%w: reading file '%s': %w", ErrInternalTool, filePath, readErr)
		}
		content = string(contentBytes)
	} else if contentHasValue {
		sourceName = "<content string>"
	} else {
		return nil, fmt.Errorf("GoParseFile: Internal logic error determining input source: %w", ErrInternalTool)
	} // Return Go error for internal issues
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, sourceName, content, parser.ParseComments)
	if err != nil {
		return fmt.Sprintf("GoParseFile failed: %s", err.Error()), fmt.Errorf("%w: %w", ErrGoParseFailed, err)
	} // Return wrapped Go error
	cachedData := CachedAst{File: astFile, Fset: fset}
	handleID := interpreter.storeObjectInCache(cachedData, golangASTTypeTag)
	interpreter.logger.Printf("[TOOL GoParseFile] Successfully parsed '%s'. Stored AST+FileSet with handle ID: %s", sourceName, handleID)
	return handleID, nil
}

// --- GoFormatAST Tool ---
// (implementation unchanged except for logging fix)
func toolGoFormatAST(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// *** FIXED: Added missing 'args' argument to Printf call ***
	interpreter.logger.Printf("[TOOL GoFormatAST ENTRY] Received args: %v", args)
	// Use defined errors for validation failures
	if len(args) != 1 {
		return nil, fmt.Errorf("GoFormatAST: Requires exactly 1 argument: handleID (string): %w", ErrValidationArgCount)
	}
	handleID, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("GoFormatAST: Expected string handle ID, got %T: %w", args[0], ErrValidationTypeMismatch)
	}
	if handleID == "" {
		return nil, fmt.Errorf("GoFormatAST: Handle ID cannot be empty: %w", ErrValidationRequiredArgNil)
	}
	interpreter.logger.Printf("[TOOL GoFormatAST] Validated handle ID: %s", handleID)

	obj, err := interpreter.retrieveObjectFromCache(handleID, golangASTTypeTag)
	if err != nil {
		// Wrap error correctly
		return nil, fmt.Errorf("GoFormatAST: %w", errors.Join(ErrGoFormatFailed, err)) // Use Join if Go 1.20+
	}
	cachedAst, ok := obj.(CachedAst)
	if !ok || cachedAst.File == nil || cachedAst.Fset == nil {
		errInternal := fmt.Errorf("internal error - retrieved object for handle '%s' is invalid (%T)", handleID, obj)
		// Wrap internal error with a relevant execution error
		return nil, fmt.Errorf("%w: %w", ErrGoFormatFailed, errInternal) // Or ErrInternalTool? ErrGoFormatFailed seems more relevant.
	}
	interpreter.logger.Printf("[TOOL GoFormatAST] Successfully retrieved CachedAst for handle '%s'.", handleID)

	var buf bytes.Buffer
	cfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	err = cfg.Fprint(&buf, cachedAst.Fset, cachedAst.File)
	if err != nil {
		// Wrap error correctly
		errInternal := fmt.Errorf("failed to format AST for handle '%s': %w", handleID, err)
		return nil, fmt.Errorf("%w: %w", ErrGoFormatFailed, errInternal)
	}
	formattedCode := buf.String()
	interpreter.logger.Printf("[TOOL GoFormatAST] Successfully formatted AST. Returning code string (len: %d).", len(formattedCode))
	return formattedCode, nil
}

// --- Registration ---
// Registers ALL Go AST tools (Parse, Modify, Format, Find)
func registerGoAstTools(registry *ToolRegistry) error {
	// GoParseFile registration
	err := registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name: "GoParseFile", Description: "Parses Go source code from path or content string. Returns AST handle.",
			Args: []ArgSpec{{Name: "path", Type: ArgTypeString, Required: false}, {Name: "content", Type: ArgTypeString, Required: false}}, ReturnType: ArgTypeString,
		}, Func: toolGoParseFile,
	})
	if err != nil {
		fmt.Printf("!!! CRITICAL: Failed to register Go AST tool GoParseFile: %v\n", err)
		return fmt.Errorf("failed to register GoParseFile: %w", err)
	}

	// GoModifyAST registration (Func defined in tools_go_ast_modify.go)
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name: "GoModifyAST", Description: "Modifies Go AST (handle) using directives (change_package, add/remove/replace_import, replace_id). Returns NEW handle on success.",
			Args: []ArgSpec{{Name: "handle", Type: ArgTypeString, Required: true}, {Name: "modifications", Type: ArgTypeAny, Required: true}}, ReturnType: ArgTypeString,
		}, Func: toolGoModifyAST, // Assumes toolGoModifyAST is accessible (defined in another file in the same package)
	})
	if err != nil {
		fmt.Printf("!!! CRITICAL: Failed to register Go AST tool GoModifyAST: %v\n", err)
		return fmt.Errorf("failed to register GoModifyAST: %w", err)
	}

	// GoFormatAST registration
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			// VVV CHANGE THIS LINE VVV
			Name: "GoFormatASTNode", // Changed from "GoFormatAST"
			// ^^^ CHANGE THIS LINE ^^^
			Description: "Formats Go AST (handle). Returns formatted code string.",
			Args:        []ArgSpec{{Name: "handle", Type: ArgTypeString, Required: true}}, ReturnType: ArgTypeString,
		}, Func: toolGoFormatAST, // Keep the function name toolGoFormatAST
	})
	if err != nil {
		fmt.Printf("!!! CRITICAL: Failed to register Go AST tool GoFormatASTNode: %v\n", err) // Updated log message
		return fmt.Errorf("failed to register GoFormatASTNode: %w", err)                      // Updated error message
	}

	// --- GoFindIdentifiers registration ---
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GoFindIdentifiers",
			Description: "Finds occurrences of qualified identifiers (pkg.Symbol) in a Go AST (handle). Returns list of positions.",
			Args: []ArgSpec{
				{Name: "handle", Type: ArgTypeString, Required: true, Description: "Handle for the AST."},
				{Name: "pkg_name", Type: ArgTypeString, Required: true, Description: "Package name part (e.g., 'fmt')."},
				{Name: "identifier", Type: ArgTypeString, Required: true, Description: "Identifier name part (e.g., 'Println')."},
			},
			// *** FIXED: Replaced ArgTypeList with ArgTypeSliceAny ***
			ReturnType: ArgTypeSliceAny, // Returns a list of maps [{filename, line, column}, ...]
		},
		Func: toolGoFindIdentifiers, // Assumes toolGoFindIdentifiers is accessible (defined in another file in the same package)
	})
	if err != nil {
		fmt.Printf("!!! CRITICAL: Failed to register Go AST tool GoFindIdentifiers: %v\n", err)
		return fmt.Errorf("failed to register GoFindIdentifiers: %w", err)
	}
	// --- END NEW ---

	return nil
}
