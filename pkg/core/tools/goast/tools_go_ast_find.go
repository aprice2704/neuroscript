// filename: pkg/core/tools_go_ast_find.go
// NEW FILE
// UPDATED: Use GetHandleValue
package goast

import (
	"errors"
	"fmt"
	"go/ast"
	// "golang.org/x/tools/go/ast/astutil" // Not needed for finding
)

// --- TOOL.GoFindIdentifiers Implementation ---

// toolGoFindIdentifiers finds occurrences of pkg_name.identifier in the AST.
func toolGoFindIdentifiers(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// 1. Validate Arguments
	if len(args) != 3 {
		// Should be caught by validation layer based on spec, but double-check
		return nil, fmt.Errorf("internal error: %w, expected 3 args (handle, pkg_name, identifier), got %d", ErrValidationArgCount, len(args))
	}
	handleID, okH := args[0].(string)
	pkgName, okP := args[1].(string)
	identifier, okI := args[2].(string)

	if !okH || !okP || !okI {
		// Should be caught by validation layer
		return nil, fmt.Errorf("internal error: %w, expected string args", ErrValidationTypeMismatch)
	}
	if handleID == "" {
		return nil, fmt.Errorf("handle ID cannot be empty: %w", ErrValidationRequiredArgNil) // Or custom error?
	}
	if pkgName == "" || identifier == "" {
		// Use specific defined error
		return nil, fmt.Errorf("pkg_name (%q) and identifier (%q) must be non-empty: %w", pkgName, identifier, ErrGoInvalidIdentifierFormat)
	}
	interpreter.logger.Printf("[TOOL GoFindIdentifiers] Args validated: handle=%s, pkg=%s, id=%s", handleID, pkgName, identifier)

	// 2. Retrieve AST from Cache
	// *** UPDATED CALL ***
	obj, err := interpreter.GetHandleValue(handleID, golangASTTypeTag)
	if err != nil {
		// Wrap error for context
		return nil, fmt.Errorf("failed to retrieve AST for handle '%s': %w", handleID, errors.Join(ErrGoModifyFailed, err)) // Reuse ErrGoModifyFailed? Or new Find error? Let's reuse for now.
	}
	// *** END UPDATE ***
	cachedAst, ok := obj.(CachedAst)
	if !ok || cachedAst.File == nil || cachedAst.Fset == nil {
		errInternal := fmt.Errorf("internal error - retrieved object for handle '%s' is invalid (%T)", handleID, obj)
		return nil, fmt.Errorf("%w: %w", ErrInternalTool, errInternal) // Use ErrInternalTool here
	}
	interpreter.logger.Printf("[TOOL GoFindIdentifiers] Retrieved AST for handle '%s'", handleID)

	// 3. Find Identifiers using ast.Inspect
	var foundPositions []map[string]interface{} // Store results as maps

	ast.Inspect(cachedAst.File, func(n ast.Node) bool {
		// Check if node is a Selector Expression (like pkg.Ident)
		selExpr, isSelExpr := n.(*ast.SelectorExpr)
		if !isSelExpr {
			return true // Continue traversal
		}

		// Check if the part before the dot (X) is an Identifier matching pkgName
		xIdent, isXIdent := selExpr.X.(*ast.Ident)
		if !isXIdent || xIdent.Name != pkgName {
			return true // Continue traversal
		}

		// Check if the part after the dot (Sel) matches the identifier
		if selExpr.Sel.Name == identifier {
			// Match found! Get position info.
			pos := cachedAst.Fset.Position(selExpr.Pos()) // Use position of the selector expression start

			positionMap := map[string]interface{}{
				// Filename might be absent if parsed from string
				"filename": pos.Filename,
				"line":     pos.Line,
				"column":   pos.Column,
				// Optionally add offset: "offset": pos.Offset,
			}
			foundPositions = append(foundPositions, positionMap)
			interpreter.logger.Printf("[TOOL GoFindIdentifiers] Found %s.%s at %s:%d:%d", pkgName, identifier, pos.Filename, pos.Line, pos.Column)
		}
		return true // Continue traversal
	})

	interpreter.logger.Printf("[TOOL GoFindIdentifiers] Found %d occurrences.", len(foundPositions))

	// 4. Return the list of positions (empty list if none found)
	return foundPositions, nil
}
