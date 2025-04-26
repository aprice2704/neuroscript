// filename: pkg/core/tools_go_ast_modify.go
// UPDATED: Add replace_identifier directive handling
// UPDATED: Use RegisterHandle and GetHandleValue
package goast

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast" // Keep ast import
	"go/parser"
	"go/printer"
	"go/token"
	"strings" // Keep strings import

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
	// --- NEW: Variables for replace_identifier ---
	var replaceIdentOldPkg string
	var replaceIdentOldID string
	var replaceIdentNewPkg string
	var replaceIdentNewID string
	// --- END NEW ---
	knownDirectiveFound := false

	// Parse change_package
	if cpVal, ok := modifications["change_package"]; ok {
		knownDirectiveFound = true
		cpName, isString := cpVal.(string)
		if !isString || cpName == "" {
			return nil, fmt.Errorf("invalid value for 'change_package': expected non-empty string, got %T: %w", cpVal, ErrGoModifyInvalidDirectiveValue)
		}
		changePackageName = cpName
	}
	// Parse add_import
	if aiVal, ok := modifications["add_import"]; ok {
		knownDirectiveFound = true
		aiPath, isString := aiVal.(string)
		if !isString || aiPath == "" {
			return nil, fmt.Errorf("invalid value for 'add_import': expected non-empty string import path, got %T: %w", aiVal, ErrGoModifyInvalidDirectiveValue)
		}
		addImportPath = aiPath
	}
	// Parse remove_import
	if riVal, ok := modifications["remove_import"]; ok {
		knownDirectiveFound = true
		riPath, isString := riVal.(string)
		if !isString || riPath == "" {
			return nil, fmt.Errorf("invalid value for 'remove_import': expected non-empty string import path, got %T: %w", riVal, ErrGoModifyInvalidDirectiveValue)
		}
		removeImportPath = riPath
	}
	// Parse replace_import
	if repiVal, ok := modifications["replace_import"]; ok {
		knownDirectiveFound = true
		repiMap, isMap := repiVal.(map[string]interface{})
		if !isMap {
			return nil, fmt.Errorf("invalid value for 'replace_import': expected map, got %T: %w", repiVal, ErrGoModifyInvalidDirectiveValue)
		}
		oldPathVal, okOld := repiMap["old_path"]
		newPathVal, okNew := repiMap["new_path"]
		if !okOld || !okNew {
			return nil, fmt.Errorf("'replace_import' map requires 'old_path' and 'new_path' keys: %w", ErrGoModifyMissingMapKey)
		}
		oldPath, isStringOld := oldPathVal.(string)
		newPath, isStringNew := newPathVal.(string)
		if !isStringOld || oldPath == "" || !isStringNew || newPath == "" {
			return nil, fmt.Errorf("invalid values in 'replace_import' map: both 'old_path' (%T) and 'new_path' (%T) must be non-empty strings: %w", oldPathVal, newPathVal, ErrGoModifyInvalidDirectiveValue)
		}
		replaceImportOldPath = oldPath
		replaceImportNewPath = newPath
	}
	// --- NEW: Parse replace_identifier ---
	if repidVal, ok := modifications["replace_identifier"]; ok {
		knownDirectiveFound = true
		repidMap, isMap := repidVal.(map[string]interface{})
		if !isMap {
			return nil, fmt.Errorf("invalid value for 'replace_identifier': expected map, got %T: %w", repidVal, ErrGoModifyInvalidDirectiveValue)
		}
		oldFullIDVal, okOld := repidMap["old"]
		newFullIDVal, okNew := repidMap["new"]
		if !okOld || !okNew {
			return nil, fmt.Errorf("'replace_identifier' map requires 'old' ('pkg.Symbol') and 'new' ('pkg.Symbol') keys: %w", ErrGoModifyMissingMapKey)
		}
		oldFullID, isStringOld := oldFullIDVal.(string)
		newFullID, isStringNew := newFullIDVal.(string)
		if !isStringOld || !isStringNew {
			return nil, fmt.Errorf("invalid values in 'replace_identifier' map: 'old' (%T) and 'new' (%T) must be strings: %w", oldFullIDVal, newFullIDVal, ErrGoModifyInvalidDirectiveValue)
		}

		// Parse old and new identifiers
		oldParts := strings.SplitN(oldFullID, ".", 2)
		newParts := strings.SplitN(newFullID, ".", 2)
		if len(oldParts) != 2 || oldParts[0] == "" || oldParts[1] == "" || len(newParts) != 2 || newParts[0] == "" || newParts[1] == "" {
			return nil, fmt.Errorf("invalid format in 'replace_identifier' map: 'old' (%q) and 'new' (%q) must be in 'package.Symbol' format with non-empty parts: %w", oldFullID, newFullID, ErrGoInvalidIdentifierFormat)
		}
		replaceIdentOldPkg = oldParts[0]
		replaceIdentOldID = oldParts[1]
		replaceIdentNewPkg = newParts[0]
		replaceIdentNewID = newParts[1]
	}
	// --- END NEW PARSE ---

	// Add validation for other directives here

	if !knownDirectiveFound {
		return nil, ErrGoModifyUnknownDirective
	}
	interpreter.logger.Printf("[TOOL GoModifyAST] Parsed directives: changePkg=%q, add=%q, remove=%q, replaceOldImp=%q, replaceNewImp=%q, replaceIdentOld=%s.%s, replaceIdentNew=%s.%s",
		changePackageName, addImportPath, removeImportPath, replaceImportOldPath, replaceImportNewPath, replaceIdentOldPkg, replaceIdentOldID, replaceIdentNewPkg, replaceIdentNewID) // Updated log
	// --- END UPDATED Validation ---

	// Retrieve Original AST
	// *** UPDATED CALL ***
	obj, err := interpreter.GetHandleValue(handleID, golangASTTypeTag)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve AST for handle '%s': %w", handleID, errors.Join(ErrGoModifyFailed, err))
	}
	// *** END UPDATE ***
	originalCachedAst, ok := obj.(CachedAst)
	if !ok {
		errInternal := fmt.Errorf("internal error - retrieved object for handle '%s' is not CachedAst (%T)", handleID, obj)
		return nil, fmt.Errorf("%w: %w", ErrGoModifyFailed, errInternal)
	}
	interpreter.logger.Printf("[TOOL GoModifyAST] Successfully retrieved original CachedAst for handle '%s'. Package: %s", handleID, originalCachedAst.File.Name.Name)

	// Deep Copy via Print/Reparse
	interpreter.logger.Printf("[TOOL GoModifyAST] Performing deep copy via print/reparse for handle '%s'.", handleID)
	var buf bytes.Buffer
	printCfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	err = printCfg.Fprint(&buf, originalCachedAst.Fset, originalCachedAst.File)
	if err != nil {
		errInternal := fmt.Errorf("printing AST for copy: %w", err)
		return nil, fmt.Errorf("failed during AST print for deep copy (handle '%s'): %w", handleID, errors.Join(ErrGoModifyFailed, errInternal))
	}
	originalSource := buf.String()
	newFset := token.NewFileSet()
	newAstFile, err := parser.ParseFile(newFset, "reparsed_copy.go", originalSource, parser.ParseComments)
	if err != nil {
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
	// --- NEW: 5. Replace Identifier ---
	if replaceIdentOldPkg != "" && replaceIdentOldID != "" && replaceIdentNewPkg != "" && replaceIdentNewID != "" {
		modificationApplied = true
		interpreter.logger.Printf("[TOOL GoModifyAST] Applying 'replace_identifier': %s.%s -> %s.%s", replaceIdentOldPkg, replaceIdentOldID, replaceIdentNewPkg, replaceIdentNewID)

		// Define the post-visitor function for astutil.Apply
		postVisit := func(cursor *astutil.Cursor) bool {
			node := cursor.Node()
			selExpr, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true // Continue traversal if not a selector expression
			}

			// Check if X is an identifier (package name or alias)
			xIdent, okX := selExpr.X.(*ast.Ident)
			if !okX {
				return true // Continue if X is not a simple identifier
			}

			// Check if the selector matches the old pkg.ID
			if xIdent.Name == replaceIdentOldPkg && selExpr.Sel.Name == replaceIdentOldID {
				interpreter.logger.Printf("[TOOL GoModifyAST ReplaceIdent] Found match at line %d: Replacing %s.%s", newFset.Position(selExpr.Pos()).Line, xIdent.Name, selExpr.Sel.Name)
				// Create new identifiers for replacement
				newX := ast.NewIdent(replaceIdentNewPkg)
				newSel := ast.NewIdent(replaceIdentNewID)
				// Replace nodes in the cursor (this modifies the AST)
				cursor.Replace(&ast.SelectorExpr{X: newX, Sel: newSel})
				actionTaken = true // Record that a change was made
				interpreter.logger.Printf("[TOOL GoModifyAST ReplaceIdent] Replaced with %s.%s", newX.Name, newSel.Name)
			}
			return true // Continue traversal
		}

		// Apply the visitor
		astutil.Apply(newAstFile, nil, postVisit)
		// Check actionTaken *after* applying the visitor
		if !actionTaken && modificationApplied { // Ensure we only log 'not found' if this was the only mod requested or others failed
			interpreter.logger.Printf("[TOOL GoModifyAST] Info: No occurrences of %s.%s found, no action taken for 'replace_identifier'.", replaceIdentOldPkg, replaceIdentOldID)
		}
	}
	// --- END NEW REPLACE IDENTIFIER ---
	// Add logic for other modifications here

	if !modificationApplied {
		// This should theoretically not be reached if initial validation passed
		return nil, fmt.Errorf("%w: no known directive provided despite passing initial check", ErrInternalTool)
	}
	if !actionTaken {
		interpreter.logger.Printf("[TOOL GoModifyAST] Warning: Modification(s) requested, but no action taken (e.g., target not found, or already correct). Returning original handle.")
		return handleID, nil // Return original handle if no changes made
	}

	// Store New AST & Return New Handle
	interpreter.logger.Printf("[TOOL GoModifyAST] Modification successful. Storing new AST.")
	newCachedAst := CachedAst{File: newAstFile, Fset: newFset}
	// *** UPDATED CALL ***
	newHandleID, err := interpreter.RegisterHandle(newCachedAst, golangASTTypeTag)
	if err != nil {
		interpreter.logger.Printf("[ERROR] Failed to register modified AST handle for original handle '%s': %v", handleID, err)
		return nil, fmt.Errorf("failed to register modified AST handle: %w", err) // Return internal error
	}
	// *** END UPDATE ***

	// Invalidation skipped for diagnostics - Consider adding RemoveHandle(handleID) here in production
	interpreter.logger.Printf("[TOOL GoModifyAST] SKIPPED invalidation of old handle '%s'. New handle is '%s'.", handleID, newHandleID)
	interpreter.logger.Printf("[TOOL GoModifyAST] Returning new handle '%s' successfully.", newHandleID)
	return newHandleID, nil
}
