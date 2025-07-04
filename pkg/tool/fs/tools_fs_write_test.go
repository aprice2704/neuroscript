// NeuroScript Version: 0.4.0
// File version: 4
// Purpose: Corrected to use the local fs test suite helpers, resolving all compiler errors.
// nlines: 100 // Approximate
// risk_rating: MEDIUM // Test file for a destructive operation
// filename: pkg/tool/fs/tools_fs_write_test.go
package fs

import (
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/lang"
)

func TestToolWriteFileFunctional(t *testing.T) {
	tests := []fsTestCase{
		{
			name:        "Write to new file",
			toolName:    "FS.Write",
			args:        []interface{}{"new_file.txt", "Hello World"},
			wantResult:  "OK",
			wantContent: "Hello World",
		},
		{
			name:     "Overwrite existing file",
			toolName: "FS.Write",
			args:     []interface{}{"existing_file.txt", "New Content"},
			setupFunc: func(s string) error {
				mustWriteFile(t, filepath.Join(s, "existing_file.txt"), "Old Content")
				return nil
			},
			wantResult:  "OK",
			wantContent: "New Content",
		},
		{
			name:        "Create parent directories",
			toolName:    "FS.Write",
			args:        []interface{}{"new/nested/dir/file.txt", "Nested Content"},
			wantResult:  "OK",
			wantContent: "Nested Content",
		},
		{
			name:     "Error on writing to a directory",
			toolName: "FS.Write",
			args:     []interface{}{"a_directory", "some content"},
			setupFunc: func(s string) error {
				mustMkdir(t, filepath.Join(s, "a_directory"))
				return nil
			},
			wantToolErrIs: lang.ErrPathNotFile,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newFsTestInterpreter(t)
			testFsToolHelper(t, interp, tt)
		})
	}
}

func TestToolAppendFileFunctional(t *testing.T) {
	tests := []fsTestCase{
		{
			name:        "Append to new file",
			toolName:    "FS.Append",
			args:        []interface{}{"append_new.txt", "First Line\n"},
			wantResult:  "OK",
			wantContent: "First Line\n",
		},
		{
			name:     "Append to existing file",
			toolName: "FS.Append",
			args:     []interface{}{"append_existing.txt", "Second Line\n"},
			setupFunc: func(s string) error {
				mustWriteFile(t, filepath.Join(s, "append_existing.txt"), "First Line\n")
				return nil
			},
			wantResult:  "OK",
			wantContent: "First Line\nSecond Line\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			interp := newFsTestInterpreter(t)
			testFsToolHelper(t, interp, tt)
		})
	}
}
