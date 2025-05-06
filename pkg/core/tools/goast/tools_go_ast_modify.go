// NeuroScript Version: 0.3.1
// File version: 0.0.2
// Fix undefined error variable core.ErrOperationFailed.
// filename: pkg/core/tools/goast/tools_go_ast_modify.go

package goast

import (
	"bytes"
	"fmt"
	"go/ast" // Keep ast import
	"go/parser"
	"go/printer"
	"go/token"
	"strings" // Keep strings import

	"github.com/aprice2704/neuroscript/pkg/core"
	"golang.org/x/tools/go/ast/astutil" // Import astutil
)

// Define ToolImplementation
var toolGoModifyASTImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name: "GoModifyAST",
		Description: "Modifies Go AST (referenced by handle) using directives like change_package, add/remove/replace_import, replace_identifier. " +
			"Returns a NEW handle referencing the modified AST on success.",
		Args: []core.ArgSpec{
			{Name: "handle", Type: core.ArgTypeString, Required: true, Description: "Handle for the original AST (from GoParseFile)."},
			{Name: "modifications", Type: core.ArgTypeMap, Required: true, Description: "Map describing modifications (e.g., {'change_package': 'newname', 'add_import': 'path/to/pkg', 'replace_identifier': {'old':'pkg.Old', 'new':'pkg.New'}})."},
		},
		ReturnType: core.ArgTypeString, // Returns NEW Handle ID
	},
	Func: toolGoModifyAST,
}

// --- GoModifyAST Tool Implementation ---
func toolGoModifyAST(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	logger := interpreter.Logger()
	logger.Debug("Tool: GoModifyAST ENTRY] Received raw args (len %d): %v", len(args), args)

	handleID := args[0].(string)
	modsArg := args[1]

	modifications, ok := modsArg.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%w: expected map modifications as second argument, got %T", core.ErrInvalidArgument, modsArg)
	}
	logger.Debug("Tool: GoModifyAST] Validated initial arg types. Handle '%s', %d directives.", handleID, len(modifications))

	if len(modifications) == 0 {
		return nil, fmt.Errorf("%w: modifications map cannot be empty", core.ErrInvalidArgument)
	}

	var changePackageName string
	var addImportPath string
	var removeImportPath string
	var replaceImportOldPath string
	var replaceImportNewPath string
	var replaceIdentOldPkg string
	var replaceIdentOldID string
	var replaceIdentNewPkg string
	var replaceIdentNewID string
	knownDirectiveFound := false

	// --- Parse Directives (Error handling uses core.ErrInvalidArgument) ---
	if cpVal, ok := modifications["change_package"]; ok {
		knownDirectiveFound = true
		cpName, isString := cpVal.(string)
		if !isString || cpName == "" {
			return nil, fmt.Errorf("%w: invalid value for 'change_package': expected non-empty string, got %T", core.ErrInvalidArgument, cpVal)
		}
		changePackageName = cpName
	}
	if aiVal, ok := modifications["add_import"]; ok {
		knownDirectiveFound = true
		aiPath, isString := aiVal.(string)
		if !isString || aiPath == "" {
			return nil, fmt.Errorf("%w: invalid value for 'add_import': expected non-empty string import path, got %T", core.ErrInvalidArgument, aiVal)
		}
		addImportPath = aiPath
	}
	if riVal, ok := modifications["remove_import"]; ok {
		knownDirectiveFound = true
		riPath, isString := riVal.(string)
		if !isString || riPath == "" {
			return nil, fmt.Errorf("%w: invalid value for 'remove_import': expected non-empty string import path, got %T", core.ErrInvalidArgument, riVal)
		}
		removeImportPath = riPath
	}
	if repiVal, ok := modifications["replace_import"]; ok {
		knownDirectiveFound = true
		repiMap, isMap := repiVal.(map[string]interface{})
		if !isMap {
			return nil, fmt.Errorf("%w: invalid value for 'replace_import': expected map, got %T", core.ErrInvalidArgument, repiVal)
		}
		oldPathVal, okOld := repiMap["old_path"]
		newPathVal, okNew := repiMap["new_path"]
		if !okOld || !okNew {
			return nil, fmt.Errorf("%w: 'replace_import' map requires 'old_path' and 'new_path' keys", core.ErrInvalidArgument)
		}
		oldPath, isStringOld := oldPathVal.(string)
		newPath, isStringNew := newPathVal.(string)
		if !isStringOld || oldPath == "" || !isStringNew || newPath == "" {
			return nil, fmt.Errorf("%w: invalid values in 'replace_import' map: both 'old_path' (%T) and 'new_path' (%T) must be non-empty strings", core.ErrInvalidArgument, oldPathVal, newPathVal)
		}
		replaceImportOldPath = oldPath
		replaceImportNewPath = newPath
	}
	if repidVal, ok := modifications["replace_identifier"]; ok {
		knownDirectiveFound = true
		repidMap, isMap := repidVal.(map[string]interface{})
		if !isMap {
			return nil, fmt.Errorf("%w: invalid value for 'replace_identifier': expected map, got %T", core.ErrInvalidArgument, repidVal)
		}
		oldFullIDVal, okOld := repidMap["old"]
		newFullIDVal, okNew := repidMap["new"]
		if !okOld || !okNew {
			return nil, fmt.Errorf("%w: 'replace_identifier' map requires 'old' ('pkg.Symbol') and 'new' ('pkg.Symbol') keys", core.ErrInvalidArgument)
		}
		oldFullID, isStringOld := oldFullIDVal.(string)
		newFullID, isStringNew := newFullIDVal.(string)
		if !isStringOld || !isStringNew {
			return nil, fmt.Errorf("%w: invalid values in 'replace_identifier' map: 'old' (%T) and 'new' (%T) must be strings", core.ErrInvalidArgument, oldFullIDVal, newFullIDVal)
		}
		oldParts := strings.SplitN(oldFullID, ".", 2)
		newParts := strings.SplitN(newFullID, ".", 2)
		if len(oldParts) != 2 || oldParts[0] == "" || oldParts[1] == "" || len(newParts) != 2 || newParts[0] == "" || newParts[1] == "" {
			return nil, fmt.Errorf("%w: invalid format in 'replace_identifier' map: 'old' (%q) and 'new' (%q) must be in 'package.Symbol' format with non-empty parts", core.ErrInvalidArgument, oldFullID, newFullID)
		}
		replaceIdentOldPkg = oldParts[0]
		replaceIdentOldID = oldParts[1]
		replaceIdentNewPkg = newParts[0]
		replaceIdentNewID = newParts[1]
	}
	// --- End Directive Parsing ---

	if !knownDirectiveFound {
		return nil, fmt.Errorf("%w: no known modification directive found in map (e.g., 'change_package', 'add_import', etc.)", core.ErrInvalidArgument)
	}
	logger.Debug("Tool: GoModifyAST] Parsed directives: changePkg=%q, add=%q, remove=%q, replaceOldImp=%q, replaceNewImp=%q, replaceIdentOld=%s.%s, replaceIdentNew=%s.%s",
		changePackageName, addImportPath, removeImportPath, replaceImportOldPath, replaceImportNewPath, replaceIdentOldPkg, replaceIdentOldID, replaceIdentNewPkg, replaceIdentNewID)

	// Retrieve Original AST
	obj, err := interpreter.GetHandleValue(handleID, golangASTTypeTag)
	if err != nil {
		return nil, fmt.Errorf("GoModifyAST failed to retrieve AST for handle '%s': %w", handleID, err) // Wrapped error from GetHandleValue
	}
	originalCachedAst, ok := obj.(CachedAst)
	if !ok || originalCachedAst.File == nil || originalCachedAst.Fset == nil {
		return nil, fmt.Errorf("%w: GoModifyAST retrieved invalid object for handle '%s' (%T)", core.ErrInternalTool, handleID, obj)
	}
	logger.Debug("Tool: GoModifyAST] Successfully retrieved original CachedAst for handle '%s'. Package: %s", handleID, originalCachedAst.File.Name.Name)

	// Deep Copy via Print/Reparse
	logger.Debug("Tool: GoModifyAST] Performing deep copy via print/reparse for handle '%s'.", handleID)
	var buf bytes.Buffer
	printCfg := printer.Config{Mode: printer.UseSpaces | printer.TabIndent, Tabwidth: 8}
	err = printCfg.Fprint(&buf, originalCachedAst.Fset, originalCachedAst.File)
	if err != nil {
		// Use GoModifyFailed for errors during the modification process
		return nil, fmt.Errorf("%w: printing AST for copy (handle '%s'): %w", core.ErrGoModifyFailed, handleID, err)
	}
	originalSource := buf.String()
	newFset := token.NewFileSet()
	// Preserve original filename if available in the fileset
	originalFilename := "<content string>" // Default if not found
	if fileToken := originalCachedAst.Fset.File(originalCachedAst.File.Pos()); fileToken != nil {
		originalFilename = fileToken.Name()
	}
	newAstFile, err := parser.ParseFile(newFset, originalFilename, originalSource, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("%w: re-parsing printed AST for copy (handle '%s'): %w", core.ErrGoModifyFailed, handleID, err)
	}
	logger.Debug("Tool: GoModifyAST] Deep copy successful for handle '%s'. New AST created.", handleID)

	// Apply Modification(s) to the NEW AST
	modificationDirectivePresent := knownDirectiveFound // If we parsed a known directive, assume modification was intended
	actionTaken := false                                // Did astutil or direct manipulation actually change the AST?

	// 1. Change Package Name
	if changePackageName != "" {
		if newAstFile.Name == nil {
			// Use GoModifyFailed as the operation cannot proceed
			return nil, fmt.Errorf("%w: cannot change package name, AST has no package declaration (handle '%s')", core.ErrGoModifyFailed, handleID)
		}
		if originalCachedAst.File.Name.Name != changePackageName { // Compare with original name from copied AST
			newAstFile.Name.Name = changePackageName
			logger.Debug("Tool: GoModifyAST] Applied modification: Changed package name from '%s' to '%s'.", originalCachedAst.File.Name.Name, changePackageName)
			actionTaken = true
		} else {
			logger.Info("Tool: GoModifyAST] Info: Package name already '%s', no action needed for 'change_package'.", changePackageName)
		}
	}
	// 2. Add Import
	if addImportPath != "" {
		if added := astutil.AddImport(newFset, newAstFile, addImportPath); added {
			logger.Debug("Tool: GoModifyAST] Applied modification: Added import '%s'.", addImportPath)
			actionTaken = true
		} else {
			logger.Info("Tool: GoModifyAST] Info: Import '%s' already exists, no action needed for 'add_import'.", addImportPath)
		}
	}
	// 3. Remove Import
	if removeImportPath != "" {
		if deleted := astutil.DeleteImport(newFset, newAstFile, removeImportPath); deleted {
			logger.Debug("Tool: GoModifyAST] Applied modification: Removed import '%s'.", removeImportPath)
			actionTaken = true
		} else {
			logger.Info("Tool: GoModifyAST] Info: Import '%s' not found, no action needed for 'remove_import'.", removeImportPath)
		}
	}
	// 4. Replace Import
	if replaceImportOldPath != "" && replaceImportNewPath != "" {
		if rewritten := astutil.RewriteImport(newFset, newAstFile, replaceImportOldPath, replaceImportNewPath); rewritten {
			logger.Debug("Tool: GoModifyAST] Applied modification: Replaced import '%s' with '%s'.", replaceImportOldPath, replaceImportNewPath)
			actionTaken = true
		} else {
			logger.Info("Tool: GoModifyAST] Info: Import '%s' not found or already matches '%s', no action needed for 'replace_import'.", replaceImportOldPath, replaceImportNewPath)
		}
	}
	// 5. Replace Identifier
	if replaceIdentOldPkg != "" && replaceIdentOldID != "" && replaceIdentNewPkg != "" && replaceIdentNewID != "" {
		localActionTaken := false
		logger.Debug("Tool: GoModifyAST] Applying 'replace_identifier': %s.%s -> %s.%s", replaceIdentOldPkg, replaceIdentOldID, replaceIdentNewPkg, replaceIdentNewID)
		postVisit := func(cursor *astutil.Cursor) bool {
			node := cursor.Node()
			selExpr, ok := node.(*ast.SelectorExpr)
			if !ok {
				return true
			}
			xIdent, okX := selExpr.X.(*ast.Ident)
			if !okX {
				return true
			}
			if xIdent.Name == replaceIdentOldPkg && selExpr.Sel.Name == replaceIdentOldID {
				logger.Debug("Tool: GoModifyAST ReplaceIdent] Found match at line %d: Replacing %s.%s", newFset.Position(selExpr.Pos()).Line, xIdent.Name, selExpr.Sel.Name)
				newX := ast.NewIdent(replaceIdentNewPkg)
				newSel := ast.NewIdent(replaceIdentNewID)
				cursor.Replace(&ast.SelectorExpr{X: newX, Sel: newSel})
				localActionTaken = true
				logger.Debug("Tool: GoModifyAST ReplaceIdent] Replaced with %s.%s", newX.Name, newSel.Name)
			}
			return true
		}
		astutil.Apply(newAstFile, nil, postVisit)
		actionTaken = actionTaken || localActionTaken
		if !localActionTaken {
			logger.Info("Tool: GoModifyAST] Info: No occurrences of %s.%s found, no action taken for 'replace_identifier'.", replaceIdentOldPkg, replaceIdentOldID)
		}
	}

	// --- Final Checks ---
	if !modificationDirectivePresent { // Should have been caught earlier by knownDirectiveFound check
		return nil, fmt.Errorf("%w: no known directive provided but directive parsing passed", core.ErrInternalTool)
	}
	if !actionTaken {
		logger.Info("Tool: GoModifyAST] Warning: Modification(s) requested, but no changes made to the AST (e.g., target not found, or already correct). Returning original handle.")
		return handleID, nil
	}

	// Store New AST & Return New Handle
	logger.Debug("Tool: GoModifyAST] Modification successful. Storing new AST.")
	newCachedAst := CachedAst{File: newAstFile, Fset: newFset}
	newHandleID, err := interpreter.RegisterHandle(newCachedAst, golangASTTypeTag)
	if err != nil {
		logger.Error("Failed to register modified AST handle for original handle '%s': %v", handleID, err)
		return nil, fmt.Errorf("%w: failed to register modified AST handle", core.ErrInternalTool)
	}

	logger.Info("Tool: GoModifyAST] Returning new handle '%s' successfully.", newHandleID)
	return newHandleID, nil
}
