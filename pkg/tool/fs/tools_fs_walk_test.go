// NeuroScript Version: 0.3.1
// File version: 4
// Purpose: Fixed variable shadowing bug by correctly handling the error return from NewDefaultTestInterpreter.
// filename: pkg/tool/fs/tools_fs_walk_test.go
// nlines: 105 // Approximate
// risk_rating: LOW

package fs

import (
	"os"
	"path/filepath"
	"sort"
	"testing"
)

func makeWalkResultChecker(expected []map[string]interface{}) func(t *testing.T, interp *Interpreter, result interface{}, err error, ctx interface{}) {
	return func(t *testing.T, interp *Interpreter, actual interface{}, err error, ctx interface{}) {
		t.Helper()
		AssertNoError(t, err)

		actualSlice, ok := actual.([]map[string]interface{})
		if !ok {
			ifaceSlice, ok2 := actual.([]interface{})
			if !ok2 {
				t.Fatalf("Actual result is not []map[string]interface{} or []interface{}, got %T", actual)
			}
			actualSlice = make([]map[string]interface{}, len(ifaceSlice))
			for i, v := range ifaceSlice {
				actualSlice[i] = v.(map[string]interface{})
			}
		}

		if len(actualSlice) != len(expected) {
			t.Fatalf("Expected %d items, got %d. Actual: %+v", len(expected), len(actualSlice), actualSlice)
		}

		sort.Slice(actualSlice, func(i, j int) bool {
			return actualSlice[i]["path_relative"].(string) < actualSlice[j]["path_relative"].(string)
		})
		sort.Slice(expected, func(i, j int) bool {
			return expected[i]["path_relative"].(string) < expected[j]["path_relative"].(string)
		})

		for i := range expected {
			if actualSlice[i]["path_relative"] != expected[i]["path_relative"] ||
				actualSlice[i]["is_dir"] != expected[i]["is_dir"] {
				t.Errorf("Mismatch at index %d.\nGot:  %+v\nWant: %+v", i, actualSlice[i], expected[i])
			}
		}
	}
}

func TestToolWalkDir(t *testing.T) {
	setupFunc := func(sandboxRoot string) error {
		if err := os.MkdirAll(filepath.Join(sandboxRoot, "dir1", "subdir1"), 0755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(sandboxRoot, "dir1", "file1.txt"), []byte(""), 0644); err != nil {
			return err
		}
		return nil
	}

	tests := []fsTestCase{
		{
			name:		"Walk_from_root",
			toolName:	"FS.Walk",
			args:		MakeArgs("."),
			setupFunc:	setupFunc,
			checkFunc: makeWalkResultChecker([]map[string]interface{}{
				{"path_relative": "dir1", "is_dir": true},
				{"path_relative": filepath.Join("dir1", "file1.txt"), "is_dir": false},
				{"path_relative": filepath.Join("dir1", "subdir1"), "is_dir": true},
			}),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp, err := NewDefaultTestInterpreter(t)
			if err != nil {
				t.Fatalf("NewDefaultTestInterpreter failed: %v", err)
			}
			sb := interp.SandboxDir()

			if tt.setupFunc != nil {
				if err := tt.setupFunc(sb); err != nil {
					t.Fatalf("Setup failed: %v", err)
				}
			}
			testFsToolHelper(t, interp, tt)
		})
	}
}