// filename: pkg/core/tools_go_ast_package_helpers.go
package core

import (
	// Removed "bytes" import as no longer needed here
	"fmt"
	"go/ast"

	// Removed "go/format"
	// Removed "go/parser"
	"go/token"
	// Removed "os"
	// Removed "os/exec"
	// Removed "path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	// analyzeImportsAndSymbols moved to tools_go_ast_analyze.go
	// processGoFileForImportUpdate removed
	// determineRefactoredDir removed
)

// fileModificationStatus - Can potentially remove this if status handling is fully within main loop
type fileModificationStatus string

// ... const definitions ... (can be removed if not used)

// applyAstImportChanges modifies the import declarations in the AST.
// (Implementation remains the same as previous step)
func applyAstImportChanges(fset *token.FileSet, astFile *ast.File, oldPath string, newImports map[string]string) error {
	fmt.Printf("[DEBUG applyAstImportChanges] Applying import changes: remove %s, add %d new paths\n", oldPath, len(newImports))
	didDelete := astutil.DeleteImport(fset, astFile, oldPath)
	if !didDelete {
		fmt.Printf("[WARN applyAstImportChanges] Expected to delete import '%s' but DeleteImport returned false.\n", oldPath)
	} else {
		fmt.Printf("[DEBUG applyAstImportChanges] Deleted import '%s'.\n", oldPath)
	}
	failedAdds := []string{}
	addedCount := 0
	for path, alias := range newImports {
		var added bool
		if alias != "" {
			added = astutil.AddNamedImport(fset, astFile, alias, path)
			fmt.Printf("[DEBUG applyAstImportChanges] Adding named import: %s as %s\n", path, alias)
		} else {
			added = astutil.AddImport(fset, astFile, path)
			fmt.Printf("[DEBUG applyAstImportChanges] Adding import: %s\n", path)
		}
		if added {
			addedCount++
		} else {
			failedAdd := fmt.Sprintf("'%s'", path)
			if alias != "" {
				failedAdd = fmt.Sprintf("'%s' as '%s'", path, alias)
			}
			failedAdds = append(failedAdds, failedAdd)
			fmt.Printf("[WARN applyAstImportChanges] astutil add command returned false for import %s. It might already exist or conflict.\n", failedAdd)
		}
	}
	if len(failedAdds) > 0 {
		fmt.Printf("[WARN applyAstImportChanges] Failed to add %d required imports (might be pre-existing/duplicates): %s\n", len(failedAdds), strings.Join(failedAdds, ", "))
	}
	return nil
}
