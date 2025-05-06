// NeuroScript Version: 0.3.1
// File version: 0.3.3
// Add ToolImplementation definition.
// filename: pkg/core/tools_go_find_declarations.go

package core

import (
	"fmt"
	"go/token"
	"go/types" // Needed for MaxInt32
	"path/filepath"
	"strings"
	// "golang.org/x/tools/go/ast/astutil" // No longer needed
)

// +++ ADDED ToolImplementation definition +++
var toolGoFindDeclarationsImpl = ToolImplementation{
	Spec: ToolSpec{
		Name:        "GoFindDeclarations",
		Description: "Finds the declaration location of the Go symbol at the specified file position using a semantic index handle.",
		Args: []ArgSpec{
			// Note: Renamed handle arg to 'index_handle' for clarity vs other handle types
			{Name: "index_handle", Type: ArgTypeString, Required: true, Description: "Handle returned by GoIndexCode."},
			{Name: "path", Type: ArgTypeString, Required: true, Description: "File path relative to the indexed directory root."},
			{Name: "line", Type: ArgTypeInt, Required: true, Description: "1-based line number of the symbol."},
			{Name: "column", Type: ArgTypeInt, Required: true, Description: "1-based column number of the symbol."},
		},
		ReturnType: ArgTypeMap, // Returns map {path, line, column, name, kind} or nil
	},
	Func: toolGoFindDeclarations,
}

// toolGoFindDeclarations finds the declaration location of a Go symbol.
// ... (rest of the function implementation remains unchanged) ...
func toolGoFindDeclarations(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	logger := interpreter.Logger()

	// --- Argument Parsing and Validation ---
	if len(args) != 4 {
		// Adjust error message if argument name changed in spec (index_handle vs handle)
		return nil, fmt.Errorf("%w: GoFindDeclarations requires 4 arguments (index_handle, path, line, column)", ErrInvalidArgument)
	}
	// Use index 0 for handle based on ArgSpecs order
	handle, okH := args[0].(string)
	pathRel, okP := args[1].(string)
	lineRaw, okL := args[2].(int64)
	colRaw, okC := args[3].(int64)

	if !okH || !okP || !okL || !okC {
		return nil, fmt.Errorf("%w: GoFindDeclarations invalid argument types (expected string, string, int64, int64)", ErrInvalidArgument)
	}
	if lineRaw <= 0 || colRaw <= 0 {
		return nil, fmt.Errorf("%w: GoFindDeclarations line/column must be positive", ErrInvalidArgument)
	}
	// Convert safely to int for internal use
	line, col := int(lineRaw), int(colRaw)

	// Use correct arg name in log if changed
	logger.Debug("[TOOL-GOFINDDECL] Request", "handle", handle, "path", pathRel, "L", line, "C", col)

	// --- Get Index ---
	indexValue, err := interpreter.GetHandleValue(handle, "semantic_index")
	if err != nil {
		logger.Error("[TOOL-GOFINDDECL] Failed get handle", "handle", handle, "error", err)
		return nil, err // Propagate handle error
	}
	index, ok := indexValue.(*SemanticIndex)
	if !ok {
		logger.Error("[TOOL-GOFINDDECL] Handle not *SemanticIndex", "handle", handle, "type", fmt.Sprintf("%T", indexValue))
		return nil, fmt.Errorf("%w: handle '%s' not a SemanticIndex", ErrHandleWrongType, handle)
	}
	if index.Fset == nil {
		logger.Error("[TOOL-GOFINDDECL] Index FileSet nil", "handle", handle)
		return nil, fmt.Errorf("%w: index '%s' has nil FileSet", ErrInternal, handle)
	}

	// --- Resolve Path and Position ---
	absPath, pathErr := ResolveAndSecurePath(pathRel, index.LoadDir)
	if pathErr != nil {
		logger.Error("[TOOL-GOFINDDECL] Path resolve failed", "path", pathRel, "load_dir", index.LoadDir, "error", pathErr)
		// Return specific path error
		return nil, fmt.Errorf("%w: path '%s': %w", ErrInvalidPath, pathRel, pathErr)
	}

	pos, posErr := findPosInFileSet(index.Fset, absPath, line, col)
	if posErr != nil {
		// This indicates a more serious issue translating line/col to pos
		logger.Error("[TOOL-GOFINDDECL] Error resolving position", "path", absPath, "L", line, "C", col, "error", posErr)
		return nil, fmt.Errorf("%w: error finding position in fileset for %s:%d:%d: %w", ErrInternal, absPath, line, col, posErr)
	}
	if pos == token.NoPos {
		// This means the line/col is outside the file's range according to the fileset
		logger.Warn("[TOOL-GOFINDDECL] Position not found in FileSet (likely out of bounds)", "path", absPath, "L", line, "C", col)
		// Return empty map instead of nil for consistency? Test expects nil currently.
		// For now, keep returning nil to match existing test expectation.
		return nil, nil // Not found
	}

	logger.Debug("[TOOL-GOFINDDECL] Resolved", "abs_path", absPath, "pos", pos)

	// --- Find AST Node & Package using ast.Inspect ---
	targetIdent, targetPkg, findErr := findIdentAndPackageAtPos(index, pos)
	if findErr != nil {
		// findIdentAndPackageAtPos logs details internally if needed
		return nil, findErr // Propagate internal errors
	}
	if targetIdent == nil || targetPkg == nil || targetPkg.TypesInfo == nil {
		logger.Warn("[TOOL-GOFINDDECL] Could not find identifier or package/type info at target position", "pos", pos)
		// Return empty map instead of nil? Keep returning nil for now.
		return nil, nil // Target identifier not found or package info missing
	}

	// --- Find Declaration Object ---
	logger.Debug("[TOOL-GOFINDDECL] Looking up object", "identifier", targetIdent.Name, "pos", targetIdent.Pos())

	obj := targetPkg.TypesInfo.ObjectOf(targetIdent) // Best case: direct definition/use
	if obj == nil {
		obj = targetPkg.TypesInfo.Uses[targetIdent] // Check uses map
	}
	if obj == nil {
		obj = targetPkg.TypesInfo.Defs[targetIdent] // Check definitions map
	}

	if obj == nil {
		logger.Warn("[TOOL-GOFINDDECL] Could not find type object via ObjectOf, Uses, or Defs", "ident", targetIdent.Name, "identPos", index.Fset.Position(targetIdent.Pos()))
		// Check if it might be an unresolved package name identifier
		if targetPkg.Types != nil {
			scope := targetPkg.Types.Scope()
			if scope != nil {
				_, objInScope := scope.LookupParent(targetIdent.Name, targetIdent.Pos())
				if pkgName, isPkg := objInScope.(*types.PkgName); isPkg {
					logger.Info("[TOOL-GOFINDDECL] Identifier is likely an unresolved PkgName", "ident", targetIdent.Name, "pkgPath", pkgName.Imported().Path())
					return nil, nil // Treat package names as "not found" for declarations
				}
			}
		}
		// Return empty map instead of nil? Keep returning nil for now.
		return nil, nil // Truly not found
	}

	logger.Debug("[TOOL-GOFINDDECL] Found object", "identifier", targetIdent.Name, "objName", obj.Name(), "objType", fmt.Sprintf("%T", obj), "objPos", index.Fset.Position(obj.Pos()))

	if _, isPkgName := obj.(*types.PkgName); isPkgName {
		logger.Info("[TOOL-GOFINDDECL] Identifier is PkgName, ignoring.", "ident", targetIdent.Name, "pkgPath", obj.(*types.PkgName).Imported().Path())
		// Return empty map instead of nil? Keep returning nil for now.
		return nil, nil // Don't return declarations for package names
	}

	// --- Determine Declaration Position & Filter ---
	declPos := obj.Pos()
	kind := getObjectKind(obj)

	logger.Debug("[TOOL-GOFINDDECL] Using obj.Pos() for declaration", "kind", kind, "pos", declPos, "objReported", index.Fset.Position(obj.Pos()))

	if !declPos.IsValid() {
		logger.Warn("[TOOL-GOFINDDECL] Declaration position invalid", "object", obj.Name())
		// Return empty map instead of nil? Keep returning nil for now.
		return nil, nil // Invalid position
	}

	declPosition := index.Fset.Position(declPos)
	declFilenameAbs := declPosition.Filename

	if declFilenameAbs == "" {
		logger.Warn("[TOOL-GOFINDDECL] Declaration has no filename", "object", obj.Name())
		// Return empty map instead of nil? Keep returning nil for now.
		return nil, nil // No filename associated
	}

	relDeclPathCheck, errCheck := filepath.Rel(index.LoadDir, declFilenameAbs)
	cleanLoadDir := filepath.Clean(index.LoadDir)
	cleanDeclFilenameAbs := filepath.Clean(declFilenameAbs)

	if errCheck != nil || (!strings.HasPrefix(cleanDeclFilenameAbs, cleanLoadDir+string(filepath.Separator)) && cleanDeclFilenameAbs != cleanLoadDir) {
		logger.Info("[TOOL-GOFINDDECL] Declaration outside indexed dir, filtering.", "object", obj.Name(), "decl_path", declFilenameAbs, "load_dir", cleanLoadDir)
		// Return empty map instead of nil? Keep returning nil for now.
		return nil, nil // Outside indexed scope
	}

	relDeclPath := filepath.ToSlash(relDeclPathCheck)

	name := obj.Name()
	if name == "" {
		name = targetIdent.Name
		logger.Debug("[TOOL-GOFINDDECL] Object name is empty, using identifier name as fallback", "fallback_name", name)
	}

	// --- Return Result ---
	logger.Debug("[TOOL-GOFINDDECL] Found declaration", "name", name, "kind", kind, "path", relDeclPath, "L", declPosition.Line, "C", declPosition.Column)
	return map[string]interface{}{
		"path":   relDeclPath,
		"line":   int64(declPosition.Line),
		"column": int64(declPosition.Column),
		"name":   name,
		"kind":   kind,
	}, nil
}

// findPosInFileSet (helper function remains unchanged)
func findPosInFileSet(fset *token.FileSet, absPath string, line, col int) (token.Pos, error) {
	// ... (implementation unchanged) ...
	var foundFile *token.File = nil
	fset.Iterate(func(f *token.File) bool {
		if filepath.Clean(f.Name()) == filepath.Clean(absPath) {
			foundFile = f
			return false
		}
		return true
	})
	if foundFile == nil {
		return token.NoPos, nil
	}
	if line <= 0 || col <= 0 {
		return token.NoPos, nil
	}
	if line > foundFile.LineCount() {
		return token.NoPos, nil
	}
	lineStartPos := foundFile.LineStart(line)
	if !lineStartPos.IsValid() {
		return token.NoPos, fmt.Errorf("invalid line start position for %s:%d", absPath, line)
	}
	finalPos := lineStartPos + token.Pos(col-1)
	fileBasePos := token.Pos(foundFile.Base())
	fileEndPos := fileBasePos + token.Pos(foundFile.Size())
	if finalPos < fileBasePos || finalPos >= fileEndPos {
		return token.NoPos, nil
	}
	return finalPos, nil
}

// getObjectKind (helper function remains unchanged)
func getObjectKind(obj types.Object) string {
	// ... (implementation unchanged) ...
	if obj == nil {
		return "unknown"
	}
	switch o := obj.(type) {
	case *types.Var:
		if o.IsField() {
			return "field"
		}
		return "variable"
	case *types.Func:
		if sig := o.Type().(*types.Signature); sig.Recv() != nil {
			return "method"
		}
		return "function"
	case *types.TypeName:
		return "type"
	case *types.Const:
		return "constant"
	case *types.PkgName:
		return "package"
	case *types.Label:
		return "label"
	case *types.Builtin:
		return "builtin"
	case *types.Nil:
		return "nil"
	default:
		return "unknown"
	}
}
