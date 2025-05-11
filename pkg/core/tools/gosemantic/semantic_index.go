// NeuroScript Version: 0.3.1
// File version: 0.0.3 // Define PackageInfo locally, use it.
// Define SemanticIndex and GoIndexCode tool.
// filename: pkg/core/tools/gosemantic/semantic_index.go

// Package gosemantic provides tools for semantic analysis of Go code using go/packages.
package gosemantic

import (
	"fmt"
	"go/token"
	"log" // Keep log for temporary debugging from original file
	"os"  // Keep os for reading file content from original file
	"path/filepath"
	"strings" // Keep strings for line counting from original file

	"golang.org/x/tools/go/packages"

	"github.com/aprice2704/neuroscript/pkg/core" // Import core for Interpreter, errors etc.
)

// PackageInfo wraps a *packages.Package to potentially add more metadata later.
// Define this locally within gosemantic as it's tightly coupled with SemanticIndex.
type PackageInfo struct {
	*packages.Package // Embed the original package info
	// Add custom fields here if needed in the future
}

// SemanticIndex holds the results of a `go/packages` load operation.
type SemanticIndex struct {
	Fset     *token.FileSet
	Packages []*PackageInfo // Use local PackageInfo type
	LoadDir  string         // Absolute path to the directory that was loaded
	LoadErrs []error        // Any non-fatal errors encountered during loading
}

// Type tag for handles containing SemanticIndex
const semanticIndexTypeTag = "semantic_index"

// --- Tool: GoIndexCode ---

func toolGoIndexCode(interpreter *core.Interpreter, args []interface{}) (interface{}, error) {
	logger := interpreter.Logger()
	targetDirRel := "."
	if len(args) > 0 && args[0] != nil {
		dirStr, ok := args[0].(string)
		if !ok {
			errMsg := fmt.Sprintf("directory argument type mismatch, expected string, got %T", args[0])
			logger.Error("[TOOL-GOINDEX] " + errMsg)
			return "", fmt.Errorf("%w: %s", core.ErrInvalidArgument, errMsg)
		}
		targetDirRel = dirStr
	}

	sandboxRoot := interpreter.SandboxDir()
	absValidatedDir, pathErr := core.ResolveAndSecurePath(targetDirRel, sandboxRoot)
	if pathErr != nil {
		errMsg := fmt.Sprintf("path validation failed for directory %q (relative to sandbox %q): %v", targetDirRel, sandboxRoot, pathErr)
		logger.Error("[TOOL-GOINDEX] " + errMsg)
		return "", fmt.Errorf("%w: %s", core.ErrInvalidPath, errMsg)
	}

	logger.Debug("[TOOL-GOINDEX] Starting indexing", "relative_dir", targetDirRel, "absolute_dir", absValidatedDir)

	// +++ DEBUG code from original file (keep for now) +++
	debugFilePath := filepath.Join(absValidatedDir, "main.go") // Assuming the test file is main.go
	if _, statErr := os.Stat(debugFilePath); statErr == nil {
		contentBytes, readErr := os.ReadFile(debugFilePath)
		if readErr == nil {
			contentStr := string(contentBytes)
			actualLineCount := strings.Count(contentStr, "\n") + 1
			actualSize := len(contentBytes)
			log.Printf("[DEBUG GoIndexCode] Pre-Load Check for %s: Size=%d, ActualLines=%d", debugFilePath, actualSize, actualLineCount)
		} else {
			log.Printf("[DEBUG GoIndexCode] Pre-Load Check: Error reading %s: %v", debugFilePath, readErr)
		}
	} else {
		log.Printf("[DEBUG GoIndexCode] Pre-Load Check: Test file %s not found or error: %v", debugFilePath, statErr)
	}
	// +++ END DEBUG +++

	fset := token.NewFileSet()

	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax |
			packages.NeedTypes | packages.NeedTypesInfo | packages.NeedTypesSizes |
			packages.NeedImports | packages.NeedDeps,
		Dir:   absValidatedDir,
		Fset:  fset,
		Tests: true,
	}
	pkgs, loadErr := packages.Load(cfg, "./...")

	// Wrap loaded packages into local PackageInfo and collect errors
	packageInfos := make([]*PackageInfo, 0, len(pkgs)) // Use local PackageInfo
	loadErrs := []error{}
	if loadErr != nil {
		logger.Warn("[TOOL-GOINDEX] Error during packages.Load execution", "error", loadErr)
		loadErrs = append(loadErrs, fmt.Errorf("packages.Load execution error: %w", loadErr))
	}
	// Visit packages to collect detailed errors and build PackageInfo slice
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		// *** FIXED: Create and append local PackageInfo struct ***
		packageInfos = append(packageInfos, &PackageInfo{Package: pkg}) // Wrap *packages.Package

		for _, err := range pkg.Errors {
			logger.Warn("[TOOL-GOINDEX] Error loading package", "package", pkg.ID, "error", err)
			loadErrs = append(loadErrs, fmt.Errorf("package %s: %w", pkg.ID, err))
		}
	})

	// +++ DEBUG code from original file (keep for now) +++
	//	log.Printf("[DEBUG GoIndexCode] Post-Load Check: Iterating populated FileSet (fset base: %d)", fset.Base())
	fset.Iterate(func(f *token.File) bool {
		if f != nil {
			//		log.Printf("[DEBUG GoIndexCode]   - File: Name=%q, Size=%d, LineCount=%d, Base=%d", f.Name(), f.Size(), f.LineCount(), f.Base())
		} else {
			log.Printf("[DEBUG GoIndexCode]   - Nil file encountered in FileSet")
		}
		return true
	})
	// +++ END DEBUG +++

	// Check for fatal loading failure
	if len(pkgs) == 0 && len(loadErrs) > 0 {
		isNoGoFilesError := false
		for _, err := range loadErrs {
			if strings.Contains(err.Error(), "no Go files") {
				isNoGoFilesError = true
				break
			}
		}
		if isNoGoFilesError {
			logger.Warn("[TOOL-GOINDEX] packages.Load reported 'no Go files' - this might be expected.", "dir", absValidatedDir)
		} else {
			logger.Error("[TOOL-GOINDEX] Failed to load any packages.", "first_error", loadErrs[0])
			return "", fmt.Errorf("failed to load any packages: %w", loadErrs[0])
		}
	}

	logger.Debug("[TOOL-GOINDEX] Package loading complete", "loaded_packages", len(pkgs), "package_infos", len(packageInfos), "errors_encountered", len(loadErrs))

	// Create and register the index
	index := &SemanticIndex{
		Fset:     fset,
		Packages: packageInfos, // Store the slice of *PackageInfo
		LoadDir:  absValidatedDir,
		LoadErrs: loadErrs,
	}
	handle, err := interpreter.RegisterHandle(index, semanticIndexTypeTag)
	if err != nil {
		logger.Error("[TOOL-GOINDEX] Failed to register handle", "error", err)
		return "", fmt.Errorf("%w: failed to register index handle: %w", core.ErrInternal, err)
	}

	logger.Debug("[TOOL-GOINDEX] Semantic index created successfully", "handle", handle)
	return handle, nil
}
