// NeuroScript Version: 0.5.0
// File version: 3
// Purpose: Allow sub-packages of the 'api' package to be imported.
// filename: cmd/ng/imports_test.go
// nlines: 75
// risk_rating: LOW

package main

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNgImports ensures that the 'ng' command only imports allowed packages
// from the neuroscript project, maintaining a clean architectural boundary.
func TestNgImports(t *testing.T) {
	allowedImports := map[string]bool{
		"github.com/aprice2704/neuroscript/pkg/api":     true,
		"github.com/aprice2704/neuroscript/pkg/neurogo": true, // Allowed for flag helpers
		"github.com/aprice2704/neuroscript/pkg/types":   true, // Often needed for basic types
		"github.com/aprice2704/neuroscript/pkg/logging": true, // For logger interfaces/setup
		"github.com/aprice2704/neuroscript/pkg/version": true, // For version info
	}

	projectPrefix := "github.com/aprice2704/neuroscript/pkg/"
	apiPrefix := "github.com/aprice2704/neuroscript/pkg/api/"

	// Get the current directory
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	// Read all go files in the current directory
	files, err := filepath.Glob(filepath.Join(dir, "*.go"))
	if err != nil {
		t.Fatalf("Failed to glob for go files: %v", err)
	}

	fset := token.NewFileSet()

	for _, file := range files {
		// Skip test files to avoid cycles
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		node, err := parser.ParseFile(fset, file, nil, parser.ImportsOnly)
		if err != nil {
			t.Errorf("Failed to parse %s: %v", file, err)
			continue
		}

		for _, imp := range node.Imports {
			path := strings.Trim(imp.Path.Value, `"`)

			// Check if it's a project import that needs validation
			if strings.HasPrefix(path, projectPrefix) {
				if _, isAllowed := allowedImports[path]; isAllowed {
					continue // Exactly allowed.
				}

				// Also allow sub-packages of the 'api' package.
				if strings.HasPrefix(path, apiPrefix) {
					continue
				}

				t.Errorf("File '%s' has a disallowed import: '%s'. It should interact via the 'api' or 'neurogo' packages.", filepath.Base(file), path)
			}
		}
	}
}
