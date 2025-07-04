// NeuroScript Version: 0.4.0
// File version: 3
// Purpose: Refactored to be self-contained within the fs package test suite.
// nlines: 180
// risk_rating: LOW
// filename: pkg/tool/fs/tools_fs_move_test.go
package fs

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/aprice2704/neuroscript/pkg/interpreter"
	"github.com/aprice2704/neuroscript/pkg/lang"
	"github.com/aprice2704/neuroscript/pkg/tool"
)

// --- MoveFile Functional Tests ---
func TestToolMoveFileFunctional(t *testing.T) {
	// --- Test Cases ---
	testCases := []fsTestCase{
		{
			name:     "Success: Correct Args",
			toolName: "FS.Move",
			args:     []interface{}{"source.txt", "destination.txt"},
			setupFunc: func(s string) error {
				mustWriteFile(t, filepath.Join(s, "source.txt"), "content")
				return nil
			},
			wantToolErrIs: nil,
			checkFunc: func(t *testing.T, interp tool.Runtime, res interface{}, err error, ctx interface{}) {
				if err != nil {
					t.Fatalf("Unexpected error: %v", err)
				}
				interpImpl := interp.(*interpreter.Interpreter)
				if _, err := os.Stat(filepath.Join(interpImpl.SandboxDir(), "source.txt")); !errors.Is(err, os.ErrNotExist) {
					t.Errorf("Source file 'source.txt' should not exist after move")
				}
				if _, err := os.Stat(filepath.Join(interpImpl.SandboxDir(), "destination.txt")); err != nil {
					t.Errorf("Destination file 'destination.txt' should exist after move: %v", err)
				}
			},
		},
		{
			name:          "Fail: Source does not exist",
			toolName:      "FS.Move",
			args:          []interface{}{"nonexistent_source.txt", "any_dest.txt"},
			wantToolErrIs: lang.ErrFileNotFound,
		},
		{
			name:     "Fail: Destination exists",
			toolName: "FS.Move",
			args:     []interface{}{"src_exists.txt", "dest_exists.txt"},
			setupFunc: func(s string) error {
				mustWriteFile(t, filepath.Join(s, "src_exists.txt"), "content3")
				mustWriteFile(t, filepath.Join(s, "dest_exists.txt"), "content4")
				return nil
			},
			wantToolErrIs: lang.ErrPathExists,
		},
		{
			name:          "Fail: Path outside sandbox (Source)",
			toolName:      "FS.Move",
			args:          []interface{}{"../outside_src.txt", "dest.txt"},
			wantToolErrIs: lang.ErrPathViolation,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			interp := newFsTestInterpreter(t)
			testFsToolHelper(t, interp, tc)
		})
	}
}
