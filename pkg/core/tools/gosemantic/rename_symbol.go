// NeuroScript Version: 0.3.1
// File version: 0.0.2 // Fix undefined error: Use ErrInvalidArgument for non-renameable symbols.
// Initial implementation stub for GoRenameSymbol tool.
// filename: pkg/core/tools/gosemantic/rename_symbol.go

package gosemantic

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser" // Need parser to validate new name is an identifier
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// --- Tool Definition: GoRenameSymbol ---

var toolGoRenameSymbolImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name: "GoRenameSymbol",
		Description: "Finds a Go symbol via semantic query and generates a list of specific text replacement operations needed to rename it and all its usages within the indexed scope.\n" +
			"Does not modify files directly. Returns a list of patch operations.",
		Args: []core.ArgSpec{
			{Name: "index_handle", Type: core.ArgTypeString, Required: true, Description: "Handle returned by GoIndexCode."},
			{Name: "query", Type: core.ArgTypeString, Required: true, Description: "Semantic query string identifying the symbol to rename (e.g., 'package:my/pkg; function:MyFunc')."},
			{Name: "new_name", Type: core.ArgTypeString, Required: true, Description: "The desired new name for the symbol. Must be a valid Go identifier."},
		},
		// Returns a list of maps, each representing a patch operation:
		// {"path": string, "offset_start": int64, "offset_end": int64, "original_text": string, "new_text": string}
		ReturnType: core.ArgTypeSliceMap, // Confirmed from tools_types.go
	},
	Func: toolGoRenameSymbol,
}

// toolGoRenameSymbol implements the GoRenameSymbol tool logic.
func toolGoRenameSymbol(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	logger := interpreter.Logger()

	// --- Argument Parsing and Validation ---
	if len(args) != 3 {
		return nil, fmt.Errorf("%w: GoRenameSymbol requires 3 arguments (index_handle, query, new_name)", core.ErrInvalidArgument)
	}
	handle, okH := args[0].(string)
	query, okQ := args[1].(string)
	newName, okN := args[2].(string)

	if !okH || !okQ || !okN {
		return nil, fmt.Errorf("%w: GoRenameSymbol invalid argument types (expected string, string, string)", core.ErrInvalidArgument)
	}
	if query == "" {
		return nil, fmt.Errorf("%w: GoRenameSymbol query cannot be empty", core.ErrInvalidArgument)
	}
	if newName == "" {
		return nil, fmt.Errorf("%w: GoRenameSymbol new_name cannot be empty", core.ErrInvalidArgument)
	}

	// Validate new_name is a valid Go identifier
	if !isValidGoIdentifier(newName) {
		return nil, fmt.Errorf("%w: new_name '%s' is not a valid Go identifier", core.ErrInvalidArgument, newName)
	}

	logger.Debug("[TOOL-GORENAME] Request", "handle", handle, "query", query, "newName", newName)

	// --- Get Index ---
	indexValue, err := interpreter.GetHandleValue(handle, semanticIndexTypeTag)
	if err != nil {
		logger.Error("[TOOL-GORENAME] Failed get handle", "handle", handle, "error", err)
		return nil, err
	}
	index, ok := indexValue.(*SemanticIndex)
	if !ok {
		logger.Error("[TOOL-GORENAME] Handle not *SemanticIndex", "handle", handle, "type", fmt.Sprintf("%T", indexValue))
		return nil, fmt.Errorf("%w: handle '%s' is not a SemanticIndex", core.ErrHandleWrongType, handle)
	}
	if index.Fset == nil || index.Packages == nil {
		logger.Error("[TOOL-GORENAME] Index FileSet or Packages nil", "handle", handle)
		return nil, fmt.Errorf("%w: index '%s' has nil FileSet or Packages", ErrIndexNotReady, handle)
	}

	// --- Find the Declaration Object based on the Query ---
	parsedQuery, parseErr := parseSemanticQuery(query)
	if parseErr != nil {
		logger.Error("[TOOL-GORENAME] Failed to parse query", "query", query, "error", parseErr)
		if errors.Is(parseErr, ErrInvalidQueryFormat) {
			return nil, fmt.Errorf("%w: %w", core.ErrInvalidArgument, parseErr)
		}
		return nil, fmt.Errorf("failed to parse query '%s': %w", query, parseErr)
	}
	logger.Debug("[TOOL-GORENAME] Parsed query", "parsed", parsedQuery)

	declarationObj, findErr := findObjectInIndex(index, parsedQuery, logger)
	if findErr != nil {
		if errors.Is(findErr, ErrSymbolNotFound) || errors.Is(findErr, ErrPackageNotFound) || errors.Is(findErr, ErrWrongKind) {
			logger.Warn("[TOOL-GORENAME] Target symbol not found via query", "query", query, "error", findErr)
			return []interface{}{}, nil // Return empty list if target symbol not found
		}
		logger.Error("[TOOL-GORENAME] Error finding target symbol via query", "query", query, "error", findErr)
		return nil, findErr // Propagate other errors
	}
	if declarationObj == nil {
		logger.Error("[TOOL-GORENAME] Internal inconsistency: findObjectInIndex returned nil object without error", "query", query)
		return nil, fmt.Errorf("%w: findObjectInIndex returned nil object unexpectedly for query '%s'", core.ErrInternal, query)
	}

	// Prevent renaming built-in types or nil, etc.
	if _, isBuiltin := declarationObj.(*types.Builtin); isBuiltin {
		// *** FIXED: Use ErrInvalidArgument ***
		return nil, fmt.Errorf("%w: cannot rename built-in symbol '%s'", core.ErrInvalidArgument, declarationObj.Name())
	}
	if _, isNil := declarationObj.(*types.Nil); isNil {
		// *** FIXED: Use ErrInvalidArgument ***
		return nil, fmt.Errorf("%w: cannot rename 'nil'", core.ErrInvalidArgument)
	}
	if pkgNameObj, isPkgName := declarationObj.(*types.PkgName); isPkgName {
		logger.Warn("[TOOL-GORENAME] Query resolved to a PkgName, cannot rename package import aliases this way.", "query", query, "pkgPath", pkgNameObj.Imported().Path())
		return []interface{}{}, nil
	}
	if !declarationObj.Pos().IsValid() {
		logger.Warn("[TOOL-GORENAME] Declaration object has invalid position", "objName", declarationObj.Name(), "objType", fmt.Sprintf("%T", declarationObj))
		return nil, fmt.Errorf("%w: declaration object for query '%s' has invalid position, cannot reliably find declaration node", core.ErrInternalTool, query)
	}

	originalName := declarationObj.Name()
	if originalName == newName {
		logger.Info("[TOOL-GORENAME] Original name is the same as new name, no rename needed.", "name", originalName)
		return []interface{}{}, nil // No changes needed
	}

	logger.Debug("[TOOL-GORENAME] Identified declaration object to rename", "origName", originalName, "newName", newName, "declObjType", fmt.Sprintf("%T", declarationObj), "declObjPos", index.Fset.Position(declarationObj.Pos()))

	patchOperations := make([]interface{}, 0)
	locationsFound := make(map[token.Pos]bool) // Track positions to avoid duplicates

	// --- Step 1: Find and Add Patch for the Declaration Site ---
	declPos := declarationObj.Pos()
	declPosition := index.Fset.Position(declPos)
	declFilenameAbs := declPosition.Filename

	if declFilenameAbs == "" {
		// This might happen for certain kinds of objects (e.g., implicitly declared methods for embedded fields?)
		// Or for built-ins which we already checked. For safety, return error.
		logger.Error("[TOOL-GORENAME] Declaration object has no filename associated with its position", "query", query, "objName", declarationObj.Name(), "pos", declPos)
		return nil, fmt.Errorf("%w: declaration object for query '%s' has no filename", core.ErrInternalTool, query)
	}

	// Check if declaration is within scope
	relDeclPath, err := filterAndRelativizePath(declFilenameAbs, index.LoadDir, logger, "[TOOL-GORENAME|DECL]")
	if err != nil || relDeclPath == "" {
		logger.Warn("[TOOL-GORENAME] Declaration is outside indexed scope or path error", "declPath", declFilenameAbs, "loadDir", index.LoadDir, "error", err)
		return []interface{}{}, nil // Cannot rename if declaration is outside scope
	}

	// Add declaration patch operation
	// Calculate offsets relative to the start of the file
	declOffsetStart := declPosition.Offset
	if declOffsetStart < 0 {
		logger.Error("[TOOL-GORENAME] Calculated negative start offset for declaration", "query", query, "pos", declPosition)
		return nil, fmt.Errorf("%w: calculated negative offset for declaration position", core.ErrInternalTool)
	}
	// We assume the length of the identifier token matches the name length.
	// This might be wrong for edge cases with weird unicode, but likely okay.
	declOffsetEnd := declOffsetStart + len(originalName)
	declPatch := map[string]interface{}{
		"path":          relDeclPath,
		"offset_start":  int64(declOffsetStart),
		"offset_end":    int64(declOffsetEnd),
		"original_text": originalName,
		"new_text":      newName,
	}
	patchOperations = append(patchOperations, declPatch)
	locationsFound[declPos] = true // Mark declaration position as handled
	logger.Debug("[TOOL-GORENAME] Added patch for declaration", "patch", declPatch)

	// --- Step 2: Find and Add Patches for Usage Sites ---
	logger.Debug("[TOOL-GORENAME] Starting usage search across indexed packages", "numPackages", len(index.Packages))

	for _, pkgInfo := range index.Packages {
		if pkgInfo == nil || pkgInfo.TypesInfo == nil || pkgInfo.TypesInfo.Uses == nil {
			continue
		}
		logger.Debug("[TOOL-GORENAME] Searching package for usages", "pkgId", pkgInfo.ID, "numUses", len(pkgInfo.TypesInfo.Uses))

		for identNode, usedObj := range pkgInfo.TypesInfo.Uses {
			// Check if the object used by this identifier is the same as our target declaration object
			if usedObj == declarationObj {
				usagePos := identNode.Pos()
				if !usagePos.IsValid() {
					continue
				}

				// Skip if we already added a patch for this exact position (e.g., if declaration was also listed in Uses)
				if locationsFound[usagePos] {
					// logger.Debug("[TOOL-GORENAME] Skipping usage at already handled position", "pos", index.Fset.Position(usagePos)) // Can be noisy
					continue
				}

				usagePosition := index.Fset.Position(usagePos)
				usageFilenameAbs := usagePosition.Filename
				if usageFilenameAbs == "" {
					continue
				}

				// Filter: Ensure the usage is within the indexed directory scope
				relUsagePath, err := filterAndRelativizePath(usageFilenameAbs, index.LoadDir, logger, "[TOOL-GORENAME|USAGE]")
				if err != nil || relUsagePath == "" {
					continue // Skip usages outside scope or path errors
				}

				// Verify the identifier name matches (sanity check)
				if identNode.Name != originalName {
					logger.Warn("[TOOL-GORENAME] Mismatched identifier name at usage site", "expected", originalName, "found", identNode.Name, "pos", usagePosition)
					continue // Skip for safety
				}

				// Add valid usage patch operation
				usageOffsetStart := usagePosition.Offset
				if usageOffsetStart < 0 {
					logger.Warn("[TOOL-GORENAME] Calculated negative start offset for usage", "query", query, "usagePos", usagePosition)
					continue // Skip invalid offset
				}
				usageOffsetEnd := usageOffsetStart + len(identNode.Name)
				usagePatch := map[string]interface{}{
					"path":          relUsagePath,
					"offset_start":  int64(usageOffsetStart),
					"offset_end":    int64(usageOffsetEnd),
					"original_text": identNode.Name,
					"new_text":      newName,
				}
				patchOperations = append(patchOperations, usagePatch)
				locationsFound[usagePos] = true // Mark as handled
				logger.Debug("[TOOL-GORENAME] Added patch for usage", "pkgId", pkgInfo.ID, "patch", usagePatch)
			}
		}
	}

	logger.Info("[TOOL-GORENAME] Rename analysis complete", "target", originalName, "newName", newName, "patchesGenerated", len(patchOperations))
	return patchOperations, nil
}

// --- Helper Functions ---

// isValidGoIdentifier checks if a string is a valid Go identifier.
func isValidGoIdentifier(name string) bool {
	// Fast path for empty string
	if name == "" {
		return false
	}
	// Use ParseExpr to leverage Go's lexer. We expect *ast.Ident if valid.
	expr, err := parser.ParseExprFrom(token.NewFileSet(), "", []byte(name), 0)
	if err != nil {
		return false // Parsing failed, not a valid expr/identifier
	}
	_, isIdent := expr.(*ast.Ident)
	// Ensure it's exactly an identifier (not composite like "a.b" which ParseExpr allows)
	return isIdent
}

// filterAndRelativizePath checks if absPath is within loadDir and returns the relative path or ""
func filterAndRelativizePath(absPath, loadDir string, logger logging.Logger, logPrefix string) (string, error) {
	relPathCheck, errCheck := filepath.Rel(loadDir, absPath)
	if errCheck != nil {
		logger.Warn(logPrefix+" Path relativization error", "absPath", absPath, "loadDir", loadDir, "error", errCheck)
		return "", errCheck // Return error if Rel fails
	}

	cleanLoadDir := filepath.Clean(loadDir)
	cleanAbsPath := filepath.Clean(absPath)

	// Check if cleanAbsPath starts with cleanLoadDir or is exactly cleanLoadDir
	if !strings.HasPrefix(cleanAbsPath, cleanLoadDir+string(filepath.Separator)) && cleanAbsPath != cleanLoadDir {
		logger.Debug(logPrefix+" Path outside indexed dir, filtering.", "path", absPath, "load_dir", cleanLoadDir)
		return "", nil // Return empty path and nil error to indicate filtering
	}

	return filepath.ToSlash(relPathCheck), nil // Return relative path
}
