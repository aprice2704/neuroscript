// NeuroScript Go Indexer - Parser
// File version: 2.2.4 // Aligned with types.go v2.0.0 and uses formatters.go
// Purpose: Parses Go source files and extracts detailed information for the index.
// filename: cmd/goindexer/parser.go
package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/goindex" // Ensure this import path is correct
)

// In cmd/goindexer/parser.go

// processFile parses a single Go file and populates the given ComponentIndex.
// repoModulePath is the base module path for the repository (e.g., "github.com/user/repo").
// sourceFileRelToComponent is the path of the file relative to its component's root.
func processFile(fset *token.FileSet, absFilePath string, repoModulePath string, sourceFileRelToComponent string, componentIndex *goindex.ComponentIndex) {
	node, err := parser.ParseFile(fset, absFilePath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("    Error parsing file %q: %v", absFilePath, err)
		return
	}

	pkgNameFromAST := ""
	if node.Name != nil {
		pkgNameFromAST = node.Name.Name
	} else {
		log.Printf("    Skipping file with no package declaration: %q", sourceFileRelToComponent)
		return
	}

	// Determine full package path
	// Uses the package name from AST and its directory relative to module root for full path
	pkgDirAbs := filepath.Dir(absFilePath)
	pkgDirRelToRoot, err := filepath.Rel(repoRootPath, pkgDirAbs)
	if err != nil {
		// Fallback if Rel fails (e.g. absFilePath is not under repoRootPath, though it should be)
		log.Printf("    Warning: Could not determine relative package dir for %s using repoRootPath %s: %v. Falling back.", absFilePath, repoRootPath, err)
		// Fallback: use the directory of the file relative to its component, then join with module path
		// This assumes ComponentPath in ComponentIndex is relative to module root OR repoModulePath itself is the component path
		// This part is tricky and depends heavily on how component paths are defined and used.
		// A simple approach if sourceFileRelToComponent is like "mypackage/myfile.go" for a component like "pkg/mycomponent"
		// is to use the directory part of sourceFileRelToComponent as the package's sub-path within the component.
		pkgDirRelToComponent := filepath.Dir(sourceFileRelToComponent)
		if pkgDirRelToComponent == "." { // Package is at the root of the component
			// If component path is "pkg/core", module path is "github.com/mod", then pkg path is "github.com/mod/pkg/core"
			// This assumes componentIndex.ComponentPath is relative to module path or can be joined.
			// Let's use a more direct approach based on where the file is relative to module root.
			// This was the previous robust logic:
			// pkgDirRelToRoot was already calculated using repoRootPath.
			// If that failed, we need a consistent fallback for fullPackagePath.
			// The most straightforward path is based on the file's directory relative to the module root.
			containingDirRelToComponentRoot := filepath.Dir(sourceFileRelToComponent)
			if componentIndex.ComponentPath != "" && componentIndex.ComponentPath != "." {
				pkgDirRelToRoot = filepath.ToSlash(filepath.Join(componentIndex.ComponentPath, containingDirRelToComponentRoot))
			} else {
				pkgDirRelToRoot = filepath.ToSlash(containingDirRelToComponentRoot)
			}
		}
		// If err is still not nil, this means pkgDirRelToRoot might be unreliable
		log.Printf("    Using derived package directory relative to root: %s", pkgDirRelToRoot)
	}
	fullPackagePath := filepath.ToSlash(filepath.Join(repoModulePath, pkgDirRelToRoot))
	// Ensure it doesn't end with a trailing slash if pkgDirRelToRoot was "."
	if pkgDirRelToRoot == "." || pkgDirRelToRoot == "" {
		fullPackagePath = repoModulePath // Or repoModulePath + "/" + pkgNameFromAST if packages are directly under module root
		// If files are at the very root of the module, their package path is just the module path.
		// However, Go packages usually have their own directory.
		// If pkgNameFromAST is "main" and it's at root, fullPackagePath could be just repoModulePath.
		// This logic often simplifies if actual import paths are resolved.
		// For now: if pkgDirRelToRoot is ".", the full path is the module path itself,
		// and the actual package name (e.g. "main") is within that.
		// The key for the map should be the importable package path.
		// If a file `main.go` is at the root of module `example.com/mymod`, its package path is `example.com/mymod`.
		if pkgDirRelToRoot == "." {
			fullPackagePath = repoModulePath // Package is the module itself
		}
	}

	currentPkgDetail, ok := componentIndex.Packages[fullPackagePath]
	if !ok {
		currentPkgDetail = &goindex.PackageDetail{
			PackagePath:  fullPackagePath,
			PackageName:  pkgNameFromAST,
			Functions:    make([]goindex.FunctionDetail, 0),
			Methods:      make([]goindex.MethodDetail, 0),
			Structs:      make([]goindex.StructDetail, 0),
			Interfaces:   make([]goindex.InterfaceDetail, 0),
			GlobalVars:   make([]goindex.GlobalVarDetail, 0),
			GlobalConsts: make([]goindex.GlobalConstDetail, 0),
			TypeAliases:  make([]goindex.TypeAliasDetail, 0),
		}
		componentIndex.Packages[fullPackagePath] = currentPkgDetail
	}

	// ... (rest of the ast.Inspect and other functions in parser.go remain the same as the previous version I provided)
	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch decl := n.(type) {
		case *ast.FuncDecl:
			funcName := decl.Name.Name
			if !ast.IsExported(funcName) {
				return true
			}

			params := formatFieldList(fset, decl.Type.Params)
			var returns []string
			if decl.Type.Results != nil {
				for _, field := range decl.Type.Results.List {
					typeNameStr := formatNode(fset, field.Type)
					if len(field.Names) > 0 {
						for range field.Names {
							returns = append(returns, typeNameStr)
						}
					} else {
						returns = append(returns, typeNameStr)
					}
				}
			}

			// Construct fully qualified function/method name using the determined fullPackagePath
			qualifiedPrefix := fullPackagePath + "."
			if pkgNameFromAST == "main" && fullPackagePath == repoModulePath { // Special case for main package at module root
				qualifiedPrefix = pkgNameFromAST + "." // e.g. "main.MyFunction" instead of "github.com/mymodule.MyFunction"
			}

			if decl.Recv == nil { // Function
				currentPkgDetail.Functions = append(currentPkgDetail.Functions, goindex.FunctionDetail{
					Name:       qualifiedPrefix + funcName,
					SourceFile: sourceFileRelToComponent,
					Parameters: params,
					Returns:    returns,
				})
			} else { // Method
				recvNameStr, recvTypeStr := formatReceiver(fset, decl.Recv.List[0])
				// Ensure recvTypeStr is fully qualified if it's a local package type
				// This is complex and usually requires go/types. formatNode might give pkg.Type.
				// For simplicity, we assume formatNode gives a usable type string.
				currentPkgDetail.Methods = append(currentPkgDetail.Methods, goindex.MethodDetail{
					ReceiverName: recvNameStr,
					ReceiverType: recvTypeStr, // This should be the fully qualified type if possible
					Name:         funcName,
					SourceFile:   sourceFileRelToComponent,
					Parameters:   params,
					Returns:      returns,
				})
			}

		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					typeName := s.Name.Name
					if !ast.IsExported(typeName) {
						continue
					}

					switch typeKind := s.Type.(type) {
					case *ast.StructType:
						var fields []goindex.FieldDetail
						for _, field := range typeKind.Fields.List {
							fieldTypeStr := formatNode(fset, field.Type)
							var fieldTagStr string
							if field.Tag != nil {
								fieldTagStr, _ = strconv.Unquote(field.Tag.Value)
							}
							if len(field.Names) > 0 {
								for _, fieldName := range field.Names {
									isExported := ast.IsExported(fieldName.Name)
									if isExported {
										fields = append(fields, goindex.FieldDetail{
											Name:     fieldName.Name,
											Type:     fieldTypeStr,
											Tags:     fieldTagStr,
											Exported: isExported,
										})
									}
								}
							} else {
								fields = append(fields, goindex.FieldDetail{
									Name:     getBaseTypeName(fieldTypeStr),
									Type:     fieldTypeStr,
									Tags:     fieldTagStr,
									Exported: true,
								})
							}
						}
						currentPkgDetail.Structs = append(currentPkgDetail.Structs, goindex.StructDetail{
							Name:       typeName,
							SourceFile: sourceFileRelToComponent,
							Fields:     fields,
						})
					case *ast.InterfaceType:
						var interfaceMethods []goindex.MethodDetail
						if typeKind.Methods != nil {
							for _, field := range typeKind.Methods.List {
								if len(field.Names) > 0 {
									for _, methodNameIdent := range field.Names {
										methodName := methodNameIdent.Name
										if funcType, ok := field.Type.(*ast.FuncType); ok {
											mParams := formatFieldList(fset, funcType.Params)
											var mReturns []string
											if funcType.Results != nil {
												for _, resField := range funcType.Results.List {
													resTypeStr := formatNode(fset, resField.Type)
													if len(resField.Names) > 0 {
														for range resField.Names {
															mReturns = append(mReturns, resTypeStr)
														}
													} else {
														mReturns = append(mReturns, resTypeStr)
													}
												}
											}
											interfaceMethods = append(interfaceMethods, goindex.MethodDetail{
												Name:       methodName,
												SourceFile: sourceFileRelToComponent,
												Parameters: mParams,
												Returns:    mReturns,
											})
										}
									}
								} else {
									embeddedInterfaceName := formatNode(fset, field.Type)
									interfaceMethods = append(interfaceMethods, goindex.MethodDetail{
										Name:       embeddedInterfaceName,
										SourceFile: sourceFileRelToComponent,
									})
								}
							}
						}
						currentPkgDetail.Interfaces = append(currentPkgDetail.Interfaces, goindex.InterfaceDetail{
							Name:       typeName,
							SourceFile: sourceFileRelToComponent,
							Methods:    interfaceMethods,
						})
					default:
						underlyingTypeStr := formatNode(fset, s.Type)
						currentPkgDetail.TypeAliases = append(currentPkgDetail.TypeAliases, goindex.TypeAliasDetail{
							Name:           typeName,
							UnderlyingType: underlyingTypeStr,
							SourceFile:     sourceFileRelToComponent,
						})
					}

				case *ast.ValueSpec:
					typeStr := ""
					if s.Type != nil {
						typeStr = formatNode(fset, s.Type)
					}
					for i, nameIdent := range s.Names {
						if !ast.IsExported(nameIdent.Name) {
							continue
						}
						name := nameIdent.Name
						valueStr := ""
						if i < len(s.Values) {
							if basicLit, ok := s.Values[i].(*ast.BasicLit); ok {
								valueStr = basicLit.Value
							}
						}

						if decl.Tok == token.VAR {
							currentPkgDetail.GlobalVars = append(currentPkgDetail.GlobalVars, goindex.GlobalVarDetail{
								Name:       name,
								Type:       typeStr,
								SourceFile: sourceFileRelToComponent,
								Value:      valueStr,
							})
						} else if decl.Tok == token.CONST {
							currentPkgDetail.GlobalConsts = append(currentPkgDetail.GlobalConsts, goindex.GlobalConstDetail{
								Name:       name,
								Type:       typeStr,
								Value:      valueStr,
								SourceFile: sourceFileRelToComponent,
							})
						}
					}
				}
			}
			return true
		}
		return true
	})
}

// getRelativePackagePath function would be defined elsewhere, e.g., in main.go or a utility file.
// For processFile to use repoRootPath, it might need to be passed in or accessed globally.
// Assuming repoRootPath is accessible here for getRelativePackagePath.
// getBaseTypeName extracts the base type name from a potentially qualified or pointer type string.
// e.g., "*pkg.MyType" -> "MyType", "[]pkg.MyType" -> "MyType"
func getBaseTypeName(qualifiedName string) string {
	name := qualifiedName
	if idx := strings.LastIndex(name, "."); idx != -1 {
		name = name[idx+1:]
	}
	name = strings.TrimPrefix(name, "*")
	name = strings.TrimPrefix(name, "[]") // Basic slice handling
	return name
}
