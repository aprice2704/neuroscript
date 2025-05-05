// NeuroScript Version: 0.3.1
// Last Modified: 2025-05-04 12:47:54 PDT // Refactored from tools_go_semantic.go, AST Dump commented out
// filename: pkg/core/tools_go_find_declarations.go

package core

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"math" // Needed for MaxInt32
	"path/filepath"
	"strings"

	// "golang.org/x/tools/go/ast/astutil" // No longer needed
	"golang.org/x/tools/go/packages"
)

// toolGoFindDeclarations finds the declaration location of a Go symbol.
// It takes a semantic index handle, a file path, line, and column number.
// It uses the index to locate the identifier at the given position and
// then finds its declaration site using type information.
// Args:
//   - handle (string): Handle to the SemanticIndex.
//   - path (string): Relative path to the file containing the symbol.
//   - line (int64): Line number (1-based).
//   - column (int64): Column number (1-based).
//
// Returns:
//   - map[string]interface{}: A map containing declaration details:
//     {"path": string, "line": int64, "column": int64, "name": string, "kind": string}
//   - nil: If the symbol or declaration is not found, or if it's outside the indexed directory.
//   - error: For argument errors, handle errors, or internal issues.
func toolGoFindDeclarations(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	logger := interpreter.Logger()

	// --- Argument Parsing and Validation ---
	if len(args) != 4 {
		return nil, fmt.Errorf("%w: GoFindDeclarations requires 4 arguments (handle, path, line, column)", ErrInvalidArgument)
	}
	handle, okH := args[0].(string)
	pathRel, okP := args[1].(string)
	lineRaw, okL := args[2].(int64) // Assume user provides int64
	colRaw, okC := args[3].(int64)  // Assume user provides int64

	if !okH || !okP || !okL || !okC {
		return nil, fmt.Errorf("%w: GoFindDeclarations invalid argument types (expected string, string, int64, int64)", ErrInvalidArgument)
	}
	if lineRaw <= 0 || colRaw <= 0 {
		return nil, fmt.Errorf("%w: GoFindDeclarations line/column must be positive", ErrInvalidArgument)
	}
	// Convert safely to int for internal use
	line, col := int(lineRaw), int(colRaw)

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
		return nil, nil // Not found, return nil result
	}

	logger.Debug("[TOOL-GOFINDDECL] Resolved", "abs_path", absPath, "pos", pos)

	// --- Find AST Node & Package using ast.Inspect ---
	var targetIdent *ast.Ident
	var targetPkg *packages.Package
	var candidateIdents []*ast.Ident // Collect all matching idents

	foundFileNode := false
	for _, pkg := range index.Packages {
		if pkg == nil {
			continue // Skip nil packages
		}
		// Check for package errors if needed, although toolGoIndex should handle critical ones
		// if len(pkg.Errors) > 0 { continue }

		for _, fileNode := range pkg.Syntax {
			if fileNode == nil {
				continue // Skip nil file nodes
			}
			tokenFile := index.Fset.File(fileNode.Pos())
			// Check tokenFile is not nil before accessing Name()
			if tokenFile == nil || filepath.Clean(tokenFile.Name()) != filepath.Clean(absPath) {
				continue // Not the file we're looking for
			}

			// Found the correct file node
			foundFileNode = true
			targetPkg = pkg // Store the package containing this file

			// --- AST Inspection ---
			// >>> AST DUMP COMMENTED OUT TO REDUCE VERBOSITY <<<
			/*
				logger.Debug("[TOOL-GOFINDDECL] --- AST Dump Start ---", "file", absPath)
				printErr := ast.Print(index.Fset, fileNode) // Use ast.Print for detailed dump
				if printErr != nil {
					logger.Error("[TOOL-GOFINDDECL] Error printing AST", "error", printErr)
					// Log the error but continue; AST dump is for debugging
				}
				logger.Debug("[TOOL-GOFINDDECL] --- AST Dump End ---", "file", absPath)
			*/

			// Use ast.Inspect to find the *most specific* identifier containing the position.
			ast.Inspect(fileNode, func(n ast.Node) bool {
				if n == nil {
					return true // Continue traversal if node is nil
				}

				// We are only interested in identifiers
				ident, ok := n.(*ast.Ident)
				if !ok {
					return true // Continue traversal for non-identifier nodes
				}

				// Get the start and end position of the identifier node
				startPos := ident.Pos()
				endPos := ident.End()

				// Check if the identifier's range is valid
				if !startPos.IsValid() || !endPos.IsValid() {
					return true // Skip identifiers with invalid positions
				}

				// Check if the target position `pos` falls within the identifier's range [start, end)
				// The standard AST range is half-open: [start, end)
				if pos >= startPos && pos < endPos {
					logger.Debug("[TOOL-GOFINDDECL] Found candidate ident via Inspect", "name", ident.Name, "pos", pos, "start", startPos, "end", endPos, "match_logic", "pos>=start && pos<end")
					candidateIdents = append(candidateIdents, ident)
					// Do *not* return false here, we want the innermost identifier.
					// Returning true ensures we visit children of this node.
				}

				// Always return true to continue inspecting the rest of the AST subtree,
				// ensuring we find the most specific (innermost) node if nodes are nested.
				return true
			})

			break // Stop searching through files once the target file is processed
		} // end for fileNode

		if foundFileNode {
			break // Stop searching through packages once the target file is found
		}
	} // end for pkg

	// --- Select the best (smallest range) candidate ---
	if len(candidateIdents) > 0 {
		minRange := int32(math.MaxInt32) // Use int32 for range comparison
		for _, ident := range candidateIdents {
			// Double-check validity before calculating range
			if !ident.Pos().IsValid() || !ident.End().IsValid() {
				continue
			}
			currentRange := int32(ident.End() - ident.Pos()) // Calculate range
			if currentRange < 0 {
				continue // Should not happen, but safeguard
			}

			// If this identifier has a smaller range, it's a better candidate
			if currentRange < minRange {
				minRange = currentRange
				targetIdent = ident // Update the best candidate found so far
			}
		}
		if targetIdent != nil {
			logger.Debug("[TOOL-GOFINDDECL] Selected best ident from candidates", "name", targetIdent.Name, "pos", targetIdent.Pos(), "end", targetIdent.End(), "range", minRange)
		}
	}

	// --- Check if identifier and package were found ---
	if targetIdent == nil {
		// This means Inspect did not find any identifier containing the pos
		logger.Warn("[TOOL-GOFINDDECL] No suitable identifier found after Inspect.", "path", absPath, "L", line, "C", col, "candidates_found", len(candidateIdents))
		return nil, nil // Not found
	}
	if targetPkg == nil || targetPkg.TypesInfo == nil {
		errMsg := "Package info missing or incomplete."
		if targetPkg == nil {
			errMsg = "Containing package not found."
		} else if targetPkg.TypesInfo == nil {
			// This can happen if the package had loading errors.
			errMsg = fmt.Sprintf("Package '%s' lacks type info (check index load errors).", targetPkg.PkgPath)
		}
		logger.Warn("[TOOL-GOFINDDECL] "+errMsg, "path", absPath, "L", line, "C", col)
		return nil, nil // Cannot proceed without type info
	}

	// --- Find Declaration Object ---
	logger.Debug("[TOOL-GOFINDDECL] Looking up object", "identifier", targetIdent.Name, "pos", targetIdent.Pos())

	// Attempt to find the object definition or use using TypesInfo
	obj := targetPkg.TypesInfo.ObjectOf(targetIdent) // Best case: direct definition/use

	// Fallbacks if ObjectOf returns nil
	if obj == nil {
		obj = targetPkg.TypesInfo.Uses[targetIdent] // Check uses map
	}
	if obj == nil {
		obj = targetPkg.TypesInfo.Defs[targetIdent] // Check definitions map
	}

	// If still not found, it's an error or unresolved identifier
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
		return nil, nil // Truly not found
	}

	logger.Debug("[TOOL-GOFINDDECL] Found object", "identifier", targetIdent.Name, "objName", obj.Name(), "objType", fmt.Sprintf("%T", obj), "objPos", index.Fset.Position(obj.Pos()))

	// Handle specific cases like package names which don't have a user-code declaration site
	if _, isPkgName := obj.(*types.PkgName); isPkgName {
		logger.Info("[TOOL-GOFINDDECL] Identifier is PkgName, ignoring.", "ident", targetIdent.Name, "pkgPath", obj.(*types.PkgName).Imported().Path())
		return nil, nil // Don't return declarations for package names
	}

	// --- Determine Declaration Position & Filter ---
	declPos := obj.Pos() // Get the position of the object's declaration
	kind := getObjectKind(obj)

	// --- Simplified: Always use obj.Pos() for now ---
	logger.Debug("[TOOL-GOFINDDECL] Using obj.Pos() for declaration", "kind", kind, "pos", declPos, "objReported", index.Fset.Position(obj.Pos()))

	if !declPos.IsValid() {
		logger.Warn("[TOOL-GOFINDDECL] Declaration position invalid", "object", obj.Name())
		return nil, nil // Invalid position
	}

	declPosition := index.Fset.Position(declPos)
	declFilenameAbs := declPosition.Filename

	// Check if the declaration is within the indexed directory structure
	if declFilenameAbs == "" {
		logger.Warn("[TOOL-GOFINDDECL] Declaration has no filename", "object", obj.Name())
		return nil, nil // No filename associated
	}

	// Ensure the declaration path is within the original LoadDir used for indexing
	relDeclPathCheck, errCheck := filepath.Rel(index.LoadDir, declFilenameAbs)
	cleanLoadDir := filepath.Clean(index.LoadDir)
	cleanDeclFilenameAbs := filepath.Clean(declFilenameAbs)

	// Check if relative path calculation failed OR if the absolute path is not prefixed by the load dir
	// (allowing for the case where the file *is* the load dir itself, e.g., single file package).
	if errCheck != nil || (!strings.HasPrefix(cleanDeclFilenameAbs, cleanLoadDir+string(filepath.Separator)) && cleanDeclFilenameAbs != cleanLoadDir) {
		logger.Info("[TOOL-GOFINDDECL] Declaration outside indexed dir, filtering.", "object", obj.Name(), "decl_path", declFilenameAbs, "load_dir", cleanLoadDir)
		return nil, nil // Outside indexed scope
	}

	relDeclPath := filepath.ToSlash(relDeclPathCheck) // Use forward slashes for consistency

	// Determine the name to report (usually obj.Name, fallback to targetIdent.Name if empty)
	name := obj.Name()
	if name == "" {
		name = targetIdent.Name
		logger.Debug("[TOOL-GOFINDDECL] Object name is empty, using identifier name as fallback", "fallback_name", name)
	}

	// --- Return Result ---
	// Changed from Info to Debug to further reduce noise on success
	logger.Debug("[TOOL-GOFINDDECL] Found declaration", "name", name, "kind", kind, "path", relDeclPath, "L", declPosition.Line, "C", declPosition.Column)
	return map[string]interface{}{
		"path":   relDeclPath,
		"line":   int64(declPosition.Line),   // Return as int64
		"column": int64(declPosition.Column), // Return as int64
		"name":   name,
		"kind":   kind,
	}, nil
}

// findPosInFileSet converts a 1-based line and column number in a given absolute file path
// into a token.Pos within the provided FileSet.
// It returns token.NoPos if the file is not found in the FileSet or if the
// line/column is invalid or out of bounds.
func findPosInFileSet(fset *token.FileSet, absPath string, line, col int) (token.Pos, error) {
	var foundFile *token.File = nil

	// Iterate through the FileSet to find the token.File matching the absolute path.
	fset.Iterate(func(f *token.File) bool {
		// Normalize paths before comparing to handle OS differences.
		if filepath.Clean(f.Name()) == filepath.Clean(absPath) {
			foundFile = f
			return false // Stop iteration once the file is found
		}
		return true // Continue iteration
	})

	// If the file wasn't found in the FileSet.
	if foundFile == nil {
		// Considered not found, not an error state for this function's purpose.
		return token.NoPos, nil
	}

	// Validate line and column numbers (must be positive).
	if line <= 0 || col <= 0 {
		// Invalid input, return NoPos.
		return token.NoPos, nil
	}

	// Check if the line number exceeds the file's line count.
	if line > foundFile.LineCount() {
		// Out of bounds, return NoPos.
		return token.NoPos, nil
	}

	// Get the starting position of the target line.
	// Note: token.File.LineStart returns the position of the first character of the line.
	lineStartPos := foundFile.LineStart(line)

	// Check if the obtained line start position is valid.
	if !lineStartPos.IsValid() {
		// If LineStart is invalid, something is wrong with the FileSet or line number.
		// This might indicate an internal issue or corrupted FileSet state.
		return token.NoPos, fmt.Errorf("invalid line start position for %s:%d", absPath, line)
	}

	// Calculate the final position by adding the 0-based column offset to the line start position.
	// token.Pos is essentially an offset from the start of the FileSet.
	finalPos := lineStartPos + token.Pos(col-1) // Convert 1-based column to 0-based offset.

	// --- Boundary Checks ---
	// It's crucial to ensure the calculated position is actually within the bounds of the file.
	fileBasePos := token.Pos(foundFile.Base())              // Start position of the file in the FileSet.
	fileEndPos := fileBasePos + token.Pos(foundFile.Size()) // End position (exclusive) of the file.

	// Check if the final position is before the file starts or at/after the file ends.
	// The range is [fileBasePos, fileEndPos).
	if finalPos < fileBasePos || finalPos >= fileEndPos {
		// Calculated position falls outside the file's actual content range.
		// This can happen if the column number is beyond the line length.
		return token.NoPos, nil // Treat as out of bounds, return NoPos.
	}

	// The calculated position is valid and within the file's bounds.
	return finalPos, nil
}

// getObjectKind determines a simple string representation of the kind of Go object.
func getObjectKind(obj types.Object) string {
	if obj == nil {
		return "unknown"
	}
	switch o := obj.(type) {
	case *types.Var:
		if o.IsField() {
			return "field" // Struct field
		}
		return "variable" // Package-level or local variable
	case *types.Func:
		// Check if it's a method by looking for a receiver
		if sig := o.Type().(*types.Signature); sig.Recv() != nil {
			return "method"
		}
		return "function" // Regular function
	case *types.TypeName:
		// Could be an interface, struct, alias, or basic type definition
		// Further checks on o.Type() could distinguish these if needed.
		return "type"
	case *types.Const:
		return "constant"
	case *types.PkgName:
		// Represents an imported package name (e.g., 'fmt' in fmt.Println)
		return "package"
	case *types.Label:
		return "label" // goto label
	case *types.Builtin:
		// Builtin function like make, new, append etc.
		return "builtin"
	case *types.Nil:
		// The predefined 'nil' object
		return "nil"
	default:
		// Fallback for any other types.Object implementations
		return "unknown"
	}
}
