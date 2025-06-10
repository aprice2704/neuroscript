// filename: pkg/core/tools_go_ast_path_helper.go
package goast

import (
	"errors"
	"fmt"
	"path"          // Use standard 'path' for joining/cleaning import paths
	"path/filepath" // Use filepath for OS-specific operations
	"strings"

	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

// debugCalculateCanonicalPath tries various methods to calculate the Go import path
// for a directory and logs the results for debugging.
// It returns the result of the refined conditional logic (Take 7 - Final).
func debugCalculateCanonicalPath(modulePath, moduleRootDir, dirPath string, logger interfaces.Logger) (string, error) {
	logPrefix := "[DEBUG PATH CALC]"
	if logger == nil {
		panic("Path calc needs a valid logger")
	}
	logger.Debug("%s Inputs: modulePath=%q, moduleRootDir=%q, dirPath=%q", logPrefix, modulePath, moduleRootDir, dirPath)

	// --- Input Validation ---
	if moduleRootDir == "" {
		err := errors.New("moduleRootDir cannot be empty")
		logger.Debug("%s FAIL: %v", logPrefix, err)
		return "", err
	}
	if modulePath == "" {
		err := errors.New("modulePath cannot be empty")
		logger.Debug("%s FAIL: %v", logPrefix, err)
		return "", err
	}
	if dirPath == "" {
		err := errors.New("dirPath cannot be empty")
		logger.Debug("%s FAIL: %v", logPrefix, err)
		return "", err
	}
	// --- End Input Validation ---

	// Calculate relative path first
	relFromModuleRoot, relErr := filepath.Rel(moduleRootDir, dirPath)
	if relErr != nil {
		logger.Debug("%s Base FAIL: filepath.Rel(%q, %q) error: %v", logPrefix, moduleRootDir, dirPath, relErr)
		return "", fmt.Errorf("failed to determine relative path from module root: %w", relErr)
	}
	relFromModuleRootSlash := filepath.ToSlash(relFromModuleRoot)

	// --- Log Results of Different Methods for Debugging ---
	// Method 1 Log: Just Rel from Root, then ToSlash -> path.Clean
	m1Path := relFromModuleRootSlash
	if m1Path == "." {
		m1Path = modulePath
	}
	m1Path = path.Clean(m1Path)
	logger.Debug("%s Method 1 Result (RelFromRoot -> ToSlash -> path.Clean): %q", logPrefix, m1Path)

	// Method 2 Log: path.Join(module, RelFromRootSlash)
	m2Path := path.Join(modulePath, relFromModuleRootSlash) // Use path.Join
	m2Path = path.Clean(m2Path)
	logger.Debug("%s Method 2 Result (path.Join(module, RelFromRootSlash)): %q", logPrefix, m2Path)

	// Method 3 Log: String concat modulePath + "/" + RelFromRootSlash
	m3Path := ""
	if relFromModuleRootSlash == "." {
		m3Path = modulePath
	} else {
		m3Path = strings.TrimSuffix(modulePath, "/") + "/" + strings.TrimPrefix(relFromModuleRootSlash, "/")
		m3Path = strings.ReplaceAll(m3Path, "//", "/") // Basic cleaning
	}
	logger.Debug("%s Method 3 Result (String Concat): %q", logPrefix, m3Path)

	// --- Final Conditional Logic (Derived from Debug Output) ---
	var canonicalPath string
	if relFromModuleRootSlash == "." {
		// Case 1: The directory *is* the module root.
		canonicalPath = modulePath
		logger.Debug("%s Final Logic: Path is module root, using modulePath: %q", logPrefix, canonicalPath)
	} else {
		// Case 2: Subdirectory. Check if joining is needed or if Rel path is sufficient.
		// If Method 1's result is the expected one (like "testtool/refactored/sub1"), use it.
		// Otherwise, Method 2's result (like "example.com/mymodule/pkg/subpkg") is correct.
		// We check if Method 2's result contains the doubled module path component.
		doublePrefix := modulePath + "/" + modulePath + "/"
		if strings.HasPrefix(m2Path, doublePrefix) {
			// If Method 2 doubled the path, Method 1 was likely correct.
			canonicalPath = m1Path
			logger.Debug("%s Final Logic: Method 2 result %q detected doubling, using Method 1 result: %q", logPrefix, m2Path, canonicalPath)
		} else {
			// Otherwise, Method 2 (path.Join) is the standard, correct way.
			canonicalPath = m2Path
			logger.Debug("%s Final Logic: Method 2 result %q seems correct, using it.", logPrefix, canonicalPath)
		}
	}

	// Final clean just in case (should be redundant if m1Path/m2Path are clean)
	canonicalPath = path.Clean(canonicalPath)
	logger.Debug("%s Final Logic Returning Path: %q", logPrefix, canonicalPath)

	// Return the result from the chosen logic
	return canonicalPath, nil // No error if Rel succeeded
}
