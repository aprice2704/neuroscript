// NeuroScript Version: 0.3.0
// File version: 0.1.9
// Purpose: Update test expectations for Concat; guide verification for ListTools.
// filename: pkg/goindex/reader_tools_test.go
// nlines: 190 // Approximate
// risk_rating: MEDIUM
package goindex

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/logging"
)

// TestGetNeuroScriptTool remains the same, using the mock index.
func TestGetNeuroScriptTool(t *testing.T) {
	indexDir, cleanup := setupTestIndex(t)
	defer cleanup()
	reader, _ := NewIndexReader("test_project_root", indexDir)
	if err := reader.LoadProjectIndex(); err != nil {
		t.Fatalf("Failed to load project index: %v", err)
	}
	if _, err := reader.GetComponentIndex("core"); err != nil {
		t.Fatalf("Failed to load 'core' component: %v", err)
	}
	componentName := "core"
	toolName := "Core.ToolFromIndexDirectly"
	tool, err := reader.GetNeuroScriptTool(componentName, toolName)
	if err != nil {
		t.Fatalf("GetNeuroScriptTool(%s, %s) failed: %v", componentName, toolName, err)
	}
	if tool == nil {
		t.Fatalf("GetNeuroScriptTool(%s, %s) returned nil tool detail", componentName, toolName)
	}
	if tool.Name != toolName {
		t.Errorf("Expected tool name '%s', got '%s'", toolName, tool.Name)
	}
}

func TestGenerateEnhancedToolDetails_WithSelectedRealTools(t *testing.T) {
	actualProjectRoot := "../.."
	actualIndexDir := filepath.Join(actualProjectRoot, "pkg/codebase-indices/")

	reader, err := NewIndexReader(actualProjectRoot, actualIndexDir)
	if err != nil {
		t.Fatalf("NewIndexReader for actual index failed: %v", err)
	}
	if err := reader.LoadProjectIndex(); err != nil {
		t.Fatalf("LoadProjectIndex for actual index failed: %v. Ensure index exists at '%s'", err, actualIndexDir)
	}
	if err := reader.LoadAllComponentIndexes(); err != nil {
		t.Fatalf("LoadAllComponentIndexes for actual index failed: %v", err)
	}

	coreComponentIndex, err := reader.GetComponentIndex("core")
	if err != nil {
		t.Logf("[DEBUG] Could not load 'core' component index for debugging: %v", err)
	} else if coreComponentIndex != nil {
		targetPackagePathKey := "github.com/aprice2704/neuroscript/pkg/core"
		projectIdx := reader.GetProjectIndex()
		moduleRelativePkgPath := "pkg/core"

		foundPkgDetailKey := ""
		if projectIdx != nil && projectIdx.ProjectRootModulePath != "" {
			potentialKey := filepath.ToSlash(filepath.Join(projectIdx.ProjectRootModulePath, moduleRelativePkgPath))
			if _, exists := coreComponentIndex.Packages[potentialKey]; exists {
				targetPackagePathKey = potentialKey
				foundPkgDetailKey = potentialKey
			}
		}
		if foundPkgDetailKey == "" { // Fallback if module path based key not found
			if _, exists := coreComponentIndex.Packages[targetPackagePathKey]; exists {
				foundPkgDetailKey = targetPackagePathKey
			}
		}

		if pkgDetail, exists := coreComponentIndex.Packages[foundPkgDetailKey]; exists && pkgDetail != nil && pkgDetail.Functions != nil {
			t.Logf("[DEBUG] Functions FQNs found in index for component 'core', package '%s':", foundPkgDetailKey)
			for _, fn := range pkgDetail.Functions {
				t.Logf("  - %s (Source: %s)", fn.Name, fn.SourceFile)
			}
		} else {
			t.Logf("[DEBUG] No functions found in index for component 'core', package '%s' (key used: '%s'). Or package/functions list is nil.", moduleRelativePkgPath, foundPkgDetailKey)
			t.Logf("[DEBUG] Available packages in 'core' component index ('%s'):", coreComponentIndex.ComponentName)
			for pkgFQN := range coreComponentIndex.Packages {
				t.Logf("  - %s", pkgFQN)
			}
		}
	}

	var logger logging.Logger
	llmClient := adapters.NewNoOpLLMClient()
	workspacePath := t.TempDir()
	var configMap map[string]interface{}
	var bootstrapScripts []string

	interpreter, err := core.NewInterpreter(
		logger, llmClient, workspacePath, configMap, bootstrapScripts,
	)
	if err != nil {
		t.Fatalf("core.NewInterpreter failed: %v", err)
	}

	numInterpreterTools := len(interpreter.ListTools())
	if numInterpreterTools == 0 {
		t.Fatalf("Interpreter has no tools listed. Check tool registration in core.")
	}
	t.Logf("Interpreter has %d tools listed. Processing details...", numInterpreterTools)

	details, errGED := reader.GenerateEnhancedToolDetails(interpreter)
	if errGED != nil {
		t.Fatalf("GenerateEnhancedToolDetails returned an unexpected error: %v", errGED)
	}
	if len(details) != numInterpreterTools {
		t.Errorf("Expected %d details (matching interpreter tools), got %d", numInterpreterTools, len(details))
	}

	type toolCheck struct {
		specName                               string
		expectedRuntimeFQN                     string
		expectedSourceFileNotEmpty             bool
		expectedComponentPath                  string
		expectedGoImplementingFuncReturnsError bool
	}

	checks := []toolCheck{
		{
			specName:                               "Concat",
			expectedRuntimeFQN:                     "github.com/aprice2704/neuroscript/pkg/core.toolStringConcat",
			expectedSourceFileNotEmpty:             true, // Changed: Expecting it to be found now
			expectedComponentPath:                  "pkg/core",
			expectedGoImplementingFuncReturnsError: true, // Because func toolStringConcat(...) (interface{}, error)
		},
		{
			specName:                               "Meta.ListTools",
			expectedRuntimeFQN:                     "github.com/aprice2704/neuroscript/pkg/core.toolListTools",
			expectedSourceFileNotEmpty:             false, // Keep as false until user confirms it's in index like Concat
			expectedComponentPath:                  "pkg/core",
			expectedGoImplementingFuncReturnsError: true, // Verify: Assume (interface{}, error) or similar
		},
		{
			specName:                               "AIWorkerDefinition.Get",
			expectedRuntimeFQN:                     "github.com/aprice2704/neuroscript/pkg/core.init.func6",
			expectedSourceFileNotEmpty:             false,
			expectedComponentPath:                  "pkg/core",
			expectedGoImplementingFuncReturnsError: true, // Verify: Assume (*Def, error) or similar
		},
	}

	foundChecks := 0
	for _, check := range checks {
		t.Run(fmt.Sprintf("CheckingTool_%s", check.specName), func(t *testing.T) {
			var detailForCheck *NeuroScriptToolDetail
			for i := range details {
				if details[i].Name == check.specName {
					detailForCheck = &details[i]
					break
				}
			}

			if detailForCheck == nil {
				t.Errorf("Tool '%s' not found in generated details", check.specName)
				return
			}
			foundChecks++

			if detailForCheck.ImplementingGoFunctionFullName != check.expectedRuntimeFQN {
				t.Errorf("Expected runtime FQN '%s', got '%s'", check.expectedRuntimeFQN, detailForCheck.ImplementingGoFunctionFullName)
			}

			if check.expectedSourceFileNotEmpty {
				if detailForCheck.ImplementingGoFunctionSourceFile == "" {
					t.Errorf("Tool '%s': Expected SourceFile to be non-empty, but it was. Linking failed. (Runtime FQN used for lookup: '%s')",
						check.specName, detailForCheck.ImplementingGoFunctionFullName)
				}
				// Only check component path if source file was expected and found
				if detailForCheck.ComponentPath != check.expectedComponentPath {
					t.Errorf("Tool '%s': Expected ComponentPath '%s', got '%s'",
						check.specName, check.expectedComponentPath, detailForCheck.ComponentPath)
				}
			} else { // If we didn't expect to find the source file (e.g., known indexing issue or init.funcX)
				if detailForCheck.ImplementingGoFunctionSourceFile != "" {
					t.Logf("INFO: Tool '%s' (runtime FQN: %s) unexpectedly HAD SourceFile: '%s' and ComponentPath: '%s'. This is good if correct!",
						check.specName, detailForCheck.ImplementingGoFunctionFullName, detailForCheck.ImplementingGoFunctionSourceFile, detailForCheck.ComponentPath)
				}
			}

			// Check GoImplementingFunctionReturnsError
			// This is most reliable if the function was found (SourceFileNotEmpty is true)
			if detailForCheck.GoImplementingFunctionReturnsError != check.expectedGoImplementingFuncReturnsError {
				if check.expectedSourceFileNotEmpty && detailForCheck.ImplementingGoFunctionSourceFile != "" { // Be strict if we expected to find it and did
					t.Errorf("Tool '%s': Expected GoImplementingFunctionReturnsError to be %v, got %v",
						check.specName, check.expectedGoImplementingFuncReturnsError, detailForCheck.GoImplementingFunctionReturnsError)
				} else if !check.expectedSourceFileNotEmpty { // If we didn't expect to find source, this is more of an observation
					t.Logf("INFO: Tool '%s': GoImplementingFunctionReturnsError is %v (expected based on assumption: %v). Source file not linked.",
						check.specName, detailForCheck.GoImplementingFunctionReturnsError, check.expectedGoImplementingFuncReturnsError)
				}
			}
		})
	}

	if foundChecks != len(checks) {
		t.Errorf("Expected to find and check %d specific tools, but only processed %d", len(checks), foundChecks)
	}
	t.Logf("Finished checking %d specific tools. Other warnings about unlinked tools are expected for functions not perfectly matched by the indexer.", foundChecks)
}
