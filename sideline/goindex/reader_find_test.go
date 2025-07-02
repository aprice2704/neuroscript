// filename: pkg/goindex/reader_find_test.go
package goindex

import (
	"strings"
	"testing"
)

func TestFindFunction(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()
	reader, _ := NewIndexReader("test_project_root", indexDir)
	if err := reader.LoadProjectIndex(); err != nil {
		t.Fatalf("Failed to load project index: %v", err)
	}
	if err := reader.LoadAllComponentIndexes(); err != nil { // Ensure components are loaded
		t.Fatalf("LoadAllComponentIndexes failed: %v", err)
	}

	componentName := "core"
	packageName := "example.com/testproj/pkg/core"
	funcName := "PublicFunction"

	fn, err := reader.FindFunction(componentName, packageName, funcName)
	if err != nil {
		t.Fatalf("FindFunction(%s, %s, %s) failed: %v", componentName, packageName, funcName, err)
	}
	if fn == nil {
		t.Fatalf("FindFunction returned nil function detail for %s.%s", packageName, funcName)
	}
	expectedFQN := "example.com/testproj/pkg/core.PublicFunction"
	if fn.Name != expectedFQN {
		t.Errorf("Expected function FQN '%s', got '%s'", expectedFQN, fn.Name)
	}
	if fn.SourceFile != "public.go" {
		t.Errorf("Expected source file 'public.go', got '%s'", fn.SourceFile)
	}
}

func TestFindStruct(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()
	reader, _ := NewIndexReader("test_project_root", indexDir)
	if err := reader.LoadProjectIndex(); err != nil {
		t.Fatalf("Failed to load project index: %v", err)
	}
	if err := reader.LoadAllComponentIndexes(); err != nil { // Ensure components are loaded
		t.Fatalf("LoadAllComponentIndexes failed: %v", err)
	}

	coreCompName := "core"
	corePkgPath := "example.com/testproj/pkg/core"
	coreStructName := "CoreStruct"

	strct, err := reader.FindStruct(coreCompName, corePkgPath, coreStructName)
	if err != nil {
		t.Fatalf("FindStruct(%s, %s, %s) failed: %v", coreCompName, corePkgPath, coreStructName, err)
	}
	if strct == nil {
		t.Fatalf("FindStruct returned nil struct detail for %s", coreStructName)
	}
	if strct.Name != coreStructName {
		t.Errorf("Expected struct name '%s', got '%s'", coreStructName, strct.Name)
	}
	expectedFQN := "example.com/testproj/pkg/core.CoreStruct"
	if strct.FQN != expectedFQN {
		t.Errorf("Expected struct FQN '%s', got '%s'", expectedFQN, strct.FQN)
	}
	if strct.SourceFile != "structs.go" {
		t.Errorf("Expected source file 'structs.go', got '%s'", strct.SourceFile)
	}

	anotherCompName := "anothercomp"
	anotherPkgPath := "example.com/testproj/pkg/another"
	anotherStructName := "UtilType"

	strctAnother, err := reader.FindStruct(anotherCompName, anotherPkgPath, anotherStructName)
	if err != nil {
		t.Fatalf("FindStruct(%s, %s, %s) failed: %v", anotherCompName, anotherPkgPath, anotherStructName, err)
	}
	if strctAnother == nil {
		t.Fatalf("FindStruct returned nil struct detail for %s", anotherStructName)
	}
	if strctAnother.Name != anotherStructName {
		t.Errorf("Expected struct name '%s', got '%s'", anotherStructName, strctAnother.Name)
	}
	expectedAnotherFQN := "example.com/testproj/pkg/another.UtilType"
	if strctAnother.FQN != expectedAnotherFQN {
		t.Errorf("Expected struct FQN '%s', got '%s'", expectedAnotherFQN, strctAnother.FQN)
	}
	if strctAnother.SourceFile != "othertypes.go" {
		t.Errorf("Expected source file 'othertypes.go', got '%s'", strctAnother.SourceFile)
	}
}

func TestFindMethod(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()
	reader, _ := NewIndexReader("test_project_root", indexDir)
	if err := reader.LoadProjectIndex(); err != nil {
		t.Fatalf("Failed to load project index: %v", err)
	}
	if err := reader.LoadAllComponentIndexes(); err != nil {
		t.Fatalf("Failed to load component indexes for method search: %v", err)
	}

	tests := []struct {
		name                 string
		componentName        string
		packagePath          string
		methodFQNToSearch    string
		expectedMethodName   string
		expectedReceiverType string
		expectedSourceFile   string
		expectFound          bool
	}{
		{
			name:                 "Pointer receiver method in core",
			componentName:        "core",
			packagePath:          "example.com/testproj/pkg/core",
			methodFQNToSearch:    "example.com/testproj/pkg/core.(*CoreStruct).GetValue",
			expectedMethodName:   "GetValue",
			expectedReceiverType: "*CoreStruct",
			expectedSourceFile:   "structs.go",
			expectFound:          true,
		},
		{
			name:                 "Value receiver method in core",
			componentName:        "core",
			packagePath:          "example.com/testproj/pkg/core",
			methodFQNToSearch:    "example.com/testproj/pkg/core.(CoreStruct).SetValue",
			expectedMethodName:   "SetValue",
			expectedReceiverType: "CoreStruct",
			expectedSourceFile:   "structs.go",
			expectFound:          true,
		},
		{
			name:                 "Method for a type in 'another' package",
			componentName:        "anothercomp",
			packagePath:          "example.com/testproj/pkg/another",
			methodFQNToSearch:    "example.com/testproj/pkg/another.(UtilType).Process",
			expectedMethodName:   "Process",
			expectedReceiverType: "UtilType",
			expectedSourceFile:   "othertypes.go",
			expectFound:          true,
		},
		{
			name:              "Non-existent method FQN in core",
			componentName:     "core",
			packagePath:       "example.com/testproj/pkg/core",
			methodFQNToSearch: "example.com/testproj/pkg/core.(*CoreStruct).NonExistentMethod",
			expectFound:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			method, err := reader.FindMethod(tt.componentName, tt.packagePath, tt.methodFQNToSearch)
			if tt.expectFound {
				if err != nil {
					t.Fatalf("FindMethod for FQN '%s' failed: %v", tt.methodFQNToSearch, err)
				}
				if method == nil {
					t.Fatalf("Expected method for FQN '%s', got nil", tt.methodFQNToSearch)
				}
				if method.FQN != tt.methodFQNToSearch {
					t.Errorf("Expected method FQN '%s', got '%s'", tt.methodFQNToSearch, method.FQN)
				}
				if method.Name != tt.expectedMethodName {
					t.Errorf("Expected method simple name '%s', got '%s'", tt.expectedMethodName, method.Name)
				}
				if method.SourceFile != tt.expectedSourceFile {
					t.Errorf("Expected source file '%s', got '%s'", tt.expectedSourceFile, method.SourceFile)
				}
				if method.ReceiverType != tt.expectedReceiverType {
					t.Errorf("Expected receiver type (simple) '%s', got '%s'", tt.expectedReceiverType, method.ReceiverType)
				}
			} else {
				if err == nil {
					t.Errorf("Expected error for FQN %s, got nil method", tt.methodFQNToSearch)
				} else if !strings.Contains(strings.ToLower(err.Error()), "not found") {
					t.Errorf("Expected a 'not found' error for FQN %s, got: %v", tt.methodFQNToSearch, err)
				}
			}
		})
	}
}
