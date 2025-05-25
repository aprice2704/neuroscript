// NeuroScript Go Indexer - Index Reader Tests
// File version: 1.0.3 // Corrected mock data structure and test for cross-package conceptual method.
// filename: pkg/goindex/reader_test.go
package goindex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// setupTestIndex creates a temporary directory with mock index files for testing.
// It returns the path to the temporary directory and a cleanup function.
func setupTestIndex(t *testing.T) (indexDir string, cleanup func()) {
	tmpDir, err := os.MkdirTemp("", "goindex_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	now := time.Now().Format(time.RFC3339)
	gitBranch := "test-branch"
	gitCommit := "testcommit123"

	// Mock ProjectIndex data
	projectIdx := ProjectIndex{
		ProjectRootModulePath: "example.com/testproj",
		IndexSchemaVersion:    "project_index_v2.0.0",
		LastIndexedTimestamp:  now,
		GitBranch:             gitBranch,
		GitCommitHash:         gitCommit,
		Components: map[string]ComponentIndexFileEntry{
			"core": {
				Name:        "core",
				Path:        "pkg/core", // This is ComponentRelPath for ComponentDefinition
				IndexFile:   "core_index.json",
				Description: "Core functionality",
			},
			"utils": { // Added a separate component for utils to make pkg/another more distinct
				Name:        "utils",
				Path:        "pkg/utils",
				IndexFile:   "utils_index.json",
				Description: "Utility functions",
			},
			"anothercomp": { // Component that would "own" pkg/another
				Name:        "anothercomp",
				Path:        "pkg/another",
				IndexFile:   "anothercomp_index.json",
				Description: "Another component",
			},
		},
	}
	projectIdxData, _ := json.MarshalIndent(projectIdx, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "project_index.json"), projectIdxData, 0644); err != nil {
		t.Fatalf("Failed to write project_index.json: %v", err)
	}

	// Mock Core ComponentIndex data
	coreComponentIdx := ComponentIndex{
		ComponentName:      "core",
		ComponentPath:      "pkg/core",
		IndexSchemaVersion: "component_index_v2.0.0",
		LastIndexed:        now,
		GitBranch:          gitBranch,
		GitCommitHash:      gitCommit,
		Packages: map[string]*PackageDetail{
			"example.com/testproj/pkg/core": {
				PackagePath: "example.com/testproj/pkg/core",
				PackageName: "core",
				Functions: []FunctionDetail{
					{Name: "example.com/testproj/pkg/core.PublicFunction", SourceFile: "public.go", Parameters: []ParamDetail{{Name: "input", Type: "string"}}, Returns: []string{"string", "error"}},
				},
				Structs: []StructDetail{
					{Name: "CoreStruct", SourceFile: "structs.go", Fields: []FieldDetail{{Name: "ID", Type: "int", Exported: true, Tags: `json:"id"`}}},
				},
				Methods: []MethodDetail{
					{ReceiverName: "s", ReceiverType: "*example.com/testproj/pkg/core.CoreStruct", Name: "GetValue", SourceFile: "structs.go", Parameters: []ParamDetail{}, Returns: []string{"int"}},
					{ReceiverName: "s", ReceiverType: "example.com/testproj/pkg/core.CoreStruct", Name: "SetValue", SourceFile: "structs.go", Parameters: []ParamDetail{{Name: "id", Type: "int"}}, Returns: []string{}},
				},
				TypeAliases: []TypeAliasDetail{
					{Name: "CoreID", UnderlyingType: "int", SourceFile: "aliases.go"},
				},
			},
		},
		NeuroScriptTools: []NeuroScriptToolDetail{
			{Name: "Core.Tool1", Description: "A core tool", Category: "CoreUtils", Example: "TOOL.Core.Tool1()", ImplementingGoFunctionFullName: "example.com/testproj/pkg/core.tool1ImplFunc", GoImplementingFunctionReturnsError: true, ReturnType: "string"},
		},
	}
	coreComponentData, _ := json.MarshalIndent(coreComponentIdx, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "core_index.json"), coreComponentData, 0644); err != nil {
		t.Fatalf("Failed to write core_index.json: %v", err)
	}

	// Mock anothercomp ComponentIndex data
	anotherComponentIdx := ComponentIndex{
		ComponentName:      "anothercomp",
		ComponentPath:      "pkg/another",
		IndexSchemaVersion: "component_index_v2.0.0",
		LastIndexed:        now,
		GitBranch:          gitBranch,
		GitCommitHash:      gitCommit,
		Packages: map[string]*PackageDetail{
			"example.com/testproj/pkg/another": {
				PackagePath: "example.com/testproj/pkg/another",
				PackageName: "another",
				Structs: []StructDetail{
					{Name: "UtilType", SourceFile: "othertypes.go"},
				},
				Methods: []MethodDetail{
					{ReceiverName: "ut", ReceiverType: "example.com/testproj/pkg/another.UtilType", Name: "Process", SourceFile: "othertypes.go", Parameters: []ParamDetail{}, Returns: []string{"bool"}},
				},
			},
		},
	}
	anotherComponentData, _ := json.MarshalIndent(anotherComponentIdx, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "anothercomp_index.json"), anotherComponentData, 0644); err != nil {
		t.Fatalf("Failed to write anothercomp_index.json: %v", err)
	}

	// Mock Utils ComponentIndex data (minimal, no methods for this test)
	utilsComponentIdx := ComponentIndex{
		ComponentName:      "utils",
		ComponentPath:      "pkg/utils",
		IndexSchemaVersion: "component_index_v2.0.0",
		LastIndexed:        now,
		Packages:           map[string]*PackageDetail{},
	}
	utilsComponentData, _ := json.MarshalIndent(utilsComponentIdx, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "utils_index.json"), utilsComponentData, 0644); err != nil {
		t.Fatalf("Failed to write utils_index.json: %v", err)
	}

	return tmpDir, func() {
		os.RemoveAll(tmpDir)
	}
}

func TestLoadProjectIndex(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()

	projectIndex, err := LoadProjectIndex(filepath.Join(indexDir, "project_index.json"))
	if err != nil {
		t.Fatalf("LoadProjectIndex failed: %v", err)
	}

	if projectIndex.ProjectRootModulePath != "example.com/testproj" {
		t.Errorf("Expected module path 'example.com/testproj', got '%s'", projectIndex.ProjectRootModulePath)
	}
	if len(projectIndex.Components) != 3 { // core, utils, anothercomp
		t.Errorf("Expected 3 components, got %d", len(projectIndex.Components))
	}
	if _, ok := projectIndex.Components["core"]; !ok {
		t.Error("Component 'core' not found in project index")
	}
	if _, ok := projectIndex.Components["anothercomp"]; !ok {
		t.Error("Component 'anothercomp' not found in project index")
	}
}

func TestGetComponentIndex(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()

	reader, err := NewIndexReader("test_project_root", indexDir)
	if err != nil {
		t.Fatalf("NewIndexReader failed: %v", err)
	}

	coreIndex, err := reader.GetComponentIndex("core")
	if err != nil {
		t.Fatalf("GetComponentIndex('core') failed: %v", err)
	}
	if coreIndex.ComponentName != "core" {
		t.Errorf("Expected component name 'core', got '%s'", coreIndex.ComponentName)
	}
	if len(coreIndex.Packages) != 1 { // Only core package defined within core_index.json
		t.Errorf("Expected 1 package in core component, got %d", len(coreIndex.Packages))
	}

	anothercompIndex, err := reader.GetComponentIndex("anothercomp")
	if err != nil {
		t.Fatalf("GetComponentIndex('anothercomp') failed: %v", err)
	}
	if anothercompIndex.ComponentName != "anothercomp" {
		t.Errorf("Expected component name 'anothercomp', got '%s'", anothercompIndex.ComponentName)
	}
	if len(anothercompIndex.Packages) != 1 {
		t.Errorf("Expected 1 package in anothercomp component, got %d", len(anothercompIndex.Packages))
	}
	if _, ok := anothercompIndex.Packages["example.com/testproj/pkg/another"]; !ok {
		t.Errorf("Package 'example.com/testproj/pkg/another' not found in 'anothercomp' component index")
	}

	// Test caching
	coreIndexAgain, err := reader.GetComponentIndex("core")
	if err != nil {
		t.Fatalf("GetComponentIndex('core') second call failed: %v", err)
	}
	if coreIndexAgain != coreIndex {
		t.Error("Expected GetComponentIndex to return cached instance")
	}

	_, err = reader.GetComponentIndex("nonexistent")
	if err == nil {
		t.Error("Expected error for nonexistent component, got nil")
	}
}

func TestFindFunction(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()
	reader, _ := NewIndexReader("test_project_root", indexDir)

	fqn := "example.com/testproj/pkg/core.PublicFunction"
	fn, comp, pkg, err := reader.FindFunction(fqn)
	if err != nil {
		t.Fatalf("FindFunction(%s) failed: %v", fqn, err)
	}
	if fn == nil {
		t.Fatalf("FindFunction(%s) returned nil function detail", fqn)
	}
	if fn.Name != fqn {
		t.Errorf("Expected function name '%s', got '%s'", fqn, fn.Name)
	}
	if comp == nil || comp.ComponentName != "core" {
		t.Errorf("Expected component 'core', got '%v'", comp)
	}
	if pkg == nil || pkg.PackageName != "core" {
		t.Errorf("Expected package 'core', got '%v'", pkg)
	}
}

func TestFindStruct(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()
	reader, _ := NewIndexReader("test_project_root", indexDir)

	// Test struct in core package
	coreStructFQN := "example.com/testproj/pkg/core.CoreStruct"
	strct, comp, pkg, err := reader.FindStruct(coreStructFQN)
	if err != nil {
		t.Fatalf("FindStruct(%s) failed: %v", coreStructFQN, err)
	}
	if strct == nil {
		t.Fatalf("FindStruct(%s) returned nil struct detail", coreStructFQN)
	}
	if strct.Name != "CoreStruct" {
		t.Errorf("Expected struct name 'CoreStruct', got '%s'", strct.Name)
	}
	if comp.ComponentName != "core" {
		t.Errorf("Expected component 'core' for CoreStruct, got '%s'", comp.ComponentName)
	}
	if pkg.PackageName != "core" {
		t.Errorf("Expected package 'core' for CoreStruct, got '%s'", pkg.PackageName)
	}

	// Test struct in another package (within anothercomp)
	anotherStructFQN := "example.com/testproj/pkg/another.UtilType"
	strct, comp, pkg, err = reader.FindStruct(anotherStructFQN)
	if err != nil {
		t.Fatalf("FindStruct(%s) failed: %v", anotherStructFQN, err)
	}
	if strct == nil {
		t.Fatalf("FindStruct(%s) returned nil struct detail", anotherStructFQN)
	}
	if strct.Name != "UtilType" {
		t.Errorf("Expected struct name 'UtilType', got '%s'", strct.Name)
	}
	if comp.ComponentName != "anothercomp" {
		t.Errorf("Expected component 'anothercomp' for UtilType, got '%s'", comp.ComponentName)
	}
	if pkg.PackageName != "another" {
		t.Errorf("Expected package 'another' for UtilType, got '%s'", pkg.PackageName)
	}
}

func TestFindMethod(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()
	reader, _ := NewIndexReader("test_project_root", indexDir)

	tests := []struct {
		name               string
		receiverTypeFQN    string
		methodName         string
		expectFound        bool
		expectedMethodName string
		expectedSourceFile string
		expectedPkgName    string
		expectedCompName   string
	}{
		{
			name:               "Pointer receiver method in core",
			receiverTypeFQN:    "*example.com/testproj/pkg/core.CoreStruct",
			methodName:         "GetValue",
			expectFound:        true,
			expectedMethodName: "GetValue",
			expectedSourceFile: "structs.go",
			expectedPkgName:    "core",
			expectedCompName:   "core",
		},
		{
			name:               "Value receiver method in core",
			receiverTypeFQN:    "example.com/testproj/pkg/core.CoreStruct",
			methodName:         "SetValue",
			expectFound:        true,
			expectedMethodName: "SetValue",
			expectedSourceFile: "structs.go",
			expectedPkgName:    "core",
			expectedCompName:   "core",
		},
		{
			name:               "Method for a type in 'another' package (in 'anothercomp' component)",
			receiverTypeFQN:    "example.com/testproj/pkg/another.UtilType",
			methodName:         "Process",
			expectFound:        true,
			expectedMethodName: "Process",
			expectedSourceFile: "othertypes.go",
			expectedPkgName:    "another",
			expectedCompName:   "anothercomp",
		},
		{
			name:            "Non-existent method in core",
			receiverTypeFQN: "*example.com/testproj/pkg/core.CoreStruct",
			methodName:      "NonExistentMethod",
			expectFound:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, comp, pkg, err := reader.FindMethod(tt.receiverTypeFQN, tt.methodName)
			if tt.expectFound {
				if err != nil {
					t.Fatalf("FindMethod failed: %v", err) // This is where the previous failure occurred
				}
				if method == nil {
					t.Fatal("Expected method, got nil")
				}
				if method.Name != tt.expectedMethodName {
					t.Errorf("Expected method name '%s', got '%s'", tt.expectedMethodName, method.Name)
				}
				if method.SourceFile != tt.expectedSourceFile {
					t.Errorf("Expected source file '%s', got '%s'", tt.expectedSourceFile, method.SourceFile)
				}
				if comp == nil || comp.ComponentName != tt.expectedCompName {
					t.Errorf("Expected component '%s', got comp '%v'", tt.expectedCompName, comp)
				}
				if pkg == nil || pkg.PackageName != tt.expectedPkgName {
					t.Errorf("Expected package '%s', got pkg '%v'", tt.expectedPkgName, pkg)
				}
			} else {
				if err == nil {
					t.Error("Expected error for non-existent method/receiver, got nil")
				}
			}
		})
	}
}

func TestGetNeuroScriptTool(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()
	reader, _ := NewIndexReader("test_project_root", indexDir)

	toolName := "Core.Tool1"
	tool, comp, err := reader.GetNeuroScriptTool(toolName)
	if err != nil {
		t.Fatalf("GetNeuroScriptTool(%s) failed: %v", toolName, err)
	}
	if tool == nil {
		t.Fatalf("GetNeuroScriptTool(%s) returned nil tool detail", toolName)
	}
	if tool.Name != toolName {
		t.Errorf("Expected tool name '%s', got '%s'", toolName, tool.Name)
	}
	if comp == nil || comp.ComponentName != "core" {
		t.Errorf("Expected component 'core', got '%v'", comp)
	}
}
