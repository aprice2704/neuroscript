// main.go
// This program scans all Go source files in the current directory, parses them,
// and outputs a JSON object containing the names of the constants, types,
// package-level variables, and functions declared in each file.
package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// FileDeclarations holds the lists of symbols found in a single Go file.
// The `json:"..."` tags are used by the json encoder to name the fields in the output.
type FileDeclarations struct {
	Constants []string `json:"constants"`
	Variables []string `json:"variables"`
	Types     []string `json:"types"`
	Functions []string `json:"functions"`
}

// NewFileDeclarations creates and initializes a FileDeclarations struct.
func NewFileDeclarations() *FileDeclarations {
	return &FileDeclarations{
		Constants: make([]string, 0),
		Variables: make([]string, 0),
		Types:     make([]string, 0),
		Functions: make([]string, 0),
	}
}

func main() {
	// Get the current working directory.
	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get current directory: %v", err)
	}

	// Read all entries in the current directory.
	files, err := os.ReadDir(wd)
	if err != nil {
		log.Fatalf("Failed to read directory contents: %v", err)
	}

	// This map will hold all the declarations, with the filename as the key.
	// It will be the final structure to be converted to JSON.
	allDeclarations := make(map[string]*FileDeclarations)

	// A FileSet is needed by the parser to associate tokens with file positions.
	fset := token.NewFileSet()

	// Iterate over each file/directory entry.
	for _, file := range files {
		// We only care about files that end with the ".go" extension.
		// We also exclude this file itself from the output.
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".go") && file.Name() != "main.go" {
			filePath := filepath.Join(wd, file.Name())

			// Parse the Go source file. The parser returns an abstract syntax tree (AST).
			node, err := parser.ParseFile(fset, filePath, nil, 0)
			if err != nil {
				log.Printf("Error parsing file %s: %v", file.Name(), err)
				continue
			}

			// Extract declarations from the AST and store them in our map.
			allDeclarations[file.Name()] = extractDeclarations(node.Decls)
		}
	}

	// Marshal the map into a pretty-printed JSON byte slice.
	output, err := json.MarshalIndent(allDeclarations, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal data to JSON: %v", err)
	}

	// Print the final JSON to standard output.
	fmt.Println(string(output))
}

// extractDeclarations iterates through the AST declarations of a file
// and returns a FileDeclarations struct containing the found symbols.
func extractDeclarations(decls []ast.Decl) *FileDeclarations {
	fileDecls := NewFileDeclarations()

	// A top-level declaration can be a function (FuncDecl) or a
	// general declaration (GenDecl) for consts, vars, types, and imports.
	for _, decl := range decls {
		switch d := decl.(type) {
		case *ast.FuncDecl:
			// This is a function declaration.
			if d.Name != nil {
				fileDecls.Functions = append(fileDecls.Functions, d.Name.Name)
			}
		case *ast.GenDecl:
			// This is a general declaration (import, const, var, type).
			for _, spec := range d.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					// This is a type declaration (`type MyType int`).
					if s.Name != nil {
						fileDecls.Types = append(fileDecls.Types, s.Name.Name)
					}
				case *ast.ValueSpec:
					// This is a constant or variable declaration.
					// `d.Tok` tells us if it's `const` or `var`.
					for _, name := range s.Names {
						// s.Names is a list of identifiers.
						if d.Tok == token.CONST {
							fileDecls.Constants = append(fileDecls.Constants, name.Name)
						} else if d.Tok == token.VAR {
							fileDecls.Variables = append(fileDecls.Variables, name.Name)
						}
					}
				}
			}
		}
	}
	return fileDecls
}
