// filename: pkg/core/tools_go_ast_package.go
package goast

import (
	"bytes"
	"errors"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser" // Needed for deep copy via print/reparse
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/core"
	"golang.org/x/tools/go/ast/astutil"
)

// --- Centralized Version Constant ---
const packageToolDebugVersion = "v13_ERR_NILMAP_TESTFIX"

// toolGoUpdateImportsForMovedPackage Tool - Uses AST analysis only.
// v13: Returns nil map on error. Test cases need updating. Uses helpers.
func toolGoUpdateImportsForMovedPackage(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	logPrefix := fmt.Sprintf("[TOOL GoUpdateImports %s]", packageToolDebugVersion)
	interpreter.Logger().Debug("%s ENTRY] Received args: %v", logPrefix, args)

	// --- Argument Parsing and Validation ---
	if len(args) != 2 {
		return nil, fmt.Errorf("expected 2 arguments, got %d", len(args))
	}
	refactoredPkgPath, ok1 := args[0].(string)
	scanScope, ok2 := args[1].(string)
	if !ok1 || !ok2 {
		return nil, fmt.Errorf("invalid argument types: expected (string, string)")
	}
	if refactoredPkgPath == "" || scanScope == "" {
		return nil, fmt.Errorf("arguments cannot be empty")
	}
	interpreter.Logger().Debug("%s Validated args: refactored_package_path='%s', scan_scope='%s'", logPrefix, refactoredPkgPath, scanScope)

	sandboxRoot := interpreter.SandboxDir()
	if sandboxRoot == "" {
		return nil, errors.New("interpreter sandbox directory not set")
	}
	validatedScanScope, scopeErr := core.SecureFilePath(scanScope, sandboxRoot)
	if scopeErr != nil {
		err := fmt.Errorf("scan_scope validation failed: %w", scopeErr)
		interpreter.Logger().Debug("%s [ERROR] %s", logPrefix, err)
		return nil, err
	}
	interpreter.Logger().Debug("%s Validated scan scope (absolute): '%s'", logPrefix, validatedScanScope)

	// --- Initialize Result Tracking ---
	modifiedFilesList := []string{}
	skippedFilesMap := make(map[string]string)
	failedFilesMap := make(map[string]string)
	var firstProcessingError error

	// --- Build Symbol Map ---
	interpreter.Logger().Debug("%s === Calling buildSymbolMap (Manual) ===", logPrefix)
	// Assuming buildSymbolMap is defined elsewhere (e.g., tools_go_ast_symbol_map.go)
	symbolMap, buildMapErr := buildSymbolMap(refactoredPkgPath, interpreter)
	interpreter.Logger().Debug("%s === buildSymbolMap returned ===", logPrefix)
	if buildMapErr != nil {
		err := fmt.Errorf("failed to build symbol map for '%s': %w", refactoredPkgPath, buildMapErr)
		interpreter.Logger().Debug("%s [ERROR] %s", logPrefix, err)
		return nil, err // Return nil map, error
	}
	interpreter.Logger().Debug("%s Symbol map built successfully. Size: %d", logPrefix, len(symbolMap))
	if len(symbolMap) == 0 { // Handle empty symbol map (success, but nothing else to do)
		message := fmt.Sprintf("No exported symbols in '%s'. No files needed modification.", refactoredPkgPath)
		interpreter.Logger().Debug("%s [INFO] %s", logPrefix, message)
		return map[string]interface{}{"modified_files": []interface{}{}, "skipped_files": map[string]interface{}{}, "failed_files": map[string]interface{}{}, "error": nil, "message": message}, nil
	}

	// --- Collect Go Files (using helper) ---
	refactoredDirAbs := filepath.Join(sandboxRoot, filepath.FromSlash(refactoredPkgPath))
	interpreter.Logger().Debug("%s Calculated refactored dir path (absolute): '%s'", logPrefix, refactoredDirAbs)
	goFilePaths, walkErr := collectGoFiles(validatedScanScope, refactoredDirAbs, interpreter) // Call helper
	if walkErr != nil {
		interpreter.Logger().Debug("%s [ERROR] %s", logPrefix, walkErr)
		return nil, walkErr
	} // Return nil map, error
	interpreter.Logger().Debug("%s Collected %d potentially relevant Go files.", logPrefix, len(goFilePaths))
	if len(goFilePaths) == 0 { // Handle no files found (success, but nothing else to do)
		message := "No Go files found in scan_scope (excluding refactored package)."
		interpreter.Logger().Debug("%s %s Exiting.", logPrefix, message)
		return map[string]interface{}{"modified_files": []interface{}{}, "skipped_files": map[string]interface{}{}, "failed_files": map[string]interface{}{}, "error": nil, "message": message}, nil
	}

	// --- Process Each Go File ---
	interpreter.Logger().Debug("%s === Parsing, Analyzing, and Modifying Files ===", logPrefix)
	for _, filePath := range goFilePaths {
		relPath, relErr := filepath.Rel(sandboxRoot, filePath)
		if relErr != nil {
			relPath = filePath
		}
		relPathSlash := filepath.ToSlash(relPath)
		interpreter.Logger().Debug("%s Processing file: %s", logPrefix, relPathSlash)
		fileFset := token.NewFileSet() // Use new FileSet for each file
		astFile, parseErr := parser.ParseFile(fileFset, filePath, nil, parser.ParseComments)
		if parseErr != nil {
			failReason := fmt.Sprintf("Failed to parse: %v", parseErr)
			failedFilesMap[relPathSlash] = failReason
			interpreter.Logger().Debug("%s [ERROR] %s: %s", logPrefix, failReason, relPathSlash)
			if firstProcessingError == nil {
				firstProcessingError = fmt.Errorf("[%s] %w", relPathSlash, parseErr)
			}
			continue // Process next file
		}
		// Analyze imports and symbols (using helper)
		oldAlias, needsMod, requiredNewImports, analysisErr := analyzeImportsAndSymbols(astFile, fileFset, refactoredPkgPath, symbolMap) // Call helper
		if analysisErr != nil {
			failReason := fmt.Sprintf("Analysis failed: %v", analysisErr)
			failedFilesMap[relPathSlash] = failReason
			interpreter.Logger().Debug("%s [ERROR] Failed analysis '%s': %s", logPrefix, relPathSlash, failReason)
			if firstProcessingError == nil {
				firstProcessingError = fmt.Errorf("[%s] analysis failed: %w", relPathSlash, analysisErr)
			}
			continue // Process next file
		}

		if needsMod {
			// STEP 1: Modify Imports (using helper)
			modifyErr := applyAstImportChanges(fileFset, astFile, refactoredPkgPath, requiredNewImports, interpreter) // Call helper
			if modifyErr != nil {                                                                                     // Should currently be nil, but handle defensively
				failReason := fmt.Sprintf("Import modification failed: %v", modifyErr)
				failedFilesMap[relPathSlash] = failReason
				interpreter.Logger().Debug("%s [ERROR] Failed import mod '%s': %s", logPrefix, relPathSlash, failReason)
				if firstProcessingError == nil {
					firstProcessingError = fmt.Errorf("[%s] import mod failed: %w", relPathSlash, modifyErr)
				}
				continue // Process next file
			}
			interpreter.Logger().Debug("%s Applied import changes for %s", logPrefix, relPathSlash)

			// STEP 2: Rewrite Qualifiers (Two-Pass)
			rewriteOccurred := false
			replacements := map[ast.Node]ast.Node{}
			importAliases := make(map[string]string)
			// Build import alias map
			for _, imp := range astFile.Imports {
				if imp.Path == nil {
					continue
				}
				impPath := strings.Trim(imp.Path.Value, `"`)
				name := ""
				if imp.Name != nil {
					name = imp.Name.Name
				} else {
					parts := strings.Split(impPath, "/")
					if len(parts) > 0 {
						name = parts[len(parts)-1]
						name = strings.ReplaceAll(name, "-", "_")
						name = strings.ReplaceAll(name, ".", "_")
					}
				}
				if name != "" {
					importAliases[impPath] = name
				}
			}
			// Pass 1: Collect replacements
			ast.Inspect(astFile, func(node ast.Node) bool {
				selExpr, ok := node.(*ast.SelectorExpr)
				if !ok {
					return true
				}
				ident, okX := selExpr.X.(*ast.Ident)
				if !okX || ident.Name != oldAlias {
					return true
				}
				symbolName := selExpr.Sel.Name
				newPath, exists := symbolMap[symbolName]
				if !exists {
					return false
				} // Symbol not in map, leave it
				newPkgName, aliasFound := importAliases[newPath]
				if !aliasFound { // Fallback inference
					parts := strings.Split(newPath, "/")
					if len(parts) > 0 {
						newPkgName = parts[len(parts)-1]
						newPkgName = strings.ReplaceAll(newPkgName, "-", "_")
						newPkgName = strings.ReplaceAll(newPkgName, ".", "_")
						interpreter.Logger().Debug("%s [WARN] Inferred alias '%s' for '%s.%s'", logPrefix, newPkgName, newPath, symbolName)
					} else {
						newPkgName = ""
					}
				}
				if newPkgName == "" {
					interpreter.Logger().Debug("%s [WARN] Cannot find pkg name for '%s'. Skipping '%s.%s'", logPrefix, newPath, oldAlias, symbolName)
					return false
				}
				if newPkgName != oldAlias { // If different, plan replacement
					interpreter.Logger().Debug("   Planning rewrite: '%s.%s' -> '%s.%s'", oldAlias, symbolName, newPkgName, symbolName)
					newIdent := ast.NewIdent(newPkgName)
					newSelExpr := &ast.SelectorExpr{X: newIdent, Sel: selExpr.Sel}
					replacements[selExpr] = newSelExpr
					rewriteOccurred = true
				} else {
					interpreter.Logger().Debug("%s    No rewrite needed for '%s.%s'", logPrefix, oldAlias, symbolName)
				} // Same name, no rewrite needed
				return false // Stop descent into this specific selector's children
			})
			// Pass 2: Apply replacements
			if rewriteOccurred {
				appliedNode := astutil.Apply(astFile, func(cursor *astutil.Cursor) bool {
					if newNode, ok := replacements[cursor.Node()]; ok {
						cursor.Replace(newNode)
						return false
					}
					return true
				}, nil)
				if newAstFile, ok := appliedNode.(*ast.File); ok && newAstFile != astFile {
					astFile = newAstFile
				} // Handle rare case where root is replaced
				interpreter.Logger().Debug("%s Applied %d qualifier rewrites in %s", logPrefix, len(replacements), relPathSlash)
			} else {
				interpreter.Logger().Debug("%s [INFO] No qualifiers needed rewriting in %s.", logPrefix, relPathSlash)
			}

			// STEP 3: Format and Write Back
			var buf bytes.Buffer
			formatErr := format.Node(&buf, fileFset, astFile) // Use the file-specific FileSet
			if formatErr != nil {
				failReason := fmt.Sprintf("Failed to format: %v", formatErr)
				failedFilesMap[relPathSlash] = failReason
				interpreter.Logger().Debug("%s [ERROR] Failed formatting '%s': %s", logPrefix, relPathSlash, failReason)
				if firstProcessingError == nil {
					firstProcessingError = fmt.Errorf("[%s] formatting failed: %w", relPathSlash, formatErr)
				}
				continue // Process next file
			}
			info, statErr := os.Stat(filePath)
			perm := os.FileMode(0644) // Default perm
			if statErr == nil {
				perm = info.Mode().Perm()
			} else {
				interpreter.Logger().Debug("%s [WARN] Cannot stat '%s': %v. Using default perm.", logPrefix, filePath, statErr)
			}
			writeErr := os.WriteFile(filePath, []byte(buf.String()), perm)
			if writeErr != nil {
				failReason := fmt.Sprintf("Failed to write: %v", writeErr)
				failedFilesMap[relPathSlash] = failReason
				interpreter.Logger().Debug("%s [ERROR] Failed writing '%s': %s", logPrefix, relPathSlash, failReason)
				if firstProcessingError == nil {
					firstProcessingError = fmt.Errorf("[%s] writing failed: %w", relPathSlash, writeErr)
				}
				// Continue processing other files even if write fails
			} else {
				modifiedFilesList = append(modifiedFilesList, relPathSlash)
				interpreter.Logger().Debug("%s Modified and wrote file '%s'", logPrefix, relPathSlash)
			}
		} else { // needsMod == false
			skipReason := "Original package not imported"
			if oldAlias != "" {
				skipReason = "Original package imported, but no relevant symbols found/used"
			}
			skippedFilesMap[relPathSlash] = skipReason
			interpreter.Logger().Debug("%s Skipped '%s': %s", logPrefix, relPathSlash, skipReason)
		}
	} // End file processing loop

	// --- Final Results Construction ---
	if firstProcessingError != nil {
		interpreter.Logger().Debug("%s EXIT] Results: modified=%d, skipped=%d, failed=%d", logPrefix, len(modifiedFilesList), len(skippedFilesMap), len(failedFilesMap))
		interpreter.Logger().Debug("%s EXIT] Returning First Processingcore.Error: %v", logPrefix, firstProcessingError)
		return nil, firstProcessingError // Return NIL map and the error
	}
	// --- Success Case ---
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
	} // Include files that failed parsing/writing
	result := map[string]interface{}{"modified_files": finalModifiedFiles, "skipped_files": finalSkippedFiles, "failed_files": finalFailedFiles, "error": nil}
	interpreter.Logger().Debug("%s EXIT] Results: modified=%d, skipped=%d, failed=%d", logPrefix, len(modifiedFilesList), len(skippedFilesMap), len(failedFilesMap))
	interpreter.Logger().Debug("%s EXIT] Success.", logPrefix)
	return result, nil // Return populated map and nil error
}

// --- Registration ---
// func registerGoAstPackageTools(registry *core.ToolRegistry) error {
// 	err := registry.RegisterTool(core.ToolcoreImplementation{
// 		Spec: core.ToolSpec{
// 			Name:        "GoUpdateImportsForMovedPackage",
// 			Description: fmt.Sprintf("Version: %s. Analyzes Go files, updating imports/qualifiers for symbols moved into sub-packages. Uses AST, two-pass rewrite, returns nil map on error. Refactored helpers.", packageToolDebugVersion),
// 			Args: []core.ArgSpec{
// 				{Name: "refactored_package_path", Type: core.ArgTypeString, Required: true, Description: "Original import path now containing sub-packages."},
// 				{Name: "scan_scope", Type: core.ArgTypeString, Required: true, Description: "Directory path (relative to sandbox root) to scan."},
// 			},
// 			ReturnType: core.ArgTypeAny, // map[string]interface{} or nil on error
// 		}, Func: toolGoUpdateImportsForMovedPackage,
// 	})
// 	if err != nil {
// 		return fmt.Errorf("failed to register tool GoUpdateImportsForMovedPackage: %w", err)
// 	}
// 	return nil
// }
