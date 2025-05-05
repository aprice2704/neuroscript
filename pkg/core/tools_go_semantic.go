// NeuroScript Version: 0.3.0_patch1
// Last Modified: 2025-05-04 00:45:00 PDT // Simplify Inspect check, keep AST dump ON
// filename: pkg/core/tools_go_semantic.go

package core

import (
	"fmt"
	"go/token"

	// Needed for MaxInt32
	// Needed for ast.Print

	// "golang.org/x/tools/go/ast/astutil" // No longer needed
	"golang.org/x/tools/go/packages"
)

// SemanticIndex struct definition (unchanged)
type SemanticIndex struct {
	Fset     *token.FileSet
	Packages []*packages.Package
	LoadDir  string
	LoadErrs []error
}

// toolGoIndexCode implementation (unchanged)
func toolGoIndexCode(interpreter *Interpreter, args []interface{}) (interface{}, error) {
	// ... (implementation unchanged) ...
	logger := interpreter.Logger()
	targetDirRel := "."
	if len(args) > 0 && args[0] != nil {
		dirStr, ok := args[0].(string)
		if !ok {
			errMsg := fmt.Sprintf("GoIndexCode: directory argument type mismatch, expected string, got %T", args[0])
			logger.Error("[TOOL-GOINDEX] %s", errMsg)
			return "", fmt.Errorf("%w: %s", ErrInvalidArgument, errMsg)
		}
		targetDirRel = dirStr
	}
	sandboxRoot := interpreter.SandboxDir()
	absValidatedDir, pathErr := ResolveAndSecurePath(targetDirRel, sandboxRoot)
	if pathErr != nil {
		errMsg := fmt.Sprintf("GoIndexCode path validation failed for directory %q (relative to sandbox %q): %v", targetDirRel, sandboxRoot, pathErr)
		logger.Error("[TOOL-GOINDEX] %s", errMsg)
		return "", fmt.Errorf("%w: %s", ErrInvalidPath, errMsg)
	}
	logger.Info("[TOOL-GOINDEX] Starting indexing", "relative_dir", targetDirRel, "absolute_dir", absValidatedDir)
	fset := token.NewFileSet()
	cfg := &packages.Config{Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedTypesSizes | packages.NeedImports | packages.NeedDeps, Dir: absValidatedDir, Fset: fset, Tests: true}
	pkgs, loadErr := packages.Load(cfg, "./...")
	loadErrs := []error{}
	if loadErr != nil {
		logger.Warn("[TOOL-GOINDEX] Error during packages.Load execution", "error", loadErr)
		loadErrs = append(loadErrs, fmt.Errorf("packages.Load execution error: %w", loadErr))
	}
	packages.Visit(pkgs, nil, func(pkg *packages.Package) {
		for _, err := range pkg.Errors {
			logger.Warn("[TOOL-GOINDEX] Error loading package", "package", pkg.ID, "error", err)
			loadErrs = append(loadErrs, fmt.Errorf("package %s: %w", pkg.ID, err))
		}
	})
	if len(pkgs) == 0 && len(loadErrs) > 0 {
		logger.Error("[TOOL-GOINDEX] Failed to load any packages.", "first_error", loadErrs[0])
		return "", fmt.Errorf("failed to load any packages: %w", loadErrs[0])
	}
	logger.Info("[TOOL-GOINDEX] Package loading complete", "loaded_packages", len(pkgs), "errors_encountered", len(loadErrs))
	index := &SemanticIndex{Fset: fset, Packages: pkgs, LoadDir: absValidatedDir, LoadErrs: loadErrs}
	handle, err := interpreter.RegisterHandle(index, "semantic_index")
	if err != nil {
		logger.Error("[TOOL-GOINDEX] Failed to register handle", "error", err)
		return "", fmt.Errorf("%w: failed to register index handle: %w", ErrInternal, err)
	}
	logger.Info("[TOOL-GOINDEX] Semantic index created successfully", "handle", handle)
	return handle, nil
}
