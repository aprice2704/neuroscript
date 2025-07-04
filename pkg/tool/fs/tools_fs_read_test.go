// filename: pkg/tool/fs/tools_fs_read_test.go
package fs

import (
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolReadFile(t *testing.T) {
	tests := []fsTestCase{
		{
			name:     "Read Existing File",
			toolName: "FS.Read",
			args:     []interface{}{"read_test.txt"},
			setupFunc: func(s string) error {
				mustWriteFile(t, filepath.Join(s, "read_test.txt"), "hello world")
				return nil
			},
			wantResult: "hello world",
		},
		{
			name:          "Read Non-Existent File",
			toolName:      "FS.Read",
			args:          []interface{}{"non_existent.txt"},
			wantToolErrIs: lang.ErrFileNotFound,
		},
		{
			name:     "Read from Directory",
			toolName: "FS.Read",
			args:     []interface{}{"a_dir"},
			setupFunc: func(s string) error {
				mustMkdir(t, filepath.Join(s, "a_dir"))
				return nil
			},
			wantToolErrIs: lang.ErrPathNotFile,
		},
		{
			name:          "Read Path Outside Sandbox",
			toolName:      "FS.Read",
			args:          []interface{}{"../outside.txt"},
			wantToolErrIs: lang.ErrPathViolation,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newFsTestInterpreter(t)
			testFsToolHelper(t, interp, tt)
		})
	}
}
