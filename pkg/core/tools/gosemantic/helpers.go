// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Use local PackageInfo, remove unused var, remove DefaultLogger calls.
// Shared helper functions for the gosemantic package.
// filename: pkg/core/tools/gosemantic/helpers.go

package gosemantic

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"path/filepath"

	"github.com/aprice2704/neuroscript/pkg/core" // Still need core for some types/errors
	// Need this for packages.Package access
)

// getObjectKind determines a simple string representation of a types.Object kind.
// (Implementation unchanged)
func getObjectKind(obj types.Object) string {
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

// findPosInFileSet converts an absolute path, line, and column into a token.Pos
// (Implementation unchanged)
func findPosInFileSet(fset *token.FileSet, absPath string, line, col int) (token.Pos, error) {
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
		return token.NoPos, fmt.Errorf("internal error: invalid line start position for %s:%d", absPath, line)
	}
	finalPos := lineStartPos + token.Pos(col-1)
	fileBasePos := token.Pos(foundFile.Base())
	fileEndPos := fileBasePos + token.Pos(foundFile.Size())
	if finalPos < fileBasePos || finalPos >= fileEndPos {
		return token.NoPos, nil
	}
	return finalPos, nil
}

// findIdentAndPackageAtPos finds the *ast.Ident node and its containing *PackageInfo
// at a given token.Pos within the SemanticIndex.
func findIdentAndPackageAtPos(index *SemanticIndex, pos token.Pos) (*ast.Ident, *PackageInfo, error) { // Return local PackageInfo
	if index == nil || index.Packages == nil {
		return nil, nil, fmt.Errorf("%w: semantic index is nil or has no packages", core.ErrInternal)
	}
	if pos == token.NoPos {
		return nil, nil, fmt.Errorf("%w: invalid token position provided", core.ErrInternal)
	}

	var foundIdent *ast.Ident
	var foundPkgInfo *PackageInfo // Use local PackageInfo
	// var foundASTFile *ast.File // *** REMOVED: Unused variable ***

	// Iterate through packages to find the one containing the position
	for _, pkgInfo := range index.Packages { // pkgInfo is now *gosemantic.PackageInfo
		// Check TypeInfo first as it's crucial
		// Access embedded *packages.Package fields directly
		if pkgInfo.TypesInfo == nil {
			continue // Skip packages without type info
		}

		for _, astFile := range pkgInfo.Syntax {
			// Check if the position falls within this file's range
			if pos >= astFile.Pos() && pos < astFile.End() {
				foundPkgInfo = pkgInfo
				// foundASTFile = astFile // *** REMOVED: Unused assignment ***

				// Found the relevant package and file, now inspect the AST
				ast.Inspect(astFile, func(n ast.Node) bool {
					if n == nil {
						return true
					}
					if pos >= n.Pos() && pos < n.End() {
						if ident, ok := n.(*ast.Ident); ok {
							if foundIdent == nil || (ident.Pos() >= foundIdent.Pos() && ident.End() <= foundIdent.End()) {
								foundIdent = ident
							}
						}
						return true
					}
					return false
				})
				goto found // Use goto for cleaner exit from nested loops
			}
		}
	}

found:
	if foundIdent == nil || foundPkgInfo == nil {
		// *** REMOVED: Call to core.DefaultLogger() ***
		// core.DefaultLogger().Debug("Identifier or package not found at position", "pos", pos, "foundIdent", foundIdent != nil, "foundPkg", foundPkgInfo != nil)
		return nil, nil, nil // Not found, return nil without error
	}

	// Defensive check: Ensure the found package actually has TypeInfo (should be redundant)
	if foundPkgInfo.TypesInfo == nil {
		// *** REMOVED: Call to core.DefaultLogger() ***
		// core.DefaultLogger().Error("Internal inconsistency: Found identifier in package without TypeInfo", "pkgId", foundPkgInfo.ID)
		return nil, nil, fmt.Errorf("%w: package '%s' lacks TypeInfo despite containing position %v", core.ErrInternal, foundPkgInfo.ID, pos)
	}

	return foundIdent, foundPkgInfo, nil
}
