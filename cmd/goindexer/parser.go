package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"strings"
	// Requires formatters.go and resolvers.go in the same package
)

// processFile parses a single Go file and adds its info to the index.
// *** MODIFIED: Store only base name for functions in FunctionShortNames ***
func processFile(fset *token.FileSet, filePath string, index *Index) {
	// log.Printf("  Processing file: %s", filePath) // Keep logging minimal now
	node, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("    Error parsing file %q: %v", filePath, err)
		return
	}

	// Determine Paths
	pkgPathRel, err := getRelativePackagePath(filePath) // Uses global repoRootPath
	if err != nil {
		log.Printf("    Warn: Could not determine relative package path for %q: %v", filePath, err)
		if node.Name != nil {
			pkgPathRel = node.Name.Name
		} else {
			log.Printf("    Error: Cannot determine package for %q. Skipping file.", filePath)
			return
		}
	}
	filePathRel, err := getRelativeFilePath(filePath) // Uses global repoRootPath
	if err != nil {
		log.Printf("    Warn: Could not determine relative file path for %q: %v", filePath, err)
		filePathRel = filePath
	}

	// Ensure Map Entries Exist
	if _, ok := index.Packages[pkgPathRel]; !ok {
		index.Packages[pkgPathRel] = PackageInfo{Files: make(map[string]FileInfo)}
	}
	if _, ok := index.Packages[pkgPathRel].Files[filePathRel]; !ok {
		index.Packages[pkgPathRel].Files[filePathRel] = FileInfo{
			Methods: []MethodInfo{},
		}
	}

	var functionShortNames []string
	var methods []MethodInfo

	// Extract Imports
	imports := make(map[string]string)
	for _, imp := range node.Imports {
		fullPath, err := strings.Trim(imp.Path.Value, `"`), error(nil)
		if err != nil {
			log.Printf("    Warn: Could not unquote import path %s: %v", imp.Path.Value, err)
			continue
		}
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		} else {
			parts := strings.Split(fullPath, "/")
			alias = parts[len(parts)-1]
		}
		if alias != "." && alias != "_" {
			imports[alias] = fullPath
		}
	}

	// Traverse AST
	ast.Inspect(node, func(n ast.Node) bool {
		switch decl := n.(type) {
		case *ast.FuncDecl:
			funcName := decl.Name.Name
			if funcName == "" {
				return false
			}

			isMethod := decl.Recv != nil && len(decl.Recv.List) > 0

			if isMethod {
				calls := []string{}
				if decl.Body != nil {
					ast.Inspect(decl.Body, func(callNode ast.Node) bool {
						if callExpr, ok := callNode.(*ast.CallExpr); ok {
							callTarget := resolveCallTarget(fset, callExpr, pkgPathRel, imports)
							if callTarget != "" {
								calls = append(calls, callTarget)
							}
						}
						return true
					})
				}
				receiverStr := formatReceiver(fset, decl.Recv.List[0].Type)
				methods = append(methods, MethodInfo{
					Receiver: receiverStr,
					Name:     funcName,
					Calls:    calls,
				})
			} else {
				// Function - store only the base name
				functionShortNames = append(functionShortNames, funcName) // CHANGED
			}
			return false

		case *ast.GenDecl:
			// Type collection removed previously
			return true
		}
		return true
	})

	// Update the index map entry
	tempFileInfo := index.Packages[pkgPathRel].Files[filePathRel]
	if len(functionShortNames) > 0 {
		tempFileInfo.FunctionShortNames = functionShortNames
	}
	if len(methods) > 0 {
		tempFileInfo.Methods = methods
	}
	index.Packages[pkgPathRel].Files[filePathRel] = tempFileInfo
}
