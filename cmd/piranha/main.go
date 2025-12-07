// :: product: FDM/NS
// :: majorVersion: 1
// :: fileVersion: 7
// :: description: A query-based Go symbol scanner with glob/multi-query support.
// :: latestChange: Added -C flag for case-sensitive matching; made case-insensitive the default.
// :: filename: piranha/main.go
// :: serialization: go

package main

import (
	"encoding/csv"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// --- Configuration ---
// Add glob patterns for any paths (files or dirs) you want to skip.
// Uses filepath.Match syntax.
var excludedPatterns = []string{
	".*", // Skip all hidden files and dirs (.git, .vscode)
	"vendor",
	"node_modules",
	"pkg/antlr", // Skip ANTLR-generated files
	"build",
	"bin",
	"third_party",
	"third-party",
	"thirdparty",
	"testdata",
	"test-fixtures",
	"test_fixtures",
	"test-data",
	"__pycache__",
	".venv",
	"venv",
	".terraform",
	"dist",
	"coverage*",
	"*_test.go", // Skip all test files
}

const helpText = `Piranha: NeuroScript Go Symbol Finder

Finds Go symbol **definitions** (not usages) in a repository.
It outputs a CSV (path, package, kind, symbol) for all
exported symbols and unexported functions/methods,
skipping paths matching glob patterns.

USAGE:
  piranha [flags] [q1] [q2]...

FLAGS:
  -C                Enable case-sensitive matching (default is case-insensitive)
  -h, --help        Show this help message

QUERY SYNTAX:
  Queries use glob matching (e.g., "Load*", "*Unit", "api.*").
  By default, matching is CASE-INSENSITIVE. Use -C to respect case.

EXAMPLE (for Gemini):
  To find 'LoadFromUnit' and 'ExecWithInterpreter' (insensitive), run:
  
  piranha *loadfromunit* *execwithinterpreter*`

// --- End Configuration ---

func main() {
	// Configure logging to stderr to keep stdout clean for CSV
	log.SetOutput(os.Stderr)
	log.SetFlags(0)

	// --- Argument Parsing ---
	var queries []string
	caseSensitive := false

	if len(os.Args) > 1 {
		for _, arg := range os.Args[1:] {
			if arg == "-h" || arg == "--help" {
				fmt.Println(helpText)
				os.Exit(0)
			}
			if arg == "-C" {
				caseSensitive = true
				continue
			}
			queries = append(queries, arg)
		}
	}

	// Set up CSV writer to stdout
	csvWriter := csv.NewWriter(os.Stdout)
	if len(queries) == 0 {
		// Only write header if we are dumping the whole file (no queries)
		if err := csvWriter.Write([]string{"path", "package", "kind", "symbol"}); err != nil {
			log.Fatalf("Failed to write CSV header: %v", err)
		}
	}

	root := "."
	fset := token.NewFileSet()

	// Walk the directory tree starting from current dir
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// --- Path Filtering ---
		if path != "." {
			for _, pattern := range excludedPatterns {
				if matched, _ := filepath.Match(pattern, path); matched {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil // Skip file
				}
				if matched, _ := filepath.Match(pattern, d.Name()); matched {
					if d.IsDir() {
						return filepath.SkipDir
					}
					return nil // Skip file
				}
			}
		}

		if d.IsDir() {
			return nil // It's a directory we want to scan
		}

		if !strings.HasSuffix(path, ".go") {
			return nil // Skip non-Go files
		}

		// --- File Parsing ---
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			log.Printf("ERROR: Failed to parse %s: %v", path, err)
			return nil // Continue walking
		}

		// --- Symbol Processing ---
		if err := processFile(path, f, csvWriter, queries, caseSensitive); err != nil {
			log.Printf("ERROR: Failed to process file %s: %v", path, err)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("FATAL: Error walking directory: %v", err)
	}

	csvWriter.Flush()
	if err := csvWriter.Error(); err != nil {
		log.Fatalf("FATAL: Failed to flush CSV writer: %v", err)
	}
}

// matchesQueries checks if a symbol matches ANY of the provided glob queries.
func matchesQueries(queries []string, name, symbolName string, caseSensitive bool) bool {
	if len(queries) == 0 {
		return true // No queries means dump all
	}

	for _, query := range queries {
		targetName := name
		targetSymbol := symbolName
		matchQuery := query

		if !caseSensitive {
			targetName = strings.ToLower(name)
			targetSymbol = strings.ToLower(symbolName)
			matchQuery = strings.ToLower(query)
		}

		// Check against the simple name
		if matched, _ := filepath.Match(matchQuery, targetName); matched {
			return true
		}
		// Check against the fully qualified (if exported) name
		if matched, _ := filepath.Match(matchQuery, targetSymbol); matched {
			return true
		}
	}
	return false // No query matched
}

// processFile extracts symbols and writes them if they match the queries.
func processFile(path string, f *ast.File, writer *csv.Writer, queries []string, caseSensitive bool) error {
	pkgName := f.Name.Name

	for _, decl := range f.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// RULE: KEEP all funcs/methods, exported or not.
			name := d.Name.Name
			symbolName := getSymbolName(pkgName, name)

			if !matchesQueries(queries, name, symbolName, caseSensitive) {
				continue
			}

			kind := "fn"
			if d.Recv != nil && len(d.Recv.List) > 0 {
				kind = "method"
			}
			if err := writer.Write([]string{path, pkgName, kind, symbolName}); err != nil {
				return err
			}

		case *ast.GenDecl:
			// Handle var, const, type
			kind := d.Tok.String() // "var", "const", "type"

			for _, spec := range d.Specs {
				var names []*ast.Ident

				switch s := spec.(type) {
				case *ast.TypeSpec:
					// RULE: KEEP exported types only.
					if ast.IsExported(s.Name.Name) {
						names = append(names, s.Name)
					}
				case *ast.ValueSpec:
					// RULE: KEEP exported vars/consts only.
					for _, nameIdent := range s.Names {
						if ast.IsExported(nameIdent.Name) {
							names = append(names, nameIdent)
						}
					}
				}

				// Process all names that passed the filter
				for _, nameIdent := range names {
					name := nameIdent.Name
					symbolName := getSymbolName(pkgName, name)

					if !matchesQueries(queries, name, symbolName, caseSensitive) {
						continue
					}

					if err := writer.Write([]string{path, pkgName, kind, symbolName}); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

// getSymbolName formats the symbol as "package.Name" if exported, or "name" if not.
func getSymbolName(pkgName, name string) string {
	if ast.IsExported(name) {
		return fmt.Sprintf("%s.%s", pkgName, name)
	}
	// Per your request, unexported symbols do not get the package prefix
	return name
}
