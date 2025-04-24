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

	// "os/exec" // Keep commented unless needed
	"path/filepath" // Still needed by other parts of the file
	"strings"
	// Removed astutil import if only used by the moved function
	// "golang.org/x/tools/go/ast/astutil"
)

const packageToolDebugVersion = "v6d_AST_IMPORTS_HELPER_MOVED" // Incremented version

// analyzeImportsAndSymbols identifies if the old import exists and which new imports are needed
// based on symbol usage.
func analyzeImportsAndSymbols(astFile *ast.File, fset *token.FileSet, oldPath string, symbolMap map[string]string) (bool, map[string]string, error) {
	// (Content of function remains the same)
	needsMod := false
	oldImportAliasOrName := "" // Store the alias or package name used for the old import
	oldImportFound := false
	requiredNewImports := make(map[string]string) // New path -> ""

	// First pass: Check if the old import exists and find its alias/name
	for _, impSpec := range astFile.Imports {
		if impSpec.Path != nil && strings.Trim(impSpec.Path.Value, `"`) == oldPath {
			oldImportFound = true
			if impSpec.Name != nil {
				oldImportAliasOrName = impSpec.Name.Name
			} else {
				// If no alias, infer package name from path (simplified)
				parts := strings.Split(oldPath, "/")
				if len(parts) > 0 {
					oldImportAliasOrName = parts[len(parts)-1]
				} else {
					oldImportAliasOrName = oldPath // Fallback
				}

			}
			break // Found the import
		}
	}

	if !oldImportFound {
		return false, nil, nil // No modification needed if old import isn't present
	}

	// Second pass: Find usages of symbols potentially from the old package
	ast.Inspect(astFile, func(node ast.Node) bool {
		selExpr, ok := node.(*ast.SelectorExpr)
		if !ok {
			return true // Continue inspecting other nodes
		}

		// Check if the expression uses the alias/name of the old package
		ident, ok := selExpr.X.(*ast.Ident)
		if !ok || ident.Name != oldImportAliasOrName {
			return true // Not accessing via the old package import, continue
		}

		// Check if the selected symbol is one that was moved
		symbolName := selExpr.Sel.Name
		if newPath, exists := symbolMap[symbolName]; exists {
			// Found a symbol from the old package that has moved!
			needsMod = true                  // Mark file for modification
			requiredNewImports[newPath] = "" // Add the new path requirement
		}
		// Don't traverse deeper into SelectorExpr
		return false // Stop descent here
	})

	if !needsMod {
		// Old import exists, but no relevant symbols found/used
		return false, nil, nil
	}

	return needsMod, requiredNewImports, nil
}

// --- applyAstImportChanges REMOVED from this file ---

// toolGoUpdateImportsForMovedPackage Tool - Uses AST analysis only.
func toolGoUpdateImportsForMovedPackage(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// (Function body remains the same as previous version, but ensure the call
	// to applyAstImportChanges is correct - it should not have changed)
	logPrefix := fmt.Sprintf("[TOOL GoUpdateImports %s]", packageToolDebugVersion)
	interpreter.logger.Printf("%s ENTRY] Received args: %v", logPrefix, args)

	// --- Argument Validation ---
	if len(args) != 2 {
		return nil, fmt.Errorf("%w: GoUpdateImportsForMovedPackage requires exactly 2 arguments (refactored_package_path, scan_scope)", ErrValidationArgCount)
	}
	refactoredPkgPath, ok := args[0].(string)
	if !ok || refactoredPkgPath == "" {
		return nil, fmt.Errorf("%w: GoUpdateImportsForMovedPackage: refactored_package_path must be a non-empty string", ErrValidationArgValue)
	}
	scanScope, ok := args[1].(string)
	if !ok || scanScope == "" {
		return nil, fmt.Errorf("%w: GoUpdateImportsForMovedPackage: scan_scope must be a non-empty string", ErrValidationArgValue)
	}
	interpreter.logger.Printf("%s Validated args: refactored_package_path='%s', scan_scope='%s'", logPrefix, refactoredPkgPath, scanScope)

	// --- Security Validation ---
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

	// --- Initialize Results & Shared State ---
	modifiedFilesList := []string{}
	skippedFilesMap := make(map[string]string)
	failedFilesMap := make(map[string]string)
	var topLevelError error
	fset := token.NewFileSet()

	// --- Build Symbol Map ---
	interpreter.logger.Printf("%s === Calling buildSymbolMap (Manual) ===", logPrefix)
	symbolMap, err := buildSymbolMap(refactoredPkgPath, interpreter)
	interpreter.logger.Printf("%s === buildSymbolMap returned ===", logPrefix)
	if err != nil {
		errMsg := ""
		if errors.Is(err, ErrRefactoredPathNotFound) {
			errMsg = fmt.Sprintf("cannot update imports: the refactored package path '%s' does not exist or is not accessible within the sandbox", refactoredPkgPath)
		} else {
			errMsg = fmt.Sprintf("failed to build symbol map for '%s': %v", refactoredPkgPath, err)
		}
		topLevelError = errors.New(errMsg)
		interpreter.logger.Printf("%s [ERROR] %s", logPrefix, topLevelError.Error())
		return map[string]interface{}{
			"modified_files": []interface{}{},
			"skipped_files":  map[string]interface{}{},
			"failed_files":   map[string]interface{}{},
			"error":          topLevelError.Error(),
		}, nil
	}
	interpreter.logger.Printf("%s Symbol map built successfully. Size: %d", logPrefix, len(symbolMap))
	if len(symbolMap) == 0 {
		interpreter.logger.Printf("%s [INFO] Symbol map is empty for '%s'. No files needed modification.", logPrefix, refactoredPkgPath)
		return map[string]interface{}{
			"modified_files": []interface{}{},
			"skipped_files":  map[string]interface{}{},
			"failed_files":   map[string]interface{}{},
			"error":          nil,
			"message":        fmt.Sprintf("No exported symbols found in sub-packages of '%s'. No files needed modification.", refactoredPkgPath),
		}, nil
	}

	// --- Determine Refactored Package Dir for Exclusion ---
	refactoredDirAbs := filepath.Join(sandboxRoot, filepath.FromSlash(refactoredPkgPath))
	interpreter.logger.Printf("%s Calculated refactored dir path (absolute): '%s'", logPrefix, refactoredDirAbs)

	// --- Step 1: Collect target Go file paths ---
	goFilePaths, walkErr := collectGoFiles(validatedScanScope, refactoredDirAbs, interpreter)
	if walkErr != nil {
		topLevelError = fmt.Errorf("file collection failed: %w", walkErr)
		interpreter.logger.Printf("%s [ERROR] %s", logPrefix, topLevelError.Error())
		return map[string]interface{}{
			"modified_files": []interface{}{},
			"skipped_files":  map[string]interface{}{},
			"failed_files":   map[string]interface{}{},
			"error":          topLevelError.Error(),
		}, nil
	}
	interpreter.logger.Printf("%s Collected %d potentially relevant Go files.", logPrefix, len(goFilePaths))
	if len(goFilePaths) == 0 {
		interpreter.logger.Printf("%s No .go files found in scan scope to analyze. Exiting.", logPrefix)
		return map[string]interface{}{
			"modified_files": []interface{}{},
			"skipped_files":  map[string]interface{}{},
			"failed_files":   map[string]interface{}{},
			"error":          nil,
			"message":        "No Go files found in the specified scan_scope (excluding the refactored package directory).",
		}, nil
	}

	// --- Step 2: Parse target files and analyze AST ---
	interpreter.logger.Printf("%s === Parsing and Analyzing Files (AST ONLY) ===", logPrefix)
	for _, filePath := range goFilePaths {
		relPath, relErr := filepath.Rel(sandboxRoot, filePath)
		if relErr != nil {
			interpreter.logger.Printf("%s [WARN] Could not make path relative '%s': %v. Using absolute path.", logPrefix, filePath, relErr)
			relPath = filePath
		}
		relPathSlash := filepath.ToSlash(relPath)
		interpreter.logger.Printf("%s Processing file: %s", logPrefix, relPathSlash)

		// Parse the file
		astFile, parseErr := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
		if parseErr != nil {
			failReason := fmt.Sprintf("Failed to parse file: %v", parseErr)
			failedFilesMap[relPathSlash] = failReason
			interpreter.logger.Printf("%s [ERROR] %s: %s", logPrefix, failReason, relPathSlash)
			continue
		}

		// Analyze imports and symbols
		interpreter.logger.Printf("%s Analyzing file: %s", logPrefix, relPathSlash)
		needsMod, requiredNewImports, analysisErr := analyzeImportsAndSymbols(astFile, fset, refactoredPkgPath, symbolMap)
		if analysisErr != nil {
			failReason := fmt.Sprintf("Analysis failed: %v", analysisErr)
			failedFilesMap[relPathSlash] = failReason
			interpreter.logger.Printf("%s [ERROR] Failed analysis for '%s': %s", logPrefix, relPathSlash, failReason)
			continue
		}

		if needsMod {
			// Apply changes using the function now in the helpers file
			modifyErr := applyAstImportChanges(fset, astFile, refactoredPkgPath, requiredNewImports) // This call assumes applyAstImportChanges is accessible
			if modifyErr != nil {
				failReason := fmt.Sprintf("Failed to apply AST import changes: %v", modifyErr)
				failedFilesMap[relPathSlash] = failReason
				interpreter.logger.Printf("%s [ERROR] Failed AST modification for '%s': %s", logPrefix, relPathSlash, failReason)
				continue
			}

			// Format and write back
			var buf strings.Builder
			formatErr := format.Node(&buf, fset, astFile)
			if formatErr != nil {
				failReason := fmt.Sprintf("Failed to format modified AST: %v", formatErr)
				failedFilesMap[relPathSlash] = failReason
				interpreter.logger.Printf("%s [ERROR] Failed formatting for '%s': %s", logPrefix, relPathSlash, failReason)
				continue
			}

			// Write file
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
			// Determine skip reason
			importExists := false
			for _, impSpec := range astFile.Imports {
				if impSpec.Path != nil {
					if strings.Trim(impSpec.Path.Value, `"`) == refactoredPkgPath {
						importExists = true
						break
					}
				}
			}
			if !importExists {
				skippedFilesMap[relPathSlash] = "Original package not imported"
			} else {
				skippedFilesMap[relPathSlash] = "Original package imported, but no relevant symbols found/used"
			}
			interpreter.logger.Printf("%s Skipped '%s': %s", logPrefix, relPathSlash, skippedFilesMap[relPathSlash])
		}
	} // End loop through files

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
	interpreter.logger.Printf("%s EXIT] Results: modified=%d, skipped=%d, failed=%d, error='%v'", logPrefix, len(modifiedFilesList), len(skippedFilesMap), len(failedFilesMap), finalErrorValue)
	return result, nil
}

// --- Helper: collectGoFiles ---
func collectGoFiles(scanScopeAbs, excludeDirAbs string, interpreter *Interpreter) ([]string, error) {
	// (Content of function remains the same as previous version)
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
				return nil // Continue walk, but files inside will be skipped
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

// --- Registration ---
func registerGoAstPackageTools(registry *ToolRegistry) error {
	// (Content of function remains the same)
	err := registry.RegisterTool(ToolImplementation{
		Spec: ToolSpec{
			Name:        "GoUpdateImportsForMovedPackage",
			Description: fmt.Sprintf("Version: %s. Analyzes Go files within a scope, updating import paths for symbols moved from an 'original' package path into its sub-packages. Uses AST analysis and a symbol map. Removes the old import and adds necessary new ones. Does NOT update code qualifiers.", packageToolDebugVersion),
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

// --- Variable checks ---
// REMOVED check for applyAstImportChanges
var _ func(astFile *ast.File, fset *token.FileSet, oldPath string, symbolMap map[string]string) (bool, map[string]string, error) = analyzeImportsAndSymbols
