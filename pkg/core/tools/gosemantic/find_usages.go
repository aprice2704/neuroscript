// NeuroScript Version: 0.3.1
// File version: 0.0.4 // Use correct ReturnType constant: ArgTypeSliceMap.
// Implements the GoFindUsages tool.
// filename: pkg/core/tools/gosemantic/find_usages.go

package gosemantic

import (
	"errors"
	"fmt"
	"go/types"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // Import core
)

// --- Tool Definition: GoFindUsages ---

var toolGoFindUsagesImpl = core.ToolImplementation{
	Spec: core.ToolSpec{
		Name: "GoFindUsages",
		Description: "Finds all usages of the Go symbol identified by a semantic query string.\n" +
			"Requires a semantic index handle created by GoIndexCode.\n" +
			"Returns a list of locations where the symbol is used. The query format is the same as GoGetDeclarationOfSymbol.",
		Args: []core.ArgSpec{
			{Name: "index_handle", Type: core.ArgTypeString, Required: true, Description: "Handle returned by GoIndexCode."},
			{Name: "query", Type: core.ArgTypeString, Required: true, Description: "Semantic query string identifying the symbol (e.g., 'package:my/pkg; function:MyFunc')."},
		},
		// Returns a list of maps: [{"path": string, "line": int64, "column": int64, "name": string, "kind": string}]
		// *** FIXED: Use correct constant value 'ArgTypeSliceMap' from tools_types.go ***
		ReturnType: core.ArgTypeSliceMap,
	},
	Func: toolGoFindUsages,
}

func toolGoFindUsages(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	logger := interpreter.Logger()

	// --- Argument Parsing and Validation ---
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: GoFindUsages requires 2 arguments (index_handle, query)", core.ErrInvalidArgument)
	}
	handle, okH := args[0].(string)
	query, okQ := args[1].(string)

	if !okH || !okQ {
		return nil, fmt.Errorf("%w: GoFindUsages invalid argument types (expected string, string)", core.ErrInvalidArgument)
	}
	if query == "" {
		return nil, fmt.Errorf("%w: GoFindUsages query cannot be empty", core.ErrInvalidArgument)
	}

	logger.Debug("[TOOL-GOFINDUSAGES] Request", "handle", handle, "query", query)

	// --- Get Index ---
	indexValue, err := interpreter.GetHandleValue(handle, semanticIndexTypeTag)
	if err != nil {
		logger.Error("[TOOL-GOFINDUSAGES] Failed get handle", "handle", handle, "error", err)
		return nil, err
	}
	index, ok := indexValue.(*SemanticIndex)
	if !ok {
		logger.Error("[TOOL-GOFINDUSAGES] Handle not *SemanticIndex", "handle", handle, "type", fmt.Sprintf("%T", indexValue))
		return nil, fmt.Errorf("%w: handle '%s' is not a SemanticIndex", core.ErrHandleWrongType, handle)
	}
	if index.Fset == nil || index.Packages == nil {
		logger.Error("[TOOL-GOFINDUSAGES] Index FileSet or Packages nil", "handle", handle)
		return nil, fmt.Errorf("%w: index '%s' has nil FileSet or Packages", ErrIndexNotReady, handle)
	}

	// --- Find the Declaration Object based on the Query ---
	parsedQuery, parseErr := parseSemanticQuery(query)
	if parseErr != nil {
		logger.Error("[TOOL-GOFINDUSAGES] Failed to parse query", "query", query, "error", parseErr)
		if errors.Is(parseErr, ErrInvalidQueryFormat) {
			return nil, fmt.Errorf("%w: %w", core.ErrInvalidArgument, parseErr)
		}
		return nil, fmt.Errorf("failed to parse query '%s': %w", query, parseErr)
	}
	logger.Debug("[TOOL-GOFINDUSAGES] Parsed query", "parsed", parsedQuery)

	declarationObj, findErr := findObjectInIndex(index, parsedQuery, logger)
	if findErr != nil {
		if errors.Is(findErr, ErrSymbolNotFound) || errors.Is(findErr, ErrPackageNotFound) || errors.Is(findErr, ErrWrongKind) {
			logger.Warn("[TOOL-GOFINDUSAGES] Target symbol not found via query", "query", query, "error", findErr)
			return []interface{}{}, nil
		}
		logger.Error("[TOOL-GOFINDUSAGES] Error finding target symbol via query", "query", query, "error", findErr)
		return nil, findErr
	}
	if declarationObj == nil {
		logger.Error("[TOOL-GOFINDUSAGES] Internal inconsistency: findObjectInIndex returned nil object without error", "query", query)
		return nil, fmt.Errorf("%w: findObjectInIndex returned nil object unexpectedly for query '%s'", core.ErrInternal, query)
	}

	// Ignore PkgName objects if the query somehow resolved to one
	if pkgNameObj, isPkgName := declarationObj.(*types.PkgName); isPkgName {
		logger.Info("[TOOL-GOFINDUSAGES] Query resolved to a PkgName, cannot find usages.", "query", query, "pkgPath", pkgNameObj.Imported().Path())
		return []interface{}{}, nil
	}

	logger.Debug("[TOOL-GOFINDUSAGES] Identified declaration object to find usages for", "declObjName", declarationObj.Name(), "declObjType", fmt.Sprintf("%T", declarationObj), "declObjPos", index.Fset.Position(declarationObj.Pos()))

	// --- Core Logic: Find Usages across all indexed packages ---
	usages := make([]interface{}, 0)
	declKind := getObjectKind(declarationObj)

	logger.Debug("[TOOL-GOFINDUSAGES] Starting search across indexed packages", "numPackages", len(index.Packages))

	for _, pkgInfo := range index.Packages {
		if pkgInfo == nil || pkgInfo.TypesInfo == nil || pkgInfo.TypesInfo.Uses == nil {
			// Log skipping packages
			continue
		}
		logger.Debug("[TOOL-GOFINDUSAGES] Searching package", "pkgId", pkgInfo.ID, "numUses", len(pkgInfo.TypesInfo.Uses))

		for identNode, usedObj := range pkgInfo.TypesInfo.Uses {
			if usedObj == declarationObj {
				usagePos := identNode.Pos()
				if !usagePos.IsValid() {
					logger.Warn("[TOOL-GOFINDUSAGES] Found usage with invalid position", "pkgId", pkgInfo.ID, "identName", identNode.Name)
					continue
				}

				usagePosition := index.Fset.Position(usagePos)
				usageFilenameAbs := usagePosition.Filename
				if usageFilenameAbs == "" {
					logger.Warn("[TOOL-GOFINDUSAGES] Found usage with no filename", "pkgId", pkgInfo.ID, "identName", identNode.Name, "pos", usagePos)
					continue
				}

				// Filter: Ensure the usage is within the indexed directory scope
				relUsagePathCheck, errCheck := filepath.Rel(index.LoadDir, usageFilenameAbs)
				cleanLoadDir := filepath.Clean(index.LoadDir)
				cleanUsageFilenameAbs := filepath.Clean(usageFilenameAbs)
				isOutside := errCheck != nil || (!strings.HasPrefix(cleanUsageFilenameAbs, cleanLoadDir+string(filepath.Separator)) && cleanUsageFilenameAbs != cleanLoadDir)

				if isOutside {
					logger.Debug("[TOOL-GOFINDUSAGES] Filtering usage outside indexed dir", "usagePath", usageFilenameAbs, "loadDir", index.LoadDir)
					continue
				}

				// Add valid usage to results
				relUsagePath := filepath.ToSlash(relUsagePathCheck)
				usageMap := map[string]interface{}{
					"path":   relUsagePath,
					"line":   int64(usagePosition.Line),
					"column": int64(usagePosition.Column),
					"name":   identNode.Name,
					"kind":   declKind,
				}
				usages = append(usages, usageMap)
				logger.Debug("[TOOL-GOFINDUSAGES] Found valid usage", "pkgId", pkgInfo.ID, "usage", usageMap)
			}
		}
	}

	logger.Info("[TOOL-GOFINDUSAGES] Search complete", "target", declarationObj.Name(), "usagesFound", len(usages))
	return usages, nil
}
