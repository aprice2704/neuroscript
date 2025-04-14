// filename: pkg/core/tools_go_ast_modify.go
// UPDATED: Return defined errors instead of strings for validation failures
package core

import (
	"bytes"
	"errors"
	"fmt"
	"go/parser"
	"go/printer"
	"go/token"

	// "strconv" // Not needed here anymore

	"golang.org/x/tools/go/ast/astutil" // Import astutil
)

// --- GoModifyAST Tool Implementation ---
func toolGoModifyAST(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	interpreter.logger.Printf("[TOOL GoModifyAST ENTRY] Received raw args (len %d): %v", len(args), args)
	// Argument Validation (by ValidateAndConvertArgs) ensures args[0] is string, args[1] is interface{}

	handleID := args[0].(string) // Safe assertion after validation
	modsArg := args[1]           // Keep as interface{} for now

	// --- UPDATED: Internal Type Check & Validation ---
	// Check modifications type explicitly here, as ValidateAndConvertArgs doesn't check ArgTypeAny
	modifications, ok := modsArg.(map[string]interface{})
	if !ok {
		// Return a defined error type
		return nil, fmt.Errorf("expected map modifications as second argument, got %T: %w", modsArg, ErrValidationTypeMismatch)
	}
	interpreter.logger.Printf("[TOOL GoModifyAST] Validated initial arg types. Handle '%s', %d directives.", handleID, len(modifications))

	// Modifications Validation & Parsing
	if len(modifications) == 0 {
		// Return defined error
		return nil, ErrGoModifyEmptyMap
	}

	var changePackageName string
	var addImportPath string
	var removeImportPath string
	var replaceImportOldPath string
	var replaceImportNewPath string
	knownDirectiveFound := false

	// Parse change_package
	if cpVal, ok := modifications["change_package"]; ok {
		knownDirectiveFound = true
		cpName, isString := cpVal.(string)
		if !isString || cpName == "" {
			// Return defined error with context
			return nil, fmt.Errorf("invalid value for 'change_package': expected non-empty string, got %T: %w", cpVal, ErrGoModifyInvalidDirectiveValue)
		}
		changePackageName = cpName
	}
	// Parse add_import
	if aiVal, ok := modifications["add_import"]; ok {
		knownDirectiveFound = true
		aiPath, isString := aiVal.(string)
		if !isString || aiPath == "" {
			// Return defined error with context
			return nil, fmt.Errorf("invalid value for 'add_import': expected non-empty string import path, got %T: %w", aiVal, ErrGoModifyInvalidDirectiveValue)
		}
		addImportPath = aiPath
	}
	// Parse remove_import
	if riVal, ok := modifications["remove_import"]; ok {
		knownDirectiveFound = true
		riPath, isString := riVal.(string)
		if !isString || riPath == "" {
			// Return defined error with context
			return nil, fmt.Errorf("invalid value for 'remove_import': expected non-empty string import path, got %T: %w", riVal, ErrGoModifyInvalidDirectiveValue)
		}
		removeImportPath = riPath
	}
	// Parse replace_import
	if repiVal, ok := modifications["replace_import"]; ok {
		knownDirectiveFound = true
		repiMap, isMap := repiVal.(map[string]interface{})
		if !isMap {
			// Return defined error with context
			return nil, fmt.Errorf("invalid value for 'replace_import': expected map, got %T: %w", repiVal, ErrGoModifyInvalidDirectiveValue)
		}
		oldPathVal, okOld := repiMap["old_path"]
		newPathVal, okNew := repiMap["new_path"]
		if !okOld || !okNew {
			// Return defined error
			return nil, fmt.Errorf("'replace_import' map requires 'old_path' and 'new_path' keys: %w", ErrGoModifyMissingMapKey)
		}
		oldPath, isStringOld := oldPathVal.(string)
		newPath, isStringNew := newPathVal.(string)
		if !isStringOld || oldPath == "" || !isStringNew || newPath == "" {
			// Return defined error with context
			return nil, fmt.Errorf("invalid values in 'replace_import' map: both 'old_path' (%T) and 'new_path' (%T) must be non-empty strings: %w", oldPathVal, newPathVal, ErrGoModifyInvalidDirectiveValue)
		}
		replaceImportOldPath = oldPath
		replaceImportNewPath = newPath
	}
	// Add validation for other directives here

	if !knownDirectiveFound {
		// Return defined error
		return nil, ErrGoModifyUnknownDirective
	}
	interpreter.logger.Printf("[TOOL GoModifyAST] Parsed directives: changePkg=%q, add=%q, remove=%q, replaceOld=%q, replaceNew=%q", changePackageName, addImportPath, removeImportPath, replaceImportOldPath, replaceImportNewPath)
	// --- END UPDATED Validation ---

	// Retrieve Original AST
	obj, err := interpreter.retrieveObjectFromCache(handleID, golangASTTypeTag)
	if err != nil {
		// Wrap the underlying error for context, keep ErrGoModifyFailed primary
		return nil, fmt.Errorf("failed to retrieve AST for handle '%s': %w", handleID, errors.Join(ErrGoModifyFailed, err)) // Use Join if Go 1.20+
	}
	originalCachedAst, ok := obj.(CachedAst)
	if !ok {
		// This indicates an internal problem, maybe wrap ErrInternalTool?
		errInternal := fmt.Errorf("internal error - retrieved object for handle '%s' is not CachedAst (%T)", handleID, obj)
		return nil, fmt.Errorf("%w: %w", ErrGoModifyFailed, errInternal) // Still primarily a modification failure
	}
	interpreter.logger.Printf("[TOOL GoModifyAST] Successfully retrieved original CachedAst for handle '%s'. Package: %s", handleID, originalCachedAst.File.Name.Name)

	// Deep Copy via Print/Reparse
	interpreter.logger.Printf("[TOOL GoModifyAST] Performing deep copy via print/reparse for handle '%s'.", handleID)
	var buf bytes.Buffer
	printCfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	err = printCfg.Fprint(&buf, originalCachedAst.Fset, originalCachedAst.File)
	if err != nil {
		// Wrap internal error
		errInternal := fmt.Errorf("printing AST for copy: %w", err)
		return nil, fmt.Errorf("failed during AST print for deep copy (handle '%s'): %w", handleID, errors.Join(ErrGoModifyFailed, errInternal))
	}
	originalSource := buf.String()
	newFset := token.NewFileSet()
	newAstFile, err := parser.ParseFile(newFset, "reparsed_copy.go", originalSource, parser.ParseComments)
	if err != nil {
		// Wrap internal error
		errInternal := fmt.Errorf("re-parsing printed AST: %w", err)
		return nil, fmt.Errorf("failed during AST re-parse for deep copy (handle '%s'): %w", handleID, errors.Join(ErrGoModifyFailed, errInternal))
	}
	interpreter.logger.Printf("[TOOL GoModifyAST] Deep copy successful for handle '%s'. New AST created.", handleID)

	// Apply Modification(s) to the NEW AST
	modificationApplied := false // Was a modification directive present?
	actionTaken := false         // Did astutil actually change the AST?

	// 1. Change Package Name
	if changePackageName != "" {
		modificationApplied = true
		if newAstFile.Name == nil {
			// Return defined error? Or a specific ErrGoModifyFailed sub-type?
			return nil, fmt.Errorf("cannot change package name, AST has no package declaration (handle '%s'): %w", handleID, ErrGoModifyFailed)
		}
		originalName := newAstFile.Name.Name
		if originalName != changePackageName {
			newAstFile.Name.Name = changePackageName
			interpreter.logger.Printf("[TOOL GoModifyAST] Applied modification: Changed package name from '%s' to '%s'.", originalName, changePackageName)
			actionTaken = true
		} else {
			interpreter.logger.Printf("[TOOL GoModifyAST] Info: Package name already '%s', no action taken for 'change_package'.", changePackageName)
		}
	}
	// 2. Add Import
	if addImportPath != "" {
		modificationApplied = true
		if added := astutil.AddImport(newFset, newAstFile, addImportPath); added {
			interpreter.logger.Printf("[TOOL GoModifyAST] Applied modification: Added import '%s'.", addImportPath)
			actionTaken = true
		} else {
			interpreter.logger.Printf("[TOOL GoModifyAST] Info: Import '%s' already exists, no action taken for 'add_import'.", addImportPath)
		}
	}
	// 3. Remove Import
	if removeImportPath != "" {
		modificationApplied = true
		if deleted := astutil.DeleteImport(newFset, newAstFile, removeImportPath); deleted {
			interpreter.logger.Printf("[TOOL GoModifyAST] Applied modification: Removed import '%s'.", removeImportPath)
			actionTaken = true
		} else {
			interpreter.logger.Printf("[TOOL GoModifyAST] Info: Import '%s' not found, no action taken for 'remove_import'.", removeImportPath)
		}
	}
	// 4. Replace Import
	if replaceImportOldPath != "" && replaceImportNewPath != "" {
		modificationApplied = true
		if rewritten := astutil.RewriteImport(newFset, newAstFile, replaceImportOldPath, replaceImportNewPath); rewritten {
			interpreter.logger.Printf("[TOOL GoModifyAST] Applied modification: Replaced import '%s' with '%s'.", replaceImportOldPath, replaceImportNewPath)
			actionTaken = true
		} else {
			interpreter.logger.Printf("[TOOL GoModifyAST] Info: Import '%s' not found or already matches '%s', no action taken for 'replace_import'.", replaceImportOldPath, replaceImportNewPath)
		}
	}
	// Add logic for other modifications here

	if !modificationApplied {
		// Should be unreachable due to earlier check, return internal error
		return nil, fmt.Errorf("%w: no known directive provided despite passing initial check", ErrInternalTool)
	}
	if !actionTaken {
		interpreter.logger.Printf("[TOOL GoModifyAST] Warning: Modification(s) requested, but no action taken. Returning original handle.")
		return handleID, nil // Return original handle if no effective change occurred
	}

	// Store New AST & Invalidate Old Handle
	interpreter.logger.Printf("[TOOL GoModifyAST] Modification successful. Storing new AST.")
	newCachedAst := CachedAst{File: newAstFile, Fset: newFset}
	newHandleID := interpreter.storeObjectInCache(newCachedAst, golangASTTypeTag)
	// Invalidation skipped for diagnostics
	interpreter.logger.Printf("[TOOL GoModifyAST] SKIPPED invalidation of old handle '%s'. New handle is '%s'.", handleID, newHandleID)
	interpreter.logger.Printf("[TOOL GoModifyAST] Returning new handle '%s' successfully.", newHandleID)
	return newHandleID, nil
}
