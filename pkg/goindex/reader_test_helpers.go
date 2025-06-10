// filename: pkg/goindex/reader_test_helpers.go
package goindex

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
	"time"

	"github.com/aprice2704/neuroscript/pkg/adapters"
	"github.com/aprice2704/neuroscript/pkg/core"
	"github.com/aprice2704/neuroscript/pkg/interfaces"
)

var logger interfaces.Logger

func init() {
	logger = adapters.NewNoOpLogger()
}

// --- Mock Go functions for tool implementations, defined at package level for stable FQNs ---
func mockToolImplFuncForTest(i *core.Interpreter, args []interface{}) (interface{}, error) {
	return "mockToolImplFuncForTest_ok", nil
}

type mockToolStruct struct{}

func (m *mockToolStruct) MockToolImplMethodForTest(i *core.Interpreter, args []interface{}) (interface{}, error) {
	return "mockToolImplMethodForTest_ok", nil
}
func mockToolImplWhichIsNotInIndex(i *core.Interpreter, args []interface{}) (interface{}, error) {
	return "mockToolImplWhichIsNotInIndex_ok", nil
}

// --- End Mock Go functions ---

// testNoOpLogger provides a minimal implementation of core.Logger for tests.
type testNoOpLogger struct{}

func (l *testNoOpLogger) Debug(msg string, args ...interface{})  {}
func (l *testNoOpLogger) Info(msg string, args ...interface{})   {}
func (l *testNoOpLogger) Warn(msg string, args ...interface{})   {}
func (l *testNoOpLogger) Error(msg string, args ...interface{})  {}
func (l *testNoOpLogger) Debugf(format string, v ...interface{}) {}
func (l *testNoOpLogger) Infof(format string, v ...interface{})  {}
func (l *testNoOpLogger) Warnf(format string, v ...interface{})  {}
func (l *testNoOpLogger) Errorf(format string, v ...interface{}) {}
func (l *testNoOpLogger) SetLevel(levelStr string)               {} // Assuming core.LogLevel might be represented/set via string

// setupTestIndex creates a temporary directory with mock index files for testing.
// It returns the path to the temporary directory and a cleanup function.
func setupTestIndex(t *testing.T) (indexDir string, cleanup func()) {
	mustParseTime := func(layout, value string) string {
		return "2024-05-25T12:00:00Z"
	}

	tmpDir, err := os.MkdirTemp("", "goindex_test_")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}

	nowStr := mustParseTime(time.RFC3339, "2024-05-25T12:00:00Z")
	gitBranch := "test-branch"
	gitCommit := "testcommit123"

	projectIdx := ProjectIndex{
		ProjectRootModulePath: "example.com/testproj",
		IndexSchemaVersion:    "project_index_v2.0.0",
		LastIndexedTimestamp:  nowStr,
		GitBranch:             gitBranch,
		GitCommitHash:         gitCommit,
		Components: map[string]ComponentIndexFileEntry{
			"core": {
				Name:        "core",
				Path:        "pkg/core",
				IndexFile:   "core_index.json",
				Description: "Core functionality",
			},
			"utils": {
				Name:      "utils",
				Path:      "pkg/utils",
				IndexFile: "utils_index.json",
			},
			"anothercomp": {
				Name:      "anothercomp",
				Path:      "pkg/another",
				IndexFile: "anothercomp_index.json",
			},
		},
	}
	projectIdxData, _ := json.MarshalIndent(projectIdx, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "project_index.json"), projectIdxData, 0644); err != nil {
		t.Fatalf("Failed to write project_index.json: %v", err)
	}

	runtimeFQNToolFunc := runtime.FuncForPC(reflect.ValueOf(mockToolImplFuncForTest).Pointer()).Name()
	runtimeFQNToolMethodCanonical := runtime.FuncForPC(reflect.ValueOf((*mockToolStruct).MockToolImplMethodForTest).Pointer()).Name()

	coreComponentIdx := ComponentIndex{
		ComponentName:      "core",
		ComponentPath:      "pkg/core",
		IndexSchemaVersion: "component_index_v2.0.0",
		LastIndexed:        nowStr,
		GitBranch:          gitBranch,
		GitCommitHash:      gitCommit,
		Packages: map[string]*PackageDetail{
			"example.com/testproj/pkg/core": {
				PackagePath: "example.com/testproj/pkg/core",
				PackageName: "core",
				Functions: []FunctionDetail{
					{Name: "example.com/testproj/pkg/core.PublicFunction", SourceFile: "public.go", Parameters: []ParamDetail{{Name: "input", Type: "string"}}, Returns: []string{"string", "error"}, IsExported: true},
					{Name: "example.com/testproj/pkg/core.tool1ImplFunc", SourceFile: "tools_impl.go", Returns: []string{"string", "error"}, IsExported: true},
					{Name: runtimeFQNToolFunc, SourceFile: "test_tool_funcs.go", Returns: []string{"string", "error"}, IsExported: true},
				},
				Structs: []StructDetail{
					{Name: "CoreStruct", FQN: "example.com/testproj/pkg/core.CoreStruct", SourceFile: "structs.go", Fields: []FieldDetail{{Name: "ID", Type: "int", Exported: true, Tags: `json:"id"`}}},
				},
				Methods: []MethodDetail{
					{ReceiverName: "s", ReceiverType: "*CoreStruct", Name: "GetValue", FQN: "example.com/testproj/pkg/core.(*CoreStruct).GetValue", SourceFile: "structs.go", Parameters: []ParamDetail{}, Returns: []string{"int"}, IsExported: true},
					{ReceiverName: "s", ReceiverType: "CoreStruct", Name: "SetValue", FQN: "example.com/testproj/pkg/core.(CoreStruct).SetValue", SourceFile: "structs.go", Parameters: []ParamDetail{{Name: "id", Type: "int"}}, Returns: []string{}, IsExported: true},
					{ReceiverName: "m", ReceiverType: "*mockToolStruct", Name: "MockToolImplMethodForTest", FQN: runtimeFQNToolMethodCanonical, SourceFile: "test_tool_funcs.go", Returns: []string{"string", "error"}, IsExported: true},
				},
				TypeAliases: []TypeAliasDetail{
					{Name: "CoreID", FQN: "example.com/testproj/pkg/core.CoreID", UnderlyingType: "int", SourceFile: "aliases.go"},
				},
			},
		},
		NeuroScriptTools: []NeuroScriptToolDetail{
			{Name: "Core.ToolFromIndexDirectly", Description: "A core tool from index", ImplementingGoFunctionFullName: "example.com/testproj/pkg/core.tool1ImplFunc", ComponentPath: "pkg/core"},
		},
	}
	coreComponentData, _ := json.MarshalIndent(coreComponentIdx, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "core_index.json"), coreComponentData, 0644); err != nil {
		t.Fatalf("Failed to write core_index.json: %v", err)
	}

	anotherComponentIdx := ComponentIndex{
		ComponentName: "anothercomp", ComponentPath: "pkg/another", IndexSchemaVersion: "component_index_v2.0.0", LastIndexed: nowStr, GitBranch: gitBranch, GitCommitHash: gitCommit,
		Packages: map[string]*PackageDetail{
			"example.com/testproj/pkg/another": {
				PackagePath: "example.com/testproj/pkg/another", PackageName: "another",
				Structs: []StructDetail{{Name: "UtilType", FQN: "example.com/testproj/pkg/another.UtilType", SourceFile: "othertypes.go"}},
				Methods: []MethodDetail{{ReceiverName: "ut", ReceiverType: "UtilType", Name: "Process", FQN: "example.com/testproj/pkg/another.(UtilType).Process", SourceFile: "othertypes.go", IsExported: true}},
			},
		},
	}
	anotherComponentData, _ := json.MarshalIndent(anotherComponentIdx, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "anothercomp_index.json"), anotherComponentData, 0644); err != nil {
		t.Fatalf("Failed to write anothercomp_index.json: %v", err)
	}

	utilsComponentIdx := ComponentIndex{
		ComponentName: "utils", ComponentPath: "pkg/utils", IndexSchemaVersion: "component_index_v2.0.0", LastIndexed: nowStr,
		Packages: map[string]*PackageDetail{},
	}
	utilsComponentData, _ := json.MarshalIndent(utilsComponentIdx, "", "  ")
	if err := os.WriteFile(filepath.Join(tmpDir, "utils_index.json"), utilsComponentData, 0644); err != nil {
		t.Fatalf("Failed to write utils_index.json: %v", err)
	}

	return tmpDir, func() {
		os.RemoveAll(tmpDir)
	}
}
