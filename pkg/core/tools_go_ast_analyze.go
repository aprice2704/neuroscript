// filename: pkg/core/tools_go_ast_analyze.go
package core

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types" // Import go/types for type information
	"strings"
	// No need for astutil here anymore
)

// analyzeImportsAndSymbols - Analyzes AST using type information for usage of symbols
// from refactoredPkgPath and determines the new imports required based on the symbolMap.
func analyzeImportsAndSymbols(
	astFile *ast.File,
	fset *token.FileSet,
	refactoredPkgPath string,
	symbolMap map[string]string,
	typesInfo *types.Info, // <<< Added types.Info parameter
) (needsModification bool, requiredNewImports map[string]string, analysisErr error) {

	fmt.Printf("[DEBUG analyzeImportsAndSymbols] Analyzing file '%s' for target '%s' using type info.\n", fset.File(astFile.Pos()).Name(), refactoredPkgPath)

	// Input validation
	if typesInfo == nil {
		analysisErr = fmt.Errorf("internal error: typesInfo provided to analyzeImportsAndSymbols is nil")
		fmt.Printf("[ERROR analyzeImportsAndSymbols] %v\n", analysisErr)
		return false, nil, analysisErr
	}

	requiredNewImports = make(map[string]string) // {"new/path": ""}
	foundUsage := false
	foundImport := false // Track if the original import exists at all

	// --- Step 1: Check if the target import path exists in the file ---
	// (Still useful to know for determining skip reason later if no usage found)
	for _, impSpec := range astFile.Imports {
		if impSpec.Path != nil {
			if strings.Trim(impSpec.Path.Value, `"`) == refactoredPkgPath {
				foundImport = true
				break
			}
		}
	}

	// --- Step 2: Walk the AST using type info to find usages ---
	ast.Inspect(astFile, func(n ast.Node) bool {
		if analysisErr != nil {
			return false
		} // Stop if error occurred

		// Look for identifiers used as the 'X' in a selector expression 'X.Sel'
		if selExpr, ok := n.(*ast.SelectorExpr); ok {
			// Get the object corresponding to the 'X' part (the package identifier)
			obj := typesInfo.ObjectOf(selExpr.X.(*ast.Ident)) // Use ObjectOf for identifiers

			// Check if the object represents a PkgName
			if pkgName, ok := obj.(*types.PkgName); ok {
				// Get the actual imported package path
				importedPkg := pkgName.Imported()
				if importedPkg == nil {
					// Should not happen for valid code resolved by go/packages
					fmt.Printf("[WARN analyzeImportsAndSymbols] Found PkgName object for '%s' but Imported() is nil. Skipping.\n", pkgName.Name())
					return true // Continue inspection
				}
				resolvedPath := importedPkg.Path()

				// Compare the *resolved* path with our target path
				if resolvedPath == refactoredPkgPath {
					// This selector expression definitely refers to our target package!
					symbolName := selExpr.Sel.Name // The symbol being used

					fmt.Printf("[DEBUG analyzeImportsAndSymbols] Found usage via type info: %s.%s (resolved path: %s)\n", pkgName.Name(), symbolName, resolvedPath)

					// Look up the symbol in the map built from the *new* packages
					if newPath, mapped := symbolMap[symbolName]; mapped {
						// Symbol found in one of the new locations!
						fmt.Printf("[DEBUG analyzeImportsAndSymbols] Symbol '%s' mapped to new path '%s'\n", symbolName, newPath)
						requiredNewImports[newPath] = "" // Alias "" for now
						foundUsage = true

					} else {
						// Symbol resolved to the target package, but isn't in the map of moved symbols.
						// This likely means the symbol was *not* moved, or is unexported,
						// or it's a method call, or the symbol map is wrong.
						// For this tool's purpose (updating imports for *moved* symbols),
						// we should probably *not* treat this as an error, but rather just
						// ignore this specific usage - it doesn't require changing the import *for this symbol*.
						// However, if *other* symbols *are* moved, the import still needs changing.
						// If *no* symbols that *are* in the map are used, then the import *might*
						// be removable if this was the only reason it was imported.
						// Let's just log for now. The logic relies on finding *at least one* mapped symbol.
						fmt.Printf("[DEBUG analyzeImportsAndSymbols] Symbol '%s' from '%s' used, but not found in symbolMap. Ignoring usage.\n", symbolName, resolvedPath)

						// // --- OLD ERROR LOGIC ---
						// // fmt.Printf("[ERROR analyzeImportsAndSymbols] Symbol '%s' used with package '%s' but not found in the built symbol map!\n", symbolName, pkgName.Name())
						// // analysisErr = fmt.Errorf("%w: symbol '%s' used from package '%s' (%s) but not found in map of new locations", ErrSymbolNotFoundInMap, symbolName, pkgName.Name(), refactoredPkgPath)
						// // return false // Stop inspection on error
					}
				}
			} // end if pkgName
		} // end if selExpr
		return true // Continue inspection
	})
	// --- End AST Walk ---

	if analysisErr != nil {
		// This should only happen now if there's an internal error, not ErrSymbolNotFoundInMap
		fmt.Printf("[DEBUG analyzeImportsAndSymbols] Analysis failed: %v\n", analysisErr)
		return false, nil, analysisErr
	}

	// Determine final result:
	// We need modification only if the original import exists AND we found usage of at least one symbol that *is* in the symbolMap.
	if foundImport && foundUsage {
		fmt.Printf("[DEBUG analyzeImportsAndSymbols] Modification needed. Required new imports count: %d\n", len(requiredNewImports))
		return true, requiredNewImports, nil
	} else {
		// Either the import wasn't found, or no usage of *mapped* symbols was found.
		fmt.Printf("[DEBUG analyzeImportsAndSymbols] No modification needed (Import found: %v, Found mapped usage: %v)\n", foundImport, foundUsage)
		return false, nil, nil
	}
}
