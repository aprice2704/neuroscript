// filename: pkg/core/tools_go_ast_package.go
package core

import (
	"fmt"
	"go/ast"
	"go/format"
	"go/token" // Need types
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

// toolGoUpdateImportsForMovedPackage Tool - Uses type checking via go/packages.
func toolGoUpdateImportsForMovedPackage(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage ENTRY] Received args: %v", args)

	// --- Argument Validation ---
	if len(args) != 2 {
		return nil, fmt.Errorf("GoUpdateImportsForMovedPackage: %w", ErrValidationArgCount)
	}
	refactoredPkgPath, ok := args[0].(string)
	if !ok {
		return nil, fmt.Errorf("GoUpdateImportsForMovedPackage: refactored_package_path %w", ErrValidationTypeMismatch)
	}
	if refactoredPkgPath == "" {
		return nil, fmt.Errorf("GoUpdateImportsForMovedPackage: refactored_package_path %w", ErrValidationRequiredArgNil)
	}
	scanScope, ok := args[1].(string)
	if !ok {
		return nil, fmt.Errorf("GoUpdateImportsForMovedPackage: scan_scope %w", ErrValidationTypeMismatch)
	}
	if scanScope == "" {
		return nil, fmt.Errorf("GoUpdateImportsForMovedPackage: scan_scope %w", ErrValidationRequiredArgNil)
	}
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Validated args: refactored_package_path='%s', scan_scope='%s'", refactoredPkgPath, scanScope)

	// --- Security Validation ---
	sandboxRoot := interpreter.sandboxDir
	if sandboxRoot == "" {
		sandboxRoot = "."
	}
	validatedScanScope, scopeErr := SecureFilePath(scanScope, sandboxRoot)
	if scopeErr != nil {
		errMsg := fmt.Sprintf("scan_scope validation failed: %s", scopeErr.Error())
		interpreter.logger.Printf("[ERROR GoUpdateImportsForMovedPackage] %s", errMsg)
		return map[string]interface{}{"error": errMsg}, nil
	}
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Validated scan scope: '%s'", validatedScanScope)

	// --- Initialize Results & Shared State ---
	modifiedFiles := []string{}
	skippedFiles := make(map[string]string)
	failedFiles := make(map[string]string)
	var topLevelError string
	fset := token.NewFileSet()

	// --- Build Symbol Map ---
	// Pass interpreter which now contains sandboxDir needed by buildSymbolMap
	symbolMap, err := buildSymbolMap(refactoredPkgPath, interpreter)
	if err != nil {
		topLevelError = fmt.Sprintf("Failed to build symbol map: %v", err)
		interpreter.logger.Printf("[ERROR GoUpdateImportsForMovedPackage] %s", topLevelError)
		return map[string]interface{}{"error": topLevelError}, nil
	}
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Symbol map built. Size: %d", len(symbolMap))
	if len(symbolMap) == 0 {
		interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] Symbol map is empty. No symbols found in refactored sub-packages.")
	}

	// --- FIX: Determine Refactored Package Dir for Exclusion (No helper needed) ---
	// We need the relative path part of the import path to join with the scan scope.
	// Assumes refactoredPkgPath is like "module/path/to/package"
	// TODO: Robustly get module path instead of assuming structure/hardcoding.
	// For test fixtures (module testtool), we strip "testtool/".
	// A better way might be needed for real projects.
	modulePrefix := ""
	// Attempt to get module path - could run 'go list -m'
	// For now, assume standard structure or known prefix based on context
	parts := strings.SplitN(refactoredPkgPath, "/", 2)
	if len(parts) == 2 { // Assuming first part is module path
		modulePrefix = parts[0] + "/"
	} else {
		interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] Could not determine module prefix from '%s'. Exclusion logic might be inaccurate.", refactoredPkgPath)
	}
	relativePkgDir := strings.TrimPrefix(refactoredPkgPath, modulePrefix)
	if relativePkgDir == "" && modulePrefix != "" {
		interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] Relative package directory is empty after stripping module prefix from '%s'. Exclusion logic might be inaccurate.", refactoredPkgPath)
		// Use the original path fragment if stripping failed unexpectedly
		relativePkgDir = filepath.Base(refactoredPkgPath)
	}

	refactoredDirAbs := filepath.Join(validatedScanScope, filepath.FromSlash(relativePkgDir))
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Calculated refactored dir path for exclusion: '%s'", refactoredDirAbs)
	// --- END FIX ---

	// --- Step 1: Collect target Go file paths ---
	goFilePaths := []string{}
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Starting file walk to collect .go files in '%s'", validatedScanScope)
	walkErr := filepath.WalkDir(validatedScanScope, func(path string, d os.DirEntry, walkErrInCb error) error {
		// (Walk logic remains the same as before)
		if walkErrInCb != nil {
			interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] Error accessing path %q during collection: %v", path, walkErrInCb)
			if d != nil && d.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if name == "vendor" || name == ".git" || name == "testdata" {
				return filepath.SkipDir
			}
			cleanedPath := filepath.Clean(path)
			cleanedRefactoredDir := filepath.Clean(refactoredDirAbs)
			// Ensure we compare absolute paths if possible, or consistently relative
			if strings.HasPrefix(cleanedPath, cleanedRefactoredDir) {
				interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Skipping directory within refactored scope: %s", path)
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".go") || strings.HasSuffix(d.Name(), "_test.go") {
			return nil
		}
		absPath, pathErr := filepath.Abs(path)
		if pathErr != nil {
			interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] Failed to get absolute path for %s: %v. Skipping.", path, pathErr)
			return nil
		}
		goFilePaths = append(goFilePaths, absPath)
		return nil
	})
	if walkErr != nil {
		topLevelError = fmt.Sprintf("File collection walk failed: %v", walkErr)
		interpreter.logger.Printf("[ERROR GoUpdateImportsForMovedPackage] %s", topLevelError)
		return map[string]interface{}{"error": topLevelError}, nil
	}
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Collected %d potentially relevant Go files.", len(goFilePaths))
	if len(goFilePaths) == 0 {
		interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] No .go files found in scan scope. Exiting.")
		return map[string]interface{}{}, nil
	}

	// --- Step 2: Load packages with type information ---
	// (Loading logic remains the same as before)
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Loading packages for analysis...")
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo, Fset: fset, Dir: validatedScanScope}
	loadPatterns := make([]string, 0, len(goFilePaths))
	for _, p := range goFilePaths {
		loadPatterns = append(loadPatterns, "file="+p)
	}
	pkgs, loadErr := packages.Load(cfg, loadPatterns...)
	if loadErr != nil {
		topLevelError = fmt.Sprintf("Failed to load packages: %v", loadErr)
		interpreter.logger.Printf("[ERROR GoUpdateImportsForMovedPackage] %s", topLevelError)
		return map[string]interface{}{"error": topLevelError}, nil
	}
	loadErrors := []string{}
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			loadErrors = append(loadErrors, fmt.Sprintf("file=%s: %v", pkg.ID, err))
		}
	})
	if len(loadErrors) > 0 {
		topLevelError = fmt.Sprintf("errors encountered loading packages: %s", strings.Join(loadErrors, "; "))
		interpreter.logger.Printf("[ERROR GoUpdateImportsForMovedPackage] %s", topLevelError)
		return map[string]interface{}{"error": topLevelError}, nil
	}
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Successfully loaded %d package(s) for analysis.", len(pkgs))

	// --- Step 3: Process loaded packages/files ---
	// (Processing logic remains the same as before)
	processedFiles := make(map[string]bool)
	for _, pkg := range pkgs {
		interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Processing package: %s", pkg.ID) // Use pkg.ID which might be file path
		typesInfo := pkg.TypesInfo
		if typesInfo == nil {
			interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] TypesInfo is nil for package %s. Skipping analysis for files in this package.", pkg.ID)
			for _, astFile := range pkg.Syntax {
				filePath := fset.Position(astFile.Pos()).Filename
				if !processedFiles[filePath] {
					failedFiles[filePath] = "Failed to get type information for analysis"
					processedFiles[filePath] = true
				}
			}
			continue
		}
		for _, astFile := range pkg.Syntax {
			tokenFile := fset.File(astFile.Pos())
			if tokenFile == nil {
				interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] Could not get file info from token.Pos for AST node in package %s. Skipping.", pkg.ID)
				continue
			}
			filePath := tokenFile.Name()
			if processedFiles[filePath] {
				continue
			}
			processedFiles[filePath] = true
			interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Analyzing file: %s", filePath)
			needsMod, requiredNewImports, analysisErr := analyzeImportsAndSymbols(astFile, fset, refactoredPkgPath, symbolMap, typesInfo)
			if analysisErr != nil {
				failReason := fmt.Sprintf("Analysis failed: %v", analysisErr)
				failedFiles[filePath] = failReason
				interpreter.logger.Printf("[ERROR GoUpdateImportsForMovedPackage] Failed analysis for '%s': %s", filePath, failReason)
				continue
			}
			if !needsMod {
				importExists := false
				astutil.Apply(astFile, func(c *astutil.Cursor) bool {
					if impSpec, ok := c.Node().(*ast.ImportSpec); ok {
						if strings.Trim(impSpec.Path.Value, `"`) == refactoredPkgPath {
							importExists = true
							return false
						}
					}
					return true
				}, nil)
				if !importExists {
					skippedFiles[filePath] = "Original package not imported"
				} else {
					skippedFiles[filePath] = "Original package imported, but no relevant symbols found/mapped"
				}
				interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Skipped '%s': %s", filePath, skippedFiles[filePath])
				continue
			}
			modifyErr := applyAstImportChanges(fset, astFile, refactoredPkgPath, requiredNewImports)
			if modifyErr != nil {
				failReason := fmt.Sprintf("Failed to apply AST import changes: %v", modifyErr)
				failedFiles[filePath] = failReason
				interpreter.logger.Printf("[ERROR GoUpdateImportsForMovedPackage] Failed AST modification for '%s': %s", filePath, failReason)
				continue
			}
			var buf strings.Builder
			formatErr := format.Node(&buf, fset, astFile)
			if formatErr != nil {
				failReason := fmt.Sprintf("Failed to format modified AST: %v", formatErr)
				failedFiles[filePath] = failReason
				interpreter.logger.Printf("[ERROR GoUpdateImportsForMovedPackage] Failed formatting for '%s': %s", filePath, failReason)
				continue
			}
			formattedCode := buf.String()
			info, statErr := os.Stat(filePath)
			perm := os.FileMode(0644)
			if statErr == nil {
				perm = info.Mode().Perm()
			} else {
				interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] Could not stat original file '%s': %v. Using default 0644.", filePath, statErr)
			}
			// Convert absolute path back to relative for reporting
			relPath, relErr := filepath.Rel(validatedScanScope, filePath)
			if relErr != nil {
				interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] Could not get relative path for reporting modified file %s: %v", filePath, relErr)
				relPath = filePath // Use absolute path if relativization fails
			}
			writeErr := os.WriteFile(filePath, []byte(formattedCode), perm)
			if writeErr != nil {
				failReason := fmt.Sprintf("Failed to write modified file: %v", writeErr)
				failedFiles[relPath] = failReason
				interpreter.logger.Printf("[ERROR GoUpdateImportsForMovedPackage] Failed writing '%s': %s", relPath, failReason)
			} else {
				modifiedFiles = append(modifiedFiles, relPath)
				interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage] Successfully modified and wrote file '%s'", relPath)
			}
		}
	}
	// Final check for files not processed
	for _, fp := range goFilePaths {
		if !processedFiles[fp] {
			relPath, _ := filepath.Rel(validatedScanScope, fp)
			if _, exists := failedFiles[relPath]; !exists {
				failedFiles[relPath] = "File was not processed by packages.Load (check build constraints or parse errors)"
				interpreter.logger.Printf("[WARN GoUpdateImportsForMovedPackage] File collected but not processed: %s", relPath)
			}
		}
	}

	// --- Return Results ---
	result := map[string]interface{}{"modified_files": interface{}(modifiedFiles), "skipped_files": interface{}(skippedFiles), "failed_files": interface{}(failedFiles), "error": nil}
	if topLevelError != "" {
		result["error"] = topLevelError
	}
	interpreter.logger.Printf("[TOOL GoUpdateImportsForMovedPackage EXIT] Results: modified=%d, skipped=%d, failed=%d, error='%v'", len(modifiedFiles), len(skippedFiles), len(failedFiles), topLevelError)
	return result, nil
}

// --- Registration --- (Remains the same)
func registerGoAstPackageTools(registry *ToolRegistry) error { /* ... */ return nil } // Elided for brevity
