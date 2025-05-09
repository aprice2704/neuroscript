// NeuroScript Version: 0.3.1
// File version: 0.3.4 // Moved from core, changed package, adjusted imports and helper calls.
// Contains GoFindDeclarations tool (Line/Column based).
// filename: pkg/core/tools/gosemantic/find_declarations_lc.go

// CAUTION: NOT RELIABLE
// It turns out this is very hard to do well, so we moved to semantic addressing instead
// Use with the expectation that it is APPROXIMATE

package gosemantic

import (
	"fmt"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // Import core
)

// --- Tool: GoFindDeclarations ---

func toolGoFindDeclarations(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	logger := interpreter.Logger()

	// --- Argument Parsing and Validation ---
	if len(args) != 4 {
		return nil, fmt.Errorf("%w: GoFindDeclarations requires 4 arguments (index_handle, path, line, column)", core.ErrInvalidArgument)
	}
	handle, okH := args[0].(string)
	pathRel, okP := args[1].(string)
	lineRaw, okL := args[2].(int64)
	colRaw, okC := args[3].(int64)

	if !okH || !okP || !okL || !okC {
		return nil, fmt.Errorf("%w: GoFindDeclarations invalid argument types (expected string, string, int64, int64)", core.ErrInvalidArgument)
	}
	if lineRaw <= 0 || colRaw <= 0 {
		return nil, fmt.Errorf("%w: GoFindDeclarations line/column must be positive", core.ErrInvalidArgument)
	}
	line, col := int(lineRaw), int(colRaw)

	logger.Debug("[TOOL-GOFINDDECL-LC] Request", "handle", handle, "path", pathRel, "L", line, "C", col)

	// --- Get Index ---
	indexValue, err := interpreter.GetHandleValue(handle, semanticIndexTypeTag) // Use consistent tag
	if err != nil {
		logger.Error("[TOOL-GOFINDDECL-LC] Failed get handle", "handle", handle, "error", err)
		return nil, err // Propagate core handle error
	}
	index, ok := indexValue.(*SemanticIndex)
	if !ok {
		logger.Error("[TOOL-GOFINDDECL-LC] Handle not *SemanticIndex", "handle", handle, "type", fmt.Sprintf("%T", indexValue))
		return nil, fmt.Errorf("%w: handle '%s' is not a SemanticIndex", core.ErrHandleWrongType, handle)
	}
	if index.Fset == nil || index.Packages == nil { // Check Packages as well
		logger.Error("[TOOL-GOFINDDECL-LC] Index FileSet or Packages nil", "handle", handle)
		return nil, fmt.Errorf("%w: index '%s' has nil FileSet or Packages", ErrIndexNotReady, handle) // Use specific error
	}

	// --- Resolve Path and Position ---
	absPath, pathErr := core.ResolveAndSecurePath(pathRel, index.LoadDir) // Use core helper
	if pathErr != nil {
		logger.Error("[TOOL-GOFINDDECL-LC] Path resolve failed", "path", pathRel, "load_dir", index.LoadDir, "error", pathErr)
		return nil, fmt.Errorf("%w: path '%s': %w", core.ErrInvalidPath, pathRel, pathErr)
	}

	// Use helper from this package
	pos, posErr := findPosInFileSet(index.Fset, absPath, line, col)
	if posErr != nil {
		logger.Error("[TOOL-GOFINDDECL-LC] Error resolving position", "path", absPath, "L", line, "C", col, "error", posErr)
		return nil, fmt.Errorf("%w: error finding position in fileset for %s:%d:%d: %w", core.ErrInternal, absPath, line, col, posErr)
	}
	if pos == token.NoPos {
		logger.Warn("[TOOL-GOFINDDECL-LC] Position not found in FileSet (likely out of bounds)", "path", absPath, "L", line, "C", col)
		return nil, nil // Not found
	}

	logger.Debug("[TOOL-GOFINDDECL-LC] Resolved", "abs_path", absPath, "pos", pos)

	// --- Find AST Node & Package ---
	// Use helper from this package
	targetIdent, targetPkgInfo, findErr := findIdentAndPackageAtPos(index, pos)
	if findErr != nil {
		// Propagate internal errors from helper
		return nil, findErr
	}
	// Check if target or critical info is missing
	if targetIdent == nil || targetPkgInfo == nil || targetPkgInfo.TypesInfo == nil {
		logger.Warn("[TOOL-GOFINDDECL-LC] Could not find identifier or package/type info at target position", "pos", pos)
		return nil, nil // Target identifier not found or package info missing
	}

	// --- Find Declaration Object ---
	logger.Debug("[TOOL-GOFINDDECL-LC] Looking up object", "identifier", targetIdent.Name, "pos", targetIdent.Pos())

	// Use TypesInfo from the resolved core.PackageInfo
	obj := targetPkgInfo.TypesInfo.ObjectOf(targetIdent)
	if obj == nil {
		obj = targetPkgInfo.TypesInfo.Uses[targetIdent]
	}
	if obj == nil {
		obj = targetPkgInfo.TypesInfo.Defs[targetIdent]
	}

	if obj == nil {
		logger.Warn("[TOOL-GOFINDDECL-LC] Could not find type object via ObjectOf, Uses, or Defs", "ident", targetIdent.Name, "identPos", index.Fset.Position(targetIdent.Pos()))
		// Simplified check for unresolved package name
		if pkgName, isPkg := targetIdent.Obj.Decl.(*types.PkgName); isPkg { // Check if the Ident itself resolves to a PkgName obj
			logger.Info("[TOOL-GOFINDDECL-LC] Identifier is likely an unresolved PkgName", "ident", targetIdent.Name, "pkgPath", pkgName.Imported().Path())
			return nil, nil
		}
		return nil, nil // Truly not found
	}

	logger.Debug("[TOOL-GOFINDDECL-LC] Found object", "identifier", targetIdent.Name, "objName", obj.Name(), "objType", fmt.Sprintf("%T", obj), "objPos", index.Fset.Position(obj.Pos()))

	if _, isPkgName := obj.(*types.PkgName); isPkgName {
		logger.Info("[TOOL-GOFINDDECL-LC] Identifier is PkgName, ignoring.", "ident", targetIdent.Name, "pkgPath", obj.(*types.PkgName).Imported().Path())
		return nil, nil
	}

	// --- Determine Declaration Position & Filter ---
	declPos := obj.Pos()
	kind := getObjectKind(obj) // Use helper from this package

	logger.Debug("[TOOL-GOFINDDECL-LC] Using obj.Pos() for declaration", "kind", kind, "pos", declPos, "objReported", index.Fset.Position(obj.Pos()))

	if !declPos.IsValid() {
		logger.Warn("[TOOL-GOFINDDECL-LC] Declaration position invalid", "object", obj.Name())
		return nil, nil
	}

	declPosition := index.Fset.Position(declPos)
	declFilenameAbs := declPosition.Filename

	if declFilenameAbs == "" {
		logger.Warn("[TOOL-GOFINDDECL-LC] Declaration has no filename", "object", obj.Name())
		return nil, nil
	}

	// Path filtering logic (relative to index.LoadDir)
	relDeclPathCheck, errCheck := filepath.Rel(index.LoadDir, declFilenameAbs)
	cleanLoadDir := filepath.Clean(index.LoadDir)
	cleanDeclFilenameAbs := filepath.Clean(declFilenameAbs)

	// Check if the declaration is within the LoadDir or is the LoadDir itself
	if errCheck != nil || (!strings.HasPrefix(cleanDeclFilenameAbs, cleanLoadDir+string(filepath.Separator)) && cleanDeclFilenameAbs != cleanLoadDir) {
		logger.Info("[TOOL-GOFINDDECL-LC] Declaration outside indexed dir, filtering.", "object", obj.Name(), "decl_path", declFilenameAbs, "load_dir", cleanLoadDir)
		return nil, nil // Outside indexed scope
	}

	relDeclPath := filepath.ToSlash(relDeclPathCheck)
	name := obj.Name()
	// Fallback name logic (seems less likely needed with types.Object)
	// if name == "" {
	// 	name = targetIdent.Name
	// 	logger.Debug("[TOOL-GOFINDDECL-LC] Object name is empty, using identifier name as fallback", "fallback_name", name)
	// }

	// --- Return Result ---
	logger.Debug("[TOOL-GOFINDDECL-LC] Found declaration", "name", name, "kind", kind, "path", relDeclPath, "L", declPosition.Line, "C", declPosition.Column)
	return map[string]interface{}{
		"path":   relDeclPath,
		"line":   int64(declPosition.Line),
		"column": int64(declPosition.Column),
		"name":   name,
		"kind":   kind,
	}, nil
}
