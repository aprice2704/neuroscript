// filename: pkg/goindex/reader_load_test.go
package goindex

import (
	"testing"
)

func TestLoadProjectIndexMethod(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()

	reader, err := NewIndexReader("test_project_root", indexDir)
	if err != nil {
		t.Fatalf("NewIndexReader failed: %v", err)
	}

	err = reader.LoadProjectIndex()
	if err != nil {
		t.Fatalf("reader.LoadProjectIndex() failed: %v", err)
	}

	projectIndex := reader.GetProjectIndex()
	if projectIndex == nil {
		t.Fatal("reader.GetProjectIndex() returned nil after successful load")
	}

	if projectIndex.ProjectRootModulePath != "example.com/testproj" {
		t.Errorf("Expected module path 'example.com/testproj', got '%s'", projectIndex.ProjectRootModulePath)
	}
	if len(projectIndex.Components) != 3 {
		t.Errorf("Expected 3 components, got %d", len(projectIndex.Components))
	}
	if _, ok := projectIndex.Components["core"]; !ok {
		t.Error("Component 'core' not found in project index")
	}
}

func TestGetComponentIndex(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()

	reader, err := NewIndexReader("test_project_root", indexDir)
	if err != nil {
		t.Fatalf("NewIndexReader failed: %v", err)
	}
	if err := reader.LoadProjectIndex(); err != nil {
		t.Fatalf("Failed to load project index: %v", err)
	}

	coreIndex, err := reader.GetComponentIndex("core")
	if err != nil {
		t.Fatalf("GetComponentIndex('core') failed: %v", err)
	}
	if coreIndex.ComponentName != "core" {
		t.Errorf("Expected component name 'core', got '%s'", coreIndex.ComponentName)
	}
	if len(coreIndex.Packages) != 1 {
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
