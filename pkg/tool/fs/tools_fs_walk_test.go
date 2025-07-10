// filename: pkg/tool/fs/tools_fs_walk_test.go
package fs

import (
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/tool"
)

func TestToolWalk(t *testing.T) {
	setup := func(s string) error {
		// Structure:
		// /
		// |- file1.txt
		// |- dir1/
		//    |- file2.txt
		// |- dir2/
		mustWriteFile(t, filepath.Join(s, "file1.txt"), "f1")
		mustMkdir(t, filepath.Join(s, "dir1"))
		mustWriteFile(t, filepath.Join(s, "dir1", "file2.txt"), "f2")
		mustMkdir(t, filepath.Join(s, "dir2"))
		return nil
	}

	tests := []fsTestCase{
		{
			name:      "Walk directory",
			toolName:  "Walk",
			args:      []interface{}{"."},
			setupFunc: setup,
			checkFunc: func(t *testing.T, interp tool.Runtime, result interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				resList, ok := result.([]map[string]interface{})
				if !ok {
					t.Fatalf("Expected result to be a slice of maps, got %T", result)
				}
				// Expect 4 entries: file1.txt, dir1, dir1/file2.txt, dir2
				if len(resList) != 4 {
					t.Errorf("Expected 4 entries from walk, got %d", len(resList))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newFsTestInterpreter(t)
			testFsToolHelper(t, interp, tt)
		})
	}
}
