// filename: pkg/tool/fs/tools_fs_stat_test.go
package fs

import (
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolStat(t *testing.T) {
	testFileName := "stat_test_file.txt"
	testDirName := "stat_test_dir"
	testContent := "hello"

	setup := func(s string) error {
		mustWriteFile(t, filepath.Join(s, testFileName), testContent)
		mustMkdir(t, filepath.Join(s, testDirName))
		return nil
	}

	tests := []fsTestCase{
		{
			name:      "Stat Existing File",
			toolName:  "FS.Stat",
			args:      []interface{}{testFileName},
			setupFunc: setup,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				resMap, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected result to be a map, got %T", result)
				}
				if resMap["name"] != testFileName {
					t.Errorf("Expected name %s, got %s", testFileName, resMap["name"])
				}
				if resMap["size_bytes"].(int64) != int64(len(testContent)) {
					t.Errorf("Expected size %d, got %d", len(testContent), resMap["size_bytes"])
				}
				if resMap["is_dir"].(bool) {
					t.Error("Expected is_dir to be false for a file")
				}
			},
		},
		{
			name:      "Stat Existing Directory",
			toolName:  "FS.Stat",
			args:      []interface{}{testDirName},
			setupFunc: setup,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				resMap, ok := result.(map[string]interface{})
				if !ok {
					t.Fatalf("Expected result to be a map, got %T", result)
				}
				if resMap["name"] != testDirName {
					t.Errorf("Expected name %s, got %s", testDirName, resMap["name"])
				}
				if !resMap["is_dir"].(bool) {
					t.Error("Expected is_dir to be true for a directory")
				}
			},
		},
		{
			name:          "Stat Non-Existent Path",
			toolName:      "FS.Stat",
			args:          []interface{}{"no_such_file.txt"},
			setupFunc:     setup,
			wantToolErrIs: lang.ErrFileNotFound,
		},
		{
			name:          "Stat Invalid Path",
			toolName:      "FS.Stat",
			args:          []interface{}{"../invalid.txt"},
			setupFunc:     setup,
			wantToolErrIs: lang.ErrPathViolation,
		},
		{
			name:          "Stat Empty Path",
			toolName:      "FS.Stat",
			args:          []interface{}{""},
			setupFunc:     setup,
			wantToolErrIs: lang.ErrInvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newFsTestInterpreter(t)
			testFsToolHelper(t, interp, tt)
		})
	}
}
