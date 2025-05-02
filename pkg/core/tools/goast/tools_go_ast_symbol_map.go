// filename: pkg/core/tools_go_ast_symbol_map.go
// UPDATED: Call core.FindAndParseGoMod (Exported version)
package goast

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core" // Added core import
	"golang.org/x/mod/modfile"
)

// buildSymbolMap analyzes the sub-packages of a given original package path
// and creates a map of exported symbols to their new full import paths.
func buildSymbolMap(refactoredPkgPathRel string, interp *core.Interpreter) (map[string]string, error) {
	logPrefix := "[buildSymbolMap MANUAL]"
	logger := interp.Logger() // Use logger directly
	logger.Debug("%s Building symbol map for relative package path: %s", logPrefix, refactoredPkgPathRel)
	symbolMap := make(map[string]string)
	ambiguousSymbols := make(map[string]string)
	foundSymbols := false
	goFilesProcessed := false
	var err error // Declare err here for broader scope

	if interp.SandboxDir() == "" {
		return nil, fmt.Errorf("%w: interpreter sandboxDir is empty", core.ErrSymbolMappingFailed)
	}

	// --- Path Validation ---
	absBaseScanDir, secErr := core.SecureFilePath(refactoredPkgPathRel, interp.SandboxDir())
	if secErr != nil {
		if errors.Is(secErr, core.ErrPathViolation) {
			return nil, fmt.Errorf("%w: %w", core.ErrInvalidPath, secErr)
		}
		return nil, fmt.Errorf("%w: security error validating refactored package path '%s': %w", core.ErrSymbolMappingFailed, refactoredPkgPathRel, secErr)
	}

	dirInfo, statErr := os.Stat(absBaseScanDir)
	if errors.Is(statErr, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: base directory '%s' corresponding to package '%s' not found", core.ErrRefactoredPathNotFound, absBaseScanDir, refactoredPkgPathRel)
	} else if statErr != nil {
		return nil, fmt.Errorf("%w: error stating base directory '%s': %v", core.ErrSymbolMappingFailed, absBaseScanDir, statErr)
	}
	if !dirInfo.IsDir() {
		return nil, fmt.Errorf("%w: path '%s' is not a directory", core.ErrSymbolMappingFailed, absBaseScanDir)
	}
	logger.Debug("%s Base directory for sub-package scan: %s", logPrefix, absBaseScanDir)

	// --- Determine Module Path using the core helper ---
	var modF *modfile.File // To hold the parsed file info if needed later
	var modulePath string
	var moduleRootDir string // Directory containing the go.mod

	// Start search from the directory being scanned
	modF, moduleRootDir, err = core.FindAndParseGoMod(absBaseScanDir, logger) // <-- Use exported name
	if err != nil {
		// Check if it was just not found
		if errors.Is(err, os.ErrNotExist) {
			logger.Debug("%s [WARN] Could not find go.mod in or above '%s'. Cannot determine module path.", logPrefix, absBaseScanDir)
			return nil, fmt.Errorf("%w: could not find go.mod required for symbol mapping: %v", core.ErrSymbolMappingFailed, err)
		}
		// Handle other errors (read, parse)
		logger.Debug("%s [ERROR] Error finding/parsing go.mod: %v", logPrefix, err)
		return nil, fmt.Errorf("%w: failed to find or parse go.mod: %w", core.ErrSymbolMappingFailed, err)
	}

	// Successfully found and parsed go.mod
	if modF != nil && modF.Module != nil && modF.Module.Mod.Path != "" {
		modulePath = modF.Module.Mod.Path
		logger.Debug("%s Found module path '%s' from go.mod at '%s'", logPrefix, modulePath, moduleRootDir)
	} else {
		goModPath := filepath.Join(moduleRootDir, "go.mod") // Reconstruct path for logging
		logger.Debug("%s [ERROR] Could not find module path declaration in parsed %s", logPrefix, goModPath)
		return nil, fmt.Errorf("%w: could not find module declaration in go.mod at %s", core.ErrSymbolMappingFailed, goModPath)
	}
	// --- End Module Path Determination ---

	fset := token.NewFileSet()

	// --- Function to process a directory ---
	// --- (processDirectory function remains unchanged from previous version) ---
	processDirectory := func(dirPath string) error {
		// *** CALL DEBUG HELPER for Canonical Path ***
		// Now passing the dynamically found moduleRootDir
		canonicalPkgPath, pathErr := debugCalculateCanonicalPath(modulePath, moduleRootDir, dirPath, logger)
		if pathErr != nil {
			// Log the error from the helper and skip this directory
			logger.Debug("%s [WARN] Skipping directory '%s' due to canonical path error: %v", logPrefix, dirPath, pathErr)
			return nil // Don't stop the whole scan, just skip this dir
		}
		// *** END CALL DEBUG HELPER ***

		logger.Debug("%s Processing directory: %s (Canonical Path: %s)", logPrefix, dirPath, canonicalPkgPath)

		// Rest of the directory processing logic remains the same...
		pkgs, parseErr := parser.ParseDir(fset, dirPath, func(fi os.FileInfo) bool {
			return !fi.IsDir() && strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
		}, parser.ParseComments)

		if parseErr != nil {
			// Use errors.Is for more robust checking if specific error types are expected
			if strings.Contains(parseErr.Error(), "no buildable Go source files") || errors.Is(parseErr, fs.ErrNotExist) { // Handle case where dir might vanish between checks
				logger.Debug("%s [INFO] No buildable Go source files in %s or directory not found.", logPrefix, dirPath)
			} else {
				logger.Debug("%s [WARN] Error parsing directory %s: %v. Skipping symbols.", logPrefix, dirPath, parseErr)
			}
			return nil // Continue scanning other directories
		}

		for _, pkg := range pkgs {
			goFilesProcessed = true
			for fileName, astFile := range pkg.Files {
				logger.Debug("%s   Processing symbols in file: %s", logPrefix, fileName)
				ast.Inspect(astFile, func(node ast.Node) bool {
					checkAndAddSymbol := func(ident *ast.Ident, nodeType string) {
						if ident != nil && ident.IsExported() {
							symbolName := ident.Name
							foundSymbols = true
							if existingPath, exists := symbolMap[symbolName]; exists {
								cleanedExisting := path.Clean(existingPath)
								cleanedCurrent := path.Clean(canonicalPkgPath)
								if cleanedExisting != cleanedCurrent {
									if _, ambigExists := ambiguousSymbols[symbolName]; !ambigExists {
										ambiguousSymbols[symbolName] = fmt.Sprintf("found in %s and %s", existingPath, canonicalPkgPath)
										logger.Debug("%s [WARN] AMBIGUITY DETECTED for %s '%s': %s", logPrefix, nodeType, symbolName, ambiguousSymbols[symbolName])
									}
								}
							} else {
								symbolMap[symbolName] = canonicalPkgPath
								logger.Debug("%s     Found exported %s: %s in %s", logPrefix, nodeType, symbolName, canonicalPkgPath)
							}
						}
					}

					switch decl := node.(type) {
					case *ast.FuncDecl:
						// Check if it's a function (not a method)
						if decl.Recv == nil {
							checkAndAddSymbol(decl.Name, "func")
						}
						// No need to recurse into function bodies for top-level symbols
						return false // Stop descent for this node
					case *ast.GenDecl:
						// Handle top-level var, const, type declarations
						for _, spec := range decl.Specs {
							switch specificSpec := spec.(type) {
							case *ast.TypeSpec: // Type declarations (struct, interface, etc.)
								checkAndAddSymbol(specificSpec.Name, "type")
							case *ast.ValueSpec: // Var/Const declarations
								for _, nameIdent := range specificSpec.Names {
									checkAndAddSymbol(nameIdent, "value")
								}
							}
						}
						// No need to recurse further into these declarations
						return false // Stop descent for this node
					}
					// Continue inspection for other node types if necessary
					return true
				})
			}
		}
		return nil
	}

	// Process baseScanDir and subdirectories (logic unchanged, uses filepath.WalkDir)
	// --- (filepath.WalkDir loop remains unchanged from previous version) ---
	err = processDirectory(absBaseScanDir)
	if err != nil {
		logger.Debug("%s [ERROR] processing base directory %s: %v", logPrefix, absBaseScanDir, err)
		// Decide if this error should halt the entire process or just be logged
		// For now, let's log it and continue with subdirs
	}

	// Walk the directory tree instead of just immediate subdirs
	walkErr := filepath.WalkDir(absBaseScanDir, func(currentPath string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			logger.Debug("%s [WARN] Error accessing path %q during walk: %v", logPrefix, currentPath, walkErr)
			return walkErr // Propagate error if needed, or return nil to continue
		}
		// Skip the root directory itself (already processed)
		if currentPath == absBaseScanDir {
			return nil
		}
		// Skip non-directories and hidden directories (like .git)
		if !d.IsDir() || strings.HasPrefix(d.Name(), ".") {
			// If it's a hidden dir, skip traversing into it
			if d.IsDir() && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}
			return nil // Skip files or hidden non-dirs
		}

		// Process this subdirectory
		err = processDirectory(currentPath)
		if err != nil {
			logger.Debug("%s [ERROR] processing subdirectory %s: %v", logPrefix, currentPath, err)
			// Decide whether to stop the walk on error or continue
		}
		return nil // Continue walking
	})

	if walkErr != nil {
		logger.Debug("%s [ERROR] Error walking directory tree starting at %s: %v", logPrefix, absBaseScanDir, walkErr)
		// Return an error if the walk failed significantly
		return nil, fmt.Errorf("%w: error walking directory '%s': %w", core.ErrSymbolMappingFailed, absBaseScanDir, walkErr)
	}

	// Final Ambiguity Check (unchanged)
	// --- (Ambiguity check logic remains unchanged) ---
	if len(ambiguousSymbols) > 0 {
		errorList := []string{}
		sortedSymbols := make([]string, 0, len(ambiguousSymbols))
		for symbol := range ambiguousSymbols {
			sortedSymbols = append(sortedSymbols, symbol)
		}
		sort.Strings(sortedSymbols)
		for _, symbol := range sortedSymbols {
			locations := ambiguousSymbols[symbol]
			errorList = append(errorList, fmt.Sprintf("symbol '%s' (%s)", symbol, locations))
		}
		errMsg := fmt.Sprintf("ambiguous exported symbols found: %s", strings.Join(errorList, "; "))
		logger.Debug("%s [ERROR] %s", logPrefix, errMsg)
		return nil, fmt.Errorf("%w: %s", core.ErrAmbiguousSymbol, errMsg)
	}

	// Final Logging (unchanged)
	// --- (Final logging remains unchanged) ---
	if !foundSymbols && goFilesProcessed {
		logger.Debug("%s [WARN] No exported symbols found in any Go files under %s and its subdirectories.", logPrefix, absBaseScanDir)
	} else if !goFilesProcessed {
		logger.Debug("%s [WARN] No Go files processed under %s and its subdirectories.", logPrefix, absBaseScanDir)
	}

	logger.Debug("%s Finished building map. Total unique symbols found: %d", logPrefix, len(symbolMap))
	return symbolMap, nil
}
