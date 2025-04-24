// filename: pkg/core/tools_go_ast_package_helpers.go
package core

import (
	"errors" // <-- ADDED IMPORT (needed for errors.New)
	"fmt"
	"go/ast"
	"go/token"
	"sort" // <-- ADDED IMPORT (needed for sorting new paths)
	"strings"

	"golang.org/x/tools/go/ast/astutil"
)

// --- applyAstImportChanges REINSTATED in this file ---
// applyAstImportChanges modifies the import declarations in the AST.
func applyAstImportChanges(fset *token.FileSet, astFile *ast.File, oldPath string, newImports map[string]string) error {
	// Use the debug version from the main file for consistency, or define one here
	logPrefix := fmt.Sprintf("[DEBUG applyAstImportChanges %s]", packageToolDebugVersion) // Assumes packageToolDebugVersion is accessible or redefine
	fmt.Printf("%s Applying import changes: remove %s, add %d new paths\n", logPrefix, oldPath, len(newImports))

	// Delete the old import correctly
	deleted := false
	for _, impSpec := range astFile.Imports {
		if impSpec.Path != nil && strings.Trim(impSpec.Path.Value, `"`) == oldPath {
			if impSpec.Name != nil {
				// Has an alias (e.g., original "path")
				fmt.Printf("%s Deleting named import '%s' \"%s\".\n", logPrefix, impSpec.Name.Name, oldPath)
				if astutil.DeleteNamedImport(fset, astFile, impSpec.Name.Name, oldPath) {
					deleted = true
					fmt.Printf("%s Successfully deleted named import '%s' %s\n", logPrefix, impSpec.Name.Name, oldPath)
					break // Assume only one import spec matches
				} else {
					fmt.Printf("%s [WARN] astutil.DeleteNamedImport returned false for '%s' \"%s\".\n", logPrefix, impSpec.Name.Name, oldPath)
				}
			} else {
				// No alias
				fmt.Printf("%s Deleting import '%s'.\n", logPrefix, oldPath)
				if astutil.DeleteImport(fset, astFile, oldPath) {
					deleted = true
					fmt.Printf("%s Successfully deleted import %s\n", logPrefix, oldPath)
					break // Assume only one import spec matches
				} else {
					fmt.Printf("%s [WARN] astutil.DeleteImport returned false for '%s'.\n", logPrefix, oldPath)
				}
			}
		}
	}
	if !deleted {
		fmt.Printf("%s [WARN] Expected to delete import '%s' but failed to find/delete it via astutil.\n", logPrefix, oldPath)
		// Continue, maybe it was already gone or AddImport will handle conflicts
	}

	// Add new imports
	failedToAddCount := 0
	failedPaths := []string{}
	sortedNewPaths := make([]string, 0, len(newImports)) // Sort for deterministic order
	for newPath := range newImports {
		sortedNewPaths = append(sortedNewPaths, newPath)
	}
	sort.Strings(sortedNewPaths)

	for _, newPath := range sortedNewPaths {
		fmt.Printf("%s Adding import: %s\n", logPrefix, newPath)
		added := astutil.AddImport(fset, astFile, newPath)
		if !added {
			exists := false
			for _, imp := range astFile.Imports {
				if imp.Path != nil && strings.Trim(imp.Path.Value, `"`) == newPath {
					exists = true
					break
				}
			}
			if !exists {
				fmt.Printf("%s [WARN] astutil add command returned false for import '%s'. It might conflict.\n", logPrefix, newPath)
				failedToAddCount++
				failedPaths = append(failedPaths, fmt.Sprintf("'%s'", newPath))
			} else {
				fmt.Printf("%s [DEBUG] Import '%s' already exists or added successfully.\n", logPrefix, newPath)
			}
		} else {
			fmt.Printf("%s Successfully added import %s\n", logPrefix, newPath)
		}
	}

	if failedToAddCount > 0 {
		errMsg := fmt.Sprintf("failed to add %d required imports due to potential conflicts: %s", failedToAddCount, strings.Join(failedPaths, ", "))
		fmt.Printf("%s [ERROR] %s\n", logPrefix, errMsg)
		return errors.New(errMsg) // Use errors.New
	}

	// Ensure imports are organized
	ast.SortImports(fset, astFile)

	return nil
}

// (Keep other helpers if they exist)
