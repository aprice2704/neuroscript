// NeuroScript Version: 0.1.0
// File version: 3
// Purpose: Provides tests for tool registry. Fixed missing Func in registration test cases.
// filename: pkg/tool/tools_registry_test.go
// nlines: 104

package tool

import (
	"testing"

	"github.com/aprice2704/neuroscript/pkg/types"
)

// TestToolRegistry_LookupCanonicalization verifies that the registry can find
// a tool using various non-canonical name formats during lookup.
func TestToolRegistry_LookupCanonicalization(t *testing.T) {
	// 1. Setup
	registry := NewToolRegistry(nil)
	testTool := ToolImplementation{
		Spec: ToolSpec{
			Group: "test",
			Name:  "dummy",
		},
		Func: func(rt Runtime, args []interface{}) (interface{}, error) {
			return "dummy output", nil
		},
	}

	registeredTool, err := registry.RegisterTool(testTool)
	if err != nil {
		t.Fatalf("RegisterTool failed unexpectedly: %v", err)
	}

	// ASSERT: The FullName stored on the struct itself should be the canonical name.
	expectedCanonicalName := "tool.test.dummy"
	if registeredTool.FullName != types.FullName(expectedCanonicalName) {
		t.Fatalf("Expected registered tool FullName to be '%s', but got '%s'", expectedCanonicalName, registeredTool.FullName)
	}

	// 2. Test Cases for Lookup
	testCases := []struct {
		lookupName string
	}{
		{"test.dummy"},                // No prefix
		{"tool.test.dummy"},           // Correct canonical prefix
		{"tool.tool.test.dummy"},      // Double prefix
		{"tool.tool.tool.test.dummy"}, // Triple prefix
	}

	// 3. Execution and Assertion
	for _, tc := range testCases {
		t.Run(tc.lookupName, func(t *testing.T) {
			foundTool, found := registry.GetTool(types.FullName(tc.lookupName))

			if !found {
				t.Errorf("expected to find tool with lookup name '%s', but it was not found", tc.lookupName)
				return
			}
			if foundTool.FullName != types.FullName(expectedCanonicalName) {
				t.Errorf("retrieved tool has wrong FullName: expected '%s', got '%s'", expectedCanonicalName, foundTool.FullName)
			}
		})
	}
}

// TestToolRegistry_RegistrationCanonicalization verifies that tools registered
// with non-canonical group/name combos are stored with a correct canonical FullName.
func TestToolRegistry_RegistrationCanonicalization(t *testing.T) {
	dummyFunc := func(rt Runtime, args []interface{}) (interface{}, error) { return nil, nil }

	testCases := []struct {
		name         string
		inputTool    ToolImplementation
		expectedName types.FullName
	}{
		{
			name: "Clean simple name",
			inputTool: ToolImplementation{
				Spec: ToolSpec{Group: "fs", Name: "read"},
				Func: dummyFunc,
			},
			expectedName: "tool.fs.read",
		},
		{
			name: "Group already has tool prefix",
			inputTool: ToolImplementation{
				Spec: ToolSpec{Group: "tool.fs", Name: "write"},
				Func: dummyFunc,
			},
			expectedName: "tool.fs.write",
		},
		{
			name: "Group has duplicated tool prefix",
			inputTool: ToolImplementation{
				Spec: ToolSpec{Group: "tool.tool.fs", Name: "list"},
				Func: dummyFunc,
			},
			expectedName: "tool.fs.list",
		},
		{
			name: "Group contains dots",
			inputTool: ToolImplementation{
				Spec: ToolSpec{Group: "system.io", Name: "get"},
				Func: dummyFunc,
			},
			expectedName: "tool.system.io.get",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			registry := NewToolRegistry(nil)
			registeredTool, err := registry.RegisterTool(tc.inputTool)
			if err != nil {
				t.Fatalf("RegisterTool failed unexpectedly: %v", err)
			}

			// Assert that the returned tool has the correct canonical name.
			if registeredTool.FullName != tc.expectedName {
				t.Errorf("expected canonical name to be '%s', but got '%s'", tc.expectedName, registeredTool.FullName)
			}

			// Assert that the tool can be retrieved using its canonical name.
			_, found := registry.GetTool(tc.expectedName)
			if !found {
				t.Errorf("could not look up tool using its new canonical name '%s'", tc.expectedName)
			}
		})
	}
}
