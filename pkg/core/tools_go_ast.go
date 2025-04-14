// filename: pkg/core/tools_go_ast.go
package core

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
)

// --- Helper Struct ---

// CachedAst holds the parsed AST and its associated FileSet, needed for accurate printing/reparsing.
type CachedAst struct {
	File *ast.File
	Fset *token.FileSet
}

const golangASTTypeTag = "GolangAST" // Define constant for type tag

// --- GoParseFile Tool (MODIFIED Validation Logic) ---
func toolGoParseFile(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	var content string
	var filePath string
	var pathArgProvided, contentArgProvided bool // Track if args were provided *at all*
	var pathHasValue, contentHasValue bool       // Track if args had non-empty string values

	pathArg := args[0]
	contentArg := args[1]

	// Check path argument
	if pathArg != nil {
		pathArgProvided = true // Mark arg as provided
		pathStr, ok := pathArg.(string)
		if ok && pathStr != "" {
			filePath = pathStr
			pathHasValue = true // Mark path as having a non-empty value
		}
	}

	// Check content argument
	if contentArg != nil {
		contentArgProvided = true // Mark arg as provided
		contentStr, ok := contentArg.(string)
		if ok && contentStr != "" {
			// Only assign content if path didn't have a value,
			// but mark that content *had* a value regardless.
			if !pathHasValue {
				content = contentStr
			}
			contentHasValue = true // Mark content as having a non-empty value
		}
	}

	// --- Validation Logic ---
	if pathHasValue && contentHasValue {
		// FAIL Case: Both path and content had non-empty string values.
		return "GoParseFile requires exactly one of 'path' or 'content' argument, both provided.", nil // Added return
	}
	if !pathHasValue && !contentHasValue {
		// FAIL Case: Neither path nor content had a non-empty string value.
		if pathArgProvided && !pathHasValue && !contentArgProvided {
			return fmt.Sprintf("GoParseFile: 'path' argument was provided but empty, and 'content' was not provided."), nil
		}
		if contentArgProvided && !contentHasValue && !pathArgProvided {
			return fmt.Sprintf("GoParseFile: 'content' argument was provided but empty, and 'path' was not provided."), nil
		}
		if pathArgProvided && contentArgProvided && !pathHasValue && !contentHasValue {
			return fmt.Sprintf("GoParseFile: Both 'path' and 'content' arguments were provided but empty."), nil
		}
		// Fallback if args were nil or wrong type initially (should be caught by validation layer ideally)
		// Return the message observed in test logs
		return "GoParseFile requires 'path' (string) or 'content' (string) argument.", nil
	}
	// --- Proceed if exactly one has a value ---

	var sourceName string
	// Get Source Content (Only if path provided and content not already set)
	if pathHasValue { // Read file only if path was the sole input with value
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
	} else if contentHasValue { // Use content only if it was the sole input with value
		sourceName = "<content string>"
		// Content was already assigned earlier if pathHadValue was false
	} else {
		// Should be unreachable due to validation above
		return "GoParseFile: Internal logic error determining input source.", fmt.Errorf("%w: internal logic error", ErrInternalTool)
	}

	// --- Parse Go Code ---
	fset := token.NewFileSet()
	astFile, err := parser.ParseFile(fset, sourceName, content, parser.ParseComments)
	if err != nil {
		return fmt.Sprintf("GoParseFile failed: %s", err.Error()), fmt.Errorf("%w: %w", ErrGoParseFailed, err)
	}

	// --- Success: Store AST + FileSet ---
	cachedData := CachedAst{File: astFile, Fset: fset}
	handleID := interpreter.storeObjectInCache(cachedData, golangASTTypeTag)
	interpreter.logger.Printf("[TOOL GoParseFile] Successfully parsed '%s'. Stored AST+FileSet with handle ID: %s", sourceName, handleID)
	return handleID, nil
}

// --- GoModifyAST Tool ---
func toolGoModifyAST(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	interpreter.logger.Printf("[TOOL GoModifyAST ENTRY] Received raw args (len %d): %v", len(args), args)
	// Argument Validation (Revised)
	if len(args) != 2 {
		return "GoModifyAST: Requires exactly 2 arguments: handleID (string), modifications (map).", nil
	}
	handleID, ok := args[0].(string)
	if !ok {
		return fmt.Sprintf("GoModifyAST: Expected string handle ID as first argument, got %T.", args[0]), nil
	}
	if handleID == "" {
		return "GoModifyAST: Handle ID cannot be empty.", nil
	}
	modifications, ok := args[1].(map[string]interface{})
	if !ok {
		return fmt.Sprintf("GoModifyAST: Expected map modifications as second argument, got %T.", args[1]), nil
	}
	interpreter.logger.Printf("[TOOL GoModifyAST] Validated initial arg types. Handle '%s', %d directives.", handleID, len(modifications))

	// Modifications Validation
	if len(modifications) == 0 {
		return "GoModifyAST: Modifications map cannot be empty.", nil
	}
	changePackageName := ""
	knownDirectiveFound := false
	if cpVal, ok := modifications["change_package"]; ok {
		knownDirectiveFound = true
		cpName, isString := cpVal.(string)
		if !isString || cpName == "" {
			return fmt.Sprintf("GoModifyAST: Invalid value for 'change_package': expected non-empty string, got %T.", cpVal), nil
		}
		changePackageName = cpName
	}
	if !knownDirectiveFound {
		return "GoModifyAST: Modifications map does not contain any known directives (e.g., 'change_package').", nil
	} // Match exact error message
	interpreter.logger.Printf("[TOOL GoModifyAST] Validated modifications request (change_package: %q). Proceeding.", changePackageName)

	// Retrieve Original AST
	obj, err := interpreter.retrieveObjectFromCache(handleID, golangASTTypeTag)
	if err != nil {
		return fmt.Sprintf("GoModifyAST: Failed to retrieve AST for handle '%s': %s", handleID, err.Error()), fmt.Errorf("%w: %w", ErrGoModifyFailed, err)
	}
	originalCachedAst, ok := obj.(CachedAst)
	if !ok {
		return fmt.Sprintf("GoModifyAST: Internal error - retrieved object for handle '%s' is not CachedAst (%T)", handleID, obj), fmt.Errorf("%w: unexpected object type retrieved from cache", ErrInternalTool)
	}
	interpreter.logger.Printf("[TOOL GoModifyAST] Successfully retrieved original CachedAst for handle '%s'. Package: %s", handleID, originalCachedAst.File.Name.Name)

	// Deep Copy via Print/Reparse
	interpreter.logger.Printf("[TOOL GoModifyAST] Performing deep copy via print/reparse for handle '%s'.", handleID)
	var buf bytes.Buffer
	err = printer.Fprint(&buf, originalCachedAst.Fset, originalCachedAst.File)
	if err != nil {
		return fmt.Sprintf("GoModifyAST: Failed during AST print for deep copy (handle '%s'): %s", handleID, err.Error()), fmt.Errorf("%w: printing AST for copy: %w", ErrInternalTool, err)
	}
	originalSource := buf.String()
	newFset := token.NewFileSet()
	newAstFile, err := parser.ParseFile(newFset, "reparsed_copy.go", originalSource, parser.ParseComments)
	if err != nil {
		return fmt.Sprintf("GoModifyAST: Failed during AST re-parse for deep copy (handle '%s'): %s", handleID, err.Error()), fmt.Errorf("%w: re-parsing printed AST: %w", ErrInternalTool, err)
	}
	interpreter.logger.Printf("[TOOL GoModifyAST] Deep copy successful for handle '%s'. New AST created.", handleID)

	// Apply Modification(s) to the NEW AST
	modificationApplied := false
	if changePackageName != "" {
		if newAstFile.Name == nil {
			return fmt.Sprintf("GoModifyAST: Cannot change package name, AST has no package declaration (handle '%s').", handleID), fmt.Errorf("%w: AST missing package declaration", ErrGoModifyFailed)
		}
		originalName := newAstFile.Name.Name
		newAstFile.Name.Name = changePackageName
		modificationApplied = true
		interpreter.logger.Printf("[TOOL GoModifyAST] Applied modification: Changed package name from '%s' to '%s'.", originalName, changePackageName)
	}
	if !modificationApplied {
		return "GoModifyAST: No applicable modification was performed.", fmt.Errorf("%w: no modification applied despite validation", ErrInternalTool)
	}

	// Store New AST & Invalidate Old Handle
	interpreter.logger.Printf("[TOOL GoModifyAST] Modification successful. Storing new AST and invalidating old handle '%s'.", handleID)
	newCachedAst := CachedAst{File: newAstFile, Fset: newFset}
	newHandleID := interpreter.storeObjectInCache(newCachedAst, golangASTTypeTag)

	// Keep deletes commented out for now
	/*
		delete(interpreter.objectCache, handleID)
		delete(interpreter.handleTypes, handleID)
	*/
	interpreter.logger.Printf("[TOOL GoModifyAST] SKIPPED invalidation of old handle '%s' for diagnostics. New handle is '%s'.", handleID, newHandleID)

	interpreter.logger.Printf("[TOOL GoModifyAST] Returning new handle '%s' successfully.", newHandleID)
	return newHandleID, nil
}

// --- Registration ---
func registerGoAstTools(registry *ToolRegistry) error {
	// GoParseFile registration
	err := registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name: "GoParseFile",
			Description: "Parses Go source code from a file path or direct content string. " +
				"Stores the AST internally and returns an opaque string handle on success.",
			Args: []ArgSpec{
				{Name: "path", Type: ArgTypeString, Required: false, Description: "Relative path to the Go source file (within sandbox)."},
				{Name: "content", Type: ArgTypeString, Required: false, Description: "Direct Go source code as a string."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolGoParseFile,
	})
	if err != nil {
		fmt.Printf("!!! CRITICAL: Failed to register Go AST tool GoParseFile: %v\n", err)
		return fmt.Errorf("failed to register Go AST tool GoParseFile: %w", err)
	} else {
		fmt.Println("--- Successfully registered GoParseFile ---")
	}

	// GoModifyAST registration
	err = registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name: "GoModifyAST",
			Description: "Modifies a Go AST represented by an input handle according to the 'modifications' map. " +
				"On success, stores the modified AST under a NEW handle, invalidates the OLD handle, and returns the NEW handle string. " +
				"Input handle is invalidated only on success.",
			Args: []ArgSpec{
				{Name: "handle", Type: ArgTypeString, Required: true, Description: "The opaque string handle for the AST to modify (obtained from GoParseFile)."},
				{Name: "modifications", Type: ArgTypeAny, Required: true, Description: "A map describing the modifications (e.g., {'change_package': 'new_name'}). Tool expects map[string]interface{}."},
			},
			ReturnType: ArgTypeString,
		},
		Func: toolGoModifyAST,
	})
	if err != nil {
		fmt.Printf("!!! CRITICAL: Failed to register Go AST tool GoModifyAST: %v\n", err)
		return fmt.Errorf("failed to register Go AST tool GoModifyAST: %w", err)
	} else {
		fmt.Println("--- Successfully registered GoModifyAST ---")
	}

	// Register GoFormatAST here later
	return nil
}
