// filename: pkg/core/tools_go_ast_symbol_map.go
package core

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// buildSymbolMap analyzes the sub-packages of a given original package path
// and creates a map of exported symbols to their new full import paths.
// Uses direct AST parsing instead of packages.Load for simplicity in test environments.
func buildSymbolMap(refactoredPkgPath string, interp *Interpreter) (map[string]string, error) {
	interp.logger.Printf("[buildSymbolMap MANUAL] Building symbol map for package path: %s", refactoredPkgPath)
	symbolMap := make(map[string]string)
	ambiguousSymbols := make(map[string]string) // Stores symbol -> "found in path1 and path2"
	foundSymbols := false
	goFilesProcessed := false

	if interp.sandboxDir == "" {
		return nil, fmt.Errorf("%w: interpreter sandboxDir is empty", ErrSymbolMappingFailed)
	}

	baseScanDir := filepath.Join(interp.sandboxDir, filepath.FromSlash(refactoredPkgPath))
	interp.logger.Printf("[buildSymbolMap MANUAL] Base directory for sub-package scan: %s", baseScanDir)

	dirInfo, err := os.Stat(baseScanDir)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: base directory '%s' corresponding to package '%s' not found", ErrRefactoredPathNotFound, baseScanDir, refactoredPkgPath)
	} else if err != nil {
		return nil, fmt.Errorf("%w: error stating base directory '%s': %v", ErrSymbolMappingFailed, baseScanDir, err)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("%w: path '%s' is not a directory", ErrSymbolMappingFailed, baseScanDir)
	}

	subDirs, err := os.ReadDir(baseScanDir)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read base directory '%s': %v", ErrSymbolMappingFailed, baseScanDir, err)
	}

	fset := token.NewFileSet()

	for _, subEntry := range subDirs {
		if !subEntry.IsDir() {
			continue
		}
		subPkgName := subEntry.Name()
		subDirPath := filepath.Join(baseScanDir, subPkgName)

		// Construct canonicalPkgPath (forward slashes for import paths)
		canonicalPkgPath := refactoredPkgPath + "/" + subPkgName

		interp.logger.Printf("[buildSymbolMap MANUAL] Scanning subdir: %s (Canonical Path: %s)", subDirPath, canonicalPkgPath)

		filesInSubDir, err := os.ReadDir(subDirPath)
		if err != nil {
			interp.logger.Printf("[WARN buildSymbolMap MANUAL] Failed to read subdir %s: %v. Skipping.", subDirPath, err)
			continue
		}

		pkgHasGoFiles := false
		for _, fileEntry := range filesInSubDir {
			if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".go") || strings.HasSuffix(fileEntry.Name(), "_test.go") {
				continue
			}
			pkgHasGoFiles = true
			goFilesProcessed = true
			filePath := filepath.Join(subDirPath, fileEntry.Name())
			interp.logger.Printf("[buildSymbolMap MANUAL]   Parsing file: %s", filePath)

			astFile, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments) // ParseComments might be useful later if needed
			if err != nil {
				interp.logger.Printf("[WARN buildSymbolMap MANUAL] Failed to parse file %s: %v. Skipping symbols from this file.", filePath, err)
				continue
			}

			// --- Inspect AST for EXPORTED declarations ---
			ast.Inspect(astFile, func(node ast.Node) bool {
				var symbolName string
				var isExported bool

				// Identify potential exported symbols
				switch d := node.(type) {
				case *ast.FuncDecl:
					// Only consider top-level functions (no methods)
					if d.Recv == nil && d.Name != nil {
						symbolName = d.Name.Name
						isExported = d.Name.IsExported()
					}
					// Don't traverse into function bodies
					return false // Stop descent for FuncDecl
				case *ast.TypeSpec:
					// Consider top-level type definitions
					if d.Name != nil {
						symbolName = d.Name.Name
						isExported = d.Name.IsExported()
					}
					// Don't traverse into type specs further (like struct fields)
					return false // Stop descent for TypeSpec
				case *ast.ValueSpec:
					// Consider top-level variables and constants
					for _, name := range d.Names {
						if name != nil && name.IsExported() {
							// Handle potential multiple declarations (var a, b int)
							valSymbolName := name.Name
							interp.logger.Printf("[buildSymbolMap MANUAL]     Found exported value spec: %s in %s", valSymbolName, canonicalPkgPath)
							// Process each exported name from ValueSpec immediately
							// This ensures ambiguity is checked correctly even within a single ValueSpec
							foundSymbols = true
							if existingPath, exists := symbolMap[valSymbolName]; exists {
								if existingPath != canonicalPkgPath {
									// Ambiguity detected across packages
									if _, ambigExists := ambiguousSymbols[valSymbolName]; !ambigExists {
										ambiguousSymbols[valSymbolName] = fmt.Sprintf("found in %s and %s", existingPath, canonicalPkgPath)
										interp.logger.Printf("[WARN buildSymbolMap MANUAL] AMBIGUITY DETECTED for symbol '%s': %s", valSymbolName, ambiguousSymbols[valSymbolName])
									}
								}
								// If exists and same path, do nothing (already recorded)
							} else {
								// Symbol not seen before, add it
								symbolMap[valSymbolName] = canonicalPkgPath
							}
						}
					}
					// Don't traverse into value specs further
					return false // Stop descent for ValueSpec
				default:
					// Continue traversal for other node types
					return true
				}

				// --- FIX: Process identified FuncDecl/TypeSpec symbols HERE ---
				if symbolName != "" && isExported {
					interp.logger.Printf("[buildSymbolMap MANUAL]     Found exported %T: %s in %s", node, symbolName, canonicalPkgPath)
					foundSymbols = true
					if existingPath, exists := symbolMap[symbolName]; exists {
						if existingPath != canonicalPkgPath {
							// Ambiguity detected across packages
							if _, ambigExists := ambiguousSymbols[symbolName]; !ambigExists {
								ambiguousSymbols[symbolName] = fmt.Sprintf("found in %s and %s", existingPath, canonicalPkgPath)
								interp.logger.Printf("[WARN buildSymbolMap MANUAL] AMBIGUITY DETECTED for symbol '%s': %s", symbolName, ambiguousSymbols[symbolName])
							}
						}
						// If exists and same path, do nothing (already recorded)
					} else {
						// Symbol not seen before, add it
						symbolMap[symbolName] = canonicalPkgPath
					}
					// For FuncDecl/TypeSpec, we already returned false, so no further descent needed.
				}

				// This point is reached only if the switch didn't handle the node
				// or didn't explicitly return false (e.g. for non-exported items).
				// We generally want to continue searching the file.
				return true
			}) // End ast.Inspect
		} // End file loop

		if !pkgHasGoFiles {
			interp.logger.Printf("[buildSymbolMap MANUAL]   No Go files found in %s", subDirPath)
		}
	} // End subdir loop

	// --- Check for Ambiguity ---
	if len(ambiguousSymbols) > 0 {
		errorList := []string{}
		// Sort symbols for consistent error messages
		sortedSymbols := make([]string, 0, len(ambiguousSymbols))
		for symbol := range ambiguousSymbols {
			sortedSymbols = append(sortedSymbols, symbol)
		}
		sort.Strings(sortedSymbols) // Requires "sort" import

		for _, symbol := range sortedSymbols {
			locations := ambiguousSymbols[symbol]
			errorList = append(errorList, fmt.Sprintf("symbol '%s' (%s)", symbol, locations))
		}
		errMsg := fmt.Sprintf("ambiguous exported symbols found: %s", strings.Join(errorList, "; "))
		interp.logger.Printf("[ERROR buildSymbolMap MANUAL] %s", errMsg)
		// Return the map containing non-ambiguous symbols found *before* ambiguity was detected?
		// Or return nil? Returning nil seems safer as the map is incomplete/unreliable.
		// Also wrap the specific error message within the general failure error.
		return nil, fmt.Errorf("%w: %w", ErrSymbolMappingFailed, errors.New(errMsg))
	}

	if !foundSymbols && goFilesProcessed {
		interp.logger.Printf("[WARN buildSymbolMap MANUAL] No exported symbols found in any Go files under %s.", baseScanDir)
		// Return empty map, not an error
	} else if !goFilesProcessed {
		interp.logger.Printf("[WARN buildSymbolMap MANUAL] No .go files processed under %s.", baseScanDir)
		// Return empty map, not an error
	}

	interp.logger.Printf("[buildSymbolMap MANUAL] Finished building map. Total unique symbols found: %d", len(symbolMap))
	return symbolMap, nil
}
