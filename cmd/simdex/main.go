// main.go
// This program scans, parses, and type-checks all Go source files in the specified packages
// (and subdirectories if -recurse is used). It outputs a JSON object containing a per-file
// breakdown of symbols (with line numbers), dependencies, a global symbol index, and a
// report of which types implement which interfaces.
//
// To run this, you will need to install the doublestar library:
// go get github.com/bmatcuk/doublestar/v4
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/bmatcuk/doublestar/v4"
	"golang.org/x/tools/go/packages"
)

// SymbolInfo holds the name and line number for a declared symbol.
type SymbolInfo struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

// InterfaceDef holds the definition of an interface, including its required methods.
type InterfaceDef struct {
	Name    string       `json:"name"`
	Line    int          `json:"line"`
	Methods []SymbolInfo `json:"methods"`
}

// FileDeclarations holds the lists of symbols and imports found in a single Go file.
type FileDeclarations struct {
	Imports    []string       `json:"imports"`
	Constants  []SymbolInfo   `json:"constants"`
	Variables  []SymbolInfo   `json:"variables"`
	Types      []SymbolInfo   `json:"types"`
	Functions  []SymbolInfo   `json:"functions"`
	Interfaces []InterfaceDef `json:"interfaces"`
}

// SymbolIndex maps symbol names to the file path where they are declared.
type SymbolIndex struct {
	Constants  map[string]string `json:"constants"`
	Variables  map[string]string `json:"variables"`
	Types      map[string]string `json:"types"`
	Functions  map[string]string `json:"functions"`
	Interfaces map[string]string `json:"interfaces"`
}

// AnalysisResult is the top-level structure for the JSON output.
type AnalysisResult struct {
	Files           map[string]*FileDeclarations `json:"files"`
	Implementations map[string][]string          `json:"implementations"`
	Index           *SymbolIndex                 `json:"index,omitempty"`
}

// NewSymbolIndex creates and initializes a SymbolIndex struct.
func NewSymbolIndex() *SymbolIndex {
	return &SymbolIndex{
		Constants:  make(map[string]string),
		Variables:  make(map[string]string),
		Types:      make(map[string]string),
		Functions:  make(map[string]string),
		Interfaces: make(map[string]string),
	}
}

// extractDeclarations iterates through the AST and extracts symbol information for a single file.
func extractDeclarations(fset *token.FileSet, file *ast.File) *FileDeclarations {
	fileDecls := &FileDeclarations{
		Imports:    make([]string, 0),
		Constants:  make([]SymbolInfo, 0),
		Variables:  make([]SymbolInfo, 0),
		Types:      make([]SymbolInfo, 0),
		Functions:  make([]SymbolInfo, 0),
		Interfaces: make([]InterfaceDef, 0),
	}

	// Extract imports
	for _, importSpec := range file.Imports {
		path, err := strconv.Unquote(importSpec.Path.Value)
		if err == nil {
			fileDecls.Imports = append(fileDecls.Imports, path)
		}
	}

	for _, decl := range file.Decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			if d.Name != nil {
				fileDecls.Functions = append(fileDecls.Functions, SymbolInfo{
					Name: d.Name.Name,
					Line: fset.Position(d.Pos()).Line,
				})
			}
		case *ast.GenDecl:
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					if s.Name == nil {
						continue
					}
					line := fset.Position(s.Pos()).Line
					if iface, ok := s.Type.(*ast.InterfaceType); ok {
						methods := make([]SymbolInfo, 0)
						if iface.Methods != nil {
							for _, method := range iface.Methods.List {
								if len(method.Names) > 0 {
									for _, name := range method.Names {
										methods = append(methods, SymbolInfo{
											Name: name.Name,
											Line: fset.Position(name.Pos()).Line,
										})
									}
								}
							}
						}
						fileDecls.Interfaces = append(fileDecls.Interfaces, InterfaceDef{
							Name:    s.Name.Name,
							Line:    line,
							Methods: methods,
						})
					} else {
						fileDecls.Types = append(fileDecls.Types, SymbolInfo{
							Name: s.Name.Name,
							Line: line,
						})
					}
				case *ast.ValueSpec:
					line := fset.Position(s.Pos()).Line
					for _, name := range s.Names {
						symbol := SymbolInfo{Name: name.Name, Line: line}
						if d.Tok == token.CONST {
							fileDecls.Constants = append(fileDecls.Constants, symbol)
						} else if d.Tok == token.VAR {
							fileDecls.Variables = append(fileDecls.Variables, symbol)
						}
					}
				}
			}
		}
	}
	return fileDecls
}

func main() {
	// Define command-line flags.
	recurse := flag.Bool("recurse", false, "Scan all subdirectories.")
	revlook := flag.Bool("revlook", false, "Include reverse lookup index (symbol -> file) in the output.")
	exclude := flag.String("exclude", "", "Doublestar glob pattern for file paths to exclude (e.g., 'pkg/parser/**' or '**/*_test.go').")
	flag.Parse()

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	filesByDir := make(map[string][]string)
	addFile := func(path string) {
		if !strings.HasSuffix(path, ".go") {
			return
		}
		if *exclude != "" {
			relPath, err := filepath.Rel(wd, path)
			if err != nil {
				log.Printf("Warning: could not get relative path for %s: %v", path, err)
				relPath = path
			}
			matched, err := doublestar.Match(*exclude, relPath)
			if err != nil {
				log.Printf("Warning: invalid exclude glob pattern '%s': %v", *exclude, err)
			}
			if matched {
				return
			}
		}
		dir := filepath.Dir(path)
		filesByDir[dir] = append(filesByDir[dir], path)
	}

	if *recurse {
		err := filepath.WalkDir(wd, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				addFile(path)
			}
			return nil
		})
		if err != nil {
			log.Fatalf("Error walking directory %s: %v", wd, err)
		}
	} else {
		entries, err := os.ReadDir(wd)
		if err != nil {
			log.Fatalf("Failed to read directory contents: %v", err)
		}
		for _, entry := range entries {
			if !entry.IsDir() {
				addFile(filepath.Join(wd, entry.Name()))
			}
		}
	}

	cfg := &packages.Config{
		Mode:  packages.NeedName | packages.NeedFiles | packages.NeedSyntax | packages.NeedTypes | packages.NeedTypesInfo,
		Tests: true, // Include test files in the analysis.
	}

	var allPkgs []*packages.Package
	for dir := range filesByDir {
		pkgs, err := packages.Load(cfg, dir)
		if err != nil {
			log.Printf("Warning: could not load package for directory %s: %v", dir, err)
			continue
		}
		allPkgs = append(allPkgs, pkgs...)
	}

	result := &AnalysisResult{
		Files:           make(map[string]*FileDeclarations),
		Implementations: make(map[string][]string),
	}
	if *revlook {
		result.Index = NewSymbolIndex()
	}

	var allNamedTypes []*types.Named
	var allInterfaces []*types.Interface

	for _, pkg := range allPkgs {
		if len(pkg.Errors) > 0 {
			log.Printf("Warning: package %s has errors, analysis may be incomplete:", pkg.PkgPath)
			for _, e := range pkg.Errors {
				log.Printf("  - %s", e)
			}
		}

		for i, file := range pkg.Syntax {
			// If a file has a syntax error, its AST might be nil. Skip it.
			if file == nil {
				continue
			}
			filePath := pkg.GoFiles[i]
			relPath, err := filepath.Rel(wd, filePath)
			if err != nil {
				relPath = filePath
			}

			fileDecls := extractDeclarations(pkg.Fset, file)
			result.Files[relPath] = fileDecls

			if result.Index != nil {
				for _, s := range fileDecls.Constants {
					result.Index.Constants[s.Name] = relPath
				}
				for _, s := range fileDecls.Variables {
					result.Index.Variables[s.Name] = relPath
				}
				for _, s := range fileDecls.Types {
					result.Index.Types[s.Name] = relPath
				}
				for _, s := range fileDecls.Functions {
					result.Index.Functions[s.Name] = relPath
				}
				for _, idef := range fileDecls.Interfaces {
					result.Index.Interfaces[idef.Name] = relPath
				}
			}
		}

		// Type analysis for implementation checking might be incomplete if the package had errors,
		// but we proceed to get as much information as possible.
		scope := pkg.Types.Scope()
		for _, name := range scope.Names() {
			if obj := scope.Lookup(name); obj != nil {
				if tn, ok := obj.(*types.TypeName); ok {
					if named, ok := tn.Type().(*types.Named); ok {
						allNamedTypes = append(allNamedTypes, named)
						if iface, ok := named.Underlying().(*types.Interface); ok {
							allInterfaces = append(allInterfaces, iface)
						}
					}
				}
			}
		}
	}

	for _, namedType := range allNamedTypes {
		if _, isIface := namedType.Underlying().(*types.Interface); isIface {
			continue
		}
		if namedType.Obj() == nil || namedType.Obj().Pkg() == nil {
			continue // Skip types without package info (e.g., from incomplete sources).
		}

		for _, iface := range allInterfaces {
			if types.Implements(namedType, iface) {
				for _, otherNamedType := range allNamedTypes {
					if otherNamedType.Underlying() == iface {
						if otherNamedType.Obj() == nil || otherNamedType.Obj().Pkg() == nil {
							continue
						}
						ifaceName := otherNamedType.Obj().Pkg().Path() + "." + otherNamedType.Obj().Name()
						typeName := namedType.Obj().Pkg().Path() + "." + namedType.Obj().Name()
						result.Implementations[ifaceName] = append(result.Implementations[ifaceName], typeName)
						break
					}
				}
			}
		}
	}

	output, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal data to JSON: %v", err)
	}
	fmt.Println(string(output))
}
