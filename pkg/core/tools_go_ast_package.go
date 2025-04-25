// filename: pkg/core/tools_go_ast_package.go
package core

import (
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// --- FIXED VERSION ---
const packageToolDebugVersion = "v11_QUALIFIER_REWRITE_REVISED"

// analyzeImportsAndSymbols identifies if the old import exists and which new imports are needed
// based on symbol usage. Returns: oldImportAliasOrName, needsMod, requiredNewImports, error
// (Unchanged from previous version)
func analyzeImportsAndSymbols(astFile *ast.File, fset *token.FileSet, oldPath string, symbolMap map[string]string) (string, bool, map[string]string, error) {
	needsMod := false
	oldImportAliasOrName := ""
	oldImportFound := false
	requiredNewImports := make(map[string]string)

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

	// Pass 2: Find usages of symbols from the old package
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
		return false
	})

	if !needsMod {
		return oldImportAliasOrName, false, nil, nil
	}

	return oldImportAliasOrName, true, requiredNewImports, nil
}

// toolGoUpdateImportsForMovedPackage Tool - Uses AST analysis only.
func toolGoUpdateImportsForMovedPackage(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	logPrefix := fmt.Sprintf("[TOOL GoUpdateImports %s]", packageToolDebugVersion)
	interpreter.logger.Printf("%s ENTRY] Received args: %v", logPrefix, args)

	refactoredPkgPath := args[0].(string)
	scanScope := args[1].(string)

	interpreter.logger.Printf("%s Validated args: refactored_package_path='%s', scan_scope='%s'", logPrefix, refactoredPkgPath, scanScope)

	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		return map[string]interface{}{"error": "Interpreter sandbox directory is not set"}, nil
	}
	validatedScanScope, scopeErr := SecureFilePath(scanScope, sandboxRoot)
	if scopeErr != nil {
		errMsg := fmt.Sprintf("scan_scope validation failed: %s", scopeErr.Error())
		interpreter.logger.Printf("%s [ERROR] %s", logPrefix, errMsg)
		return map[string]interface{}{"error": errMsg}, nil
	}
	interpreter.logger.Printf("%s Validated scan scope (absolute): '%s'", logPrefix, validatedScanScope)

	modifiedFilesList := []string{}
	skippedFilesMap := make(map[string]string)
	failedFilesMap := make(map[string]string)
	var topLevelError error
	//fset := token.NewFileSet()

	interpreter.logger.Printf("%s === Calling buildSymbolMap (Manual) ===", logPrefix)
	symbolMap, err := buildSymbolMap(refactoredPkgPath, interpreter)
	interpreter.logger.Printf("%s === buildSymbolMap returned ===", logPrefix)
	if err != nil {
		errMsg := fmt.Sprintf("failed to build symbol map for '%s': %v", refactoredPkgPath, err) // Simplified error
		topLevelError = errors.New(errMsg)
		interpreter.logger.Printf("%s [ERROR] %s", logPrefix, topLevelError.Error())
		return map[string]interface{}{"error": topLevelError.Error()}, nil
	}
	interpreter.logger.Printf("%s Symbol map built successfully. Size: %d", logPrefix, len(symbolMap))
	if len(symbolMap) == 0 {
		interpreter.logger.Printf("%s [INFO] Symbol map is empty for '%s'. No files needed modification.", logPrefix, refactoredPkgPath)
		return map[string]interface{}{"modified_files": []interface{}{}, "skipped_files": map[string]interface{}{}, "failed_files": map[string]interface{}{}, "error": nil, "message": fmt.Sprintf("No exported symbols found in sub-packages of '%s'. No files needed modification.", refactoredPkgPath)}, nil
	}

	refactoredDirAbs := filepath.Join(sandboxRoot, filepath.FromSlash(refactoredPkgPath))
	interpreter.logger.Printf("%s Calculated refactored dir path (absolute): '%s'", logPrefix, refactoredDirAbs)

	goFilePaths, walkErr := collectGoFiles(validatedScanScope, refactoredDirAbs, interpreter)
	if walkErr != nil {
		topLevelError = fmt.Errorf("file collection failed: %w", walkErr)
		interpreter.logger.Printf("%s [ERROR] %s", logPrefix, topLevelError.Error())
		return map[string]interface{}{"error": topLevelError.Error()}, nil
	}
	interpreter.logger.Printf("%s Collected %d potentially relevant Go files.", logPrefix, len(goFilePaths))
	if len(goFilePaths) == 0 {
		interpreter.logger.Printf("%s No .go files found in scan scope to analyze. Exiting.", logPrefix)
		return map[string]interface{}{"modified_files": []interface{}{}, "skipped_files": map[string]interface{}{}, "failed_files": map[string]interface{}{}, "error": nil, "message": "No Go files found in the specified scan_scope (excluding the refactored package directory)."}, nil
	}

	interpreter.logger.Printf("%s === Parsing, Analyzing, and Modifying Files ===", logPrefix)
	for _, filePath := range goFilePaths {
		relPath, relErr := filepath.Rel(sandboxRoot, filePath)
		if relErr != nil {
			interpreter.logger.Printf("%s [WARN] Could not make path relative '%s': %v. Using absolute path.", logPrefix, filePath, relErr)
			relPath = filePath
		}
		relPathSlash := filepath.ToSlash(relPath)
		interpreter.logger.Printf("%s Processing file: %s", logPrefix, relPathSlash)

		// --- IMPORTANT: Use a new FileSet for each file to handle positions correctly after modification ---
		fileFset := token.NewFileSet()
		astFile, parseErr := parser.ParseFile(fileFset, filePath, nil, parser.ParseComments)
		if parseErr != nil {
			failReason := fmt.Sprintf("Failed to parse file: %v", parseErr)
			failedFilesMap[relPathSlash] = failReason
			interpreter.logger.Printf("%s [ERROR] %s: %s", logPrefix, failReason, relPathSlash)
			continue
		}

		interpreter.logger.Printf("%s Analyzing file: %s", logPrefix, relPathSlash)
		oldAlias, needsMod, requiredNewImports, analysisErr := analyzeImportsAndSymbols(astFile, fileFset, refactoredPkgPath, symbolMap)
		if analysisErr != nil {
			failReason := fmt.Sprintf("Analysis failed: %v", analysisErr)
			failedFilesMap[relPathSlash] = failReason
			interpreter.logger.Printf("%s [ERROR] Failed analysis for '%s': %s", logPrefix, relPathSlash, failReason)
			continue
		}

		if needsMod {
			// --- STEP 1: Modify Imports ---
			modifyErr := applyAstImportChanges(fileFset, astFile, refactoredPkgPath, requiredNewImports)
			if modifyErr != nil {
				failReason := fmt.Sprintf("Failed to apply AST import changes: %v", modifyErr)
				failedFilesMap[relPathSlash] = failReason
				interpreter.logger.Printf("%s [ERROR] Failed import modification for '%s': %s", logPrefix, relPathSlash, failReason)
				continue
			}
			interpreter.logger.Printf("%s Successfully applied import changes for %s", logPrefix, relPathSlash)

			// --- STEP 2: Rewrite Qualifiers ---
			rewriteOccurred := false
			// Create a map to find the required alias for newly added imports *if* base name conflicts
			importAliases := make(map[string]string) // path -> alias/name
			for _, imp := range astFile.Imports {
				if imp.Path == nil {
					continue
				}
				impPath := strings.Trim(imp.Path.Value, `"`)
				name := ""
				if imp.Name != nil {
					name = imp.Name.Name // Explicit alias
				} else {
					parts := strings.Split(impPath, "/")
					if len(parts) > 0 {
						name = parts[len(parts)-1] // Inferred name
						name = strings.ReplaceAll(name, "-", "_")
						name = strings.ReplaceAll(name, ".", "_")
					}
				}
				if name != "" {
					importAliases[impPath] = name
				}
			}

			postVisit := func(cursor *astutil.Cursor) bool {
				node := cursor.Node()
				selExpr, ok := node.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, okX := selExpr.X.(*ast.Ident)
				if !okX || ident.Name != oldAlias {
					return true
				} // Match using the original alias found for THIS file

				symbolName := selExpr.Sel.Name
				if newPath, exists := symbolMap[symbolName]; exists {
					// Determine the correct new package name/alias to use
					// Check if an alias was assigned automatically by AddImport if the base name conflicts
					newPkgName, aliasFound := importAliases[newPath]
					if !aliasFound {
						// Fallback: Infer from path if not found in import list (should not happen often after AddImport)
						parts := strings.Split(newPath, "/")
						if len(parts) > 0 {
							newPkgName = parts[len(parts)-1]
							newPkgName = strings.ReplaceAll(newPkgName, "-", "_")
							newPkgName = strings.ReplaceAll(newPkgName, ".", "_")
						} else {
							newPkgName = newPath // Should not happen
						}
						interpreter.logger.Printf("%s [WARN] Could not find alias for new import '%s' in AST, inferred '%s'", logPrefix, newPath, newPkgName)
					}

					if newPkgName == "" {
						interpreter.logger.Printf("%s [WARN] Inferred empty package name for new path '%s'. Skipping rewrite for '%s.%s'", logPrefix, newPath, oldAlias, symbolName)
						return false
					}

					// Only replace if the new name is different from the old one
					if newPkgName != oldAlias {
						interpreter.logger.Printf("   Rewriting '%s.%s' -> '%s.%s'", oldAlias, symbolName, newPkgName, symbolName)
						newIdent := ast.NewIdent(newPkgName)
						cursor.Replace(&ast.SelectorExpr{X: newIdent, Sel: selExpr.Sel})
						rewriteOccurred = true
					} else {
						// This case might happen if the refactored package name is the same as the original alias
						interpreter.logger.Printf("%s   Qualifier '%s.%s' already matches target '%s'. No rewrite needed.", logPrefix, oldAlias, symbolName, newPkgName)
					}
				}
				return false // Stop descent
			}
			astutil.Apply(astFile, nil, postVisit)

			if rewriteOccurred {
				interpreter.logger.Printf("%s Successfully rewrote qualifiers in %s", logPrefix, relPathSlash)
			} else {
				interpreter.logger.Printf("%s [INFO] No qualifiers needed rewriting in %s", logPrefix, relPathSlash)
			}
			// --- END STEP 2 ---

			// --- STEP 3: Format and Write Back ---
			var buf strings.Builder
			// Use the FileSet specific to this file for formatting
			formatErr := format.Node(&buf, fileFset, astFile) // Use fileFset
			if formatErr != nil {
				failReason := fmt.Sprintf("Failed to format modified AST: %v", formatErr)
				failedFilesMap[relPathSlash] = failReason
				interpreter.logger.Printf("%s [ERROR] Failed formatting for '%s': %s", logPrefix, relPathSlash, failReason)
				continue
			}
			info, statErr := os.Stat(filePath)
			perm := os.FileMode(0644)
			if statErr == nil {
				perm = info.Mode().Perm()
			} else {
				interpreter.logger.Printf("%s [WARN] Could not stat original file '%s' to get permissions: %v. Using default 0644.", logPrefix, filePath, statErr)
			}
			writeErr := os.WriteFile(filePath, []byte(buf.String()), perm)
			if writeErr != nil {
				failReason := fmt.Sprintf("Failed to write modified file: %v", writeErr)
				failedFilesMap[relPathSlash] = failReason
				interpreter.logger.Printf("%s [ERROR] Failed writing '%s': %s", logPrefix, relPathSlash, failReason)
			} else {
				modifiedFilesList = append(modifiedFilesList, relPathSlash)
				interpreter.logger.Printf("%s Successfully modified and wrote file '%s'", logPrefix, relPathSlash)
			}
		} else {
			skipReason := "Original package not imported"
			if oldAlias != "" {
				skipReason = "Original package imported, but no relevant symbols found/used"
			}
			skippedFilesMap[relPathSlash] = skipReason
			interpreter.logger.Printf("%s Skipped '%s': %s", logPrefix, relPathSlash, skipReason)
		}
	}

	// --- Return Results ---
	var finalErrorValue interface{}
	if topLevelError != nil {
		finalErrorValue = topLevelError.Error()
	}
	finalModifiedFiles := make([]interface{}, len(modifiedFilesList))
	for i, v := range modifiedFilesList {
		finalModifiedFiles[i] = v
	}
	finalSkippedFiles := make(map[string]interface{})
	for k, v := range skippedFilesMap {
		finalSkippedFiles[k] = v
	}
	finalFailedFiles := make(map[string]interface{})
	for k, v := range failedFilesMap {
		finalFailedFiles[k] = v
	}
	result := map[string]interface{}{
		"modified_files": finalModifiedFiles,
		"skipped_files":  finalSkippedFiles,
		"failed_files":   finalFailedFiles,
		"error":          finalErrorValue,
	}
	if len(symbolMap) == 0 && topLevelError == nil && len(modifiedFilesList) == 0 && len(failedFilesMap) == 0 {
		result["message"] = fmt.Sprintf("No exported symbols found in sub-packages of '%s'. No files needed modification.", refactoredPkgPath)
	}
	interpreter.logger.Printf("%s EXIT] Results: modified=%d, skipped=%d, failed=%d", // Corrected format string
		logPrefix, len(modifiedFilesList), len(skippedFilesMap), len(failedFilesMap))
	if finalErrorValue != nil {
		interpreter.logger.Printf("%s EXIT] Error: '%v'", logPrefix, finalErrorValue)
	}
	return result, nil
}

// --- Helper: collectGoFiles --- (remains unchanged)
func collectGoFiles(scanScopeAbs, excludeDirAbs string, interpreter *Interpreter) ([]string, error) {
	logPrefix := fmt.Sprintf("[collectGoFiles %s]", packageToolDebugVersion)
	goFilePaths := []string{}
	interpreter.logger.Printf("%s Starting file walk in '%s' (excluding '%s')", logPrefix, scanScopeAbs, excludeDirAbs)

	walkErr := filepath.WalkDir(scanScopeAbs, func(path string, d fs.DirEntry, walkErrInCb error) error {
		absPath := path
		if !filepath.IsAbs(path) {
			absPath = filepath.Join(scanScopeAbs, path)
		}
		if walkErrInCb != nil {
			interpreter.logger.Printf("%s [WARN] Error accessing path %q during walk: %v", logPrefix, absPath, walkErrInCb)
			if d != nil && d.IsDir() {
				interpreter.logger.Printf("%s Skipping directory due to error: %s", logPrefix, absPath)
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			cleanedPath := filepath.Clean(absPath)
			cleanedExcludeDir := filepath.Clean(excludeDirAbs)
			if cleanedPath == cleanedExcludeDir {
				interpreter.logger.Printf("%s Skipping excluded directory: %s", logPrefix, absPath)
				return filepath.SkipDir
			}
			if strings.HasPrefix(cleanedPath, cleanedExcludeDir+string(filepath.Separator)) {
				interpreter.logger.Printf("%s Skipping subdirectory of excluded dir: %s", logPrefix, absPath)
				return nil
			}
			dirName := d.Name()
			if dirName == "vendor" || dirName == ".git" || dirName == "testdata" {
				interpreter.logger.Printf("%s Skipping special directory: %s", logPrefix, absPath)
				return filepath.SkipDir
			}
			return nil
		}
		fileName := d.Name()
		if strings.HasSuffix(fileName, ".go") && !strings.HasSuffix(fileName, "_test.go") {
			cleanedPath := filepath.Clean(absPath)
			cleanedExcludeDir := filepath.Clean(excludeDirAbs)
			if !strings.HasPrefix(cleanedPath, cleanedExcludeDir+string(filepath.Separator)) && cleanedPath != cleanedExcludeDir {
				goFilePaths = append(goFilePaths, absPath)
			}
		}
		return nil
	})

	if walkErr != nil {
		return nil, fmt.Errorf("file collection walk failed: %w", walkErr)
	}

	interpreter.logger.Printf("%s Collected %d Go files.", logPrefix, len(goFilePaths))
	return goFilePaths, nil
}

// --- Registration --- (remains unchanged)
func registerGoAstPackageTools(registry *ToolRegistry) error {
	err := registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GoUpdateImportsForMovedPackage",
			Description: fmt.Sprintf("Version: %s. Analyzes Go files within a scope, updating import paths AND code qualifiers for symbols moved from an 'original' package path into its sub-packages. Uses AST analysis and a symbol map.", packageToolDebugVersion),
			Args: []ArgSpec{
				{Name: "refactored_package_path", Type: ArgTypeString, Required: true, Description: "The original import path that now contains sub-packages (e.g., 'github.com/org/repo/original')."},
				{Name: "scan_scope", Type: ArgTypeString, Required: true, Description: "Directory path (relative to sandbox root) to scan for Go files (e.g., '.', './cmd/app')."},
			},
			ReturnType: ArgTypeAny, // map[string]interface{}
		},
		Func: toolGoUpdateImportsForMovedPackage,
	})
	if err != nil {
		return fmt.Errorf("failed to register tool GoUpdateImportsForMovedPackage: %w", err)
	}
	return nil
}
