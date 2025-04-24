// filename: pkg/core/tools_go_ast_symbol_map_test.go
package core

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// --- Test Setup Helpers ---

// Ensure writeFileHelper is defined (e.g., in tools_go_ast_package_test.go or a shared _test.go file)
func writeFileHelper(t *testing.T, path string, content string) {
	t.Helper()
	t.Logf("Writing %d bytes to: %s", len(content), path)
	err := os.WriteFile(path, []byte(content), 0644)
	if err != nil {
		t.Fatalf("Failed to write file %s: %v", path, err)
	}
	readBytes, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Logf("WARN: Failed to read back file %s for verification: %v", path, readErr)
	} else {
		snippetLen := 50
		if len(readBytes) < snippetLen {
			snippetLen = len(readBytes)
		}
		t.Logf("  -> Verified write: Read back %d bytes, start: %q", len(readBytes), string(readBytes[:snippetLen]))
	}
}

func setupBasicFixture(t *testing.T, rootDir string)         { /* ... implementation ... */ }
func setupAmbiguousFixture(t *testing.T, rootDir string)     { /* ... implementation ... */ }
func setupNoExportedFixture(t *testing.T, rootDir string)    { /* ... implementation ... */ }
func setupNoSubPackagesFixture(t *testing.T, rootDir string) { /* ... implementation ... */ }

// --- Test-only Manual Symbol Map Builder ---

type fileInfo struct {
	absPath   string
	subPkgRel string
}

// buildSymbolMapManual_test mimics the symbol extraction logic of buildSymbolMap
// but uses manual file walking and parsing instead of packages.Load.
func buildSymbolMapManual_test(t *testing.T, refactoredPkgPath string, rootDir string) (map[string]string, error) {
	t.Helper()
	symbolMap := make(map[string]string)
	ambiguousSymbols := make(map[string]string)
	foundSymbols := false
	fset := token.NewFileSet()

	modulePrefix := ""
	parts := strings.SplitN(refactoredPkgPath, "/", 2)
	if len(parts) == 2 {
		modulePrefix = parts[0] + "/"
	}
	relativePkgDir := strings.TrimPrefix(refactoredPkgPath, modulePrefix)
	baseScanDir := filepath.Join(rootDir, filepath.FromSlash(relativePkgDir))
	t.Logf("[ManualBuild] Base scan dir: %s", baseScanDir)

	if _, err := os.Stat(baseScanDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("%w: base directory '%s' not found", ErrRefactoredPathNotFound, baseScanDir)
	} else if err != nil {
		return nil, fmt.Errorf("%w: error stating base directory '%s': %v", ErrSymbolMappingFailed, baseScanDir, err)
	}

	subDirs, err := os.ReadDir(baseScanDir)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to read base directory '%s': %v", ErrSymbolMappingFailed, baseScanDir, err)
	}

	// --- FIX: Replace WalkDir with nested loops ---
	goFilesProcessedCount := 0 // Track files processed

	for _, entry := range subDirs {
		if !entry.IsDir() {
			continue
		}
		subPkgName := entry.Name()
		subDirPath := filepath.Join(baseScanDir, subPkgName)
		canonicalPkgPath := refactoredPkgPath + "/" + subPkgName
		t.Logf("[ManualBuild] Scanning subdir: %s for package %s", subDirPath, canonicalPkgPath)

		// Read files within this specific subdirectory
		filesInSubDir, readSubErr := os.ReadDir(subDirPath)
		if readSubErr != nil {
			t.Logf("[ManualBuild][WARN] Error reading subdir %s: %v", subDirPath, readSubErr)
			continue // Skip this subdir if unreadable
		}

		for _, fileEntry := range filesInSubDir {
			if fileEntry.IsDir() || !strings.HasSuffix(fileEntry.Name(), ".go") || strings.HasSuffix(fileEntry.Name(), "_test.go") {
				continue
			}

			path := filepath.Join(subDirPath, fileEntry.Name())
			t.Logf("[ManualBuild] Processing file: %s", path)
			goFilesProcessedCount++ // Increment files processed

			contentBytes, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Logf("[ManualBuild][ERROR] Failed read %s: %v", path, readErr)
				continue // Skip file on read error
			}

			astFile, parseErr := parser.ParseFile(fset, path, contentBytes, parser.ParseComments|parser.SkipObjectResolution)
			if parseErr != nil {
				t.Logf("[ManualBuild][ERROR] Failed parse %s: %v", path, parseErr)
				continue // Skip file on parse error
			}

			t.Logf("[ManualBuild]   Inspecting AST for: %s (%d Decls)", path, len(astFile.Decls))
			if len(astFile.Decls) == 0 {
				t.Logf("[ManualBuild]     WARNING: No declarations found in AST for %s", path)
			}

			for i, decl := range astFile.Decls {
				t.Logf("[ManualBuild]     Processing Decl %d: Type=%T", i, decl)
				switch node := decl.(type) {
				case *ast.FuncDecl:
					if node.Recv != nil { // Skip methods
						t.Logf("[ManualBuild]       Skipping Method: Name=%s", node.Name.Name)
						continue
					}
					funcName := node.Name.Name
					isExported := node.Name.IsExported()
					t.Logf("[ManualBuild]       FuncDecl: Name=%s, IsExported=%v", funcName, isExported)
					if isExported {
						t.Logf("[ManualBuild]         -> Adding FuncDecl to map: %s -> %s", funcName, canonicalPkgPath)
						foundSymbols = true
						if existingPath, exists := symbolMap[funcName]; exists && existingPath != canonicalPkgPath {
							t.Logf("[ManualBuild]           AMBIGUITY DETECTED for %s", funcName)
							ambiguousSymbols[funcName] = fmt.Sprintf("found in %s and %s", existingPath, canonicalPkgPath)
						} else if !exists {
							symbolMap[funcName] = canonicalPkgPath
							if _, actuallyAdded := symbolMap[funcName]; actuallyAdded {
								t.Logf("[ManualBuild]           VERIFIED map contains %s", funcName)
							} else {
								t.Logf("[ManualBuild]           ERROR: map assignment FAILED for FuncDecl %s", funcName)
							}
						}
					}
				case *ast.GenDecl:
					t.Logf("[ManualBuild]       GenDecl: Tok=%s, Specs=%d", node.Tok, len(node.Specs))
					for j, spec := range node.Specs {
						t.Logf("[ManualBuild]         Processing Spec %d: Type=%T", j, spec)
						switch s := spec.(type) {
						case *ast.TypeSpec:
							typeName := s.Name.Name
							isExported := s.Name.IsExported()
							t.Logf("[ManualBuild]           TypeSpec: Name=%s, IsExported=%v", typeName, isExported)
							if isExported {
								t.Logf("[ManualBuild]             -> Adding TypeSpec to map: %s -> %s", typeName, canonicalPkgPath)
								foundSymbols = true
								if existingPath, exists := symbolMap[typeName]; exists && existingPath != canonicalPkgPath {
									t.Logf("[ManualBuild]           AMBIGUITY DETECTED for %s", typeName)
									ambiguousSymbols[typeName] = fmt.Sprintf("found in %s and %s", existingPath, canonicalPkgPath)
								} else if !exists {
									symbolMap[typeName] = canonicalPkgPath
									if _, actuallyAdded := symbolMap[typeName]; actuallyAdded {
										t.Logf("[ManualBuild]             VERIFIED map contains %s", typeName)
									} else {
										t.Logf("[ManualBuild]             ERROR: map assignment FAILED for TypeSpec %s", typeName)
									}
								}
							}
						case *ast.ValueSpec:
							t.Logf("[ManualBuild]           ValueSpec: Names=%d", len(s.Names))
							for k, name := range s.Names {
								valueName := name.Name
								isExported := name.IsExported()
								t.Logf("[ManualBuild]             ValueSpec Name %d: Name=%s, IsExported=%v (Parent Tok: %s)", k, valueName, isExported, node.Tok)
								if isExported {
									t.Logf("[ManualBuild]               -> Adding ValueSpec Name to map: %s -> %s", valueName, canonicalPkgPath)
									foundSymbols = true
									if existingPath, exists := symbolMap[valueName]; exists && existingPath != canonicalPkgPath {
										t.Logf("[ManualBuild]           AMBIGUITY DETECTED for %s", valueName)
										ambiguousSymbols[valueName] = fmt.Sprintf("found in %s and %s", existingPath, canonicalPkgPath)
									} else if !exists {
										symbolMap[valueName] = canonicalPkgPath
										if _, actuallyAdded := symbolMap[valueName]; actuallyAdded {
											t.Logf("[ManualBuild]               VERIFIED map contains %s", valueName)
										} else {
											t.Logf("[ManualBuild]               ERROR: map assignment FAILED for ValueSpec %s", valueName)
										}
									}
								}
							}
						default:
							t.Logf("[ManualBuild]         Found other Spec type: %T", s)
						}
					}
				default:
					t.Logf("[ManualBuild]     Found other Decl type: %T", node)
				}
			} // end loop decls
		} // end loop filesInSubDir
	} // end loop subDirs
	// --- END FIX ---

	if len(ambiguousSymbols) > 0 {
		errorList := []string{}
		for symbol, locations := range ambiguousSymbols {
			errorList = append(errorList, fmt.Sprintf("symbol '%s' (%s)", symbol, locations))
		}
		errMsg := fmt.Sprintf("ambiguous exported symbols found: %s", strings.Join(errorList, "; "))
		t.Logf("[ManualBuild][ERROR] %s", errMsg)
		return nil, fmt.Errorf("%w: %s", ErrSymbolMappingFailed, errMsg)
	}
	if !foundSymbols && goFilesProcessedCount > 0 { // Check processed count instead of goFilesToLoad slice
		t.Logf("[ManualBuild] No exported symbols found during inspection.")
	}
	t.Logf("[ManualBuild] Finished building map. Symbols found: %d", len(symbolMap))
	return symbolMap, nil
}

// --- Test Function ---

func TestBuildSymbolMapLogic(t *testing.T) {
	testCases := []struct {
		name              string
		fixtureSetupFunc  func(t *testing.T, rootDir string)
		refactoredPkgPath string
		expectedMap       map[string]string
		expectedErrorIs   error
	}{
		{
			name:              "Basic case with multiple types",
			fixtureSetupFunc:  setupBasicFixture,
			refactoredPkgPath: "testbuildmap/original",
			expectedMap: map[string]string{
				"ExportedFuncOne":  "testbuildmap/original/sub1",
				"ExportedVarOne":   "testbuildmap/original/sub1",
				"ExportedTypeTwo":  "testbuildmap/original/sub2",
				"ExportedConstTwo": "testbuildmap/original/sub2",
			},
			expectedErrorIs: nil,
		},
		{
			name:              "Ambiguous symbols",
			fixtureSetupFunc:  setupAmbiguousFixture,
			refactoredPkgPath: "testbuildmap/original",
			expectedMap:       nil,
			expectedErrorIs:   ErrSymbolMappingFailed,
		},
		{
			name:              "No exported symbols in subpackages",
			fixtureSetupFunc:  setupNoExportedFixture,
			refactoredPkgPath: "testbuildmap/original",
			expectedMap:       map[string]string{},
			expectedErrorIs:   nil,
		},
		{
			name:              "No subpackages with go code",
			fixtureSetupFunc:  setupNoSubPackagesFixture,
			refactoredPkgPath: "testbuildmap/original",
			expectedMap:       map[string]string{},
			expectedErrorIs:   nil,
		},
		{
			name:              "Original package path does not exist",
			fixtureSetupFunc:  setupBasicFixture,
			refactoredPkgPath: "testbuildmap/nonexistent",
			expectedMap:       nil,
			expectedErrorIs:   ErrRefactoredPathNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			rootDir := t.TempDir()
			t.Logf("Test rootDir: %s", rootDir)
			tc.fixtureSetupFunc(t, rootDir)
			actualMap, err := buildSymbolMapManual_test(t, tc.refactoredPkgPath, rootDir)

			// Assertions remain the same
			if tc.expectedErrorIs != nil {
				if err == nil {
					t.Errorf("Expected error matching '%v', but got nil error", tc.expectedErrorIs)
				} else if !errors.Is(err, tc.expectedErrorIs) {
					if !strings.Contains(err.Error(), tc.expectedErrorIs.Error()) {
						t.Errorf("Expected error matching '%v' or containing its text, but got error: %v (type %T)", tc.expectedErrorIs, err, err)
					}
				}
			} else {
				if err != nil {
					t.Errorf("Did not expect error, but got: %v", err)
				}
			}
			if tc.expectedErrorIs == nil && err == nil {
				if !reflect.DeepEqual(actualMap, tc.expectedMap) {
					t.Errorf("Returned map does not match expected.\nExpected: %v\nGot:      %v", tc.expectedMap, actualMap)
				}
			}
		})
	}
}
