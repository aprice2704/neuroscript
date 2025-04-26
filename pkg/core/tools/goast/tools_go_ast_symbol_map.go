// filename: pkg/core/tools_go_ast_symbol_map.go
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

	"github.com/aprice2704/neuroscript/pkg/core"
	"golang.org/x/mod/modfile"
)

// buildSymbolMap analyzes the sub-packages of a given original package path
// and creates a map of exported symbols to their new full import paths.
func buildSymbolMap(refactoredPkgPathRel string, interp *core.Interpreter) (map[string]string, error) {
	logPrefix := "[buildSymbolMap MANUAL]"
	interp.Logger().Debug("%s Building symbol map for relative package path: %s", logPrefix, refactoredPkgPathRel)
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
	interp.Logger().Debug("%s Base directory for sub-package scan: %s", logPrefix, absBaseScanDir)

	// --- Determine Module Path using modfile package ---
	modulePath := ""
	moduleRootDir := interp.SandboxDir()
	goModPath := filepath.Join(moduleRootDir, "go.mod")

	modContent, modErr := os.ReadFile(goModPath)
	if modErr == nil {
		modF, parseModErr := modfile.Parse(goModPath, modContent, nil)
		if parseModErr != nil {
			interp.Logger().Debug("%s [ERROR] Could not parse %s using modfile: %v", logPrefix, goModPath, parseModErr)
			return nil, fmt.Errorf("%w: failed to parse go.mod at %s using modfile: %w", core.ErrSymbolMappingFailed, goModPath, parseModErr)
		}
		if modF.Module != nil && modF.Module.Mod.Path != "" {
			modulePath = modF.Module.Mod.Path
			interp.Logger().Debug("%s Found module path from go.mod using modfile: %s", logPrefix, modulePath)
		} else {
			interp.Logger().Debug("%s [ERROR] Could not find module path declaration in parsed %s", logPrefix, goModPath)
			return nil, fmt.Errorf("%w: could not find module declaration in go.mod at %s", core.ErrSymbolMappingFailed, goModPath)
		}
	} else {
		interp.Logger().Debug("%s [ERROR] Could not read %s to determine module path: %v.", logPrefix, goModPath, modErr)
		return nil, fmt.Errorf("%w: failed to read go.mod at %s: %w", core.ErrSymbolMappingFailed, goModPath, modErr)
	}
	// --- End Module Path Determination ---

	fset := token.NewFileSet()

	// --- Function to process a directory ---
	processDirectory := func(dirPath string) error {
		// *** CALL DEBUG HELPER for Canonical Path ***
		canonicalPkgPath, pathErr := debugCalculateCanonicalPath(modulePath, moduleRootDir, dirPath, interp.Logger())
		if pathErr != nil {
			// Log the error from the helper and skip this directory
			interp.Logger().Debug("%s [WARN] Skipping directory '%s' due to canonical path error: %v", logPrefix, dirPath, pathErr)
			return nil // Don't stop the whole scan, just skip this dir
		}
		// *** END CALL DEBUG HELPER ***

		interp.Logger().Debug("%s Processing directory: %s (Canonical Path: %s)", logPrefix, dirPath, canonicalPkgPath)

		// Rest of the directory processing logic remains the same...
		pkgs, parseErr := parser.ParseDir(fset, dirPath, func(fi os.FileInfo) bool {
			return !fi.IsDir() && strings.HasSuffix(fi.Name(), ".go") && !strings.HasSuffix(fi.Name(), "_test.go")
		}, parser.ParseComments)

		if parseErr != nil {
			if !strings.Contains(parseErr.Error(), "no buildable Go source files") {
				interp.Logger().Debug("%s [WARN]core.Error parsing directory %s: %v. Skipping symbols.", logPrefix, dirPath, parseErr)
			} else {
				interp.Logger().Debug("%s [INFO] No buildable Go source files in %s.", logPrefix, dirPath)
			}
			return nil
		}

		for _, pkg := range pkgs {
			goFilesProcessed = true
			for fileName, astFile := range pkg.Files {
				interp.Logger().Debug("%s   Processing symbols in file: %s", logPrefix, fileName)
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
										interp.Logger().Debug("%s [WARN] AMBIGUITY DETECTED for %s '%s': %s", logPrefix, nodeType, symbolName, ambiguousSymbols[symbolName])
									}
								}
							} else {
								symbolMap[symbolName] = canonicalPkgPath
								interp.Logger().Debug("%s     Found exported %s: %s in %s", logPrefix, nodeType, symbolName, canonicalPkgPath)
							}
						}
					}

					switch decl := node.(type) {
					case *ast.FuncDecl:
						if decl.Recv == nil {
							checkAndAddSymbol(decl.Name, "func")
						}
						return false
					case *ast.GenDecl:
						for _, spec := range decl.Specs {
							switch specificSpec := spec.(type) {
							case *ast.TypeSpec:
								checkAndAddSymbol(specificSpec.Name, "type")
							case *ast.ValueSpec:
								for _, nameIdent := range specificSpec.Names {
									checkAndAddSymbol(nameIdent, "value")
								}
							}
						}
						return false
					}
					return true
				})
			}
		}
		return nil
	}

	// Process baseScanDir and subdirectories (logic unchanged)
	err = processDirectory(absBaseScanDir)
	if err != nil {
		interp.Logger().Debug("%s [ERROR] processing base directory %s: %v", logPrefix, absBaseScanDir, err)
	}

	var subDirs []fs.DirEntry
	subDirs, err = os.ReadDir(absBaseScanDir)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read base directory '%s': %v", core.ErrSymbolMappingFailed, absBaseScanDir, err)
	}

	for _, subEntry := range subDirs {
		if !subEntry.IsDir() {
			continue
		}
		subDirPath := filepath.Join(absBaseScanDir, subEntry.Name())
		err = processDirectory(subDirPath)
		if err != nil {
			interp.Logger().Debug("%s [ERROR] processing subdirectory %s: %v", logPrefix, subDirPath, err)
		}
	}

	// Final Ambiguity Check (unchanged)
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
		interp.Logger().Debug("%s [ERROR] %s", logPrefix, errMsg)
		return nil, fmt.Errorf("%w: %s", core.ErrAmbiguousSymbol, errMsg)
	}

	// Final Logging (unchanged)
	if !foundSymbols && goFilesProcessed {
		interp.Logger().Debug("%s [WARN] No exported symbols found in any Go files under %s and its subdirectories.", logPrefix, absBaseScanDir)
	} else if !goFilesProcessed {
		interp.Logger().Debug("%s [WARN] No Go files processed under %s and its subdirectories.", logPrefix, absBaseScanDir)
	}

	interp.Logger().Debug("%s Finished building map. Total unique symbols found: %d", logPrefix, len(symbolMap))
	return symbolMap, nil
}
