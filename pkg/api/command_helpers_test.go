// NeuroScript Version: 0.6.0
// File version: 1
// Purpose: Provides tests for the command helper functions.
// filename: pkg/api/command_helpers_test.go
// nlines: 46
// risk_rating: LOW

package api

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/ast"
)

func TestHasCommandBlock(t *testing.T) {
	testCases := []struct {
		name     string
		program  *ast.Program
		expected bool
	}{
		{
			name:     "Nil Program",
			program:  nil,
			expected: false,
		},
		{
			name:     "Program with no commands",
			program:  ast.NewProgram(),
			expected: false,
		},
		{
			name: "Program with one command block",
			program: &ast.Program{
				Commands: []*ast.CommandNode{{}},
			},
			expected: true,
		},
		{
			name: "Program with multiple command blocks",
			program: &ast.Program{
				Commands: []*ast.CommandNode{{}, {}},
			},
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got := HasCommandBlock(tc.program)
			if got != tc.expected {
				t.Errorf("HasCommandBlock() = %v; want %v", got, tc.expected)
			}
		})
	}
}
