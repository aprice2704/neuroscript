// NeuroScript Go Indexer - Parser
// File version: 2.2.7 // Relaxed filter for unexported functions to aid tool linking.
// Purpose: Parses Go source files and extracts detailed information for the index.
// filename: cmd/goindexer/parser.go
package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aprice2704/neuroscript/pkg/goindex"
)

// processFile parses a single Go file and populates the given ComponentIndex.
// repoRootPath is the absolute path to the root of the repository on the filesystem.
// currentRepoModulePath is the base module path for the repository (e.g., "github.com/user/repo").
// sourceFileRelToComponent is the path of the file relative to its component's root.
func processFile(fset *token.FileSet, absFilePath string, repoRootPath string, currentRepoModulePath string, sourceFileRelToComponent string, componentIndex *goindex.ComponentIndex) {
	node, err := parser.ParseFile(fset, absFilePath, nil, parser.ParseComments)
	if err != nil {
		log.Printf("    Error parsing file %q: %v", absFilePath, err)
		return
	}

	pkgNameFromAST := ""
	if node.Name != nil {
		pkgNameFromAST = node.Name.Name
	} else {
		return
	}

	pkgDirAbs := filepath.Dir(absFilePath)
	var fullPackagePath string

	if repoRootPath == "" {
		log.Panicf("repoRootPath is not set in call to processFile for %s. This is required.", absFilePath)
	}

	pkgDirRelToRoot, err := filepath.Rel(repoRootPath, pkgDirAbs)
	if err != nil {
		log.Printf("    Warning: Could not determine relative package dir for %s using repoRootPath %s: %v. This might lead to incorrect package paths.", absFilePath, repoRootPath, err)
		pkgDirInComponent := filepath.Dir(sourceFileRelToComponent)
		if componentIndex.ComponentPath != "" && componentIndex.ComponentPath != "." {
			fullPackagePath = filepath.ToSlash(filepath.Join(currentRepoModulePath, componentIndex.ComponentPath, pkgDirInComponent))
		} else {
			fullPackagePath = filepath.ToSlash(filepath.Join(currentRepoModulePath, pkgDirInComponent))
		}
		if pkgDirInComponent == "." {
			if componentIndex.ComponentPath != "" && componentIndex.ComponentPath != "." {
				fullPackagePath = filepath.ToSlash(filepath.Join(currentRepoModulePath, componentIndex.ComponentPath))
			} else {
				fullPackagePath = currentRepoModulePath
			}
		}
		fullPackagePath = filepath.Clean(fullPackagePath)

	} else {
		fullPackagePath = filepath.ToSlash(filepath.Join(currentRepoModulePath, pkgDirRelToRoot))
	}
	if strings.HasSuffix(fullPackagePath, "/.") {
		fullPackagePath = strings.TrimSuffix(fullPackagePath, "/.")
	}
	if pkgDirRelToRoot == "." && fullPackagePath == filepath.ToSlash(currentRepoModulePath+"/.") { // Handle root package of module
		fullPackagePath = currentRepoModulePath
	}
	fullPackagePath = filepath.Clean(fullPackagePath)

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

	ast.Inspect(node, func(n ast.Node) bool {
		if n == nil {
			return false
		}

		switch decl := n.(type) {
		case *ast.FuncDecl:
			simpleFuncName := decl.Name.Name
			isFuncExported := ast.IsExported(simpleFuncName)

			// MODIFIED FILTER:
			// Index all top-level functions. The IsExported field will denote visibility.
			// This ensures tool-implementing functions (exported or not) are indexed.
			// If you still want to skip certain unexported functions not related to tools,
			// you might need a more sophisticated condition here.
			// For now, to ensure toolStringConcat and toolListTools are caught:
			// (Original: if !isFuncExported && !strings.HasPrefix(simpleFuncName, "init") { return true } )
			// No filter based on export status here anymore for *all* top-level functions.
			// If the function is "init", it will also be processed.

			if simpleFuncName == "_" { // Skip blank identifier functions
				return true
			}

			params := formatFieldList(fset, decl.Type.Params)
			var returns []string
			if decl.Type.Results != nil {
				for _, field := range decl.Type.Results.List {
					typeNameStr := formatNode(fset, field.Type) // Uses formatters.go
					if len(field.Names) > 0 {
						for range field.Names {
							returns = append(returns, typeNameStr)
						}
					} else {
						returns = append(returns, typeNameStr)
					}
				}
			}

			fqn := fullPackagePath + "." + simpleFuncName
			if decl.Recv == nil { // Function
				// Enhanced logging for specific functions of interest
				//	debugMsg := ""
				// if simpleFuncName == "toolStringConcat" || simpleFuncName == "toolListTools" {
				// 	debugMsg = "--> FOUND AND ADDING TOOL CANDIDATE: "
				// }

				currentPkgDetail.Functions = append(currentPkgDetail.Functions, goindex.FunctionDetail{
					Name:       fqn,
					SourceFile: sourceFileRelToComponent,
					Parameters: params,
					Returns:    returns,
					IsExported: isFuncExported,
					IsMethod:   false,
				})
				//	log.Printf("    [PARSER_DEBUG] %sAdded Function: FQN=%s (Exported: %v) (File: %s, Pkg: %s)", debugMsg, fqn, isFuncExported, sourceFileRelToComponent, fullPackagePath)
			} else { // Method
				_, receiverBaseTypeStr := formatReceiver(fset, decl.Recv.List[0])
				methodFQN := fmt.Sprintf("%s.(%s).%s", fullPackagePath, receiverBaseTypeStr, simpleFuncName)

				currentPkgDetail.Methods = append(currentPkgDetail.Methods, goindex.MethodDetail{
					ReceiverName: formatReceiverName(decl.Recv.List[0]),
					ReceiverType: receiverBaseTypeStr,
					Name:         simpleFuncName,
					FQN:          methodFQN,
					SourceFile:   sourceFileRelToComponent,
					Parameters:   params,
					Returns:      returns,
					IsExported:   isFuncExported,
				})
				// log.Printf("    [PARSER_DEBUG] Added Method: FQN=%s (Simple: %s on %s) (File: %s, Pkg: %s)", methodFQN, simpleFuncName, receiverBaseTypeStr, sourceFileRelToComponent, fullPackagePath)
			}

		case *ast.GenDecl:
			for _, spec := range decl.Specs {
				switch s := spec.(type) {
				case *ast.TypeSpec:
					typeName := s.Name.Name
					// isTypeExported := ast.IsExported(typeName) // Already captured below
					typeFQN := fullPackagePath + "." + typeName

					switch typeKind := s.Type.(type) {
					case *ast.StructType:
						var fields []goindex.FieldDetail
						if typeKind.Fields != nil {
							for _, field := range typeKind.Fields.List {
								fieldTypeStr := formatNode(fset, field.Type)
								var fieldTagStr string
								if field.Tag != nil {
									fieldTagStr, _ = strconv.Unquote(field.Tag.Value)
								}
								if len(field.Names) > 0 {
									for _, fieldName := range field.Names {
										fieldIsExported := ast.IsExported(fieldName.Name)
										fields = append(fields, goindex.FieldDetail{
											Name:     fieldName.Name,
											Type:     fieldTypeStr,
											Tags:     fieldTagStr,
											Exported: fieldIsExported,
										})
									}
								} else {
									baseTypeName := getBaseTypeName(fieldTypeStr)
									fields = append(fields, goindex.FieldDetail{
										Name:     baseTypeName,
										Type:     fieldTypeStr,
										Tags:     fieldTagStr,
										Exported: ast.IsExported(baseTypeName),
									})
								}
							}
						}
						currentPkgDetail.Structs = append(currentPkgDetail.Structs, goindex.StructDetail{
							Name:       typeName,
							FQN:        typeFQN,
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
											interfaceMethodFQN := typeFQN + "." + methodName
											interfaceMethods = append(interfaceMethods, goindex.MethodDetail{
												Name:       methodName,
												FQN:        interfaceMethodFQN,
												SourceFile: sourceFileRelToComponent,
												Parameters: mParams,
												Returns:    mReturns,
												IsExported: ast.IsExported(methodName),
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
							FQN:        typeFQN,
							SourceFile: sourceFileRelToComponent,
							Methods:    interfaceMethods,
						})
					default:
						// Only add to TypeAliases if it's an exported type, common practice.
						// If you need unexported type aliases, remove this ast.IsExported check.
						if ast.IsExported(typeName) {
							underlyingTypeStr := formatNode(fset, s.Type)
							currentPkgDetail.TypeAliases = append(currentPkgDetail.TypeAliases, goindex.TypeAliasDetail{
								Name:           typeName,
								FQN:            typeFQN,
								UnderlyingType: underlyingTypeStr,
								SourceFile:     sourceFileRelToComponent,
							})
						}
					}

				case *ast.ValueSpec: // Variables and Constants
					typeStr := ""
					if s.Type != nil {
						typeStr = formatNode(fset, s.Type)
					}
					for i, nameIdent := range s.Names {
						name := nameIdent.Name
						// For vars and consts, usually only exported ones are of interest for an index.
						if !ast.IsExported(name) {
							continue
						}
						fqn := fullPackagePath + "." + name
						valueStr := ""
						if i < len(s.Values) {
							if basicLit, ok := s.Values[i].(*ast.BasicLit); ok {
								valueStr = basicLit.Value
							}
						}

						if decl.Tok == token.VAR {
							currentPkgDetail.GlobalVars = append(currentPkgDetail.GlobalVars, goindex.GlobalVarDetail{
								Name:       name,
								FQN:        fqn,
								Type:       typeStr,
								SourceFile: sourceFileRelToComponent,
								Value:      valueStr,
							})
						} else if decl.Tok == token.CONST {
							currentPkgDetail.GlobalConsts = append(currentPkgDetail.GlobalConsts, goindex.GlobalConstDetail{
								Name:       name,
								FQN:        fqn,
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
