// filename: pkg/core/tools_go_ast_package_helpers.go
package goast

import (
	"fmt"
	"go/ast"
	"go/token"
	"io/fs" // Needed for os.FileInfo
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// analyzeImportsAndSymbols identifies if the old import exists and which new imports are needed
// based on symbol usage. Returns: oldImportAliasOrName, needsMod, requiredNewImports, error
// (Doesn't need logging)
func analyzeImportsAndSymbols(astFile *ast.File, fset *token.FileSet, oldPath string, symbolMap map[string]string) (string, bool, map[string]string, error) {
	needsMod := false
	oldImportAliasOrName := ""
	oldImportFound := false
	requiredNewImports := make(map[string]string) // Map new path -> ""

	// Pass 1: Find old import and its alias/name
	for _, impSpec := range astFile.Imports {
		if impSpec.Path == nil {
			continue
		}
		importPath := strings.Trim(impSpec.Path.Value, `"`)
		if importPath == oldPath {
			oldImportFound = true
			if impSpec.Name != nil {
				oldImportAliasOrName = impSpec.Name.Name
			} else {
				parts := strings.Split(oldPath, "/")
				if len(parts) > 0 {
					oldImportAliasOrName = parts[len(parts)-1]
					oldImportAliasOrName = strings.ReplaceAll(oldImportAliasOrName, "-", "_")
					oldImportAliasOrName = strings.ReplaceAll(oldImportAliasOrName, ".", "_")
				} else {
					oldImportAliasOrName = oldPath
				}
			}
			break
		}
	}
	if !oldImportFound {
		return "", false, nil, nil
	}

	// Pass 2: Find usages of symbols from the old package alias/name
	ast.Inspect(astFile, func(node ast.Node) bool {
		selExpr, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, okX := selExpr.X.(*ast.Ident)
		if !okX || ident.Name != oldImportAliasOrName {
			return true
		}
		symbolName := selExpr.Sel.Name
		if newPath, exists := symbolMap[symbolName]; exists {
			needsMod = true
			requiredNewImports[newPath] = ""
		}
		return false // Stop descent into selector children
	})

	if !needsMod {
		return oldImportAliasOrName, false, nil, nil
	}
	return oldImportAliasOrName, true, requiredNewImports, nil
}

// applyAstImportChanges modifies the AST in place to remove the old import and add the new ones.
// Needs interpreter for logging.
func applyAstImportChanges(fset *token.FileSet, f *ast.File, oldImportPath string, requiredNewImports map[string]string, interpreter *Interpreter) error {
	// Use constant from main file (assuming same package)
	logPrefix := fmt.Sprintf("[applyAstImportChanges %s]", packageToolDebugVersion)
	logger := interpreter.logger // Use interpreter's logger
	logger.Printf("%s Applying import changes: remove %s, add %d new paths", logPrefix, oldImportPath, len(requiredNewImports))

	// Delete old import (try path first)
	deleted := astutil.DeleteImport(fset, f, oldImportPath)
	if deleted {
		logger.Printf("%s Successfully deleted import %s", logPrefix, oldImportPath)
	} else {
		// Try deleting by name if path failed
		var oldImportName string
		deletedNamed := false
		for _, impSpec := range f.Imports {
			if impSpec.Path != nil && strings.Trim(impSpec.Path.Value, `"`) == oldImportPath && impSpec.Name != nil {
				oldImportName = impSpec.Name.Name
				logger.Printf("%s Attempting to delete named import '%s' \"%s\".", logPrefix, oldImportName, oldImportPath)
				deletedNamed = astutil.DeleteNamedImport(fset, f, oldImportName, oldImportPath)
				if deletedNamed {
					break
				}
			}
		}
		if deletedNamed {
			logger.Printf("%s Successfully deleted named import '%s' %s", logPrefix, oldImportName, oldImportPath)
		} else {
			logger.Printf("%s [WARN] Could not find or delete import '%s'", logPrefix, oldImportPath)
		}
	}

	// Add new imports
	for newPath := range requiredNewImports {
		logger.Printf("%s Adding import: %s", logPrefix, newPath)
		added := astutil.AddImport(fset, f, newPath) // Handles name collisions
		if added {
			logger.Printf("%s Successfully added import %s", logPrefix, newPath)
		} else {
			logger.Printf("%s [INFO] Import '%s' was not added (likely already present).", logPrefix, newPath)
		}
	}

	// Sort and clean imports
	ast.SortImports(fset, f)
	return nil
}

// collectGoFiles walks the scan scope and collects paths to .go files, excluding specific directories.
// Needs interpreter for logging.
func collectGoFiles(scanScopeAbs, excludeDirAbs string, interpreter *Interpreter) ([]string, error) {
	// Use constant from main file (assuming same package)
	logPrefix := fmt.Sprintf("[collectGoFiles %s]", packageToolDebugVersion)
	goFilePaths := []string{}
	interpreter.logger.Printf("%s Starting file walk in '%s' (excluding '%s')", logPrefix, scanScopeAbs, excludeDirAbs)
	cleanedExcludeDir := filepath.Clean(excludeDirAbs)

	walkErr := filepath.WalkDir(scanScopeAbs, func(path string, d fs.DirEntry, walkErrInCb error) error {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(scanScopeAbs, path)
		}
		absPath = filepath.Clean(absPath)

		if walkErrInCb != nil { // Handle access errors
			interpreter.logger.Printf("%s [WARN] Error accessing path %q: %v", logPrefix, absPath, walkErrInCb)
			if d != nil && d.IsDir() {
				interpreter.logger.Printf("%s Skipping dir due to error: %s", logPrefix, absPath)
				return filepath.SkipDir
			}
			return nil // Skip entry, continue walk
		}

		if d.IsDir() { // Handle directories
			if absPath == cleanedExcludeDir {
				interpreter.logger.Printf("%s Skipping excluded directory: %s", logPrefix, absPath)
				return filepath.SkipDir
			}
			dirName := d.Name()
			if dirName == "vendor" || dirName == ".git" || dirName == "testdata" {
				interpreter.logger.Printf("%s Skipping special directory: %s", logPrefix, absPath)
				return filepath.SkipDir
			}
			return nil // Continue into directory
		}

		// Handle files
		fileName := d.Name()
		// Check if it's a Go file, not a test file, and not inside the excluded dir tree
		if strings.HasSuffix(fileName, ".go") &&
			!strings.HasSuffix(fileName, "_test.go") &&
			!strings.HasPrefix(absPath, cleanedExcludeDir+string(filepath.Separator)) {
			goFilePaths = append(goFilePaths, absPath)
		}
		return nil // Continue walk
	})

	if walkErr != nil {
		return nil, fmt.Errorf("file collection walk failed: %w", walkErr)
	}
	interpreter.logger.Printf("%s Collected %d Go files.", logPrefix, len(goFilePaths))
	return goFilePaths, nil
}
