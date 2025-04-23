// filename: pkg/core/tools_go_ast_symbol_map.go
package core

import (
	// Need errors for IsNotExist check
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// buildSymbolMap analyzes the sub-packages of a given original package path
// and creates a map of exported symbols to their new full import paths.
// It uses the go/packages library for robust package loading and analysis.
func buildSymbolMap(refactoredPkgPath string, interp *Interpreter) (map[string]string, error) {
	interp.logger.Printf("[buildSymbolMap] Building symbol map for refactored package: %s", refactoredPkgPath)
	symbolMap := make(map[string]string)

	if interp.sandboxDir == "" { /* ... */
		return nil, fmt.Errorf("%w: interpreter sandboxDir is empty", ErrSymbolMappingFailed)
	}

	modulePrefix := ""
	parts := strings.SplitN(refactoredPkgPath, "/", 2)
	if len(parts) == 2 {
		modulePrefix = parts[0] + "/"
	}
	relativePkgDir := strings.TrimPrefix(refactoredPkgPath, modulePrefix)
	baseScanDir := filepath.Join(interp.sandboxDir, filepath.FromSlash(relativePkgDir))
	interp.logger.Printf("[buildSymbolMap] Base directory for sub-package scan: %s", baseScanDir)
	if _, err := os.Stat(baseScanDir); os.IsNotExist(err) { /* ... */
		return nil, fmt.Errorf("%w: base directory '%s' for package '%s' not found", ErrRefactoredPathNotFound, baseScanDir, refactoredPkgPath)
	} else if err != nil { /* ... */
		return nil, fmt.Errorf("%w: error stating base directory '%s': %v", ErrSymbolMappingFailed, baseScanDir, err)
	}

	// Collect Go files and their intended sub-package relative paths
	type fileInfo struct {
		absPath   string
		subPkgRel string // e.g., "sub1", "sub2" relative to baseScanDir
	}
	goFilesToLoad := []fileInfo{}
	subDirs, err := os.ReadDir(baseScanDir)
	if err != nil { /* ... */
		return nil, fmt.Errorf("%w: failed to read base directory '%s': %v", ErrSymbolMappingFailed, baseScanDir, err)
	}
	for _, entry := range subDirs {
		if !entry.IsDir() {
			continue
		}
		subPkgName := entry.Name() // e.g., "sub1"
		subDirPath := filepath.Join(baseScanDir, subPkgName)
		filepath.WalkDir(subDirPath, func(path string, d os.DirEntry, walkErr error) error {
			if walkErr != nil {
				interp.logger.Printf("[WARN buildSymbolMap] Error accessing path %q: %v", path, walkErr)
				return nil
			}
			if !d.IsDir() && strings.HasSuffix(d.Name(), ".go") && !strings.HasSuffix(d.Name(), "_test.go") {
				absPath, pathErr := filepath.Abs(path)
				if pathErr != nil {
					interp.logger.Printf("[WARN buildSymbolMap] Failed to get abs path for %s: %v.", path, pathErr)
					return nil
				}
				goFilesToLoad = append(goFilesToLoad, fileInfo{absPath: absPath, subPkgRel: subPkgName})
			}
			return nil
		})
	}

	if len(goFilesToLoad) == 0 { /* ... */
		return symbolMap, nil
	}

	// Load the specific Go files
	loadPatterns := make([]string, 0, len(goFilesToLoad))
	for _, fi := range goFilesToLoad {
		loadPatterns = append(loadPatterns, "file="+fi.absPath)
	}
	interp.logger.Printf("[buildSymbolMap] Loading %d specific Go files...", len(loadPatterns))
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedImports, Fset: token.NewFileSet(), Dir: interp.sandboxDir}
	pkgs, err := packages.Load(cfg, loadPatterns...)
	if err != nil { /* ... */
		return nil, fmt.Errorf("%w: failed to load specific Go files: %v", ErrSymbolMappingFailed, err)
	}

	interp.logger.Printf("[buildSymbolMap] packages.Load returned %d packages.", len(pkgs))
	for i, pkg := range pkgs {
		interp.logger.Printf("[buildSymbolMap]   Loaded Pkg %d: ID='%s', Name='%s', PkgPath='%s', Files=%d", i, pkg.ID, pkg.Name, pkg.PkgPath, len(pkg.Syntax)) /* ... error logging ... */
	}

	loadErrors := []string{}
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			loadErrors = append(loadErrors, fmt.Sprintf("package %s (ID: %s): %v", pkg.PkgPath, pkg.ID, err))
		}
	})
	if len(loadErrors) > 0 { /* ... */
		return nil, fmt.Errorf("%w: errors loading packages: %s", ErrSymbolMappingFailed, strings.Join(loadErrors, "; "))
	}
	if len(pkgs) == 0 { /* ... */
		return symbolMap, nil
	}

	interp.logger.Printf("[buildSymbolMap] Successfully loaded %d package(s). Processing...", len(pkgs))

	foundSymbols := false
	ambiguousSymbols := make(map[string]string)
	processedFiles := make(map[string]bool) // Track processed files to derive subPkgRel once per file group

	for _, pkg := range pkgs {
		// --- FIX: Construct canonical path manually ---
		var canonicalPkgPath string
		var fileProcessedInThisPkg bool // Flag to find subPkgRel only once per pkg

		for _, goFile := range pkg.GoFiles {
			// Only process once per file, even if loaded multiple times by packages.Load
			if processedFiles[goFile] {
				continue
			}
			// Find the corresponding fileInfo to get the subPkgRel path
			var currentFi *fileInfo
			for i := range goFilesToLoad {
				if goFilesToLoad[i].absPath == goFile {
					currentFi = &goFilesToLoad[i]
					break
				}
			}

			if currentFi == nil {
				interp.logger.Printf("[WARN buildSymbolMap] Could not find original fileInfo for loaded file '%s'. Cannot determine canonical path.", goFile)
				continue // Skip files we can't map back
			}

			// Construct the expected import path
			// Ensure joining with "/" for import path syntax
			canonicalPkgPath = refactoredPkgPath + "/" + currentFi.subPkgRel
			interp.logger.Printf("[buildSymbolMap] Processing file '%s' belonging to constructed path '%s'", goFile, canonicalPkgPath)
			fileProcessedInThisPkg = true // Mark that we processed at least one file and determined path
			break                         // Found the subPkgRel for this group of files, assuming all files in pkg share it
		}

		if !fileProcessedInThisPkg || canonicalPkgPath == "" {
			interp.logger.Printf("[WARN buildSymbolMap] Could not determine canonical package path for pkg ID '%s'. Skipping.", pkg.ID)
			continue
		}
		// --- END FIX ---

		// Mark all files in this package as processed (using the paths from pkg.GoFiles)
		for _, goFile := range pkg.GoFiles {
			processedFiles[goFile] = true
		}

		// Extract exported symbols using the constructed canonicalPkgPath
		for _, fileNode := range pkg.Syntax {
			for _, decl := range fileNode.Decls {
				switch d := decl.(type) {
				case *ast.FuncDecl:
					if d.Name.IsExported() { /* ... map symbol ... */
						foundSymbols = true
						symbolName := d.Name.Name
						if existingPath, exists := symbolMap[symbolName]; exists && existingPath != canonicalPkgPath {
							ambiguousSymbols[symbolName] = fmt.Sprintf("found in %s and %s", existingPath, canonicalPkgPath)
						} else if !exists {
							symbolMap[symbolName] = canonicalPkgPath
						}
					}
				case *ast.GenDecl:
					for _, spec := range d.Specs {
						switch s := spec.(type) {
						case *ast.TypeSpec:
							if s.Name.IsExported() { /* ... map symbol ... */
								foundSymbols = true
								symbolName := s.Name.Name
								if existingPath, exists := symbolMap[symbolName]; exists && existingPath != canonicalPkgPath {
									ambiguousSymbols[symbolName] = fmt.Sprintf("found in %s and %s", existingPath, canonicalPkgPath)
								} else if !exists {
									symbolMap[symbolName] = canonicalPkgPath
								}
							}
						case *ast.ValueSpec:
							for _, name := range s.Names {
								if name.IsExported() { /* ... map symbol ... */
									foundSymbols = true
									symbolName := name.Name
									if existingPath, exists := symbolMap[symbolName]; exists && existingPath != canonicalPkgPath {
										ambiguousSymbols[symbolName] = fmt.Sprintf("found in %s and %s", existingPath, canonicalPkgPath)
									} else if !exists {
										symbolMap[symbolName] = canonicalPkgPath
									}
								}
							}
						}
					}
				}
			}
		}
	} // End loop through loaded packages

	// Final checks (ambiguity, no symbols found) and return logic remain the same...
	if len(ambiguousSymbols) > 0 { /* ... return ambiguity error ... */
		errorList := []string{}
		for symbol, locations := range ambiguousSymbols {
			errorList = append(errorList, fmt.Sprintf("symbol '%s' (%s)", symbol, locations))
		}
		errMsg := fmt.Sprintf("ambiguous exported symbols found: %s", strings.Join(errorList, "; "))
		interp.logger.Printf("[ERROR buildSymbolMap] %s", errMsg)
		return nil, fmt.Errorf("%w: %s", ErrSymbolMappingFailed, errMsg)
	}
	if !foundSymbols && len(goFilesToLoad) > 0 {
		interp.logger.Printf("[WARN buildSymbolMap] No exported symbols found in collected Go files under %s.", baseScanDir)
	}
	interp.logger.Printf("[buildSymbolMap] Finished building map. Total unique symbols found: %d", len(symbolMap))
	return symbolMap, nil
}
